package specdoc

import (
	"bytes"
	"strings"
	"unicode/utf8"

	"github.com/chiisen/speclet/internal/i18n"
)

type OperationKind int

const (
	Added OperationKind = iota + 1
	Modified
	Removed
)

type Operation struct {
	Kind OperationKind
	Name string
	Raw  []byte
}

// ParseDelta parses the explicitly marked delta sections and requirement blocks.
func ParseDelta(src []byte) ([]Operation, error) {
	if !utf8.Valid(src) {
		return nil, i18n.Errorf("delta is not valid UTF-8")
	}
	var operations []Operation
	var current *Operation
	var raw bytes.Buffer
	var section OperationKind
	var fenceChar byte
	fenceLength := 0
	lineNumber := 0
	paragraphStart := 0
	paragraphActive := false
	seenNames := make(map[string]bool)
	flush := func() error {
		if current == nil {
			return nil
		}
		if seenNames[current.Name] {
			return i18n.Errorf("duplicate delta Requirement %q", current.Name)
		}
		seenNames[current.Name] = true
		current.Raw = append([]byte(nil), raw.Bytes()...)
		operations = append(operations, *current)
		current = nil
		raw.Reset()
		return nil
	}

	for _, rawLine := range bytes.SplitAfter(src, []byte("\n")) {
		if len(rawLine) == 0 {
			continue
		}
		lineNumber++
		line := strings.TrimSuffix(strings.TrimSuffix(string(rawLine), "\n"), "\r")
		if lineNumber == 1 {
			line = strings.TrimPrefix(line, "\uFEFF")
		}
		if markdownFence(line, &fenceChar, &fenceLength) {
			if current != nil {
				raw.Write(rawLine)
			}
			paragraphActive = false
			continue
		}
		if fenceChar != 0 {
			if current != nil {
				raw.Write(rawLine)
			}
			paragraphActive = false
			continue
		}
		level, title, isHeading := parseHeading(line)
		marker, name := structuralMarker(title)
		if marker != "" {
			if marker == "Requirement" && level != 3 {
				return nil, i18n.Errorf("line %d: delta Requirement heading must use level 3", lineNumber)
			}
			if marker == "Scenario" && level != 4 {
				return nil, i18n.Errorf("line %d: delta Scenario heading must use level 4", lineNumber)
			}
			if name == "" {
				return nil, i18n.Errorf("line %d: %s name is empty", lineNumber, marker)
			}
		}
		if isHeading && level == 2 {
			if err := flush(); err != nil {
				return nil, err
			}
			switch title {
			case "ADDED":
				section = Added
			case "MODIFIED":
				section = Modified
			case "REMOVED":
				section = Removed
			default:
				return nil, i18n.Errorf("line %d: unknown delta section %q", lineNumber, title)
			}
			paragraphActive = false
			continue
		}
		if isHeading && level <= 1 {
			if err := flush(); err != nil {
				return nil, err
			}
			section = 0
			paragraphActive = false
			continue
		}
		if isSetextUnderline(line) && paragraphActive {
			if current != nil {
				raw.Truncate(paragraphStart)
				if err := flush(); err != nil {
					return nil, err
				}
			}
			section = 0
			paragraphActive = false
			continue
		}
		if marker == "Requirement" {
			if section == 0 {
				return nil, i18n.Errorf("line %d: Requirement is outside a delta section", lineNumber)
			}
			if err := flush(); err != nil {
				return nil, err
			}
			current = &Operation{Kind: section, Name: name}
			raw.Write(rawLine)
			paragraphActive = false
			continue
		}
		if marker == "Scenario" && current == nil {
			return nil, i18n.Errorf("line %d: Scenario is outside a Requirement", lineNumber)
		}
		if strings.TrimSpace(line) == "" || isHeading || isMarkdownBlockStart(line) {
			paragraphActive = false
		} else {
			if !paragraphActive {
				paragraphStart = raw.Len()
			}
			paragraphActive = true
		}
		if current != nil {
			raw.Write(rawLine)
		}
	}
	if fenceChar != 0 {
		return nil, i18n.Errorf("unclosed code fence")
	}
	if err := flush(); err != nil {
		return nil, err
	}
	return operations, nil
}

// ApplyDelta validates every operation, then applies it while retaining unrelated Markdown.
func ApplyDelta(spec, delta []byte) ([]byte, error) {
	doc, err := ParseSpec(spec)
	if err != nil {
		return nil, i18n.Errorf("parse formal spec: %w", err)
	}
	operations, err := ParseDelta(delta)
	if err != nil {
		return nil, i18n.Errorf("parse delta: %w", err)
	}
	existing := make(map[string]bool, len(doc.Requirements))
	for _, requirement := range doc.Requirements {
		existing[requirement.Name] = true
	}
	byName := make(map[string]Operation, len(operations))
	for _, operation := range operations {
		if operation.Kind == Added && existing[operation.Name] {
			return nil, i18n.Errorf("cannot add Requirement %q: it already exists", operation.Name)
		}
		if (operation.Kind == Modified || operation.Kind == Removed) && !existing[operation.Name] {
			return nil, i18n.Errorf("cannot %s Requirement %q: it does not exist", operationName(operation.Kind), operation.Name)
		}
		byName[operation.Name] = operation
	}

	var output bytes.Buffer
	for _, part := range doc.segments {
		ensureSegmentBoundary(&output)
		if !part.isRequirement {
			output.Write(part.raw)
			continue
		}
		operation, hasOperation := byName[part.requirementName]
		if !hasOperation {
			output.Write(part.raw)
			continue
		}
		switch operation.Kind {
		case Modified:
			output.Write(normalizeDeltaHeadings(operation.Raw))
		case Removed:
			// The requirement block is omitted; adjacent ordinary Markdown remains.
		}
	}
	for _, operation := range operations {
		if operation.Kind != Added {
			continue
		}
		writeBlockSeparator(&output)
		output.Write(normalizeDeltaHeadings(operation.Raw))
	}
	result := output.Bytes()
	if _, err := ParseSpec(result); err != nil {
		return nil, i18n.Errorf("delta produces invalid formal spec: %w", err)
	}
	return result, nil
}

func normalizeDeltaHeadings(src []byte) []byte {
	var output bytes.Buffer
	var fenceChar byte
	fenceLength := 0
	for _, rawLine := range bytes.SplitAfter(src, []byte("\n")) {
		if len(rawLine) == 0 {
			continue
		}
		lineBytes := rawLine
		ending := []byte(nil)
		if bytes.HasSuffix(lineBytes, []byte("\r\n")) {
			lineBytes = lineBytes[:len(lineBytes)-2]
			ending = []byte("\r\n")
		} else if bytes.HasSuffix(lineBytes, []byte("\n")) {
			lineBytes = lineBytes[:len(lineBytes)-1]
			ending = []byte("\n")
		}
		line := string(lineBytes)
		if markdownFence(line, &fenceChar, &fenceLength) || fenceChar != 0 {
			output.Write(rawLine)
			continue
		}
		level, title, isHeading := parseHeading(line)
		marker, _ := structuralMarker(title)
		if isHeading && marker == "Requirement" && level == 3 {
			output.WriteString(replaceHeadingLevel(line, 3, 2))
		} else if isHeading && marker == "Scenario" && level == 4 {
			output.WriteString(replaceHeadingLevel(line, 4, 3))
		} else {
			output.Write(rawLine)
			continue
		}
		output.Write(ending)
	}
	return output.Bytes()
}

func replaceHeadingLevel(line string, from, to int) string {
	trimmed := strings.TrimLeft(line, " ")
	indent := len(line) - len(trimmed)
	return line[:indent] + strings.Repeat("#", to) + trimmed[from:]
}

func writeBlockSeparator(output *bytes.Buffer) {
	if output.Len() == 0 {
		return
	}
	data := output.Bytes()
	if bytes.Equal(data, []byte("\uFEFF")) {
		return
	}
	if data[len(data)-1] != '\n' {
		output.WriteString("\n\n")
		return
	}
	if len(data) < 2 || data[len(data)-2] != '\n' {
		output.WriteByte('\n')
	}
}

func ensureSegmentBoundary(output *bytes.Buffer) {
	if output.Len() > 0 && !bytes.Equal(output.Bytes(), []byte("\uFEFF")) && output.Bytes()[output.Len()-1] != '\n' {
		output.WriteString("\n\n")
	}
}

func operationName(kind OperationKind) string {
	switch kind {
	case Modified:
		return "modify"
	case Removed:
		return "remove"
	default:
		return "apply"
	}
}

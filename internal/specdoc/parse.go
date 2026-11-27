package specdoc

import (
	"bytes"
	"strings"
	"unicode/utf8"

	"github.com/chiisen/speclet/internal/i18n"
)

type Requirement struct {
	Name string
	Raw  []byte
}

type Document struct {
	Prefix       []byte
	Requirements []Requirement
	Suffix       []byte
	segments     []segment
}

type segment struct {
	raw             []byte
	requirementName string
	isRequirement   bool
}

// ParseSpec locates formal requirement blocks while retaining all source bytes.
func ParseSpec(src []byte) (Document, error) {
	if !utf8.Valid(src) {
		return Document{}, i18n.Errorf("formal spec is not valid UTF-8")
	}
	var doc Document
	var segments []segment
	var current bytes.Buffer
	currentRequirement := ""
	seenRequirement := false
	seenNames := make(map[string]bool)
	var fenceChar byte
	fenceLength := 0
	lineNumber := 0
	paragraphStart := 0
	paragraphActive := false

	flush := func() {
		if current.Len() == 0 {
			return
		}
		part := segment{raw: append([]byte(nil), current.Bytes()...)}
		if currentRequirement != "" {
			part.requirementName = currentRequirement
			part.isRequirement = true
			doc.Requirements = append(doc.Requirements, Requirement{Name: currentRequirement, Raw: append([]byte(nil), part.raw...)})
		} else if seenRequirement {
			doc.Suffix = append(doc.Suffix, part.raw...)
		} else {
			doc.Prefix = append(doc.Prefix, part.raw...)
		}
		segments = append(segments, part)
		current.Reset()
		currentRequirement = ""
	}

	for _, rawLine := range bytes.SplitAfter(src, []byte("\n")) {
		if len(rawLine) == 0 {
			continue
		}
		lineNumber++
		line := strings.TrimSuffix(strings.TrimSuffix(string(rawLine), "\n"), "\r")
		if lineNumber == 1 && strings.HasPrefix(line, "\uFEFF") {
			current.WriteString("\uFEFF")
			flush()
			rawLine = rawLine[len("\uFEFF"):]
			line = strings.TrimPrefix(line, "\uFEFF")
			if len(rawLine) == 0 {
				continue
			}
		}
		if markdownFence(line, &fenceChar, &fenceLength) {
			current.Write(rawLine)
			paragraphActive = false
			continue
		}
		if fenceChar != 0 {
			current.Write(rawLine)
			paragraphActive = false
			continue
		}
		level, title, isHeading := parseHeading(line)
		if isSetextUnderline(line) && currentRequirement != "" && paragraphActive {
			contents := current.Bytes()
			heading := append([]byte(nil), contents[paragraphStart:]...)
			current.Truncate(paragraphStart)
			flush()
			current.Write(heading)
			current.Write(rawLine)
			paragraphActive = false
			continue
		}
		marker, markerName := structuralMarker(title)
		if marker != "" {
			if marker == "Requirement" && level != 2 {
				return Document{}, i18n.Errorf("line %d: Requirement heading must use level 2", lineNumber)
			}
			if marker == "Scenario" && level != 3 {
				return Document{}, i18n.Errorf("line %d: Scenario heading must use level 3", lineNumber)
			}
			if markerName == "" {
				return Document{}, i18n.Errorf("line %d: %s name is empty", lineNumber, marker)
			}
			if marker == "Scenario" && currentRequirement == "" {
				return Document{}, i18n.Errorf("line %d: Scenario is outside a Requirement", lineNumber)
			}
		}
		if isHeading && level <= 2 {
			flush()
			paragraphActive = false
			if marker == "Requirement" {
				if seenNames[markerName] {
					return Document{}, i18n.Errorf("line %d: duplicate Requirement %q", lineNumber, markerName)
				}
				seenNames[markerName] = true
				seenRequirement = true
				currentRequirement = markerName
			}
		}
		if strings.TrimSpace(line) == "" || isHeading || isMarkdownBlockStart(line) {
			paragraphActive = false
		} else if !paragraphActive {
			paragraphStart = current.Len()
			paragraphActive = true
		}
		current.Write(rawLine)
	}
	if fenceChar != 0 {
		return Document{}, i18n.Errorf("unclosed code fence")
	}
	flush()
	doc.segments = segments
	return doc, nil
}

func parseHeading(line string) (level int, title string, ok bool) {
	trimmed := strings.TrimLeft(line, " ")
	indent := len(line) - len(trimmed)
	if indent > 3 || !strings.HasPrefix(trimmed, "#") {
		return 0, "", false
	}
	for level < len(trimmed) && trimmed[level] == '#' {
		level++
	}
	if level == 0 || level >= len(trimmed) || (trimmed[level] != ' ' && trimmed[level] != '\t') {
		return 0, "", false
	}
	return level, strings.TrimSpace(trimmed[level:]), true
}

func isSetextUnderline(line string) bool {
	trimmed := strings.TrimLeft(line, " ")
	if len(line)-len(trimmed) > 3 || len(trimmed) == 0 {
		return false
	}
	marker := trimmed[0]
	if marker != '=' && marker != '-' {
		return false
	}
	length := 0
	for length < len(trimmed) && trimmed[length] == marker {
		length++
	}
	return length > 0 && strings.TrimSpace(trimmed[length:]) == ""
}

func isMarkdownBlockStart(line string) bool {
	trimmed := strings.TrimLeft(line, " ")
	if len(line)-len(trimmed) > 3 {
		return false
	}
	if strings.HasPrefix(trimmed, ">") || strings.HasPrefix(trimmed, "- ") || strings.HasPrefix(trimmed, "+ ") || strings.HasPrefix(trimmed, "* ") {
		return true
	}
	index := 0
	for index < len(trimmed) && trimmed[index] >= '0' && trimmed[index] <= '9' {
		index++
	}
	return index > 0 && index+1 < len(trimmed) && (trimmed[index] == '.' || trimmed[index] == ')') && (trimmed[index+1] == ' ' || trimmed[index+1] == '\t')
}

func structuralMarker(title string) (marker, name string) {
	for _, candidate := range []string{"Requirement", "Scenario"} {
		prefix := candidate + ":"
		if strings.HasPrefix(title, prefix) {
			return candidate, strings.TrimSpace(strings.TrimPrefix(title, prefix))
		}
	}
	return "", ""
}

func markdownFence(line string, fenceChar *byte, fenceLength *int) bool {
	trimmed := strings.TrimLeft(line, " ")
	indent := len(line) - len(trimmed)
	if indent > 3 || len(trimmed) < 3 || (trimmed[0] != '`' && trimmed[0] != '~') {
		return false
	}
	marker := trimmed[0]
	length := 0
	for length < len(trimmed) && trimmed[length] == marker {
		length++
	}
	if length < 3 {
		return false
	}
	if marker == '`' && strings.Contains(trimmed[length:], "`") {
		return false
	}
	if *fenceChar == 0 {
		*fenceChar, *fenceLength = marker, length
		return true
	}
	if marker == *fenceChar && length >= *fenceLength && strings.TrimSpace(trimmed[length:]) == "" {
		*fenceChar, *fenceLength = 0, 0
	}
	return true
}

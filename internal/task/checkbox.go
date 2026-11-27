package task

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"unicode/utf8"
)

var checkboxLine = regexp.MustCompile(`^([ \t]*)[-*+] \[([ xX])\](?:[ \t]|$)`)
var taskIDLine = regexp.MustCompile(`^[-*+] \[[ xX]\] ([^ \t]+)(?:[ \t]+.*)?$`)

// Count returns completed and total flat checkboxes in a task document.
func Count(src []byte) (done, total int, err error) {
	if !utf8.Valid(src) {
		return 0, 0, fmt.Errorf("tasks document is not valid UTF-8")
	}
	var fenceChar byte
	fenceLength := 0
	seenIDs := make(map[string]bool)
	for lineNumber, raw := range bytes.Split(src, []byte("\n")) {
		line := strings.TrimSuffix(string(raw), "\r")
		if lineNumber == 0 {
			line = strings.TrimPrefix(line, "\uFEFF")
		}
		trimmed := strings.TrimLeft(line, " ")
		indent := len(line) - len(trimmed)
		if indent <= 3 && len(trimmed) >= 3 && (strings.HasPrefix(trimmed, "```") || strings.HasPrefix(trimmed, "~~~")) {
			marker := trimmed[0]
			markerLength := 0
			for markerLength < len(trimmed) && trimmed[markerLength] == marker {
				markerLength++
			}
			if marker != '`' || !strings.Contains(trimmed[markerLength:], "`") {
				if fenceChar == 0 {
					fenceChar, fenceLength = marker, markerLength
				} else if marker == fenceChar && markerLength >= fenceLength && strings.TrimSpace(trimmed[markerLength:]) == "" {
					fenceChar, fenceLength = 0, 0
				}
				continue
			}
		}
		if fenceChar != 0 {
			continue
		}
		match := checkboxLine.FindStringSubmatch(line)
		if match == nil {
			continue
		}
		if match[1] != "" {
			return 0, 0, fmt.Errorf("nested checkbox at line %d is unsupported", lineNumber+1)
		}
		idMatch := taskIDLine.FindStringSubmatch(line)
		if idMatch == nil {
			return 0, 0, fmt.Errorf("task checkbox at line %d has no task ID", lineNumber+1)
		}
		if seenIDs[idMatch[1]] {
			return 0, 0, fmt.Errorf("task ID %q is duplicated at line %d", idMatch[1], lineNumber+1)
		}
		seenIDs[idMatch[1]] = true
		total++
		if match[2] != " " {
			done++
		}
	}
	if fenceChar != 0 {
		return 0, 0, fmt.Errorf("unclosed code fence")
	}
	return done, total, nil
}

// SetDone updates only the checkbox for id after validating the complete task file.
func SetDone(tasksPath, id string, done bool) error {
	if strings.TrimSpace(id) == "" || strings.ContainsAny(id, " \t\r\n") {
		return fmt.Errorf("invalid task ID %q", id)
	}
	info, err := os.Lstat(tasksPath)
	if err != nil {
		return fmt.Errorf("inspect tasks file: %w", err)
	}
	if info.Mode()&os.ModeSymlink != 0 || !info.Mode().IsRegular() {
		return fmt.Errorf("tasks file is not a regular file")
	}
	parentInfo, err := os.Lstat(filepath.Dir(tasksPath))
	if err != nil {
		return fmt.Errorf("inspect tasks directory: %w", err)
	}
	if parentInfo.Mode()&os.ModeSymlink != 0 || !parentInfo.IsDir() {
		return fmt.Errorf("tasks directory is not a regular directory")
	}
	src, err := os.ReadFile(tasksPath)
	if err != nil {
		return fmt.Errorf("read tasks file: %w", err)
	}
	if !utf8.Valid(src) {
		return fmt.Errorf("tasks document is not valid UTF-8")
	}

	chunks := bytes.SplitAfter(src, []byte("\n"))
	var fenceChar byte
	fenceLength := 0
	matchedLine := -1
	statusOffset := 0
	seenIDs := make(map[string]bool)
	for lineNumber, chunk := range chunks {
		if len(chunk) == 0 {
			continue
		}
		body := chunk
		if body[len(body)-1] == '\n' {
			body = body[:len(body)-1]
		}
		line := strings.TrimSuffix(string(body), "\r")
		lineOffset := 0
		if lineNumber == 0 && strings.HasPrefix(line, "\uFEFF") {
			line = strings.TrimPrefix(line, "\uFEFF")
			lineOffset = len("\uFEFF")
		}
		trimmed := strings.TrimLeft(line, " ")
		indent := len(line) - len(trimmed)
		if indent <= 3 && len(trimmed) >= 3 && (strings.HasPrefix(trimmed, "```") || strings.HasPrefix(trimmed, "~~~")) {
			marker := trimmed[0]
			markerLength := 0
			for markerLength < len(trimmed) && trimmed[markerLength] == marker {
				markerLength++
			}
			if marker != '`' || !strings.Contains(trimmed[markerLength:], "`") {
				if fenceChar == 0 {
					fenceChar, fenceLength = marker, markerLength
				} else if marker == fenceChar && markerLength >= fenceLength && strings.TrimSpace(trimmed[markerLength:]) == "" {
					fenceChar, fenceLength = 0, 0
				}
				continue
			}
		}
		if fenceChar != 0 {
			continue
		}
		match := checkboxLine.FindStringSubmatch(line)
		if match == nil {
			continue
		}
		if match[1] != "" {
			return fmt.Errorf("nested checkbox at line %d is unsupported", lineNumber+1)
		}
		idMatch := taskIDLine.FindStringSubmatch(line)
		if idMatch == nil {
			return fmt.Errorf("task checkbox at line %d has no task ID", lineNumber+1)
		}
		if seenIDs[idMatch[1]] {
			return fmt.Errorf("task ID %q is duplicated", idMatch[1])
		}
		seenIDs[idMatch[1]] = true
		if idMatch[1] == id {
			matchedLine = lineNumber
			statusOffset = lineOffset
		}
	}
	if fenceChar != 0 {
		return fmt.Errorf("unclosed code fence")
	}
	if matchedLine < 0 {
		return fmt.Errorf("task ID %q was not found", id)
	}

	line := chunks[matchedLine]
	status := byte(' ')
	if done {
		status = 'x'
	}
	line[statusOffset+3] = status
	chunks[matchedLine] = line
	updated := bytes.Join(chunks, nil)
	file, err := os.CreateTemp(filepath.Dir(tasksPath), ".speclet-task-*")
	if err != nil {
		return fmt.Errorf("stage tasks file: %w", err)
	}
	tempPath := file.Name()
	defer os.Remove(tempPath)
	if _, err := file.Write(updated); err != nil {
		_ = file.Close()
		return fmt.Errorf("write staged tasks file: %w", err)
	}
	if err := file.Chmod(info.Mode().Perm()); err != nil {
		_ = file.Close()
		return fmt.Errorf("set staged tasks file mode: %w", err)
	}
	if err := file.Sync(); err != nil {
		_ = file.Close()
		return fmt.Errorf("sync staged tasks file: %w", err)
	}
	if err := file.Close(); err != nil {
		return fmt.Errorf("close staged tasks file: %w", err)
	}
	if err := os.Rename(tempPath, tasksPath); err != nil {
		return fmt.Errorf("replace tasks file: %w", err)
	}
	return nil
}

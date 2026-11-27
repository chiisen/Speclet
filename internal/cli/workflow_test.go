package cli

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestWorkflowInitializesTracksAndArchivesChange(t *testing.T) {
	cwd := t.TempDir()
	run := func(args ...string) (int, string, string) {
		t.Helper()
		var stdout, stderr bytes.Buffer
		code := Run(args, &stdout, &stderr, cwd)
		return code, stdout.String(), stderr.String()
	}
	if code, stdout, stderr := run("init"); code != 0 || stdout == "" || stderr != "" {
		t.Fatalf("init: code=%d stdout=%q stderr=%q", code, stdout, stderr)
	}
	name := "付款功能 甲"
	if code, stdout, stderr := run("change", "new", name); code != 0 || !strings.Contains(stdout, name) || stderr != "" {
		t.Fatalf("change new: code=%d stdout=%q stderr=%q", code, stdout, stderr)
	}
	root := filepath.Join(cwd, "speclet")
	changeDir := filepath.Join(root, "changes", name)
	if err := os.WriteFile(filepath.Join(changeDir, "proposal.md"), []byte("# Proposal\n\nAdd payment flow.\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(changeDir, "tasks.md"), []byte("- [ ] T001 Implement flow\n- [ ] T002 Update spec\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	deltaPath := filepath.Join(changeDir, "specs", "payments", "spec.md")
	if err := os.MkdirAll(filepath.Dir(deltaPath), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(deltaPath, []byte("## ADDED\n### Requirement: Payment can be started\nA customer can start a payment.\n#### Scenario: Start accepted\n- **WHEN** valid details are submitted\n- **THEN** a payment is created\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if code, stdout, stderr := run("change", "list"); code != 0 || !strings.Contains(stdout, "0/2 tasks done") || stderr != "" {
		t.Fatalf("change list: code=%d stdout=%q stderr=%q", code, stdout, stderr)
	}
	if code, stdout, stderr := run("change", "show", name); code != 0 || !strings.Contains(stdout, "proposal.md") || !strings.Contains(stdout, "tasks.md") || !strings.Contains(stdout, "spec.md") || stderr != "" {
		t.Fatalf("change show: code=%d stdout=%q stderr=%q", code, stdout, stderr)
	}
	for _, id := range []string{"T001", "T002"} {
		if code, stdout, stderr := run("task", "done", name, id); code != 0 || !strings.Contains(stdout, id) || stderr != "" {
			t.Fatalf("task done %s: code=%d stdout=%q stderr=%q", id, code, stdout, stderr)
		}
	}
	if code, stdout, stderr := run("archive", name); code != 0 || stdout == "" || stderr != "" {
		t.Fatalf("archive: code=%d stdout=%q stderr=%q", code, stdout, stderr)
	}
	formal, err := os.ReadFile(filepath.Join(root, "specs", "payments", "spec.md"))
	if err != nil {
		t.Fatal(err)
	}
	for _, expected := range []string{"## Requirement: Payment can be started", "### Scenario: Start accepted", "- **THEN** a payment is created"} {
		if !strings.Contains(string(formal), expected) {
			t.Errorf("formal spec missing %q: %q", expected, formal)
		}
	}
	archivePath := filepath.Join(root, "changes", "archive", time.Now().Format("2006-01-02")+"-"+name)
	if _, err := os.Stat(archivePath); err != nil {
		t.Fatalf("archived change missing: %v", err)
	}
}

func TestRunHelpDocumentsMVPCommands(t *testing.T) {
	var stdout, stderr bytes.Buffer
	code := Run([]string{"help"}, &stdout, &stderr, t.TempDir())
	if code != 0 || stdout.Len() == 0 || stderr.Len() != 0 {
		t.Fatalf("help: code=%d stdout=%q stderr=%q", code, stdout.String(), stderr.String())
	}
	for _, expected := range []string{"speclet init", "speclet change new", "speclet task done", "speclet archive"} {
		if !strings.Contains(stdout.String(), expected) {
			t.Errorf("help output missing %q: %q", expected, stdout.String())
		}
	}
}

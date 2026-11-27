package cli

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestRunInitCreatesProjectAndWritesSuccessToStdout(t *testing.T) {
	cwd := t.TempDir()
	var stdout, stderr bytes.Buffer

	code := Run([]string{"init"}, &stdout, &stderr, cwd)
	if code != 0 {
		t.Fatalf("Run() = %d, stderr = %q", code, stderr.String())
	}
	if stdout.Len() == 0 || stderr.Len() != 0 {
		t.Fatalf("stdout = %q, stderr = %q; want success on stdout only", stdout.String(), stderr.String())
	}
	if _, err := os.Stat(filepath.Join(cwd, "speclet", "config.yaml")); err != nil {
		t.Fatalf("init did not create config.yaml: %v", err)
	}
	stdout.Reset()
	stderr.Reset()
	code = Run([]string{"init"}, &stdout, &stderr, cwd)
	if code != 0 || !strings.Contains(strings.ToLower(stdout.String()), "already initialized") || stderr.Len() != 0 {
		t.Fatalf("second init: code=%d stdout=%q stderr=%q", code, stdout.String(), stderr.String())
	}
}

func TestRunRejectsInvalidCommandsOnStderr(t *testing.T) {
	for _, args := range [][]string{nil, {"unknown"}, {"init", "one", "two"}} {
		var stdout, stderr bytes.Buffer
		code := Run(args, &stdout, &stderr, t.TempDir())
		if code == 0 {
			t.Errorf("Run(%v) = 0, want failure", args)
		}
		if stdout.Len() != 0 || stderr.Len() == 0 {
			t.Errorf("Run(%v): stdout = %q, stderr = %q; want error on stderr only", args, stdout.String(), stderr.String())
		}
	}
}

func TestRunChangeNewListAndShow(t *testing.T) {
	cwd := t.TempDir()
	var stdout, stderr bytes.Buffer
	if code := Run([]string{"init"}, &stdout, &stderr, cwd); code != 0 {
		t.Fatalf("init = %d, stderr = %q", code, stderr.String())
	}

	stdout.Reset()
	if code := Run([]string{"change", "new", "feature one"}, &stdout, &stderr, cwd); code != 0 {
		t.Fatalf("change new = %d, stderr = %q", code, stderr.String())
	}
	if !strings.Contains(stdout.String(), "feature one") || stderr.Len() != 0 {
		t.Fatalf("change new output: stdout = %q stderr = %q", stdout.String(), stderr.String())
	}

	stdout.Reset()
	if code := Run([]string{"change", "list"}, &stdout, &stderr, cwd); code != 0 {
		t.Fatalf("change list = %d, stderr = %q", code, stderr.String())
	}
	if !strings.Contains(stdout.String(), "feature one") || !strings.Contains(stdout.String(), "0/0") {
		t.Fatalf("change list output = %q, want feature name and task count", stdout.String())
	}

	stdout.Reset()
	if code := Run([]string{"change", "show", "feature one"}, &stdout, &stderr, cwd); code != 0 {
		t.Fatalf("change show = %d, stderr = %q", code, stderr.String())
	}
	for _, expected := range []string{"feature one", "proposal.md", "tasks.md"} {
		if !strings.Contains(stdout.String(), expected) {
			t.Errorf("change show output %q missing %q", stdout.String(), expected)
		}
	}
}

func TestRunChangeListEmptyAndShowMissing(t *testing.T) {
	cwd := t.TempDir()
	var stdout, stderr bytes.Buffer
	if code := Run([]string{"init"}, &stdout, &stderr, cwd); code != 0 {
		t.Fatal(stderr.String())
	}
	stdout.Reset()
	if code := Run([]string{"change", "list"}, &stdout, &stderr, cwd); code != 0 || !strings.Contains(strings.ToLower(stdout.String()), "no active changes") {
		t.Fatalf("empty change list: code=%d stdout=%q stderr=%q", code, stdout.String(), stderr.String())
	}
	stdout.Reset()
	stderr.Reset()
	if code := Run([]string{"change", "show", "missing"}, &stdout, &stderr, cwd); code == 0 || stdout.Len() != 0 || stderr.Len() == 0 {
		t.Fatalf("missing show: code=%d stdout=%q stderr=%q", code, stdout.String(), stderr.String())
	}
}

func TestRunChangeListNamesOnly(t *testing.T) {
	for _, tc := range []struct {
		name      string
		args      []string
		changes   []string
		want      string
		wantError bool
	}{
		{
			name:    "sorted names with spaces",
			args:    []string{"--names-only"},
			changes: []string{"zeta", "alpha one"},
			want:    "alpha one\nzeta\n",
		},
		{name: "empty list", args: []string{"--names-only"}},
		{name: "unknown flag", args: []string{"--unknown"}, wantError: true},
		{name: "extra argument", args: []string{"--names-only", "extra"}, wantError: true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			cwd := t.TempDir()
			var stdout, stderr bytes.Buffer
			if code := Run([]string{"init"}, &stdout, &stderr, cwd); code != 0 {
				t.Fatalf("init = %d, stderr = %q", code, stderr.String())
			}
			for _, name := range tc.changes {
				if code := Run([]string{"change", "new", name}, &stdout, &stderr, cwd); code != 0 {
					t.Fatalf("change new %q = %d, stderr = %q", name, code, stderr.String())
				}
			}
			stdout.Reset()
			stderr.Reset()
			code := Run(append([]string{"change", "list"}, tc.args...), &stdout, &stderr, cwd)
			if tc.wantError {
				if code == 0 || stdout.Len() != 0 || !strings.Contains(stderr.String(), "usage: speclet change list") {
					t.Fatalf("code = %d, stdout = %q, stderr = %q; want usage on stderr with a nonzero exit", code, stdout.String(), stderr.String())
				}
				return
			}
			if code != 0 || stdout.String() != tc.want || stderr.Len() != 0 {
				t.Fatalf("code = %d, stdout = %q, stderr = %q; want success with stdout %q", code, stdout.String(), stderr.String(), tc.want)
			}
		})
	}
}

func TestRunChangeNewRejectsPathsWithoutEscaping(t *testing.T) {
	cwd := t.TempDir()
	var stdout, stderr bytes.Buffer
	if code := Run([]string{"init"}, &stdout, &stderr, cwd); code != 0 {
		t.Fatal(stderr.String())
	}
	absolute := filepath.Join(t.TempDir(), "outside")
	for _, name := range []string{"../escape", absolute, "nested/name", `nested\name`} {
		stdout.Reset()
		stderr.Reset()
		code := Run([]string{"change", "new", name}, &stdout, &stderr, cwd)
		if code == 0 || stdout.Len() != 0 || stderr.Len() == 0 {
			t.Errorf("change new %q: code=%d stdout=%q stderr=%q", name, code, stdout.String(), stderr.String())
		}
	}
	for _, path := range []string{
		filepath.Join(cwd, "speclet", "escape"),
		absolute,
	} {
		if _, err := os.Lstat(path); !os.IsNotExist(err) {
			t.Errorf("unexpected path created outside changes: %s (err=%v)", path, err)
		}
	}
}

func TestRunTaskDoneAndUndone(t *testing.T) {
	cwd := t.TempDir()
	var stdout, stderr bytes.Buffer
	for _, args := range [][]string{{"init"}, {"change", "new", "tasks"}} {
		if code := Run(args, &stdout, &stderr, cwd); code != 0 {
			t.Fatalf("Run(%v) = %d, stderr = %q", args, code, stderr.String())
		}
		stdout.Reset()
		stderr.Reset()
	}
	tasksPath := filepath.Join(cwd, "speclet", "changes", "tasks", "tasks.md")
	original := []byte("- [ ] T001 first\n- [ ] T002 second\n")
	if err := os.WriteFile(tasksPath, original, 0o644); err != nil {
		t.Fatal(err)
	}
	for _, test := range []struct {
		verb string
		want string
	}{
		{verb: "done", want: "- [x] T002 second\n"},
		{verb: "undone", want: "- [ ] T002 second\n"},
	} {
		stdout.Reset()
		stderr.Reset()
		code := Run([]string{"task", test.verb, "tasks", "T002"}, &stdout, &stderr, cwd)
		if code != 0 || stderr.Len() != 0 || !strings.Contains(stdout.String(), "tasks") || !strings.Contains(stdout.String(), "T002") {
			t.Fatalf("task %s: code=%d stdout=%q stderr=%q", test.verb, code, stdout.String(), stderr.String())
		}
		got, err := os.ReadFile(tasksPath)
		if err != nil {
			t.Fatal(err)
		}
		if !strings.Contains(string(got), test.want) || !strings.Contains(string(got), "- [ ] T001 first\n") {
			t.Fatalf("task %s produced %q", test.verb, got)
		}
	}
}

func TestRunTaskRejectsMissingChangeAndBadArguments(t *testing.T) {
	var stdout, stderr bytes.Buffer
	for _, args := range [][]string{{"task", "done", "missing", "T001"}, {"task", "done"}, {"task", "other", "change", "T001"}} {
		stdout.Reset()
		stderr.Reset()
		code := Run(args, &stdout, &stderr, t.TempDir())
		if code == 0 || stdout.Len() != 0 || stderr.Len() == 0 {
			t.Errorf("Run(%v): code=%d stdout=%q stderr=%q", args, code, stdout.String(), stderr.String())
		}
	}
}

func TestRunArchiveAppliesAndMovesChange(t *testing.T) {
	cwd := t.TempDir()
	var stdout, stderr bytes.Buffer
	for _, args := range [][]string{{"init"}, {"change", "new", "archive-me"}} {
		if code := Run(args, &stdout, &stderr, cwd); code != 0 {
			t.Fatalf("Run(%v) = %d, stderr = %q", args, code, stderr.String())
		}
		stdout.Reset()
		stderr.Reset()
	}
	root := filepath.Join(cwd, "speclet")
	if err := os.WriteFile(filepath.Join(root, "changes", "archive-me", "tasks.md"), []byte("- [x] T001 Done\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	deltaPath := filepath.Join(root, "changes", "archive-me", "specs", "payments", "spec.md")
	if err := os.MkdirAll(filepath.Dir(deltaPath), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(deltaPath, []byte("## ADDED\n### Requirement: New\nText\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	code := Run([]string{"archive", "archive-me"}, &stdout, &stderr, cwd)
	if code != 0 || stdout.Len() == 0 || stderr.Len() != 0 {
		t.Fatalf("archive: code=%d stdout=%q stderr=%q", code, stdout.String(), stderr.String())
	}
	formal, err := os.ReadFile(filepath.Join(root, "specs", "payments", "spec.md"))
	if err != nil || !strings.Contains(string(formal), "## Requirement: New") {
		t.Fatalf("formal spec = %q, err=%v", formal, err)
	}
	archived := filepath.Join(root, "changes", "archive", time.Now().Format("2006-01-02")+"-archive-me")
	if _, err := os.Stat(archived); err != nil {
		t.Fatalf("archive directory missing: %v", err)
	}
}

func TestRunArchiveFailureUsesStderrAndKeepsChange(t *testing.T) {
	cwd := t.TempDir()
	var stdout, stderr bytes.Buffer
	for _, args := range [][]string{{"init"}, {"change", "new", "pending"}} {
		if code := Run(args, &stdout, &stderr, cwd); code != 0 {
			t.Fatal(stderr.String())
		}
		stdout.Reset()
		stderr.Reset()
	}
	root := filepath.Join(cwd, "speclet")
	if err := os.WriteFile(filepath.Join(root, "changes", "pending", "tasks.md"), []byte("- [ ] T001 Pending\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	code := Run([]string{"archive", "pending"}, &stdout, &stderr, cwd)
	if code == 0 || stdout.Len() != 0 || stderr.Len() == 0 {
		t.Fatalf("archive failure: code=%d stdout=%q stderr=%q", code, stdout.String(), stderr.String())
	}
	if _, err := os.Stat(filepath.Join(root, "changes", "pending")); err != nil {
		t.Fatalf("failed archive moved change: %v", err)
	}
}

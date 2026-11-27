package cli

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/chiisen/speclet/internal/version"
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

func TestRunHelpDocumentsJSONChangeList(t *testing.T) {
	var stdout, stderr bytes.Buffer
	if code := Run([]string{"help"}, &stdout, &stderr, t.TempDir()); code != 0 {
		t.Fatalf("help = %d, stderr = %q", code, stderr.String())
	}
	if !strings.Contains(stdout.String(), "--json") || stderr.Len() != 0 {
		t.Fatalf("help output = %q, stderr = %q; want --json on stdout", stdout.String(), stderr.String())
	}
}

func TestRunVersionAndHelpShowCurrentVersion(t *testing.T) {
	var versionOut, versionErr bytes.Buffer
	if code := Run([]string{"--version"}, &versionOut, &versionErr, t.TempDir()); code != 0 {
		t.Fatalf("version = %d, stderr = %q", code, versionErr.String())
	}
	line := strings.TrimSpace(versionOut.String())
	if line != "Speclet version: "+version.String() && line != "Speclet 版本："+version.String() || versionErr.Len() != 0 {
		t.Fatalf("version output = %q, stderr = %q", versionOut.String(), versionErr.String())
	}

	for _, flag := range []string{"-h", "--help"} {
		var stdout, stderr bytes.Buffer
		if code := Run([]string{flag}, &stdout, &stderr, t.TempDir()); code != 0 {
			t.Fatalf("%s = %d, stderr = %q", flag, code, stderr.String())
		}
		if !strings.Contains(stdout.String(), line) || !strings.Contains(stdout.String(), "speclet init") || stderr.Len() != 0 {
			t.Errorf("%s output = %q, want version %q and existing usage", flag, stdout.String(), line)
		}
	}
}

func TestRunVersionTextUsesTraditionalChinese(t *testing.T) {
	configDir := t.TempDir()
	t.Setenv("HOME", configDir)
	t.Setenv("APPDATA", configDir)
	t.Setenv("XDG_CONFIG_HOME", configDir)
	userConfigDir, err := os.UserConfigDir()
	if err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Join(userConfigDir, "speclet"), 0o700); err != nil {
		t.Fatal(err)
	}
	settings := filepath.Join(userConfigDir, "speclet", "config.json")
	if err := os.WriteFile(settings, []byte(`{"language":"zh-TW"}`), 0o600); err != nil {
		t.Fatal(err)
	}

	var stdout, stderr bytes.Buffer
	if code := Run([]string{"--version"}, &stdout, &stderr, t.TempDir()); code != 0 {
		t.Fatalf("version = %d, stderr = %q", code, stderr.String())
	}
	if !strings.HasPrefix(stdout.String(), "Speclet 版本：") {
		t.Fatalf("version output = %q, want Traditional Chinese label", stdout.String())
	}

	stdout.Reset()
	stderr.Reset()
	if code := Run([]string{"--version", "extra"}, &stdout, &stderr, t.TempDir()); code == 0 {
		t.Fatal("version with extra arguments succeeded")
	}
	if !strings.Contains(stderr.String(), "用法：speclet --version") {
		t.Fatalf("usage error = %q, want Traditional Chinese usage", stderr.String())
	}

	stdout.Reset()
	stderr.Reset()
	if code := Run([]string{"-h"}, &stdout, &stderr, t.TempDir()); code != 0 {
		t.Fatalf("help = %d, stderr = %q", code, stderr.String())
	}
	if !strings.Contains(stdout.String(), "Speclet 版本：") || !strings.Contains(stdout.String(), "用法：") {
		t.Fatalf("help output = %q, want Traditional Chinese version and usage", stdout.String())
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

func TestRunChangeListJSON(t *testing.T) {
	type summary struct {
		Name  string `json:"name"`
		Done  int    `json:"done"`
		Total int    `json:"total"`
	}
	for _, tc := range []struct {
		name      string
		changes   map[string]string
		args      []string
		want      []summary
		wantError bool
		listError bool
	}{
		{name: "empty list", args: []string{"--json"}, want: []summary{}},
		{
			name: "names with spaces and unicode use task counter",
			args: []string{"--json"},
			changes: map[string]string{
				"payment checkout": "- [x] T001 done\n- [ ] T002 todo\n- [X] T003 also done\n",
				"付款 甲":             "- [ ] T001 待辦\n",
			},
			want: []summary{
				{Name: "payment checkout", Done: 2, Total: 3},
				{Name: "付款 甲", Done: 0, Total: 1},
			},
		},
		{name: "conflicts with names only", args: []string{"--json", "--names-only"}, wantError: true},
		{name: "unsupported argument", args: []string{"--unknown"}, wantError: true},
		{name: "invalid task file is an error", args: []string{"--json"}, changes: map[string]string{"broken": "  - [ ] T001 nested\n"}, listError: true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			cwd := t.TempDir()
			var stdout, stderr bytes.Buffer
			if code := Run([]string{"init"}, &stdout, &stderr, cwd); code != 0 {
				t.Fatalf("init = %d, stderr = %q", code, stderr.String())
			}
			for name, tasks := range tc.changes {
				if code := Run([]string{"change", "new", name}, &stdout, &stderr, cwd); code != 0 {
					t.Fatalf("change new %q = %d, stderr = %q", name, code, stderr.String())
				}
				if err := os.WriteFile(filepath.Join(cwd, "speclet", "changes", name, "tasks.md"), []byte(tasks), 0o644); err != nil {
					t.Fatal(err)
				}
			}
			stdout.Reset()
			stderr.Reset()
			code := Run(append([]string{"change", "list"}, tc.args...), &stdout, &stderr, cwd)
			if tc.listError {
				if code == 0 || stdout.Len() != 0 || stderr.Len() == 0 {
					t.Fatalf("code = %d, stdout = %q, stderr = %q; want task error on stderr and no stdout", code, stdout.String(), stderr.String())
				}
				return
			}
			if tc.wantError {
				if code == 0 || stdout.Len() != 0 || !strings.Contains(stderr.String(), "usage: speclet change list") {
					t.Fatalf("code = %d, stdout = %q, stderr = %q; want usage on stderr and no stdout", code, stdout.String(), stderr.String())
				}
				return
			}
			if code != 0 || stderr.Len() != 0 {
				t.Fatalf("code = %d, stdout = %q, stderr = %q", code, stdout.String(), stderr.String())
			}
			if len(tc.want) == 0 && stdout.String() != "[]\n" {
				t.Fatalf("empty JSON output = %q, want [] followed by newline", stdout.String())
			}
			var got []summary
			if err := json.Unmarshal(stdout.Bytes(), &got); err != nil {
				t.Fatalf("stdout is not JSON: %q (%v)", stdout.String(), err)
			}
			if len(tc.want) > 0 {
				for _, key := range []string{`"name":`, `"done":`, `"total":`} {
					if !strings.Contains(stdout.String(), key) {
						t.Errorf("JSON output %q is missing lower-camel-case field %s", stdout.String(), key)
					}
				}
				if strings.Contains(stdout.String(), `"Name":`) || strings.Contains(stdout.String(), `"Done":`) || strings.Contains(stdout.String(), `"Total":`) {
					t.Errorf("JSON output uses non-lower-camel-case keys: %q", stdout.String())
				}
			}
			if got == nil || len(got) != len(tc.want) {
				t.Fatalf("JSON summaries = %#v, want %#v", got, tc.want)
			}
			for i := range got {
				if got[i] != tc.want[i] {
					t.Errorf("summary[%d] = %#v, want %#v", i, got[i], tc.want[i])
				}
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

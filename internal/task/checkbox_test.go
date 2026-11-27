package task

import (
	"os"
	"path/filepath"
	"testing"
)

func TestCountFlatCheckboxes(t *testing.T) {
	src := []byte("# Tasks\n\n- [ ] T001 First\n- [x] T002 Second\n- [X] T003 Third\n")
	done, total, err := Count(src)
	if err != nil {
		t.Fatal(err)
	}
	if done != 2 || total != 3 {
		t.Fatalf("Count() = (%d, %d), want (2, 3)", done, total)
	}
}

func TestCountRejectsNestedCheckboxOutsideCodeFence(t *testing.T) {
	src := []byte("- [ ] T001 Parent\n  - [ ] T002 Child\n")
	if _, _, err := Count(src); err == nil {
		t.Fatal("Count() error = nil, want nested checkbox error")
	}
}

func TestCountIgnoresCheckboxesInCodeFences(t *testing.T) {
	src := []byte("- [ ] T001 Real\n```markdown\n- [x] T002 Example\n```\n")
	done, total, err := Count(src)
	if err != nil {
		t.Fatal(err)
	}
	if done != 0 || total != 1 {
		t.Fatalf("Count() = (%d, %d), want (0, 1)", done, total)
	}
}

func TestCountRejectsMissingAndDuplicateTaskIDs(t *testing.T) {
	for _, src := range []string{
		"- [x]\n",
		"- [x] T001 first\n- [ ] T001 duplicate\n",
	} {
		if _, _, err := Count([]byte(src)); err == nil {
			t.Errorf("Count(%q) error = nil, want invalid task ID error", src)
		}
	}
}

func TestCountAcceptsStableIDsWithoutPrescribingTheirSuffix(t *testing.T) {
	done, total, err := Count([]byte("- [x] review-docs Update guidance\n"))
	if err != nil || done != 1 || total != 1 {
		t.Fatalf("Count() = (%d, %d, %v), want (1, 1, nil)", done, total, err)
	}
}

func TestCountDoesNotTreatInlineCodeAsAFenceAndRejectsUnclosedFence(t *testing.T) {
	src := []byte("``` - [ ] inline example ```\n- [ ] T001 real\n")
	_, total, err := Count(src)
	if err != nil || total != 1 {
		t.Fatalf("Count() = (_, %d, %v), want (_, 1, nil)", total, err)
	}
	if _, _, err := Count([]byte("```markdown\n- [x] T001 hidden\n")); err == nil {
		t.Fatal("Count() error = nil for unclosed code fence")
	}
}

func TestSetDoneAndUndonePreservesOtherBytesAndCRLF(t *testing.T) {
	path := filepath.Join(t.TempDir(), "tasks.md")
	original := []byte("# Tasks\r\n\r\n- [ ] T001 First task  \r\n- [x] T002 Second task\r\n")
	if err := os.WriteFile(path, original, 0o640); err != nil {
		t.Fatal(err)
	}
	if err := SetDone(path, "T001", true); err != nil {
		t.Fatal(err)
	}
	got, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	want := []byte("# Tasks\r\n\r\n- [x] T001 First task  \r\n- [x] T002 Second task\r\n")
	if string(got) != string(want) {
		t.Fatalf("SetDone() bytes = %q, want %q", got, want)
	}
	if err := SetDone(path, "T001", false); err != nil {
		t.Fatal(err)
	}
	got, err = os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != string(original) {
		t.Fatalf("SetDone(false) bytes = %q, want %q", got, original)
	}
}

func TestSetDoneRejectsUnknownDuplicateAndNestedIDsWithoutWriting(t *testing.T) {
	for _, test := range []struct {
		name string
		src  string
		id   string
	}{
		{name: "unknown", src: "- [ ] T001 Existing\n", id: "T002"},
		{name: "duplicate", src: "- [ ] T001 First\n- [x] T001 Again\n", id: "T001"},
		{name: "nested", src: "- [ ] T001 Parent\n  - [ ] T002 Child\n", id: "T001"},
	} {
		t.Run(test.name, func(t *testing.T) {
			path := filepath.Join(t.TempDir(), "tasks.md")
			original := []byte(test.src)
			if err := os.WriteFile(path, original, 0o644); err != nil {
				t.Fatal(err)
			}
			if err := SetDone(path, test.id, true); err == nil {
				t.Fatal("SetDone() error = nil, want validation error")
			}
			got, err := os.ReadFile(path)
			if err != nil {
				t.Fatal(err)
			}
			if string(got) != string(original) {
				t.Fatalf("SetDone() modified file: got %q, want %q", got, original)
			}
		})
	}
}

func TestSetDoneIgnoresBlankLinesAndCodeFences(t *testing.T) {
	path := filepath.Join(t.TempDir(), "tasks.md")
	original := []byte("\n- [ ] T001 Real\n\n```markdown\n- [ ] T001 Example\n```\n")
	if err := os.WriteFile(path, original, 0o644); err != nil {
		t.Fatal(err)
	}
	if err := SetDone(path, "T001", true); err != nil {
		t.Fatal(err)
	}
	got, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	want := []byte("\n- [x] T001 Real\n\n```markdown\n- [ ] T001 Example\n```\n")
	if string(got) != string(want) {
		t.Fatalf("SetDone() bytes = %q, want %q", got, want)
	}
}

func TestCountAndSetDoneHandleUTF8BOM(t *testing.T) {
	original := []byte("\xef\xbb\xbf- [ ] T001 pending\r\n")
	done, total, err := Count(original)
	if err != nil || done != 0 || total != 1 {
		t.Fatalf("Count() = (%d, %d, %v), want (0, 1, nil)", done, total, err)
	}
	path := filepath.Join(t.TempDir(), "tasks.md")
	if err := os.WriteFile(path, original, 0o644); err != nil {
		t.Fatal(err)
	}
	if err := SetDone(path, "T001", true); err != nil {
		t.Fatal(err)
	}
	got, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	want := []byte("\xef\xbb\xbf- [x] T001 pending\r\n")
	if string(got) != string(want) {
		t.Fatalf("SetDone() = %q, want %q", got, want)
	}
}

func TestCountAndSetDoneRejectInvalidUTF8(t *testing.T) {
	invalid := []byte{'-', ' ', '[', 'x', ']', ' ', 'T', '0', '0', '1', ' ', 0xff, '\n'}
	if _, _, err := Count(invalid); err == nil {
		t.Fatal("Count() accepted invalid UTF-8")
	}
	path := filepath.Join(t.TempDir(), "tasks.md")
	if err := os.WriteFile(path, invalid, 0o644); err != nil {
		t.Fatal(err)
	}
	if err := SetDone(path, "T001", true); err == nil {
		t.Fatal("SetDone() accepted invalid UTF-8")
	}
}

func TestSetDoneRejectsDuplicateOrMissingOtherTaskIDs(t *testing.T) {
	for _, src := range []string{
		"- [ ] T001 selected\n- [ ] T002\n- [x] T002 duplicate\n",
		"- [ ] T001 selected\n- [ ]\n",
	} {
		path := filepath.Join(t.TempDir(), "tasks.md")
		if err := os.WriteFile(path, []byte(src), 0o644); err != nil {
			t.Fatal(err)
		}
		if err := SetDone(path, "T001", true); err == nil {
			t.Errorf("SetDone() error = nil for malformed task list %q", src)
		}
		got, err := os.ReadFile(path)
		if err != nil {
			t.Fatal(err)
		}
		if string(got) != src {
			t.Errorf("SetDone() modified invalid task list: got %q, want %q", got, src)
		}
	}
}

func TestSetDoneRejectsTasksFileUnderSymlinkedParent(t *testing.T) {
	parent := t.TempDir()
	targetDir := t.TempDir()
	tasks := filepath.Join(targetDir, "tasks.md")
	const original = "- [ ] T001 outside\n"
	if err := os.WriteFile(tasks, []byte(original), 0o644); err != nil {
		t.Fatal(err)
	}
	link := filepath.Join(parent, "linked")
	if err := os.Symlink(targetDir, link); err != nil {
		t.Skipf("symlinks unavailable: %v", err)
	}
	if err := SetDone(filepath.Join(link, "tasks.md"), "T001", true); err == nil {
		t.Fatal("SetDone() error = nil for symlinked parent directory")
	}
	got, err := os.ReadFile(tasks)
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != original {
		t.Fatalf("external tasks file was modified: %q", got)
	}
}

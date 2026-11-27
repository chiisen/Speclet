package change

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func newSpecletRoot(t *testing.T) string {
	t.Helper()
	root := filepath.Join(t.TempDir(), "speclet")
	if err := os.MkdirAll(filepath.Join(root, "changes", "archive"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Join(root, "specs"), 0o755); err != nil {
		t.Fatal(err)
	}
	return root
}

func TestCreateChangeAndRejectDuplicate(t *testing.T) {
	root := newSpecletRoot(t)
	if err := Create(root, "feature change"); err != nil {
		t.Fatal(err)
	}
	for _, name := range []string{"proposal.md", "tasks.md", "specs"} {
		if _, err := os.Stat(filepath.Join(root, "changes", "feature change", name)); err != nil {
			t.Errorf("expected %s: %v", name, err)
		}
	}
	if err := Create(root, "feature change"); err == nil {
		t.Fatal("Create() error = nil for duplicate name")
	}
}

func TestCreateRejectsUnsafeNames(t *testing.T) {
	root := newSpecletRoot(t)
	for _, name := range []string{"", ".", "..", "../escape", "/tmp/escape", `nested\\name`, "two..dots"} {
		if err := Create(root, name); err == nil {
			t.Errorf("Create(%q) error = nil, want unsafe name error", name)
		}
	}
	if _, err := os.Stat(filepath.Join(filepath.Dir(root), "escape")); !os.IsNotExist(err) {
		t.Fatalf("unexpected escaped path state: %v", err)
	}
}

func TestCreatePreservesUnicodeAndSpaces(t *testing.T) {
	root := newSpecletRoot(t)
	name := "新增功能 甲"
	if err := Create(root, name); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(filepath.Join(root, "changes", name, "proposal.md")); err != nil {
		t.Fatal(err)
	}
}

func TestListEmptyAndCountsTasks(t *testing.T) {
	root := newSpecletRoot(t)
	summaries, err := List(root)
	if err != nil {
		t.Fatal(err)
	}
	if len(summaries) != 0 {
		t.Fatalf("List() returned %d entries, want empty", len(summaries))
	}
	if err := Create(root, "alpha"); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, "changes", "alpha", "tasks.md"), []byte("- [x] T001 done\n- [ ] T002 todo\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	summaries, err = List(root)
	if err != nil {
		t.Fatal(err)
	}
	if len(summaries) != 1 || summaries[0] != (Summary{Name: "alpha", Done: 1, Total: 2}) {
		t.Fatalf("List() = %#v, want one alpha summary with 1/2 tasks", summaries)
	}
}

func TestListIncludesDotPrefixedChangeNames(t *testing.T) {
	root := newSpecletRoot(t)
	if err := Create(root, ".draft"); err != nil {
		t.Fatal(err)
	}
	summaries, err := List(root)
	if err != nil {
		t.Fatal(err)
	}
	if len(summaries) != 1 || summaries[0].Name != ".draft" {
		t.Fatalf("List() = %#v, want .draft", summaries)
	}
}

func TestListRejectsMalformedNestedTasks(t *testing.T) {
	root := newSpecletRoot(t)
	if err := Create(root, "alpha"); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, "changes", "alpha", "tasks.md"), []byte("- [ ] T001 parent\n  - [ ] T002 child\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if _, err := List(root); err == nil || !strings.Contains(err.Error(), "nested") {
		t.Fatalf("List() error = %v, want nested checkbox error", err)
	}
}

func TestReadReturnsPathsAndRejectsMissingChange(t *testing.T) {
	root := newSpecletRoot(t)
	if _, err := Read(root, "missing"); err == nil {
		t.Fatal("Read() error = nil for missing change")
	}
	if err := Create(root, "feature"); err != nil {
		t.Fatal(err)
	}
	capability := filepath.Join(root, "changes", "feature", "specs", "payments")
	if err := os.MkdirAll(capability, 0o755); err != nil {
		t.Fatal(err)
	}
	delta := filepath.Join(capability, "spec.md")
	if err := os.WriteFile(delta, []byte(""), 0o644); err != nil {
		t.Fatal(err)
	}
	detail, err := Read(root, "feature")
	if err != nil {
		t.Fatal(err)
	}
	if detail.Name != "feature" || detail.ProposalPath != filepath.Join(root, "changes", "feature", "proposal.md") || detail.TasksPath != filepath.Join(root, "changes", "feature", "tasks.md") {
		t.Fatalf("Read() = %#v, unexpected paths or name", detail)
	}
	if len(detail.DeltaPaths) != 1 || detail.DeltaPaths[0] != delta {
		t.Fatalf("DeltaPaths = %#v, want [%q]", detail.DeltaPaths, delta)
	}
}

func TestReadRejectsMissingFiles(t *testing.T) {
	root := newSpecletRoot(t)
	if err := os.Mkdir(filepath.Join(root, "changes", "broken"), 0o755); err != nil {
		t.Fatal(err)
	}
	if _, err := Read(root, "broken"); err == nil {
		t.Fatal("Read() error = nil for change missing proposal/tasks")
	}
}

func TestReadRejectsSymlinkedChangesDirectory(t *testing.T) {
	externalRoot := newSpecletRoot(t)
	if err := Create(externalRoot, "external"); err != nil {
		t.Fatal(err)
	}
	root := newSpecletRoot(t)
	changesPath := filepath.Join(root, "changes")
	if err := os.RemoveAll(changesPath); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(filepath.Join(externalRoot, "changes"), changesPath); err != nil {
		t.Skipf("symlinks unavailable: %v", err)
	}
	if _, err := Read(root, "external"); err == nil {
		t.Fatal("Read() accepted change data through symlinked changes directory")
	}
}

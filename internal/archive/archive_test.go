package archive

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/chiisen/speclet/internal/change"
	"github.com/chiisen/speclet/internal/project"
)

func archiveFixture(t *testing.T, name string) string {
	t.Helper()
	base := t.TempDir()
	if err := project.Init(base); err != nil {
		t.Fatal(err)
	}
	root := filepath.Join(base, "speclet")
	if err := change.Create(root, name); err != nil {
		t.Fatal(err)
	}
	return root
}

func writeChangeFile(t *testing.T, root, name, rel, contents string) string {
	t.Helper()
	path := filepath.Join(root, "changes", name, rel)
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(contents), 0o644); err != nil {
		t.Fatal(err)
	}
	return path
}

func TestRunAppliesAllDeltaKindsAndArchivesChange(t *testing.T) {
	root := archiveFixture(t, "release")
	formalDir := filepath.Join(root, "specs", "payments")
	if err := os.MkdirAll(formalDir, 0o755); err != nil {
		t.Fatal(err)
	}
	formal := filepath.Join(formalDir, "spec.md")
	original := "# Payment rules\n\nKeep preface.\n\n## Requirement: Existing\nOld behavior.\n\n## Requirement: Retired\nRemove this.\n\n## Notes\nKeep these notes.\n"
	if err := os.WriteFile(formal, []byte(original), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, "changes", "release", "tasks.md"), []byte("- [x] T001 Implement\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	writeChangeFile(t, root, "release", filepath.Join("specs", "payments", "spec.md"), "## MODIFIED\n### Requirement: Existing\nNew behavior.\n#### Scenario: New\n- **THEN** works\n## REMOVED\n### Requirement: Retired\n")
	writeChangeFile(t, root, "release", filepath.Join("specs", "refunds", "spec.md"), "## ADDED\n### Requirement: Refund\nRefund behavior.\n#### Scenario: Requested\n- **THEN** accepted\n")

	if err := Run(root, "release"); err != nil {
		t.Fatal(err)
	}
	updated, err := os.ReadFile(formal)
	if err != nil {
		t.Fatal(err)
	}
	text := string(updated)
	for _, expected := range []string{"Keep preface.", "## Requirement: Existing\nNew behavior.", "### Scenario: New", "## Notes\nKeep these notes."} {
		if !strings.Contains(text, expected) {
			t.Errorf("updated spec missing %q: %q", expected, text)
		}
	}
	if strings.Contains(text, "Old behavior") || strings.Contains(text, "Remove this") || strings.Contains(text, "#### Scenario") {
		t.Fatalf("updated spec contains old or unnormalized content: %q", text)
	}
	refundSpec, err := os.ReadFile(filepath.Join(root, "specs", "refunds", "spec.md"))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(refundSpec), "## Requirement: Refund") || !strings.Contains(string(refundSpec), "### Scenario: Requested") {
		t.Fatalf("new capability spec missing added requirement: %q", refundSpec)
	}
	archivePath := filepath.Join(root, "changes", "archive", time.Now().Format("2006-01-02")+"-release")
	if _, err := os.Stat(archivePath); err != nil {
		t.Fatalf("archived change missing: %v", err)
	}
	if _, err := os.Stat(filepath.Join(root, "changes", "release")); !os.IsNotExist(err) {
		t.Fatalf("active change still exists, stat error = %v", err)
	}
	if _, err := os.Stat(filepath.Join(archivePath, "proposal.md")); err != nil {
		t.Fatalf("archive omitted proposal: %v", err)
	}
}

func TestRunTreatsZeroTasksAsComplete(t *testing.T) {
	root := archiveFixture(t, "empty")
	if err := Run(root, "empty"); err != nil {
		t.Fatalf("Run() with empty task list: %v", err)
	}
	archivePath := filepath.Join(root, "changes", "archive", time.Now().Format("2006-01-02")+"-empty")
	if _, err := os.Stat(archivePath); err != nil {
		t.Fatalf("archived change missing: %v", err)
	}
}

func TestRunDoesNotArchiveBOMTaskFileWithPendingTask(t *testing.T) {
	root := archiveFixture(t, "bom-pending")
	tasks := filepath.Join(root, "changes", "bom-pending", "tasks.md")
	if err := os.WriteFile(tasks, []byte("\xef\xbb\xbf- [ ] T001 pending\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	writeChangeFile(t, root, "bom-pending", filepath.Join("specs", "payments", "spec.md"), "## ADDED\n### Requirement: New\nText.\n")
	if err := Run(root, "bom-pending"); err == nil {
		t.Fatal("Run() archived a change with a BOM-prefixed pending task")
	}
	if _, err := os.Stat(filepath.Join(root, "changes", "bom-pending")); err != nil {
		t.Fatalf("pending change was moved: %v", err)
	}
	if _, err := os.Stat(filepath.Join(root, "specs", "payments", "spec.md")); !os.IsNotExist(err) {
		t.Fatalf("formal spec was written for pending task: err=%v", err)
	}
}

func TestRunRejectsInvalidInputsWithoutChangingSpecsOrChange(t *testing.T) {
	for _, test := range []struct {
		name       string
		tasks      string
		delta      string
		formal     string
		makeTarget func(t *testing.T, root string)
	}{
		{name: "incomplete task", tasks: "- [ ] T001 pending\n", delta: "## ADDED\n### Requirement: New\nText\n", formal: "# Rules\n"},
		{name: "invalid delta", tasks: "- [x] T001 done\n", delta: "## UNKNOWN\n### Requirement: New\nText\n", formal: "# Rules\n"},
		{name: "missing modified target", tasks: "- [x] T001 done\n", delta: "## MODIFIED\n### Requirement: Missing\nText\n", formal: "# Rules\n"},
		{name: "duplicate formal target", tasks: "- [x] T001 done\n", delta: "## MODIFIED\n### Requirement: Existing\nNew\n", formal: "## Requirement: Existing\nOne\n## Requirement: Existing\nTwo\n"},
		{name: "spec target is directory", tasks: "- [x] T001 done\n", delta: "## ADDED\n### Requirement: New\nText\n", makeTarget: func(t *testing.T, root string) {
			t.Helper()
			if err := os.MkdirAll(filepath.Join(root, "specs", "payments", "spec.md"), 0o755); err != nil {
				t.Fatal(err)
			}
		}},
		{name: "archive target exists", tasks: "- [x] T001 done\n", delta: "## ADDED\n### Requirement: New\nText\n", formal: "# Rules\n", makeTarget: func(t *testing.T, root string) {
			t.Helper()
			if err := os.Mkdir(filepath.Join(root, "changes", "archive", time.Now().Format("2006-01-02")+"-guard"), 0o755); err != nil {
				t.Fatal(err)
			}
		}},
	} {
		t.Run(test.name, func(t *testing.T) {
			root := archiveFixture(t, "guard")
			changeDir := filepath.Join(root, "changes", "guard")
			if err := os.WriteFile(filepath.Join(changeDir, "tasks.md"), []byte(test.tasks), 0o644); err != nil {
				t.Fatal(err)
			}
			deltaPath := writeChangeFile(t, root, "guard", filepath.Join("specs", "payments", "spec.md"), test.delta)
			var formalPath string
			if test.formal != "" {
				formalPath = filepath.Join(root, "specs", "payments", "spec.md")
				if err := os.MkdirAll(filepath.Dir(formalPath), 0o755); err != nil {
					t.Fatal(err)
				}
				if err := os.WriteFile(formalPath, []byte(test.formal), 0o644); err != nil {
					t.Fatal(err)
				}
			}
			if test.makeTarget != nil {
				test.makeTarget(t, root)
			}
			var originalFormal []byte
			if formalPath != "" {
				originalFormal, _ = os.ReadFile(formalPath)
			}
			originalDelta, err := os.ReadFile(deltaPath)
			if err != nil {
				t.Fatal(err)
			}
			if err := Run(root, "guard"); err == nil {
				t.Fatal("Run() error = nil, want archive validation error")
			}
			if _, err := os.Stat(changeDir); err != nil {
				t.Fatalf("active change was moved: %v", err)
			}
			gotDelta, err := os.ReadFile(deltaPath)
			if err != nil || string(gotDelta) != string(originalDelta) {
				t.Fatalf("delta changed: %q, err=%v", gotDelta, err)
			}
			if formalPath != "" {
				gotFormal, err := os.ReadFile(formalPath)
				if err != nil || string(gotFormal) != string(originalFormal) {
					t.Fatalf("formal spec changed: %q, err=%v", gotFormal, err)
				}
			}
		})
	}
}

func TestRunPrevalidatesEveryDeltaBeforeUpdatingAnySpec(t *testing.T) {
	root := archiveFixture(t, "multi")
	formalPath := filepath.Join(root, "specs", "payments", "spec.md")
	if err := os.MkdirAll(filepath.Dir(formalPath), 0o755); err != nil {
		t.Fatal(err)
	}
	const original = "## Requirement: Existing\nOriginal.\n"
	if err := os.WriteFile(formalPath, []byte(original), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, "changes", "multi", "tasks.md"), []byte("- [x] T001 done\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	writeChangeFile(t, root, "multi", filepath.Join("specs", "payments", "spec.md"), "## MODIFIED\n### Requirement: Existing\nNew.\n")
	writeChangeFile(t, root, "multi", filepath.Join("specs", "refunds", "spec.md"), "## UNKNOWN\n### Requirement: New\nInvalid.\n")
	if err := Run(root, "multi"); err == nil {
		t.Fatal("Run() error = nil, want second delta validation error")
	}
	got, err := os.ReadFile(formalPath)
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != original {
		t.Fatalf("first formal spec changed before all deltas validated: %q", got)
	}
}

func TestRunRejectsAnExistingWriteLockWithoutChangingFiles(t *testing.T) {
	root := archiveFixture(t, "locked")
	formalPath := filepath.Join(root, "specs", "payments", "spec.md")
	if err := os.MkdirAll(filepath.Dir(formalPath), 0o755); err != nil {
		t.Fatal(err)
	}
	const original = "## Requirement: Existing\nOriginal.\n"
	if err := os.WriteFile(formalPath, []byte(original), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, "changes", "locked", "tasks.md"), []byte("- [x] T001 done\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	writeChangeFile(t, root, "locked", filepath.Join("specs", "payments", "spec.md"), "## MODIFIED\n### Requirement: Existing\nNew.\n")
	if err := os.Mkdir(filepath.Join(root, ".operation.lock"), 0o700); err != nil {
		t.Fatal(err)
	}
	if err := Run(root, "locked"); err == nil {
		t.Fatal("Run() error = nil while the project write lock is held")
	}
	got, err := os.ReadFile(formalPath)
	if err != nil || string(got) != original {
		t.Fatalf("formal spec changed while lock was held: %q, err=%v", got, err)
	}
	if _, err := os.Stat(filepath.Join(root, "changes", "locked")); err != nil {
		t.Fatalf("active change was moved while lock was held: %v", err)
	}
}

func TestRunRollsBackEarlierReplacementWhenLaterReplacementFails(t *testing.T) {
	root := archiveFixture(t, "rollback")
	formalPath := filepath.Join(root, "specs", "payments", "spec.md")
	if err := os.MkdirAll(filepath.Dir(formalPath), 0o755); err != nil {
		t.Fatal(err)
	}
	const original = "## Requirement: Existing\nOriginal.\n"
	if err := os.WriteFile(formalPath, []byte(original), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, "changes", "rollback", "tasks.md"), []byte("- [x] T001 done\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	writeChangeFile(t, root, "rollback", filepath.Join("specs", "payments", "spec.md"), "## MODIFIED\n### Requirement: Existing\nUpdated.\n")
	writeChangeFile(t, root, "rollback", filepath.Join("specs", "refunds", "spec.md"), "## ADDED\n### Requirement: Refund\nNew.\n")
	originalReplace := replaceStagedFile
	replaceStagedFile = func(old, new string) error {
		if new == filepath.Join(root, "specs", "refunds", "spec.md") {
			return errors.New("injected replacement failure")
		}
		return originalReplace(old, new)
	}
	defer func() { replaceStagedFile = originalReplace }()
	if err := Run(root, "rollback"); err == nil {
		t.Fatal("Run() error = nil after injected replacement failure")
	}
	got, err := os.ReadFile(formalPath)
	if err != nil || string(got) != original {
		t.Fatalf("first spec was not restored: %q, err=%v", got, err)
	}
	if _, err := os.Stat(filepath.Join(root, "specs", "refunds", "spec.md")); !os.IsNotExist(err) {
		t.Fatalf("second spec unexpectedly remains: err=%v", err)
	}
	if _, err := os.Stat(filepath.Join(root, "changes", "rollback")); err != nil {
		t.Fatalf("active change was moved despite replacement failure: %v", err)
	}
	entries, err := os.ReadDir(filepath.Join(root, "specs", "payments"))
	if err != nil {
		t.Fatal(err)
	}
	if len(entries) != 1 || entries[0].Name() != "spec.md" {
		t.Fatalf("temporary files remain after rollback: %#v", entries)
	}
}

func TestRunRollsBackSpecsWhenFinalArchiveMoveFails(t *testing.T) {
	root := archiveFixture(t, "move-failure")
	formalPath := filepath.Join(root, "specs", "payments", "spec.md")
	if err := os.MkdirAll(filepath.Dir(formalPath), 0o755); err != nil {
		t.Fatal(err)
	}
	const original = "## Requirement: Existing\nOriginal.\n"
	if err := os.WriteFile(formalPath, []byte(original), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, "changes", "move-failure", "tasks.md"), []byte("- [x] T001 done\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	writeChangeFile(t, root, "move-failure", filepath.Join("specs", "payments", "spec.md"), "## MODIFIED\n### Requirement: Existing\nUpdated.\n")
	originalMove := moveChangeDirectory
	moveChangeDirectory = func(old, new string) error { return errors.New("injected archive move failure") }
	defer func() { moveChangeDirectory = originalMove }()
	if err := Run(root, "move-failure"); err == nil {
		t.Fatal("Run() error = nil after injected archive move failure")
	}
	got, err := os.ReadFile(formalPath)
	if err != nil || string(got) != original {
		t.Fatalf("formal spec was not restored: %q, err=%v", got, err)
	}
	if _, err := os.Stat(filepath.Join(root, "changes", "move-failure")); err != nil {
		t.Fatalf("active change was moved despite archive move failure: %v", err)
	}
}

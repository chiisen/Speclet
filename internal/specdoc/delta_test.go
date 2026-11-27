package specdoc

import (
	"strings"
	"testing"
)

func TestParseDeltaSectionsRequirementsAndScenarios(t *testing.T) {
	src := []byte("## ADDED\n### Requirement: New capability\nText.\n#### Scenario: Works\n- **WHEN** called\n- **THEN** works\n## MODIFIED\n### Requirement: Existing\nUpdated.\n")
	ops, err := ParseDelta(src)
	if err != nil {
		t.Fatal(err)
	}
	if len(ops) != 2 || ops[0].Kind != Added || ops[0].Name != "New capability" || ops[1].Kind != Modified || ops[1].Name != "Existing" {
		t.Fatalf("ParseDelta() = %#v", ops)
	}
	if !strings.Contains(string(ops[0].Raw), "#### Scenario: Works") {
		t.Fatalf("delta requirement lost scenario: %q", ops[0].Raw)
	}
}

func TestParseDeltaRejectsUnknownSectionsAndInvalidNames(t *testing.T) {
	for _, src := range []string{
		"## UNKNOWN\n### Requirement: R\n",
		"## ADDED\n### Requirement:\n",
		"## ADDED\n#### Scenario: Orphan\n",
		"## ADDED\n### Requirement: R\n## ADDED\n### Requirement: R\n",
		"### Requirement: No section\n",
		"## REMOVED\n### Requirement: B\n# Notes\n### Requirement: A\n",
		"## REMOVED\n### Requirement: B\nAppendix\n=======\n### Requirement: A\n",
		"## REMOVED\n\nAppendix\n=======\n\n### Requirement: A\n",
	} {
		if _, err := ParseDelta([]byte(src)); err == nil {
			t.Errorf("ParseDelta(%q) error = nil, want validation error", src)
		}
	}
}

func TestParseDeltaSetextAppendixEndsOperationAndIsNotRequirementBody(t *testing.T) {
	delta := []byte("## MODIFIED\n### Requirement: B\nnew\n\nAppendix\n=======\nnot part of B\n")
	ops, err := ParseDelta(delta)
	if err != nil {
		t.Fatal(err)
	}
	if len(ops) != 1 || ops[0].Name != "B" {
		t.Fatalf("ParseDelta() = %#v, want one Modified B operation", ops)
	}
	want := "### Requirement: B\nnew\n\n"
	if string(ops[0].Raw) != want {
		t.Fatalf("operation Raw = %q, want exact bytes %q", ops[0].Raw, want)
	}
}

func TestApplyDeltaRemoveThenAddPreservesBOMWithoutExtraBlankLines(t *testing.T) {
	spec := []byte("\uFEFF## Requirement: B\nOld.\n")
	delta := []byte("## REMOVED\n### Requirement: B\n## ADDED\n### Requirement: A\nNew.\n")
	got, err := ApplyDelta(spec, delta)
	if err != nil {
		t.Fatal(err)
	}
	want := "\uFEFF## Requirement: A\nNew.\n"
	if string(got) != want {
		t.Fatalf("ApplyDelta() = %q, want %q", got, want)
	}
}

func TestParseDeltaIgnoresStructuralHeadingsInsideFences(t *testing.T) {
	src := []byte("## ADDED\n### Requirement: New\nExample follows.\n```markdown\n## MODIFIED\n### Requirement: Fake\n#### Scenario: Fake\n```\nActual text.\n")
	ops, err := ParseDelta(src)
	if err != nil {
		t.Fatal(err)
	}
	if len(ops) != 1 || ops[0].Name != "New" || ops[0].Kind != Added {
		t.Fatalf("ParseDelta() = %#v, want one Added New operation", ops)
	}
	got, err := ApplyDelta(nil, src)
	if err != nil {
		t.Fatal(err)
	}
	for _, expected := range []string{"```markdown\n## MODIFIED\n### Requirement: Fake\n#### Scenario: Fake\n```", "## Requirement: New"} {
		if !strings.Contains(string(got), expected) {
			t.Errorf("ApplyDelta() lost or changed fenced text %q in %q", expected, got)
		}
	}
}

func TestParseDeltaParsesEveryRequirementInOneSection(t *testing.T) {
	delta := []byte("## ADDED\n### Requirement: First\nFirst body.\n### Requirement: Second\nSecond body.\n")
	ops, err := ParseDelta(delta)
	if err != nil {
		t.Fatal(err)
	}
	if len(ops) != 2 || ops[0].Name != "First" || ops[1].Name != "Second" {
		t.Fatalf("ParseDelta() = %#v, want First and Second", ops)
	}
	if _, err := ApplyDelta([]byte("## Requirement: First\nExisting.\n"), delta); err == nil {
		t.Fatal("ApplyDelta() error = nil; duplicate ADDED First must be rejected")
	}
}

func TestApplyDeltaPreservesLineEndingsAndSeparatesFollowingBlocks(t *testing.T) {
	spec := []byte("## Requirement: Existing\r\nOld.\r\n\r\n## Notes\r\nKeep this table.\r\n| A | B |\r\n\n## Requirement: Unchanged\r\nLeave byte-for-byte.\r\n")
	delta := []byte("## MODIFIED\r\n### Requirement: Existing\r\nNew body without final newline")
	got, err := ApplyDelta(spec, delta)
	if err != nil {
		t.Fatal(err)
	}
	text := string(got)
	if !strings.Contains(text, "New body without final newline\n\n## Notes\r\n") {
		t.Fatalf("replacement is joined to following content: %q", text)
	}
	if !strings.Contains(text, "| A | B |\r\n") || !strings.Contains(text, "## Requirement: Unchanged\r\nLeave byte-for-byte.\r\n") {
		t.Fatalf("ApplyDelta() changed untouched CRLF content: %q", text)
	}
	if !strings.Contains(text, "New body without final newline") || strings.Contains(text, "New body without final newline\r\n") {
		t.Fatalf("delta body line ending changed unexpectedly: %q", text)
	}
}

func TestApplyDeltaAddedModifiedRemovedAndPreservesOutsideText(t *testing.T) {
	spec := []byte("# Product\n\nKeep preface.\n\n## Requirement: Existing\nOld text.\n### Scenario: Old\nOld outcome.\n\n## Requirement: Removed\nDelete this.\n\n## Notes\nKeep this table.\n| left | right |\n")
	delta := []byte("## ADDED\n### Requirement: New\nNew text.\n#### Scenario: Added\nResult.\n## MODIFIED\n### Requirement: Existing\nUpdated text.\n#### Scenario: Changed\nNew outcome.\n## REMOVED\n### Requirement: Removed\n")
	got, err := ApplyDelta(spec, delta)
	if err != nil {
		t.Fatal(err)
	}
	text := string(got)
	for _, expected := range []string{"Keep preface.", "## Requirement: Existing\nUpdated text.", "### Scenario: Changed", "## Requirement: New", "### Scenario: Added", "## Notes\nKeep this table.", "| left | right |"} {
		if !strings.Contains(text, expected) {
			t.Errorf("ApplyDelta() missing %q in %q", expected, text)
		}
	}
	if strings.Contains(text, "Old text") || strings.Contains(text, "Delete this") || strings.Contains(text, "#### Scenario") {
		t.Fatalf("ApplyDelta() retained old content or delta heading level: %q", text)
	}
}

func TestApplyDeltaValidatesAllNamesBeforeChanging(t *testing.T) {
	for _, test := range []struct {
		name  string
		spec  string
		delta string
	}{
		{name: "added exists", spec: "## Requirement: Same\nText\n", delta: "## ADDED\n### Requirement: Same\nNew\n"},
		{name: "modified missing", spec: "", delta: "## MODIFIED\n### Requirement: Missing\nText\n"},
		{name: "removed missing", spec: "", delta: "## REMOVED\n### Requirement: Missing\n"},
		{name: "duplicate spec", spec: "## Requirement: Same\nA\n## Requirement: Same\nB\n", delta: "## MODIFIED\n### Requirement: Same\nC\n"},
	} {
		t.Run(test.name, func(t *testing.T) {
			if _, err := ApplyDelta([]byte(test.spec), []byte(test.delta)); err == nil {
				t.Fatal("ApplyDelta() error = nil, want validation error")
			}
		})
	}
}

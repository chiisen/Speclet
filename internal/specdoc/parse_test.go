package specdoc

import (
	"bytes"
	"strings"
	"testing"
)

func TestParseSpecFindsRequirementsAndPreservesOrdinaryMarkdown(t *testing.T) {
	src := []byte("# Payments\n\nIntro table.\n\n## Requirement: Transfer\nText.\n\n### Scenario: Accepted\n- **WHEN** sent\n- **THEN** accepted\n\n```markdown\n## Requirement: Example\n### Scenario: Example\n```\n\n## Notes\nKeep this table.\n| A | B |\n\n## Requirement: Refund\nRefund text.\n")
	doc, err := ParseSpec(src)
	if err != nil {
		t.Fatal(err)
	}
	if len(doc.Requirements) != 2 {
		t.Fatalf("got %d requirements, want 2", len(doc.Requirements))
	}
	if doc.Requirements[0].Name != "Transfer" || doc.Requirements[1].Name != "Refund" {
		t.Fatalf("requirement names = %q, %q", doc.Requirements[0].Name, doc.Requirements[1].Name)
	}
	if !strings.Contains(string(doc.Prefix), "Intro table.") {
		t.Fatalf("prefix lost ordinary introduction: %q", doc.Prefix)
	}
	if !strings.Contains(string(doc.Requirements[0].Raw), "### Scenario: Accepted") || !strings.Contains(string(doc.Requirements[0].Raw), "Requirement: Example") {
		t.Fatalf("first requirement raw block is incorrect: %q", doc.Requirements[0].Raw)
	}
	if !strings.Contains(string(doc.Suffix), "## Notes\nKeep this table.") {
		t.Fatalf("suffix lost ordinary section: %q", doc.Suffix)
	}
}

func TestParseSpecRejectsInvalidAndDuplicateMarkers(t *testing.T) {
	for _, src := range []string{
		"## Requirement:\ntext\n",
		"## Requirement: Same\nA\n## Requirement: Same\nB\n",
		"### Scenario: Orphan\ntext\n",
		"## Requirement: R\n#### Scenario: Wrong level\n",
	} {
		if _, err := ParseSpec([]byte(src)); err == nil {
			t.Errorf("ParseSpec(%q) error = nil, want validation error", src)
		}
	}
}

func TestParseSpecDoesNotParseMarkersInsideFences(t *testing.T) {
	src := []byte("## Requirement: Real\n```\n## Requirement: Fake\n### Scenario: Fake\n```\n")
	doc, err := ParseSpec(src)
	if err != nil {
		t.Fatal(err)
	}
	if len(doc.Requirements) != 1 || doc.Requirements[0].Name != "Real" {
		t.Fatalf("requirements = %#v, want only Real", doc.Requirements)
	}
}

func TestApplyRemovedRequirementPreservesFollowingH1Appendix(t *testing.T) {
	spec := []byte("## Requirement: Remove me\nRequirement text.\n\n# Appendix\nKeep this appendix.\n")
	delta := []byte("## REMOVED\n### Requirement: Remove me\n")
	got, err := ApplyDelta(spec, delta)
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(got), "Remove me") || !strings.Contains(string(got), "# Appendix\nKeep this appendix.") {
		t.Fatalf("ApplyDelta() = %q; removed requirement but should retain appendix", got)
	}
}

func TestApplyDeltaRejectsUnclosedFenceBeforeAppendingRequirement(t *testing.T) {
	spec := []byte("# Rules\n\n## Notes\n```markdown\n")
	delta := []byte("## ADDED\n### Requirement: New\nText\n")
	if _, err := ApplyDelta(spec, delta); err == nil {
		t.Fatal("ApplyDelta() error = nil, want unclosed formal code fence error")
	}
}

func TestApplyRemovedPreservesTabAndSetextAppendices(t *testing.T) {
	for _, appendix := range []string{
		"#\tAppendix\nKeep tab heading.\n",
		"Appendix\n=======\nKeep setext heading.\n",
		"Appendix of the\nrelease\n=======\nKeep multi-line setext heading.\n",
	} {
		spec := []byte("## Requirement: Remove me\nRequirement text.\n\n" + appendix)
		got, err := ApplyDelta(spec, []byte("## REMOVED\n### Requirement: Remove me\n"))
		if err != nil {
			t.Fatalf("ApplyDelta(%q): %v", appendix, err)
		}
		if strings.Contains(string(got), "Requirement text") || !strings.Contains(string(got), appendix) {
			t.Errorf("Appendix not preserved exactly: got %q, want %q", got, appendix)
		}
	}
}

func TestApplyRemovedDoesNotMisreadListItemAndHorizontalRuleAsSetext(t *testing.T) {
	spec := []byte("## Requirement: Remove me\nRequirement text.\n- one list item\n---\n")
	got, err := ApplyDelta(spec, []byte("## REMOVED\n### Requirement: Remove me\n"))
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(got), "one list item") || strings.Contains(string(got), "---") {
		t.Fatalf("removed requirement content remains: %q", got)
	}
}

func TestApplyDeltaPreservesBOMPrefaceBytes(t *testing.T) {
	spec := []byte("\xef\xbb\xbf# Readme\n\n## Requirement: Existing\nOld.\n")
	delta := []byte("## MODIFIED\n### Requirement: Existing\nNew.\n")
	got, err := ApplyDelta(spec, delta)
	if err != nil {
		t.Fatal(err)
	}
	wantPrefix := "\xef\xbb\xbf# Readme\n\n## Requirement: Existing\nNew.\n"
	if string(got) != wantPrefix {
		t.Fatalf("ApplyDelta() = %q, want exact BOM preface and update %q", got, wantPrefix)
	}
}

func TestApplyDeltaRejectsScenarioAfterTopLevelAppendix(t *testing.T) {
	delta := []byte("## ADDED\n### Requirement: A\nBody.\n#\tAppendix\n#### Scenario: Outside requirement\n")
	if _, err := ApplyDelta(nil, delta); err == nil {
		t.Fatal("ApplyDelta() error = nil for a Scenario outside its Requirement")
	}
}

func TestParseSpecAndDeltaRejectInvalidUTF8(t *testing.T) {
	invalid := []byte{0xff}
	if _, err := ParseSpec(invalid); err == nil {
		t.Fatal("ParseSpec() accepted invalid UTF-8")
	}
	if _, err := ParseDelta(invalid); err == nil {
		t.Fatal("ParseDelta() accepted invalid UTF-8")
	}
}

func TestParseAndApplyDeltaPreserveLeadingUTF8BOM(t *testing.T) {
	spec := []byte("\xef\xbb\xbf## Requirement: Existing\nOld.\n")
	delta := []byte("\xef\xbb\xbf## MODIFIED\n### Requirement: Existing\nNew.\n")
	if _, err := ParseDelta(delta); err != nil {
		t.Fatalf("ParseDelta() rejected UTF-8 BOM: %v", err)
	}
	got, err := ApplyDelta(spec, delta)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.HasPrefix(got, []byte("\xef\xbb\xbf")) || !strings.Contains(string(got), "## Requirement: Existing\nNew.") {
		t.Fatalf("ApplyDelta() lost BOM or update: %q", got)
	}
}

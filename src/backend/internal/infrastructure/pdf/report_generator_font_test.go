package pdf

import "testing"

func TestBuildFontCandidates_OrderAndDedupe(t *testing.T) {
	configured := "./fonts/SourceHanSans-Regular.ttf"
	env := "./fonts/SourceHanSans-Regular.ttf"

	got := buildFontCandidates(configured, env)
	if len(got) == 0 {
		t.Fatal("expected non-empty candidates")
	}

	if got[0] != configured {
		t.Fatalf("expected first candidate to be %q, got %q", configured, got[0])
	}

	count := 0
	for _, item := range got {
		if item == configured {
			count++
		}
	}
	if count != 1 {
		t.Fatalf("expected configured font path to appear once, got %d", count)
	}
}

func TestBuildFontCandidates_IncludesConfiguredSiblingTTC(t *testing.T) {
	configured := "/opt/bridge/fonts/SourceHanSans-Regular.ttf"
	got := buildFontCandidates(configured, "")

	want := "/opt/bridge/fonts/wqy-microhei.ttc"
	found := false
	for _, item := range got {
		if item == want {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("expected candidates to include sibling TTC font %q, got %v", want, got)
	}
}

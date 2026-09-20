package scopecheck

import (
	"testing"

	"go-pr-review-assistant/internal/gitdiff"
)

// TestClassifyAndAnalyze covers path classification, area detection and the two
// scope heuristics plus the missing-test warning.
func TestClassifyAndAnalyze(t *testing.T) {
	classify := []struct {
		path string
		want string
	}{
		{"internal/report/report.go", CatSource},
		{"internal/report/report_test.go", CatTests},
		{"tests/integration/run.py", CatTests},
		{"scripts/deploy.sh", CatSource},
		{"docs/usage.md", CatDocs},
		{"README.md", CatDocs},
		{"go.mod", CatConfig},
		{".gitignore", CatConfig},
		{".github/workflows/ci.yml", CatConfig},
		{"config/settings.yaml", CatConfig},
		{"assets/logo.png", CatOther},
	}
	for _, tc := range classify {
		if got := Classify(tc.path); got != tc.want {
			t.Errorf("Classify(%q) = %q, want %q", tc.path, got, tc.want)
		}
	}

	areas := []struct {
		path string
		want string
	}{
		{"internal/report/report.go", "internal/report"},
		{"internal/gitdiff/gitdiff.go", "internal/gitdiff"},
		{"cmd/review/main.go", "cmd/review"},
		{"server/handler.go", "server"},
		{"main.go", "(repo root)"},
	}
	for _, tc := range areas {
		if got := Area(tc.path); got != tc.want {
			t.Errorf("Area(%q) = %q, want %q", tc.path, got, tc.want)
		}
	}

	// One source area with a real test change: no findings, no warning.
	tidy := Analyze([]gitdiff.Change{
		{Status: "M", Path: "internal/report/report.go"},
		{Status: "M", Path: "internal/report/report_test.go"},
	})
	if len(tidy.Findings) != 0 {
		t.Errorf("tidy change should have no findings, got %v", tidy.Findings)
	}
	if tidy.TestWarning != "" {
		t.Errorf("tidy change should have no test warning, got %q", tidy.TestWarning)
	}
	if len(tidy.SourceAreas) != 1 || tidy.SourceAreas[0] != "internal/report" {
		t.Errorf("unexpected source areas: %v", tidy.SourceAreas)
	}
	if len(tidy.Categories[CatTests]) != 1 {
		t.Errorf("expected 1 test file, got %v", tidy.Categories[CatTests])
	}

	// Three source areas and only a deleted test: both signals should fire.
	spread := Analyze([]gitdiff.Change{
		{Status: "M", Path: "internal/report/report.go"},
		{Status: "M", Path: "internal/gitdiff/gitdiff.go"},
		{Status: "A", Path: "cmd/review/main.go"},
		{Status: "D", Path: "internal/report/report_test.go"},
	})
	if len(spread.SourceAreas) != 3 {
		t.Errorf("expected 3 source areas, got %v", spread.SourceAreas)
	}
	if len(spread.Findings) != 1 {
		t.Errorf("expected exactly the area finding, got %v", spread.Findings)
	}
	if spread.TestWarning == "" {
		t.Error("deleted-only test change should still warn about missing tests")
	}

	// More than MaxChangedFiles files in a single area triggers the size finding.
	var many []gitdiff.Change
	for _, p := range []string{"a", "b", "c", "d", "e", "f", "g", "h", "i", "j", "k"} {
		many = append(many, gitdiff.Change{Status: "M", Path: "internal/report/" + p + ".go"})
	}
	many = append(many, gitdiff.Change{Status: "M", Path: "internal/report/report_test.go"})
	size := Analyze(many)
	if len(size.Findings) != 1 {
		t.Errorf("expected exactly the file-count finding, got %v", size.Findings)
	}
	if size.TestWarning != "" {
		t.Errorf("tests were modified, so no warning expected, got %q", size.TestWarning)
	}
}

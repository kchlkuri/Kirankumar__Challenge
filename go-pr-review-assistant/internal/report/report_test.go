package report

import (
	"os"
	"strconv"
	"strings"
	"testing"

	"go-pr-review-assistant/internal/gitdiff"
	"go-pr-review-assistant/internal/history"
	"go-pr-review-assistant/internal/scopecheck"
)

// fixturePath is the shared fixture at the repository root.
const fixturePath = "../../testdata/sample_changes.tsv"

// TestRender checks the Markdown for a fixture diff and for an empty diff.
func TestRender(t *testing.T) {
	changes := loadFixture(t)
	if len(changes) != 8 {
		t.Fatalf("fixture should describe 8 changes, got %d", len(changes))
	}

	md := Render(Data{
		Repo:    "/tmp/demo",
		BaseRef: "main",
		BaseSHA: "1111111111111111111111111111111111111111",
		HeadSHA: "2222222222222222222222222222222222222222",
		Dirty:   true,
		Changes: changes,
		Scope:   scopecheck.Analyze(changes),
		Histories: []history.PathHistory{
			{Path: "cmd/review/main.go", Commits: []history.Commit{
				{SHA: "abcdef12", Author: "Dana", Date: "2026-09-01", Subject: "Add flag parsing"},
			}},
			{Path: "internal/scopecheck/scopecheck.go"}, // new file, no prior commits
		},
	})

	sections := []string{
		"# PR Review Assistant Report",
		"## Summary", "## Changed Files", "## Scope Findings",
		"## Test Warnings", "## Historical Context", "## Review Recommendations",
	}
	for _, s := range sections {
		if !strings.Contains(md, s) {
			t.Errorf("report is missing section %q", s)
		}
	}

	wants := []string{
		"Changed files: 8",
		"+207 / -141",   // line totals, binary file excluded
		"no merge base", // comparison method is stated
		"1111111111111111111111111111111111111111",
		"uncommitted changes", // dirty worktree note
		"`internal/gitdiff/gitdiff.go`",
		"| deleted | source | `internal/legacy/old_helper.go` |",
		"| added | other | `assets/logo.png` | binary |",
		"Source changes span 4 areas",        // heuristic fired
		"heuristics, not conclusions",        // hedged language
		"no test file was added or modified", // only a deleted test in the fixture
		"### `internal/scopecheck/scopecheck.go`",
		"`abcdef12` 2026-09-01: Add flag parsing (Dana)",
		"No prior commits at the base ref",
		"Files were deleted", // recommendation
	}
	for _, w := range wants {
		if !strings.Contains(md, w) {
			t.Errorf("report is missing %q\n---\n%s", w, md)
		}
	}
	if strings.Contains(md, "docs updates") {
		t.Error("fixture changes docs, so the docs recommendation should be absent")
	}

	// Markdown-hostile names and subjects must not break a table or list row.
	messy := []gitdiff.Change{{Status: "M", Path: "pkg/we|ird`name.go"}}
	escaped := Render(Data{
		Repo:    "/tmp/demo",
		BaseRef: "main",
		Changes: messy,
		Scope:   scopecheck.Analyze(messy),
		Histories: []history.PathHistory{{Path: messy[0].Path, Commits: []history.Commit{
			{SHA: "1234abcd", Author: "Lee", Date: "2026-09-02", Subject: "fix a|b\nand *c*"},
		}}},
	})
	if !strings.Contains(escaped, "`pkg/we\\|ird'name.go`") {
		t.Errorf("path was not escaped for Markdown\n---\n%s", escaped)
	}
	if !strings.Contains(escaped, "fix a\\|b and \\*c\\*") {
		t.Errorf("commit subject was not escaped or flattened\n---\n%s", escaped)
	}

	// An empty diff must still render an honest, complete report.
	empty := Render(Data{
		Repo:    "/tmp/demo",
		BaseRef: "main",
		BaseSHA: "3333333333333333333333333333333333333333",
		HeadSHA: "3333333333333333333333333333333333333333",
		Scope:   scopecheck.Analyze(nil),
	})
	for _, s := range sections {
		if !strings.Contains(empty, s) {
			t.Errorf("empty report is missing section %q", s)
		}
	}
	for _, w := range []string{"Changed files: 0", "same commit", "No files changed", "Nothing to review"} {
		if !strings.Contains(empty, w) {
			t.Errorf("empty report is missing %q\n---\n%s", w, empty)
		}
	}
	if strings.Contains(empty, "Source changes span") {
		t.Error("empty diff must not produce scope findings")
	}
}

// loadFixture parses testdata/sample_changes.tsv into changes.
func loadFixture(t *testing.T) []gitdiff.Change {
	t.Helper()
	raw, err := os.ReadFile(fixturePath)
	if err != nil {
		t.Fatalf("reading fixture: %v", err)
	}
	var changes []gitdiff.Change
	for _, line := range strings.Split(string(raw), "\n") {
		if strings.TrimSpace(line) == "" || strings.HasPrefix(line, "#") {
			continue
		}
		f := strings.Split(line, "\t")
		if len(f) != 4 {
			t.Fatalf("bad fixture line %q", line)
		}
		changes = append(changes, gitdiff.Change{
			Status:  f[0],
			Path:    f[1],
			Added:   count(f[2]),
			Deleted: count(f[3]),
		})
	}
	return changes
}

// count turns a fixture column into a line count ("-" means binary).
func count(s string) int {
	n, err := strconv.Atoi(strings.TrimSpace(s))
	if err != nil {
		return -1
	}
	return n
}

// Package report renders review notes as Markdown.
package report

import (
	"fmt"
	"strings"

	"go-pr-review-assistant/internal/gitdiff"
	"go-pr-review-assistant/internal/history"
	"go-pr-review-assistant/internal/scopecheck"
)

// Data is everything collected about one base..HEAD comparison.
type Data struct {
	Repo      string
	BaseRef   string
	BaseSHA   string
	HeadSHA   string
	Dirty     bool // uncommitted changes exist and are excluded here
	Changes   []gitdiff.Change
	Scope     scopecheck.Result
	Histories []history.PathHistory
}

// Render builds the full Markdown report.
func Render(d Data) string {
	var b strings.Builder
	b.WriteString("# PR Review Assistant Report\n\n")
	writeSummary(&b, d)
	writeChangedFiles(&b, d)
	writeScope(&b, d)
	writeTestWarnings(&b, d)
	writeHistory(&b, d)
	writeRecommendations(&b, d)
	return b.String()
}

func writeSummary(b *strings.Builder, d Data) {
	b.WriteString("## Summary\n\n")
	fmt.Fprintf(b, "- Repository: %s\n", code(d.Repo))
	fmt.Fprintf(b, "- Base ref: %s (resolved to %s)\n", code(d.BaseRef), code(d.BaseSHA))
	fmt.Fprintf(b, "- Head commit: %s\n", code(d.HeadSHA))
	fmt.Fprintf(b, "- Comparison: direct %s (no merge base)\n",
		code("git diff "+short(d.BaseSHA)+" "+short(d.HeadSHA)))
	fmt.Fprintf(b, "- Changed files: %d\n", len(d.Changes))

	if len(d.Changes) == 0 {
		if d.BaseSHA == d.HeadSHA {
			b.WriteString("- Base and head are the same commit, so there is nothing to review.\n")
		} else {
			b.WriteString("- No committed differences between these two commits.\n")
		}
	} else {
		added, deleted := totals(d.Changes)
		fmt.Fprintf(b, "- Line changes: +%d / -%d (binary files excluded)\n", added, deleted)
		fmt.Fprintf(b, "- Categories: %s\n", categorySummary(d.Scope))
	}
	if d.Dirty {
		b.WriteString("- Note: the working tree has uncommitted changes. " +
			"Only committed work is included in this report.\n")
	}
	b.WriteString("\n")
}

func writeChangedFiles(b *strings.Builder, d Data) {
	b.WriteString("## Changed Files\n\n")
	if len(d.Changes) == 0 {
		b.WriteString("No files changed between the base and head commits.\n\n")
		return
	}
	b.WriteString("Renames are reported as a delete plus an add.\n\n")
	b.WriteString("| Status | Category | File | +/- |\n|---|---|---|---|\n")
	for _, c := range d.Changes {
		fmt.Fprintf(b, "| %s | %s | %s | %s |\n",
			statusLabel(c.Status), scopecheck.Classify(c.Path), code(c.Path), lineStat(c))
	}
	b.WriteString("\n")
}

func writeScope(b *strings.Builder, d Data) {
	b.WriteString("## Scope Findings\n\n")
	if len(d.Changes) == 0 {
		b.WriteString("Nothing to assess: the diff is empty.\n\n")
		return
	}
	if len(d.Scope.SourceAreas) > 0 {
		b.WriteString("Source areas touched: " + codeList(d.Scope.SourceAreas) + "\n\n")
	}
	if len(d.Scope.Findings) == 0 {
		fmt.Fprintf(b, "No scope concerns from the current heuristics (flag at %d+ source areas "+
			"or more than %d changed files).\n\n", scopecheck.MaxSourceAreas, scopecheck.MaxChangedFiles)
		return
	}
	b.WriteString("These are heuristics, not conclusions. A wide change can still be correct.\n\n")
	for _, f := range d.Scope.Findings {
		b.WriteString("- " + text(f) + "\n")
	}
	b.WriteString("\n")
}

func writeTestWarnings(b *strings.Builder, d Data) {
	b.WriteString("## Test Warnings\n\n")
	switch {
	case len(d.Changes) == 0:
		b.WriteString("No changes, so no test warnings.\n\n")
	case d.Scope.TestWarning != "":
		b.WriteString("- " + text(d.Scope.TestWarning) + "\n\n")
	default:
		b.WriteString("No test warnings: either no source changed, or tests were added " +
			"or modified alongside it.\n\n")
	}
}

func writeHistory(b *strings.Builder, d Data) {
	b.WriteString("## Historical Context\n\n")
	if len(d.Histories) == 0 {
		b.WriteString("No prior commit history was collected for the changed paths.\n\n")
		return
	}
	fmt.Fprintf(b, "Up to %d commits per path, taken from %s (the state before this change). "+
		"Subjects are quoted from git; no reasoning is inferred.\n\n",
		history.MaxCommitsPerPath, code(d.BaseRef))
	for _, h := range d.Histories {
		fmt.Fprintf(b, "### %s\n\n", code(h.Path))
		if len(h.Commits) == 0 {
			b.WriteString("- No prior commits at the base ref (likely a new file).\n\n")
			continue
		}
		for _, c := range h.Commits {
			fmt.Fprintf(b, "- %s %s: %s (%s)\n",
				code(c.SHA), text(c.Date), text(c.Subject), text(c.Author))
		}
		b.WriteString("\n")
	}
}

func writeRecommendations(b *strings.Builder, d Data) {
	b.WriteString("## Review Recommendations\n\n")
	if len(d.Changes) == 0 {
		b.WriteString("- Nothing to review. Point `--base` at the branch or commit this work forked from.\n")
		return
	}
	var recs []string
	if len(d.Scope.Findings) > 0 {
		recs = append(recs, "Check whether the flagged breadth above is one coherent change or several.")
	}
	if d.Scope.TestWarning != "" {
		recs = append(recs, "Ask for a test covering the changed behaviour, or a note on why one is not needed.")
	}
	if len(d.Scope.Categories[scopecheck.CatConfig]) > 0 {
		recs = append(recs, "Config files changed: review them for environment or deployment impact.")
	}
	if len(d.Scope.Categories[scopecheck.CatSource]) > 0 && len(d.Scope.Categories[scopecheck.CatDocs]) == 0 {
		recs = append(recs, "Source changed without docs updates; confirm no user-facing docs need a change.")
	}
	if deletions(d.Changes) > 0 {
		recs = append(recs, "Files were deleted; confirm nothing still references them.")
	}
	recs = append(recs, "Read the diff itself: this report summarises shape, not correctness.")
	for _, r := range recs {
		b.WriteString("- " + r + "\n")
	}
}

func totals(changes []gitdiff.Change) (int, int) {
	var added, deleted int
	for _, c := range changes {
		if c.Added > 0 {
			added += c.Added
		}
		if c.Deleted > 0 {
			deleted += c.Deleted
		}
	}
	return added, deleted
}

func deletions(changes []gitdiff.Change) int {
	n := 0
	for _, c := range changes {
		if c.Status == "D" {
			n++
		}
	}
	return n
}

func categorySummary(s scopecheck.Result) string {
	order := []string{scopecheck.CatSource, scopecheck.CatTests, scopecheck.CatDocs,
		scopecheck.CatConfig, scopecheck.CatOther}
	var parts []string
	for _, cat := range order {
		if n := len(s.Categories[cat]); n > 0 {
			parts = append(parts, fmt.Sprintf("%s %d", cat, n))
		}
	}
	if len(parts) == 0 {
		return "none"
	}
	return strings.Join(parts, ", ")
}

func statusLabel(status string) string {
	switch status {
	case "A":
		return "added"
	case "M":
		return "modified"
	case "D":
		return "deleted"
	case "T":
		return "type changed"
	default:
		return text(status)
	}
}

func lineStat(c gitdiff.Change) string {
	if c.Added < 0 || c.Deleted < 0 {
		return "binary"
	}
	return fmt.Sprintf("+%d / -%d", c.Added, c.Deleted)
}

func codeList(items []string) string {
	out := make([]string, 0, len(items))
	for _, it := range items {
		out = append(out, code(it))
	}
	return strings.Join(out, ", ")
}

// code renders a value as an inline code span. Backticks become quotes and pipes
// are escaped so table rows stay well formed.
func code(s string) string {
	s = flatten(s)
	s = strings.ReplaceAll(s, "`", "'")
	return "`" + strings.ReplaceAll(s, "|", "\\|") + "`"
}

// text escapes git-supplied prose (commit subjects, authors) for Markdown.
func text(s string) string {
	s = flatten(s)
	for _, ch := range []string{"\\", "`", "*", "_", "|", "<", ">"} {
		s = strings.ReplaceAll(s, ch, "\\"+ch)
	}
	return s
}

// flatten collapses characters that would break a line or table cell.
func flatten(s string) string {
	s = strings.ReplaceAll(s, "\r\n", " ")
	for _, ch := range []string{"\r", "\n", "\t"} {
		s = strings.ReplaceAll(s, ch, " ")
	}
	return strings.TrimSpace(s)
}

func short(sha string) string {
	if len(sha) > 8 {
		return sha[:8]
	}
	return sha
}

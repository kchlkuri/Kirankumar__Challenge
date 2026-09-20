// Command review compares a base commit with HEAD in a local git repo and writes
// a Markdown review report. Local git only; no network calls.
//
// Usage: go run ./cmd/review --repo PATH --base main --output PATH
package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"

	"go-pr-review-assistant/internal/gitdiff"
	"go-pr-review-assistant/internal/history"
	"go-pr-review-assistant/internal/report"
	"go-pr-review-assistant/internal/scopecheck"
)

func main() {
	repo := flag.String("repo", ".", "path to the local git repository")
	base := flag.String("base", "main", "base ref to compare against (branch, tag, or SHA)")
	output := flag.String("output", "review-report.md", "path to write the Markdown report to")
	flag.Parse()

	if err := run(*repo, *base, *output); err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
}

func run(repo, base, output string) error {
	absRepo, err := filepath.Abs(repo)
	if err != nil {
		return err
	}
	if !gitdiff.IsRepo(absRepo) {
		return fmt.Errorf("%s is not a git working tree", absRepo)
	}

	baseSHA, err := gitdiff.ResolveCommit(absRepo, base)
	if err != nil {
		return fmt.Errorf("resolving base ref: %w", err)
	}
	headSHA, err := gitdiff.ResolveCommit(absRepo, "HEAD")
	if err != nil {
		return fmt.Errorf("resolving HEAD: %w", err)
	}

	changes, err := gitdiff.Changes(absRepo, baseSHA, headSHA)
	if err != nil {
		return err
	}
	dirty, err := gitdiff.WorktreeDirty(absRepo)
	if err != nil {
		return err
	}
	histories, err := history.ForPaths(absRepo, baseSHA, paths(changes))
	if err != nil {
		return err
	}

	md := report.Render(report.Data{
		Repo:      absRepo,
		BaseRef:   base,
		BaseSHA:   baseSHA,
		HeadSHA:   headSHA,
		Dirty:     dirty,
		Changes:   changes,
		Scope:     scopecheck.Analyze(changes),
		Histories: histories,
	})
	if err := writeNew(output, md); err != nil {
		return err
	}

	fmt.Printf("wrote %s (%d changed files, base %s, head %s)\n",
		output, len(changes), short(baseSHA), short(headSHA))
	return nil
}

func paths(changes []gitdiff.Change) []string {
	out := make([]string, 0, len(changes))
	for _, c := range changes {
		out = append(out, c.Path)
	}
	return out
}

// writeNew creates output only if it does not already exist, so an earlier
// report is never overwritten.
func writeNew(path, content string) error {
	if dir := filepath.Dir(path); dir != "." {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			return err
		}
	}
	f, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0o644)
	if err != nil {
		if os.IsExist(err) {
			return fmt.Errorf("%s already exists; pass a different --output path", path)
		}
		return err
	}
	defer f.Close()
	if _, err := f.WriteString(content); err != nil {
		return err
	}
	return f.Close()
}

func short(sha string) string {
	if len(sha) > 8 {
		return sha[:8]
	}
	return sha
}

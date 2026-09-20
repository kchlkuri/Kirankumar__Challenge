// Package gitdiff collects committed changes between two commits using local git.
package gitdiff

import (
	"bytes"
	"fmt"
	"os/exec"
	"strconv"
	"strings"
)

// Change is one file changed between base and head. Renames are split into D + A.
type Change struct {
	Status  string // A, M, D, T
	Path    string
	Added   int // -1 for binary
	Deleted int
}

// Run executes git inside repo. --literal-pathspecs stops paths with glob
// characters from being treated as patterns.
func Run(repo string, args ...string) (string, error) {
	full := append([]string{"-C", repo, "--literal-pathspecs"}, args...)
	cmd := exec.Command("git", full...)
	var out, errBuf bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &errBuf
	if err := cmd.Run(); err != nil {
		msg := strings.TrimSpace(errBuf.String())
		if msg == "" {
			msg = err.Error()
		}
		return "", fmt.Errorf("git %s: %s", strings.Join(args, " "), msg)
	}
	return out.String(), nil
}

// ResolveCommit turns a ref into a full commit SHA. --end-of-options keeps a ref
// starting with "-" from being read as a flag.
func ResolveCommit(repo, ref string) (string, error) {
	out, err := Run(repo, "rev-parse", "--verify", "--quiet", "--end-of-options", ref+"^{commit}")
	sha := strings.TrimSpace(out)
	if err != nil || sha == "" {
		return "", fmt.Errorf("%q is not a commit, branch, or tag in this repo", ref)
	}
	return sha, nil
}

// IsRepo reports whether repo is a git working tree.
func IsRepo(repo string) bool {
	out, err := Run(repo, "rev-parse", "--is-inside-work-tree")
	return err == nil && strings.TrimSpace(out) == "true"
}

// WorktreeDirty reports whether uncommitted changes exist; they are excluded
// from the report, so it is worth telling the reader.
func WorktreeDirty(repo string) (bool, error) {
	out, err := Run(repo, "status", "--porcelain")
	if err != nil {
		return false, err
	}
	return strings.TrimSpace(out) != "", nil
}

// Changes diffs two commits directly, not via a merge base.
func Changes(repo, baseSHA, headSHA string) ([]Change, error) {
	diffArgs := []string{"diff", "--no-ext-diff", "--no-textconv", "--no-renames"}
	nameStatus, err := Run(repo, append(diffArgs, "--name-status", "-z", baseSHA, headSHA)...)
	if err != nil {
		return nil, err
	}
	numstat, err := Run(repo, append(diffArgs, "--numstat", "-z", baseSHA, headSHA)...)
	if err != nil {
		return nil, err
	}

	changes := parseNameStatus(nameStatus)
	stats := parseNumstat(numstat)
	for i := range changes {
		if s, ok := stats[changes[i].Path]; ok {
			changes[i].Added, changes[i].Deleted = s.added, s.deleted
		}
	}
	return changes, nil
}

// parseNameStatus reads -z name-status: status NUL path NUL ...
func parseNameStatus(out string) []Change {
	fields := splitNUL(out)
	changes := make([]Change, 0, len(fields)/2)
	for i := 0; i+1 < len(fields); i += 2 {
		if fields[i] == "" || fields[i+1] == "" {
			continue
		}
		changes = append(changes, Change{Status: fields[i][:1], Path: fields[i+1]})
	}
	return changes
}

type lineStat struct{ added, deleted int }

// parseNumstat reads -z numstat records of "added\tdeleted\tpath", one per NUL
// field. The path keeps its exact bytes, including any tabs.
func parseNumstat(out string) map[string]lineStat {
	stats := map[string]lineStat{}
	for _, field := range splitNUL(out) {
		parts := strings.SplitN(field, "\t", 3)
		if len(parts) < 3 || parts[2] == "" {
			continue
		}
		stats[parts[2]] = lineStat{added: parseCount(parts[0]), deleted: parseCount(parts[1])}
	}
	return stats
}

// parseCount returns -1 for binary files, which git reports as "-".
func parseCount(s string) int {
	n, err := strconv.Atoi(s)
	if err != nil {
		return -1
	}
	return n
}

func splitNUL(out string) []string {
	var fields []string
	for _, f := range strings.Split(out, "\x00") {
		if f != "" {
			fields = append(fields, f)
		}
	}
	return fields
}

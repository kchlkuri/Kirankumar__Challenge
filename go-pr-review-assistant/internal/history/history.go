// Package history reports recent commits for changed paths at the base ref.
package history

import (
	"strings"

	"go-pr-review-assistant/internal/gitdiff"
)

// MaxCommitsPerPath caps how much history is shown per file.
const MaxCommitsPerPath = 3

// Commit holds fields copied straight from git; no rationale is inferred.
type Commit struct {
	SHA     string
	Author  string
	Date    string
	Subject string
}

// PathHistory groups prior commits for one changed path.
type PathHistory struct {
	Path    string
	Commits []Commit
}

// ForPaths returns up to MaxCommitsPerPath commits per path, looking back from
// baseRef. Rename following is off, so output matches the path as given.
func ForPaths(repo, baseRef string, paths []string) ([]PathHistory, error) {
	out := make([]PathHistory, 0, len(paths))
	for _, p := range paths {
		commits, err := forPath(repo, baseRef, p)
		if err != nil {
			return nil, err
		}
		out = append(out, PathHistory{Path: p, Commits: commits})
	}
	return out, nil
}

func forPath(repo, baseRef, p string) ([]Commit, error) {
	// Unit separator keeps subjects containing tabs or pipes intact.
	out, err := gitdiff.Run(repo, "log", "--no-follow", "--no-merges",
		"-n", "3", "--date=short", "--format=%H\x1f%an\x1f%ad\x1f%s",
		// gitdiff.Run passes --literal-pathspecs, so p is matched byte for byte.
		baseRef, "--", p)
	if err != nil {
		return nil, err
	}
	var commits []Commit
	for _, line := range strings.Split(strings.TrimSpace(out), "\n") {
		fields := strings.SplitN(line, "\x1f", 4)
		if len(fields) < 4 {
			continue
		}
		commits = append(commits, Commit{
			SHA:     shortSHA(fields[0]),
			Author:  fields[1],
			Date:    fields[2],
			Subject: fields[3],
		})
		if len(commits) == MaxCommitsPerPath {
			break
		}
	}
	return commits, nil
}

func shortSHA(sha string) string {
	if len(sha) > 8 {
		return sha[:8]
	}
	return sha
}

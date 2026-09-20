// Package scopecheck classifies changed files and applies simple, deterministic
// scope heuristics. Every finding here is a hint for a human reviewer, not a fact.
package scopecheck

import (
	"path"
	"sort"
	"strconv"
	"strings"

	"go-pr-review-assistant/internal/gitdiff"
)

// Categories a changed file can fall into.
const (
	CatSource = "source"
	CatTests  = "tests"
	CatDocs   = "docs"
	CatConfig = "config"
	CatOther  = "other"
)

// Thresholds for the two scope-creep hints.
const (
	MaxSourceAreas  = 3  // >= this many distinct source areas is flagged
	MaxChangedFiles = 10 // more than this many changed files is flagged
)

// Result is everything the report needs about scope.
type Result struct {
	Categories  map[string][]string // category -> sorted paths
	SourceAreas []string            // sorted directory areas touched by source files
	Findings    []string            // scope-creep hints
	TestWarning string              // empty when no warning applies
}

// Classify buckets a path by name and extension. Test detection wins over the
// language extension so that foo_test.go counts as tests, not source.
func Classify(p string) string {
	clean := strings.ToLower(path.Clean(strings.ReplaceAll(p, "\\", "/")))
	base := path.Base(clean)
	ext := path.Ext(base)

	if isTest(clean, base) {
		return CatTests
	}
	if isDocs(clean, base, ext) {
		return CatDocs
	}
	if isConfig(clean, base, ext) {
		return CatConfig
	}
	if sourceExts[ext] {
		return CatSource
	}
	return CatOther
}

func isTest(clean, base string) bool {
	if strings.HasSuffix(base, "_test.go") || strings.HasSuffix(base, "_test.py") ||
		strings.HasPrefix(base, "test_") || strings.HasSuffix(base, "test.java") {
		return true
	}
	if strings.Contains(base, ".test.") || strings.Contains(base, ".spec.") {
		return true
	}
	for _, dir := range []string{"test/", "tests/", "testing/", "__tests__/", "spec/", "testdata/"} {
		if strings.HasPrefix(clean, dir) || strings.Contains(clean, "/"+dir) {
			return true
		}
	}
	return false
}

func isDocs(clean, base, ext string) bool {
	if docExts[ext] && !configNames[base] {
		return true
	}
	if base == "license" || base == "notice" || base == "changelog" {
		return true
	}
	return strings.HasPrefix(clean, "docs/") || strings.Contains(clean, "/docs/")
}

func isConfig(clean, base, ext string) bool {
	if configExts[ext] {
		return true
	}
	if configNames[base] {
		return true
	}
	// Dotfiles at any level (.gitignore, .editorconfig, ...) and CI folders.
	if strings.HasPrefix(base, ".") && !strings.Contains(base[1:], ".") {
		return true
	}
	return strings.HasPrefix(clean, ".github/") || strings.Contains(clean, "/.github/")
}

var sourceExts = map[string]bool{
	".go": true, ".py": true, ".js": true, ".jsx": true, ".ts": true, ".tsx": true,
	".java": true, ".kt": true, ".rb": true, ".rs": true, ".c": true, ".h": true,
	".cc": true, ".cpp": true, ".hpp": true, ".cs": true, ".php": true, ".swift": true,
	".scala": true, ".sh": true, ".sql": true,
}

var docExts = map[string]bool{
	".md": true, ".markdown": true, ".rst": true, ".adoc": true, ".txt": true,
}

var configExts = map[string]bool{
	".yml": true, ".yaml": true, ".json": true, ".toml": true, ".ini": true,
	".cfg": true, ".conf": true, ".properties": true, ".tf": true, ".env": true,
}

var configNames = map[string]bool{
	"dockerfile": true, "makefile": true, "go.mod": true, "go.sum": true,
	"requirements.txt": true, "package.json": true, "package-lock.json": true,
}

// Area is the rough project area of a path, used to judge how spread out a
// change is. Nested container dirs keep their child so internal/a and
// internal/b count as two areas.
func Area(p string) string {
	clean := path.Clean(strings.ReplaceAll(p, "\\", "/"))
	parts := strings.Split(clean, "/")
	if len(parts) < 2 {
		return "(repo root)"
	}
	if containerDirs[strings.ToLower(parts[0])] && len(parts) > 2 {
		return parts[0] + "/" + parts[1]
	}
	return parts[0]
}

var containerDirs = map[string]bool{
	"internal": true, "cmd": true, "src": true, "pkg": true, "lib": true, "app": true, "apps": true,
}

// Analyze classifies every change and applies the scope heuristics.
func Analyze(changes []gitdiff.Change) Result {
	res := Result{Categories: map[string][]string{}}
	areaSeen := map[string]bool{}
	testsTouched := false

	for _, c := range changes {
		cat := Classify(c.Path)
		res.Categories[cat] = append(res.Categories[cat], c.Path)
		if cat == CatSource {
			areaSeen[Area(c.Path)] = true
		}
		// A deleted test file is not evidence that the change was tested.
		if cat == CatTests && c.Status != "D" {
			testsTouched = true
		}
	}
	for _, paths := range res.Categories {
		sort.Strings(paths)
	}
	for area := range areaSeen {
		res.SourceAreas = append(res.SourceAreas, area)
	}
	sort.Strings(res.SourceAreas)

	res.Findings = findings(len(changes), res.SourceAreas)
	if len(res.Categories[CatSource]) > 0 && !testsTouched {
		res.TestWarning = "Source files changed but no test file was added or modified. " +
			"Confirm the behaviour is covered before approving."
	}
	return res
}

func findings(fileCount int, areas []string) []string {
	var out []string
	if len(areas) >= MaxSourceAreas {
		out = append(out, "Source changes span "+strconv.Itoa(len(areas))+" areas ("+strings.Join(areas, ", ")+
			"). A change touching this many areas may be doing more than one thing.")
	}
	if fileCount > MaxChangedFiles {
		out = append(out, "This change touches "+strconv.Itoa(fileCount)+" files (threshold "+
			strconv.Itoa(MaxChangedFiles)+"). Consider whether it can be split into smaller reviews.")
	}
	return out
}

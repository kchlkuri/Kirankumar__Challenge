# Go PR Review Assistant

A local Go CLI that compares a repository's current commit with `main` and writes
a Markdown review report. It summarizes changed files, flags broad changes and
missing test updates, and includes recent history for touched paths.

The tool uses the Go standard library and local Git commands. It does not call
GitHub, run an LLM, post comments, or require an MCP server.

## Problem statement

A branch review involves more than reading a patch. Reviewers also need to see
which parts of the repository changed, whether tests changed with the source,
and what earlier commits say about the affected files.

This project collects those signals in one report. It provides a starting point
for a human review, not an automated approval or a substitute for running tests.

## Architecture

`cmd/review` runs a single process. `internal/gitdiff` reads the committed diff;
`internal/scopecheck` classifies paths and applies review rules;
`internal/history` reads earlier commits; and `internal/report` renders Markdown.
The CLI writes the report to the selected output path.

There are no external Go packages, persistent indexes, or background processes.
The target repository is read-only: the tool does not switch branches, edit source,
or execute the target project's code.

## Setup

Install Go 1.20 or newer and Git. From the repository root:

```bash
cd go-pr-review-assistant
go version
git --version
```

No API keys or services are needed. The module uses only standard-library imports,
so there are no third-party packages to download.

## Usage

Check out the branch to review in the target repository. From this project directory:

```bash
go run ./cmd/review \
  --repo "/absolute/path/to/target-repo" \
  --base main \
  --output /tmp/go-branch-review.md
```

The base must already exist locally; the tool does not fetch it. To use a different
base, pass a local ref such as `--base develop`.

To review this repository from the project directory:

```bash
go run ./cmd/review \
  --repo .. \
  --base main \
  --output /tmp/go-current-repo-review.md
```

Choose a new output filename for each run. Existing files are not overwritten.
The default output is `review-report.md` in the current directory. The included
`sample_output/sample_review.md` is a saved example; use a different output path
for your own review.

## Review rules

- **File types:** Paths are classified as source, tests, docs, config, or other.
  Classification uses filenames and directories, not language parsing.
- **Scope:** Three or more source path areas, or more than ten changed files,
  trigger a review warning. These thresholds do not prove that work is unrelated.
- **Tests:** Source changes without an accompanying added or modified test file
  trigger a warning. A deleted test does not count as added coverage.
- **History:** The report includes up to three non-merge commits per touched path,
  read from the base ref. Commit subjects are context, not verified intent.

## Tests

```bash
go test ./...
```

There are exactly two test functions: one for classification and scope/test warnings,
and one for Markdown report output. Test fixtures live in `testdata/`.

## Sample report

[The sample report](sample_output/sample_review.md) comes from a run against this
repository before the project was committed. The checkout was on `main`, so `main`
and `HEAD` pointed to the same commit and there were no committed changes to review.

The new project files were uncommitted and excluded from that comparison. This is
an actual empty-diff result, not a fabricated pull request; the report test covers
a populated report using local fixture data.

## Limitations

- **Comparison:** Uses the direct base-to-HEAD diff, not a merge-base comparison.
  If a branch is behind its base, base-only changes may appear reversed.
- **Uncommitted work:** Staged, unstaged, and untracked changes are not reviewed.
  Run the tool after committing the changes you want to inspect.
- **Coverage:** Test warnings are filename heuristics. The tool does not measure
  coverage, inspect assertions, or establish that changed behavior is tested.
- **Scope:** No PR description or intended scope is supplied. Related changes can
  span several directories, and unrelated edits can stay within one.
- **History:** Limited to locally available commits at the current path. Renames
  appear as deletion plus addition, and earlier names are not followed.
- **Review depth:** No patch-level correctness, security, or performance analysis.
  Read the actual diff and run relevant tests before approving a change.

## Future improvements

These are possible follow-ups, not part of version 1:

- Add merge-base comparison for branches that have diverged from their base.
- Evaluate test-file matching against repositories with different naming conventions.
- Add a small labeled set of branch reviews before changing the scope thresholds.

## Resume bullets

- Built a Go CLI that combines local branch diffs, file classification, and per-file commit history into Markdown review reports.
- Implemented explicit heuristics for broad change scope and source changes without test updates, using only the Go standard library and Git.
- Validated classification, review rules, and Markdown output with two automated tests and captured a review of a real local repository.

## References

Only these three reference paths were inspected. The implementation is original;
the references supplied ideas, not copied code or infrastructure.

- [GitHub MCP Agent](https://github.com/Shubhamsaboo/awesome-llm-apps/tree/main/mcp_ai_agents/github_mcp_agent):
  repository information as review context, without the API or MCP layer.
- [Scope Creep Detector](https://github.com/Shubhamsaboo/awesome-llm-apps/tree/main/agent_skills/scope-creep-detector):
  explicit rules for changes that deserve a scope check.
- [Commit Archaeologist](https://github.com/Shubhamsaboo/awesome-llm-apps/tree/main/agent_skills/commit-archaeologist):
  earlier file history as evidence for a reviewer.

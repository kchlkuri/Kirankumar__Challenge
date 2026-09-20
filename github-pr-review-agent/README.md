# GitHub PR Review Agent

A local Python CLI that reviews committed branch changes and writes a Markdown
report. Despite the name, version 1 does not connect to GitHub or post PR comments.
It uses local Git commands and the Python standard library.

## What it checks

- **Changed files:** Status and added/removed line counts.
- **Scope:** Broad changes across source directories, large diffs, and source changes
  accompanied by dependency-manifest edits.
- **Tests:** Source changes without test-file changes, tests without added lines,
  and source files without a similarly named changed test.
- **History:** Up to three commits per touched path from before the branch diverged.

The report includes a summary, findings, historical context, and review recommendations.
It is a review checklist, not an automated approval or measured test-coverage report.

## Setup

Use Python 3.10 or newer and a local Git installation. There are no Python packages
to install, credentials to configure, or services to start.

From this repository's root:

```bash
cd github-pr-review-agent
python --version
git --version
```

Use `python3` instead of `python` if needed on your machine. Run the commands below
from the `github-pr-review-agent` directory.

## Usage

The target repository must already have the review branch checked out and a local
base ref. The CLI compares committed changes only; it does not switch branches or
fetch anything.

Review the target's current branch against `main`:

```bash
python -m src.main \
  --repo "/absolute/path/to/target-repo" \
  --base main \
  --output /tmp/branch-review.md
```

If the target uses `develop` as its base, select that local ref explicitly:

```bash
python -m src.main \
  --repo "/absolute/path/to/target-repo" \
  --base develop \
  --output /tmp/develop-review.md
```

Replace the repository path with your checkout. A successful run prints
`Wrote /tmp/branch-review.md` for the first example. The output file must not already
exist; use a new filename for each run.

`--base` defaults to `main`. Always pass `--output` for your own review because the
default, `sample_output/sample_review.md`, is already included in the project.
Errors return exit code 1; review warnings do not fail the command.

## Architecture

`main.py` runs the review in one process. `git_diff.py` reads changes from the common
ancestor of the base and HEAD; `scope_check.py` applies path, size, and test-file
rules; `history.py` reads earlier commits for touched paths. `report.py` turns those
results into Markdown, while `utils.py` handles Git commands and Markdown escaping.

Git access is read-only, and the CLI writes only the requested report. It does not
run target-project code or tests, change branches, call an API, or start background work.

## Rules and limitations

- **Diff semantics:** Equivalent to the file changes in `main...HEAD`, not a comparison
  of working-tree contents. Staged, unstaged, and untracked files are excluded and
  their presence is noted. Running on `main` normally gives an empty report.
- **Scope thresholds:** Flags at least three source path areas, more than ten changed
  files, or more than 500 added-plus-removed text lines. Common container directories
  such as `src` and `services` are stripped before counting areas. Flat files form
  one root area. Manifest edits are flagged alongside source edits without claiming
  that a dependency was added.
- **Scope uncertainty:** No PR description or intent is supplied. A wide diff may be
  correct; each finding asks for review rather than declaring scope creep.
- **Test heuristics:** Uses conventional test directories and filenames for common
  source languages. Similar names are only a hint. Existing tests, integration
  coverage, assertion quality, fixtures, and language-specific coverage are not analyzed.
  More changed test lines do not prove better coverage.
- **History:** Uses only locally available history, before branch changes. Renames
  are shown as deletion plus addition, and history is not followed across old paths.
  Commit messages are not treated as proof of why a decision was made.
- **Diff detail:** Summarizes file-level statistics, not patch semantics. Binary files
  are listed without line counts. This is not a correctness or security review.

## Tests and sample report

From the project directory:

```bash
python -m unittest discover -s tests -v
```

Exactly two test methods cover the scope/test heuristics and an end-to-end CLI
report. The integration test creates an isolated temporary repository; it does not
change this repository's branches or commits.

[The sample report](sample_output/sample_review.md) was generated from a fictional
temporary repository. Its branch changes three services but updates only an auth
test, demonstrating broad scope, unmatched test filenames, and prior file history.
Commit hashes and dates belong to that fixture, not to a real pull request.

## Project files

```text
github-pr-review-agent/
  README.md
  requirements.txt
  .gitignore
  sample_output/
    sample_review.md
  src/
    main.py
    git_diff.py
    scope_check.py
    history.py
    report.py
    utils.py
  tests/
    test_scope_check.py
    test_report.py
```

## Resume bullets

- Built a standard-library Python CLI that combines Git branch diffs and per-file commit history into Markdown review reports.
- Implemented deterministic checks for cross-directory changes, oversized diffs, dependency-manifest edits, and missing test-file updates.
- Validated review rules and end-to-end report generation with two automated tests, including an isolated Git fixture and output-overwrite protection.

## References

Only these three reference paths were inspected. The implementation is original,
with no copied code and none of the reference projects' service requirements.

- [GitHub MCP Agent](https://github.com/Shubhamsaboo/awesome-llm-apps/tree/main/mcp_ai_agents/github_mcp_agent):
  repository intelligence, reduced here to local Git data without MCP.
- [Scope Creep Detector](https://github.com/Shubhamsaboo/awesome-llm-apps/tree/main/agent_skills/scope-creep-detector):
  small deterministic checks for changes that need review.
- [Commit Archaeologist](https://github.com/Shubhamsaboo/awesome-llm-apps/tree/main/agent_skills/commit-archaeologist):
  use file history as review context without inventing intent.

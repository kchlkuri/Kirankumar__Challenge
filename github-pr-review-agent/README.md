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

## Setup and run

Use Python 3.10 or newer with Git installed. No packages, credentials, model downloads,
MCP server, or hosted services are needed.

From this repository's root:

```bash
cd github-pr-review-agent
python --version
git --version
```

Check out the branch you want to review in your target repository, then run:

```bash
python -m src.main \
  --repo /absolute/path/to/target-repo \
  --base main \
  --output /tmp/branch-review.md
```

Use `python3` if that is your local Python command. The target needs a local `main`
ref and at least one commit; `--base` can select another local ref. Nothing is fetched.
The output path must not already exist, so the command cannot overwrite existing files.
If you repeat the command, choose a new report filename.

The default output path is `sample_output/sample_review.md`, relative to your current
directory. This repository includes that sample, so pass `--output` for your own review.
Errors return exit code 1. Findings do not produce a failing exit code.

## Architecture

`main.py` runs one local process:

1. `git_diff.py` resolves the base and HEAD, finds their common ancestor, and reads
   file status and line counts for the ancestor-to-HEAD diff.
2. `scope_check.py` applies path, size, and test-filename rules.
3. `history.py` reads earlier commits for each touched path at the common ancestor.
4. `report.py` formats the findings and recommendations as Markdown.
5. `utils.py` runs Git without a shell and escapes Markdown content.

The target repository is read-only. The CLI writes only the requested report, does
not execute project tests or code, and never changes branches or creates commits.
There is no database, UI, authentication layer, scheduler, or background worker.

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

- Built a local Python CLI that summarizes branch diffs and file history into structured Markdown review reports using Git and the standard library.
- Implemented review heuristics for broad change scope and missing test-file updates, with explicit limits separating warnings from measured coverage.
- Added focused heuristic tests and a temporary-repository integration test covering diff analysis, historical context, report generation, and non-overwriting output.

## References

Only these three reference paths were inspected. The implementation is original,
with no copied code and none of the reference projects' service requirements.

- [GitHub MCP Agent](https://github.com/Shubhamsaboo/awesome-llm-apps/tree/main/mcp_ai_agents/github_mcp_agent):
  repository intelligence, reduced here to local Git data without MCP.
- [Scope Creep Detector](https://github.com/Shubhamsaboo/awesome-llm-apps/tree/main/agent_skills/scope-creep-detector):
  small deterministic checks for changes that need review.
- [Commit Archaeologist](https://github.com/Shubhamsaboo/awesome-llm-apps/tree/main/agent_skills/commit-archaeologist):
  use file history as review context without inventing intent.

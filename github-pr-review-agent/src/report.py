"""Render a local branch review as Markdown."""

from .git_diff import BranchDiff
from .utils import markdown


def finding_section(title: str, findings: list[str], empty: str) -> list[str]:
    return [f"## {title}", ""] + (
        [f"- {markdown(finding)}" for finding in findings] if findings else [empty]
    ) + [""]


def generate_report(
    diff: BranchDiff,
    scope: list[str],
    tests: list[str],
    history: dict[str, list[str]],
) -> str:
    added = sum(file.added or 0 for file in diff.files)
    removed = sum(file.removed or 0 for file in diff.files)
    lines = [
        "# Local Branch Review",
        "",
        "A rule-based review of committed changes. No network or LLM calls were made.",
        "",
        "## Summary",
        "",
        f"- Changed files: {len(diff.files)}",
        f"- Text lines: +{added} / -{removed} (binary changes excluded)",
        f"- Scope findings: {len(scope)}",
        f"- Test warnings: {len(tests)}",
        f"- Base branch commit: `{diff.base_commit[:12]}`",
        f"- Common ancestor: `{diff.merge_base[:12]}`",
        f"- Reviewed HEAD: `{diff.head[:12]}`",
        "",
    ]
    if diff.dirty:
        lines.extend(["Uncommitted or untracked files were present and were not reviewed.", ""])
    lines.extend(["## Changed files", ""])
    if diff.files:
        lines.extend(["| Status | File | Added | Removed |", "| --- | --- | ---: | ---: |"])
        for file in diff.files:
            added_count = "binary" if file.added is None else str(file.added)
            removed_count = "binary" if file.removed is None else str(file.removed)
            lines.append(
                f"| {file.status} | {markdown(file.path)} | {added_count} | {removed_count} |"
            )
        lines.append("")
    else:
        lines.extend(["No committed changes relative to the common ancestor.", ""])
    lines.extend(finding_section(
        "Scope creep findings", scope,
        "No scope rules triggered. This does not establish that the change is in scope.",
    ))
    lines.extend(finding_section(
        "Test coverage warnings", tests,
        "No test-file warnings triggered. Test execution and actual coverage were not measured.",
    ))
    lines.extend([
        "## Historical context", "",
        "Up to three earlier commits per touched path, as of the common ancestor. "
        "Commit subjects are context, not verified explanations of intent.", "",
    ])
    for path, commits in history.items():
        lines.extend([f"### {markdown(path)}", ""])
        lines.extend(
            [f"- {markdown(commit)}" for commit in commits]
            or ["No earlier history found at this path."]
        )
        lines.append("")
    if not history:
        lines.extend(["No touched paths to inspect.", ""])
    recommendations = []
    if scope:
        recommendations.append("Confirm each flagged area belongs to the same change; split unrelated work.")
    if tests:
        recommendations.append("Check the test warnings and run relevant tests before merging.")
    if diff.files:
        recommendations.append("Read the actual patch and earlier commits before approving behavior changes.")
    else:
        recommendations.append("Check out the branch to review if a non-empty diff was expected.")
    lines.extend(finding_section("Review recommendations", recommendations, ""))
    lines.extend([
        "## Limits", "",
        "These are path and size heuristics, not a correctness, security, or coverage audit. "
        "No PR intent is supplied. Renames appear as deletion plus addition; history does "
        "not follow renames. Only local committed history is available.", "",
    ])
    return "\n".join(lines)

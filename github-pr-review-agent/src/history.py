"""Read recent file history from before the branch diverged."""

from pathlib import Path

from .git_diff import BranchDiff
from .utils import run_git


def file_history(repo: Path, diff: BranchDiff) -> dict[str, list[str]]:
    history = {}
    for file in diff.files:
        output = run_git(
            repo, "log", "-3", "--no-show-signature",
            "--format=%h %ad %s", "--date=short",
            diff.merge_base, "--", file.path,
        )
        history[file.path] = output.splitlines()
    return history

"""Read committed changes between a base branch and HEAD."""

from dataclasses import dataclass
from pathlib import Path

from .utils import run_git


@dataclass(frozen=True)
class ChangedFile:
    path: str
    status: str
    added: int | None
    removed: int | None


@dataclass(frozen=True)
class BranchDiff:
    base_commit: str
    merge_base: str
    head: str
    dirty: bool
    files: list[ChangedFile]


def read_diff(repo: Path, base: str = "main") -> BranchDiff:
    base_commit = run_git(repo, "rev-parse", "--verify", "--end-of-options", f"{base}^{{commit}}").strip()
    head = run_git(repo, "rev-parse", "--verify", "HEAD").strip()
    merge_base = run_git(repo, "merge-base", base_commit, head).strip()
    options = (
        "diff", "--no-ext-diff", "--no-textconv", "--no-renames",
        merge_base, head,
    )
    stats = {}
    for entry in run_git(repo, *options, "--numstat", "-z", "--").split("\0"):
        if entry:
            added, removed, path = entry.split("\t", 2)
            stats[path] = (
                None if added == "-" else int(added),
                None if removed == "-" else int(removed),
            )
    entries = run_git(repo, *options, "--name-status", "-z", "--").split("\0")
    files = []
    for index in range(0, len(entries) - 1, 2):
        status, path = entries[index:index + 2]
        added, removed = stats[path]
        files.append(ChangedFile(path, status, added, removed))
    dirty = bool(run_git(repo, "status", "--porcelain", "--untracked-files=normal").strip())
    return BranchDiff(base_commit, merge_base, head, dirty, files)

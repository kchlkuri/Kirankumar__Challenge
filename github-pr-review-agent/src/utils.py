"""Local Git execution and Markdown escaping."""

import os
import re
import subprocess
from pathlib import Path


def run_git(repo: Path, *args: str) -> str:
    environment = os.environ.copy()
    environment["GIT_LITERAL_PATHSPECS"] = "1"
    environment["GIT_TERMINAL_PROMPT"] = "0"
    result = subprocess.run(
        ["git", "--no-pager", "-C", str(repo), *args],
        capture_output=True, text=True, encoding="utf-8", errors="replace",
        env=environment, timeout=30,
    )
    if result.returncode:
        raise ValueError(result.stderr.strip() or "Git command failed.")
    return result.stdout


def markdown(text: str) -> str:
    text = " ".join(text.splitlines())
    return re.sub(r"([\\`*_\[\]<>|])", r"\\\1", text)

"""Write a review of a local branch diff."""

import argparse
import subprocess
from pathlib import Path

from .git_diff import read_diff
from .history import file_history
from .report import generate_report
from .scope_check import scope_findings, test_warnings


def main() -> None:
    cli = argparse.ArgumentParser(description="Review a local branch against main without network access.")
    cli.add_argument("--repo", type=Path, required=True, help="Local Git repository to review")
    cli.add_argument("--base", default="main", help="Local base ref (default: main)")
    cli.add_argument("--output", type=Path, default=Path("sample_output/sample_review.md"))
    args = cli.parse_args()
    try:
        diff = read_diff(args.repo, args.base)
        report = generate_report(
            diff, scope_findings(diff.files), test_warnings(diff.files),
            file_history(args.repo, diff),
        )
        if args.output.exists():
            raise ValueError(f"Output already exists: {args.output}. Choose a new --output path.")
        args.output.parent.mkdir(parents=True, exist_ok=True)
        with args.output.open("x", encoding="utf-8") as output:
            output.write(report)
    except (OSError, ValueError, subprocess.TimeoutExpired) as error:
        cli.exit(1, f"Error: {error}\n")
    print(f"Wrote {args.output}")


if __name__ == "__main__":
    main()

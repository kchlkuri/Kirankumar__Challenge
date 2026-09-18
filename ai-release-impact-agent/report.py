"""Generate an offline Markdown report; this module is also the CLI."""

import argparse
from pathlib import Path

from analyzer import Finding, analyze
from parser import parse_requirements


def generate_report(findings: list[Finding], count: int) -> str:
    lines = [
        "# AI Release Impact Agent: Offline Dependency Report",
        "",
        "This is a rule-based manifest review. No AI model or network service was used.",
        "It does not check current releases, release notes, vulnerabilities, or installed versions.",
        "",
        f"Requirement entries read: {count}",
        f"Findings: {len(findings)}",
        "",
    ]
    for finding in findings:
        lines.extend([
            f"## {finding.package}: {finding.level}",
            "",
            f"- Lines: {finding.lines}",
            f"- Finding: {finding.reason}",
            f"- Action: {finding.action}",
            "",
        ])
    if not findings:
        lines.extend(["No findings from the limited checks performed.", ""])
    lines.extend([
        "## Limits",
        "",
        "An exact pin is not proof that a dependency is safe or current.",
        "Only disagreements between simple numeric exact pins are checked for conflicts.",
        "Ranges are flagged for review, not solved or checked against exact pins.",
        "Transitive dependencies and compatibility with application code are not analyzed.",
        "",
    ])
    return "\n".join(lines)


def main() -> None:
    cli = argparse.ArgumentParser(description="Review requirements.txt without network access.")
    cli.add_argument("input", nargs="?", type=Path, default=Path("requirements.txt"))
    cli.add_argument("--output", type=Path, default=Path("output/sample_report.md"))
    args = cli.parse_args()
    try:
        if args.input.resolve() == args.output.resolve():
            raise ValueError("Input and output must be different files.")
        requirements = parse_requirements(args.input)
        report = generate_report(analyze(requirements), len(requirements))
        args.output.parent.mkdir(parents=True, exist_ok=True)
        args.output.write_text(report, encoding="utf-8")
    except (OSError, UnicodeError, ValueError) as error:
        cli.exit(1, f"Error: {error}\n")
    print(f"Wrote {args.output}")


if __name__ == "__main__":
    main()

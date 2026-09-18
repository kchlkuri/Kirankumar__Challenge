"""Describe direct dependency risks without network calls."""

from dataclasses import dataclass

from parser import Requirement


@dataclass(frozen=True)
class Finding:
    package: str
    lines: str
    level: str
    reason: str
    action: str


def version_key(version: str) -> tuple[int, ...]:
    parts = [int(part) for part in version.split(".")]
    while len(parts) > 1 and parts[-1] == 0:
        parts.pop()
    return tuple(parts)


def analyze(requirements: list[Requirement]) -> list[Finding]:
    groups: dict[str, list[Requirement]] = {}
    for requirement in requirements:
        groups.setdefault(requirement.name, []).append(requirement)

    findings = []
    for name, entries in groups.items():
        pins = {
            version_key(entry.exact_version)
            for entry in entries
            if entry.exact_version is not None
        }
        if len(pins) > 1:
            pinned_entries = [entry for entry in entries if entry.exact_version is not None]
            findings.append(Finding(
                name,
                ", ".join(str(entry.line) for entry in pinned_entries),
                "High",
                "Different exact pins for the same package cannot both be satisfied.",
                "Choose one tested version and remove the conflicting pins.",
            ))
        # Repeated requirements are combined by an installer. An exact pin
        # constrains the package even if another line is unpinned or a range.
        if pins:
            continue
        constraints = [entry.constraint for entry in entries if entry.constraint]
        lines = ", ".join(str(entry.line) for entry in entries)
        if not constraints:
            findings.append(Finding(
                name, lines, "Medium",
                "No version constraint is declared.",
                "Record a tested exact version for repeatable direct dependency selection.",
            ))
        else:
            findings.append(Finding(
                name, lines, "Review",
                "Constraints are present, but there is no single exact pin.",
                "Review the allowed versions; use a tested exact pin if reproducibility is required.",
            ))
    priority = {"High": 0, "Medium": 1, "Review": 2}
    return sorted(findings, key=lambda finding: (priority[finding.level], finding.package))

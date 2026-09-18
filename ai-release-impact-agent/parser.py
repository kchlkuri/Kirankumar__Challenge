"""Read a small, explicit subset of requirements.txt syntax."""

import re
from dataclasses import dataclass
from pathlib import Path


NAME = r"[A-Za-z0-9](?:[A-Za-z0-9._-]*[A-Za-z0-9])?"
CONSTRAINT = re.compile(r"(==|!=|>=|<=|~=|>|<)(\d+(?:\.\d+)*)(\.\*)?")


@dataclass(frozen=True)
class Requirement:
    name: str
    constraint: str
    line: int

    @property
    def exact_version(self) -> str | None:
        if re.fullmatch(r"==\d+(?:\.\d+)*", self.constraint):
            return self.constraint[2:]
        return None


def parse_requirements(path: Path) -> list[Requirement]:
    requirements = []
    for line_number, raw_line in enumerate(path.read_text(encoding="utf-8").splitlines(), 1):
        text = raw_line.split("#", 1)[0].strip()
        if not text:
            continue
        match = re.fullmatch(rf"({NAME})\s*(.*)", text)
        if not match:
            raise ValueError(f"Line {line_number}: unsupported requirement: {text}")
        name, constraint = match.groups()
        constraint = re.sub(r"\s+", "", constraint)
        if constraint:
            for part in constraint.split(","):
                version_match = CONSTRAINT.fullmatch(part)
                if not version_match:
                    raise ValueError(f"Line {line_number}: unsupported constraint: {part}")
                operator, version, wildcard = version_match.groups()
                if wildcard and operator not in ("==", "!="):
                    raise ValueError(f"Line {line_number}: wildcard requires == or !=")
                if operator == "~=" and "." not in version:
                    raise ValueError(f"Line {line_number}: ~= needs at least two version parts")
        name = re.sub(r"[-_.]+", "-", name).lower()
        requirements.append(Requirement(name, constraint, line_number))
    return requirements

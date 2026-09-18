# AI Release Impact Agent: Offline Dependency Report

This offline, rule-based review found three dependency-risk findings across five
requirement entries. No AI model or network service was used, and no input files
were changed.

## Summary

| Measure | Result |
| --- | --- |
| Input | `requirements.txt` |
| Requirement entries read | 5 |
| Total findings | 3 |
| High | 1 |
| Medium | 1 |
| Review | 1 |

## Findings

### Conflicting exact pins

- **Package:** `pydantic`
- **Priority:** High
- **Input lines:** 6, 7
- **Finding:** Different exact pins for the same package cannot both be satisfied.
- **Action:** Choose one tested version and remove the conflicting pins.

### Unconstrained dependency

- **Package:** `flask`
- **Priority:** Medium
- **Input line:** 4
- **Finding:** No version constraint is declared.
- **Action:** Record a tested exact version for repeatable direct dependency selection.

### Version range to review

- **Package:** `fastapi`
- **Priority:** Review
- **Input line:** 5
- **Finding:** Constraints are present, but there is no single exact pin.
- **Action:** Review the allowed versions; use a tested exact pin if reproducibility is required.

## Limits

An exact pin is not proof that a dependency is safe or current.
A review finding is not a claim that a version range is invalid.

- **Conflict coverage:** Only disagreements between simple numeric exact pins are checked.
- **Version ranges:** Flagged for review, not solved or checked against exact pins.
- **Release and security coverage:** Current releases, release notes, vulnerabilities, and installed versions are not checked.
- **Compatibility coverage:** Transitive dependencies and compatibility with application code are not analyzed.

## Reproducing the findings

Run `python3 -B report.py` from `ai-release-impact-agent/` to regenerate this report.
The findings stay the same for the supplied sample input; the CLI overwrites this
presentation-edited example with its original, simpler Markdown layout.

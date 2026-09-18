# ai-release-impact-agent

A small, fully offline Python CLI that turns `requirements.txt` into an actionable
Markdown dependency-risk report. It flags conflicting exact pins, unconstrained
dependencies, and version ranges that need review.

The implementation uses only the Python standard library. Despite the project
name, it is rule-based: it does not use AI or fetch actual releases.

## Architecture

The tool runs as a single local process with three small modules. The CLI entry
point lives in `report.py`; there is no separate service or orchestration layer.

```text
python3 -B report.py
        |
        v
requirements.txt
        |
        v
parser.py       Validate supported syntax and normalize package names
        |       Return Requirement records with source line numbers
        v
analyzer.py     Group requirements and apply dependency-risk rules
        |       Return findings ordered by priority, then package name
        v
report.py      Format findings and write the Markdown report
        |
        v
output/sample_report.md
```

The parser rejects unsupported syntax rather than skipping entries. The analyzer
does not modify the input, and the report generator writes only the selected
output file. All processing stays local, with no network calls or third-party
runtime dependencies.

## Run

Use Python 3.10 or newer. Everything uses the standard library; there is nothing
to install. The included `requirements.txt` is sample data, not tool dependencies.

From the repository root:

```bash
cd ai-release-impact-agent
python3 -B report.py
```

The command prints `Wrote output/sample_report.md`. Paths are relative to your
current directory. To review another file:

```bash
python3 -B report.py /path/to/requirements.txt --output output/report.md
```

The output directory is created when needed. An existing report is overwritten,
but the input file cannot also be the output file. Invalid input returns exit
code 1 with a line-numbered error; findings themselves do not cause a failing exit.

## Sample report

The included sample has five requirement entries and produces three findings.
See the [formatted sample report](output/sample_report.md) for the full explanation
and suggested actions.

| Priority | Package | Finding |
| --- | --- | --- |
| High | `pydantic` | Two different exact pins |
| Medium | `flask` | No version constraint |
| Review | `fastapi` | Version range without a single exact pin |

The checked-in sample report has a presentation-only formatting pass. Running
the CLI regenerates the same findings in the generator's original, simpler layout.

## Checks

- **High:** different numeric exact pins for the same normalized package name.
- **Medium:** a package has no version constraint on any of its entries.
- **Review:** constraints exist, but there is no single exact pin.

Repeated entries are grouped by package name. Case and runs of dots, underscores,
and hyphens are normalized. Numeric pins such as `1.0` and `1.0.0` compare equally.
If a package has an exact pin, another unpinned or ranged entry does not generate
a version-drift finding. Range compatibility with that pin is not evaluated.

An exact pin is not a safety guarantee. This tool does not resolve dependencies,
detect vulnerabilities, check current releases, analyze transitive dependencies,
or prove compatibility with application code. A range may be intentional; a
review finding is not a claim that the range is invalid.

## Input format

Supports blank lines, `#` comments, plain package names, numeric release versions,
and comma-separated constraints using `==`, `!=`, `>=`, `<=`, `>`, `<`, and `~=`.
Numeric versions can have multiple parts, such as `1`, `1.2`, or `1.2.3`.
Trailing wildcards are accepted only with `==` and `!=`; `~=` needs at least
two numeric parts.

This is deliberately not a full pip requirements parser. Extras, environment
markers, URLs, editable installs, includes, hashes, line continuations, prereleases,
post/dev/local versions, and epochs are rejected, not silently ignored.
Only simple standalone exact pins are compared for conflicts. Complex constraint
combinations are left for review.

## Tests

Run from this directory:

```bash
python3 -B -m unittest -v
```

Exactly two test methods cover parsing and rejected syntax, then dependency-risk
analysis and Markdown generation. Tests use temporary local files and no network.

## Resume bullets

- Built a fully offline Python CLI that converts `requirements.txt` into prioritized Markdown dependency-risk reports using only the standard library.
- Implemented requirement parsing with package-name normalization, line-numbered validation errors, and rule-based checks for conflicting exact pins, unpinned dependencies, and version ranges.
- Separated parsing, analysis, and reporting into three focused modules and validated the workflow with two automated tests covering supported inputs, rejected syntax, risk classification, and Markdown output.

## Project files

```text
parser.py                  Read and validate the supported requirement syntax
analyzer.py                Find direct manifest risks
report.py                  Generate Markdown and provide the CLI
test_agent.py              Two basic tests
requirements.txt           Sample input
output/sample_report.md    Generated findings with presentation-only formatting
README.md                  Usage and limitations
```

## References

Original implementation, inspired only by the manifest-to-brief idea in
[Release Radar Agent](https://github.com/Shubhamsaboo/awesome-llm-apps/tree/main/always_on_agents/release_radar_agent)
and the offline direct-manifest checks in
[Dependency Doctor](https://github.com/Shubhamsaboo/awesome-llm-apps/tree/main/agent_skills/dependency-doctor).
No reference code was copied. No UI, Docker, cloud service, database, API, LLM,
scheduler, or third-party dependency is included.

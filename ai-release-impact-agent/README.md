# ai-release-impact-agent

A small offline Python CLI that reads `requirements.txt` and writes a dependency-risk
report to `output/sample_report.md`. Despite the project name, it uses simple rules,
not an AI model. It does not fetch or assess actual releases.

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

## Files

```text
parser.py                  Read and validate the supported requirement syntax
analyzer.py                Find direct manifest risks
report.py                  Generate Markdown and provide the CLI
test_agent.py              Two basic tests
requirements.txt           Sample input
output/sample_report.md    Generated sample report
README.md                  Usage and limitations
```

## References

Original implementation, inspired only by the manifest-to-brief idea in
[Release Radar Agent](https://github.com/Shubhamsaboo/awesome-llm-apps/tree/main/always_on_agents/release_radar_agent)
and the offline direct-manifest checks in
[Dependency Doctor](https://github.com/Shubhamsaboo/awesome-llm-apps/tree/main/agent_skills/dependency-doctor).
No reference code was copied. No UI, Docker, cloud service, database, API, LLM,
scheduler, or third-party dependency is included.

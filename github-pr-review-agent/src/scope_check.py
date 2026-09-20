"""Flag broad diffs and test-file gaps using path-based rules."""

from pathlib import PurePosixPath

from .git_diff import ChangedFile


SOURCE_SUFFIXES = {
    ".py", ".go", ".java", ".js", ".jsx", ".ts", ".tsx",
    ".rs", ".c", ".cpp", ".h", ".cs", ".kt",
}
MANIFESTS = {"requirements.txt", "pyproject.toml", "package.json", "go.mod", "Cargo.toml"}


def is_test(path: str) -> bool:
    file = PurePosixPath(path)
    name = file.name.lower()
    return (
        bool({"test", "tests", "__tests__"} & set(file.parts))
        or name.startswith("test_")
        or file.stem.endswith(("_test", ".test", ".spec", "Test", "Tests"))
    )


def source_files(files: list[ChangedFile]) -> list[ChangedFile]:
    return [
        file for file in files
        if PurePosixPath(file.path).suffix in SOURCE_SUFFIXES and not is_test(file.path)
    ]


def area(path: str) -> str:
    parts = list(PurePosixPath(path).parts)
    if parts[0] in {"src", "lib", "app", "services", "packages"}:
        parts = parts[1:]
    return parts[0] if len(parts) > 1 else "(root)"


def scope_findings(files: list[ChangedFile]) -> list[str]:
    findings = []
    areas = sorted({area(file.path) for file in source_files(files)})
    if len(areas) >= 3:
        findings.append(
            f"Source changes span {len(areas)} path areas: {', '.join(areas)}. "
            "Check whether they belong to one change."
        )
    if len(files) > 10:
        findings.append(f"{len(files)} files changed. Consider whether the review can be split.")
    lines = sum((file.added or 0) + (file.removed or 0) for file in files)
    if lines > 500:
        findings.append(f"{lines} text lines changed. Separate mechanical edits if possible.")
    manifests = [file.path for file in files if PurePosixPath(file.path).name in MANIFESTS]
    if manifests and source_files(files):
        findings.append(
            "Source and dependency manifests changed together: "
            + ", ".join(manifests) + ". Confirm the manifest edits are necessary."
        )
    return findings


def test_name(path: str) -> str:
    stem = PurePosixPath(path).stem.lower().removeprefix("test_")
    for suffix in ("_test", ".test", ".spec", "tests", "test"):
        if stem.endswith(suffix):
            return stem[:-len(suffix)]
    return stem


def test_warnings(files: list[ChangedFile]) -> list[str]:
    sources = source_files(files)
    tests = [file for file in files if is_test(file.path) and file.status != "D"]
    if not sources:
        return []
    if not tests:
        return [
            f"{len(sources)} source file(s) changed without accompanying test-file changes. "
            "Existing tests may cover the behavior; verify that explicitly."
        ]
    warnings = []
    if not any((file.added or 0) > 0 for file in tests):
        warnings.append("Test files changed, but no added test lines were detected.")
    names = {test_name(file.path) for file in tests}
    for file in sources:
        if test_name(file.path) not in names:
            warnings.append(
                f"No similarly named test changed for {file.path}. "
                "Check for integration coverage or add a focused test."
            )
    return warnings

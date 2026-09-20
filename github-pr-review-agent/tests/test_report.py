import subprocess
import sys
import tempfile
import unittest
from pathlib import Path

from src.git_diff import read_diff
from src.utils import run_git


def create_sample_repo(repo: Path) -> None:
    """Create fictional commits for the integration test and sample report."""
    repo.mkdir()
    run_git(repo, "init", "-b", "main")
    run_git(repo, "config", "user.name", "Example Developer")
    run_git(repo, "config", "user.email", "example@example.invalid")
    for name in ("auth", "payments", "orders"):
        path = repo / "services" / name / f"{name}.py"
        path.parent.mkdir(parents=True)
        path.write_text(f"def {name}_enabled():\n    return False\n", encoding="utf-8")
    tests = repo / "tests"
    tests.mkdir()
    (tests / "test_auth.py").write_text(
        "def test_auth_disabled():\n    assert True\n", encoding="utf-8",
    )
    run_git(repo, "add", ".")
    run_git(repo, "commit", "-m", "Add fictional service stubs")
    auth = repo / "services/auth/auth.py"
    auth.write_text("def auth_enabled():\n    return bool(0)\n", encoding="utf-8")
    run_git(repo, "add", ".")
    run_git(repo, "commit", "-m", "Clarify the default auth state")
    run_git(repo, "checkout", "-b", "review-demo")
    for name in ("auth", "payments", "orders"):
        (repo / "services" / name / f"{name}.py").write_text(
            f"def {name}_enabled():\n    return True\n", encoding="utf-8",
        )
    (tests / "test_auth.py").write_text(
        "from services.auth.auth import auth_enabled\n\n"
        "def test_auth_enabled():\n    assert auth_enabled()\n", encoding="utf-8",
    )
    run_git(repo, "add", ".")
    run_git(repo, "commit", "-m", "Enable three services with an auth-only test")


class ReportTests(unittest.TestCase):
    def test_cli_report_uses_branch_diff_and_prior_history(self):
        project = Path(__file__).resolve().parents[1]
        with tempfile.TemporaryDirectory() as directory:
            repo = Path(directory) / "sample-repo"
            create_sample_repo(repo)
            original_head = run_git(repo, "rev-parse", "HEAD")
            self.assertEqual(read_diff(repo).files[0].path, "services/auth/auth.py")
            (repo / "untracked.txt").write_text("Not reviewed.", encoding="utf-8")
            output = Path(directory) / "review.md"
            command = [
                sys.executable, "-B", "-m", "src.main", "--repo", str(repo),
                "--output", str(output),
            ]
            result = subprocess.run(command, cwd=project, capture_output=True, text=True)
            self.assertEqual(result.returncode, 0, result.stderr)
            report = output.read_text(encoding="utf-8")
            for heading in ("Summary", "Changed files", "Scope creep findings",
                            "Test coverage warnings", "Historical context", "Review recommendations"):
                self.assertIn(f"## {heading}", report)
            self.assertIn("Changed files: 4", report)
            self.assertIn("3 path areas", report)
            self.assertIn("Test warnings: 2", report)
            self.assertIn("Clarify the default auth state", report)
            self.assertNotIn("Enable three services with an auth-only test", report)
            self.assertIn("were not reviewed", report)
            self.assertEqual(run_git(repo, "rev-parse", "HEAD"), original_head)
            retry = subprocess.run(command, cwd=project, capture_output=True, text=True)
            self.assertNotEqual(retry.returncode, 0)
            self.assertEqual(output.read_text(encoding="utf-8"), report)

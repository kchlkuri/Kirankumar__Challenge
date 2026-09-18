import tempfile
import unittest
from pathlib import Path

from analyzer import analyze
from parser import Requirement, parse_requirements
from report import generate_report


class AgentTests(unittest.TestCase):
    def test_parser_reads_requirements_and_rejects_unsupported_syntax(self):
        with tempfile.TemporaryDirectory() as directory:
            path = Path(directory) / "requirements.txt"
            path.write_text(
                "# Sample\n\nMy_Package == 1.2.0 # tested\nflask\nfastapi>=0.100,<1\n",
                encoding="utf-8",
            )
            self.assertEqual(parse_requirements(path), [
                Requirement("my-package", "==1.2.0", 3),
                Requirement("flask", "", 4),
                Requirement("fastapi", ">=0.100,<1", 5),
            ])
            for text in ("-r other.txt", "requests[security]", "pkg>=1.*", "pkg~=1",
                         "pkg==1.0; python_version<'3.12'", "pkg==banana"):
                with self.subTest(text=text):
                    path.write_text(text, encoding="utf-8")
                    with self.assertRaisesRegex(ValueError, "Line 1:"):
                        parse_requirements(path)

    def test_analysis_and_report_cover_manifest_risks(self):
        with tempfile.TemporaryDirectory() as directory:
            path = Path(directory) / "requirements.txt"
            path.write_text(
                "requests==2.32.3\nflask\nfastapi>=0.100,<1\n"
                "pydantic==2.8.0\npydantic==2.9.0\n"
                "same==1.0\nsame==1.0.0\nsame\nwild==1.*\n",
                encoding="utf-8",
            )
            requirements = parse_requirements(path)
            findings = analyze(requirements)
            self.assertEqual(
                [(item.package, item.level) for item in findings],
                [("pydantic", "High"), ("flask", "Medium"),
                 ("fastapi", "Review"), ("wild", "Review")],
            )
            self.assertEqual(findings[0].lines, "4, 5")
            report = generate_report(findings, len(requirements))
            self.assertIn("Requirement entries read: 9", report)
            self.assertIn("Findings: 4", report)
            self.assertIn("## pydantic: High", report)
            self.assertIn("No AI model or network service was used.", report)
            self.assertIn("No findings", generate_report(analyze([]), 0))


if __name__ == "__main__":
    unittest.main()

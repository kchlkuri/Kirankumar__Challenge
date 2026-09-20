import unittest

from src.git_diff import ChangedFile
from src.scope_check import scope_findings, test_warnings


class ScopeTests(unittest.TestCase):
    def test_scope_and_test_file_heuristics(self):
        files = [
            ChangedFile("services/auth/auth.py", "M", 2, 1),
            ChangedFile("services/payments/payments.py", "M", 2, 1),
            ChangedFile("services/orders/orders.py", "M", 2, 1),
        ]
        self.assertIn("3 path areas", scope_findings(files)[0])
        self.assertIn("without accompanying", test_warnings(files)[0])
        files.append(ChangedFile("tests/test_auth.py", "M", 2, 0))
        warnings = test_warnings(files)
        self.assertEqual(len(warnings), 2)
        self.assertTrue(any("payments.py" in warning for warning in warnings))
        self.assertTrue(any("orders.py" in warning for warning in warnings))
        self.assertEqual(test_warnings([ChangedFile("README.md", "M", 5, 1)]), [])
        self.assertTrue(scope_findings([ChangedFile("large.py", "M", 501, 0)]))
        self.assertTrue(test_warnings([
            ChangedFile("auth.py", "M", 1, 1),
            ChangedFile("tests/test_auth.py", "D", 0, 4),
        ]))

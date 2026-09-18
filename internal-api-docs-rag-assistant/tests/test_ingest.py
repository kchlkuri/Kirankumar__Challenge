import tempfile
import unittest
from pathlib import Path

from src.ingest import load_documents


class IngestTests(unittest.TestCase):
    def test_ingestion_preserves_sections_and_bounds_chunks(self):
        with tempfile.TemporaryDirectory() as directory:
            docs = Path(directory)
            (docs / "auth.md").write_text(
                "# Auth\n## Tokens\none two three four five six seven\n"
                "## Errors\nMissing tokens return 401.\n",
                encoding="utf-8",
            )
            (docs / "notes.txt").write_text("Plain text works too.", encoding="utf-8")
            (docs / "skip.csv").write_text("not,documentation", encoding="utf-8")
            chunks = load_documents(docs, max_words=4)
            self.assertEqual(len(chunks), 4)
            self.assertEqual(chunks[0].section, "Auth > Tokens")
            self.assertEqual(chunks[1].section, "Auth > Tokens")
            self.assertEqual(chunks[2].section, "Auth > Errors")
            self.assertEqual(chunks[3].section, "Document")
            self.assertEqual(chunks[3].filename, "notes.txt")
            self.assertTrue(all(len(chunk.text.split()) <= 4 for chunk in chunks))
            self.assertEqual(
                " ".join(chunk.text for chunk in chunks[:2]),
                "one two three four five six seven",
            )

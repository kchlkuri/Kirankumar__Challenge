import unittest
from pathlib import Path

from src.answer import answer_question
from src.ingest import load_documents
from src.retriever import retrieve


class RetrieverTests(unittest.TestCase):
    def test_retrieval_citations_and_weak_evidence_refusal(self):
        docs = Path(__file__).resolve().parents[1] / "docs"
        chunks = load_documents(docs)
        cases = [
            ("How does service authentication work?", "auth.md", "Service authentication"),
            ("How are duplicate payment requests handled?", "payments.md", "Duplicate payment requests"),
            ("How can I cancel an order?", "orders.md", "Order cancellation"),
        ]
        for question, filename, section in cases:
            with self.subTest(question=question):
                matches = retrieve(question, chunks)
                self.assertEqual(matches[0].chunk.filename, filename)
                answer = answer_question(question, chunks)
                self.assertIn(filename, answer)
                self.assertIn(section, answer)
                quoted = "\n".join(
                    line[2:] for line in answer.splitlines() if line.startswith("> ")
                )
                self.assertEqual(quoted, matches[0].chunk.text)
        for question in (
            "What is the payments data retention period?",
            "How does service authentication use biometrics?",
            "What is tomorrow's weather?",
            "authentication",
            "",
        ):
            with self.subTest(question=question):
                self.assertEqual(answer_question(question, chunks), "I don't know")
        self.assertEqual(answer_question("service authentication", []), "I don't know")

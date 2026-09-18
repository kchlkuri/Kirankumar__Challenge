"""Return a cited excerpt or decline when retrieval has no evidence."""

from .ingest import Chunk
from .retriever import retrieve


def answer_question(question: str, chunks: list[Chunk]) -> str:
    matches = retrieve(question, chunks)
    if not matches:
        return "I don't know"
    chunk = matches[0].chunk
    excerpt = "\n".join("> " + line for line in chunk.text.splitlines())
    return (
        f"Relevant documentation:\n\n{excerpt}\n\n"
        f"Source: [{chunk.filename} :: {chunk.section}]"
    )

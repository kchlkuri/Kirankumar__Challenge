"""Ask one question against the project's local docs."""

import argparse
from pathlib import Path

from .answer import answer_question
from .ingest import load_documents


def main() -> None:
    cli = argparse.ArgumentParser(description="Search local API docs and return a cited excerpt.")
    cli.add_argument("--question", required=True, help="Question to ask the local documentation")
    args = cli.parse_args()
    if not args.question.strip():
        cli.error("Question must not be empty.")
    directory = Path(__file__).resolve().parents[1] / "docs"
    try:
        chunks = load_documents(directory)
        if not chunks:
            raise ValueError("No non-empty Markdown or text documents found in docs/.")
    except (OSError, UnicodeError, ValueError) as error:
        cli.exit(1, f"Error: {error}\n")
    print(answer_question(args.question, chunks))


if __name__ == "__main__":
    main()

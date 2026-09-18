"""Read local documents into section-aware chunks."""

import re
from dataclasses import dataclass
from pathlib import Path


@dataclass(frozen=True)
class Chunk:
    filename: str
    section: str
    text: str


def split_sections(text: str, markdown: bool) -> list[tuple[str, str]]:
    sections = []
    headings: list[tuple[int, str]] = []
    lines: list[str] = []
    section = "Document"
    fence = ""
    for line in text.splitlines():
        stripped = line.lstrip()
        if markdown and stripped.startswith(("```", "~~~")):
            marker = stripped[:3]
            if not fence:
                fence = marker
            elif fence == marker:
                fence = ""
        heading = re.match(r"^(#{1,6})\s+(.+?)\s*#*\s*$", line)
        if markdown and heading and not fence:
            sections.append((section, "\n".join(lines).strip()))
            level = len(heading[1])
            headings = [(depth, title) for depth, title in headings if depth < level]
            headings.append((level, heading[2]))
            section = " > ".join(title for _, title in headings)
            lines = []
        else:
            lines.append(line)
    sections.append((section, "\n".join(lines).strip()))
    return [(title, body) for title, body in sections if body]


def chunk_text(text: str, max_words: int) -> list[str]:
    words = list(re.finditer(r"\S+", text))
    chunks = []
    for start in range(0, len(words), max_words):
        end = min(start + max_words, len(words)) - 1
        chunks.append(text[words[start].start():words[end].end()])
    return chunks


def load_documents(directory: Path, max_words: int = 140) -> list[Chunk]:
    if max_words < 1:
        raise ValueError("Chunk size must be at least one word.")
    chunks = []
    for path in sorted(directory.iterdir()):
        if not path.is_file() or path.is_symlink() or path.suffix.lower() not in {".md", ".txt"}:
            continue
        text = path.read_text(encoding="utf-8")
        for section, body in split_sections(text, path.suffix.lower() == ".md"):
            chunks.extend(
                Chunk(path.name, section, part)
                for part in chunk_text(body, max_words)
            )
    return chunks

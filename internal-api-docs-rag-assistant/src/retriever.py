"""Rank chunks with local term-frequency cosine similarity."""

import math
from collections import Counter
from dataclasses import dataclass

from .ingest import Chunk
from .utils import tokenize


@dataclass(frozen=True)
class Match:
    chunk: Chunk
    score: float


def cosine_score(left: Counter, right: Counter) -> float:
    numerator = sum(count * right[term] for term, count in left.items())
    denominator = math.sqrt(
        sum(count * count for count in left.values())
        * sum(count * count for count in right.values())
    )
    return numerator / denominator if denominator else 0.0


def retrieve(question: str, chunks: list[Chunk]) -> list[Match]:
    query = Counter(tokenize(question))
    if len(query) < 2:
        return []
    matches = []
    for chunk in chunks:
        terms = Counter(tokenize(chunk.section + "\n" + chunk.text))
        if not set(query).issubset(terms):
            continue
        score = cosine_score(query, terms)
        if score >= 0.12:
            matches.append(Match(chunk, score))
    return sorted(
        matches,
        key=lambda match: (-match.score, match.chunk.filename, match.chunk.section),
    )[:3]

"""Text normalization shared by retrieval."""

import re


STOP_WORDS = frozenset(
    "a an the and or to of in on for with from by at is are was were be "
    "do does did how what which when where who why can could should would "
    "i we you our my it its this that these those use used using work works "
    "handled please tell me about".split()
)


def tokenize(text: str) -> list[str]:
    return [
        word for word in re.findall(r"[a-z0-9]+", text.lower())
        if word not in STOP_WORDS
    ]

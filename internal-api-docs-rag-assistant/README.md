# Internal API Docs RAG Assistant

A fully local Python CLI that answers questions by retrieving excerpts from Markdown
and text files. Every answer includes the source filename and section; weak matches
return `I don't know`.

Version 1 uses the Python standard library only. It demonstrates a retrieval-grounded
question-answering pipeline, not LLM generation: answers are document excerpts, not
model-written summaries. All sample API documentation is fictional.

## Problem statement

API behavior is often scattered across documents: authentication headers, payment
idempotency, and order-state rules. This project provides a small way to find the
relevant passage and check its source without uploading documentation to a service.

The goal is a readable local baseline, not a production search platform. There are
no network calls, model downloads, API keys, or external services.

## Architecture

```text
docs/*.md and docs/*.txt
          |
          v
ingest.py       Read UTF-8 files, track headings, split into bounded chunks
          |
          v
retriever.py    Filter by query-term coverage, rank by cosine similarity
          |
          v
answer.py       Quote the best chunk with a citation, or say "I don't know"
          |
          v
main.py         Print the answer to the terminal
```

- **Ingestion:** Reads direct `.md` and `.txt` files in `docs/`, skipping symlinks and
  other formats. Markdown hash headings form section paths; plain text uses `Document`.
  Each section is split into non-overlapping chunks of at most 140 whitespace-delimited words.
- **Retrieval:** Lowercases text, removes a small list of common words, and builds
  term-frequency counters. A candidate must contain every remaining query term in
  its section heading or body. Queries need at least two distinct meaningful terms.
  Candidates with cosine similarity below 0.12 are discarded; at most three are returned.
- **Answering:** Quotes only the highest-ranked chunk and includes a citation such
  as `[auth.md :: Authentication > Service authentication]`. No new claims are generated.
- **Execution:** `main.py` runs the pipeline once and exits. `utils.py` holds shared
  tokenization. The in-memory index is rebuilt on every invocation; nothing runs in
  the background.

The coverage rule and score cutoff are conservative heuristics, not calibrated
confidence measures. They favor refusal over loosely related matches.

## Setup

Use Python 3.10 or newer. From the repository root:

```bash
cd internal-api-docs-rag-assistant
python --version
```

There are no dependencies to install. `requirements.txt` documents that choice;
use `python3` instead of `python` if that is your local Python command.

## Usage

Run from the project directory:

```bash
python -m src.main --question "How does service authentication work?"
python -m src.main --question "How are duplicate payment requests handled?"
python -m src.main --question "What is the payments data retention period?"
```

The first two questions return cited excerpts. The third asks for information not
covered by the sample docs and returns `I don't know`; it must not confuse payment
data retention with the documented idempotency-key lifetime.

Add local UTF-8 `.md` or `.txt` files directly to `docs/` to change the corpus.
The CLI does not modify documents or save answers automatically. Missing or unreadable
docs produce an error and exit code 1; an empty question produces an argument error.
A normal answer or refusal exits successfully.

See [the three captured CLI runs](output/sample_queries.md) for sample output.

## Tests

```bash
python -m unittest discover -s tests -v
```

There are exactly two test methods:

- **Ingestion:** Checks Markdown section paths, plain-text support, ignored file types,
  word limits, and preservation of text across chunk boundaries.
- **Retrieval:** Checks all three sample document topics, filename-and-section
  citations, exact excerpt grounding, and refusal for unrelated or weakly supported queries.

Both tests run locally without packages or credentials.

## Project files

```text
internal-api-docs-rag-assistant/
  README.md
  requirements.txt
  .gitignore
  docs/
    auth.md
    payments.md
    orders.md
  src/
    main.py
    ingest.py
    retriever.py
    answer.py
    utils.py
  tests/
    test_ingest.py
    test_retriever.py
  output/
    sample_queries.md
```

## Limitations

- **Lexical matching:** No embeddings, stemming, or synonym handling. Useful paraphrases,
  single-keyword questions, and facts split across chunks may be refused.
- **Evidence strength:** Matching all terms does not prove a passage answers a question.
  The gate can still accept a misleading match; quoting a source does not validate its accuracy.
- **Answer format:** Returns one excerpt, not a synthesized explanation. There is no
  multi-document reasoning, contradiction detection, or conversation history.
- **Document parsing:** Supports hash-style Markdown headings, not full Markdown syntax.
  Long passages can split mid-sentence or inside a code block. Text files use a single section.
- **Scale:** Reads the whole corpus on each run. Intended for small local documentation sets,
  not large repositories, PDFs, or nested folder trees.
- **Security:** Fictional authentication docs are sample content, not an authentication
  implementation. The CLI has no access-control layer and relies on local file permissions.

## Future improvements

These are possible follow-ups, not implemented features. Version 1 intentionally
stops at the local, single-process pipeline above.

- **Evaluation:** Add a labeled question set before adjusting the refusal threshold.
- **Chunking:** Preserve full sentences and code blocks when splitting longer sections.
- **Retrieval:** Evaluate local TF-IDF weighting against this term-frequency baseline.

## Resume bullets

- Built a fully local Python documentation assistant with section-aware ingestion, bounded text chunking, and cosine-ranked retrieval using only the standard library.
- Implemented evidence-gated answers with filename-and-section citations and verbatim source excerpts, returning explicit refusals when keyword coverage or similarity was insufficient.
- Validated ingestion, retrieval, source grounding, and weak-evidence refusal with two automated tests, and captured three reproducible CLI examples against fictional internal API documentation.

## Reference ideas

Only the following four folders in `awesome-llm-apps` were inspected. This is an
original implementation; no reference code was copied.

- [RAG database routing](https://github.com/Shubhamsaboo/awesome-llm-apps/tree/main/rag_tutorials/rag_database_routing):
  use relevant context, reduced here to one local corpus with no router or fallback web search.
- [Typed RAG with Pydantic AI](https://github.com/Shubhamsaboo/awesome-llm-apps/tree/main/rag_tutorials/agentic_typed_rag_pydanticai):
  carry source metadata and reject weak retrieval before answering.
- [Knowledge graph RAG citations](https://github.com/Shubhamsaboo/awesome-llm-apps/tree/main/rag_tutorials/knowledge_graph_rag_citations):
  make answers traceable to documents, without adopting a graph database.
- [RAG failure diagnostics clinic](https://github.com/Shubhamsaboo/awesome-llm-apps/tree/main/rag_tutorials/rag_failure_diagnostics_clinic):
  treat grounding and chunk-boundary failures as explicit limitations.

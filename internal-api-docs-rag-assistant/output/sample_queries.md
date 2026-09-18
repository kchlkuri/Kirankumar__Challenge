# Internal API Docs RAG Assistant: Sample Queries

These outputs were captured from three local CLI runs against the included fictional
API documentation. No model, network service, or external database was used.

## Service authentication

Command, run from `internal-api-docs-rag-assistant/`:

```bash
python -B -m src.main --question "How does service authentication work?"
```

Output:

```text
Relevant documentation:

> Service authentication uses a bearer token issued by the fictional Atlas identity service.
> Send the token in the Authorization header as `Bearer <token>` on every internal API request.
> Tokens expire after 15 minutes. The calling service must obtain a new token before expiry.
> Never put tokens in query parameters or application logs.

Source: [auth.md :: Authentication > Service authentication]
```

## Duplicate payment requests

Command:

```bash
python -B -m src.main --question "How are duplicate payment requests handled?"
```

Output:

```text
Relevant documentation:

> Duplicate payment requests are handled with an `Idempotency-Key` header on `POST /payments`.
> The fictional Atlas payments service stores the key and original response for 24 hours.
> Repeating the same key and request body returns the original response without creating another payment.
> Reusing the key with a different request body returns HTTP 409 with the code `IDEMPOTENCY_CONFLICT`.

Source: [payments.md :: Payments API > Duplicate payment requests]
```

## Unsupported payment-retention question

Command:

```bash
python -B -m src.main --question "What is the payments data retention period?"
```

Output:

```text
I don't know
```

The docs describe the lifetime of an idempotency key, not a payment-data retention
policy. The question fails the retrieval coverage check, so the assistant refuses
instead of returning a loosely related passage.

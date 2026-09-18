# Payments API

## Duplicate payment requests

Duplicate payment requests are handled with an `Idempotency-Key` header on `POST /payments`.
The fictional Atlas payments service stores the key and original response for 24 hours.
Repeating the same key and request body returns the original response without creating another payment.
Reusing the key with a different request body returns HTTP 409 with the code `IDEMPOTENCY_CONFLICT`.

## Payment status

Call `GET /payments/{payment_id}` to read the payment status.
Supported states are `pending`, `succeeded`, and `failed`.
The response includes `payment_id`, `status`, `amount`, and `currency`.

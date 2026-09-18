# Authentication

## Service authentication

Service authentication uses a bearer token issued by the fictional Atlas identity service.
Send the token in the Authorization header as `Bearer <token>` on every internal API request.
Tokens expire after 15 minutes. The calling service must obtain a new token before expiry.
Never put tokens in query parameters or application logs.

## Authentication errors

A missing or expired bearer token returns HTTP 401 with the code `AUTH_TOKEN_INVALID`.
A valid token without the required scope returns HTTP 403 with the code `AUTH_SCOPE_MISSING`.
The response includes a request ID for troubleshooting.

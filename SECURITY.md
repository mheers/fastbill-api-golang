# Security Policy

## Supported versions

Only the latest release receives security fixes.

## Reporting a vulnerability

Please do not open a public issue for a security problem. Send a description
and, if possible, a reproduction to **marcel@heers.it**. You will get an
acknowledgement as soon as possible, and the fix and disclosure will be
coordinated with you.

## Handling credentials

FastBill uses HTTP Basic authentication with the account email and API key.
Never put credentials in code, logs, issues or test fixtures. Load them from
the environment or a secret store.

The client never logs request bodies or credentials. `APIError` and
`StatusError` contain only the service name, FastBill's error messages and the
truncated HTTP body.

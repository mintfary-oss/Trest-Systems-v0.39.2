# Stage 2 — Authentication and Roles

Stage 2 adds a self-hosted authentication foundation without a third-party identity provider.

## Endpoints

- `POST /api/v1/auth/register` — create customer, contractor, or supplier account.
- `POST /api/v1/auth/login` — issue a signed 24-hour bearer token.
- `GET /api/v1/auth/me` — authenticated profile endpoint.

## Security

Passwords are stored using PBKDF2-HMAC-SHA-256 with a per-user random salt. Access tokens are HMAC-SHA-256 signed and expire after 24 hours. The signing secret must be supplied from configuration before production use.

## Roles

`customer`, `contractor`, `supplier`, `admin` remain the canonical roles. Public registration cannot create an `admin` account.

## Production requirement

Set a strong `TREST_AUTH_SECRET` and rotate it according to the deployment security policy. This stage intentionally does not add external OAuth/SSO.

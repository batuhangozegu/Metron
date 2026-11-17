# Accounts and sign-in backend

The Go API exclusively writes `app`; market data ownership remains with the engine. Run `task migrate` before using auth. Migration histories are `public.goose_market_version` and `public.goose_app_version`. Application times are UTC `timestamptz`. Responses use UTC timestamps and decimal strings.

Configure `APP_ENCRYPTION_KEY` as base64 of exactly 32 random bytes (`openssl rand -base64 32`). Missing/invalid keys log a startup warning and make all auth, `/v1/me` and accounts endpoints return 503 with a configuration message. Health and markets remain available. Keep this key private and stable: key rotation and recovery codes are future work; replacing it makes existing TOTP secrets unreadable.

## HTTP contract

All responses and errors are JSON (`{"error":"message"}`), with `Cache-Control: no-store`. Unknown JSON fields, trailing JSON values and bodies over 4 KiB are rejected. Every mutation requires `Content-Type: application/json` (parameters such as `charset=utf-8` are allowed). This prevents cross-site HTML form submissions from using cookie auth; do not configure credentialed CORS for untrusted origins. There is no permissive CORS middleware. Clients on another origin must use an appropriately restricted proxy/CORS configuration; same-origin clients send cookies normally.

| Method and path | Request | Success |
| --- | --- | --- |
| POST `/v1/auth/register` | `{email, password}` | 201, user and session cookie |
| POST `/v1/auth/login` | `{email, password}` | 200, user and session cookie, or `{two_factor_required: true}` and pending cookie |
| POST `/v1/auth/login/2fa` | `{code}` and pending cookie | 200, user and full session cookie; pending cookie cleared |
| POST `/v1/auth/logout` | `{}` | 200 `{ok: true}`; deletes presented full/pending sessions and clears cookies; idempotent |
| GET `/v1/me` | session cookie | 200 `{id, email, two_factor_enabled, created_at}` |
| POST `/v1/auth/2fa/setup` | `{}`, session cookie | 200 `{secret, otpauth_url}`; encrypted secret stored but not enabled |
| POST `/v1/auth/2fa/enable` | `{code}`, session cookie | 200 `{two_factor_enabled: true}` |
| POST `/v1/auth/2fa/disable` | `{code, password}`, session cookie | 200 `{two_factor_enabled: false}` |
| POST `/v1/accounts` | `{name, paper_balance}`, session cookie | 201 account |
| GET `/v1/accounts` | session cookie | 200 account array (empty array when none) |

Emails are trimmed, lowercased and parsed as addresses without display names, with a 254-byte limit. Passwords contain 12–128 Unicode characters and are never trimmed. Invalid registration/balance/name/JSON inputs return 400; missing/expired sessions and invalid passwords/codes return 401; duplicate registration and invalid 2FA state return 409; rate limits return 429 (`Retry-After: 900`); unavailable configuration/database returns 503. Unknown email and incorrect password have the same error response and both run Argon2id verification.

Accounts return `{id, user_id, type, name, quote_asset, paper_balance, created_at}`. Names are trimmed, 1–100 characters. Balance input must be a plain positive decimal string, at most 10,000,000 with up to eight fractional digits (no exponent/sign/whitespace). It is parsed by shopspring/decimal, stored as `numeric` and returned as a canonical decimal string. Only paper/USDT creation is exposed. The database supports `live` with NULL balance for step 8; the API rejects extra type/quote/user fields. Lists always filter by the authenticated user.

## Sessions and two-factor state

`metron_session` is a random 256-bit opaque token, encoded as base64url. Only its SHA-256 digest is stored in `app.sessions`. Full sessions expire 30 days after last use; authenticated requests atomically check expiry before extending it and refresh the cookie. Expired rows are never revived. Sessions record user agent and the socket peer IP. Forwarded headers are not trusted. Expired rows for a user are pruned on login; deployments can periodically delete rows where `expires_at <= now()`.

The same table has `pending_2fa=true` rows with a separate `metron_pending` cookie and fixed five-minute expiry. This keeps FK cleanup and token handling shared without granting partial logins access to protected endpoints. Completing 2FA atomically consumes the pending row and counter, then creates a fresh full session/token. A pending token is rejected even if moved into `metron_session`. Login session creation locks the user row and checks current 2FA state, preventing races with enable/disable.

Both cookies use Path=/, HttpOnly, SameSite=Lax, no Domain, and Secure only when `COOKIE_SECURE=true`. Logout deletes the supplied tokens; enabling/disabling 2FA revokes every other full and pending session while preserving the authenticated caller. Setup cannot overwrite an enabled secret. Setup stores a new secret each time, so use the latest returned URL/code.

TOTP uses pquerna/otp: SHA-1, six digits, 30-second steps, ±1 step clock tolerance. The highest accepted counter is compared and updated atomically with the encrypted-secret snapshot and expected enabled state. This rejects code replay across sessions, concurrent requests and state changes. After enabling or logging in, wait for a later authenticator code before another operation requiring a code; accepting a future-window code also advances the counter. Disable requires both password and a fresh code and removes the secret. There is no recovery/reset UI yet.

Passwords use PHC-style Argon2id v19 hashes: 64 MiB memory, three passes, two lanes, a fresh 16-byte salt and 32-byte output, compared in constant time. `internal/cryptobox` uses AES-256-GCM with a fresh random nonce, a version byte and authenticated purpose (`totp:<user UUID>`). Its format and API can also encrypt exchange credentials with a separate purpose. Plaintext secrets are returned only at setup and never stored or logged.

An in-memory mutex-protected limiter allows ten password attempts per normalized email and per peer IP in 15 minutes (registration shares these buckets). TOTP attempts have separate per-user and per-IP buckets, also ten per 15 minutes. Buckets expire and memory is bounded. Limits reset on API restart and are per process; a shared limiter is needed before running multiple API replicas. A reverse proxy needs explicit trusted-client-IP handling; currently requests behind one share its IP budget.

## Verification

From `api/`, run `go generate ./...`, `gofmt -l .` (empty), `go vet ./...` and `go test ./...`. Unit tests use fakes for validation, password login, TOTP/replay/state transitions, session expiry, balance precision, cookie flags, HTTP policies and rate limits. Crypto tests cover nonce uniqueness, tamper detection and purpose binding.

Integration tests skip unless `TEST_DATABASE_URL` is set. Point it at a dedicated Postgres database with the `db/app` migration applied (for example `goose -dir ../db/app -table goose_app_version postgres "$TEST_DATABASE_URL" up`), then run `go test ./...` or `go test -run TestPostgresAccountsAndSignIn -v ./internal/auth`. They create uniquely named users, exercise real HTTP handlers and PostgreSQL transactions/constraints, test account isolation, expiry, 2FA replay including concurrent code use, and logout. Cleanup deletes only those users, cascading their accounts and sessions; no truncation or schema deletion.

The web/mobile work is to connect sign-in, the pending-code prompt, setup/enable/disable and paper account creation/list screens, keeping codes/balances as strings and sending JSON with cookies. No web or mobile files change in this backend step.

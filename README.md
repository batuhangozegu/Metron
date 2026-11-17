# Metron

Metron is the second generation of [AETHER](https://github.com/batuhangozegu/AETHER) (formerly referred to as AETHER v2), a crypto risk management and portfolio platform. v1 (Java/Spring Boot + Flutter) is complete and archived; Metron is a ground-up rewrite with a Go + Rust backend and web + mobile clients.

The name comes from the Greek *metron*, "measure": measure the risk before every order, and act with measure.

The design is entirely new and the scope is more detailed. The main focus is a risk engine that calculates stop-loss and position size according to each coin's own characteristics.

## Status

Market data (roadmap step 1) connects Binance candles and live prices to the web Markets page, with volume-based coin classes and daily ATR. Other web and mobile screens still use sample data; see [ROADMAP.md](docs/ROADMAP.md).

Accounts and sign-in (step 2 backend) now support cookie sessions, optional TOTP two-factor authentication and paper accounts. Web and mobile sign-in screens still need connecting.

## Running locally

`task up` builds and starts Postgres, Redis, migrations, the Rust engine, the Go API and the Next.js web app through Docker Compose:

```sh
cp deploy/.env.example deploy/.env   # change the password and set APP_ENCRYPTION_KEY
task up                            # or: docker compose -f deploy/compose.yaml up -d --build
curl http://127.0.0.1:8080/healthz   # the API answers once it can reach the engine
curl http://127.0.0.1:8080/v1/markets
```

Open [Markets](http://localhost:3000/markets). Pairs appear as the first candle sync completes. The page refreshes every five seconds while visible, showing live prices, classes, Wilder daily ATR(14) and suggested stops using the temporary system default of 1.5 times ATR. Stablecoins have no stop. Optional watched base assets and Binance endpoints are documented in `deploy/.env.example`. Passwords containing URL-reserved characters must be percent-encoded in the connection URL; use a URL-safe password for Compose's interpolated URL.

Common jobs are in `Taskfile.yml`. Install [Task](https://taskfile.dev) once with `go install github.com/go-task/task/v3/cmd/task@latest`, then run `task` to list them. `task migrate` applies migrations separately. For frontend development, run `docker compose -f deploy/compose.yaml stop web`, then `task web` (requires the web npm dependencies installed). It uses `METRON_API_URL`, defaulting to `http://127.0.0.1:8080`; Compose uses `http://api:8080`.

For the Go API outside Docker, set `DATABASE_URL=postgres://metron:<password>@127.0.0.1:5433/metron?sslmode=disable` and optionally `REDIS_URL=redis://127.0.0.1:6379`. There is no default database password: without `DATABASE_URL`, the API starts and `/v1/markets` returns a JSON 503 until configured. Missing Redis prices appear as empty values while database statistics remain available.

## Accounts with curl

Generate `APP_ENCRYPTION_KEY` with `openssl rand -base64 32`, put it in `deploy/.env` before starting the API, and retain it across restarts. Outside Compose, export it in the API process environment. Missing or invalid keys leave auth endpoints returning JSON 503. `task migrate` applies both `market` and `app` migrations with separate goose version tables.

The following shell commands use a cookie jar. Every mutation requires `Content-Type: application/json`, including logout and 2FA setup. Money is always a JSON string.

```sh
API=http://127.0.0.1:8080
curl -i -c cookies.txt -H 'Content-Type: application/json' \
  -d '{"email":"you@example.com","password":"a long local password"}' "$API/v1/auth/register"
curl -b cookies.txt "$API/v1/me"
curl -b cookies.txt -c cookies.txt -H 'Content-Type: application/json' \
  -d '{"name":"Practice","paper_balance":"10000.12345678"}' "$API/v1/accounts"
curl -b cookies.txt "$API/v1/accounts"

# Add the returned secret/otpauth_url to an authenticator, then enter its code.
curl -b cookies.txt -c cookies.txt -H 'Content-Type: application/json' \
  -d '{}' "$API/v1/auth/2fa/setup"
curl -b cookies.txt -c cookies.txt -H 'Content-Type: application/json' \
  -d '{"code":"123456"}' "$API/v1/auth/2fa/enable"
curl -b cookies.txt -c cookies.txt -H 'Content-Type: application/json' \
  -d '{}' "$API/v1/auth/logout"
curl -c cookies.txt -H 'Content-Type: application/json' \
  -d '{"email":"you@example.com","password":"a long local password"}' "$API/v1/auth/login"
# When two_factor_required is true, wait for a new 30-second code after enable.
curl -b cookies.txt -c cookies.txt -H 'Content-Type: application/json' \
  -d '{"code":"654321"}' "$API/v1/auth/login/2fa"
curl -b cookies.txt "$API/v1/me"
```

Replace sample codes with current authenticator codes; a code cannot be used twice. Session cookies are HttpOnly, SameSite=Lax and slide for 30 days; use `-b cookies.txt -c cookies.txt` to retain refreshed expiry in long-running curl use. Set `COOKIE_SECURE=true` for HTTPS (default false for local HTTP). The pending 2FA cookie expires after five minutes and cannot authorize account operations. See [backend auth details and tests](docs/AUTH.md).

## Target technology

- **Go:** API gateway, authentication, user/portfolio CRUD, REST endpoints.
- **Rust:** risk engine, live price feed (WebSocket), PnL/position calculations.
- **Clients:** web: Next.js (React + TypeScript), mobile: Flutter.

## Relationship to v1

No code will be carried over from v1. What will be used as reference: the API contract, the Flyway SQL schema and the risk calculation rules.

## Documents

- [Feature list](docs/FEATURES.md)
- [Architecture decisions](docs/ARCHITECTURE.md)
- [First release roadmap](docs/ROADMAP.md)

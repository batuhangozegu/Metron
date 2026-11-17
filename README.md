# Metron

Metron is the second generation of [AETHER](https://github.com/batuhangozegu/AETHER) (formerly referred to as AETHER v2), a crypto risk management and portfolio platform. v1 (Java/Spring Boot + Flutter) is complete and archived; Metron is a ground-up rewrite with a Go + Rust backend and web + mobile clients.

The name comes from the Greek *metron*, "measure": measure the risk before every order, and act with measure.

The design is entirely new and the scope is more detailed. The main focus is a risk engine that calculates stop-loss and position size according to each coin's own characteristics.

## Status

Market data (roadmap step 1) connects Binance candles and live prices to the web Markets page, with volume-based coin classes and daily ATR. Other web and mobile screens still use sample data; see [ROADMAP.md](docs/ROADMAP.md).

## Running locally

`task up` builds and starts Postgres, Redis, migrations, the Rust engine, the Go API and the Next.js web app through Docker Compose:

```sh
cp deploy/.env.example deploy/.env   # then change the password
task up                            # or: docker compose -f deploy/compose.yaml up -d --build
curl http://127.0.0.1:8080/healthz   # the API answers once it can reach the engine
curl http://127.0.0.1:8080/v1/markets
```

Open [Markets](http://localhost:3000/markets). Pairs appear as the first candle sync completes. The page refreshes every five seconds while visible, showing live prices, classes, Wilder daily ATR(14) and suggested stops using the temporary system default of 1.5 times ATR. Stablecoins have no stop. Optional watched base assets and Binance endpoints are documented in `deploy/.env.example`. Passwords containing URL-reserved characters must be percent-encoded in the connection URL; use a URL-safe password for Compose's interpolated URL.

Common jobs are in `Taskfile.yml`. Install [Task](https://taskfile.dev) once with `go install github.com/go-task/task/v3/cmd/task@latest`, then run `task` to list them. `task migrate` applies migrations separately. For frontend development, run `docker compose -f deploy/compose.yaml stop web`, then `task web` (requires the web npm dependencies installed). It uses `METRON_API_URL`, defaulting to `http://127.0.0.1:8080`; Compose uses `http://api:8080`.

For the Go API outside Docker, set `DATABASE_URL=postgres://metron:<password>@127.0.0.1:5433/metron?sslmode=disable` and optionally `REDIS_URL=redis://127.0.0.1:6379`. There is no default database password: without `DATABASE_URL`, the API starts and `/v1/markets` returns a JSON 503 until configured. Missing Redis prices appear as empty values while database statistics remain available.

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

# Metron

Metron is the second generation of [AETHER](https://github.com/batuhangozegu/AETHER) (formerly referred to as AETHER v2), a crypto risk management and portfolio platform. v1 (Java/Spring Boot + Flutter) is complete and archived; Metron is a ground-up rewrite with a Go + Rust backend and web + mobile clients.

The name comes from the Greek *metron*, "measure": measure the risk before every order, and act with measure.

The design is entirely new and the scope is more detailed. The main focus is a risk engine that calculates stop-loss and position size according to each coin's own characteristics.

## Status

The web and mobile screens are built against sample data. Backend work has started with roadmap step 0 (skeleton); see [ROADMAP.md](docs/ROADMAP.md).

## Running locally

Postgres, Redis, the Rust engine and the Go API all run through Docker Compose:

```sh
cp deploy/.env.example deploy/.env   # then change the password
docker compose -f deploy/compose.yaml up -d --build
curl http://127.0.0.1:8080/healthz   # the API answers once it can reach the engine
```

Common jobs (start, stop, generate, test) are in `Taskfile.yml`. Install [Task](https://taskfile.dev) once with `go install github.com/go-task/task/v3/cmd/task@latest`, then run `task` to list them; `task up` and `task test` are the main ones.

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

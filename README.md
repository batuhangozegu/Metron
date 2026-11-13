# Metron

Metron is the second generation of [AETHER](https://github.com/batuhangozegu/AETHER) (formerly referred to as AETHER v2), a crypto risk management and portfolio platform. v1 (Java/Spring Boot + Flutter) is complete and archived; Metron is a ground-up rewrite with a Go + Rust backend and web + mobile clients.

The name comes from the Greek *metron*, "measure": measure the risk before every order, and act with measure.

The design is entirely new and the scope is more detailed. The main focus is a risk engine that calculates stop-loss and position size according to each coin's own characteristics.

## Status

Planning stage, no code yet.

## Target technology

- **Go:** API gateway, authentication, user/portfolio CRUD, REST endpoints.
- **Rust:** risk engine, live price feed (WebSocket), PnL/position calculations.
- **Clients:** web: Next.js (React + TypeScript), mobile: Flutter.

## Relationship to v1

No code will be carried over from v1. What will be used as reference: the API contract, the Flyway SQL schema and the risk calculation rules.

## Documents

- [Feature list](docs/FEATURES.md)
- [Architecture decisions](docs/ARCHITECTURE.md)

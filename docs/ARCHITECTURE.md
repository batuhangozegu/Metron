# Metron — Architecture Decisions

13 November 2026

Note: Coding has not started; this file records the decisions and their rationale. For the feature list, see [FEATURES.md](FEATURES.md).

## 1. Client frameworks

- **Web: Next.js (React + TypeScript).** The detailed part of v2 (charts, backtest results, portfolio analysis, dense tables) will live on the web. Financial chart and table/dashboard components are more mature on the JavaScript side; Flutter web is weak on data-dense screens.
- **Mobile: Flutter (as in v1, with Riverpod).** Flutter is already known. While learning two new languages on the backend (Go, Rust), a third novelty on mobile would slow the project down.
- **Cost of two clients.** Screens are written twice. To reduce this, the API is defined with OpenAPI and the TypeScript and Dart client code is generated automatically; only the UI is written by hand.
- **Rejected alternatives.** Using Flutter for the web as well (one codebase but weak web) and switching to React Native (code sharing but one more new thing to learn) were rejected.

## 2. Roles of web and mobile

The two clients will not be copies of the same application.

- **Web, the analysis desk.** Portfolio analysis, backtest, risk settings, coin class rules and the trade journal live here. Wide screen, dense data.
- **Mobile, monitoring and quick trading.** Open positions, alerts, the pre-trade check screen, quick orders and notifications live here. Heavy screens such as backtest will not be on mobile.

## 3. Design direction

- **The home screen shows risk.** The home screen shows the risk status, not prices: total open risk, the remaining share of the daily loss cap, broken rules.
- **Signature screen.** The pre-trade check screen is the center of the design and the signature screen of the application; most of the effort goes here.
- **Use of color.** Color is not just red/green: for profit/loss, signs and icons are used alongside color. There is a separate scale for risk level (safe, caution, limit exceeded).
- **Theme and numerals.** Dark theme first, light theme also supported. Numerals are set with equal widths so they stay aligned in tables.
- **One design system, two applications.** Color, type and spacing values are defined once; web and mobile use the same values.

## 4. Services and communication

- **Overall structure.** Not microservices, but two services: a Go service (API) and a Rust service (engine). Each is split into modules internally, but there will not be many separately deployed pieces; more than that is a burden in a one-person project.

| Link | Method | Purpose |
| --- | --- | --- |
| Client ↔ Go | REST + WebSocket | Normal requests over REST, live prices and alerts over WebSocket |
| Go → Rust | gRPC | Question-and-answer calls such as "calculate the risk of this order" |
| Rust → Go | Redis pub/sub | Events such as the price feed, alerts and stop triggers |

- **Why gRPC.** The contract between the two languages is defined in a single file and the Go and Rust code is generated from it; a wrong field name or type is caught at compile time.
- **Why Redis.** It also holds the latest prices; when Go needs a price, it reads it from Redis without asking Rust.

### 4.1 Who owns what

- **Rust engine.** Connects to the Binance price feed, calculates candles and ATR, runs the risk rules, and monitors the stops of open positions. It does not send orders to the exchange and does not write user data.
- **Go service.** Authentication, user and portfolio data, exchange API keys (encrypted), writing to the database, sending orders to the exchange. The only door the clients talk to.
- **Rationale.** API keys stay in a single service and orders go to the exchange from a single place.

### 4.2 The path of an order

1. The user enters the order, and the client sends it to Go.
2. Go asks Rust, over gRPC, about the order and the user's risk profile.
3. Rust returns the position size, the stop level and the rules being broken; the pre-trade check screen shows this response.
4. If the user approves, Go sends the order to the exchange and writes it to the database.

### 4.3 Two rules

- **The stop always sits on the exchange too.** If the system tracks the stop only internally, the position is left unprotected when the server crashes. A real stop order is placed on the exchange for every position; in cases such as a trailing stop, the engine updates that order.
- **Paper trading goes through the same path.** On the Go side the exchange connection is a swappable component (one is the real Binance, one is a virtual exchange). The risk engine does not distinguish between the two; paper trading is not written as a separate system.

## 5. Open decisions

1. Database: what remains from the v1 schema, and how coin classes and the "asset class" concept are modeled.
2. Repository layout: single repository or separate repositories.

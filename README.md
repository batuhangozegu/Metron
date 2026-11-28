# Metron

Metron is the second generation of [AETHER](https://github.com/batuhangozegu/AETHER) (formerly referred to as AETHER v2), a crypto risk management and portfolio platform. v1 (Java/Spring Boot + Flutter) is complete and archived; Metron is a ground-up rewrite with a Go + Rust backend and web + mobile clients.

The name comes from the Greek *metron*, "measure": measure the risk before every order, and act with measure.

The design is entirely new and the scope is more detailed. The main focus is a risk engine that calculates stop-loss and position size according to each coin's own characteristics.

## Status

The first release (roadmap steps 0–8, plus price alerts) is coded: live market data, accounts with optional 2FA, risk checks with separate spot and futures rule sets, paper trading with stop tracking, saved limits and check history, portfolio tracking and journal, the mobile app, and real Binance spot/USDT-M accounts with exchange-side stops. Real trading on Binance's test environment has not been tried yet; see [ROADMAP.md](docs/ROADMAP.md).

## Experimental: trade ideas (planned, not built)

Today Metron suggests a **stop** (ATR-based) and a **position size** for a trade the user has already chosen. It does not yet say *what* to trade. The planned experimental feature adds that step:

- **Spot buy ideas.** Long-only suggestions for spot accounts: which watched pair looks like a buy now, with entry, ATR stop, target and size from the existing risk engine.
- **Leveraged futures ideas.** Long or short suggestions for USDT-M accounts. Leverage is not picked for excitement: it follows from the stop distance and risk per trade, so the position never risks more than the account's saved limits allow, and it is capped well below the exchange maximum.
- **Method.** Time-series analysis of the hourly and daily candles the engine already stores: trend (moving-average slope and crossovers), momentum, volatility regime (ATR percentile) and volume. Each idea carries a confidence score and the reasons behind it. Other methods, such as simple forecasting models, can be compared later through backtests.
- **Guard rails.** An idea is only a suggestion: it is never sent automatically, it must pass the normal pre-trade check, and it is tried on paper accounts first. Ideas are recorded so their hit rate can be measured honestly. This is not investment advice.

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

## Paper trading with curl

After applying app migrations through `00005`, sign in with the cookie jar above and
copy an account ID from `GET /v1/accounts`. The engine must have a fresh class/ATR
snapshot and Redis must have a latest ticker for the symbol. The API chooses the
entry price and runs the risk check against the account's actual open positions
and settled UTC day/week history, using the account's saved risk profile.

```sh
ACCOUNT_ID=replace-with-your-paper-account-id
curl -b cookies.txt -c cookies.txt -H 'Content-Type: application/json' \
  -d "{\"account_id\":\"$ACCOUNT_ID\",\"symbol\":\"BTCUSDT\",\"side\":\"long\"}" \
  "$API/v1/orders"
# The 201 response includes order, fill, position and check. Copy position.id.
POSITION_ID=replace-with-the-returned-position-id
curl -b cookies.txt "$API/v1/accounts/$ACCOUNT_ID/positions"
curl -b cookies.txt -c cookies.txt -H 'Content-Type: application/json' \
  -d '{}' "$API/v1/positions/$POSITION_ID/close"
curl -b cookies.txt "$API/v1/accounts/$ACCOUNT_ID/positions?status=closed"
curl -b cookies.txt "$API/v1/accounts"
```

`manual_stop`, `quantity` and spot `leverage` are optional. A non-null
`risk_profile` on an order returns 400 `code: "risk_profile_not_allowed"`.
Quantity defaults to the suggested size; a supplied decimal string must not
exceed that size. Futures accounts require integer `leverage` (1–125); spot
permits only long positions and leverage 1. Any failed rule produces HTTP 422
with `error.id = "risk_blocked"`, the full `check`, and a rejected order record.
Other trading errors use `{error:{id,message_en,message_tr}}`.

The paper exchange charges 0.1% of notional per spot fill and 0.05% per futures
fill, in USDT. Opening requires available cash to cover notional (spot) or initial
margin (futures), plus the opening fee. Cash reserves the cost basis or margin
and opening fees of every open position. The balance changes only on close, by
directional PnL minus both fill fees in one transaction. Open-position
`unrealized_pnl` is directional PnL before settlement fees; it and `last_price`
are null if Redis has no price. Closed positions omit these fields and list
the newest 100 closes first. Repeating a manual close returns HTTP 409 with
`error.id = "position_closed"`.

The API subscribes to `EngineService.WatchStops` and reconnects with backoff.
Stop hits close at the stored stop price, including when Redis is unavailable;
duplicate hits settle once. Stop monitoring requires the step 4 engine changes.
Trusted equity is rounded down to eight decimal places, while activity
percentages and class holdings are rounded up conservatively for the engine.

## Saved risk limits and history with curl

App migration `00004_risk_profiles_checks.sql` adds persistent account limits
and check history. Use a full session cookie and an owned account ID. With no
saved profile, GET returns the defaults and `updated_at: null`. PUT requires
all eight fields; decimals are strings, counts are positive uint32 integers.

```sh
curl -b cookies.txt "$API/v1/accounts/$ACCOUNT_ID/risk-profile"
curl -b cookies.txt -c cookies.txt -X PUT -H 'Content-Type: application/json' \
  -d '{"risk_per_trade_percent":"0.1","target_risk_reward":"2","total_open_risk_limit_percent":"6","max_open_positions":5,"max_losing_trades_today":3,"daily_loss_limit_percent":"0.5","max_stopped_trades_this_week":5,"weekly_loss_limit_percent":"10"}' \
  "$API/v1/accounts/$ACCOUNT_ID/risk-profile"
# Uses the saved profile; response includes profile_source: "stored".
ENTRY_PRICE=replace-with-current-BTCUSDT-price
curl -b cookies.txt -c cookies.txt -H 'Content-Type: application/json' \
  -d "{\"account_id\":\"$ACCOUNT_ID\",\"symbol\":\"BTCUSDT\",\"side\":\"long\",\"entry_price\":\"$ENTRY_PRICE\"}" \
  "$API/v1/risk/check"
# What-if only: supplied fields override saved limits without saving them.
curl -b cookies.txt -c cookies.txt -H 'Content-Type: application/json' \
  -d "{\"account_id\":\"$ACCOUNT_ID\",\"symbol\":\"BTCUSDT\",\"side\":\"long\",\"entry_price\":\"$ENTRY_PRICE\",\"risk_profile\":{\"risk_per_trade_percent\":\"0.2\"}}" \
  "$API/v1/risk/check"
# Newest first; default limit 20, valid range 1..100.
curl -b cookies.txt "$API/v1/accounts/$ACCOUNT_ID/risk-checks?limit=20"
```

Every completed preview is recorded, including failures of risk rules. Opening
orders always use saved limits under the account lock; their checks and filled
or rejected order records commit together. Loss brakes use settled net PnL
over UTC days and Monday-start weeks. A tighter daily limit such as 0.5% blocks
new orders once that loss percentage is reached. Preview overrides cannot
authorize orders or change saved limits. See [RISK.md](docs/RISK.md).

## Portfolio tracking and journal with curl

Apply app migration `00005_portfolio.sql`, then use the same full session cookie
and owned paper account ID as positions. Other users' accounts return 404.
All amounts and percentages are decimal strings; timestamps are UTC.

```sh
# Default period is 30d; the other valid choices are 90d and all.
curl -b cookies.txt "$API/v1/accounts/$ACCOUNT_ID/portfolio?period=30d"
curl -b cookies.txt "$API/v1/accounts/$ACCOUNT_ID/portfolio?period=all"
# Settled trades, newest first; limit defaults to 50 and accepts 1..200.
curl -b cookies.txt "$API/v1/accounts/$ACCOUNT_ID/journal?limit=50"
# Copy next_before from the previous page; null means there is no next page.
BEFORE=replace-with-next_before-UTC-timestamp
curl -b cookies.txt --get --data-urlencode 'limit=50' \
  --data-urlencode "before=$BEFORE" "$API/v1/accounts/$ACCOUNT_ID/journal"
```

Portfolio responses include starting/realized balance, reserved cash, live
unrealized PnL and equity, period fees, grouped holdings and closed-trade
statistics by coin class. Equity history keeps at most 200 bucketed observations,
including a live point, and reports maximum drawdown. The BTC benchmark scales
stored BTCUSDT daily candle closes to the first equity point; missing candles
produce an empty benchmark and null return. Journal entries include both fill
fees, net PnL, R multiples and caution/fail rules from the opening order's check.

The API snapshots all paper accounts at startup and each full UTC hour, plus
after successful opens and closes. Missing Redis prices for any open position
skip that account's snapshot and log the failure; trades remain successful.
Live portfolio valuation returns 503 until all open positions have prices.
Starting balances for existing accounts are backfilled from settled net PnL.

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

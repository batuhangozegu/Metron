# Live Binance accounts

Apply app migrations through `00006_live_accounts.sql` and set the same
`APP_ENCRYPTION_KEY` used for authentication. Keys are encrypted with separate
purposes and responses expose only the final four API-key characters. Authenticate
with the existing cookie flow first; these examples use its `cookies.txt` jar.

```sh
API=http://127.0.0.1:8080
# Supply keys for the selected environment and market (spot or USDT-M futures).
curl -b cookies.txt -H 'Content-Type: application/json' "$API/v1/accounts/live" \
  -d '{"name":"Binance testnet","market":"spot","environment":"testnet","api_key":"YOUR_TESTNET_KEY","api_secret":"YOUR_TESTNET_SECRET"}'

# Copy the returned account id. Existing markets use pairs such as BTCUSDT;
# live orders also accept BTC and normalize it to BTCUSDT.
ACCOUNT_ID=REPLACE_WITH_ACCOUNT_ID
curl -b cookies.txt "$API/v1/accounts"
curl -b cookies.txt -H 'Content-Type: application/json' "$API/v1/orders" \
  -d "{\"account_id\":\"$ACCOUNT_ID\",\"symbol\":\"BTCUSDT\",\"side\":\"long\"}"

# Futures orders also require leverage, for example "leverage":3.
# Mainnet orders additionally require "confirm_live":true on every open request;
# omission returns HTTP 400 with code live_confirmation_required.
curl -b cookies.txt "$API/v1/accounts/$ACCOUNT_ID/positions"
curl -b cookies.txt "$API/v1/accounts/$ACCOUNT_ID/portfolio?period=all"

POSITION_ID=REPLACE_WITH_POSITION_ID
# A failed protective stop leaves the real fill open and visible. Inspect
# exchange_stop.status/error and retry it; only failed open live stops qualify.
curl -b cookies.txt -X POST "$API/v1/positions/$POSITION_ID/stop/retry"
curl -b cookies.txt -X POST "$API/v1/positions/$POSITION_ID/close"

# Only owned live accounts without open positions can be deleted (204).
# Open live positions and paper accounts return 409; foreign accounts return 404.
curl -b cookies.txt -X DELETE "$API/v1/accounts/$ACCOUNT_ID"
```

Live orders use exchange quote equity for risk sizing, round down to exchange
quantity/price filters and store actual execution prices, quantities and fees.
Protective stops live on Binance. The API reconciles them every 20 seconds and
immediately on engine stop events; filled stops settle with `exchange_stop` as
the close reason. Manual closes cancel the protective stop before submitting
the opposite market order (reduce-only for futures). Exchange failures skip
equity snapshots; live portfolio equity/cash use quote total/free balance.

Optional endpoint overrides: `BINANCE_SPOT_TESTNET_URL`,
`BINANCE_SPOT_MAINNET_URL`, `BINANCE_FUTURES_TESTNET_URL`, and
`BINANCE_FUTURES_MAINNET_URL`. Paper trading keeps its existing execution path.

Run `go vet ./...` and `go test ./...` from `api/`. Set `TEST_DATABASE_URL` to
a dedicated database with all migrations applied to enable PostgreSQL tests.
The contract-only `internal/binance/binance.go` placeholder must be replaced
by lane 1's implementation at integration; it intentionally cannot execute
real exchange requests.

# API walkthroughs

See the [root README](../README.md) for local setup and cookie-based sign-in.

## Price alerts

Apply app migration `00007_price_alerts.sql`. Alerts belong to the signed-in user,
independently of trading accounts. Choose an exact `symbol` from `/v1/markets`.
Prices are decimal strings, positive with at most eight decimal places; notes may
contain up to 200 characters. Each user may have 50 active alerts (409 at the cap).

```sh
API=http://127.0.0.1:8080
curl "$API/v1/markets"
# Replace BTCUSDT with a configured symbol if needed.
curl -b cookies.txt -c cookies.txt -H 'Content-Type: application/json' \
  -d '{"symbol":"BTCUSDT","direction":"above","price":"65000","note":"Review BTC"}' \
  "$API/v1/alerts"
curl -b cookies.txt -c cookies.txt "$API/v1/alerts"
# Poll every 15 seconds, retaining the prior poll's timestamp as the cursor.
curl -b cookies.txt -c cookies.txt --get \
  --data-urlencode 'triggered_since=2026-11-24T00:00:00Z' "$API/v1/alerts"
# Use the id returned by POST or GET.
ALERT_ID=replace-with-alert-uuid
curl -i -b cookies.txt -c cookies.txt -X DELETE \
  -H 'Content-Type: application/json' "$API/v1/alerts/$ALERT_ID"
```

Listing returns up to 200 alerts, newest first. `triggered_since` requires RFC3339
and returns only triggered alerts strictly after that timestamp. DELETE marks an
owned alert cancelled and returns 204; unknown or foreign ids return 404. Every
five seconds the checker reads Redis prices for active symbols: `above` triggers
at or above the target, `below` at or below. It records the observed price and
trigger time atomically, and retries on subsequent ticks after Redis errors.

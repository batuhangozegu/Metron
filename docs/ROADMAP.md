# Metron — First Release Roadmap

14 November 2026

Note: No code has been written yet; this file orders the work for the first release defined in section 9.1 of [FEATURES.md](FEATURES.md), based on the decisions in [ARCHITECTURE.md](ARCHITECTURE.md).

## 1. Principles

- **Every step ends with something demonstrable.** Not "infrastructure is ready" but "when I open X, I see Y".
- **Paper money first, real money last.** No order goes to a real exchange until the system has been shown to behave correctly in paper trading.
- **Web first, then mobile.** One UI while the backend settles; the mobile app is written quickly once the API is stable.

## 2. Steps

| Step | Work | Demonstrable result |
| --- | --- | --- |
| 0. Skeleton | Folders, Docker Compose (Postgres, Redis), the Go and Rust services, one test call between them, CI | One command starts everything; Go asks Rust a question and gets an answer |
| 1. Market data | Rust connects to Binance, stores hourly and daily candles, computes ATR, writes latest prices to Redis; coins are classified by volume | A live price list on the web with each coin's class and ATR |
| 2. Accounts and sign-in | Registration, login, two-factor authentication, creating a paper account | Sign in and create an account with a paper balance |
| 3. Risk calculation | Coin class rules, risk profile, ATR-based stop and position size | First version of the pre-trade check screen: pick a coin, see the suggested stop and size |
| 4. Paper trading | Orders, fills and positions tables; simulated exchange; stop tracking | Open an order in the paper account; the position closes when price reaches the stop |
| 5. Portfolio limits and discipline | Total open risk limit, loss brake, discipline rules, risk check records | An order over the limit is blocked and the check screen names the violated rule |
| 6. Portfolio tracking | Equity snapshots, historical performance, trade statistics, cost tracking | Equity chart, maximum drawdown, win rate |
| 7. Mobile | Flutter app: positions, alarms, pre-trade check screen, quick order | Sign in to the same account on a phone and open an order |
| 8. Real exchange | Connecting a Binance account, sending real orders, keeping the stop order on the exchange as well | Trade first on Binance's test environment, then on a real account with a small amount |

## 3. Notes on the order

- **Step 0 is small but the most important.** Go, Rust, gRPC and Docker are set up together for the first time here, and most surprises appear at this point. Finishing it before any feature work speeds up every later step.
- **v1's formula becomes tests in step 3.** The inputs and outputs of v1's `risk amount / |entry − stop|` calculation are the Rust engine's first tests, so the new engine is known to be at least as correct as the old one.
- **The product is usable at the end of step 4.** It can be shown in a portfolio from that point; steps 5 and later build on top.
- **No time estimates.** They depend on how fast Go and Rust are learned. A realistic estimate for the rest is possible once steps 0 and 1 are done.

## 4. After the first release

The second release and the idea pool are listed in sections 9.2 and 9.3 of [FEATURES.md](FEATURES.md); nothing from them starts before step 8 is finished.

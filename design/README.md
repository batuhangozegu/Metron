# Metron — Design

14 November 2026

Note: This folder holds the visual design that section 8 of [ARCHITECTURE.md](../docs/ARCHITECTURE.md) left open. It covers the web client for the first release; mobile reuses the same values.

## 1. Files

- **`tokens.json`.** The single source for color, type, spacing and radius. Web (Tailwind theme) and mobile (Flutter theme) both read their values from it.
- **`mockups/`.** One mockup per screen, exported from the design canvas. They are the reference for layout, copy and sample data; they are not application code and do not open on their own in a browser (they load a `support.js` that only exists on the canvas). Read them as source: the markup carries the layout and inline styles, and the sample data sits in `renderVals()` at the bottom of each file.

## 2. Rules

- **Neutral interface.** Surfaces and text are grey. Only two things carry color: the result of a trade and the distance to a limit.
- **Result of a trade.** `profit` and `loss`, always with an arrow (▲ ▼) and a sign (+ −), never color alone.
- **Risk scale.** `safe`, `caution`, `limit`. It is separate from profit/loss and always shown as color, icon (circle-check, triangle, octagon) and word together. Caution starts at 75% of a limit.
- **The measure.** A ruler with ten tick divisions above a bar. It always shows a real quantity against a real limit, with a caution mark at 75% and a limit mark at the end. It is the only ornament in the interface.
- **Numerals.** Every number is set in JetBrains Mono with tabular figures. Words are set in Inter.
- **Shape.** Flat cards with a one-pixel line and no shadow. Controls are 44 px high.
- **Theme.** Dark is the default; light uses the same token names.

## 3. Screens

| Mockup | Route | Purpose |
| --- | --- | --- |
| `Main.dc.html` | `/` | Risk home: status headline, three gauges, open positions, discipline |
| `PreTrade.dc.html` | `/check` | Signature screen: order form, verdict, price measure, limits after the order |
| `PreTradeBlocked.dc.html` | `/check` | The same screen when a limit is exceeded: no way to place the order, only reduce or discard |
| `Markets.dc.html` | `/markets` | Live price list with class, ATR and suggested stop |
| `Portfolio.dc.html` | `/portfolio` | Account value chart, drawdown, statistics, cost, trade journal |
| `Rules.dc.html` | `/rules` | Account limits, coin class rules, discipline rules |
| `Foundations.dc.html` | none | Token reference in both themes |

### 3.1 Mobile

Mobile is for monitoring and quick trading, not analysis (ARCHITECTURE.md section 2). Screens are 390 px wide, with a four-item bottom bar (Risk, Markets, Alerts, Rules) and the main action fixed above it.

| Mockup | Purpose |
| --- | --- |
| `MobileHome.dc.html` | Risk status, total open risk gauge, daily and weekly caps, discipline notice, open positions, "Check a new order" |
| `MobileCheck.dc.html` | Pre-trade check: pair and price, verdict, stop/entry/target measure, limits after the order, actions fixed at the bottom |
| `MobilePosition.dc.html` | One position: unrealized result, where the price sits between stop and target, stop status, tighten or close |
| `MobileAlerts.dc.html` | Today's events (cooldown, stops triggered, price alarms, blocked orders) and the price alarm list |
| `MobileMarkets.dc.html` | Search, class filter and the price list with class, ATR and suggested stop; a tap opens the check |
| `MobileRules.dc.html` | Account limits, coin classes and discipline rules, read-only (rules are changed on the web), plus account and sign out |
| `MobileCheckBlocked.dc.html` | The pre-trade check when a limit is exceeded: reduce or discard only |
| `MobileSignIn.dc.html` | Sign in, and the two-factor code step (the `step` tweak switches between them) |

## 4. Verdicts of the pre-trade check

| Verdict | When | Banner | Actions |
| --- | --- | --- | --- |
| All clear | Every check is safe | `safe` color, circle-check, "Within your limits" | Place order (primary), Discard |
| Caution | A check is past 75% of its limit, or a warn-and-record discipline rule would be broken | `caution` color, triangle | Wait (primary), Place anyway (recorded as a violation), Discard |
| Blocked | A limit is exceeded, or a blocking discipline rule would be broken | `limit` color, octagon | Reduce to fit (primary), Discard; the order cannot be placed |

Only the caution and blocked verdicts are drawn; all clear uses the same layout with the banner and actions above.

## 5. Not designed yet

- **Web sign-in, two-factor authentication and account creation.** Drawn for mobile only; the web versions follow the same form.
- **Mobile close-position and tighten-stop confirmations, and the account creation form.** Drawn before roadmap step 7.

## 6. Sample data

The mockups share one consistent sample account (25,000 USDT paper account, three open positions, a PENDLE/USDT order under check). The numbers are illustrative; they are not defaults for the risk engine.

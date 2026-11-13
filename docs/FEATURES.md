# AETHER v2 — Feature List

13 November 2026

Note: This list is at the idea stage; it is not a finalized scope.

## 1. State of v1 (starting point)

- **The stop price is entered by hand.** In v1 the user enters the stop price themselves; the system only calculates the position size as `risk amount / |entry − stop|`. The system does not know whether the stop is in the right place.
- **Risk profile fields.** Risk per trade as a percentage (default 2%), daily loss cap (5%), target risk/reward ratio (2.0), automatic stop-loss switch, maximum number of open positions (5).
- **Core idea.** Bitcoin and an altcoin should not get the same stop logic; the v2 risk engine will be built on this.

## 2. Per-coin stop-loss (priority 1)

- **Volatility-based stop (ATR).** The stop distance is not a fixed percentage; it is set from the coin's average movement over recent days. Example (illustrative figures): if BTC moves 2% a day on average, the stop is 3%; if an altcoin moves 8%, the stop is 12%. When the stop is wide, the position size shrinks automatically and the amount of money at risk stays the same.
- **Coin classes.** BTC/ETH, large-cap altcoin, small-cap altcoin, stablecoin. Each class gets its own ATR multiplier, maximum position size and risk per trade (example: BTC 2%, small-cap altcoin 0.5%).
- **Liquidity check.** On a coin with a thin order book, the price slips even when the stop triggers. The system estimates slippage from the order book and, if needed, reduces the position size or warns.

## 3. Portfolio-level risk (priority 2)

- **Total open risk limit.** An upper bound is set on what percentage of the account is lost if all stops trigger at the same time. (v1 only has a limit on the number of positions.)
- **Correlation.** Five longs in five different altcoins are not five separate risks; if BTC falls they fall together. The system calculates the total risk of open positions using correlation and warns.
- **Loss brake.** In addition to the daily cap, a weekly cap and a cooldown period after consecutive losses.
- **Market regime.** When the market as a whole is very volatile, all risk percentages drop automatically.

## 4. Stop types (priority 3)

- **Trailing stop.** As the price moves in your favor, the stop follows it.
- **Break-even stop.** Once a certain profit is reached, the stop is moved to the entry price.
- **Scaled take-profit.** The position is closed in stages rather than all at once.
- **Time stop.** If the price does not move within a set time, the position is closed.

## 5. Portfolio management

Three topics (portfolio management, stablecoin, tokenized stocks) are grouped under a single idea: target allocation.

### 5.1 Target allocation (main idea)

- **Target allocation definition.** The user defines once how they want the portfolio to look (example: 40% BTC/ETH, 20% altcoin, 25% stablecoin, 15% tokenized stocks). The system continuously compares the actual allocation with it. v1 only showed "what you hold"; v2 also answers "how far have you drifted from where you should be".
- **Drift alert.** When a class's share moves noticeably away from the target (example: altcoin from 20% to 35%), the system warns and lists the buys and sells needed to return to balance.
- **Rebalancing suggestion.** The system does not send orders itself; it prepares the suggestion, and it is applied if the user approves. Automatic execution may be left to a later stage.
- **Class limits.** The same classes as the coin classes in the risk engine are used. A rule such as "total small-cap altcoin cannot exceed 10% of the account" blocks a new order up front.

### 5.2 Stablecoin (cash role)

- **Cash portion.** Stablecoin is not a separate feature; it is treated as the cash portion of the portfolio.
- **Retreat from risk.** When the loss brake triggers or the market is very volatile, the system suggests increasing the stablecoin share. This links the risk engine and the portfolio side.
- **Parking profit.** Money from scaled take-profit sales moves into stablecoin and counts as cash in the target allocation.
- **Stablecoin's own risk.** Concentration in a single stablecoin and loss of its peg (depeg) are monitored; an alarm is raised if the price moves away from 1 dollar by a set amount.

### 5.3 Tokenized stocks (not in the first release)

- **Purpose.** Diversifies the portfolio as a non-crypto asset class, but will not be included in the first release.
- **Access is unclear.** Which exchange and which country it can be bought from varies. Whether Binance offers it and whether it is accessible from Turkey will be checked before work starts (open question).
- **Different behavior.** Tokens can trade while the stock market is closed, the price can detach from the real stock, and liquidity is low. A separate class is needed in the risk engine.
- **Additional integration.** A second exchange integration may be required.
- **Decision.** The data model will be built flexibly from the start around the concept of "asset class", and stocks will be added later.

### 5.4 Portfolio tracking

- **Historical performance.** Account value over time, maximum drawdown, and a "what if you had only held BTC" comparison.
- **Trade journal and statistics.** Win rate, average win/loss, and in which coin class you win and in which you lose. This is the part that best serves v1's goal of "preventing emotional decisions".
- **Cost tracking.** Average purchase price and commissions paid.

### 5.5 Priority order on the portfolio side

1. Historical performance and trade statistics.
2. Target allocation, drift alert and the cash role of stablecoin.
3. Rebalancing suggestion.
4. Tokenized stocks (once access is clear).

## 6. Unfinished work carried over from v1

- **Market service.** Coin list, candle data (klines), live price feed via WebSocket + Redis.
- **Shared logic.** Consolidating the user lookup logic in one place; having PnL calculation in a single shared module (in v1 it was duplicated in two places).
- **Notification system.**
- **Two-factor authentication.** In v1 the backend was ready but there was no mobile integration.

## 7. Architecture note

- All calculations in sections 2-4 are done in the Rust risk engine; the Go side handles the API and user operations.
- First step: once the scope is finalized, document the v1 API as OpenAPI and adapt the database schema.

## 8. Additional features

Features outside the risk engine and the portfolio.

- **Pre-trade check screen.** Before an order is sent, a single screen shows how much can be lost on this trade, what the total open risk will be, the drift from the target allocation, and which rules are being broken. It makes AETHER's core idea ("a risk calculation before every order") visible; the risk engine and the portfolio side already produce the data.
- **Paper trading.** Trading in a virtual account with live prices and no real money. The user tries their rules without risk, and the system is also tested without risking real money. It is the exchange-independent version of v1's Binance demo support.
- **Backtest.** Answers the question "what would have happened last month if the ATR multiplier had been 2 instead of 1.5" using historical data; stop settings are chosen with data rather than guesswork. It suits the Rust side and is the item that takes the most effort.
- **Discipline rules.** Rules the user sets for themselves (example: "no trading for 1 hour after a loss", "no orders after 00:00", "a stop cannot be widened in the loss direction"). The system either prevents the rule from being broken or at least records the violation.
- **Multiple exchanges.** A second exchange is added alongside Binance and all of them are combined in a single portfolio. It may also be needed for tokenized stocks.
- **Smart alerts.** In addition to v1's price alert, alerts for volatility spikes, approaching the stop, drift from the target allocation, and depeg. Push on mobile, instant notification on web.
- **DCA (recurring buys).** Plans such as "buy 50 dollars of BTC every week" are defined. It separates the long-term part from the trading part.
- **Weekly report.** An automatic summary is produced at the end of the week (profit/loss, broken rules, best and worst trade).
- **Tax/profit statement.** An annual realized-profit report and export are provided.
- **Futures.** Leverage, liquidation price and funding cost are added. It greatly enlarges the risk engine; it will not be entered until spot works solidly.
- **AI commentary.** An assistant that draws observations from trade history (example: "most losses are in small-cap altcoins, in trades opened at night").

## 9. Release tiers

Being written on the list does not mean it will be done. Rule: work on the second release does not start until the first release is finished.

Note: This section determines which release the priority notes in the previous sections fall into.

### 9.1 First release

- ATR-based stop and coin classes (section 2)
- Total open risk limit and loss brake (section 3)
- Portfolio tracking: historical performance, trade statistics, cost tracking (section 5.4)
- Pre-trade check screen
- Discipline rules
- Paper trading
- Market service: coin list, candle data, live price feed (section 6; prerequisite for the ATR calculation)
- Two-factor authentication (section 6)

### 9.2 Second release

- Liquidity check, correlation, market regime (sections 2 and 3)
- Stop types: trailing, break-even, scaled take-profit, time stop (section 4)
- Target allocation, drift alert, the cash role of stablecoin, rebalancing suggestion (sections 5.1, 5.2)
- Backtest
- Smart alerts and notification system
- Weekly report
- DCA (recurring buys)

### 9.3 Idea pool (no commitment)

- Multiple exchanges
- Tokenized stocks (section 5.3)
- Tax/profit statement
- Futures
- AI commentary

package portfolio

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	"github.com/batuhangozegu/metron/api/internal/risk"
	"github.com/batuhangozegu/metron/api/internal/trading"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/shopspring/decimal"
)

type Postgres struct{ Pool *pgxpool.Pool }

type account struct {
	market, currency  string
	balance, starting decimal.Decimal
}

func loadAccount(ctx context.Context, tx pgx.Tx, owner *string, id string, lock bool) (account, error) {
	query := `SELECT market,quote_asset,paper_balance::text,COALESCE(starting_balance,paper_balance)::text
		FROM app.accounts WHERE id=$1::uuid AND ($2::uuid IS NULL OR user_id=$2::uuid)
		AND type='paper'`
	if owner != nil {
		query += ` AND quote_asset='USDT'`
	}
	if lock {
		query += ` FOR UPDATE`
	}
	var a account
	var balance, starting string
	err := tx.QueryRow(ctx, query, id, owner).Scan(&a.market, &a.currency, &balance, &starting)
	if errors.Is(err, pgx.ErrNoRows) {
		return a, risk.ErrAccountNotFound
	}
	if err != nil {
		return a, err
	}
	a.balance, err = decimal.NewFromString(balance)
	if err != nil {
		return a, err
	}
	a.starting, err = decimal.NewFromString(starting)
	return a, err
}

func loadOpen(ctx context.Context, tx pgx.Tx, id string) ([]openPosition, error) {
	rows, err := tx.Query(ctx, `SELECT p.symbol,p.side,p.quantity::text,p.entry_price::text,COALESCE(p.leverage,1),
		COALESCE((SELECT sum(f.fee) FROM app.orders o JOIN app.fills f ON f.order_id=o.id
		WHERE o.position_id=p.id AND o.account_id=p.account_id AND o.purpose='open' AND o.status='filled'),0)::text
		FROM app.positions p WHERE p.account_id=$1::uuid AND p.status='open'`, id)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	positions := []openPosition{}
	for rows.Next() {
		var p openPosition
		var quantity, entry, fee string
		if err := rows.Scan(&p.Symbol, &p.Side, &quantity, &entry, &p.Leverage, &fee); err != nil {
			return nil, err
		}
		values := []*decimal.Decimal{&p.Quantity, &p.Entry, &p.Fee}
		for i, v := range []string{quantity, entry, fee} {
			value, err := decimal.NewFromString(v)
			if err != nil {
				return nil, err
			}
			*values[i] = value
		}
		positions = append(positions, p)
	}
	return positions, rows.Err()
}

func (s Service) value(ctx context.Context, positions []openPosition) (decimal.Decimal, error) {
	pnl := decimal.Zero
	prices := map[string]decimal.Decimal{}
	for _, p := range positions {
		price, ok := prices[p.Symbol]
		if !ok {
			if s.Prices == nil {
				return decimal.Zero, trading.ErrPrice
			}
			var err error
			price, err = s.Prices.Price(ctx, p.Symbol)
			if err != nil {
				return decimal.Zero, err
			}
			prices[p.Symbol] = price
		}
		delta := price.Sub(p.Entry).Mul(p.Quantity)
		if p.Side == "short" {
			delta = delta.Neg()
		}
		pnl = pnl.Add(delta)
	}
	return pnl, nil
}

func periodStart(period string, now time.Time) *time.Time {
	var days int
	switch period {
	case "30d":
		days = 30
	case "90d":
		days = 90
	default:
		return nil
	}
	start := now.AddDate(0, 0, -days)
	return &start
}

func (s Service) Portfolio(ctx context.Context, owner, id, period string) (Portfolio, error) {
	var result Portfolio
	tx, err := s.Store.Pool.BeginTx(ctx, pgx.TxOptions{IsoLevel: pgx.RepeatableRead, AccessMode: pgx.ReadOnly})
	if err != nil {
		return result, err
	}
	defer tx.Rollback(context.WithoutCancel(ctx))
	a, err := loadAccount(ctx, tx, &owner, id, false)
	if err != nil {
		return result, err
	}
	positions, err := loadOpen(ctx, tx, id)
	if err != nil {
		return result, err
	}
	unrealized, err := s.value(ctx, positions)
	if err != nil {
		return result, err
	}
	now := s.now()
	start := periodStart(period, now)
	equity := a.balance.Add(unrealized)
	capital := make([]trading.Capital, 0, len(positions))
	for _, p := range positions {
		capital = append(capital, trading.Capital{Notional: p.Quantity.Mul(p.Entry), Fee: p.Fee, Leverage: p.Leverage})
	}
	var fees string
	err = tx.QueryRow(ctx, `SELECT COALESCE(sum(f.fee),0)::text FROM app.fills f JOIN app.orders o ON o.id=f.order_id
		WHERE o.account_id=$1::uuid AND ($2::timestamptz IS NULL OR f.created_at >= $2) AND f.created_at <= $3`, id, start, now).Scan(&fees)
	if err != nil {
		return result, err
	}
	fee, err := decimal.NewFromString(fees)
	if err != nil {
		return result, err
	}
	result = Portfolio{AccountID: id, Market: a.market, Currency: a.currency, Period: period,
		Summary: Summary{StartingBalance: a.starting.String(), Balance: a.balance.String(), Cash: trading.Cash(a.balance, a.market, capital).String(),
			UnrealizedPNL: unrealized.String(), Equity: equity.String(), NetResult: equity.Sub(a.starting).String(),
			ReturnPercent: ratio(equity.Sub(a.starting).Mul(decimal.NewFromInt(100)), a.starting, 2), FeesPaid: fee.String()},
		EquitySeries: []Point{}, BenchmarkSeries: []Point{}, Holdings: holdings(positions)}
	rows, err := tx.Query(ctx, `SELECT taken_at,equity::text FROM app.equity_snapshots
		WHERE account_id=$1::uuid AND ($2::timestamptz IS NULL OR taken_at >= $2) AND taken_at <= $3 ORDER BY taken_at`, id, start, now)
	if err != nil {
		return result, err
	}
	for rows.Next() {
		var p Point
		if err := rows.Scan(&p.At, &p.Equity); err != nil {
			rows.Close()
			return result, err
		}
		p.At = p.At.UTC()
		v, err := decimal.NewFromString(p.Equity)
		if err != nil {
			rows.Close()
			return result, err
		}
		p.Equity = v.String()
		result.EquitySeries = append(result.EquitySeries, p)
	}
	err = rows.Err()
	rows.Close()
	if err != nil {
		return result, err
	}
	// Replace an observation at exactly now rather than emitting duplicate times.
	if n := len(result.EquitySeries); n > 0 && result.EquitySeries[n-1].At.Equal(now) {
		result.EquitySeries = result.EquitySeries[:n-1]
	}
	result.EquitySeries = downsample(append(result.EquitySeries, Point{At: now, Equity: equity.String()}), 200)
	result.MaxDrawdown = drawdown(result.EquitySeries)
	rows, err = tx.Query(ctx, `SELECT coin_class,realized_pnl::text FROM app.positions
		WHERE account_id=$1::uuid AND status='closed' AND ($2::timestamptz IS NULL OR closed_at >= $2) AND closed_at <= $3`, id, start, now)
	if err != nil {
		return result, err
	}
	closed := []closedPosition{}
	for rows.Next() {
		var p closedPosition
		var pnl string
		if err := rows.Scan(&p.CoinClass, &pnl); err != nil {
			rows.Close()
			return result, err
		}
		p.NetPNL, err = decimal.NewFromString(pnl)
		if err != nil {
			rows.Close()
			return result, err
		}
		closed = append(closed, p)
	}
	err = rows.Err()
	rows.Close()
	if err != nil {
		return result, err
	}
	result.Stats = statistics(closed)
	first := result.EquitySeries[0]
	// Anchor the benchmark at the first equity point with the latest closed
	// candle (hourly preferred), then daily closes, then the live price.
	rows, err = tx.Query(ctx, `(SELECT $1::timestamptz, close::text FROM market.candles
			WHERE symbol='BTCUSDT' AND close_time <= $1 ORDER BY close_time DESC, interval='1h' DESC LIMIT 1)
		UNION ALL
		(SELECT close_time,close::text FROM market.candles
			WHERE symbol='BTCUSDT' AND interval='1d' AND close_time > $1 AND close_time <= $2 ORDER BY close_time)`, first.At, now)
	if err != nil {
		return result, err
	}
	candles := []Point{}
	for rows.Next() {
		var at time.Time
		var text string
		if err := rows.Scan(&at, &text); err != nil {
			rows.Close()
			return result, err
		}
		price, err := decimal.NewFromString(text)
		if err != nil {
			rows.Close()
			return result, err
		}
		if !price.IsPositive() {
			continue
		}
		candles = append(candles, Point{At: at.UTC(), Equity: price.String()})
	}
	err = rows.Err()
	rows.Close()
	if err != nil {
		return result, err
	}
	if len(candles) > 0 && s.Prices != nil {
		if live, err := s.Prices.Price(ctx, "BTCUSDT"); err == nil && live.IsPositive() && candles[len(candles)-1].At.Before(now) {
			candles = append(candles, Point{At: now, Equity: live.String()})
		}
	}
	result.BenchmarkSeries, result.BenchmarkReturnPercent = benchmark(candles, decimal.RequireFromString(first.Equity))
	return result, tx.Commit(ctx)
}

func (p Postgres) Journal(ctx context.Context, owner, id string, limit int, before *time.Time) (Journal, error) {
	result := Journal{Trades: []Trade{}}
	tx, err := p.Pool.BeginTx(ctx, pgx.TxOptions{IsoLevel: pgx.RepeatableRead, AccessMode: pgx.ReadOnly})
	if err != nil {
		return result, err
	}
	defer tx.Rollback(context.WithoutCancel(ctx))
	if _, err := loadAccount(ctx, tx, &owner, id, false); err != nil {
		return result, err
	}
	rows, err := tx.Query(ctx, `SELECT p.id::text,p.symbol,p.side,p.coin_class,p.opened_at,p.closed_at,p.close_reason,
		p.quantity::text,p.entry_price::text,p.exit_price::text,p.realized_pnl::text,p.initial_risk::text,
		COALESCE((SELECT sum(f.fee) FROM app.orders o JOIN app.fills f ON f.order_id=o.id
		WHERE o.position_id=p.id AND o.account_id=p.account_id AND o.status='filled'),0)::text,
		COALESCE((SELECT c.result->'checks' FROM app.orders o JOIN app.risk_checks c ON c.order_id=o.id AND c.account_id=o.account_id
		WHERE o.position_id=p.id AND o.account_id=p.account_id AND o.purpose='open' AND o.status='filled' AND c.source='order'
		ORDER BY c.created_at,c.id LIMIT 1),'[]'::jsonb)
		FROM app.positions p WHERE p.account_id=$1::uuid AND p.status='closed'
		AND ($2::timestamptz IS NULL OR p.closed_at < $2) ORDER BY p.closed_at DESC,p.id DESC LIMIT $3`, id, before, limit+1)
	if err != nil {
		return result, err
	}
	defer rows.Close()
	for rows.Next() {
		var t Trade
		var initial string
		var checks []byte
		if err := rows.Scan(&t.PositionID, &t.Symbol, &t.Side, &t.CoinClass, &t.OpenedAt, &t.ClosedAt, &t.CloseReason,
			&t.Quantity, &t.EntryPrice, &t.ExitPrice, &t.NetPNL, &initial, &t.Fees, &checks); err != nil {
			return result, err
		}
		t.OpenedAt, t.ClosedAt = t.OpenedAt.UTC(), t.ClosedAt.UTC()
		for _, value := range []*string{&t.Quantity, &t.EntryPrice, &t.ExitPrice, &t.NetPNL, &t.Fees} {
			v, err := decimal.NewFromString(*value)
			if err != nil {
				return result, err
			}
			*value = v.String()
		}
		riskAmount, err := decimal.NewFromString(initial)
		if err != nil {
			return result, err
		}
		t.RMultiple = rMultiple(decimal.RequireFromString(t.NetPNL), riskAmount)
		t.Rules = []Rule{}
		var rules []struct {
			RuleID string `json:"rule_id"`
			Status string `json:"status"`
		}
		if err := json.Unmarshal(checks, &rules); err != nil {
			return result, err
		}
		for _, r := range rules {
			if r.Status == "caution" || r.Status == "fail" {
				t.Rules = append(t.Rules, Rule{Rule: r.RuleID, Status: r.Status})
			}
		}
		result.Trades = append(result.Trades, t)
	}
	if err := rows.Err(); err != nil {
		return result, err
	}
	rows.Close()
	if len(result.Trades) > limit {
		result.Trades = result.Trades[:limit]
		at := result.Trades[limit-1].ClosedAt
		result.NextBefore = &at
	}
	return result, tx.Commit(ctx)
}

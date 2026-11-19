package risk

import (
	"context"
	"errors"
	"time"

	enginev1 "github.com/batuhangozegu/metron/api/internal/gen/metron/engine/v1"
	"github.com/jackc/pgx/v5"
	"github.com/shopspring/decimal"
)

type ActivityPosition struct {
	Status, CoinClass, CloseReason                 string
	Quantity, EntryPrice, InitialRisk, RealizedPNL decimal.Decimal
	ClosedAt                                       time.Time
}

func UTCWindows(now time.Time) (day, week time.Time) {
	now = now.UTC()
	day = time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.UTC)
	week = day.AddDate(0, 0, -(int(day.Weekday())+6)%7)
	return
}

// AggregateActivity uses settled net PnL, so profitable trades offset losses.
// Equity is the account balance; open entry fees settle when the position closes.
func AggregateActivity(positions []ActivityPosition, equity decimal.Decimal, coinClass, market string, now time.Time) (*enginev1.Activity, string) {
	day, week := UTCWindows(now)
	a := &enginev1.Activity{}
	openRisk, daily, weekly, holding := decimal.Zero, decimal.Zero, decimal.Zero, decimal.Zero
	for _, p := range positions {
		if p.Status == "open" {
			a.OpenPositions++
			openRisk = openRisk.Add(p.InitialRisk)
			if market == "spot" && p.CoinClass == coinClass {
				holding = holding.Add(p.Quantity.Mul(p.EntryPrice))
			}
			continue
		}
		if p.Status != "closed" || p.ClosedAt.Before(week) || p.ClosedAt.After(now) {
			continue
		}
		weekly = weekly.Add(p.RealizedPNL)
		if p.CloseReason == "stop" {
			a.StoppedTradesThisWeek++
		}
		if !p.ClosedAt.Before(day) {
			daily = daily.Add(p.RealizedPNL)
			if p.RealizedPNL.IsNegative() {
				a.LosingTradesToday++
			}
		}
	}
	percent := func(value decimal.Decimal) string {
		if !equity.IsPositive() {
			return "0"
		} // The engine blocks non-positive equity.
		// The engine accepts at most 8 decimals; round up so limits stay conservative.
		return value.Mul(decimal.NewFromInt(100)).DivRound(equity, 28).RoundCeil(8).String()
	}
	a.OpenRiskPercent = percent(openRisk)
	a.DailyLossPercent = percent(decimal.Max(decimal.Zero, daily.Neg()))
	a.WeeklyLossPercent = percent(decimal.Max(decimal.Zero, weekly.Neg()))
	return a, holding.RoundCeil(8).String()
}

func LoadActivity(ctx context.Context, q Querier, accountID, symbol string, account Account, now time.Time) (Account, error) {
	var coinClass string
	err := q.QueryRow(ctx, `SELECT coin_class FROM market.symbols WHERE symbol=$1`, symbol).Scan(&coinClass)
	if err != nil && !errors.Is(err, pgx.ErrNoRows) {
		return Account{}, err
	}
	_, week := UTCWindows(now)
	rows, err := q.Query(ctx, `SELECT status,coin_class,quantity::text,entry_price::text,initial_risk::text,
		COALESCE(realized_pnl,0)::text,COALESCE(close_reason,''),closed_at
		FROM app.positions WHERE account_id=$1::uuid AND
		(status='open' OR (status='closed' AND closed_at >= $2 AND closed_at <= $3))`, accountID, week, now.UTC())
	if err != nil {
		return Account{}, err
	}
	defer rows.Close()
	var positions []ActivityPosition
	for rows.Next() {
		var p ActivityPosition
		var quantity, entry, risk, pnl string
		var closed *time.Time
		if err := rows.Scan(&p.Status, &p.CoinClass, &quantity, &entry, &risk, &pnl, &p.CloseReason, &closed); err != nil {
			return Account{}, err
		}
		values := []*decimal.Decimal{&p.Quantity, &p.EntryPrice, &p.InitialRisk, &p.RealizedPNL}
		for i, value := range []string{quantity, entry, risk, pnl} {
			*values[i], err = decimal.NewFromString(value)
			if err != nil {
				return Account{}, err
			}
		}
		if closed != nil {
			p.ClosedAt = *closed
		}
		positions = append(positions, p)
	}
	if err := rows.Err(); err != nil {
		return Account{}, err
	}
	equity, err := decimal.NewFromString(account.Equity)
	if err != nil {
		return Account{}, err
	}
	account.Activity, account.CurrentClassHolding = AggregateActivity(positions, equity, coinClass, account.Market, now)
	// Settled balances can carry fee/PnL precision beyond the engine's 8 decimals;
	// round down so sizing never uses more equity than the account holds.
	account.Equity = equity.RoundFloor(8).String()
	return account, nil
}

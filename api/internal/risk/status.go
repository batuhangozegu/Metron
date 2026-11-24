package risk

import (
	"context"
	"net/http"
	"strings"
	"time"

	"github.com/batuhangozegu/metron/api/internal/auth"
	"github.com/batuhangozegu/metron/api/internal/httpjson"
	"github.com/go-chi/chi/v5"
	"github.com/shopspring/decimal"
)

type Prices interface {
	Price(context.Context, string) (decimal.Decimal, error)
}

type CheckCounts struct {
	Run     int64 `json:"run"`
	Orders  int64 `json:"orders"`
	Blocked int64 `json:"blocked"`
}

type StatusStore interface {
	ChecksToday(context.Context, string, string, time.Time) (CheckCounts, error)
}

type AmountStatus struct {
	Amount       string `json:"amount"`
	Percent      string `json:"percent"`
	LimitPercent string `json:"limit_percent"`
}

type CountStatus struct {
	Count uint32 `json:"count"`
	Limit uint32 `json:"limit"`
}

type Status struct {
	AccountID             string       `json:"account_id"`
	Market                string       `json:"market"`
	AsOf                  time.Time    `json:"as_of"`
	Currency              string       `json:"currency"`
	Equity                string       `json:"equity"`
	RealizedToday         string       `json:"realized_today"`
	Unrealized            string       `json:"unrealized"`
	OpenRisk              AmountStatus `json:"open_risk"`
	DailyLoss             AmountStatus `json:"daily_loss"`
	WeeklyLoss            AmountStatus `json:"weekly_loss"`
	OpenPositions         CountStatus  `json:"open_positions"`
	LosingTradesToday     CountStatus  `json:"losing_trades_today"`
	StoppedTradesThisWeek CountStatus  `json:"stopped_trades_this_week"`
	ChecksToday           CheckCounts  `json:"checks_today"`
}

func statusFromAccount(id string, account Account, unrealized decimal.Decimal, checks CheckCounts) Status {
	p := ResolveProfile(account.Profile, nil)
	a, snapshot := account.Activity, account.Snapshot
	return Status{
		AccountID: id, Market: account.Market, AsOf: snapshot.AsOf, Currency: "USDT",
		Equity: account.Equity, RealizedToday: snapshot.DailyPNL.String(), Unrealized: unrealized.String(),
		OpenRisk:              AmountStatus{snapshot.OpenRisk.String(), a.OpenRiskPercent, p.TotalOpenRiskLimitPercent},
		DailyLoss:             AmountStatus{decimal.Max(decimal.Zero, snapshot.DailyPNL.Neg()).String(), a.DailyLossPercent, p.DailyLossLimitPercent},
		WeeklyLoss:            AmountStatus{decimal.Max(decimal.Zero, snapshot.WeeklyPNL.Neg()).String(), a.WeeklyLossPercent, p.WeeklyLossLimitPercent},
		OpenPositions:         CountStatus{a.OpenPositions, *p.MaxOpenPositions},
		LosingTradesToday:     CountStatus{a.LosingTradesToday, *p.MaxLosingTradesToday},
		StoppedTradesThisWeek: CountStatus{a.StoppedTradesThisWeek, *p.MaxStoppedTradesThisWeek},
		ChecksToday:           checks,
	}
}

func (s Service) unrealized(ctx context.Context, positions []ActivityPosition) decimal.Decimal {
	total := decimal.Zero
	if s.Prices == nil {
		return total
	}
	// Use one observation per symbol, including when several positions share it.
	prices := map[string]decimal.Decimal{}
	for _, p := range positions {
		symbol := p.Symbol
		if !strings.HasSuffix(symbol, "USDT") {
			symbol += "USDT"
		}
		price, ok := prices[symbol]
		if !ok {
			var err error
			price, err = s.Prices.Price(ctx, symbol)
			if err != nil {
				price = decimal.Zero
			}
			prices[symbol] = price
		}
		if !price.IsPositive() {
			continue
		}
		pnl := price.Sub(p.EntryPrice).Mul(p.Quantity)
		if p.Side == "short" {
			pnl = pnl.Neg()
		}
		total = total.Add(pnl)
	}
	return total
}

func (s Service) statusHTTP(w http.ResponseWriter, r *http.Request) {
	owner, id := auth.UserFromContext(r.Context()).ID, chi.URLParam(r, "id")
	if !ValidID(id) {
		storeFailure(w, ErrAccountNotFound)
		return
	}
	store, ok := s.Store.(StatusStore)
	if s.Accounts == nil || !ok {
		storeFailure(w, nil)
		return
	}
	ctx, cancel := context.WithTimeout(r.Context(), 4*time.Second)
	defer cancel()
	// An empty symbol requests account-wide activity without class holdings.
	// The configured RiskAccounts also supplies live exchange quote equity.
	account, err := s.Accounts.RiskAccount(ctx, owner, id, "")
	if err != nil {
		storeFailure(w, err)
		return
	}
	if account.Activity == nil || account.Snapshot == nil {
		storeFailure(w, nil)
		return
	}
	checks, err := store.ChecksToday(ctx, owner, id, account.Snapshot.AsOf)
	if err != nil {
		storeFailure(w, err)
		return
	}
	value := statusFromAccount(id, account, s.unrealized(ctx, account.Snapshot.Positions), checks)
	httpjson.Write(w, http.StatusOK, value)
}

func (p Postgres) ChecksToday(ctx context.Context, owner, id string, now time.Time) (CheckCounts, error) {
	if err := ownedAccount(ctx, p.Pool, owner, id, false); err != nil {
		return CheckCounts{}, err
	}
	day, _ := UTCWindows(now)
	var counts CheckCounts
	err := p.Pool.QueryRow(ctx, `SELECT count(*),count(*) FILTER (WHERE order_id IS NOT NULL),
		count(*) FILTER (WHERE NOT allowed) FROM app.risk_checks
		WHERE account_id=$1::uuid AND created_at >= $2 AND created_at <= $3`, id, day, now.UTC()).
		Scan(&counts.Run, &counts.Orders, &counts.Blocked)
	return counts, err
}

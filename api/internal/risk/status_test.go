package risk

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"net/http"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/batuhangozegu/metron/api/internal/auth"
	"github.com/batuhangozegu/metron/api/internal/cryptobox"
	"github.com/go-chi/chi/v5"
	"github.com/shopspring/decimal"
)

func statusAccount(equity string, positions []ActivityPosition, now time.Time) Account {
	a, holding, snapshot := aggregateActivity(positions, decimal.RequireFromString(equity), "", "futures", now)
	return Account{Equity: equity, Market: "futures", Activity: a, CurrentClassHolding: holding, Snapshot: snapshot}
}

func TestStatusCalculationsAndLimits(t *testing.T) {
	now := time.Date(2026, 11, 18, 1, 0, 0, 0, time.FixedZone("Istanbul", 3*60*60))
	day, week := UTCWindows(now)
	positions := []ActivityPosition{
		{Status: "open", InitialRisk: decimal.RequireFromString("1.000000001")},
		{Status: "closed", ClosedAt: week, RealizedPNL: decimal.NewFromInt(-2), CloseReason: "stop"},
		{Status: "closed", ClosedAt: day, RealizedPNL: decimal.NewFromInt(-2), CloseReason: "stop"},
		{Status: "closed", ClosedAt: now, RealizedPNL: decimal.NewFromInt(1)},
		{Status: "closed", ClosedAt: week.Add(-time.Nanosecond), RealizedPNL: decimal.NewFromInt(-999)},
		{Status: "closed", ClosedAt: now.Add(time.Nanosecond), RealizedPNL: decimal.NewFromInt(-999)},
	}
	a := statusAccount("3", positions, now)
	for _, saved := range []bool{false, true} {
		t.Run(map[bool]string{false: "defaults", true: "saved"}[saved], func(t *testing.T) {
			p := DefaultProfile()
			if saved {
				p.TotalOpenRiskLimitPercent, p.DailyLossLimitPercent, p.WeeklyLossLimitPercent = "4.25", "1.5", "7.75"
				p.MaxOpenPositions, p.MaxLosingTradesToday, p.MaxStoppedTradesThisWeek = count(1), count(2), count(^uint32(0))
				a.Profile = &p
			}
			got := statusFromAccount(accountID, a, decimal.RequireFromString("-3.000000001"), CheckCounts{4, 2, 1})
			if got.AccountID != accountID || got.Market != "futures" || got.Currency != "USDT" || got.Equity != "3" || got.AsOf.Location() != time.UTC || !got.AsOf.Equal(now) || got.RealizedToday != "-1" || got.Unrealized != "-3.000000001" {
				t.Fatalf("wrong summary: %+v", got)
			}
			// Percentages use the engine's conservative ceil to 8 places and
			// may exceed limits or 100%; amounts keep their original precision.
			if got.OpenRisk != (AmountStatus{"1.000000001", "33.33333337", p.TotalOpenRiskLimitPercent}) ||
				got.DailyLoss != (AmountStatus{"1", "33.33333334", p.DailyLossLimitPercent}) ||
				got.WeeklyLoss != (AmountStatus{"3", "100", p.WeeklyLossLimitPercent}) ||
				got.OpenPositions != (CountStatus{1, *p.MaxOpenPositions}) ||
				got.LosingTradesToday != (CountStatus{1, *p.MaxLosingTradesToday}) ||
				got.StoppedTradesThisWeek != (CountStatus{2, *p.MaxStoppedTradesThisWeek}) || got.ChecksToday != (CheckCounts{4, 2, 1}) {
				t.Fatalf("wrong activity/limits: %+v", got)
			}
		})
	}
	for _, equity := range []string{"0", "-1", "3"} {
		a := statusAccount(equity, []ActivityPosition{{Status: "closed", ClosedAt: now, RealizedPNL: decimal.NewFromInt(1)}}, now)
		got := statusFromAccount(accountID, a, decimal.Zero, CheckCounts{})
		if got.RealizedToday != "1" || got.Unrealized != "0" || got.DailyLoss.Amount != "0" || got.WeeklyLoss.Amount != "0" || got.DailyLoss.Percent != "0" || got.WeeklyLoss.Percent != "0" || got.OpenRisk.Percent != "0" {
			t.Fatalf("profit/zero equity: %+v", got)
		}
	}
	a = statusAccount("0", positions, now)
	got := statusFromAccount(accountID, a, decimal.Zero, CheckCounts{})
	if got.OpenRisk.Percent != "0" || got.DailyLoss.Percent != "0" || got.WeeklyLoss.Percent != "0" || got.DailyLoss.Amount != "1" {
		t.Fatalf("zero equity with losses: %+v", got)
	}
	got = statusFromAccount(accountID, statusAccount("0.5", positions, now), decimal.Zero, CheckCounts{})
	if got.DailyLoss.Percent != "200" {
		t.Fatalf("percentage was clamped: %+v", got)
	}
}

type statusPrices struct {
	calls map[string]int
}

func (p statusPrices) Price(_ context.Context, symbol string) (decimal.Decimal, error) {
	p.calls[symbol]++
	if symbol == "BTCUSDT" {
		return decimal.RequireFromString("110.25"), nil
	}
	if symbol == "ZEROUSDT" {
		return decimal.Zero, nil
	}
	return decimal.Zero, errors.New("price unavailable")
}

func TestStatusUnrealizedSkipsMissingPrices(t *testing.T) {
	position := func(symbol, side, quantity string) ActivityPosition {
		return ActivityPosition{Symbol: symbol, Side: side, EntryPrice: decimal.NewFromInt(100), Quantity: decimal.RequireFromString(quantity)}
	}
	positions := []ActivityPosition{position("BTCUSDT", "long", "2"), position("BTC", "short", "0.5"), position("MISSINGUSDT", "long", "9"), position("MISSINGUSDT", "short", "1"), position("ZEROUSDT", "long", "1")}
	prices := statusPrices{calls: map[string]int{}}
	got := (Service{Prices: prices}).unrealized(context.Background(), positions)
	if got.String() != "15.375" || !reflect.DeepEqual(prices.calls, map[string]int{"BTCUSDT": 1, "MISSINGUSDT": 1, "ZEROUSDT": 1}) {
		t.Fatalf("unrealized=%s calls=%v", got, prices.calls)
	}
	if got := (Service{}).unrealized(context.Background(), positions); !got.IsZero() {
		t.Fatalf("no prices: %s", got)
	}
}

type statusAccounts struct {
	*fakeAccounts
	value       Account
	counts      CheckCounts
	countsErr   error
	countsCalls int
	asOf        time.Time
}

func (s *statusAccounts) RiskAccount(ctx context.Context, owner, id, symbol string) (Account, error) {
	if _, err := s.fakeAccounts.RiskAccount(ctx, owner, id, symbol); err != nil {
		return Account{}, err
	}
	return s.value, nil
}

func (s *statusAccounts) ChecksToday(_ context.Context, _, _ string, now time.Time) (CheckCounts, error) {
	s.countsCalls++
	s.asOf = now
	return s.counts, s.countsErr
}

func statusRouter(t *testing.T, service Service, session auth.Store) http.Handler {
	t.Helper()
	box, err := cryptobox.New(base64.StdEncoding.EncodeToString(make([]byte, 32)))
	if err != nil {
		t.Fatal(err)
	}
	r := chi.NewRouter()
	auth.HTTP{Service: &auth.Service{Store: session, Box: box, Now: time.Now}}.Mount(r, service.Mount)
	return r
}

func TestStatusHTTP(t *testing.T) {
	s := &statusAccounts{fakeAccounts: &fakeAccounts{}, value: statusAccount("10000", nil, time.Now()), counts: CheckCounts{4, 2, 1}}
	h := statusRouter(t, Service{Accounts: s, Store: s}, fakeAuth{}) // No engine required.
	w := profileRequest(h, "GET", "/v1/accounts/"+accountID+"/risk-status", "", token)
	var got Status
	if err := json.Unmarshal(w.Body.Bytes(), &got); err != nil || w.Code != 200 || got.Equity != "10000" || got.ChecksToday != s.counts || got.RealizedToday != "0" || got.Unrealized != "0" || !s.asOf.Equal(s.value.Snapshot.AsOf) {
		t.Fatalf("status: %d %s %v", w.Code, w.Body, err)
	}
	if w.Header().Get("Cache-Control") != "no-store" || !strings.Contains(w.Body.String(), `"open_risk":{"amount":"0","percent":"0","limit_percent":"6"}`) || !strings.Contains(w.Body.String(), `"open_positions":{"count":0,"limit":5}`) {
		t.Fatalf("wire format: %s", w.Body)
	}
}

func TestStatusOwnershipAndUnavailable(t *testing.T) {
	for _, tc := range []struct {
		name, id, cookie      string
		accountErr, countsErr error
		code                  int
	}{
		{"foreign", foreignID, token, nil, nil, 404},
		{"unknown", "33333333-3333-4333-8333-333333333333", token, nil, nil, 404},
		{"malformed", "not-a-uuid", token, nil, nil, 404},
		{"session", accountID, "", nil, nil, 401},
		{"exchange", accountID, token, errors.New("private exchange detail"), nil, 503},
		{"checks", accountID, token, nil, errors.New("private database detail"), 503},
	} {
		t.Run(tc.name, func(t *testing.T) {
			s := &statusAccounts{fakeAccounts: &fakeAccounts{err: tc.accountErr}, value: statusAccount("1000", nil, time.Now()), countsErr: tc.countsErr}
			h := statusRouter(t, Service{Accounts: s, Store: s}, fakeAuth{})
			w := profileRequest(h, "GET", "/v1/accounts/"+tc.id+"/risk-status", "", tc.cookie)
			if w.Code != tc.code || !strings.Contains(w.Body.String(), "error_tr") || strings.Contains(w.Body.String(), "private") {
				t.Fatalf("%d %s", w.Code, w.Body)
			}
			if tc.code == 404 && (!strings.Contains(w.Body.String(), `"code":"account_not_found"`) || s.countsCalls != 0) {
				t.Fatalf("ownership: %s counts calls=%d", w.Body, s.countsCalls)
			}
		})
	}
}

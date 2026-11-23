package portfolio

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/batuhangozegu/metron/api/internal/accounts"
	"github.com/batuhangozegu/metron/api/internal/auth"
	"github.com/batuhangozegu/metron/api/internal/cryptobox"
	enginev1 "github.com/batuhangozegu/metron/api/internal/gen/metron/engine/v1"
	"github.com/batuhangozegu/metron/api/internal/risk"
	"github.com/batuhangozegu/metron/api/internal/trading"
	"github.com/go-chi/chi/v5"
	"github.com/shopspring/decimal"
	"google.golang.org/grpc"
)

type journalEngine struct{}

func (journalEngine) Check(_ context.Context, r *enginev1.CheckRequest, _ ...grpc.CallOption) (*enginev1.CheckResponse, error) {
	return &enginev1.CheckResponse{Symbol: r.Symbol, Side: r.Side, EntryPrice: r.EntryPrice, PositionSize: "1", Stop: "95", CoinClass: "core", Market: r.Market, Leverage: max(1, r.GetLeverage()), Allowed: true,
		Checks: []*enginev1.RuleCheck{{RuleId: "class_allocation", Status: "caution"}, {RuleId: "atr_stop", Status: "pass"}}}, nil
}

// Uses a migrated dedicated database; only this test's user is deleted.
func TestPostgresPortfolioSnapshotsJournalAndOwnership(t *testing.T) {
	pool := testPool(t)
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	var owner string
	if err := pool.QueryRow(ctx, `INSERT INTO app.users(email,password_hash) VALUES(gen_random_uuid()::text || '@portfolio.example.com','test-only') RETURNING id::text`).Scan(&owner); err != nil {
		t.Fatal(err)
	}
	defer func() {
		if _, err := pool.Exec(context.Background(), `DELETE FROM app.users WHERE id=$1::uuid`, owner); err != nil {
			t.Error(err)
		}
	}()
	a, err := (accounts.Postgres{Pool: pool}).CreatePaper(ctx, owner, "Portfolio", "spot", decimal.NewFromInt(1000))
	if err != nil {
		t.Fatal(err)
	}
	var starting string
	if err := pool.QueryRow(ctx, `SELECT starting_balance::text FROM app.accounts WHERE id=$1::uuid`, a.ID).Scan(&starting); err != nil || starting != "1000" {
		t.Fatalf("creation starting balance %s %v", starting, err)
	}
	now := time.Now().UTC().Truncate(time.Microsecond).Add(-time.Hour)
	prices := fixedPrices{price: decimal.NewFromInt(100)}
	s := Service{Store: &Postgres{Pool: pool}, Prices: prices, Now: func() time.Time { return now }}
	trade := trading.Service{Store: trading.Postgres{Pool: pool}, Exchange: trading.Paper{Tickers: portfolioTickers{}}, Risk: risk.Service{Engine: journalEngine{}}, Snapshots: s, Now: func() time.Time { return now }}
	// Use a pointer so snapshot timestamps advance with each trade as well.
	trade.Snapshots = &s
	open, err := trade.Open(ctx, owner, trading.Request{AccountID: a.ID, Symbol: "BTCUSDT", Side: "long"})
	if err != nil {
		t.Fatal(err)
	}
	now = now.Add(time.Minute)
	if _, err := trade.Close(ctx, owner, open.Position.ID); err != nil {
		t.Fatal(err)
	}
	now = now.Add(time.Minute)
	second, err := trade.Open(ctx, owner, trading.Request{AccountID: a.ID, Symbol: "BTCUSDT", Side: "long"})
	if err != nil {
		t.Fatal(err)
	}
	now = now.Add(time.Minute)
	if _, err := trade.Close(ctx, owner, second.Position.ID); err != nil {
		t.Fatal(err)
	}
	var count int
	if err := pool.QueryRow(ctx, `SELECT count(*) FROM app.equity_snapshots WHERE account_id=$1::uuid`, a.ID).Scan(&count); err != nil || count != 4 {
		t.Fatalf("trade snapshots %d %v", count, err)
	}
	value, err := s.Portfolio(ctx, owner, a.ID, "30d")
	if err != nil || value.Summary.Balance != "999.6" || value.Summary.FeesPaid != "0.4" || value.Summary.Cash != "999.6" || value.Stats.Trades != 2 || len(value.EquitySeries) != 4 {
		t.Fatalf("portfolio: %+v %v", value, err)
	}
	page, err := s.Store.Journal(ctx, owner, a.ID, 1, nil)
	if err != nil || len(page.Trades) != 1 || page.NextBefore == nil || page.Trades[0].PositionID != second.Position.ID || page.Trades[0].RMultiple == nil || *page.Trades[0].RMultiple != "-0.04" || page.Trades[0].Fees != "0.2" || len(page.Trades[0].Rules) != 1 || page.Trades[0].Rules[0] != (Rule{Rule: "class_allocation", Status: "caution"}) {
		t.Fatalf("journal: %+v %v", page, err)
	}
	older, err := s.Store.Journal(ctx, owner, a.ID, 1, page.NextBefore)
	if err != nil || len(older.Trades) != 1 || older.Trades[0].PositionID != open.Position.ID || older.NextBefore != nil {
		t.Fatalf("cursor: %+v %v", older, err)
	}
	if _, err := s.Portfolio(ctx, "22222222-2222-4222-8222-222222222222", a.ID, "all"); !errors.Is(err, risk.ErrAccountNotFound) {
		t.Fatalf("foreign account: %v", err)
	}
	// Missing prices skip the entire snapshot, never a partial valuation.
	now = now.Add(time.Minute)
	if _, err := trade.Open(ctx, owner, trading.Request{AccountID: a.ID, Symbol: "BTCUSDT", Side: "long"}); err != nil {
		t.Fatal(err)
	}
	var beforeCount int
	if err := pool.QueryRow(ctx, `SELECT count(*) FROM app.equity_snapshots WHERE account_id=$1::uuid`, a.ID).Scan(&beforeCount); err != nil {
		t.Fatal(err)
	}
	s.Prices = fixedPrices{err: trading.ErrPrice}
	now = now.Add(time.Minute)
	if err := s.Snapshot(ctx, a.ID); !errors.Is(err, trading.ErrPrice) {
		t.Fatalf("missing price: %v", err)
	}
	if err := pool.QueryRow(ctx, `SELECT count(*) FROM app.equity_snapshots WHERE account_id=$1::uuid`, a.ID).Scan(&count); err != nil || count != beforeCount {
		t.Fatalf("partial snapshot written %d %v", count, err)
	}
	s.Prices = prices
	if err := s.Snapshot(ctx, a.ID); err != nil {
		t.Fatal(err)
	}
	// HTTP is mounted inside the real full-session/JSON policy.
	token := strings.Repeat("p", 43)
	if _, err := pool.Exec(ctx, `INSERT INTO app.sessions(user_id,token_hash,expires_at,user_agent,ip) VALUES($1::uuid,$2,now()+interval '1 hour','portfolio-test','127.0.0.1')`, owner, auth.TokenHash(token)); err != nil {
		t.Fatal(err)
	}
	box, err := cryptobox.New(base64.StdEncoding.EncodeToString(make([]byte, 32)))
	if err != nil {
		t.Fatal(err)
	}
	authService, err := auth.New(auth.Postgres{Pool: pool}, box)
	if err != nil {
		t.Fatal(err)
	}
	r := chi.NewRouter()
	auth.HTTP{Service: authService}.Mount(r, s.Mount)
	for _, endpoint := range []string{"portfolio", "journal"} {
		for _, tc := range []struct {
			id, cookie string
			code       int
		}{{a.ID, token, 200}, {a.ID, "", 401}, {"22222222-2222-4222-8222-222222222222", token, 404}} {
			req := httptest.NewRequest("GET", "/v1/accounts/"+tc.id+"/"+endpoint, nil)
			if tc.cookie != "" {
				req.AddCookie(&http.Cookie{Name: auth.CookieName, Value: tc.cookie})
			}
			w := httptest.NewRecorder()
			r.ServeHTTP(w, req)
			if w.Code != tc.code || w.Header().Get("Cache-Control") != "no-store" {
				t.Fatalf("%s %d %s", endpoint, w.Code, w.Body)
			}
		}
	}
	// An authenticated different user sees 404 for this existing account.
	var foreign string
	if err := pool.QueryRow(ctx, `INSERT INTO app.users(email,password_hash) VALUES(gen_random_uuid()::text || '@portfolio.example.com','test-only') RETURNING id::text`).Scan(&foreign); err != nil {
		t.Fatal(err)
	}
	defer pool.Exec(context.Background(), `DELETE FROM app.users WHERE id=$1::uuid`, foreign)
	foreignToken := strings.Repeat("f", 43)
	if _, err := pool.Exec(ctx, `INSERT INTO app.sessions(user_id,token_hash,expires_at,user_agent,ip) VALUES($1::uuid,$2,now()+interval '1 hour','portfolio-test','127.0.0.1')`, foreign, auth.TokenHash(foreignToken)); err != nil {
		t.Fatal(err)
	}
	for _, endpoint := range []string{"portfolio", "journal"} {
		req := httptest.NewRequest("GET", "/v1/accounts/"+a.ID+"/"+endpoint, nil)
		req.AddCookie(&http.Cookie{Name: auth.CookieName, Value: foreignToken})
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
		if w.Code != 404 {
			t.Fatalf("foreign %s: %d %s", endpoint, w.Code, w.Body)
		}
	}
	req := httptest.NewRequest("GET", "/v1/accounts/"+a.ID+"/journal?limit=1&before="+url.QueryEscape(page.NextBefore.Format(time.RFC3339Nano)), nil)
	req.AddCookie(&http.Cookie{Name: auth.CookieName, Value: token})
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	var response Journal
	if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil || w.Code != 200 || len(response.Trades) != 1 || response.Trades[0].PositionID != open.Position.ID {
		t.Fatalf("HTTP cursor: %s %v", w.Body, err)
	}
	for _, market := range []string{"spot", "futures"} {
		balance, leverage := "200.15", uint32(1)
		if market == "futures" {
			balance, leverage = "20.075", 10
		}
		cashAccount, err := (accounts.Postgres{Pool: pool}).CreatePaper(ctx, owner, "Reserved capital", market, decimal.RequireFromString(balance))
		if err != nil {
			t.Fatal(err)
		}
		body := trading.Request{AccountID: cashAccount.ID, Symbol: "BTCUSDT", Side: "long", Leverage: &leverage}
		if _, err := trade.Open(ctx, owner, body); err != nil {
			t.Fatal(err)
		}
		if _, err := trade.Open(ctx, owner, body); !errors.Is(err, trading.ErrBalance) {
			t.Fatalf("%s reserved cash: %v", market, err)
		}
		var orders, checks, fills int
		if err := pool.QueryRow(ctx, `SELECT (SELECT count(*) FROM app.orders WHERE account_id=$1::uuid),
			(SELECT count(*) FROM app.risk_checks WHERE account_id=$1::uuid),
			(SELECT count(*) FROM app.fills f JOIN app.orders o ON o.id=f.order_id WHERE o.account_id=$1::uuid)`, cashAccount.ID).Scan(&orders, &checks, &fills); err != nil || orders != 2 || checks != 2 || fills != 1 {
			t.Fatalf("%s rejection audit: %d %d %d %v", market, orders, checks, fills, err)
		}
	}
}

type portfolioTickers struct{}

func (portfolioTickers) Tickers(context.Context, []string) (map[string]map[string]string, error) {
	return map[string]map[string]string{"BTCUSDT": {"price": "100"}}, nil
}

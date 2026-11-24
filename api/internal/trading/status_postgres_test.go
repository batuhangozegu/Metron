package trading

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/batuhangozegu/metron/api/internal/auth"
	"github.com/batuhangozegu/metron/api/internal/binance"
	"github.com/batuhangozegu/metron/api/internal/cryptobox"
	"github.com/batuhangozegu/metron/api/internal/risk"
	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/shopspring/decimal"
)

type riskStatusSession struct {
	auth.Store
	owner string
}

func (s riskStatusSession) ResolveSession(context.Context, []byte, bool, time.Time) (auth.User, error) {
	return auth.User{ID: s.owner}, nil
}

type riskStatusPrices struct{}

func (riskStatusPrices) Price(_ context.Context, symbol string) (decimal.Decimal, error) {
	if symbol == "BTCUSDT" {
		return decimal.RequireFromString("110.25"), nil
	}
	return decimal.Zero, ErrPrice
}

func TestPostgresRiskStatus(t *testing.T) {
	databaseURL := os.Getenv("TEST_DATABASE_URL")
	if databaseURL == "" {
		t.Skip("set TEST_DATABASE_URL to a migrated test Postgres database")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	pool, err := pgxpool.New(ctx, databaseURL)
	if err != nil {
		t.Fatal(err)
	}
	defer pool.Close()
	var owner, foreignOwner string
	for _, user := range []*string{&owner, &foreignOwner} {
		if err := pool.QueryRow(ctx, `INSERT INTO app.users(email,password_hash) VALUES(gen_random_uuid()::text || '@risk-status.example.com','test-only') RETURNING id::text`).Scan(user); err != nil {
			t.Fatal(err)
		}
		id := *user
		defer func() {
			if _, err := pool.Exec(context.Background(), `DELETE FROM app.users WHERE id=$1::uuid`, id); err != nil {
				t.Error(err)
			}
		}()
	}
	box, err := cryptobox.New(base64.StdEncoding.EncodeToString(make([]byte, 32)))
	if err != nil {
		t.Fatal(err)
	}
	_, _, exchange, _ := liveFixture(t)
	store := risk.Postgres{Pool: pool}
	service := risk.Service{
		Accounts: RiskAccounts{Store: Postgres{Pool: pool}, Live: fakeLiveProvider{exchange, "testnet"}},
		Store:    store, Prices: riskStatusPrices{},
	}
	request := func(owner, id string) *httptest.ResponseRecorder {
		t.Helper()
		h := chi.NewRouter()
		auth.HTTP{Service: &auth.Service{Store: riskStatusSession{owner: owner}, Box: box, Now: time.Now}}.Mount(h, service.Mount)
		r := httptest.NewRequest(http.MethodGet, "/v1/accounts/"+id+"/risk-status", nil)
		r.AddCookie(&http.Cookie{Name: auth.CookieName, Value: testToken})
		w := httptest.NewRecorder()
		h.ServeHTTP(w, r)
		return w
	}
	for _, kind := range []string{"paper", "live"} {
		t.Run(kind, func(t *testing.T) {
			var id string
			if err := pool.QueryRow(ctx, `INSERT INTO app.accounts(user_id,type,name,market,paper_balance)
				VALUES($1::uuid,$2,'Risk status','futures',CASE WHEN $2='paper' THEN 1000.123456789 ELSE NULL END) RETURNING id::text`, owner, kind).Scan(&id); err != nil {
				t.Fatal(err)
			}
			w := request(owner, id)
			var empty risk.Status
			if err := json.Unmarshal(w.Body.Bytes(), &empty); err != nil || w.Code != 200 || empty.OpenRisk.Amount != "0" || empty.DailyLoss.Amount != "0" || empty.Unrealized != "0" || empty.OpenPositions.Limit != 5 || empty.ChecksToday != (risk.CheckCounts{}) {
				t.Fatalf("empty defaults: %d %s %v", w.Code, w.Body, err)
			}
			now := time.Now().UTC()
			day, week := risk.UTCWindows(now)
			// The same rows drive status amounts, engine percentages and counts.
			for _, p := range []struct {
				symbol, side, state, risk, pnl, reason string
				closed                                 *time.Time
			}{
				{"BTCUSDT", "long", "open", "10.25", "0", "", nil},
				{"BTCUSDT", "short", "open", "5", "0", "", nil},
				{"MISSINGUSDT", "long", "open", "1", "0", "", nil},
				{"BTCUSDT", "long", "closed", "10", "-12.5", "stop", &day},
				{"BTCUSDT", "long", "closed", "10", "5", "manual", &day},
				{"BTCUSDT", "long", "closed", "10", "-999", "stop", timePointer(week.Add(-time.Second))},
				{"BTCUSDT", "long", "closed", "10", "-999", "stop", timePointer(day.AddDate(0, 0, 1))},
			} {
				if _, err := pool.Exec(ctx, `INSERT INTO app.positions(id,account_id,symbol,side,quantity,entry_price,stop,coin_class,initial_risk,status,closed_at,realized_pnl,close_reason)
					VALUES(gen_random_uuid(),$1::uuid,$2,$3,CASE WHEN $3='short' THEN 1 ELSE 2 END,100,95,'core',$4::numeric,$5,$6,$7::numeric,NULLIF($8,''))`, id, p.symbol, p.side, p.risk, p.state, p.closed, p.pnl, p.reason); err != nil {
					t.Fatal(err)
				}
			}
			if kind == "live" {
				if _, err := pool.Exec(ctx, `INSERT INTO app.positions(id,account_id,symbol,side,quantity,entry_price,stop,coin_class,initial_risk,status,closed_at,realized_pnl,close_reason)
					VALUES(gen_random_uuid(),$1::uuid,'BTCUSDT','long',1,100,95,'core',5,'closed',$2,0,'exchange_stop')`, id, day); err != nil {
					t.Fatal(err)
				}
			}
			var orderID string
			if err := pool.QueryRow(ctx, `INSERT INTO app.orders(account_id,symbol,side,kind,purpose,quantity,status)
				VALUES($1::uuid,'BTCUSDT','long','market','open',1,'rejected') RETURNING id::text`, id).Scan(&orderID); err != nil {
				t.Fatal(err)
			}
			// Exceeds the history endpoint's 100-row cap. Count orders by the
			// actual order link, including rejected orders, rather than source.
			if _, err := pool.Exec(ctx, `INSERT INTO app.risk_checks(account_id,source,order_id,symbol,side,entry_price,allowed,verdict,failed_rules,caution_rules,profile_source,result,created_at)
				SELECT $1::uuid,'check',CASE WHEN n=2 THEN $2::uuid ELSE NULL END,'BTCUSDT','long',100,n<>1,'pass','{}','{}','stored','{}',$3
				FROM generate_series(1,103) n`, id, orderID, day); err != nil {
				t.Fatal(err)
			}
			for _, at := range []time.Time{day.Add(-time.Second), day.AddDate(0, 0, 1)} {
				if _, err := pool.Exec(ctx, `INSERT INTO app.risk_checks(account_id,source,symbol,side,entry_price,allowed,verdict,failed_rules,caution_rules,profile_source,result,created_at)
					VALUES($1::uuid,'order','BTCUSDT','long',100,false,'fail','{}','{}','stored','{}',$2)`, id, at); err != nil {
					t.Fatal(err)
				}
			}
			p := risk.DefaultProfile()
			p.TotalOpenRiskLimitPercent, p.DailyLossLimitPercent, p.WeeklyLossLimitPercent = "4.5", "3.5", "8.5"
			open, losing, stopped := uint32(2), uint32(1), uint32(4)
			p.MaxOpenPositions, p.MaxLosingTradesToday, p.MaxStoppedTradesThisWeek = &open, &losing, &stopped
			if _, err := store.SaveProfile(ctx, owner, id, p); err != nil {
				t.Fatal(err)
			}
			w = request(owner, id)
			var got risk.Status
			if err := json.Unmarshal(w.Body.Bytes(), &got); err != nil || w.Code != 200 {
				t.Fatalf("status: %d %s %v", w.Code, w.Body, err)
			}
			preview, err := service.Accounts.RiskAccount(ctx, owner, id, "BTCUSDT")
			if err != nil {
				t.Fatal(err)
			}
			wantEquity, wantStopped := "1000.12345678", uint32(1)
			if kind == "live" {
				wantEquity, wantStopped = "6000", 2
			}
			if got.AccountID != id || got.Currency != "USDT" || got.Market != "futures" || got.AsOf.Location() != time.UTC || got.AsOf.Before(now) || got.Equity != wantEquity || got.RealizedToday != "-7.5" || got.Unrealized != "10.25" ||
				got.OpenRisk != (risk.AmountStatus{Amount: "16.25", Percent: preview.Activity.OpenRiskPercent, LimitPercent: "4.5"}) ||
				got.DailyLoss != (risk.AmountStatus{Amount: "7.5", Percent: preview.Activity.DailyLossPercent, LimitPercent: "3.5"}) ||
				got.WeeklyLoss != (risk.AmountStatus{Amount: "7.5", Percent: preview.Activity.WeeklyLossPercent, LimitPercent: "8.5"}) ||
				got.OpenPositions != (risk.CountStatus{Count: 3, Limit: 2}) || got.LosingTradesToday != (risk.CountStatus{Count: 1, Limit: 1}) ||
				got.StoppedTradesThisWeek != (risk.CountStatus{Count: wantStopped, Limit: 4}) || got.ChecksToday != (risk.CheckCounts{Run: 103, Orders: 1, Blocked: 1}) {
				t.Fatalf("wrong status: %+v preview=%+v", got, preview.Activity)
			}
			if w := request(foreignOwner, id); w.Code != 404 || !strings.Contains(w.Body.String(), `"code":"account_not_found"`) {
				t.Fatalf("foreign status: %d %s", w.Code, w.Body)
			}
			if _, err := store.ChecksToday(ctx, foreignOwner, id, now); err != risk.ErrAccountNotFound {
				t.Fatalf("foreign counts: %v", err)
			}
			if kind == "live" {
				exchange.errorBalance = binance.ErrUnavailable
				defer func() { exchange.errorBalance = nil }()
				w := request(owner, id)
				if w.Code != 503 {
					t.Fatalf("exchange outage: %d %s", w.Code, w.Body)
				}
			}
		})
	}
	if w := request(owner, newID()); w.Code != 404 {
		t.Fatalf("unknown status: %d %s", w.Code, w.Body)
	}
}

func timePointer(at time.Time) *time.Time { return &at }

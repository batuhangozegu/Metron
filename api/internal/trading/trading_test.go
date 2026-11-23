package trading

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/batuhangozegu/metron/api/internal/auth"
	"github.com/batuhangozegu/metron/api/internal/cryptobox"
	enginev1 "github.com/batuhangozegu/metron/api/internal/gen/metron/engine/v1"
	"github.com/batuhangozegu/metron/api/internal/risk"
	"github.com/go-chi/chi/v5"
	"github.com/shopspring/decimal"
	"google.golang.org/grpc"
)

const testAccount = "11111111-1111-4111-8111-111111111111"
const testToken = "sssssssssssssssssssssssssssssssssssssssssss"
const orderBody = `{"account_id":"` + testAccount + `","symbol":"BTCUSDT","side":"long"}`

var testNow = time.Date(2026, 11, 19, 12, 0, 0, 0, time.UTC)

type memoryStore struct {
	mu        sync.Mutex
	market    string
	balance   decimal.Decimal
	orders    []Order
	fills     []Fill
	positions map[string]Position
	failWrite bool
	commits   int
	failBegin int
	onCommit  chan struct{}
	checks    []string
	failCheck bool
	profile   *risk.Profile
}

type memoryTx struct {
	s         *memoryStore
	balance   decimal.Decimal
	orders    []Order
	fills     []Fill
	positions map[string]Position
	done      bool
	checks    []string
}

func (s *memoryStore) Begin(_ context.Context, owner, account string) (Transaction, error) {
	s.mu.Lock()
	if s.failBegin > 0 {
		s.failBegin--
		s.mu.Unlock()
		return nil, errors.New("transient database outage")
	}
	if owner != "owner" || account != testAccount {
		s.mu.Unlock()
		return nil, risk.ErrAccountNotFound
	}
	tx := &memoryTx{s: s, balance: s.balance, orders: append([]Order{}, s.orders...), fills: append([]Fill{}, s.fills...), checks: append([]string{}, s.checks...), positions: map[string]Position{}}
	for id, p := range s.positions {
		tx.positions[id] = p
	}
	return tx, nil
}
func (s *memoryStore) PositionAccount(_ context.Context, owner *string, id string) (string, string, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.positions[id]; !ok || (owner != nil && *owner != "owner") {
		return "", "", ErrNotFound
	}
	return testAccount, "owner", nil
}
func (s *memoryStore) Positions(_ context.Context, owner, account, state string) ([]Position, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if owner != "owner" || account != testAccount {
		return nil, risk.ErrAccountNotFound
	}
	result := []Position{}
	for _, p := range s.positions {
		if p.Status == state {
			result = append(result, p)
		}
	}
	return result, nil
}
func (tx *memoryTx) Account() risk.Account {
	return risk.Account{Equity: tx.balance.String(), Market: tx.s.market, Profile: tx.s.profile}
}

func (tx *memoryTx) RecordCheck(_ context.Context, body risk.Request, _ *enginev1.CheckResponse, orderID string) error {
	if tx.s.failCheck {
		return errors.New("private check write failure")
	}
	if body.EntryPrice != "100" || body.RiskProfile != nil {
		return errors.New("unexpected audit input")
	}
	tx.checks = append(tx.checks, orderID)
	return nil
}
func (tx *memoryTx) Snapshot(_ context.Context, _ string) (risk.Account, error) {
	a := tx.Account()
	var rows []risk.ActivityPosition
	for _, p := range tx.positions {
		row := risk.ActivityPosition{Status: p.Status, CoinClass: p.CoinClass, Quantity: decimal.RequireFromString(p.Quantity), EntryPrice: decimal.RequireFromString(p.EntryPrice), InitialRisk: decimal.RequireFromString(p.InitialRisk)}
		if p.ClosedAt != nil {
			row.ClosedAt = *p.ClosedAt
		}
		if p.RealizedPNL != nil {
			row.RealizedPNL = decimal.RequireFromString(*p.RealizedPNL)
		}
		if p.CloseReason != nil {
			row.CloseReason = *p.CloseReason
		}
		rows = append(rows, row)
	}
	a.Activity, a.CurrentClassHolding = risk.AggregateActivity(rows, tx.balance, "core", a.Market, testNow)
	return a, nil
}
func (tx *memoryTx) Position(_ context.Context, id string) (Position, error) {
	p, ok := tx.positions[id]
	if !ok {
		return p, ErrNotFound
	}
	return p, nil
}
func (tx *memoryTx) OpeningFee(_ context.Context, id string) (decimal.Decimal, error) {
	fee := decimal.Zero
	for _, o := range tx.orders {
		if o.PositionID == nil || *o.PositionID != id || o.Purpose != "open" {
			continue
		}
		for _, f := range tx.fills {
			if f.OrderID == o.ID {
				fee = fee.Add(decimal.RequireFromString(f.Fee))
			}
		}
	}
	return fee, nil
}
func (tx *memoryTx) Capital(ctx context.Context) ([]Capital, error) {
	var capital []Capital
	for _, p := range tx.positions {
		if p.Status != "open" {
			continue
		}
		fee, err := tx.OpeningFee(ctx, p.ID)
		if err != nil {
			return nil, err
		}
		leverage := uint32(1)
		if p.Leverage != nil {
			leverage = *p.Leverage
		}
		capital = append(capital, Capital{Notional: decimal.RequireFromString(p.Quantity).Mul(decimal.RequireFromString(p.EntryPrice)), Fee: fee, Leverage: leverage})
	}
	return capital, nil
}
func (tx *memoryTx) Reject(_ context.Context, o Order) error {
	tx.orders = append(tx.orders, o)
	return nil
}
func (tx *memoryTx) Open(_ context.Context, e Execution) error {
	tx.orders = append(tx.orders, e.Order)
	tx.fills = append(tx.fills, e.Fill)
	tx.positions[e.Position.ID] = e.Position
	if tx.s.failWrite {
		return errors.New("private database failure")
	}
	return nil
}
func (tx *memoryTx) Close(ctx context.Context, e Execution, pnl decimal.Decimal) error {
	if err := tx.Open(ctx, e); err != nil {
		return err
	}
	tx.balance = tx.balance.Add(pnl)
	return nil
}
func (tx *memoryTx) Commit(context.Context) error {
	tx.s.balance, tx.s.orders, tx.s.fills, tx.s.positions = tx.balance, tx.orders, tx.fills, tx.positions
	tx.s.checks = tx.checks
	tx.s.commits++
	if tx.s.onCommit != nil {
		select {
		case tx.s.onCommit <- struct{}{}:
		default:
		}
	}
	tx.done = true
	tx.s.mu.Unlock()
	return nil
}
func (tx *memoryTx) Rollback(context.Context) error {
	if !tx.done {
		tx.done = true
		tx.s.mu.Unlock()
	}
	return nil
}

type ticker struct {
	value string
	err   error
}

func (p *ticker) Tickers(_ context.Context, _ []string) (map[string]map[string]string, error) {
	return map[string]map[string]string{"BTCUSDT": {"price": p.value}}, p.err
}

type testEngine struct {
	requests          []*enginev1.CheckRequest
	blocked, failRule bool
	size              string
	err               error
}

func (e *testEngine) Check(_ context.Context, r *enginev1.CheckRequest, _ ...grpc.CallOption) (*enginev1.CheckResponse, error) {
	e.requests = append(e.requests, r)
	if e.err != nil {
		return nil, e.err
	}
	size := e.size
	if size == "" {
		size = "2"
	}
	stop := "95"
	if r.Side == "short" {
		stop = "105"
	}
	status := "pass"
	if e.failRule || e.blocked {
		status = "fail"
	}
	return &enginev1.CheckResponse{Symbol: r.Symbol, Side: r.Side, EntryPrice: r.EntryPrice, PositionSize: size, Stop: stop,
		CoinClass: "core", Market: r.Market, Leverage: max(1, r.GetLeverage()), Allowed: !e.blocked, Checks: []*enginev1.RuleCheck{{RuleId: "open_positions", Status: status}}}, nil
}

type sessionStore struct{ auth.Store }

func (sessionStore) ResolveSession(_ context.Context, hash []byte, pending bool, _ time.Time) (auth.User, error) {
	if pending || !bytes.Equal(hash, auth.TokenHash(testToken)) {
		return auth.User{}, auth.ErrUnauthorized
	}
	return auth.User{ID: "owner"}, nil
}

func fixture(t *testing.T) (Service, *memoryStore, *ticker, *testEngine, http.Handler) {
	t.Helper()
	store := &memoryStore{market: "spot", balance: decimal.NewFromInt(1000), positions: map[string]Position{}}
	prices := &ticker{value: "100"}
	engine := &testEngine{}
	s := Service{Store: store, Exchange: Paper{Tickers: prices}, Risk: risk.Service{Engine: engine}, Now: func() time.Time { return testNow }}
	box, err := cryptobox.New(base64.StdEncoding.EncodeToString(make([]byte, 32)))
	if err != nil {
		t.Fatal(err)
	}
	r := chi.NewRouter()
	r.Use(ErrorShape)
	auth.HTTP{Service: &auth.Service{Store: sessionStore{}, Box: box, Now: time.Now}}.Mount(r, s.Mount)
	return s, store, prices, engine, r
}

func request(h http.Handler, method, path, body, token, media string) *httptest.ResponseRecorder {
	r := httptest.NewRequest(method, path, strings.NewReader(body))
	r.Header.Set("Content-Type", media)
	if token != "" {
		r.AddCookie(&http.Cookie{Name: auth.CookieName, Value: token})
	}
	w := httptest.NewRecorder()
	h.ServeHTTP(w, r)
	return w
}
func post(h http.Handler, path, body string) *httptest.ResponseRecorder {
	return request(h, "POST", path, body, testToken, "application/json")
}
func assertError(t *testing.T, w *httptest.ResponseRecorder, code int, id string) {
	t.Helper()
	var result struct {
		Error json.RawMessage `json:"error"`
		Code  string          `json:"code"`
		TR    string          `json:"error_tr"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &result); err != nil {
		t.Fatal(err)
	}
	var legacy wireError
	var en string
	if json.Unmarshal(result.Error, &en) != nil {
		if err := json.Unmarshal(result.Error, &legacy); err != nil || legacy.ID != id || legacy.TR == "" {
			t.Fatalf("invalid bilingual error: %s", w.Body)
		}
		en = legacy.EN
	}
	if w.Code != code || result.Code != id || en == "" || result.TR == "" {
		t.Fatalf("want %d %s; got %d %s", code, id, w.Code, w.Body)
	}
}
func mustOpen(t *testing.T, s Service) Execution {
	t.Helper()
	e, err := s.Open(context.Background(), "owner", Request{AccountID: testAccount, Symbol: "BTCUSDT", Side: "long"})
	if err != nil {
		t.Fatal(err)
	}
	return e
}

func TestOrderFlow(t *testing.T) {
	for _, tc := range []struct {
		name, extra, balance, market, size string
		blocked, fail                      bool
		code                               int
		id                                 string
	}{
		{name: "success", code: 201},
		{name: "smaller quantity", extra: `,"quantity":"1"`, code: 201},
		{name: "blocked", blocked: true, code: 422, id: "risk_blocked"},
		{name: "zero size blocked", blocked: true, size: "0", code: 422, id: "risk_blocked"},
		{name: "fail always blocks", fail: true, code: 422, id: "risk_blocked"},
		{name: "too large", extra: `,"quantity":"3"`, code: 422, id: "quantity_exceeded"},
		{name: "spot fee exceeds balance", balance: "200.19999999", code: 422, id: "insufficient_balance"},
		{name: "spot exact balance", balance: "200.2", code: 201},
		{name: "futures margin plus fee", market: "futures", extra: `,"leverage":10`, balance: "20.09999999", code: 422, id: "insufficient_balance"},
		{name: "futures exact balance", market: "futures", extra: `,"leverage":10`, balance: "20.1", code: 201},
		{name: "repeating margin below", market: "futures", extra: `,"leverage":3`, balance: "66.76666666", code: 422, id: "insufficient_balance"},
		{name: "repeating margin above", market: "futures", extra: `,"leverage":3`, balance: "66.76666667", code: 201},
	} {
		t.Run(tc.name, func(t *testing.T) {
			_, store, _, engine, h := fixture(t)
			engine.blocked, engine.failRule = tc.blocked, tc.fail
			engine.size = tc.size
			if tc.balance != "" {
				store.balance = decimal.RequireFromString(tc.balance)
			}
			if tc.market != "" {
				store.market = tc.market
			}
			balance := store.balance
			w := post(h, "/v1/orders", strings.TrimSuffix(orderBody, "}")+tc.extra+"}")
			if tc.code != 201 {
				assertError(t, w, tc.code, tc.id)
				if len(store.orders) != 1 || store.orders[0].Status != "rejected" || len(store.fills) != 0 || len(store.positions) != 0 || store.commits != 1 {
					t.Fatalf("rejected order not committed: %+v", store)
				}
				if !strings.Contains(w.Body.String(), `"check":`) {
					t.Fatal("missing risk check")
				}
			} else {
				if w.Code != 201 {
					t.Fatalf("%d %s", w.Code, w.Body)
				}
				var e Execution
				if err := json.Unmarshal(w.Body.Bytes(), &e); err != nil {
					t.Fatal(err)
				}
				qty := "2"
				if tc.extra == `,"quantity":"1"` {
					qty = "1"
				}
				rate := "0.001"
				if store.market == "futures" {
					rate = "0.0005"
				}
				fee := decimal.RequireFromString(qty).Mul(decimal.NewFromInt(100)).Mul(decimal.RequireFromString(rate)).String()
				if e.Position.EntryPrice != "100" || e.Position.Quantity != qty || e.Fill.Fee != fee || e.Position.InitialRisk != decimal.RequireFromString(qty).Mul(decimal.NewFromInt(5)).String() || len(store.positions) != 1 {
					t.Fatalf("wrong execution: %s", w.Body)
				}
				if !strings.Contains(w.Body.String(), `"position_size":"2"`) || !strings.Contains(w.Body.String(), `"last_price":null`) {
					t.Fatal("wrong wire format")
				}
			}
			if !store.balance.Equal(balance) {
				t.Fatal("open changed settled balance")
			}
			if len(store.checks) != 1 || store.checks[0] != store.orders[0].ID {
				t.Fatal("order and risk check were not committed together")
			}
			if engine.requests[0].EntryPrice != "100" || engine.requests[0].AccountEquity != balance.String() {
				t.Fatal("untrusted risk input")
			}
		})
	}
}

func TestCloseAndStopSettlement(t *testing.T) {
	for _, stop := range []bool{false, true} {
		t.Run(map[bool]string{false: "manual", true: "stop"}[stop], func(t *testing.T) {
			s, store, prices, _, h := fixture(t)
			opened := mustOpen(t, s)
			prices.value = "110"
			want := "19.58"
			if stop {
				want = "-10.39"
				// Stop settlement must work even with Redis unavailable and forged
				// event levels: the authoritative stop is stored on the position.
				s.Exchange = Paper{}
				hit := &enginev1.StopHit{PositionId: opened.Position.ID, Stop: "1", Price: "999"}
				for range 2 {
					if err := s.Stop(context.Background(), hit); err != nil {
						t.Fatal(err)
					}
				}
			} else {
				w := post(h, "/v1/positions/"+opened.Position.ID+"/close", `{}`)
				if w.Code != 200 {
					t.Fatalf("%d %s", w.Code, w.Body)
				}
				assertError(t, post(h, "/v1/positions/"+opened.Position.ID+"/close", `{}`), 409, "position_closed")
			}
			p := store.positions[opened.Position.ID]
			if p.Status != "closed" || p.RealizedPNL == nil || *p.RealizedPNL != want || len(store.orders) != 2 || len(store.fills) != 2 || store.commits != 2 || !store.balance.Equal(decimal.NewFromInt(1000).Add(decimal.RequireFromString(want))) {
				t.Fatalf("wrong settlement: %+v %+v", p, store)
			}
			if p.ClosedAt.Location() != time.UTC {
				t.Fatal("time must be UTC")
			}
		})
	}
}

func TestConcurrentClosesSettleOnce(t *testing.T) {
	s, store, _, _, _ := fixture(t)
	e := mustOpen(t, s)
	var wg sync.WaitGroup
	errs := make(chan error, 20)
	for range 20 {
		wg.Go(func() {
			_, err := s.Close(context.Background(), "owner", e.Position.ID)
			if err != nil && !errors.Is(err, ErrClosed) {
				errs <- err
			}
		})
	}
	wg.Wait()
	close(errs)
	for err := range errs {
		t.Fatal(err)
	}
	if len(store.orders) != 2 || len(store.fills) != 2 || store.commits != 2 || store.balance.String() != "999.6" {
		t.Fatalf("settled more than once: %+v", store)
	}
}

func TestRollbackAndIsolation(t *testing.T) {
	s, store, _, _, h := fixture(t)
	store.failWrite = true
	w := post(h, "/v1/orders", orderBody)
	assertError(t, w, 503, "trading_unavailable")
	if len(store.orders) != 0 || len(store.fills) != 0 || len(store.positions) != 0 || store.commits != 0 || strings.Contains(w.Body.String(), "private") {
		t.Fatal("partial order escaped rollback")
	}
	store.failWrite = false
	e := mustOpen(t, s)
	store.failWrite = true
	w = post(h, "/v1/positions/"+e.Position.ID+"/close", `{}`)
	assertError(t, w, 503, "trading_unavailable")
	if store.positions[e.Position.ID].Status != "open" || len(store.orders) != 1 || store.balance.String() != "1000" {
		t.Fatal("partial close escaped rollback")
	}
	_, err := s.Close(context.Background(), "foreign", e.Position.ID)
	if !errors.Is(err, ErrNotFound) {
		t.Fatal(err)
	}
	assertError(t, post(h, "/v1/positions/22222222-2222-4222-8222-222222222222/close", `{}`), 404, "position_not_found")
	assertError(t, post(h, "/v1/orders", strings.Replace(orderBody, testAccount, "22222222-2222-4222-8222-222222222222", 1)), 404, "account_not_found")
}

func TestOrderInputAndSessionPolicy(t *testing.T) {
	for _, body := range []string{`null`, `[]`, `{`, orderBody + `{}`, strings.TrimSuffix(orderBody, "}") + `,"entry_price":"1"}`, strings.TrimSuffix(orderBody, "}") + `,"activity":{}}`, strings.TrimSuffix(orderBody, "}") + `,"quantity":2}`, strings.TrimSuffix(orderBody, "}") + `,"quantity":"0"}`, strings.TrimSuffix(orderBody, "}") + `,"quantity":"1e2"}`, strings.TrimSuffix(orderBody, "}") + `,"quantity":"0.000000001"}`} {
		_, store, _, engine, h := fixture(t)
		w := post(h, "/v1/orders", body)
		assertError(t, w, 400, "invalid_input")
		if len(store.orders) != 0 || len(engine.requests) != 0 {
			t.Fatal("invalid input reached execution")
		}
	}
	_, _, prices, _, h := fixture(t)
	assertError(t, request(h, "POST", "/v1/orders", orderBody, "", "application/json"), 401, "authentication_required")
	assertError(t, request(h, "POST", "/v1/orders", orderBody, testToken, "text/plain"), 415, "unsupported_media_type")
	prices.value = ""
	assertError(t, post(h, "/v1/orders", orderBody), 503, "price_unavailable")
}

func TestPositionsPricesAndTrustedActivity(t *testing.T) {
	s, _, prices, engine, h := fixture(t)
	e := mustOpen(t, s)
	prices.value = "101"
	w := request(h, "GET", "/v1/accounts/"+testAccount+"/positions", "", testToken, "")
	if w.Code != 200 || !strings.Contains(w.Body.String(), `"last_price":"101"`) || !strings.Contains(w.Body.String(), `"unrealized_pnl":"2"`) {
		t.Fatalf("%d %s", w.Code, w.Body)
	}
	prices.err = ErrPrice
	w = request(h, "GET", "/v1/accounts/"+testAccount+"/positions", "", testToken, "")
	if !strings.Contains(w.Body.String(), `"last_price":null`) || !strings.Contains(w.Body.String(), `"unrealized_pnl":null`) {
		t.Fatal(w.Body)
	}
	prices.err = nil
	prices.value = "100"
	mustOpen(t, s)
	r := engine.requests[1]
	if r.Activity.OpenPositions != 1 || r.Activity.OpenRiskPercent != "1" || r.CurrentClassHolding != "200" {
		t.Fatalf("missing real activity: %+v", r)
	}
	if _, err := s.Close(context.Background(), "owner", e.Position.ID); err != nil {
		t.Fatal(err)
	}
	w = request(h, "GET", "/v1/accounts/"+testAccount+"/positions?status=closed", "", testToken, "")
	if w.Code != 200 || strings.Contains(w.Body.String(), "last_price") || strings.Contains(w.Body.String(), "unrealized_pnl") {
		t.Fatalf("%d %s", w.Code, w.Body)
	}
	w = request(h, "GET", "/v1/accounts/"+testAccount+"/positions?status=all", "", testToken, "")
	assertError(t, w, 400, "invalid_input")
}

func TestPaperFeePrecision(t *testing.T) {
	for market, want := range map[string]string{"spot": "0.000000000000000001", "futures": "0.0000000000000000005"} {
		fill, err := (Paper{}).Execute(context.Background(), FillRequest{Market: market, Price: decimal.RequireFromString("0.00000001"), Quantity: decimal.RequireFromString("0.0000001")})
		if err != nil || fill.Fee != want {
			t.Fatalf("%s %+v %v", market, fill, err)
		}
	}
}

func TestFuturesShortSettlement(t *testing.T) {
	for _, stop := range []bool{false, true} {
		s, store, prices, _, _ := fixture(t)
		store.market = "futures"
		leverage := uint32(10)
		e, err := s.Open(context.Background(), "owner", Request{AccountID: testAccount, Symbol: "BTCUSDT", Side: "short", Leverage: &leverage})
		if err != nil {
			t.Fatal(err)
		}
		want := "19.81"
		if stop {
			want = "-10.205"
			err = s.Stop(context.Background(), &enginev1.StopHit{PositionId: e.Position.ID})
		} else {
			prices.value = "90"
			_, err = s.Close(context.Background(), "owner", e.Position.ID)
		}
		if err != nil || *store.positions[e.Position.ID].RealizedPNL != want || !store.balance.Equal(decimal.NewFromInt(1000).Add(decimal.RequireFromString(want))) {
			t.Fatalf("short stop=%t: %+v %v", stop, store, err)
		}
	}
}

type observedExchange struct {
	Paper
	executions int
}

func (e *observedExchange) Execute(ctx context.Context, request FillRequest) (Fill, error) {
	e.executions++
	return e.Paper.Execute(ctx, request)
}

func TestBalanceRejectionPrecedesExchangeExecution(t *testing.T) {
	s, store, prices, _, _ := fixture(t)
	store.balance = decimal.NewFromInt(100)
	exchange := &observedExchange{Paper: Paper{Tickers: prices}}
	s.Exchange = exchange
	_, err := s.Open(context.Background(), "owner", Request{AccountID: testAccount, Symbol: "BTCUSDT", Side: "long"})
	if !errors.Is(err, ErrBalance) || exchange.executions != 0 || len(store.orders) != 1 {
		t.Fatalf("unexpected execution on rejection: executions=%d err=%v", exchange.executions, err)
	}
}

func TestOrdersRejectClientProfiles(t *testing.T) {
	for _, value := range []string{`{}`, `{"daily_loss_limit_percent":"100"}`, `"invalid"`, `0`, `false`, `[]`} {
		_, store, _, engine, h := fixture(t)
		w := post(h, "/v1/orders", strings.TrimSuffix(orderBody, "}")+`,"risk_profile":`+value+`}`)
		assertError(t, w, 400, "risk_profile_not_allowed")
		if len(engine.requests) != 0 || len(store.orders) != 0 || len(store.checks) != 0 {
			t.Fatal("client profile reached execution")
		}
	}
	_, _, _, _, h := fixture(t)
	if w := post(h, "/v1/orders", strings.TrimSuffix(orderBody, "}")+`,"risk_profile":null}`); w.Code != 201 {
		t.Fatalf("null profile should be omitted: %d %s", w.Code, w.Body)
	}
}

func TestOrderUsesStoredProfileAndAuditFailureRollsBack(t *testing.T) {
	for _, blocked := range []bool{false, true} {
		_, store, _, engine, h := fixture(t)
		p := risk.DefaultProfile()
		p.DailyLossLimitPercent = "0.5"
		store.profile, store.failCheck, engine.blocked = &p, true, blocked
		w := post(h, "/v1/orders", orderBody)
		assertError(t, w, 503, "trading_unavailable")
		if len(store.orders) != 0 || len(store.fills) != 0 || len(store.positions) != 0 || store.commits != 0 || len(store.checks) != 0 {
			t.Fatal("order escaped rollback after audit failure")
		}
		if engine.requests[0].RiskProfile.DailyLossLimitPercent != "0.5" {
			t.Fatal("stored profile was not used")
		}
	}
}

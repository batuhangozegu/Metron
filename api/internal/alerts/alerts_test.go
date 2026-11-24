package alerts

import (
	"context"
	"errors"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/batuhangozegu/metron/api/internal/auth"
	"github.com/batuhangozegu/metron/api/internal/market"
	"github.com/go-chi/chi/v5"
	"github.com/shopspring/decimal"
)

type symbolsStub struct{}

func (symbolsStub) Symbols(context.Context) ([]market.Symbol, error) {
	return []market.Symbol{{Symbol: "BTCUSDT"}}, nil
}

type storeStub struct {
	Store
	created   Request
	createErr error
	prices    []string
	since     *time.Time
}

func (s *storeStub) Create(_ context.Context, _ string, r Request) (Alert, error) {
	s.created = r
	return Alert{Price: r.Price}, s.createErr
}
func (s *storeStub) List(_ context.Context, _ string, since *time.Time) ([]Alert, error) {
	s.since = since
	return nil, nil
}
func (s *storeStub) ActiveSymbols(context.Context) ([]string, error) {
	return []string{"BTCUSDT", "ETHUSDT"}, nil
}
func (s *storeStub) Trigger(_ context.Context, symbol string, p decimal.Decimal) ([]Alert, error) {
	s.prices = append(s.prices, symbol+":"+p.String())
	return nil, nil
}

func TestValidation(t *testing.T) {
	valid := Request{Symbol: "BTCUSDT", Direction: "above", Price: "65000.12345678"}
	note := strings.Repeat("ş", 201)
	for _, tc := range []struct {
		name    string
		request Request
	}{
		{"unknown", Request{Symbol: "BTC", Direction: "above", Price: "1"}},
		{"direction", Request{Symbol: "BTCUSDT", Direction: "up", Price: "1"}},
		{"zero", Request{Symbol: "BTCUSDT", Direction: "above", Price: "0"}},
		{"negative", Request{Symbol: "BTCUSDT", Direction: "above", Price: "-1"}},
		{"scale", Request{Symbol: "BTCUSDT", Direction: "above", Price: "1.000000001"}},
		{"exponent", Request{Symbol: "BTCUSDT", Direction: "above", Price: "1e8"}},
		{"nan", Request{Symbol: "BTCUSDT", Direction: "above", Price: "NaN"}},
		{"note", Request{Symbol: "BTCUSDT", Direction: "below", Price: "1", Note: &note}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			store := &storeStub{}
			_, err := (Service{Store: store, Symbols: symbolsStub{}}).Create(context.Background(), "owner", tc.request)
			var validation auth.ValidationError
			if !errors.As(err, &validation) || store.created.Symbol != "" {
				t.Fatalf("validation=%v created=%+v", err, store.created)
			}
		})
	}
	note = strings.Repeat("ş", 200)
	valid.Note = &note
	store := &storeStub{}
	if _, err := (Service{Store: store, Symbols: symbolsStub{}}).Create(context.Background(), "owner", valid); err != nil {
		t.Fatal(err)
	}
}

func TestHTTPValidationAndLimit(t *testing.T) {
	store := &storeStub{createErr: ErrLimit}
	router := chi.NewRouter()
	(Service{Store: store, Symbols: symbolsStub{}}).Mount(router)
	for _, tc := range []struct {
		method, path, body string
		status             int
	}{
		{"POST", "/v1/alerts", `{"symbol":"BTCUSDT","direction":"above","price":"1"}`, 409},
		{"POST", "/v1/alerts", `{"symbol":"BTCUSDT","direction":"above","price":1}`, 400},
		{"GET", "/v1/alerts?triggered_since=bad", "", 400},
		{"GET", "/v1/alerts?triggered_since=", "", 400},
		{"GET", "/v1/alerts?triggered_since=2026-01-01T00:00:00Z", "", 200},
		{"DELETE", "/v1/alerts/bad", "", 404},
	} {
		w := httptest.NewRecorder()
		router.ServeHTTP(w, httptest.NewRequest(tc.method, tc.path, strings.NewReader(tc.body)))
		if w.Code != tc.status {
			t.Fatalf("%s %s: %d %s", tc.method, tc.path, w.Code, w.Body.String())
		}
	}
	if store.since == nil || store.since.Year() != 2026 {
		t.Fatalf("since=%v", store.since)
	}
}

type tickerStub struct {
	fail  bool
	price string
}

func (s *tickerStub) Tickers(_ context.Context, symbols []string) (map[string]map[string]string, error) {
	if len(symbols) != 2 {
		return nil, errors.New("unexpected symbols")
	}
	if s.fail {
		return nil, errors.New("redis down")
	}
	return map[string]map[string]string{"BTCUSDT": {"price": s.price}}, nil
}

func TestCheckerRecoveryAndCancellation(t *testing.T) {
	store := &storeStub{}
	tickers := &tickerStub{fail: true, price: "100.25"}
	checker := Checker{Store: store, Tickers: tickers}
	if err := checker.Check(context.Background()); err == nil {
		t.Fatal("expected Redis error")
	}
	tickers.fail = false
	if err := checker.Check(context.Background()); err != nil {
		t.Fatal(err)
	}
	if len(store.prices) != 1 || store.prices[0] != "BTCUSDT:100.25" {
		t.Fatalf("prices=%v", store.prices)
	}
	for _, price := range []string{"bad", "0", "-1"} {
		tickers.price = price
		if err := checker.Check(context.Background()); err != nil {
			t.Fatal(err)
		}
	}
	if len(store.prices) != 1 {
		t.Fatalf("invalid prices triggered: %v", store.prices)
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	done := make(chan struct{})
	go func() { checker.Run(ctx); close(done) }()
	select {
	case <-done:
	case <-time.After(time.Second):
		t.Fatal("checker did not stop")
	}
}

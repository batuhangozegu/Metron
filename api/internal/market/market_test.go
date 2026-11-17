package market

import (
	"context"
	"encoding/json"
	"errors"
	"net/http/httptest"
	"reflect"
	"testing"
	"time"
)

func ptr(s string) *string { return &s }

func TestPresent(t *testing.T) {
	for _, tc := range []struct {
		name   string
		row    Symbol
		ticker map[string]string
		want   Market
	}{
		{
			name:   "decimal strings, negative change and default stop",
			row:    Symbol{Symbol: "BTCUSDT", BaseAsset: "BTC", QuoteAsset: "USDT", CoinClass: "core", ATR: ptr("1540.20000000"), ATRPercent: ptr("2.4108")},
			ticker: map[string]string{"price": "63880.94000000", "open": "64676.46", "quote_volume": "1845512345.12000000", "event_time_ms": "1794936345264"},
			want:   Market{Symbol: "BTCUSDT", BaseAsset: "BTC", QuoteAsset: "USDT", CoinClass: "core", Price: ptr("63880.94"), Change24hPercent: ptr("-1.23"), QuoteVolume24h: ptr("1845512345.12"), PriceTime: ptr("2026-11-17T17:25:45.264Z"), ATR: ptr("1540.2"), ATRPercent: ptr("2.4108"), SuggestedStopPercent: ptr("3.62")},
		},
		{
			name:   "stable has no stop",
			row:    Symbol{Symbol: "USDCUSDT", CoinClass: "stable", ATRPercent: ptr("0.0200")},
			ticker: map[string]string{"price": "1.00000000", "open": "1", "quote_volume": "0.00000000"},
			want:   Market{Symbol: "USDCUSDT", CoinClass: "stable", Price: ptr("1"), Change24hPercent: ptr("0.00"), QuoteVolume24h: ptr("0"), ATRPercent: ptr("0.02")},
		},
		{
			name: "missing ticker preserves statistics",
			row:  Symbol{Symbol: "ETHUSDT", CoinClass: "core", ATR: ptr("42"), ATRPercent: ptr("2")},
			want: Market{Symbol: "ETHUSDT", CoinClass: "core", ATR: ptr("42"), ATRPercent: ptr("2"), SuggestedStopPercent: ptr("3.00")},
		},
		{
			name:   "zero open cannot divide",
			ticker: map[string]string{"price": "2.5", "open": "0", "event_time_ms": "bad"},
			want:   Market{Price: ptr("2.5")},
		},
		{
			name:   "invalid fields are null",
			row:    Symbol{ATR: ptr("invalid"), ATRPercent: ptr("NaN")},
			ticker: map[string]string{"price": "NaN", "open": "1", "quote_volume": "invalid", "event_time_ms": "9223372036854775807"},
			want:   Market{},
		},
		{
			name:   "large prices stay exact and change rounds half up",
			ticker: map[string]string{"price": "9007199254740993.12500000", "open": "9007199254740993.125"},
			want:   Market{Price: ptr("9007199254740993.125"), Change24hPercent: ptr("0.00")},
		},
		{
			name:   "percentage rounds to two places",
			ticker: map[string]string{"price": "1.00005", "open": "1"},
			want:   Market{Price: ptr("1.00005"), Change24hPercent: ptr("0.01")},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			got := present(tc.row, tc.ticker)
			if !reflect.DeepEqual(got, tc.want) {
				a, _ := json.Marshal(got)
				b, _ := json.Marshal(tc.want)
				t.Fatalf("got %s\nwant %s", a, b)
			}
		})
	}
}

type fakeSymbols struct {
	rows []Symbol
	err  error
}

func (f fakeSymbols) Symbols(context.Context) ([]Symbol, error) { return f.rows, f.err }

type fakeTickers struct {
	keys []string
	err  error
}

func (f *fakeTickers) Tickers(_ context.Context, keys []string) (map[string]map[string]string, error) {
	f.keys = keys
	return nil, f.err
}

func TestHandler(t *testing.T) {
	for _, tc := range []struct {
		name    string
		symbols SymbolSource
		tickers *fakeTickers
		status  int
		count   int
	}{
		{"unconfigured", nil, nil, 503, 0},
		{"database down", fakeSymbols{err: errors.New("secret connection details")}, nil, 503, 0},
		{"empty database", fakeSymbols{}, nil, 200, 0},
		{"missing redis data", fakeSymbols{rows: []Symbol{{Symbol: "BTCUSDT"}, {Symbol: "ETHUSDT"}}}, &fakeTickers{}, 200, 2},
		{"redis down", fakeSymbols{rows: []Symbol{{Symbol: "BTCUSDT"}}}, &fakeTickers{err: errors.New("redis down")}, 200, 1},
	} {
		t.Run(tc.name, func(t *testing.T) {
			var tickers TickerSource
			if tc.tickers != nil {
				tickers = tc.tickers
			}
			w := httptest.NewRecorder()
			Handler(tc.symbols, tickers).ServeHTTP(w, httptest.NewRequest("GET", "/v1/markets", nil))
			if w.Code != tc.status || w.Header().Get("Content-Type") != "application/json" {
				t.Fatalf("status=%d headers=%v", w.Code, w.Header())
			}
			if tc.status == 503 {
				if w.Body.String() != "{\"error\":\"market data unavailable\"}\n" {
					t.Fatal(w.Body.String())
				}
				return
			}
			var response Response
			if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
				t.Fatal(err)
			}
			if response.Markets == nil || len(response.Markets) != tc.count {
				t.Fatalf("markets=%v", response.Markets)
			}
			stamp, err := time.Parse(time.RFC3339, response.AsOf)
			if err != nil || stamp.Location() != time.UTC || time.Since(stamp) > 2*time.Second {
				t.Fatalf("as_of=%s", response.AsOf)
			}
			if tc.tickers != nil && len(tc.tickers.keys) != tc.count {
				t.Fatalf("ticker keys=%v", tc.tickers.keys)
			}
			for _, row := range response.Markets {
				if row.Price != nil || row.Change24hPercent != nil || row.PriceTime != nil {
					t.Fatalf("missing ticker must have null price fields: %+v", row)
				}
			}
		})
	}
}

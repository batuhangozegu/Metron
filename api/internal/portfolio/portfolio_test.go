package portfolio

import (
	"context"
	"errors"
	"net/http/httptest"
	"testing"

	"github.com/batuhangozegu/metron/api/internal/trading"
	"github.com/go-chi/chi/v5"
	"github.com/shopspring/decimal"
)

type fixedPrices struct {
	price decimal.Decimal
	err   error
}

func (p fixedPrices) Price(context.Context, string) (decimal.Decimal, error) { return p.price, p.err }

func TestValuationAndMissingPrice(t *testing.T) {
	s := Service{Prices: fixedPrices{price: decimal.NewFromInt(110)}}
	positions := []openPosition{{Symbol: "BTCUSDT", Side: "long", Quantity: decimal.NewFromInt(2), Entry: decimal.NewFromInt(100)}, {Symbol: "SOLUSDT", Side: "short", Quantity: decimal.NewFromInt(1), Entry: decimal.NewFromInt(100)}}
	v, err := s.value(context.Background(), positions)
	if err != nil || v.String() != "10" {
		t.Fatalf("valuation %s %v", v, err)
	}
	s.Prices = fixedPrices{err: trading.ErrPrice}
	if _, err := s.value(context.Background(), positions); !errors.Is(err, trading.ErrPrice) {
		t.Fatal("missing price was silently valued")
	}
}

func TestQueryValidation(t *testing.T) {
	r := chi.NewRouter()
	(Service{}).Mount(r)
	for _, path := range []string{"portfolio?period=1d", "portfolio?period=", "portfolio?period=all&period=30d", "journal?limit=0", "journal?limit=201", "journal?limit=x", "journal?limit=1&limit=2", "journal?before=x", "journal?before="} {
		w := httptest.NewRecorder()
		r.ServeHTTP(w, httptest.NewRequest("GET", "/v1/accounts/11111111-1111-4111-8111-111111111111/"+path, nil))
		if w.Code != 400 || w.Header().Get("Cache-Control") != "no-store" {
			t.Fatalf("%s: %d %s", path, w.Code, w.Body)
		}
		if path == "portfolio?period=1d" && w.Body.String() != "{\"error\":\"invalid_period\"}\n" {
			t.Fatal(w.Body)
		}
	}
}

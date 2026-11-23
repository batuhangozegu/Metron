package trading

import (
	"github.com/shopspring/decimal"
	"testing"
)

func TestCash(t *testing.T) {
	d := decimal.RequireFromString
	positions := []Capital{{Notional: d("200"), Fee: d("0.2"), Leverage: 10}, {Notional: d("300"), Fee: d("0.3"), Leverage: 5}}
	for market, want := range map[string]string{"spot": "499.5", "futures": "919.5"} {
		if got := Cash(d("1000"), market, positions); got.String() != want {
			t.Fatalf("%s cash %s", market, got)
		}
	}
	if v := Cash(d("1"), "futures", []Capital{{Notional: d("1"), Leverage: 3}}); v.String() != "0.6666666666666666666666666666" {
		t.Fatalf("repeating cash: %s", v)
	}
	if !affordable(d("1"), "futures", []Capital{{Notional: d("1"), Leverage: 3}}, d("2"), decimal.Zero, 3) {
		t.Fatal("exact repeating margin equality rejected")
	}
	if v := Cash(d("1"), "futures", []Capital{{Notional: d("1"), Leverage: 3}, {Notional: d("2"), Leverage: 3}}); !v.IsZero() {
		t.Fatalf("combined repeating margin %s", v)
	}
}

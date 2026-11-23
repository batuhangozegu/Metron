package trading

import (
	"context"
	"errors"
	"testing"

	"github.com/shopspring/decimal"
)

type failingSnapshots struct{ calls int }

func (s *failingSnapshots) Snapshot(context.Context, string) error {
	s.calls++
	return errors.New("snapshot failed")
}

func TestReservedCapitalBalanceRejections(t *testing.T) {
	for _, market := range []string{"spot", "futures"} {
		s, store, _, engine, _ := fixture(t)
		store.market = market
		leverage := uint32(1)
		store.balance = decimal.RequireFromString("400.3")
		if market == "futures" {
			leverage = 10
			store.balance = decimal.RequireFromString("40.15")
		}
		snapshots := &failingSnapshots{}
		s.Snapshots = snapshots
		body := Request{AccountID: testAccount, Symbol: "BTCUSDT", Side: "long", Leverage: &leverage}
		first, err := s.Open(context.Background(), "owner", body)
		if err != nil {
			t.Fatal(err)
		}
		_, err = s.Open(context.Background(), "owner", body)
		if !errors.Is(err, ErrBalance) || len(store.orders) != 2 || store.orders[1].Status != "rejected" || len(store.checks) != 2 || len(store.fills) != 1 || snapshots.calls != 1 {
			t.Fatalf("%s: reserved cash rejection %v %+v", market, err, store.orders)
		}
		// Risk equity remains the settled balance even while capital is reserved.
		if engine.requests[1].AccountEquity != store.balance.String() {
			t.Fatal("risk equity changed")
		}
		if _, err := s.Close(context.Background(), "owner", first.Position.ID); err != nil {
			t.Fatal(err)
		}
		if snapshots.calls != 2 || store.positions[first.Position.ID].Status != "closed" {
			t.Fatal("snapshot failure failed the trade")
		}
	}
}

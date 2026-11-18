package accounts

import (
	"context"
	"strings"
	"testing"

	"github.com/shopspring/decimal"
)

func TestBalanceValidation(t *testing.T) {
	for _, tc := range []struct {
		value string
		valid bool
	}{
		{"1", true}, {"0.00000001", true}, {"10000000", true}, {"10000000.00000000", true}, {"123.12345678", true},
		{"0", false}, {"-1", false}, {"10000000.00000001", false}, {"0.000000001", false},
		{"1.123456789", false}, {"1e5", false}, {"NaN", false}, {"", false}, {" 1", false}, {"+1", false},
		{"1.", false}, {strings.Repeat("9", 100), false},
	} {
		t.Run(tc.value, func(t *testing.T) {
			balance, err := ValidateBalance(tc.value)
			if (err == nil) != tc.valid {
				t.Fatalf("got %s, %v", balance, err)
			}
			if tc.valid && !balance.Equal(decimal.RequireFromString(tc.value)) {
				t.Fatal("precision lost")
			}
		})
	}
}

type fakeStore struct {
	user, name string
	balance    decimal.Decimal
	calls      int
}

func (f *fakeStore) CreatePaper(_ context.Context, user, name, market string, balance decimal.Decimal) (Account, error) {
	f.user = user
	f.name = name
	f.balance = balance
	f.calls++
	value := balance.String()
	return Account{UserID: user, Type: "paper", Market: market, Name: name, QuoteAsset: "USDT", PaperBalance: &value}, nil
}
func (f *fakeStore) List(context.Context, string) ([]Account, error) { return nil, nil }

func TestCreatePaper(t *testing.T) {
	for _, tc := range []struct {
		name, balance string
		valid         bool
	}{
		{" Paper ", "1.12345678", true}, {"", "100", false}, {strings.Repeat("x", 101), "100", false}, {"Paper", "0", false},
	} {
		t.Run(tc.name+tc.balance, func(t *testing.T) {
			f := &fakeStore{}
			s := Service{Store: f}
			a, err := s.Create(context.Background(), "user-1", tc.name, tc.balance, "")
			if (err == nil) != tc.valid {
				t.Fatal(err)
			}
			if tc.valid {
				if f.user != "user-1" || f.name != "Paper" || a.Type != "paper" || a.Market != "spot" || !f.balance.Equal(decimal.RequireFromString(tc.balance)) {
					t.Fatalf("invalid account %+v", a)
				}
			} else if f.calls != 0 {
				t.Fatal("invalid data reached store")
			}
		})
	}
}

func TestAccountMarkets(t *testing.T) {
	for _, market := range []string{"", "spot", "futures", "margin", "SPOT", " futures "} {
		f := &fakeStore{}
		a, err := (Service{Store: f}).Create(context.Background(), "user-1", "Paper", "100", market)
		valid := market == "" || market == "spot" || market == "futures"
		if (err == nil) != valid || (!valid && f.calls != 0) {
			t.Fatalf("market %q: %+v %v", market, a, err)
		}
		if valid {
			want := market
			if want == "" {
				want = "spot"
			}
			if a.Market != want {
				t.Fatalf("market %q got %q", market, a.Market)
			}
		}
	}
}

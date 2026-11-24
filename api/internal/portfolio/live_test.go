package portfolio

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/batuhangozegu/metron/api/internal/accounts"
	"github.com/batuhangozegu/metron/api/internal/binance"
	"github.com/shopspring/decimal"
)

type balanceExchange struct {
	accounts.Exchange
	err error
}

func (f *balanceExchange) Balance(context.Context, string) (binance.Balance, error) {
	return binance.Balance{Free: decimal.NewFromInt(123), Total: decimal.NewFromInt(456)}, f.err
}

type balanceProvider struct{ exchange *balanceExchange }

func (f balanceProvider) Client(context.Context, string) (accounts.Exchange, string, error) {
	return f.exchange, "testnet", nil
}

func TestLiveBalanceFailureIsSanitized(t *testing.T) {
	exchange := &balanceExchange{err: errors.New("private signed URL")}
	s := Service{Live: balanceProvider{exchange}}
	if _, err := s.liveBalance(context.Background(), "id"); err == nil || err.Error() == exchange.err.Error() {
		t.Fatal("exchange error not sanitized")
	}
	exchange.err = nil
	b, err := s.liveBalance(context.Background(), "id")
	if err != nil || b.Free.String() != "123" || b.Total.String() != "456" {
		t.Fatal(b, err)
	}
}

func TestPostgresLivePortfolioAndBestEffortSnapshot(t *testing.T) {
	pool := testPool(t)
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()
	var owner string
	if err := pool.QueryRow(ctx, `INSERT INTO app.users(email,password_hash) VALUES(gen_random_uuid()::text || '@live-portfolio.example.com','test-only') RETURNING id::text`).Scan(&owner); err != nil {
		t.Fatal(err)
	}
	defer func() {
		if _, err := pool.Exec(context.Background(), `DELETE FROM app.users WHERE id=$1::uuid`, owner); err != nil {
			t.Error(err)
		}
	}()
	a, err := (accounts.Postgres{Pool: pool}).CreateLive(ctx, owner, accounts.LiveRequest{Name: "Live portfolio", Market: "futures"}, accounts.Credentials{Environment: "testnet", KeyHint: "•••TEST", Key: []byte{1}, Secret: []byte{1}}, decimal.NewFromInt(400))
	if err != nil {
		t.Fatal(err)
	}
	exchange := &balanceExchange{}
	s := Service{Store: &Postgres{Pool: pool}, Live: balanceProvider{exchange}}
	p, err := s.Portfolio(ctx, owner, a.ID, "all")
	if err != nil || p.Summary.Equity != "456" || p.Summary.Cash != "123" || p.Summary.NetResult != "56" {
		t.Fatalf("portfolio %+v %v", p, err)
	}
	if err = s.Snapshot(ctx, a.ID); err != nil {
		t.Fatal(err)
	}
	exchange.err = binance.ErrUnavailable
	if err = s.Snapshot(ctx, a.ID); err == nil {
		t.Fatal("failed exchange snapshot saved")
	}
	var count int
	if err = pool.QueryRow(ctx, `SELECT count(*) FROM app.equity_snapshots WHERE account_id=$1::uuid`, a.ID).Scan(&count); err != nil || count != 1 {
		t.Fatalf("snapshot count %d %v", count, err)
	}
}

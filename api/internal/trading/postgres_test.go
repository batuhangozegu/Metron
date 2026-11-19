package trading

import (
	"context"
	"errors"
	"os"
	"sync"
	"testing"
	"time"

	"github.com/batuhangozegu/metron/api/internal/accounts"
	enginev1 "github.com/batuhangozegu/metron/api/internal/gen/metron/engine/v1"
	"github.com/batuhangozegu/metron/api/internal/risk"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/shopspring/decimal"
)

// Uses a dedicated migrated database. Cleanup is restricted to the unique test
// user; this test never drops schemas or changes engine-owned market data.
func TestPostgresPaperSettlement(t *testing.T) {
	databaseURL := os.Getenv("TEST_DATABASE_URL")
	if databaseURL == "" {
		t.Skip("set TEST_DATABASE_URL to a migrated test Postgres database")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()
	pool, err := pgxpool.New(ctx, databaseURL)
	if err != nil {
		t.Fatal(err)
	}
	defer pool.Close()
	var owner string
	if err := pool.QueryRow(ctx, `INSERT INTO app.users(email,password_hash) VALUES($1,'test-only') RETURNING id::text`, newID()+"@paper.example.com").Scan(&owner); err != nil {
		t.Fatal(err)
	}
	defer func() {
		if _, err := pool.Exec(context.Background(), `DELETE FROM app.users WHERE id=$1::uuid`, owner); err != nil {
			t.Error(err)
		}
	}()
	account, err := (accounts.Postgres{Pool: pool}).CreatePaper(ctx, owner, "Paper settlement", "spot", decimal.NewFromInt(1000))
	if err != nil {
		t.Fatal(err)
	}
	engine := &testEngine{}
	s := Service{Store: Postgres{Pool: pool}, Exchange: Paper{Tickers: &ticker{value: "100"}}, Risk: risk.Service{Engine: engine}}
	open := func() Execution {
		t.Helper()
		e, err := s.Open(ctx, owner, Request{AccountID: account.ID, Symbol: "BTCUSDT", Side: "long"})
		if err != nil {
			t.Fatal(err)
		}
		return e
	}
	first, second := open(), open()
	if engine.requests[1].Activity.OpenPositions != 1 || engine.requests[1].Activity.OpenRiskPercent != "1" {
		t.Fatalf("wrong SQL activity %+v", engine.requests[1])
	}
	// Exercise the real account/position locks with simultaneous manual closes.
	var wg sync.WaitGroup
	errs := make(chan error, 8)
	for range 8 {
		wg.Go(func() { _, err := s.Close(ctx, owner, first.Position.ID); errs <- err })
	}
	wg.Wait()
	close(errs)
	successes := 0
	for err := range errs {
		if err == nil {
			successes++
		} else if !errors.Is(err, ErrClosed) {
			t.Fatal(err)
		}
	}
	if successes != 1 {
		t.Fatalf("concurrent close successes %d", successes)
	}
	for range 2 {
		if err := s.Stop(ctx, &enginev1.StopHit{PositionId: second.Position.ID}); err != nil {
			t.Fatal(err)
		}
	}
	var balance string
	if err := pool.QueryRow(ctx, `SELECT paper_balance::text FROM app.accounts WHERE id=$1::uuid`, account.ID).Scan(&balance); err != nil {
		t.Fatal(err)
	}
	if !decimal.RequireFromString(balance).Equal(decimal.RequireFromString("989.21")) {
		t.Fatalf("balance=%s", balance)
	}
	var orders, fills, closed int
	if err := pool.QueryRow(ctx, `SELECT (SELECT count(*) FROM app.orders WHERE account_id=$1::uuid),
		(SELECT count(*) FROM app.fills f JOIN app.orders o ON o.id=f.order_id WHERE o.account_id=$1::uuid),
		(SELECT count(*) FROM app.positions WHERE account_id=$1::uuid AND status='closed')`, account.ID).Scan(&orders, &fills, &closed); err != nil {
		t.Fatal(err)
	}
	if orders != 4 || fills != 4 || closed != 2 {
		t.Fatalf("orders=%d fills=%d positions=%d", orders, fills, closed)
	}
	a, err := (risk.Postgres{Pool: pool}).RiskAccount(ctx, owner, account.ID, "BTCUSDT")
	if err != nil {
		t.Fatal(err)
	}
	wantLoss := decimal.RequireFromString("10.79").Mul(decimal.NewFromInt(100)).DivRound(decimal.RequireFromString("989.21"), 28).String()
	if a.Activity.OpenPositions != 0 || a.Activity.LosingTradesToday != 2 || a.Activity.StoppedTradesThisWeek != 1 || a.Activity.DailyLossPercent != wantLoss || a.Activity.WeeklyLossPercent != wantLoss {
		t.Fatalf("wrong settled SQL activity %+v", a.Activity)
	}
	list, err := s.Store.Positions(ctx, owner, account.ID, "closed")
	if err != nil || len(list) != 2 || list[0].ID != second.Position.ID || list[0].ClosedAt.Location() != time.UTC {
		t.Fatalf("closed list %+v %v", list, err)
	}
	// A loss can exhaust equity and fee precision must never be rounded to fit
	// the account creation constraint.
	if _, err := pool.Exec(ctx, `UPDATE app.accounts SET paper_balance=-0.0000000000000000005 WHERE id=$1::uuid`, account.ID); err != nil {
		t.Fatal(err)
	}
	if _, err := s.Close(ctx, newID(), first.Position.ID); !errors.Is(err, ErrNotFound) {
		t.Fatalf("foreign close %v", err)
	}
	if _, err := s.Store.Positions(ctx, newID(), account.ID, "closed"); !errors.Is(err, risk.ErrAccountNotFound) {
		t.Fatalf("foreign list %v", err)
	}
	// Verify owner filtering occurs before positions are used for risk activity.
	if _, err := (risk.Postgres{Pool: pool}).RiskAccount(ctx, newID(), account.ID, "BTCUSDT"); !errors.Is(err, risk.ErrAccountNotFound) {
		t.Fatalf("foreign activity %v", err)
	}
	// A missing target symbol must remain the engine's NotFound decision rather
	// than exposing a database error.
	var unused string
	err = pool.QueryRow(ctx, `SELECT symbol FROM market.symbols WHERE symbol='MISSINGPAPERUSDT'`).Scan(&unused)
	if errors.Is(err, pgx.ErrNoRows) {
		if _, err := (risk.Postgres{Pool: pool}).RiskAccount(ctx, owner, account.ID, "MISSINGPAPERUSDT"); err != nil {
			t.Fatal(err)
		}
	}
}

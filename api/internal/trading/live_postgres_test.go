package trading

import (
	"context"
	"encoding/base64"
	"errors"
	"os"
	"testing"
	"time"

	"github.com/batuhangozegu/metron/api/internal/accounts"
	"github.com/batuhangozegu/metron/api/internal/binance"
	"github.com/batuhangozegu/metron/api/internal/cryptobox"
	"github.com/batuhangozegu/metron/api/internal/risk"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/shopspring/decimal"
)

func TestPostgresLiveStopsCredentialsAndDeletion(t *testing.T) {
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
	var owner string
	if err = pool.QueryRow(ctx, `INSERT INTO app.users(email,password_hash) VALUES(gen_random_uuid()::text || '@live.example.com','test-only') RETURNING id::text`).Scan(&owner); err != nil {
		t.Fatal(err)
	}
	defer func() {
		if _, err := pool.Exec(context.Background(), `DELETE FROM app.users WHERE id=$1::uuid`, owner); err != nil {
			t.Error(err)
		}
	}()
	_, _, exchange, engine := liveFixture(t)
	box, err := cryptobox.New(base64.StdEncoding.EncodeToString(make([]byte, 32)))
	if err != nil {
		t.Fatal(err)
	}
	manager := &accounts.LiveService{Store: accounts.Postgres{Pool: pool}, Box: box, Factory: func(binance.Config) (accounts.Exchange, error) { return exchange, nil }}
	account, err := manager.Create(ctx, owner, accounts.LiveRequest{Name: "Integration live", Market: "spot", Environment: "testnet", APIKey: "integration-key-ABCD", APISecret: "integration-secret"})
	if err != nil {
		t.Fatal(err)
	}
	credentials, err := (accounts.Postgres{Pool: pool}).Credentials(ctx, account.ID)
	if err != nil {
		t.Fatal(err)
	}
	plain, err := box.Open(credentials.Secret, "binance-api-secret")
	if err != nil || string(plain) != "integration-secret" {
		t.Fatal("credentials not encrypted", err)
	}
	s := Service{Store: Postgres{Pool: pool}, Live: manager, Exchange: Paper{}, Risk: risk.Service{Engine: engine}}
	exchange.stopErr = binance.ErrRejected
	opened, err := s.Open(ctx, owner, Request{AccountID: account.ID, Symbol: "BTCUSDT", Side: "long"})
	if err != nil {
		t.Fatal(err)
	}
	if opened.Position.ExchangeStop.Status != "failed" {
		t.Fatal("stop failure not persisted")
	}
	if err = manager.Delete(ctx, owner, account.ID); !errors.Is(err, accounts.ErrDeleteConflict) {
		t.Fatalf("deleted open account: %v", err)
	}
	exchange.stopErr = nil
	if _, err = s.RetryStop(ctx, owner, opened.Position.ID); err != nil {
		t.Fatal(err)
	}
	positions, err := s.Store.Positions(ctx, owner, account.ID, "open")
	if err != nil || len(positions) != 1 || positions[0].ExchangeStop.Status != "placed" || positions[0].ExchangeEntryOrderID == nil {
		t.Fatalf("positions %+v %v", positions, err)
	}
	if _, err = s.RetryStop(ctx, newID(), opened.Position.ID); !errors.Is(err, ErrNotFound) {
		t.Fatalf("foreign stop retry: %v", err)
	}
	exchange.stopResult = binance.OrderResult{OrderID: 42, Status: "FILLED", ExecutedQty: decimal.NewFromInt(2), AvgPrice: decimal.NewFromInt(94), Fee: decimal.RequireFromString("0.11"), FeeAsset: "USDT"}
	if err = s.Reconcile(ctx, opened.Position.ID); err != nil {
		t.Fatal(err)
	}
	positions, err = s.Store.Positions(ctx, owner, account.ID, "closed")
	if err != nil || len(positions) != 1 || *positions[0].CloseReason != "exchange_stop" || positions[0].ExchangeStop.Status != "filled" {
		t.Fatalf("settled %+v %v", positions, err)
	}
	preview, err := (RiskAccounts{Store: Postgres{Pool: pool}, Live: manager}).RiskAccount(ctx, owner, account.ID, "BTCUSDT")
	if err != nil || preview.Activity.StoppedTradesThisWeek != 1 || preview.Equity != "6000" {
		t.Fatalf("live activity %+v %v", preview, err)
	}
	var paperBalance *string
	if err = pool.QueryRow(ctx, `SELECT paper_balance::text FROM app.accounts WHERE id=$1::uuid`, account.ID).Scan(&paperBalance); err != nil || paperBalance != nil {
		t.Fatal("paper balance changed", err)
	}
	if err = manager.Delete(ctx, newID(), account.ID); !errors.Is(err, accounts.ErrAccountNotFound) {
		t.Fatal(err)
	}
	if err = manager.Delete(ctx, owner, account.ID); err != nil {
		t.Fatal(err)
	}
	if _, _, err = manager.Client(ctx, account.ID); err == nil {
		t.Fatal("deleted client cached")
	}
	var count int
	if err = pool.QueryRow(ctx, `SELECT count(*) FROM app.exchange_credentials WHERE account_id=$1::uuid`, account.ID).Scan(&count); err != nil || count != 0 {
		t.Fatal("credentials not cascaded", err)
	}
}

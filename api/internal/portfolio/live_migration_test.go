package portfolio

import (
	"context"
	"os"
	"strings"
	"testing"

	"github.com/jackc/pgx/v5"
)

func TestPostgresLiveMigrationUpAndDown(t *testing.T) {
	pool := testPool(t)
	ctx := context.Background()
	tx, err := pool.Begin(ctx)
	if err != nil {
		t.Fatal(err)
	}
	defer tx.Rollback(ctx)
	var name string
	if err = tx.QueryRow(ctx, `SELECT 'live_' || replace(gen_random_uuid()::text,'-','')`).Scan(&name); err != nil {
		t.Fatal(err)
	}
	schema := pgx.Identifier{name}.Sanitize()
	if _, err = tx.Exec(ctx, `CREATE SCHEMA `+schema+`; CREATE TABLE `+schema+`.accounts(id uuid PRIMARY KEY);
 CREATE TABLE `+schema+`.positions(close_reason text CONSTRAINT positions_close_reason_check CHECK(close_reason IN ('manual','stop')));`); err != nil {
		t.Fatal(err)
	}
	data, err := os.ReadFile("../../../db/app/00006_live_accounts.sql")
	if err != nil {
		t.Fatal(err)
	}
	parts := strings.Split(string(data), "-- +goose Down")
	if len(parts) != 2 {
		t.Fatal("missing Down migration")
	}
	if _, err = tx.Exec(ctx, strings.ReplaceAll(parts[0], "app.", schema+".")); err != nil {
		t.Fatal(err)
	}
	if _, err = tx.Exec(ctx, `INSERT INTO `+schema+`.accounts VALUES('11111111-1111-4111-8111-111111111111');
 INSERT INTO `+schema+`.exchange_credentials(account_id,environment,key_hint,api_key_enc,api_secret_enc) VALUES('11111111-1111-4111-8111-111111111111','testnet','hint','encrypted-key','encrypted-secret');
 INSERT INTO `+schema+`.positions(close_reason,exchange_entry_order_id,exchange_stop_order_id,exchange_stop_status) VALUES('exchange_stop',1,2,'filled');`); err != nil {
		t.Fatal(err)
	}
	if _, err = tx.Exec(ctx, `SAVEPOINT invalid_environment`); err != nil {
		t.Fatal(err)
	}
	if _, err = tx.Exec(ctx, `UPDATE `+schema+`.exchange_credentials SET environment='invalid'`); err == nil {
		t.Fatal("environment constraint missing")
	}
	if _, err = tx.Exec(ctx, `ROLLBACK TO SAVEPOINT invalid_environment`); err != nil {
		t.Fatal(err)
	}
	if _, err = tx.Exec(ctx, `DELETE FROM `+schema+`.accounts`); err != nil {
		t.Fatal(err)
	}
	var count int
	if err = tx.QueryRow(ctx, `SELECT count(*) FROM `+schema+`.exchange_credentials`).Scan(&count); err != nil || count != 0 {
		t.Fatal("credential cascade missing", err)
	}
	if _, err = tx.Exec(ctx, strings.ReplaceAll(parts[1], "app.", schema+".")); err != nil {
		t.Fatal(err)
	}
	var reason string
	if err = tx.QueryRow(ctx, `SELECT close_reason FROM `+schema+`.positions`).Scan(&reason); err != nil || reason != "stop" {
		t.Fatal("Down did not preserve stop reason", err)
	}
}

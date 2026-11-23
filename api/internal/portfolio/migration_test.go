package portfolio

import (
	"context"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"os"
	"strings"
	"testing"
)

func testPool(t *testing.T) *pgxpool.Pool {
	t.Helper()
	databaseURL := os.Getenv("TEST_DATABASE_URL")
	if databaseURL == "" {
		t.Skip("set TEST_DATABASE_URL to a migrated test Postgres database")
	}
	pool, err := pgxpool.New(context.Background(), databaseURL)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(pool.Close)
	return pool
}

func TestPostgresMigrationBackfillAndDown(t *testing.T) {
	pool := testPool(t)
	ctx := context.Background()
	tx, err := pool.Begin(ctx)
	if err != nil {
		t.Fatal(err)
	}
	defer tx.Rollback(ctx)
	// Exercise the actual migration in a disposable transaction-local schema.
	// No app tables, existing users, or engine-owned data are changed.
	var name string
	if err := tx.QueryRow(ctx, `SELECT 'portfolio_' || replace(gen_random_uuid()::text,'-','')`).Scan(&name); err != nil {
		t.Fatal(err)
	}
	schema := pgx.Identifier{name}.Sanitize()
	if _, err := tx.Exec(ctx, `CREATE SCHEMA `+schema+`; CREATE TABLE `+schema+`.accounts(id uuid PRIMARY KEY,type text,paper_balance numeric);
		CREATE TABLE `+schema+`.positions(account_id uuid,status text,realized_pnl numeric);
		INSERT INTO `+schema+`.accounts VALUES ('11111111-1111-4111-8111-111111111111','paper',1200),('22222222-2222-4222-8222-222222222222','paper',500),('33333333-3333-4333-8333-333333333333','live',NULL);
		INSERT INTO `+schema+`.positions VALUES ('11111111-1111-4111-8111-111111111111','closed',250),('11111111-1111-4111-8111-111111111111','closed',-50),('11111111-1111-4111-8111-111111111111','open',999);`); err != nil {
		t.Fatal(err)
	}
	file, err := os.ReadFile("../../../db/app/00005_portfolio.sql")
	if err != nil {
		t.Fatal(err)
	}
	parts := strings.Split(string(file), "-- +goose Down")
	if len(parts) != 2 {
		t.Fatal("missing Down")
	}
	if _, err := tx.Exec(ctx, strings.ReplaceAll(parts[0], "app.", schema+".")); err != nil {
		t.Fatal(err)
	}
	rows, err := tx.Query(ctx, `SELECT starting_balance::text FROM `+schema+`.accounts ORDER BY id`)
	if err != nil {
		t.Fatal(err)
	}
	var values []*string
	for rows.Next() {
		var value *string
		if err := rows.Scan(&value); err != nil {
			t.Fatal(err)
		}
		values = append(values, value)
	}
	if err := rows.Err(); err != nil {
		t.Fatal(err)
	}
	rows.Close()
	if len(values) != 3 || *values[0] != "1000" || *values[1] != "500" || values[2] != nil {
		t.Fatalf("wrong backfill: %+v", values)
	}
	if _, err := tx.Exec(ctx, strings.ReplaceAll(parts[1], "app.", schema+".")); err != nil {
		t.Fatal(err)
	}
}

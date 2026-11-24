package alerts

import (
	"context"
	"errors"
	"fmt"
	"os"
	"sync"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/shopspring/decimal"
)

// Uses an already migrated dedicated database; cleanup only deletes test users.
func TestPostgresAlerts(t *testing.T) {
	url := os.Getenv("TEST_DATABASE_URL")
	if url == "" {
		t.Skip("set TEST_DATABASE_URL to a migrated test Postgres database")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	pool, err := pgxpool.New(ctx, url)
	if err != nil {
		t.Fatal(err)
	}
	defer pool.Close()
	users := []string{}
	defer func() {
		for _, id := range users {
			if _, err := pool.Exec(context.Background(), `DELETE FROM app.users WHERE id=$1::uuid`, id); err != nil {
				t.Error(err)
			}
		}
	}()
	for range 2 {
		var id string
		err := pool.QueryRow(ctx, `INSERT INTO app.users(email,password_hash) VALUES($1,'test-only') RETURNING id::text`, fmt.Sprintf("alert-%d-%d@example.com", time.Now().UnixNano(), len(users))).Scan(&id)
		if err != nil {
			t.Fatal(err)
		}
		users = append(users, id)
	}
	store := Postgres{Pool: pool}
	// A unique symbol isolates triggering from alerts belonging to other tests/users.
	symbol := "ALERTTEST" + users[0]
	create := func(direction, price string) Alert {
		t.Helper()
		a, err := store.Create(ctx, users[0], Request{Symbol: symbol, Direction: direction, Price: price})
		if err != nil {
			t.Fatal(err)
		}
		return a
	}
	above, below := create("above", "100"), create("below", "100")
	create("above", "101")
	create("below", "99")
	if err := store.Cancel(ctx, users[1], above.ID); !errors.Is(err, ErrNotFound) {
		t.Fatalf("foreign cancel=%v", err)
	}
	foreign, err := store.List(ctx, users[1], nil)
	if err != nil || len(foreign) != 0 {
		t.Fatalf("foreign list=%v %v", foreign, err)
	}
	before := time.Now().Add(-time.Second)
	triggered, err := store.Trigger(ctx, symbol, decimal.NewFromInt(100))
	if err != nil || len(triggered) != 2 {
		t.Fatalf("equality=%v %v", triggered, err)
	}
	for _, a := range triggered {
		if (a.ID != above.ID && a.ID != below.ID) || a.TriggeredAt == nil || a.TriggeredPrice == nil || *a.TriggeredPrice != "100" || a.Status != "triggered" {
			t.Fatalf("trigger=%+v", a)
		}
	}
	again, err := store.Trigger(ctx, symbol, decimal.NewFromInt(100))
	if err != nil || len(again) != 0 {
		t.Fatalf("repeat=%v %v", again, err)
	}
	filtered, err := store.List(ctx, users[0], &before)
	if err != nil || len(filtered) != 2 {
		t.Fatalf("filter=%v %v", filtered, err)
	}
	// Exact cursor equality is excluded, as required by "after".
	for _, a := range filtered {
		list, err := store.List(ctx, users[0], a.TriggeredAt)
		if err != nil {
			t.Fatal(err)
		}
		for _, b := range list {
			if !b.TriggeredAt.After(*a.TriggeredAt) {
				t.Fatal("non-strict timestamp filter")
			}
		}
	}
	for _, price := range []int64{102, 98} {
		hits, err := store.Trigger(ctx, symbol, decimal.NewFromInt(price))
		if err != nil || len(hits) != 1 {
			t.Fatalf("crossing %d: %v %v", price, hits, err)
		}
	}
	cancelled := create("above", "1")
	for range 2 {
		if err := store.Cancel(ctx, users[0], cancelled.ID); err != nil {
			t.Fatal(err)
		}
	}
	hits, err := store.Trigger(ctx, symbol, decimal.NewFromInt(2))
	if err != nil || len(hits) != 0 {
		t.Fatalf("cancelled triggered %v %v", hits, err)
	}
	for range 49 {
		create("above", "1000")
	}
	// Several simultaneous creates race for the final available slot.
	var wg sync.WaitGroup
	results := make(chan error, 8)
	for range 8 {
		wg.Go(func() {
			_, err := store.Create(ctx, users[0], Request{Symbol: symbol, Direction: "above", Price: "1000"})
			results <- err
		})
	}
	wg.Wait()
	close(results)
	success := 0
	for err := range results {
		if err == nil {
			success++
		} else if !errors.Is(err, ErrLimit) {
			t.Fatal(err)
		}
	}
	if success != 1 {
		t.Fatalf("cap admitted %d", success)
	}
	list, err := store.List(ctx, users[0], nil)
	if err != nil {
		t.Fatal(err)
	}
	active := 0
	for i, a := range list {
		if a.Status == "active" {
			active++
		}
		if i > 0 && list[i-1].CreatedAt.Before(a.CreatedAt) {
			t.Fatal("not newest first")
		}
	}
	if active != 50 {
		t.Fatalf("active=%d", active)
	}
	// The database rejects excessive scale rather than silently rounding it.
	for _, price := range []string{"0", "-1", "1.000000001", "NaN", "Infinity"} {
		if _, err := pool.Exec(ctx, `INSERT INTO app.price_alerts(user_id,symbol,direction,price) VALUES($1::uuid,$2,'above',$3::numeric)`, users[1], symbol, price); err == nil {
			t.Fatalf("database accepted %s", price)
		}
	}
}

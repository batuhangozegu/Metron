package portfolio

import (
	"context"
	"log/slog"
	"time"

	"github.com/shopspring/decimal"
)

type Service struct {
	Store  *Postgres
	Prices interface {
		Price(context.Context, string) (decimal.Decimal, error)
	}
	Now func() time.Time
	Log *slog.Logger
}

func (s Service) now() time.Time {
	if s.Now != nil {
		return s.Now().UTC()
	}
	return time.Now().UTC()
}

func (s Service) logger() *slog.Logger {
	if s.Log != nil {
		return s.Log
	}
	return slog.Default()
}

// Snapshot serializes with trades so balance and open positions agree. Prices
// must exist for every position: a partial valuation is never stored.
func (s Service) Snapshot(ctx context.Context, id string) error {
	tx, err := s.Store.Pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(context.WithoutCancel(ctx))
	a, err := loadAccount(ctx, tx, nil, id, true)
	if err != nil {
		return err
	}
	positions, err := loadOpen(ctx, tx, id)
	if err != nil {
		return err
	}
	unrealized, err := s.value(ctx, positions)
	if err != nil {
		return err
	}
	_, err = tx.Exec(ctx, `INSERT INTO app.equity_snapshots(account_id,taken_at,balance,unrealized_pnl,equity)
		VALUES($1::uuid,$2,$3::numeric,$4::numeric,$5::numeric)
		ON CONFLICT(account_id,taken_at) DO UPDATE SET balance=EXCLUDED.balance,
		unrealized_pnl=EXCLUDED.unrealized_pnl,equity=EXCLUDED.equity`,
		id, s.now(), a.balance.String(), unrealized.String(), a.balance.Add(unrealized).String())
	if err != nil {
		return err
	}
	return tx.Commit(ctx)
}

func (s Service) snapshotAll(ctx context.Context) {
	queryCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	rows, err := s.Store.Pool.Query(queryCtx, `SELECT id::text FROM app.accounts WHERE type='paper' ORDER BY id`)
	if err != nil {
		cancel()
		s.logger().Warn("equity snapshot accounts unavailable", "err", err)
		return
	}
	var ids []string
	for rows.Next() {
		var id string
		if err = rows.Scan(&id); err != nil {
			break
		}
		ids = append(ids, id)
	}
	if err == nil {
		err = rows.Err()
	}
	rows.Close()
	cancel()
	if err != nil {
		s.logger().Warn("equity snapshot accounts unavailable", "err", err)
		return
	}
	for _, id := range ids {
		if ctx.Err() != nil {
			return
		}
		writeCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
		err := s.Snapshot(writeCtx, id)
		cancel()
		if err != nil {
			s.logger().Warn("equity snapshot skipped", "account_id", id, "err", err)
		}
	}
}

// RunSnapshots runs at startup and then at full UTC hours, until cancellation.
func (s Service) RunSnapshots(ctx context.Context) {
	if s.Store == nil {
		return
	}
	for {
		if ctx.Err() != nil {
			return
		}
		s.snapshotAll(ctx)
		now := time.Now().UTC()
		timer := time.NewTimer(now.Truncate(time.Hour).Add(time.Hour).Sub(now))
		select {
		case <-ctx.Done():
			timer.Stop()
			return
		case <-timer.C:
		}
	}
}

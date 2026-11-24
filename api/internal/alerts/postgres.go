package alerts

import (
	"context"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/shopspring/decimal"
)

type Postgres struct{ Pool *pgxpool.Pool }

const columns = `id::text,symbol,direction,price::text,note,status,created_at,triggered_at,triggered_price::text`

func scanAlert(row pgx.Row) (Alert, error) {
	var a Alert
	err := row.Scan(&a.ID, &a.Symbol, &a.Direction, &a.Price, &a.Note, &a.Status, &a.CreatedAt, &a.TriggeredAt, &a.TriggeredPrice)
	a.CreatedAt = a.CreatedAt.UTC()
	if a.TriggeredAt != nil {
		stamp := a.TriggeredAt.UTC()
		a.TriggeredAt = &stamp
	}
	if err == nil {
		p, parseErr := decimal.NewFromString(a.Price)
		if parseErr != nil {
			return Alert{}, parseErr
		}
		a.Price = p.String()
		if a.TriggeredPrice != nil {
			p, parseErr := decimal.NewFromString(*a.TriggeredPrice)
			if parseErr != nil {
				return Alert{}, parseErr
			}
			value := p.String()
			a.TriggeredPrice = &value
		}
	}
	return a, err
}

func (p Postgres) Create(ctx context.Context, userID string, r Request) (Alert, error) {
	tx, err := p.Pool.Begin(ctx)
	if err != nil {
		return Alert{}, err
	}
	defer tx.Rollback(ctx)
	// Serialize creates per user before counting, including concurrent requests.
	var owner string
	if err := tx.QueryRow(ctx, `SELECT id::text FROM app.users WHERE id=$1::uuid FOR UPDATE`, userID).Scan(&owner); err != nil {
		return Alert{}, err
	}
	var count int
	if err := tx.QueryRow(ctx, `SELECT count(*) FROM app.price_alerts WHERE user_id=$1::uuid AND status='active'`, userID).Scan(&count); err != nil {
		return Alert{}, err
	}
	if count >= 50 {
		return Alert{}, ErrLimit
	}
	a, err := scanAlert(tx.QueryRow(ctx, `INSERT INTO app.price_alerts(user_id,symbol,direction,price,note) VALUES($1::uuid,$2,$3,$4::numeric,$5) RETURNING `+columns, userID, r.Symbol, r.Direction, r.Price, r.Note))
	if err != nil {
		return Alert{}, err
	}
	return a, tx.Commit(ctx)
}

func (p Postgres) List(ctx context.Context, userID string, since *time.Time) ([]Alert, error) {
	rows, err := p.Pool.Query(ctx, `SELECT `+columns+` FROM app.price_alerts WHERE user_id=$1::uuid AND ($2::timestamptz IS NULL OR (status='triggered' AND triggered_at > $2)) ORDER BY created_at DESC,id DESC LIMIT 200`, userID, since)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return collect(rows)
}

func collect(rows pgx.Rows) ([]Alert, error) {
	result := []Alert{}
	for rows.Next() {
		a, err := scanAlert(rows)
		if err != nil {
			return nil, err
		}
		result = append(result, a)
	}
	return result, rows.Err()
}

func (p Postgres) Cancel(ctx context.Context, userID, id string) error {
	// Cancellation is owner-filtered and idempotent.
	var found string
	err := p.Pool.QueryRow(ctx, `UPDATE app.price_alerts SET status='cancelled',cancelled_at=COALESCE(cancelled_at,now()) WHERE id=$1::uuid AND user_id=$2::uuid RETURNING id::text`, id, userID).Scan(&found)
	if errors.Is(err, pgx.ErrNoRows) {
		return ErrNotFound
	}
	return err
}

func (p Postgres) ActiveSymbols(ctx context.Context) ([]string, error) {
	rows, err := p.Pool.Query(ctx, `SELECT DISTINCT symbol FROM app.price_alerts WHERE status='active' ORDER BY symbol`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	result := []string{}
	for rows.Next() {
		var symbol string
		if err := rows.Scan(&symbol); err != nil {
			return nil, err
		}
		result = append(result, symbol)
	}
	return result, rows.Err()
}

func (p Postgres) Trigger(ctx context.Context, symbol string, price decimal.Decimal) ([]Alert, error) {
	rows, err := p.Pool.Query(ctx, `UPDATE app.price_alerts SET status='triggered',triggered_at=clock_timestamp(),triggered_price=$2::numeric WHERE symbol=$1 AND status='active' AND ((direction='above' AND $2::numeric >= price) OR (direction='below' AND $2::numeric <= price)) RETURNING `+columns, symbol, price.String())
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return collect(rows)
}

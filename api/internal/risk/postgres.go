package risk

import (
	"context"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/shopspring/decimal"
)

type Postgres struct{ Pool *pgxpool.Pool }

// Querier is implemented by both a pool and a trading transaction.
type Querier interface {
	QueryRow(context.Context, string, ...any) pgx.Row
	Query(context.Context, string, ...any) (pgx.Rows, error)
}

func (p Postgres) RiskAccount(ctx context.Context, userID, accountID, symbol string) (Account, error) {
	account, err := LoadAccount(ctx, p.Pool, userID, accountID, false)
	if err != nil {
		return Account{}, err
	}
	profile, err := LoadProfile(ctx, p.Pool, accountID)
	if err != nil {
		return Account{}, err
	}
	account.Profile = &profile.Profile
	return LoadActivity(ctx, p.Pool, accountID, symbol, account, time.Now())
}

// LoadAccount optionally locks the account. Every trading mutation takes this
// lock first, serializing risk snapshots, opens and closes for the account.
func LoadAccount(ctx context.Context, q Querier, userID, accountID string, lock bool) (Account, error) {
	var value string
	var account Account
	query := `SELECT paper_balance::text,market FROM app.accounts
		WHERE user_id=$1::uuid AND id=$2::uuid AND type='paper' AND quote_asset='USDT' AND paper_balance IS NOT NULL`
	if lock {
		query += ` FOR UPDATE`
	}
	err := q.QueryRow(ctx, query, userID, accountID).Scan(&value, &account.Market)
	if errors.Is(err, pgx.ErrNoRows) {
		return Account{}, ErrAccountNotFound
	}
	if err != nil {
		return Account{}, err
	}
	balance, err := decimal.NewFromString(value)
	if err != nil {
		return Account{}, err
	}
	account.Equity = balance.String()
	return account, nil
}

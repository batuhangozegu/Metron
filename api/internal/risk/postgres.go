package risk

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/shopspring/decimal"
)

type Postgres struct{ Pool *pgxpool.Pool }

func (p Postgres) RiskAccount(ctx context.Context, userID, accountID string) (Account, error) {
	var value string
	var account Account
	err := p.Pool.QueryRow(ctx, `SELECT paper_balance::text,market FROM app.accounts
		WHERE user_id=$1 AND id=$2::uuid AND type='paper' AND quote_asset='USDT' AND paper_balance IS NOT NULL`, userID, accountID).Scan(&value, &account.Market)
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

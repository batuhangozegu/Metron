package accounts

import (
	"context"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/shopspring/decimal"
)

type Postgres struct{ Pool *pgxpool.Pool }

const columns = `id::text,user_id::text,type,market,name,quote_asset,paper_balance::text,created_at`

func scanAccount(row pgx.Row) (Account, error) {
	var a Account
	err := row.Scan(&a.ID, &a.UserID, &a.Type, &a.Market, &a.Name, &a.QuoteAsset, &a.PaperBalance, &a.CreatedAt)
	a.CreatedAt = a.CreatedAt.UTC()
	if a.PaperBalance != nil {
		balance, parseErr := decimal.NewFromString(*a.PaperBalance)
		if parseErr != nil {
			return Account{}, parseErr
		}
		value := balance.String()
		a.PaperBalance = &value
	}
	return a, err
}

func (p Postgres) CreatePaper(ctx context.Context, userID, name, market string, balance decimal.Decimal) (Account, error) {
	return scanAccount(p.Pool.QueryRow(ctx, `INSERT INTO app.accounts(user_id,type,name,market,paper_balance) VALUES($1,'paper',$2,$3,$4::numeric) RETURNING `+columns, userID, name, market, balance.String()))
}

func (p Postgres) List(ctx context.Context, userID string) ([]Account, error) {
	rows, err := p.Pool.Query(ctx, `SELECT `+columns+` FROM app.accounts WHERE user_id=$1 ORDER BY created_at,id`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	accounts := []Account{}
	for rows.Next() {
		account, err := scanAccount(rows)
		if err != nil {
			return nil, err
		}
		accounts = append(accounts, account)
	}
	return accounts, rows.Err()
}

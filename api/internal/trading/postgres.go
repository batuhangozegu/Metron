package trading

import (
	"context"
	"errors"
	"time"

	enginev1 "github.com/batuhangozegu/metron/api/internal/gen/metron/engine/v1"
	"github.com/batuhangozegu/metron/api/internal/risk"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/shopspring/decimal"
)

type Postgres struct{ Pool *pgxpool.Pool }

type postgresTx struct {
	pgx.Tx
	accountID string
	account   risk.Account
	live      bool
}

func (p Postgres) Begin(ctx context.Context, owner, accountID string) (Transaction, error) {
	tx, err := p.Pool.Begin(ctx)
	if err != nil {
		return nil, err
	}
	a, live, err := loadTradingAccount(ctx, tx, owner, accountID, true)
	if err != nil {
		_ = tx.Rollback(context.WithoutCancel(ctx))
		return nil, err
	}
	return &postgresTx{Tx: tx, accountID: accountID, account: a, live: live}, nil
}

func (p *postgresTx) Account() risk.Account { return p.account }
func (p *postgresTx) Capital(ctx context.Context) ([]Capital, error) {
	return LoadCapital(ctx, p.Tx, p.accountID)
}
func (p *postgresTx) Snapshot(ctx context.Context, symbol string) (risk.Account, error) {
	profile, err := risk.LoadProfile(ctx, p.Tx, p.accountID)
	if err != nil {
		return risk.Account{}, err
	}
	a := p.account
	a.Profile = &profile.Profile
	// Exchange-triggered stops count toward the same discipline limit.
	now := time.Now()
	a, err = risk.LoadActivity(ctx, p.Tx, p.accountID, symbol, a, now)
	if err != nil {
		return a, err
	}
	if p.live {
		_, week := risk.UTCWindows(now)
		var stopped uint32
		err = p.QueryRow(ctx, `SELECT count(*) FROM app.positions WHERE account_id=$1::uuid AND status='closed' AND close_reason='exchange_stop' AND closed_at >= $2 AND closed_at <= $3`, p.accountID, week, now).Scan(&stopped)
		a.Activity.StoppedTradesThisWeek += stopped
	}
	return a, err
}

func (p *postgresTx) RecordCheck(ctx context.Context, body risk.Request, result *enginev1.CheckResponse, orderID string) error {
	body.AccountID = p.accountID
	return risk.InsertCheck(ctx, p.Tx, body, result, "order", &orderID)
}

func (p Postgres) PositionAccount(ctx context.Context, owner *string, id string) (string, string, error) {
	var accountID, userID string
	err := p.Pool.QueryRow(ctx, `SELECT a.id::text,a.user_id::text FROM app.positions p
		JOIN app.accounts a ON a.id=p.account_id WHERE p.id=$1::uuid AND ($2::uuid IS NULL OR a.user_id=$2::uuid)
		AND a.type IN ('paper','live') AND a.quote_asset='USDT'`, id, owner).Scan(&accountID, &userID)
	if errors.Is(err, pgx.ErrNoRows) {
		err = ErrNotFound
	}
	return accountID, userID, err
}

const positionColumns = `id::text,account_id::text,symbol,side,quantity::text,entry_price::text,stop::text,
	leverage,coin_class,initial_risk::text,status,opened_at,closed_at,exit_price::text,realized_pnl::text,close_reason,exchange_entry_order_id,exchange_stop_order_id::text,exchange_stop_status,exchange_stop_error`

func scanPosition(row pgx.Row) (Position, error) {
	var p Position
	var stopID *string
	var stopStatus, stopError *string
	err := row.Scan(&p.ID, &p.AccountID, &p.Symbol, &p.Side, &p.Quantity, &p.EntryPrice, &p.Stop,
		&p.Leverage, &p.CoinClass, &p.InitialRisk, &p.Status, &p.OpenedAt, &p.ClosedAt, &p.ExitPrice, &p.RealizedPNL, &p.CloseReason, &p.ExchangeEntryOrderID, &stopID, &stopStatus, &stopError)
	if err != nil {
		return p, err
	}
	if stopStatus != nil {
		p.ExchangeStop = &ExchangeStop{Status: *stopStatus, OrderID: stopID, Error: stopError}
	}
	p.OpenedAt = p.OpenedAt.UTC()
	if p.ClosedAt != nil {
		t := p.ClosedAt.UTC()
		p.ClosedAt = &t
	}
	for _, value := range []*string{&p.Quantity, &p.EntryPrice, &p.Stop, &p.InitialRisk, p.ExitPrice, p.RealizedPNL} {
		if value == nil {
			continue
		}
		d, err := decimal.NewFromString(*value)
		if err != nil {
			return Position{}, err
		}
		*value = d.String()
	}
	return p, nil
}

func (p *postgresTx) Position(ctx context.Context, id string) (Position, error) {
	position, err := scanPosition(p.QueryRow(ctx, `SELECT `+positionColumns+` FROM app.positions
		WHERE id=$1::uuid AND account_id=$2::uuid FOR UPDATE`, id, p.accountID))
	if errors.Is(err, pgx.ErrNoRows) {
		err = ErrNotFound
	}
	return position, err
}

func (p *postgresTx) OpeningFee(ctx context.Context, id string) (decimal.Decimal, error) {
	var fee string
	err := p.QueryRow(ctx, `SELECT COALESCE(sum(f.fee),0)::text FROM app.fills f
		JOIN app.orders o ON o.id=f.order_id WHERE o.position_id=$1::uuid AND o.account_id=$2::uuid
		AND o.purpose='open' AND o.status='filled'`, id, p.accountID).Scan(&fee)
	if err != nil {
		return decimal.Zero, err
	}
	return decimal.NewFromString(fee)
}

func (p *postgresTx) Reject(ctx context.Context, o Order) error {
	_, err := p.Exec(ctx, `INSERT INTO app.orders
		(id,account_id,symbol,side,kind,purpose,quantity,leverage,stop,status,position_id,created_at)
		VALUES($1::uuid,$2::uuid,$3,$4,$5,$6,$7::numeric,$8,$9::numeric,$10,$11::uuid,$12)`,
		o.ID, p.accountID, o.Symbol, o.Side, o.Kind, o.Purpose, o.Quantity, o.Leverage, o.Stop, o.Status, o.PositionID, o.CreatedAt)
	return err
}

func (p *postgresTx) execution(ctx context.Context, e Execution) error {
	if err := p.Reject(ctx, e.Order); err != nil {
		return err
	}
	f := e.Fill
	_, err := p.Exec(ctx, `INSERT INTO app.fills(id,order_id,price,quantity,fee,fee_asset,created_at)
		VALUES($1::uuid,$2::uuid,$3::numeric,$4::numeric,$5::numeric,$6,$7)`,
		f.ID, f.OrderID, f.Price, f.Quantity, f.Fee, f.FeeAsset, f.CreatedAt)
	return err
}

func (p *postgresTx) Open(ctx context.Context, e Execution) error {
	if err := p.execution(ctx, e); err != nil {
		return err
	}
	v := e.Position
	_, err := p.Exec(ctx, `INSERT INTO app.positions
		(id,account_id,symbol,side,quantity,entry_price,stop,leverage,coin_class,initial_risk,status,opened_at,exchange_entry_order_id,exchange_stop_order_id,exchange_stop_status,exchange_stop_error)
		VALUES($1::uuid,$2::uuid,$3,$4,$5::numeric,$6::numeric,$7::numeric,$8,$9,$10::numeric,$11,$12,$13,$14::bigint,$15,$16)`,
		v.ID, p.accountID, v.Symbol, v.Side, v.Quantity, v.EntryPrice, v.Stop, v.Leverage, v.CoinClass, v.InitialRisk, v.Status, v.OpenedAt, v.ExchangeEntryOrderID, stopOrderID(v), stopStatus(v), stopError(v))
	return err
}

func (p *postgresTx) Close(ctx context.Context, e Execution, pnl decimal.Decimal) error {
	v := e.Position
	tag, err := p.Exec(ctx, `UPDATE app.positions SET status='closed',closed_at=$3,exit_price=$4::numeric,
		realized_pnl=$5::numeric,close_reason=$6,exchange_stop_status=$7,exchange_stop_error=$8 WHERE id=$1::uuid AND account_id=$2::uuid AND status='open'`,
		v.ID, p.accountID, v.ClosedAt, v.ExitPrice, v.RealizedPNL, v.CloseReason, stopStatus(v), stopError(v))
	if err != nil {
		return err
	}
	if tag.RowsAffected() != 1 {
		return ErrClosed
	}
	if err := p.execution(ctx, e); err != nil {
		return err
	}
	_, err = p.Exec(ctx, `UPDATE app.accounts SET paper_balance=paper_balance+$2::numeric WHERE id=$1::uuid AND type='paper'`, p.accountID, pnl.String())
	return err
}

func (p Postgres) Positions(ctx context.Context, owner, accountID, state string) ([]Position, error) {
	// Check account existence independently so an owned empty account returns [].
	if _, _, err := loadTradingAccount(ctx, p.Pool, owner, accountID, false); err != nil {
		return nil, err
	}
	query := `SELECT ` + positionColumns + ` FROM app.positions WHERE account_id=$1::uuid AND status=$2`
	if state == "closed" {
		query += ` ORDER BY closed_at DESC,id DESC LIMIT 100`
	} else {
		query += ` ORDER BY opened_at DESC,id DESC`
	}
	rows, err := p.Pool.Query(ctx, query, accountID, state)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	positions := []Position{}
	for rows.Next() {
		position, err := scanPosition(rows)
		if err != nil {
			return nil, err
		}
		positions = append(positions, position)
	}
	return positions, rows.Err()
}

// LoadCapital uses the same locked account and transaction as order evaluation.
func LoadCapital(ctx context.Context, q risk.Querier, accountID string) ([]Capital, error) {
	rows, err := q.Query(ctx, `SELECT (p.quantity*p.entry_price)::text,COALESCE(p.leverage,1),
		COALESCE((SELECT sum(f.fee) FROM app.orders o JOIN app.fills f ON f.order_id=o.id
		WHERE o.position_id=p.id AND o.account_id=p.account_id AND o.purpose='open' AND o.status='filled'),0)::text
		FROM app.positions p WHERE p.account_id=$1::uuid AND p.status='open'`, accountID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var positions []Capital
	for rows.Next() {
		var p Capital
		var notional, fee string
		if err := rows.Scan(&notional, &p.Leverage, &fee); err != nil {
			return nil, err
		}
		p.Notional, err = decimal.NewFromString(notional)
		if err != nil {
			return nil, err
		}
		p.Fee, err = decimal.NewFromString(fee)
		if err != nil {
			return nil, err
		}
		positions = append(positions, p)
	}
	return positions, rows.Err()
}

package risk

import (
	"context"
	"errors"

	enginev1 "github.com/batuhangozegu/metron/api/internal/gen/metron/engine/v1"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/shopspring/decimal"
)

const profileColumns = `risk_per_trade_percent::text,target_risk_reward::text,total_open_risk_limit_percent::text,
	max_open_positions,max_losing_trades_today,daily_loss_limit_percent::text,max_stopped_trades_this_week,
	weekly_loss_limit_percent::text,updated_at`

func scanProfile(row pgx.Row) (SavedProfile, error) {
	var p SavedProfile
	err := row.Scan(&p.RiskPerTradePercent, &p.TargetRiskReward, &p.TotalOpenRiskLimitPercent,
		&p.MaxOpenPositions, &p.MaxLosingTradesToday, &p.DailyLossLimitPercent, &p.MaxStoppedTradesThisWeek,
		&p.WeeklyLossLimitPercent, &p.UpdatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return SavedProfile{Profile: DefaultProfile()}, nil
	}
	if err != nil {
		return p, err
	}
	for _, value := range []*string{&p.RiskPerTradePercent, &p.TargetRiskReward, &p.TotalOpenRiskLimitPercent, &p.DailyLossLimitPercent, &p.WeeklyLossLimitPercent} {
		d, err := decimal.NewFromString(*value)
		if err != nil {
			return SavedProfile{}, err
		}
		*value = d.String()
	}
	updated := p.UpdatedAt.UTC()
	p.UpdatedAt = &updated
	return p, nil
}

// LoadProfile can use the account-locked transaction used for order evaluation.
func LoadProfile(ctx context.Context, q Querier, accountID string) (SavedProfile, error) {
	return scanProfile(q.QueryRow(ctx, `SELECT `+profileColumns+` FROM app.risk_profiles WHERE account_id=$1::uuid`, accountID))
}

func ownedAccount(ctx context.Context, q Querier, owner, accountID string, lock bool) error {
	query := `SELECT id::text FROM app.accounts WHERE user_id=$1::uuid AND id=$2::uuid`
	if lock {
		query += ` FOR UPDATE`
	}
	var id string
	err := q.QueryRow(ctx, query, owner, accountID).Scan(&id)
	if errors.Is(err, pgx.ErrNoRows) {
		return ErrAccountNotFound
	}
	return err
}

func (p Postgres) Profile(ctx context.Context, owner, accountID string) (SavedProfile, error) {
	if err := ownedAccount(ctx, p.Pool, owner, accountID, false); err != nil {
		return SavedProfile{}, err
	}
	return LoadProfile(ctx, p.Pool, accountID)
}

func (p Postgres) SaveProfile(ctx context.Context, owner, accountID string, profile Profile) (SavedProfile, error) {
	if !profile.Valid(true) {
		return SavedProfile{}, errors.New("invalid profile")
	}
	tx, err := p.Pool.Begin(ctx)
	if err != nil {
		return SavedProfile{}, err
	}
	defer tx.Rollback(context.WithoutCancel(ctx))
	if err := ownedAccount(ctx, tx, owner, accountID, true); err != nil {
		return SavedProfile{}, err
	}
	value, err := scanProfile(tx.QueryRow(ctx, `INSERT INTO app.risk_profiles
		(account_id,risk_per_trade_percent,target_risk_reward,total_open_risk_limit_percent,max_open_positions,
		max_losing_trades_today,daily_loss_limit_percent,max_stopped_trades_this_week,weekly_loss_limit_percent)
		VALUES($1::uuid,$2::numeric,$3::numeric,$4::numeric,$5,$6,$7::numeric,$8,$9::numeric)
		ON CONFLICT(account_id) DO UPDATE SET risk_per_trade_percent=EXCLUDED.risk_per_trade_percent,
		target_risk_reward=EXCLUDED.target_risk_reward,total_open_risk_limit_percent=EXCLUDED.total_open_risk_limit_percent,
		max_open_positions=EXCLUDED.max_open_positions,max_losing_trades_today=EXCLUDED.max_losing_trades_today,
		daily_loss_limit_percent=EXCLUDED.daily_loss_limit_percent,max_stopped_trades_this_week=EXCLUDED.max_stopped_trades_this_week,
		weekly_loss_limit_percent=EXCLUDED.weekly_loss_limit_percent,updated_at=clock_timestamp() RETURNING `+profileColumns,
		accountID, profile.RiskPerTradePercent, profile.TargetRiskReward, profile.TotalOpenRiskLimitPercent,
		int64(*profile.MaxOpenPositions), int64(*profile.MaxLosingTradesToday), profile.DailyLossLimitPercent,
		int64(*profile.MaxStoppedTradesThisWeek), profile.WeeklyLossLimitPercent))
	if err != nil {
		return SavedProfile{}, err
	}
	return value, tx.Commit(ctx)
}

type Execer interface {
	Exec(context.Context, string, ...any) (pgconn.CommandTag, error)
}

// InsertCheck stores the complete engine response, preserving protobuf JSON names.
// Order callers insert the order first and pass the same transaction here.
func InsertCheck(ctx context.Context, q Execer, body Request, result *enginev1.CheckResponse, source string, orderID *string) error {
	encoded, err := MarshalCheck(result)
	if err != nil {
		return err
	}
	failed, caution := []string{}, []string{}
	for _, rule := range result.Checks {
		switch rule.GetStatus() {
		case "fail":
			failed = append(failed, rule.GetRuleId())
		case "caution":
			caution = append(caution, rule.GetRuleId())
		}
	}
	_, err = q.Exec(ctx, `INSERT INTO app.risk_checks
		(account_id,source,order_id,symbol,side,entry_price,allowed,verdict,failed_rules,caution_rules,profile_source,result)
		VALUES($1::uuid,$2,$3::uuid,$4,$5,$6::numeric,$7,$8,$9,$10,$11,$12::jsonb)`,
		body.AccountID, source, orderID, body.Symbol, body.Side, body.EntryPrice, result.Allowed, result.Verdict,
		failed, caution, ProfileSource(body), encoded)
	return err
}

func (p Postgres) RecordCheck(ctx context.Context, owner, accountID string, body Request, result *enginev1.CheckResponse) error {
	if err := ownedAccount(ctx, p.Pool, owner, accountID, false); err != nil {
		return err
	}
	body.AccountID = accountID
	return InsertCheck(ctx, p.Pool, body, result, "check", nil)
}

func (p Postgres) Checks(ctx context.Context, owner, accountID string, limit int) ([]CheckRecord, error) {
	if err := ownedAccount(ctx, p.Pool, owner, accountID, false); err != nil {
		return nil, err
	}
	rows, err := p.Pool.Query(ctx, `SELECT id::text,source,order_id::text,symbol,side,entry_price::text,
		allowed,verdict,failed_rules,caution_rules,profile_source,created_at FROM app.risk_checks
		WHERE account_id=$1::uuid ORDER BY created_at DESC,id DESC LIMIT $2`, accountID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	checks := []CheckRecord{}
	for rows.Next() {
		var c CheckRecord
		if err := rows.Scan(&c.ID, &c.Source, &c.OrderID, &c.Symbol, &c.Side, &c.EntryPrice, &c.Allowed,
			&c.Verdict, &c.FailedRules, &c.CautionRules, &c.ProfileSource, &c.CreatedAt); err != nil {
			return nil, err
		}
		price, err := decimal.NewFromString(c.EntryPrice)
		if err != nil {
			return nil, err
		}
		c.EntryPrice, c.CreatedAt = price.String(), c.CreatedAt.UTC()
		checks = append(checks, c)
	}
	return checks, rows.Err()
}

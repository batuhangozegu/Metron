-- Saved limits and the audit trail are owned by the Go API.
-- +goose Up
CREATE TABLE app.risk_profiles (
    account_id uuid PRIMARY KEY REFERENCES app.accounts(id) ON DELETE CASCADE,
    risk_per_trade_percent numeric NOT NULL CHECK (risk_per_trade_percent > 0 AND risk_per_trade_percent <= 100 AND scale(risk_per_trade_percent) <= 8),
    target_risk_reward numeric NOT NULL CHECK (target_risk_reward > 0 AND target_risk_reward <= 100 AND scale(target_risk_reward) <= 8),
    total_open_risk_limit_percent numeric NOT NULL CHECK (total_open_risk_limit_percent > 0 AND total_open_risk_limit_percent <= 100 AND scale(total_open_risk_limit_percent) <= 8),
    max_open_positions bigint NOT NULL CHECK (max_open_positions BETWEEN 1 AND 4294967295),
    max_losing_trades_today bigint NOT NULL CHECK (max_losing_trades_today BETWEEN 1 AND 4294967295),
    daily_loss_limit_percent numeric NOT NULL CHECK (daily_loss_limit_percent > 0 AND daily_loss_limit_percent <= 100 AND scale(daily_loss_limit_percent) <= 8),
    max_stopped_trades_this_week bigint NOT NULL CHECK (max_stopped_trades_this_week BETWEEN 1 AND 4294967295),
    weekly_loss_limit_percent numeric NOT NULL CHECK (weekly_loss_limit_percent > 0 AND weekly_loss_limit_percent <= 100 AND scale(weekly_loss_limit_percent) <= 8),
    updated_at timestamptz NOT NULL DEFAULT now()
);

CREATE TABLE app.risk_checks (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    account_id uuid NOT NULL REFERENCES app.accounts(id) ON DELETE CASCADE,
    source text NOT NULL CHECK (source IN ('check', 'order')),
    order_id uuid REFERENCES app.orders(id),
    symbol text NOT NULL,
    side text NOT NULL CHECK (side IN ('long', 'short')),
    entry_price numeric NOT NULL,
    allowed boolean NOT NULL,
    verdict text NOT NULL,
    failed_rules text[] NOT NULL,
    caution_rules text[] NOT NULL,
    profile_source text NOT NULL CHECK (profile_source IN ('stored', 'preview')),
    result jsonb NOT NULL,
    created_at timestamptz NOT NULL DEFAULT now()
);
CREATE INDEX risk_checks_account_created ON app.risk_checks(account_id, created_at DESC);

-- +goose Down
DROP TABLE app.risk_checks;
DROP TABLE app.risk_profiles;

-- Portfolio history is owned by the Go API.
-- +goose Up
ALTER TABLE app.accounts ADD COLUMN starting_balance numeric;
UPDATE app.accounts a SET starting_balance = a.paper_balance - COALESCE(
    (SELECT sum(p.realized_pnl) FROM app.positions p
     WHERE p.account_id=a.id AND p.status='closed'), 0)
WHERE a.type='paper';

CREATE TABLE app.equity_snapshots (
    account_id uuid NOT NULL REFERENCES app.accounts(id) ON DELETE CASCADE,
    taken_at timestamptz NOT NULL,
    balance numeric NOT NULL,
    unrealized_pnl numeric NOT NULL,
    equity numeric NOT NULL,
    PRIMARY KEY (account_id, taken_at)
);

-- +goose Down
DROP TABLE app.equity_snapshots;
ALTER TABLE app.accounts DROP COLUMN starting_balance;

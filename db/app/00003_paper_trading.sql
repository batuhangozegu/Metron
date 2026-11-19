-- Paper execution is owned by Go; the engine only reads open stops.
-- +goose Up
CREATE TABLE app.orders (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    account_id uuid NOT NULL REFERENCES app.accounts(id) ON DELETE CASCADE,
    symbol text NOT NULL,
    side text CHECK (side IN ('long', 'short')),
    kind text CHECK (kind IN ('market')),
    purpose text CHECK (purpose IN ('open', 'close')),
    quantity numeric NOT NULL CHECK (quantity > 0),
    leverage int,
    stop numeric,
    status text CHECK (status IN ('filled', 'rejected')),
    position_id uuid,
    created_at timestamptz NOT NULL DEFAULT now()
);

CREATE TABLE app.fills (
    id uuid PRIMARY KEY,
    order_id uuid NOT NULL REFERENCES app.orders(id) ON DELETE CASCADE,
    price numeric NOT NULL,
    quantity numeric NOT NULL,
    fee numeric NOT NULL,
    fee_asset text NOT NULL DEFAULT 'USDT',
    created_at timestamptz NOT NULL DEFAULT now()
);

CREATE TABLE app.positions (
    id uuid PRIMARY KEY,
    account_id uuid NOT NULL REFERENCES app.accounts(id) ON DELETE CASCADE,
    symbol text NOT NULL,
    side text CHECK (side IN ('long', 'short')),
    quantity numeric NOT NULL,
    entry_price numeric NOT NULL,
    stop numeric NOT NULL,
    leverage int,
    coin_class text NOT NULL,
    initial_risk numeric NOT NULL,
    status text CHECK (status IN ('open', 'closed')),
    opened_at timestamptz NOT NULL DEFAULT now(),
    closed_at timestamptz,
    exit_price numeric,
    realized_pnl numeric,
    close_reason text CHECK (close_reason IN ('manual', 'stop'))
);
CREATE INDEX positions_account_status ON app.positions(account_id, status);
CREATE INDEX positions_open ON app.positions(status) WHERE status = 'open';

-- Creation limits belong to the API. Settled PnL can have more than eight
-- decimal places, exceed the starting balance limit, or exhaust the account.
ALTER TABLE app.accounts DROP CONSTRAINT accounts_check;
ALTER TABLE app.accounts ADD CONSTRAINT accounts_check CHECK (
    (type = 'paper' AND paper_balance IS NOT NULL)
    OR (type = 'live' AND paper_balance IS NULL)
);

-- +goose Down
ALTER TABLE app.accounts DROP CONSTRAINT accounts_check;
ALTER TABLE app.accounts ADD CONSTRAINT accounts_check CHECK (
    (type = 'paper' AND paper_balance IS NOT NULL
        AND paper_balance > 0 AND paper_balance <= 10000000
        AND scale(paper_balance) <= 8)
    OR (type = 'live' AND paper_balance IS NULL)
);
DROP TABLE app.positions;
DROP TABLE app.fills;
DROP TABLE app.orders;

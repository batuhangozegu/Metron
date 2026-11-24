-- Live credentials and exchange-owned protective stops.
-- +goose Up
CREATE TABLE app.exchange_credentials (
    account_id uuid PRIMARY KEY REFERENCES app.accounts(id) ON DELETE CASCADE,
    environment text NOT NULL CHECK (environment IN ('testnet', 'mainnet')),
    key_hint text NOT NULL,
    api_key_enc bytea NOT NULL,
    api_secret_enc bytea NOT NULL,
    created_at timestamptz NOT NULL DEFAULT now()
);
ALTER TABLE app.positions
    ADD COLUMN exchange_entry_order_id bigint,
    ADD COLUMN exchange_stop_order_id bigint,
    ADD COLUMN exchange_stop_status text CHECK (exchange_stop_status IN ('placed','failed','filled','cancelled')),
    ADD COLUMN exchange_stop_error text;
ALTER TABLE app.positions DROP CONSTRAINT positions_close_reason_check;
ALTER TABLE app.positions ADD CONSTRAINT positions_close_reason_check CHECK (close_reason IN ('manual','stop','exchange_stop'));

-- +goose Down
ALTER TABLE app.positions DROP CONSTRAINT positions_close_reason_check;
UPDATE app.positions SET close_reason='stop' WHERE close_reason='exchange_stop';
ALTER TABLE app.positions ADD CONSTRAINT positions_close_reason_check CHECK (close_reason IN ('manual','stop'));
ALTER TABLE app.positions
    DROP COLUMN exchange_entry_order_id,
    DROP COLUMN exchange_stop_order_id,
    DROP COLUMN exchange_stop_status,
    DROP COLUMN exchange_stop_error;
DROP TABLE app.exchange_credentials;

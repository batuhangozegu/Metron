-- Account market selects the spot or USDT-M futures risk rule set.
-- +goose Up
ALTER TABLE app.accounts
    ADD COLUMN market text NOT NULL DEFAULT 'spot'
    CHECK (market IN ('spot', 'futures'));

-- +goose Down
ALTER TABLE app.accounts DROP COLUMN market;

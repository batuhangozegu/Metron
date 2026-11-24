-- +goose Up
CREATE TABLE app.price_alerts (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id uuid NOT NULL REFERENCES app.users(id) ON DELETE CASCADE,
    symbol text NOT NULL,
    direction text NOT NULL CHECK (direction IN ('above', 'below')),
    price numeric NOT NULL CHECK (price > 0 AND price < 'Infinity'::numeric AND scale(price) <= 8),
    note text CHECK (length(note) <= 200),
    status text NOT NULL DEFAULT 'active' CHECK (status IN ('active', 'triggered', 'cancelled')),
    created_at timestamptz NOT NULL DEFAULT now(),
    triggered_at timestamptz,
    triggered_price numeric,
    cancelled_at timestamptz
);
CREATE INDEX price_alerts_active_symbol ON app.price_alerts(symbol) WHERE status = 'active';
CREATE INDEX price_alerts_user_created ON app.price_alerts(user_id, created_at DESC, id DESC);

-- +goose Down
DROP TABLE app.price_alerts;

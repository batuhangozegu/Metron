-- Application data: only the Go API writes this schema.
-- +goose Up
CREATE SCHEMA app;

CREATE TABLE app.users (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    email text NOT NULL,
    password_hash text NOT NULL,
    totp_secret_enc bytea,
    totp_enabled boolean NOT NULL DEFAULT false,
    -- Highest accepted 30-second counter, consumed atomically to prevent replay.
    totp_last_step bigint,
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now(),
    CHECK (NOT totp_enabled OR totp_secret_enc IS NOT NULL)
);
CREATE UNIQUE INDEX users_email_lower ON app.users (lower(email));

CREATE TABLE app.sessions (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id uuid NOT NULL REFERENCES app.users(id) ON DELETE CASCADE,
    token_hash bytea NOT NULL UNIQUE CHECK (octet_length(token_hash) = 32),
    pending_2fa boolean NOT NULL DEFAULT false,
    created_at timestamptz NOT NULL DEFAULT now(),
    expires_at timestamptz NOT NULL,
    last_seen_at timestamptz NOT NULL DEFAULT now(),
    user_agent text NOT NULL,
    ip inet NOT NULL
);
CREATE INDEX sessions_user_id ON app.sessions(user_id);
CREATE INDEX sessions_expires_at ON app.sessions(expires_at);

CREATE TABLE app.accounts (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id uuid NOT NULL REFERENCES app.users(id) ON DELETE CASCADE,
    type text NOT NULL CHECK (type IN ('paper', 'live')),
    name text NOT NULL CHECK (length(trim(name)) BETWEEN 1 AND 100),
    quote_asset text NOT NULL DEFAULT 'USDT',
    paper_balance numeric,
    created_at timestamptz NOT NULL DEFAULT now(),
    CHECK ((type = 'paper' AND paper_balance IS NOT NULL
        AND paper_balance > 0 AND paper_balance <= 10000000
        AND scale(paper_balance) <= 8)
        OR (type = 'live' AND paper_balance IS NULL))
);
CREATE INDEX accounts_user_id ON app.accounts(user_id);

-- +goose Down
DROP SCHEMA app CASCADE;

-- Market data written by the Rust engine. Only the engine writes to this
-- schema; the Go API reads it. Prices are NUMERIC, times are UTC.

-- +goose Up
CREATE SCHEMA market;

-- Closed candles from Binance. Hourly and daily only: enough for ATR, and
-- older history can be fetched again when needed (ARCHITECTURE.md 5.5).
CREATE TABLE market.candles (
    symbol       text        NOT NULL,
    interval     text        NOT NULL CHECK (interval IN ('1h', '1d')),
    open_time    timestamptz NOT NULL,
    close_time   timestamptz NOT NULL,
    open         numeric     NOT NULL,
    high         numeric     NOT NULL,
    low          numeric     NOT NULL,
    close        numeric     NOT NULL,
    volume       numeric     NOT NULL, -- in the base asset
    quote_volume numeric     NOT NULL, -- in USDT
    trades       bigint      NOT NULL,
    PRIMARY KEY (symbol, interval, open_time)
);

-- One row per watched pair: its automatic coin class and latest statistics.
-- A user's manual class override lives in the app schema, not here.
CREATE TABLE market.symbols (
    symbol                 text        PRIMARY KEY,
    base_asset             text        NOT NULL,
    quote_asset            text        NOT NULL,
    coin_class             text        NOT NULL CHECK (coin_class IN ('core', 'large', 'small', 'stable')),
    -- Mean USDT volume of the last 30 closed daily candles; drives the class.
    avg_daily_quote_volume numeric,
    atr_period             integer     NOT NULL,
    -- Wilder ATR over closed candles, and as a percentage of the last close.
    atr_1d                 numeric,
    atr_1d_percent         numeric,
    atr_1h                 numeric,
    atr_1h_percent         numeric,
    updated_at             timestamptz NOT NULL
);

-- +goose Down
DROP SCHEMA market CASCADE;

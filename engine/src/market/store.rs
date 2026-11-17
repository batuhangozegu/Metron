//! Writes to the `market` schema. The engine is its only writer.

use chrono::{DateTime, Utc};
use rust_decimal::Decimal;
use sqlx::PgPool;

use super::classify::CoinClass;
use super::{Candle, Interval};

/// Inserts closed candles, replacing any row with the same key.
pub async fn upsert_candles(
    pool: &PgPool,
    symbol: &str,
    interval: Interval,
    candles: &[Candle],
) -> sqlx::Result<()> {
    if candles.is_empty() {
        return Ok(());
    }
    // One round trip: each column goes in as an array and UNNEST zips them.
    let col = |f: fn(&Candle) -> Decimal| candles.iter().map(f).collect::<Vec<_>>();
    let open_times: Vec<DateTime<Utc>> = candles.iter().map(|c| c.open_time).collect();
    let close_times: Vec<DateTime<Utc>> = candles.iter().map(|c| c.close_time).collect();
    let trades: Vec<i64> = candles.iter().map(|c| c.trades).collect();

    sqlx::query(
        "INSERT INTO market.candles
             (symbol, interval, open_time, close_time, open, high, low, close, volume, quote_volume, trades)
         SELECT $1, $2, * FROM UNNEST(
             $3::timestamptz[], $4::timestamptz[], $5::numeric[], $6::numeric[], $7::numeric[],
             $8::numeric[], $9::numeric[], $10::numeric[], $11::bigint[])
         ON CONFLICT (symbol, interval, open_time) DO UPDATE SET
             close_time = EXCLUDED.close_time, open = EXCLUDED.open, high = EXCLUDED.high,
             low = EXCLUDED.low, close = EXCLUDED.close, volume = EXCLUDED.volume,
             quote_volume = EXCLUDED.quote_volume, trades = EXCLUDED.trades",
    )
    .bind(symbol)
    .bind(interval.as_str())
    .bind(open_times)
    .bind(close_times)
    .bind(col(|c| c.open))
    .bind(col(|c| c.high))
    .bind(col(|c| c.low))
    .bind(col(|c| c.close))
    .bind(col(|c| c.volume))
    .bind(col(|c| c.quote_volume))
    .bind(trades)
    .execute(pool)
    .await?;
    Ok(())
}

/// A pair's latest class and statistics.
#[derive(Debug, Clone)]
pub struct SymbolStats {
    pub symbol: String,
    pub base_asset: String,
    pub quote_asset: String,
    pub coin_class: CoinClass,
    pub avg_daily_quote_volume: Option<Decimal>,
    pub atr_period: i32,
    pub atr_1d: Option<Decimal>,
    pub atr_1d_percent: Option<Decimal>,
    pub atr_1h: Option<Decimal>,
    pub atr_1h_percent: Option<Decimal>,
    pub updated_at: DateTime<Utc>,
}

pub async fn upsert_symbol(pool: &PgPool, s: &SymbolStats) -> sqlx::Result<()> {
    sqlx::query(
        "INSERT INTO market.symbols
             (symbol, base_asset, quote_asset, coin_class, avg_daily_quote_volume, atr_period,
              atr_1d, atr_1d_percent, atr_1h, atr_1h_percent, updated_at)
         VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
         ON CONFLICT (symbol) DO UPDATE SET
             base_asset = EXCLUDED.base_asset, quote_asset = EXCLUDED.quote_asset,
             coin_class = EXCLUDED.coin_class,
             avg_daily_quote_volume = EXCLUDED.avg_daily_quote_volume,
             atr_period = EXCLUDED.atr_period,
             atr_1d = EXCLUDED.atr_1d, atr_1d_percent = EXCLUDED.atr_1d_percent,
             atr_1h = EXCLUDED.atr_1h, atr_1h_percent = EXCLUDED.atr_1h_percent,
             updated_at = EXCLUDED.updated_at",
    )
    .bind(&s.symbol)
    .bind(&s.base_asset)
    .bind(&s.quote_asset)
    .bind(s.coin_class.as_str())
    .bind(s.avg_daily_quote_volume)
    .bind(s.atr_period)
    .bind(s.atr_1d)
    .bind(s.atr_1d_percent)
    .bind(s.atr_1h)
    .bind(s.atr_1h_percent)
    .bind(s.updated_at)
    .execute(pool)
    .await?;
    Ok(())
}

/// Removes pairs that are no longer on the watch list. Their candles stay.
pub async fn retain_symbols(pool: &PgPool, symbols: &[String]) -> sqlx::Result<u64> {
    let result = sqlx::query("DELETE FROM market.symbols WHERE symbol <> ALL($1)")
        .bind(symbols)
        .execute(pool)
        .await?;
    Ok(result.rows_affected())
}

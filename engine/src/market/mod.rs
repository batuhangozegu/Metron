//! Market data: candles and ATR into Postgres, live prices into Redis.
//!
//! Two loops run side by side:
//! - [`sync`] fetches hourly and daily candles over REST on a timer, stores
//!   the closed ones, and recomputes each pair's ATR and coin class.
//! - [`ticker`] keeps a WebSocket open to Binance's mini-ticker stream and
//!   writes every price update to Redis.

pub mod atr;
pub mod binance;
pub mod classify;
pub mod store;
pub mod sync;
pub mod ticker;

use std::time::Duration;

use anyhow::{Context, bail};
use chrono::{DateTime, Utc};
use rust_decimal::Decimal;

/// Only USDT spot pairs are in scope for the first release.
pub const QUOTE_ASSET: &str = "USDT";

/// Watched by default. Kept to well-known crypto assets on purpose: Binance
/// also lists tokenized stocks, fiat and gold tokens under USDT, and nothing
/// in its public API marks them apart.
const DEFAULT_SYMBOLS: &str = "BTC,ETH,SOL,XRP,BNB,DOGE,ADA,TRX,LINK,AVAX,SUI,NEAR,LTC,BCH,DOT,UNI,\
AAVE,INJ,TAO,ENA,HYPE,PEPE,WLD,OP,ARB,APT,TIA,PENDLE,SEI,FET,RENDER,ONDO,JUP,USDC,FDUSD";

/// Binance's public market-data hosts: no API key, no account region checks.
const DEFAULT_REST_URL: &str = "https://data-api.binance.vision";
const DEFAULT_WS_URL: &str = "wss://data-stream.binance.vision";

/// A pair is large-cap when its 30-day mean daily USDT volume reaches this.
const DEFAULT_LARGE_CAP_MIN_DAILY_VOLUME: i64 = 10_000_000;

const DEFAULT_SYNC_SECONDS: u64 = 15 * 60;

/// One closed candle. Prices are in the quote asset (USDT).
#[derive(Debug, Clone, PartialEq)]
pub struct Candle {
    pub open_time: DateTime<Utc>,
    pub close_time: DateTime<Utc>,
    pub open: Decimal,
    pub high: Decimal,
    pub low: Decimal,
    pub close: Decimal,
    pub volume: Decimal,
    pub quote_volume: Decimal,
    pub trades: i64,
}

/// The candle intervals the engine stores.
#[derive(Debug, Clone, Copy, PartialEq, Eq)]
pub enum Interval {
    Hour,
    Day,
}

impl Interval {
    pub fn as_str(self) -> &'static str {
        match self {
            Interval::Hour => "1h",
            Interval::Day => "1d",
        }
    }
}

/// Market data settings, read from the environment.
#[derive(Debug, Clone)]
pub struct Config {
    pub database_url: String,
    pub redis_url: String,
    pub rest_url: String,
    pub ws_url: String,
    /// Base assets, e.g. `BTC`; each is traded against [`QUOTE_ASSET`].
    pub base_assets: Vec<String>,
    pub sync_every: Duration,
    pub large_cap_min_daily_volume: Decimal,
}

impl Config {
    pub fn from_env() -> anyhow::Result<Self> {
        let var = |key: &str| std::env::var(key).ok().filter(|v| !v.trim().is_empty());

        let database_url = var("DATABASE_URL").context("DATABASE_URL is required")?;
        let base_assets = parse_assets(&var("MARKET_SYMBOLS").unwrap_or(DEFAULT_SYMBOLS.into()));
        if base_assets.is_empty() {
            bail!("MARKET_SYMBOLS lists no assets");
        }
        let sync_seconds = match var("MARKET_SYNC_SECONDS") {
            Some(v) => v
                .parse()
                .context("MARKET_SYNC_SECONDS must be whole seconds")?,
            None => DEFAULT_SYNC_SECONDS,
        };
        let large_cap_min_daily_volume = match var("LARGE_CAP_MIN_DAILY_VOLUME") {
            Some(v) => v
                .parse()
                .context("LARGE_CAP_MIN_DAILY_VOLUME must be a number")?,
            None => Decimal::from(DEFAULT_LARGE_CAP_MIN_DAILY_VOLUME),
        };

        Ok(Self {
            database_url,
            redis_url: var("REDIS_URL").unwrap_or("redis://127.0.0.1:6379".into()),
            rest_url: var("BINANCE_REST_URL").unwrap_or(DEFAULT_REST_URL.into()),
            ws_url: var("BINANCE_WS_URL").unwrap_or(DEFAULT_WS_URL.into()),
            base_assets,
            sync_every: Duration::from_secs(sync_seconds.max(60)),
            large_cap_min_daily_volume,
        })
    }

    /// Exchange symbols, e.g. `BTCUSDT`.
    pub fn symbols(&self) -> Vec<String> {
        self.base_assets
            .iter()
            .map(|base| format!("{base}{QUOTE_ASSET}"))
            .collect()
    }
}

/// Splits a comma-separated list, upper-cases it, and drops blanks and repeats.
fn parse_assets(list: &str) -> Vec<String> {
    let mut assets: Vec<String> = Vec::new();
    for asset in list.split(',').map(|a| a.trim().to_uppercase()) {
        if !asset.is_empty() && !assets.contains(&asset) {
            assets.push(asset);
        }
    }
    assets
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn parse_assets_normalises_the_list() {
        assert_eq!(parse_assets(" btc, ETH,,eth ,sol"), ["BTC", "ETH", "SOL"]);
    }

    #[test]
    fn default_list_is_well_formed() {
        let assets = parse_assets(DEFAULT_SYMBOLS);
        assert!(assets.len() > 30);
        assert!(
            assets
                .iter()
                .all(|a| a.chars().all(|c| c.is_ascii_alphanumeric()))
        );
    }
}

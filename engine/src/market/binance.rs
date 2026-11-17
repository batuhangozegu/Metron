//! Binance public market data: candles over REST, mini tickers over WebSocket.

use std::str::FromStr;
use std::time::Duration;

use anyhow::{Context, anyhow, bail};
use chrono::{DateTime, Utc};
use rust_decimal::Decimal;
use serde::{Deserialize, Serialize};

use super::{Candle, Interval};

/// REST client for the public market-data endpoints.
#[derive(Clone)]
pub struct Rest {
    http: reqwest::Client,
    base_url: String,
}

impl Rest {
    pub fn new(base_url: &str) -> anyhow::Result<Self> {
        let http = reqwest::Client::builder()
            .timeout(Duration::from_secs(15))
            .user_agent(concat!("metron-engine/", env!("CARGO_PKG_VERSION")))
            .build()?;
        Ok(Self {
            http,
            base_url: base_url.trim_end_matches('/').to_string(),
        })
    }

    /// The latest `limit` candles of `symbol`, oldest first. The last one is
    /// usually still open; [`closed_only`] drops it.
    pub async fn klines(
        &self,
        symbol: &str,
        interval: Interval,
        limit: u16,
    ) -> anyhow::Result<Vec<Candle>> {
        let response = self
            .http
            .get(format!("{}/api/v3/klines", self.base_url))
            .query(&[
                ("symbol", symbol),
                ("interval", interval.as_str()),
                ("limit", &limit.to_string()),
            ])
            .send()
            .await
            .with_context(|| format!("klines request for {symbol}"))?;

        let status = response.status();
        let body = response.text().await?;
        if !status.is_success() {
            bail!("klines for {symbol}: HTTP {status}: {body}");
        }
        parse_klines(&body)
    }
}

/// A kline row as Binance sends it: a 12-element array of mixed types.
type RawKline = (
    i64,    // open time, ms
    String, // open
    String, // high
    String, // low
    String, // close
    String, // volume (base asset)
    i64,    // close time, ms
    String, // quote asset volume
    i64,    // number of trades
    String, // taker buy base volume
    String, // taker buy quote volume
    String, // unused
);

pub fn parse_klines(body: &str) -> anyhow::Result<Vec<Candle>> {
    let rows: Vec<RawKline> = serde_json::from_str(body).context("parse klines")?;
    rows.into_iter()
        .map(|r| {
            Ok(Candle {
                open_time: millis(r.0)?,
                close_time: millis(r.6)?,
                open: decimal(&r.1)?,
                high: decimal(&r.2)?,
                low: decimal(&r.3)?,
                close: decimal(&r.4)?,
                volume: decimal(&r.5)?,
                quote_volume: decimal(&r.7)?,
                trades: r.8,
            })
        })
        .collect()
}

/// Keeps the candles that had closed by `now`.
pub fn closed_only(candles: Vec<Candle>, now: DateTime<Utc>) -> Vec<Candle> {
    candles.into_iter().filter(|c| c.close_time < now).collect()
}

/// One entry of the `<symbol>@miniTicker` stream: rolling 24-hour figures.
#[derive(Debug, Clone, PartialEq, Deserialize, Serialize)]
pub struct MiniTicker {
    #[serde(rename = "E")]
    pub event_time_ms: i64,
    #[serde(rename = "s")]
    pub symbol: String,
    #[serde(rename = "c", with = "rust_decimal::serde::str")]
    pub close: Decimal,
    #[serde(rename = "o", with = "rust_decimal::serde::str")]
    pub open: Decimal,
    #[serde(rename = "h", with = "rust_decimal::serde::str")]
    pub high: Decimal,
    #[serde(rename = "l", with = "rust_decimal::serde::str")]
    pub low: Decimal,
    #[serde(rename = "v", with = "rust_decimal::serde::str")]
    pub volume: Decimal,
    #[serde(rename = "q", with = "rust_decimal::serde::str")]
    pub quote_volume: Decimal,
}

/// The envelope of a combined-stream message.
#[derive(Debug, Deserialize)]
struct Combined<T> {
    data: T,
}

pub fn parse_mini_ticker(message: &str) -> anyhow::Result<MiniTicker> {
    let envelope: Combined<MiniTicker> =
        serde_json::from_str(message).context("parse mini ticker")?;
    Ok(envelope.data)
}

/// The combined-stream URL for the mini tickers of `symbols`.
pub fn mini_ticker_url(ws_base: &str, symbols: &[String]) -> String {
    let streams: Vec<String> = symbols
        .iter()
        .map(|s| format!("{}@miniTicker", s.to_lowercase()))
        .collect();
    format!(
        "{}/stream?streams={}",
        ws_base.trim_end_matches('/'),
        streams.join("/")
    )
}

fn decimal(s: &str) -> anyhow::Result<Decimal> {
    Decimal::from_str(s).with_context(|| format!("not a decimal: {s:?}"))
}

fn millis(ms: i64) -> anyhow::Result<DateTime<Utc>> {
    DateTime::from_timestamp_millis(ms).ok_or_else(|| anyhow!("timestamp out of range: {ms}"))
}

#[cfg(test)]
mod tests {
    use super::*;
    use rust_decimal::dec;

    // Two real daily BTCUSDT candles, as returned by /api/v3/klines.
    const KLINES: &str = r#"[[1794787200000,"65351.05000000","65696.16000000","63905.01000000","64882.87000000","21929.23589000",1794873599999,"1420769295.21917790",4401625,"11262.16992000","729754303.41452570","0"],[1794873600000,"64882.88000000","65447.00000000","63336.00000000","63876.33000000","22780.65472000",1794959999999,"1470940545.02942510",3784115,"12507.26691000","807555401.12725520","0"]]"#;

    #[test]
    fn parses_klines() {
        let candles = parse_klines(KLINES).unwrap();
        assert_eq!(candles.len(), 2);
        let first = &candles[0];
        assert_eq!(first.open_time.timestamp_millis(), 1794787200000);
        assert_eq!(first.close_time.timestamp_millis(), 1794873599999);
        assert_eq!(first.high, dec!(65696.16));
        assert_eq!(first.close, dec!(64882.87));
        assert_eq!(first.quote_volume, dec!(1420769295.2191779));
        assert_eq!(first.trades, 4401625);
    }

    #[test]
    fn drops_the_open_candle() {
        let candles = parse_klines(KLINES).unwrap();
        let now = DateTime::from_timestamp_millis(1794900000000).unwrap();
        let closed = closed_only(candles, now);
        assert_eq!(closed.len(), 1);
        assert_eq!(closed[0].open_time.timestamp_millis(), 1794787200000);
    }

    #[test]
    fn parses_a_combined_mini_ticker() {
        let message = r#"{"stream":"pendleusdt@miniTicker","data":{"e":"24hrMiniTicker","E":1794936345264,"s":"PENDLEUSDT","c":"1.41800000","o":"1.43500000","h":"1.46100000","l":"1.40200000","v":"669731.80000000","q":"961426.31250000"}}"#;
        let ticker = parse_mini_ticker(message).unwrap();
        assert_eq!(ticker.symbol, "PENDLEUSDT");
        assert_eq!(ticker.close, dec!(1.418));
        assert_eq!(ticker.open, dec!(1.435));
        assert_eq!(ticker.quote_volume, dec!(961426.3125));
        assert_eq!(ticker.event_time_ms, 1794936345264);
    }

    #[test]
    fn builds_the_stream_url() {
        let url = mini_ticker_url(
            "wss://data-stream.binance.vision/",
            &["BTCUSDT".into(), "ETHUSDT".into()],
        );
        assert_eq!(
            url,
            "wss://data-stream.binance.vision/stream?streams=btcusdt@miniTicker/ethusdt@miniTicker"
        );
    }
}

//! The live mini-ticker feed. Redis reconnects through its connection manager;
//! a closed Binance stream (including the 24-hour limit) is opened again.

use std::time::{Duration, Instant};

use anyhow::{Context, bail};
use futures_util::{SinkExt, StreamExt};
use redis::aio::ConnectionManager;
use tokio_tungstenite::tungstenite::Message;
use tracing::{info, warn};

use super::{Config, binance};

pub async fn run(config: Config, mut redis: ConnectionManager) {
    let url = binance::mini_ticker_url(&config.ws_url, &config.symbols());
    let mut backoff = Duration::from_secs(1);
    loop {
        let started = Instant::now();
        if let Err(error) = stream(&url, &mut redis).await {
            warn!(%error, "mini-ticker stream disconnected");
        }
        // A healthy stream resets the retry delay; repeated short failures
        // back off even when the WebSocket handshake succeeds.
        if started.elapsed() >= Duration::from_secs(60) {
            backoff = Duration::from_secs(1);
        }
        info!(
            retry_seconds = backoff.as_secs(),
            "reconnecting mini-ticker stream"
        );
        tokio::time::sleep(backoff).await;
        backoff = (backoff * 2).min(Duration::from_secs(60));
    }
}

async fn stream(url: &str, redis: &mut ConnectionManager) -> anyhow::Result<()> {
    let (mut socket, _) = tokio_tungstenite::connect_async(url)
        .await
        .context("connect Binance mini-ticker stream")?;
    info!("mini-ticker stream connected");
    while let Some(message) = socket.next().await {
        match message? {
            Message::Text(body) => {
                let ticker = match binance::parse_mini_ticker(&body) {
                    Ok(ticker) => ticker,
                    Err(error) => {
                        warn!(%error, "ignoring invalid mini ticker");
                        continue;
                    }
                };
                ticker_pipeline(&ticker)?.query_async::<()>(redis).await?;
            }
            Message::Ping(payload) => socket.send(Message::Pong(payload)).await?,
            Message::Close(_) => break,
            _ => {}
        }
    }
    bail!("Binance closed the mini-ticker stream")
}

fn ticker_pipeline(ticker: &binance::MiniTicker) -> anyhow::Result<redis::Pipeline> {
    let fields = [
        ("price", ticker.close.normalize().to_string()),
        ("open", ticker.open.normalize().to_string()),
        ("high", ticker.high.normalize().to_string()),
        ("low", ticker.low.normalize().to_string()),
        ("volume", ticker.volume.normalize().to_string()),
        ("quote_volume", ticker.quote_volume.normalize().to_string()),
        ("event_time_ms", ticker.event_time_ms.to_string()),
    ];
    let mut pipeline = redis::pipe();
    pipeline
        .cmd("HSET")
        .arg(format!("market:ticker:{}", ticker.symbol))
        .arg(&fields)
        .ignore()
        .publish("market:tickers", serde_json::to_string(ticker)?)
        .ignore();
    Ok(pipeline)
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn pipeline_caches_and_publishes_decimal_strings() {
        let ticker = binance::parse_mini_ticker(
            r#"{"data":{"E":1794936345264,"s":"BTCUSDT","c":"63880.94000000","o":"64000","h":"65000","l":"63000","v":"2.50000000","q":"159702.35000000"}}"#,
        )
        .unwrap();
        let pipeline = ticker_pipeline(&ticker).unwrap();
        let packed = String::from_utf8(pipeline.get_packed_pipeline()).unwrap();
        assert!(packed.contains("HSET\r\n"));
        assert!(packed.contains("market:ticker:BTCUSDT"));
        assert!(packed.contains("63880.94\r\n"));
        assert!(packed.contains("PUBLISH\r\n"));
        assert!(packed.contains("market:tickers"));
        let json = serde_json::to_value(ticker).unwrap();
        assert!(json["c"].is_string());
        assert_eq!(json["E"], 1794936345264_i64);
    }
}

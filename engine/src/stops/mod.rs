//! Read-only paper stop monitoring. Go owns execution and all app-schema writes.

use std::collections::{HashMap, HashSet};
use std::pin::Pin;
use std::time::{Duration, Instant};

use anyhow::Context;
use chrono::{DateTime, SecondsFormat, Utc};
use futures_util::{Stream, stream};
use redis::aio::ConnectionManager;
use rust_decimal::Decimal;
use sqlx::{PgPool, types::Uuid};
use tokio::sync::broadcast;
use tonic::Status;
use tracing::{debug, warn};

use crate::pb::StopHit;

const POLL_INTERVAL: Duration = Duration::from_secs(2);
const RESEND_AFTER: Duration = Duration::from_secs(30);
const STREAM_CAPACITY: usize = 256;

pub type HitStream = Pin<Box<dyn Stream<Item = Result<StopHit, Status>> + Send>>;

#[derive(Debug, Clone)]
pub struct Hub {
    sender: broadcast::Sender<StopHit>,
}

impl Default for Hub {
    fn default() -> Self {
        Self {
            sender: broadcast::channel(STREAM_CAPACITY).0,
        }
    }
}

impl Hub {
    pub fn subscribe(&self) -> HitStream {
        Box::pin(stream::unfold(
            Some(self.sender.subscribe()),
            |receiver| async move {
                let mut receiver = receiver?;
                match receiver.recv().await {
                    Ok(hit) => Some((Ok(hit), Some(receiver))),
                    Err(broadcast::error::RecvError::Closed) => None,
                    // Never silently skip a stop. Reconnecting consumers will get
                    // still-open hits again when the dedupe window expires.
                    Err(broadcast::error::RecvError::Lagged(_)) => Some((
                        Err(Status::unavailable(
                            "Stop stream fell behind; reconnect. / Stop akışı geride kaldı; yeniden bağlanın.",
                        )),
                        None,
                    )),
                }
            },
        ))
    }
}

#[derive(Debug)]
struct Position {
    id: String,
    symbol: String,
    side: String,
    stop: Decimal,
}

fn crossed(side: &str, stop: Decimal, price: Decimal) -> bool {
    match side {
        "long" => price <= stop,
        "short" => price >= stop,
        _ => false,
    }
}

#[derive(Default)]
struct Watcher {
    emitted: HashMap<String, Instant>,
}

impl Watcher {
    fn scan(
        &mut self,
        positions: &[Position],
        prices: &HashMap<String, Decimal>,
        hub: &Hub,
        now: Instant,
        observed_at: DateTime<Utc>,
    ) {
        let open: HashSet<_> = positions.iter().map(|position| &position.id).collect();
        self.emitted
            .retain(|id, at| open.contains(id) && now.duration_since(*at) < RESEND_AFTER);
        for position in positions {
            let Some(&price) = prices.get(&position.symbol) else {
                continue;
            };
            if price <= Decimal::ZERO
                || position.stop <= Decimal::ZERO
                || !crossed(&position.side, position.stop, price)
                || self.emitted.contains_key(&position.id)
            {
                continue;
            }
            let hit = StopHit {
                position_id: position.id.clone(),
                symbol: position.symbol.clone(),
                stop: position.stop.normalize().to_string(),
                price: price.normalize().to_string(),
                observed_at: observed_at.to_rfc3339_opts(SecondsFormat::Millis, true),
            };
            // An absent subscriber must not suppress delivery to the next one.
            if hub.sender.send(hit).is_ok() {
                self.emitted.insert(position.id.clone(), now);
            }
        }
    }
}

async fn open_positions(pool: &PgPool) -> sqlx::Result<Vec<Position>> {
    let rows: Vec<(Uuid, String, String, Decimal)> =
        sqlx::query_as("SELECT id, symbol, side, stop FROM app.positions WHERE status = 'open'")
            .fetch_all(pool)
            .await?;
    Ok(rows
        .into_iter()
        .map(|(id, symbol, side, stop)| Position {
            id: id.to_string(),
            symbol,
            side,
            stop,
        })
        .collect())
}

async fn latest_prices(
    redis: &mut ConnectionManager,
    positions: &[Position],
) -> anyhow::Result<HashMap<String, Decimal>> {
    let symbols: HashSet<_> = positions.iter().map(|position| &position.symbol).collect();
    if symbols.is_empty() {
        return Ok(HashMap::new());
    }
    let mut pipeline = redis::pipe();
    for symbol in &symbols {
        pipeline.hget(format!("market:ticker:{symbol}"), "price");
    }
    let values: Vec<Option<String>> = pipeline
        .query_async(redis)
        .await
        .context("read latest paper stop prices")?;
    let mut prices = HashMap::new();
    for (symbol, value) in symbols.into_iter().zip(values) {
        let Some(value) = value else { continue };
        match Decimal::from_str_exact(&value) {
            Ok(price) if price > Decimal::ZERO => {
                prices.insert(symbol.clone(), price);
            }
            _ => warn!(%symbol, "ignoring invalid paper stop ticker price"),
        }
    }
    Ok(prices)
}

pub async fn run(pool: PgPool, mut redis: ConnectionManager, hub: Hub) {
    let mut watcher = Watcher::default();
    let mut missing_table_logged = false;
    let mut interval = tokio::time::interval(POLL_INTERVAL);
    interval.set_missed_tick_behavior(tokio::time::MissedTickBehavior::Skip);
    loop {
        interval.tick().await;
        let positions = match open_positions(&pool).await {
            Ok(positions) => positions,
            Err(error)
                if error.as_database_error().and_then(|error| error.code())
                    == Some("42P01".into()) =>
            {
                if !missing_table_logged {
                    debug!("app.positions is not available yet; paper stops are empty");
                    missing_table_logged = true;
                }
                Vec::new()
            }
            Err(error) => {
                warn!(%error, "could not load open paper positions");
                continue;
            }
        };
        match latest_prices(&mut redis, &positions).await {
            Ok(prices) => watcher.scan(&positions, &prices, &hub, Instant::now(), Utc::now()),
            Err(error) => warn!(%error, "could not check paper stops"),
        }
    }
}

#[cfg(test)]
mod grpc_tests;
#[cfg(test)]
mod tests;

//! The candle loop: fetch, store, recompute ATR and class, sleep, repeat.

use chrono::Utc;
use rust_decimal::Decimal;
use sqlx::PgPool;
use tracing::{info, warn};

use super::binance::{Rest, closed_only};
use super::classify::classify;
use super::store::{self, SymbolStats};
use super::{Candle, Config, Interval, QUOTE_ASSET, atr};

/// Daily candles fetched per sync: the class needs 30, ATR settles over ~60.
const DAILY_LIMIT: u16 = 61;
/// Hourly candles fetched per sync: one week.
const HOURLY_LIMIT: u16 = 169;
/// Days averaged for the volume that decides the coin class.
const VOLUME_DAYS: usize = 30;

pub async fn run(config: Config, rest: Rest, pool: PgPool) {
    let symbols = config.symbols();
    let mut timer = tokio::time::interval(config.sync_every);
    timer.set_missed_tick_behavior(tokio::time::MissedTickBehavior::Delay);
    loop {
        timer.tick().await;
        match store::retain_symbols(&pool, &symbols).await {
            Ok(0) => {}
            Ok(n) => info!(removed = n, "dropped pairs no longer watched"),
            Err(e) => warn!(error = %e, "could not prune the symbol list"),
        }

        let mut synced = 0;
        for base in &config.base_assets {
            match sync_one(&config, &rest, &pool, base).await {
                Ok(()) => synced += 1,
                Err(e) => warn!(base, error = format!("{e:#}"), "candle sync failed"),
            }
        }
        info!(synced, total = config.base_assets.len(), "candle sync done");
    }
}

async fn sync_one(config: &Config, rest: &Rest, pool: &PgPool, base: &str) -> anyhow::Result<()> {
    let symbol = format!("{base}{QUOTE_ASSET}");
    let now = Utc::now();

    let daily = closed_only(rest.klines(&symbol, Interval::Day, DAILY_LIMIT).await?, now);
    let hourly = closed_only(
        rest.klines(&symbol, Interval::Hour, HOURLY_LIMIT).await?,
        now,
    );
    store::upsert_candles(pool, &symbol, Interval::Day, &daily).await?;
    store::upsert_candles(pool, &symbol, Interval::Hour, &hourly).await?;

    let avg_daily_quote_volume = average_quote_volume(&daily, VOLUME_DAYS);
    let (atr_1d, atr_1d_percent) = atr_with_percent(&daily);
    let (atr_1h, atr_1h_percent) = atr_with_percent(&hourly);

    store::upsert_symbol(
        pool,
        &SymbolStats {
            symbol,
            base_asset: base.to_string(),
            quote_asset: QUOTE_ASSET.to_string(),
            coin_class: classify(
                base,
                avg_daily_quote_volume,
                config.large_cap_min_daily_volume,
            ),
            avg_daily_quote_volume,
            atr_period: atr::PERIOD as i32,
            atr_1d,
            atr_1d_percent,
            atr_1h,
            atr_1h_percent,
            updated_at: now,
        },
    )
    .await?;
    Ok(())
}

fn atr_with_percent(candles: &[Candle]) -> (Option<Decimal>, Option<Decimal>) {
    let Some(value) = atr::wilder(candles, atr::PERIOD) else {
        return (None, None);
    };
    let percent = candles
        .last()
        .and_then(|c| atr::percent_of_close(value, c.close));
    (Some(value.round_dp(8)), percent)
}

/// Mean quote volume of the last `days` candles; `None` until there are that many.
fn average_quote_volume(daily: &[Candle], days: usize) -> Option<Decimal> {
    if days == 0 || daily.len() < days {
        return None;
    }
    let total: Decimal = daily[daily.len() - days..]
        .iter()
        .map(|c| c.quote_volume)
        .sum();
    Some((total / Decimal::from(days)).round_dp(2))
}

#[cfg(test)]
mod tests {
    use super::*;
    use chrono::TimeZone;
    use rust_decimal::dec;

    fn day(quote_volume: Decimal) -> Candle {
        let t = Utc.timestamp_opt(0, 0).unwrap();
        Candle {
            open_time: t,
            close_time: t,
            open: dec!(1),
            high: dec!(1),
            low: dec!(1),
            close: dec!(1),
            volume: Decimal::ZERO,
            quote_volume,
            trades: 0,
        }
    }

    #[test]
    fn averages_only_the_most_recent_days() {
        let mut candles = vec![day(dec!(1000)); 5];
        candles.extend(vec![day(dec!(30)); 3]);
        assert_eq!(average_quote_volume(&candles, 3), Some(dec!(30)));
    }

    #[test]
    fn no_average_without_enough_history() {
        assert_eq!(average_quote_volume(&[day(dec!(1))], 30), None);
    }
}

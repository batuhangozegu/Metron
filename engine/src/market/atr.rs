//! Average True Range: how far a pair moves per candle, on average.

use rust_decimal::Decimal;

use super::Candle;

/// ATR look-back used everywhere for now: 14 candles, Wilder's default.
pub const PERIOD: usize = 14;

/// The largest of the candle's range and its gaps from the previous close.
pub fn true_range(candle: &Candle, previous_close: Decimal) -> Decimal {
    let range = candle.high - candle.low;
    let gap_up = (candle.high - previous_close).abs();
    let gap_down = (candle.low - previous_close).abs();
    range.max(gap_up).max(gap_down)
}

/// Wilder's ATR over `candles` (oldest first).
///
/// The first value is the mean of the first `period` true ranges; each later
/// true range is folded in as `(atr * (period - 1) + tr) / period`. Needs at
/// least `period + 1` candles, because every true range needs a previous
/// close. Older candles beyond that make the value settle, so callers pass
/// a few times `period`.
pub fn wilder(candles: &[Candle], period: usize) -> Option<Decimal> {
    if period == 0 || candles.len() < period + 1 {
        return None;
    }
    let n = Decimal::from(period);
    let mut ranges = candles.windows(2).map(|w| true_range(&w[1], w[0].close));

    let mut atr = ranges.by_ref().take(period).sum::<Decimal>() / n;
    for tr in ranges {
        atr = (atr * (n - Decimal::ONE) + tr) / n;
    }
    Some(atr)
}

/// ATR as a percentage of the last close, e.g. `2.41` for 2.41%.
pub fn percent_of_close(atr: Decimal, close: Decimal) -> Option<Decimal> {
    if close.is_zero() {
        return None;
    }
    Some((atr / close * Decimal::ONE_HUNDRED).round_dp(4))
}

#[cfg(test)]
mod tests {
    use super::*;
    use chrono::{TimeZone, Utc};
    use rust_decimal::dec;

    fn candle(high: Decimal, low: Decimal, close: Decimal) -> Candle {
        let t = Utc.timestamp_opt(0, 0).unwrap();
        Candle {
            open_time: t,
            close_time: t,
            open: close,
            high,
            low,
            close,
            volume: Decimal::ZERO,
            quote_volume: Decimal::ZERO,
            trades: 0,
        }
    }

    #[test]
    fn true_range_takes_the_widest_measure() {
        // Inside range: high - low.
        assert_eq!(
            true_range(&candle(dec!(110), dec!(100), dec!(105)), dec!(105)),
            dec!(10)
        );
        // Gap up: the move from the previous close dominates.
        assert_eq!(
            true_range(&candle(dec!(105), dec!(102), dec!(104)), dec!(100)),
            dec!(5)
        );
        // Gap down.
        assert_eq!(
            true_range(&candle(dec!(95), dec!(94), dec!(94)), dec!(100)),
            dec!(6)
        );
    }

    #[test]
    fn constant_range_gives_that_range() {
        let candles: Vec<_> = (0..30)
            .map(|_| candle(dec!(105), dec!(95), dec!(100)))
            .collect();
        assert_eq!(wilder(&candles, PERIOD), Some(dec!(10)));
    }

    #[test]
    fn wilder_smoothing_by_hand() {
        // True ranges: 2, 4, 6. Seed (2 + 4) / 2 = 3, then (3 * 1 + 6) / 2 = 4.5.
        let candles = [
            candle(dec!(100), dec!(100), dec!(100)),
            candle(dec!(101), dec!(99), dec!(100)),
            candle(dec!(102), dec!(98), dec!(100)),
            candle(dec!(103), dec!(97), dec!(100)),
        ];
        assert_eq!(wilder(&candles, 2), Some(dec!(4.5)));
    }

    #[test]
    fn too_few_candles_gives_none() {
        let candles: Vec<_> = (0..PERIOD)
            .map(|_| candle(dec!(2), dec!(1), dec!(1)))
            .collect();
        assert_eq!(wilder(&candles, PERIOD), None);
        assert_eq!(wilder(&candles, 0), None);
    }

    #[test]
    fn percent_of_close_rounds_to_four_places() {
        assert_eq!(
            percent_of_close(dec!(1540), dec!(63880)),
            Some(dec!(2.4108))
        );
        assert_eq!(percent_of_close(dec!(1), Decimal::ZERO), None);
    }
}

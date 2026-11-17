//! Automatic coin classes (ARCHITECTURE.md 5.5): by trading volume, with
//! BTC/ETH and stablecoins named outright. Users can override a class later;
//! that override lives in the app schema.

use rust_decimal::Decimal;

/// A pair's risk class. Each class later gets its own ATR multiplier, maximum
/// position size and risk per trade.
#[derive(Debug, Clone, Copy, PartialEq, Eq)]
pub enum CoinClass {
    /// BTC and ETH.
    Core,
    Large,
    Small,
    Stable,
}

impl CoinClass {
    pub fn as_str(self) -> &'static str {
        match self {
            CoinClass::Core => "core",
            CoinClass::Large => "large",
            CoinClass::Small => "small",
            CoinClass::Stable => "stable",
        }
    }
}

const CORE: &[&str] = &["BTC", "ETH"];

/// Dollar-pegged tokens traded against USDT on Binance.
const STABLECOINS: &[&str] = &[
    "USDC", "FDUSD", "TUSD", "USDP", "DAI", "USD1", "RLUSD", "USDE", "PYUSD", "XUSD", "BFUSD",
];

/// Classifies a base asset. Without volume history (a newly listed pair, or a
/// failed fetch) a non-core, non-stable asset is small-cap: the stricter class.
pub fn classify(
    base_asset: &str,
    avg_daily_quote_volume: Option<Decimal>,
    large_cap_min: Decimal,
) -> CoinClass {
    if CORE.contains(&base_asset) {
        CoinClass::Core
    } else if STABLECOINS.contains(&base_asset) {
        CoinClass::Stable
    } else if avg_daily_quote_volume.is_some_and(|v| v >= large_cap_min) {
        CoinClass::Large
    } else {
        CoinClass::Small
    }
}

#[cfg(test)]
mod tests {
    use super::*;
    use rust_decimal::dec;

    const MIN: Decimal = dec!(10_000_000);

    #[test]
    fn named_classes_ignore_volume() {
        assert_eq!(classify("BTC", Some(Decimal::ZERO), MIN), CoinClass::Core);
        assert_eq!(classify("ETH", None, MIN), CoinClass::Core);
        assert_eq!(
            classify("USDC", Some(dec!(1_000_000_000)), MIN),
            CoinClass::Stable
        );
    }

    #[test]
    fn volume_splits_large_and_small() {
        assert_eq!(
            classify("SOL", Some(dec!(229_000_000)), MIN),
            CoinClass::Large
        );
        assert_eq!(classify("LINK", Some(MIN), MIN), CoinClass::Large);
        assert_eq!(
            classify("PENDLE", Some(dec!(9_999_999.99)), MIN),
            CoinClass::Small
        );
    }

    #[test]
    fn unknown_volume_is_small_cap() {
        assert_eq!(classify("NEW", None, MIN), CoinClass::Small);
    }
}

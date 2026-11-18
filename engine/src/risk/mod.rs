//! Pure decimal risk calculation. Transport and market lookup live in `grpc`.
//! See docs/RISK.md for defaults, limits and rounding.

pub mod grpc;

#[cfg(test)]
mod tests;

#[cfg(test)]
mod grpc_tests;

use std::collections::BTreeMap;
use std::fmt;

use rust_decimal::{Decimal, RoundingStrategy, dec};

use crate::market::classify::CoinClass;

pub const QUANTITY_STEP: Decimal = dec!(0.00000001);
pub const MAINTENANCE_MARGIN_RATE: Decimal = dec!(0.005);
pub const LIQUIDATION_BUFFER_RATE: Decimal = dec!(0.01);

#[derive(Debug, Clone, Copy, PartialEq, Eq)]
pub enum AccountMarket {
    Spot,
    Futures,
}

impl AccountMarket {
    pub fn as_str(self) -> &'static str {
        match self {
            Self::Spot => "spot",
            Self::Futures => "futures",
        }
    }
}

#[derive(Debug, Clone, Default)]
pub struct Activity {
    pub open_risk_percent: Decimal,
    pub open_positions: u32,
    pub losing_trades_today: u32,
    pub daily_loss_percent: Decimal,
    pub stopped_trades_this_week: u32,
    pub weekly_loss_percent: Decimal,
}

#[derive(Debug, Clone, Copy, PartialEq, Eq)]
pub enum Side {
    Long,
    Short,
}

impl Side {
    pub fn as_str(self) -> &'static str {
        match self {
            Self::Long => "long",
            Self::Short => "short",
        }
    }
}

#[derive(Debug, Clone)]
pub struct Profile {
    pub risk_percent: Decimal,
    pub target_risk_reward: Decimal,
    pub total_open_risk_limit_percent: Decimal,
    pub max_open_positions: u32,
    pub max_losing_trades_today: u32,
    pub daily_loss_limit_percent: Decimal,
    pub max_stopped_trades_this_week: u32,
    pub weekly_loss_limit_percent: Decimal,
}

impl Default for Profile {
    fn default() -> Self {
        Self {
            risk_percent: dec!(2),
            target_risk_reward: dec!(2),
            total_open_risk_limit_percent: dec!(6),
            max_open_positions: 5,
            max_losing_trades_today: 3,
            daily_loss_limit_percent: dec!(5),
            max_stopped_trades_this_week: 5,
            weekly_loss_limit_percent: dec!(10),
        }
    }
}

#[derive(Debug, Clone)]
pub struct Input {
    pub account_market: AccountMarket,
    pub leverage: Option<u32>,
    pub current_class_holding: Decimal,
    pub activity: Activity,
    pub symbol: String,
    pub side: Side,
    pub entry: Decimal,
    pub manual_stop: Option<Decimal>,
    pub equity: Decimal,
    pub profile: Profile,
}

#[derive(Debug, Clone, Copy)]
pub struct Market {
    pub class: CoinClass,
    pub atr: Decimal,
    pub quantity_step: Decimal,
}

#[derive(Debug, Clone, Copy, PartialEq, Eq)]
pub struct ClassRules {
    pub max_risk_percent: Decimal,
    pub atr_multiplier: Decimal,
    pub max_position_percent: Decimal,
}

pub fn class_rules(class: CoinClass) -> Result<ClassRules, Error> {
    let (risk, multiplier, position) = match class {
        CoinClass::Core => (dec!(2), dec!(1.5), dec!(40)),
        CoinClass::Large => (dec!(1), dec!(1.5), dec!(15)),
        CoinClass::Small => (dec!(0.5), dec!(2), dec!(8)),
        CoinClass::Stable => return Err(Error::Stablecoin),
    };
    Ok(ClassRules {
        max_risk_percent: risk,
        atr_multiplier: multiplier,
        max_position_percent: position,
    })
}

#[derive(Debug, Clone, PartialEq, Eq)]
pub enum Error {
    InvalidSymbol,
    InvalidDecimal(&'static str),
    NonPositive(&'static str),
    OutOfRange(&'static str),
    WrongSide,
    ZeroDistance,
    Stablecoin,
    Arithmetic,
    InvalidMarket,
    InvalidLeverage,
}

impl Error {
    pub fn code(&self) -> &'static str {
        match self {
            Self::InvalidSymbol => "invalid_symbol",
            Self::InvalidDecimal(_) => "invalid_decimal",
            Self::NonPositive(_) => "non_positive",
            Self::OutOfRange(_) => "out_of_range",
            Self::WrongSide => "wrong_stop_side",
            Self::ZeroDistance => "zero_stop_distance",
            Self::Stablecoin => "stablecoin_unsupported",
            Self::Arithmetic => "arithmetic_range",
            Self::InvalidMarket => "invalid_market",
            Self::InvalidLeverage => "invalid_leverage",
        }
    }
}

impl fmt::Display for Error {
    fn fmt(&self, f: &mut fmt::Formatter<'_>) -> fmt::Result {
        write!(f, "{}", self.code())
    }
}

impl std::error::Error for Error {}

pub fn validate_symbol(symbol: &str) -> Result<(), Error> {
    if !(5..=30).contains(&symbol.len())
        || !symbol.ends_with("USDT")
        || !symbol
            .bytes()
            .all(|c| c.is_ascii_uppercase() || c.is_ascii_digit())
    {
        return Err(Error::InvalidSymbol);
    }
    Ok(())
}

/// Plain decimal strings only. No exponent, sign, whitespace or silent rescaling.
pub fn parse_decimal(value: &str, field: &'static str) -> Result<Decimal, Error> {
    let value = parse_nonnegative_decimal(value, field)?;
    positive(value, field)?;
    Ok(value)
}

pub fn parse_nonnegative_decimal(value: &str, field: &'static str) -> Result<Decimal, Error> {
    let mut parts = value.split('.');
    let whole = parts.next().unwrap_or_default();
    let fraction = parts.next();
    if whole.is_empty()
        || whole.len() > 13
        || !whole.bytes().all(|c| c.is_ascii_digit())
        || fraction
            .is_some_and(|p| p.is_empty() || p.len() > 8 || !p.bytes().all(|c| c.is_ascii_digit()))
        || parts.next().is_some()
    {
        return Err(Error::InvalidDecimal(field));
    }
    let value = Decimal::from_str_exact(value).map_err(|_| Error::InvalidDecimal(field))?;
    nonnegative(value, field)?;
    Ok(value)
}

fn nonnegative(value: Decimal, field: &'static str) -> Result<(), Error> {
    if value < Decimal::ZERO || value > dec!(1_000_000_000_000) || value.scale() > 8 {
        return Err(Error::OutOfRange(field));
    }
    Ok(())
}

fn positive(value: Decimal, field: &'static str) -> Result<(), Error> {
    if value <= Decimal::ZERO {
        return Err(Error::NonPositive(field));
    }
    if value > dec!(1_000_000_000_000) || value.scale() > 8 {
        return Err(Error::OutOfRange(field));
    }
    Ok(())
}

#[derive(Debug, Clone, Copy, PartialEq, Eq)]
pub enum Status {
    Pass,
    Caution,
    Fail,
}

impl Status {
    pub fn as_str(self) -> &'static str {
        match self {
            Self::Pass => "pass",
            Self::Caution => "caution",
            Self::Fail => "fail",
        }
    }
}

#[derive(Debug, Clone)]
pub struct Check {
    pub rule_id: &'static str,
    pub status: Status,
    pub values: BTreeMap<String, String>,
    pub message_en: &'static str,
    pub message_tr: &'static str,
}

fn check(
    rule_id: &'static str,
    status: Status,
    values: &[(&str, Decimal)],
    en: &'static str,
    tr: &'static str,
) -> Check {
    Check {
        rule_id,
        status,
        values: values
            .iter()
            .map(|(k, v)| ((*k).to_owned(), v.normalize().to_string()))
            .collect(),
        message_en: en,
        message_tr: tr,
    }
}

#[derive(Debug)]
pub struct Output {
    pub leverage: u32,
    pub liquidation_price: Option<Decimal>,
    pub initial_margin: Option<Decimal>,
    pub rules: ClassRules,
    pub atr: Decimal,
    pub suggested_stop: Decimal,
    pub stop: Decimal,
    pub distance: Decimal,
    pub stop_percent: Decimal,
    pub risk_percent: Decimal,
    pub risk_amount: Decimal,
    pub position_size: Decimal,
    pub notional: Decimal,
    pub actual_risk: Decimal,
    pub position_percent: Decimal,
    pub target_price: Decimal,
    pub checks: Vec<Check>,
}

fn mul(a: Decimal, b: Decimal) -> Result<Decimal, Error> {
    let (a, b) = (a.normalize(), b.normalize());
    let scale = a.scale() + b.scale();
    let mantissa = a
        .mantissa()
        .checked_mul(b.mantissa())
        .ok_or(Error::Arithmetic)?;
    from_parts(mantissa, scale)
}

fn from_parts(mut mantissa: i128, mut scale: u32) -> Result<Decimal, Error> {
    while scale > 0 && mantissa % 10 == 0 {
        mantissa /= 10;
        scale -= 1;
    }
    if scale > 28 {
        return Err(Error::Arithmetic);
    }
    Decimal::try_from_i128_with_scale(mantissa, scale).map_err(|_| Error::Arithmetic)
}

fn add(a: Decimal, b: Decimal) -> Result<Decimal, Error> {
    let scale = a.scale().max(b.scale());
    let align = |v: Decimal| v.mantissa().checked_mul(10_i128.pow(scale - v.scale()));
    let sum = align(a)
        .ok_or(Error::Arithmetic)?
        .checked_add(align(b).ok_or(Error::Arithmetic)?)
        .ok_or(Error::Arithmetic)?;
    from_parts(sum, scale)
}
fn div(a: Decimal, b: Decimal) -> Result<Decimal, Error> {
    a.checked_div(b).ok_or(Error::Arithmetic)
}

/// V1 formula, floored to the quantity step. Remove the remainder *before*
/// division, avoiding a rounded quotient crossing a step boundary.
pub fn position_size(budget: Decimal, distance: Decimal, step: Decimal) -> Result<Decimal, Error> {
    if budget <= Decimal::ZERO || distance <= Decimal::ZERO || step <= Decimal::ZERO {
        return Err(Error::NonPositive("sizing"));
    }
    let unit_risk = mul(distance, step)?;
    let remainder = budget.checked_rem(unit_risk).ok_or(Error::Arithmetic)?;
    let units = div(
        budget.checked_sub(remainder).ok_or(Error::Arithmetic)?,
        unit_risk,
    )?;
    let size = mul(units, step)?;
    // Refuse an inexact intermediate rather than increasing the risk budget.
    if units.fract() != Decimal::ZERO || mul(size, distance)? > budget {
        return Err(Error::Arithmetic);
    }
    Ok(size)
}

pub fn calculate(input: &Input, market: Market) -> Result<Output, Error> {
    validate_symbol(&input.symbol)?;
    positive(input.entry, "entry_price")?;
    positive(input.equity, "account_equity")?;
    positive(input.profile.risk_percent, "risk_per_trade_percent")?;
    positive(input.profile.target_risk_reward, "target_risk_reward")?;
    if input.profile.risk_percent > dec!(100) || input.profile.target_risk_reward > dec!(100) {
        return Err(Error::OutOfRange("risk_profile"));
    }
    let leverage = match (input.account_market, input.leverage) {
        (AccountMarket::Spot, None | Some(1)) => 1,
        (AccountMarket::Futures, Some(value @ 1..=125)) => value,
        _ => return Err(Error::InvalidLeverage),
    };
    nonnegative(input.current_class_holding, "current_class_holding")?;
    for value in [
        input.activity.open_risk_percent,
        input.activity.daily_loss_percent,
        input.activity.weekly_loss_percent,
    ] {
        nonnegative(value, "activity")?;
    }
    for value in [
        input.profile.total_open_risk_limit_percent,
        input.profile.daily_loss_limit_percent,
        input.profile.weekly_loss_limit_percent,
    ] {
        positive(value, "discipline_profile")?;
        if value > dec!(100) {
            return Err(Error::OutOfRange("discipline_profile"));
        }
    }
    if input.profile.max_open_positions == 0
        || input.profile.max_losing_trades_today == 0
        || input.profile.max_stopped_trades_this_week == 0
    {
        return Err(Error::OutOfRange("discipline_profile"));
    }
    // Wilder's stored ATR can have 28 fractional digits. Quantize it upward
    // explicitly so a decimal product cannot silently narrow the stop.
    if market.atr <= Decimal::ZERO {
        return Err(Error::NonPositive("atr"));
    }
    if market.atr > dec!(1_000_000_000_000) {
        return Err(Error::OutOfRange("atr"));
    }
    let atr = market
        .atr
        .round_dp_with_strategy(8, RoundingStrategy::AwayFromZero);
    positive(market.quantity_step, "quantity_step")?;
    let rules = class_rules(market.class)?;
    let atr_distance = mul(atr, rules.atr_multiplier)?;
    let suggested_stop = add(
        input.entry,
        match input.side {
            Side::Long => -atr_distance,
            Side::Short => atr_distance,
        },
    )?;
    if suggested_stop <= Decimal::ZERO {
        return Err(Error::NonPositive("suggested_stop"));
    }
    let stop = input.manual_stop.unwrap_or(suggested_stop);
    if stop <= Decimal::ZERO {
        return Err(Error::NonPositive("manual_stop"));
    }
    if let Some(stop) = input.manual_stop {
        positive(stop, "manual_stop")?;
    }
    if stop == input.entry {
        return Err(Error::ZeroDistance);
    }
    if (input.side == Side::Long && stop > input.entry)
        || (input.side == Side::Short && stop < input.entry)
    {
        return Err(Error::WrongSide);
    }
    let distance = add(input.entry, -stop)?.abs();
    let risk_percent = input.profile.risk_percent.min(rules.max_risk_percent);
    let risk_amount = div(mul(input.equity, risk_percent)?, dec!(100))?;
    let position_size = position_size(risk_amount, distance, market.quantity_step)?;
    let notional = mul(position_size, input.entry)?;
    let actual_risk = mul(position_size, distance)?;
    let position_percent = mul(div(notional, input.equity)?, dec!(100))?;
    let stop_percent = mul(div(distance, input.entry)?, dec!(100))?;
    let target_distance = mul(distance, input.profile.target_risk_reward)?;
    let target_price = add(
        input.entry,
        match input.side {
            Side::Long => target_distance,
            Side::Short => -target_distance,
        },
    )?;
    let mut checks = vec![
        check(
            "class_risk",
            if input.profile.risk_percent > rules.max_risk_percent {
                Status::Caution
            } else {
                Status::Pass
            },
            &[
                ("requested_percent", input.profile.risk_percent),
                ("max_percent", rules.max_risk_percent),
                ("effective_percent", risk_percent),
            ],
            "Risk is capped by the coin class.",
            "Risk, coin sınıfının üst sınırıyla sınırlandırılır.",
        ),
        check(
            "atr_stop",
            if distance < atr_distance {
                Status::Caution
            } else {
                Status::Pass
            },
            &[
                ("distance", distance),
                ("suggested_distance", atr_distance),
                ("atr", atr),
                ("multiplier", rules.atr_multiplier),
            ],
            "Compare the stop distance with the ATR suggestion.",
            "Stop mesafesini ATR önerisiyle karşılaştırın.",
        ),
        check(
            "quantity_step",
            if position_size.is_zero() {
                Status::Fail
            } else {
                Status::Pass
            },
            &[
                ("step", market.quantity_step),
                ("size", position_size),
                ("budget", risk_amount),
                ("actual_risk", actual_risk),
            ],
            "Size is rounded down to the quantity step.",
            "Miktar, işlem adımına aşağı yuvarlanır.",
        ),
        check(
            "target_price",
            if target_price <= Decimal::ZERO {
                Status::Fail
            } else {
                Status::Pass
            },
            &[
                ("price", target_price),
                ("risk_reward", input.profile.target_risk_reward),
            ],
            "The target price must be positive.",
            "Hedef fiyat pozitif olmalıdır.",
        ),
    ];
    let (liquidation_price, initial_margin) = if input.account_market == AccountMarket::Futures {
        let margin = div(notional, Decimal::from(leverage))?;
        // Isolated collateral = entry notional / leverage; maintenance = mark
        // notional * rate. Compare cross-products so division rounding cannot
        // accidentally permit a stop on the liquidation boundary.
        let l = Decimal::from(leverage);
        let (numerator, denominator) = match input.side {
            Side::Long => (
                mul(input.entry, add(l, -Decimal::ONE)?)?,
                mul(l, add(Decimal::ONE, -MAINTENANCE_MARGIN_RATE)?)?,
            ),
            Side::Short => (
                mul(input.entry, add(l, Decimal::ONE)?)?,
                mul(l, add(Decimal::ONE, MAINTENANCE_MARGIN_RATE)?)?,
            ),
        };
        let liquidation = div(numerator, denominator)?;
        let stop_scaled = mul(stop, denominator)?;
        let gap = match input.side {
            Side::Long => add(stop_scaled, -numerator)?,
            Side::Short => add(numerator, -stop_scaled)?,
        };
        checks.push(check(
            "liquidation_stop",
            if gap <= Decimal::ZERO { Status::Fail } else if gap <= mul(mul(input.entry, LIQUIDATION_BUFFER_RATE)?, denominator)? { Status::Caution } else { Status::Pass },
            &[("stop", stop), ("liquidation_price", liquidation), ("leverage", l), ("maintenance_margin_rate", MAINTENANCE_MARGIN_RATE), ("buffer_rate", LIQUIDATION_BUFFER_RATE)],
            "The stop must trigger before estimated isolated-margin liquidation; allow a price buffer.",
            "Stop, tahmini izole marjin tasfiyesinden önce tetiklenmelidir; fiyat payı bırakın.",
        ));
        // Compare notional with equity * leverage, avoiding a rounded margin.
        checks.push(check(
            "initial_margin",
            if notional > mul(input.equity, l)? {
                Status::Fail
            } else {
                Status::Pass
            },
            &[
                ("initial_margin", margin),
                ("equity", input.equity),
                ("leverage", l),
            ],
            "Required initial margin must fit account equity.",
            "Gerekli başlangıç marjini hesap öz sermayesini aşmamalıdır.",
        ));
        (Some(liquidation), Some(margin))
    } else {
        let class_notional = add(notional, input.current_class_holding)?;
        checks.push(check(
            "position_limit",
            if class_notional > div(mul(input.equity, rules.max_position_percent)?, dec!(100))? {
                Status::Fail
            } else {
                Status::Pass
            },
            &[
                ("notional", notional),
                ("current_class_holding", input.current_class_holding),
                ("total_class_notional", class_notional),
                ("equity", input.equity),
                ("position_percent", position_percent),
                ("max_percent", rules.max_position_percent),
            ],
            "Existing class holdings plus the proposed position must fit the spot allocation limit.",
            "Mevcut sınıf tutarı ve önerilen pozisyon spot dağılım sınırına uymalıdır.",
        ));
        checks.push(check(
            "spot_side",
            if input.side == Side::Short {
                Status::Fail
            } else {
                Status::Pass
            },
            &[],
            "Spot trading supports long orders only.",
            "Spot işlemler yalnızca uzun pozisyonları destekler.",
        ));
        (None, None)
    };
    checks.extend(discipline_checks(input, risk_percent)?);
    Ok(Output {
        leverage,
        liquidation_price,
        initial_margin,
        rules,
        atr,
        suggested_stop,
        stop,
        distance,
        stop_percent,
        risk_percent,
        risk_amount,
        position_size,
        notional,
        actual_risk,
        position_percent,
        target_price,
        checks,
    })
}

fn discipline_checks(input: &Input, new_risk: Decimal) -> Result<Vec<Check>, Error> {
    let a = &input.activity;
    let p = &input.profile;
    let mut checks = Vec::new();
    for (id, current, proposed, limit, inclusive, en, tr) in [
        (
            "total_open_risk",
            a.open_risk_percent,
            add(a.open_risk_percent, new_risk)?,
            p.total_open_risk_limit_percent,
            false,
            "Total open risk includes the proposed trade.",
            "Toplam açık risk önerilen işlemi içerir.",
        ),
        (
            "daily_loss",
            a.daily_loss_percent,
            a.daily_loss_percent,
            p.daily_loss_limit_percent,
            true,
            "Reaching the daily loss cap blocks new trades until 00:00 UTC.",
            "Günlük zarar sınırına ulaşılması UTC 00:00'a kadar yeni işlemleri engeller.",
        ),
        (
            "weekly_loss",
            a.weekly_loss_percent,
            a.weekly_loss_percent,
            p.weekly_loss_limit_percent,
            true,
            "Reaching the weekly loss cap blocks new trades until Monday 00:00 UTC.",
            "Haftalık zarar sınırına ulaşılması pazartesi UTC 00:00'a kadar yeni işlemleri engeller.",
        ),
    ] {
        let status = if proposed > limit || (inclusive && proposed == limit) {
            Status::Fail
        } else if proposed >= mul(limit, dec!(0.75))? {
            Status::Caution
        } else {
            Status::Pass
        };
        checks.push(check(
            id,
            status,
            &[
                ("current_percent", current),
                ("proposed_percent", proposed),
                ("limit_percent", limit),
                (
                    "new_trade_risk_percent",
                    if id == "total_open_risk" {
                        new_risk
                    } else {
                        Decimal::ZERO
                    },
                ),
            ],
            en,
            tr,
        ));
    }
    for (id, current, proposed, limit, inclusive, en, tr) in [
        (
            "open_positions",
            a.open_positions,
            u64::from(a.open_positions) + 1,
            p.max_open_positions,
            false,
            "The proposed trade adds one open position.",
            "Önerilen işlem bir açık pozisyon ekler.",
        ),
        (
            "losing_trades_today",
            a.losing_trades_today,
            u64::from(a.losing_trades_today),
            p.max_losing_trades_today,
            true,
            "Reaching the daily losing-trade limit blocks new trades until 00:00 UTC.",
            "Günlük zararlı işlem sınırına ulaşılması UTC 00:00'a kadar yeni işlemleri engeller.",
        ),
        (
            "stopped_trades_this_week",
            a.stopped_trades_this_week,
            u64::from(a.stopped_trades_this_week),
            p.max_stopped_trades_this_week,
            true,
            "Reaching the weekly stopped-trade limit blocks new trades until Monday 00:00 UTC.",
            "Haftalık stop olan işlem sınırına ulaşılması pazartesi UTC 00:00'a kadar yeni işlemleri engeller.",
        ),
    ] {
        let limit64 = u64::from(limit);
        let status = if proposed > limit64 || (inclusive && proposed == limit64) {
            Status::Fail
        } else if proposed >= limit64 - 1 {
            Status::Caution
        } else {
            Status::Pass
        };
        checks.push(check(
            id,
            status,
            &[
                ("current", Decimal::from(current)),
                ("proposed", Decimal::from(proposed)),
                ("limit", Decimal::from(limit)),
            ],
            en,
            tr,
        ));
    }
    Ok(checks)
}

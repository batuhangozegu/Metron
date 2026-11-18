//! HTTP-independent transport adapter: reads the engine's latest market snapshot.

use std::sync::Arc;

use chrono::{DateTime, Utc};
use rust_decimal::Decimal;
use sqlx::PgPool;
use tonic::{Request, Response, Status};

use super::{
    AccountMarket, Activity, Input, Market, Profile, QUANTITY_STEP, Side, calculate, parse_decimal,
    parse_nonnegative_decimal, validate_symbol,
};
use crate::market::classify::CoinClass;
use crate::pb::{self, risk_service_server::RiskService};

#[derive(Debug, Clone)]
pub struct Snapshot {
    pub class: CoinClass,
    pub atr: Option<Decimal>,
    pub updated_at: DateTime<Utc>,
}

#[tonic::async_trait]
pub trait MarketSource: Send + Sync {
    async fn snapshot(&self, symbol: &str) -> Result<Option<Snapshot>, Status>;
}

#[tonic::async_trait]
impl MarketSource for PgPool {
    async fn snapshot(&self, symbol: &str) -> Result<Option<Snapshot>, Status> {
        let row: Option<(String, Option<Decimal>, DateTime<Utc>)> = sqlx::query_as(
            "SELECT coin_class, atr_1d, updated_at FROM market.symbols WHERE symbol=$1 AND quote_asset='USDT'",
        ).bind(symbol).fetch_optional(self).await.map_err(|_| Status::unavailable("market_unavailable"))?;
        row.map(|(class, atr, updated_at)| {
            let class = match class.as_str() {
                "core" => CoinClass::Core,
                "large" => CoinClass::Large,
                "small" => CoinClass::Small,
                "stable" => CoinClass::Stable,
                _ => return Err(Status::unavailable("market_unavailable")),
            };
            Ok(Snapshot {
                class,
                atr,
                updated_at,
            })
        })
        .transpose()
    }
}

pub struct Service {
    pub source: Arc<dyn MarketSource>,
}

fn decimal(value: Decimal) -> String {
    value.normalize().to_string()
}

fn invalid(error: super::Error) -> Status {
    Status::invalid_argument(error.code())
}

fn input(request: pb::CheckRequest) -> Result<Input, Status> {
    validate_symbol(&request.symbol).map_err(invalid)?;
    let side = match request.side.as_str() {
        "long" => Side::Long,
        "short" => Side::Short,
        _ => return Err(Status::invalid_argument("invalid_side")),
    };
    let profile = request.risk_profile.unwrap_or_default();
    let defaults = Profile::default();
    let activity = request.activity.unwrap_or_default();
    let optional_decimal = |value: &str, default, field| {
        if value.is_empty() {
            Ok(default)
        } else {
            parse_decimal(value, field).map_err(invalid)
        }
    };
    let activity_decimal = |value: &str, field| {
        if value.is_empty() {
            Ok(Decimal::ZERO)
        } else {
            parse_nonnegative_decimal(value, field).map_err(invalid)
        }
    };
    Ok(Input {
        account_market: match request.market.as_str() {
            "" | "spot" => AccountMarket::Spot,
            "futures" => AccountMarket::Futures,
            _ => return Err(invalid(super::Error::InvalidMarket)),
        },
        leverage: request.leverage,
        current_class_holding: activity_decimal(
            &request.current_class_holding,
            "current_class_holding",
        )?,
        activity: Activity {
            open_risk_percent: activity_decimal(&activity.open_risk_percent, "open_risk_percent")?,
            open_positions: activity.open_positions,
            losing_trades_today: activity.losing_trades_today,
            daily_loss_percent: activity_decimal(
                &activity.daily_loss_percent,
                "daily_loss_percent",
            )?,
            stopped_trades_this_week: activity.stopped_trades_this_week,
            weekly_loss_percent: activity_decimal(
                &activity.weekly_loss_percent,
                "weekly_loss_percent",
            )?,
        },
        symbol: request.symbol,
        side,
        entry: parse_decimal(&request.entry_price, "entry_price").map_err(invalid)?,
        manual_stop: request
            .manual_stop
            .as_deref()
            .map(|s| parse_decimal(s, "manual_stop"))
            .transpose()
            .map_err(invalid)?,
        equity: parse_decimal(&request.account_equity, "account_equity").map_err(invalid)?,
        profile: Profile {
            total_open_risk_limit_percent: optional_decimal(
                &profile.total_open_risk_limit_percent,
                defaults.total_open_risk_limit_percent,
                "total_open_risk_limit_percent",
            )?,
            max_open_positions: profile
                .max_open_positions
                .unwrap_or(defaults.max_open_positions),
            max_losing_trades_today: profile
                .max_losing_trades_today
                .unwrap_or(defaults.max_losing_trades_today),
            daily_loss_limit_percent: optional_decimal(
                &profile.daily_loss_limit_percent,
                defaults.daily_loss_limit_percent,
                "daily_loss_limit_percent",
            )?,
            max_stopped_trades_this_week: profile
                .max_stopped_trades_this_week
                .unwrap_or(defaults.max_stopped_trades_this_week),
            weekly_loss_limit_percent: optional_decimal(
                &profile.weekly_loss_limit_percent,
                defaults.weekly_loss_limit_percent,
                "weekly_loss_limit_percent",
            )?,
            risk_percent: if profile.risk_per_trade_percent.is_empty() {
                defaults.risk_percent
            } else {
                parse_decimal(&profile.risk_per_trade_percent, "risk_per_trade_percent")
                    .map_err(invalid)?
            },
            target_risk_reward: if profile.target_risk_reward.is_empty() {
                defaults.target_risk_reward
            } else {
                parse_decimal(&profile.target_risk_reward, "target_risk_reward").map_err(invalid)?
            },
        },
    })
}

#[tonic::async_trait]
impl RiskService for Service {
    async fn check(
        &self,
        request: Request<pb::CheckRequest>,
    ) -> Result<Response<pb::CheckResponse>, Status> {
        let input = input(request.into_inner())?;
        let snapshot = self
            .source
            .snapshot(&input.symbol)
            .await?
            .ok_or_else(|| Status::not_found("symbol_not_found"))?;
        // Sync defaults to 15 minutes. Refuse stale/future observations rather
        // than silently sizing against an old or misdated volatility estimate.
        let age = Utc::now().signed_duration_since(snapshot.updated_at);
        if age > chrono::Duration::hours(1) || age < chrono::Duration::minutes(-1) {
            return Err(Status::unavailable("market_stale"));
        }
        if snapshot.class == CoinClass::Stable {
            return Err(invalid(super::Error::Stablecoin));
        }
        let atr = snapshot
            .atr
            .filter(|v| *v > Decimal::ZERO)
            .ok_or_else(|| Status::unavailable("atr_unavailable"))?;
        let output = calculate(
            &input,
            Market {
                class: snapshot.class,
                atr,
                quantity_step: QUANTITY_STEP,
            },
        )
        .map_err(invalid)?;
        let passed = output
            .checks
            .iter()
            .filter(|c| c.status == super::Status::Pass)
            .count() as u32;
        let cautions = output
            .checks
            .iter()
            .filter(|c| c.status == super::Status::Caution)
            .count() as u32;
        let failed = output
            .checks
            .iter()
            .filter(|c| c.status == super::Status::Fail)
            .count() as u32;
        let verdict = if failed > 0 {
            "fail"
        } else if cautions > 0 {
            "caution"
        } else {
            "pass"
        };
        Ok(Response::new(pb::CheckResponse {
            market: input.account_market.as_str().into(),
            leverage: output.leverage,
            liquidation_price: output.liquidation_price.map(decimal),
            initial_margin: output.initial_margin.map(decimal),
            symbol: input.symbol,
            side: input.side.as_str().into(),
            coin_class: snapshot.class.as_str().into(),
            entry_price: decimal(input.entry),
            account_equity: decimal(input.equity),
            atr: decimal(output.atr),
            source_atr: decimal(atr),
            atr_multiplier: decimal(output.rules.atr_multiplier),
            suggested_stop: decimal(output.suggested_stop),
            stop: decimal(output.stop),
            stop_distance: decimal(output.distance),
            stop_percent: decimal(output.stop_percent),
            requested_risk_percent: decimal(input.profile.risk_percent),
            risk_percent: decimal(output.risk_percent),
            max_risk_percent: decimal(output.rules.max_risk_percent),
            risk_amount: decimal(output.risk_amount),
            quantity_step: decimal(QUANTITY_STEP),
            position_size: decimal(output.position_size),
            notional: decimal(output.notional),
            actual_risk: decimal(output.actual_risk),
            position_percent: decimal(output.position_percent),
            max_position_percent: if input.account_market == AccountMarket::Spot {
                decimal(output.rules.max_position_percent)
            } else {
                String::new()
            },
            target_risk_reward: decimal(input.profile.target_risk_reward),
            target_price: decimal(output.target_price),
            checks: output
                .checks
                .into_iter()
                .map(|c| pb::RuleCheck {
                    rule_id: c.rule_id.into(),
                    status: c.status.as_str().into(),
                    values: c.values.into_iter().collect(),
                    message_en: c.message_en.into(),
                    message_tr: c.message_tr.into(),
                })
                .collect(),
            verdict: verdict.into(),
            allowed: failed == 0,
            passed,
            cautions,
            failed,
            market_updated_at: snapshot.updated_at.to_rfc3339(),
        }))
    }
}

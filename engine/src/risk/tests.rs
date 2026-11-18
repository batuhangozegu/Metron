use super::*;

fn input(side: Side, entry: Decimal, stop: Option<Decimal>) -> Input {
    Input {
        account_market: AccountMarket::Spot,
        leverage: None,
        current_class_holding: Decimal::ZERO,
        activity: Activity::default(),
        symbol: "BTCUSDT".into(),
        side,
        entry,
        manual_stop: stop,
        equity: dec!(10000),
        profile: Profile::default(),
    }
}

fn market(class: CoinClass) -> Market {
    Market {
        class,
        atr: dec!(10),
        quantity_step: QUANTITY_STEP,
    }
}

#[test]
fn v1_formula_hand_checked_for_both_sides() {
    // Equity, risk %, entry, long stop, short stop, expected budget, size.
    for (equity, pct, entry, long, short, budget, size) in [
        (
            dec!(10000),
            dec!(2),
            dec!(100),
            dec!(90),
            dec!(110),
            dec!(200),
            dec!(20),
        ),
        (
            dec!(25000),
            dec!(0.5),
            dec!(5.2),
            dec!(4.576),
            dec!(5.824),
            dec!(125),
            dec!(200.32051282),
        ),
        (
            dec!(1000),
            dec!(1),
            dec!(50),
            dec!(47),
            dec!(53),
            dec!(10),
            dec!(3.33333333),
        ),
        (
            dec!(12345.67),
            dec!(0.75),
            dec!(200),
            dec!(187.5),
            dec!(212.5),
            dec!(92.592525),
            dec!(7.407402),
        ),
    ] {
        for (side, stop) in [(Side::Long, long), (Side::Short, short)] {
            let mut input = input(side, entry, Some(stop));
            input.equity = equity;
            input.profile.risk_percent = pct;
            let mut m = market(CoinClass::Core);
            m.atr = dec!(0.1);
            let output = calculate(&input, m).unwrap();
            assert_eq!(output.risk_amount, budget);
            assert_eq!(output.position_size, size);
            assert_eq!(output.notional, size * entry);
            assert_eq!(output.actual_risk, size * (entry - stop).abs());
            assert!(output.actual_risk <= budget);
            assert!(output.actual_risk + QUANTITY_STEP * (entry - stop).abs() > budget);
        }
    }
}

#[test]
fn floors_to_arbitrary_step_and_never_crosses_boundary() {
    assert_eq!(position_size(dec!(10), dec!(3), dec!(0.05)), Ok(dec!(3.3)));
    assert_eq!(
        position_size(dec!(0.99999999), dec!(1), dec!(1)),
        Ok(dec!(0))
    );
    assert_eq!(position_size(dec!(1), dec!(1), dec!(1)), Ok(dec!(1)));
    assert_eq!(
        position_size(dec!(1.00000001), dec!(1), dec!(1)),
        Ok(dec!(1))
    );
    // A normal decimal division can round this quotient to 1 at scale 28.
    assert_eq!(
        position_size(dec!(1), dec!(1.0000000000000000000000000001), dec!(1)),
        Ok(dec!(0))
    );
    for cents in 1..500 {
        let budget = Decimal::new(cents, 2);
        for distance in [dec!(0.12345678), dec!(3), dec!(12345.67)] {
            let step = dec!(0.005);
            let size = position_size(budget, distance, step).unwrap();
            assert_eq!(size % step, Decimal::ZERO);
            assert!(size * distance <= budget);
            assert!((size + step) * distance > budget);
        }
    }
}

#[test]
fn class_defaults_caps_and_atr_stops() {
    for (class, cap, k, max) in [
        (CoinClass::Core, dec!(2), dec!(1.5), dec!(40)),
        (CoinClass::Large, dec!(1), dec!(1.5), dec!(15)),
        (CoinClass::Small, dec!(0.5), dec!(2), dec!(8)),
    ] {
        for side in [Side::Long, Side::Short] {
            let result = calculate(&input(side, dec!(100), None), market(class)).unwrap();
            assert_eq!(result.risk_percent, cap);
            assert_eq!(result.rules.atr_multiplier, k);
            assert_eq!(result.rules.max_position_percent, max);
            assert_eq!(
                result.suggested_stop,
                if side == Side::Long {
                    dec!(100) - dec!(10) * k
                } else {
                    dec!(100) + dec!(10) * k
                }
            );
            assert_eq!(result.stop, result.suggested_stop);
        }
    }
    assert_eq!(class_rules(CoinClass::Stable), Err(Error::Stablecoin));
}

#[test]
fn returns_rule_values_and_bilingual_copy() {
    let mut order = input(Side::Long, dec!(100), Some(dec!(99)));
    order.profile.risk_percent = dec!(2);
    let output = calculate(&order, market(CoinClass::Small)).unwrap();
    assert_eq!(output.checks[0].status, Status::Caution);
    assert_eq!(output.checks[0].values["effective_percent"], "0.5");
    assert_eq!(output.checks[1].status, Status::Caution);
    assert_eq!(status(&output, "position_limit"), Status::Fail);
    assert!(
        output
            .checks
            .iter()
            .all(|c| !c.message_en.is_empty() && !c.message_tr.is_empty())
    );
    order.profile.risk_percent = dec!(0.01);
    order.manual_stop = Some(dec!(80));
    let output = calculate(&order, market(CoinClass::Small)).unwrap();
    assert!(output.checks.iter().all(|c| c.status == Status::Pass));
    let mut m = market(CoinClass::Core);
    m.quantity_step = dec!(1000000);
    assert_eq!(
        status(&calculate(&order, m).unwrap(), "quantity_step"),
        Status::Fail
    );
}

#[test]
fn invalid_inputs_are_typed_errors() {
    for (side, stop, expected) in [
        (Side::Long, dec!(101), Error::WrongSide),
        (Side::Short, dec!(99), Error::WrongSide),
        (Side::Long, dec!(100), Error::ZeroDistance),
        (Side::Short, dec!(100), Error::ZeroDistance),
        (Side::Long, dec!(0), Error::NonPositive("manual_stop")),
    ] {
        assert_eq!(
            calculate(&input(side, dec!(100), Some(stop)), market(CoinClass::Core)).unwrap_err(),
            expected
        );
    }
    let mut order = input(Side::Long, dec!(100), None);
    order.entry = dec!(0);
    assert_eq!(
        calculate(&order, market(CoinClass::Core)).unwrap_err(),
        Error::NonPositive("entry_price")
    );
    order.entry = dec!(100);
    order.equity = dec!(-1);
    assert_eq!(
        calculate(&order, market(CoinClass::Core)).unwrap_err(),
        Error::NonPositive("account_equity")
    );
    order.equity = dec!(1000);
    order.profile.risk_percent = dec!(101);
    assert_eq!(
        calculate(&order, market(CoinClass::Core)).unwrap_err(),
        Error::OutOfRange("risk_profile")
    );
    order.profile.risk_percent = dec!(0);
    assert_eq!(
        calculate(&order, market(CoinClass::Core)).unwrap_err(),
        Error::NonPositive("risk_per_trade_percent")
    );
    let order = input(Side::Long, dec!(1), Some(dec!(0.5)));
    assert_eq!(
        calculate(&order, market(CoinClass::Core)).unwrap_err(),
        Error::NonPositive("suggested_stop")
    );
    assert_eq!(mul(Decimal::MAX, dec!(2)), Err(Error::Arithmetic));
    assert_eq!(
        mul(dec!(0.000000000000001), dec!(0.000000000000001)),
        Err(Error::Arithmetic)
    );
}

#[test]
fn decimal_wire_grammar_has_no_floats_or_silent_rounding() {
    for value in ["1", "0.00000001", "1000000000000", "123.12345678"] {
        assert!(parse_decimal(value, "test").is_ok(), "{value}");
    }
    for value in [
        "",
        "0",
        "-1",
        "+1",
        "1e3",
        "NaN",
        " 1",
        "1 ",
        "1.",
        ".1",
        "1.123456789",
        "1000000000001",
        "1.2.3",
    ] {
        assert!(parse_decimal(value, "test").is_err(), "{value}");
    }
    for symbol in ["btcUSDT", "BTC/USDT", "USDT", "BTCUSD", "BTCUSDT "] {
        assert_eq!(validate_symbol(symbol), Err(Error::InvalidSymbol));
    }
}

#[test]
fn full_precision_wilder_atr_rounds_up_before_sizing() {
    let raw_atr = dec!(1.2345678912345678912345678912);
    let mut m = market(CoinClass::Core);
    m.atr = raw_atr;
    for side in [Side::Long, Side::Short] {
        let result = calculate(&input(side, dec!(100), None), m).unwrap();
        assert_eq!(result.atr, dec!(1.2345679));
        assert!(result.atr >= raw_atr);
        assert_eq!(result.distance, dec!(1.85185185));
        assert!(result.actual_risk <= result.risk_amount);
    }
    m.atr = dec!(0.0000000000000000000000000001);
    assert_eq!(
        calculate(&input(Side::Long, dec!(100), None), m)
            .unwrap()
            .atr,
        QUANTITY_STEP
    );
}

#[test]
fn rejects_invalid_market_profile_and_unrepresentable_prices() {
    let order = input(Side::Long, dec!(100), None);
    for (atr, step, expected) in [
        (dec!(0), QUANTITY_STEP, Error::NonPositive("atr")),
        (dec!(-1), QUANTITY_STEP, Error::NonPositive("atr")),
        (dec!(1), dec!(0), Error::NonPositive("quantity_step")),
        (dec!(1000000000001), QUANTITY_STEP, Error::OutOfRange("atr")),
    ] {
        assert_eq!(
            calculate(
                &order,
                Market {
                    class: CoinClass::Core,
                    atr,
                    quantity_step: step
                }
            )
            .unwrap_err(),
            expected
        );
    }
    for reward in [dec!(0), dec!(-1), dec!(101)] {
        let mut order = order.clone();
        order.profile.target_risk_reward = reward;
        assert!(calculate(&order, market(CoinClass::Core)).is_err());
    }
    let short = input(Side::Short, dec!(100), Some(dec!(200)));
    let result = calculate(&short, market(CoinClass::Core)).unwrap();
    assert_eq!(result.target_price, dec!(-100));
    assert_eq!(status(&result, "target_price"), Status::Fail);
    assert_eq!(
        add(dec!(1000000000000), dec!(0.00000000000000001)),
        Err(Error::Arithmetic)
    );
    assert_eq!(
        add(dec!(1000000000000), dec!(0.0000000000000001)),
        Ok(dec!(1000000000000.0000000000000001))
    );
    assert_eq!(
        mul(dec!(0.1234567890123456789012345678), dec!(100)),
        Ok(dec!(12.34567890123456789012345678))
    );
}

fn status(output: &Output, id: &str) -> Status {
    output
        .checks
        .iter()
        .find(|c| c.rule_id == id)
        .unwrap()
        .status
}

#[test]
fn spot_allocation_includes_holdings_and_disallows_leverage() {
    let mut order = input(Side::Long, dec!(100), Some(dec!(90)));
    order.current_class_holding = dec!(2000);
    assert_eq!(
        status(
            &calculate(&order, market(CoinClass::Core)).unwrap(),
            "position_limit"
        ),
        Status::Pass
    );
    order.current_class_holding += dec!(0.00000001);
    assert_eq!(
        status(
            &calculate(&order, market(CoinClass::Core)).unwrap(),
            "position_limit"
        ),
        Status::Fail
    );
    for leverage in [None, Some(0), Some(1), Some(2), Some(125), Some(126)] {
        order.leverage = leverage;
        assert_eq!(
            calculate(&order, market(CoinClass::Core)).is_ok(),
            leverage.is_none() || leverage == Some(1)
        );
        order.account_market = AccountMarket::Futures;
        assert_eq!(
            calculate(&order, market(CoinClass::Core)).is_ok(),
            leverage.is_some_and(|v| (1..=125).contains(&v))
        );
        order.account_market = AccountMarket::Spot;
    }
}

#[test]
fn futures_uses_margin_and_liquidation_instead_of_allocation() {
    let mut order = input(Side::Long, dec!(100), Some(dec!(99)));
    order.account_market = AccountMarket::Futures;
    order.leverage = Some(10);
    let output = calculate(&order, market(CoinClass::Core)).unwrap();
    assert_eq!(output.notional, dec!(20000));
    assert_eq!(output.initial_margin, Some(dec!(2000)));
    assert_eq!(output.liquidation_price, Some(dec!(90) / dec!(0.995)));
    assert!(
        output
            .checks
            .iter()
            .all(|c| c.rule_id != "position_limit" && c.rule_id != "spot_side")
    );
    assert_eq!(status(&output, "liquidation_stop"), Status::Pass);
    assert_eq!(status(&output, "initial_margin"), Status::Pass);
    order.manual_stop = Some(dec!(99.99));
    assert_eq!(
        status(
            &calculate(&order, market(CoinClass::Core)).unwrap(),
            "initial_margin"
        ),
        Status::Fail
    );
    order.leverage = Some(1);
    assert_eq!(
        calculate(&order, market(CoinClass::Core))
            .unwrap()
            .liquidation_price,
        Some(dec!(0))
    );
}

#[test]
fn liquidation_boundaries_are_strict_in_both_directions() {
    for (side, entry, liquidation, cases) in [
        (
            Side::Long,
            dec!(199),
            dec!(180),
            [
                (dec!(179), Status::Fail),
                (dec!(180), Status::Fail),
                (dec!(181), Status::Caution),
                (dec!(183), Status::Pass),
            ],
        ),
        (
            Side::Short,
            dec!(201),
            dec!(220),
            [
                (dec!(221), Status::Fail),
                (dec!(220), Status::Fail),
                (dec!(219), Status::Caution),
                (dec!(217), Status::Pass),
            ],
        ),
    ] {
        for (stop, expected) in cases {
            let mut order = input(side, entry, Some(stop));
            order.account_market = AccountMarket::Futures;
            order.leverage = Some(10);
            let output = calculate(&order, market(CoinClass::Core)).unwrap();
            assert_eq!(output.liquidation_price, Some(liquidation));
            assert_eq!(status(&output, "liquidation_stop"), expected);
        }
    }
}

#[test]
fn discipline_percentage_thresholds_apply_to_both_markets() {
    for account_market in [AccountMarket::Spot, AccountMarket::Futures] {
        for (id, cases) in [
            (
                "total_open_risk",
                [
                    (dec!(2.49999999), Status::Pass),
                    (dec!(2.5), Status::Caution),
                    (dec!(4), Status::Caution),
                    (dec!(4.00000001), Status::Fail),
                ],
            ),
            (
                "daily_loss",
                [
                    (dec!(3.74999999), Status::Pass),
                    (dec!(3.75), Status::Caution),
                    (dec!(4.99), Status::Caution),
                    (dec!(5), Status::Fail),
                ],
            ),
            (
                "weekly_loss",
                [
                    (dec!(7.49999999), Status::Pass),
                    (dec!(7.5), Status::Caution),
                    (dec!(9.99), Status::Caution),
                    (dec!(10), Status::Fail),
                ],
            ),
        ] {
            for (value, expected) in cases {
                let mut order = input(Side::Long, dec!(100), Some(dec!(90)));
                order.account_market = account_market;
                order.leverage = Some(2);
                if account_market == AccountMarket::Spot {
                    order.leverage = None;
                }
                match id {
                    "total_open_risk" => order.activity.open_risk_percent = value,
                    "daily_loss" => order.activity.daily_loss_percent = value,
                    _ => order.activity.weekly_loss_percent = value,
                }
                let output = calculate(&order, market(CoinClass::Core)).unwrap();
                assert_eq!(status(&output, id), expected, "{id} {value}");
                assert!(output.checks.iter().all(|c| !c.message_en.is_empty()
                    && !c.message_tr.is_empty()
                    && !c.rule_id.is_empty()));
            }
        }
    }
}

#[test]
fn discipline_count_thresholds_and_custom_profiles() {
    for (id, cases) in [
        (
            "open_positions",
            [
                (2, Status::Pass),
                (3, Status::Caution),
                (4, Status::Caution),
                (5, Status::Fail),
            ],
        ),
        (
            "losing_trades_today",
            [
                (1, Status::Pass),
                (2, Status::Caution),
                (3, Status::Fail),
                (4, Status::Fail),
            ],
        ),
        (
            "stopped_trades_this_week",
            [
                (3, Status::Pass),
                (4, Status::Caution),
                (5, Status::Fail),
                (6, Status::Fail),
            ],
        ),
    ] {
        for (value, expected) in cases {
            let mut order = input(Side::Long, dec!(100), None);
            match id {
                "open_positions" => order.activity.open_positions = value,
                "losing_trades_today" => order.activity.losing_trades_today = value,
                _ => order.activity.stopped_trades_this_week = value,
            }
            assert_eq!(
                status(&calculate(&order, market(CoinClass::Core)).unwrap(), id),
                expected
            );
        }
    }
    let mut order = input(Side::Long, dec!(100), None);
    order.profile.total_open_risk_limit_percent = dec!(1);
    order.profile.max_open_positions = 1;
    order.profile.max_losing_trades_today = 1;
    order.profile.max_stopped_trades_this_week = 1;
    order.profile.daily_loss_limit_percent = dec!(1);
    order.profile.weekly_loss_limit_percent = dec!(2);
    order.activity.daily_loss_percent = dec!(1);
    order.activity.weekly_loss_percent = dec!(2);
    order.activity.losing_trades_today = 1;
    order.activity.stopped_trades_this_week = 1;
    let output = calculate(&order, market(CoinClass::Core)).unwrap();
    for id in [
        "total_open_risk",
        "daily_loss",
        "weekly_loss",
        "losing_trades_today",
        "stopped_trades_this_week",
    ] {
        assert_eq!(status(&output, id), Status::Fail);
    }
    assert_eq!(status(&output, "open_positions"), Status::Caution);
    order.activity.open_positions = u32::MAX;
    assert_eq!(
        status(
            &calculate(&order, market(CoinClass::Core)).unwrap(),
            "open_positions"
        ),
        Status::Fail
    );
    order.profile.max_open_positions = 0;
    assert!(calculate(&order, market(CoinClass::Core)).is_err());
    order.profile.max_open_positions = 5;
    order.activity.daily_loss_percent = dec!(-1);
    assert!(calculate(&order, market(CoinClass::Core)).is_err());
    assert_eq!(
        parse_nonnegative_decimal("0", "activity"),
        Ok(Decimal::ZERO)
    );
    assert!(parse_nonnegative_decimal("-1", "activity").is_err());
}

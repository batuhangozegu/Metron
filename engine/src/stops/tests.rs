use futures_util::StreamExt;
use rust_decimal::dec;

use super::*;

pub(super) fn position(id: &str, symbol: &str, side: &str, stop: Decimal) -> Position {
    Position {
        id: id.into(),
        symbol: symbol.into(),
        side: side.into(),
        stop,
    }
}

#[test]
fn stop_crossings_include_equality_in_both_directions() {
    for (side, price, expected) in [
        ("long", dec!(99), true),
        ("long", dec!(100), true),
        ("long", dec!(101), false),
        ("short", dec!(99), false),
        ("short", dec!(100), true),
        ("short", dec!(101), true),
        ("buy", dec!(99), false),
    ] {
        assert_eq!(crossed(side, dec!(100), price), expected, "{side} {price}");
    }
}

#[test]
fn comparisons_preserve_decimal_precision() {
    let stop = dec!(0.1234567890123456789012345678);
    let below = dec!(0.1234567890123456789012345677);
    let above = dec!(0.1234567890123456789012345679);
    assert!(crossed("long", stop, below));
    assert!(!crossed("long", stop, above));
    assert!(!crossed("short", stop, below));
    assert!(crossed("short", stop, above));
    assert!(crossed("long", stop, stop));
    assert!(crossed("short", stop, stop));
    assert!(!crossed(
        "long",
        dec!(1000000000000.00000001),
        dec!(1000000000000.00000002)
    ));
}

#[test]
fn dedupe_expires_and_closed_positions_are_removed() {
    let hub = Hub::default();
    let mut receiver = hub.sender.subscribe();
    let mut watcher = Watcher::default();
    let positions = [position("one", "BTCUSDT", "long", dec!(100))];
    let prices = HashMap::from([("BTCUSDT".into(), dec!(99))]);
    let now = Instant::now();
    watcher.scan(&positions, &prices, &hub, now, Utc::now());
    assert_eq!(receiver.try_recv().unwrap().position_id, "one");
    watcher.scan(
        &positions,
        &prices,
        &hub,
        now + RESEND_AFTER - Duration::from_millis(1),
        Utc::now(),
    );
    assert!(matches!(
        receiver.try_recv(),
        Err(broadcast::error::TryRecvError::Empty)
    ));
    watcher.scan(&positions, &prices, &hub, now + RESEND_AFTER, Utc::now());
    assert_eq!(receiver.try_recv().unwrap().position_id, "one");
    watcher.scan(&[], &prices, &hub, now + RESEND_AFTER, Utc::now());
    assert!(watcher.emitted.is_empty());
}

#[test]
fn missing_and_invalid_prices_do_not_trigger_or_dedupe() {
    let hub = Hub::default();
    let mut receiver = hub.sender.subscribe();
    let mut watcher = Watcher::default();
    let positions = [position("one", "BTCUSDT", "long", dec!(100))];
    for prices in [
        HashMap::new(),
        HashMap::from([("BTCUSDT".into(), dec!(0))]),
        HashMap::from([("BTCUSDT".into(), dec!(-1))]),
    ] {
        watcher.scan(&positions, &prices, &hub, Instant::now(), Utc::now());
        assert!(watcher.emitted.is_empty());
        assert!(matches!(
            receiver.try_recv(),
            Err(broadcast::error::TryRecvError::Empty)
        ));
    }
}

#[test]
fn no_subscribers_does_not_consume_a_hit() {
    let hub = Hub::default();
    let mut watcher = Watcher::default();
    let positions = [position("one", "BTCUSDT", "short", dec!(100))];
    let prices = HashMap::from([("BTCUSDT".into(), dec!(101))]);
    watcher.scan(&positions, &prices, &hub, Instant::now(), Utc::now());
    assert!(watcher.emitted.is_empty());
    let mut receiver = hub.sender.subscribe();
    watcher.scan(&positions, &prices, &hub, Instant::now(), Utc::now());
    assert_eq!(receiver.try_recv().unwrap().price, "101");
}

#[tokio::test]
async fn lagging_stream_errors_then_ends() {
    let hub = Hub::default();
    let mut stream = hub.subscribe();
    for _ in 0..=STREAM_CAPACITY {
        hub.sender.send(StopHit::default()).unwrap();
    }
    let status = stream.next().await.unwrap().unwrap_err();
    assert_eq!(status.code(), tonic::Code::Unavailable);
    assert!(stream.next().await.is_none());
}

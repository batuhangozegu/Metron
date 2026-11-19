use rust_decimal::dec;
use tonic::transport::Server;

use super::{tests::position, *};
use crate::{Engine, pb};

#[tokio::test]
async fn stop_hits_fan_out_over_grpc_and_resume_after_reconnect() {
    let engine = Engine::default();
    let hub = engine.stops.clone();
    let listener = tokio::net::TcpListener::bind("127.0.0.1:0").await.unwrap();
    let addr = listener.local_addr().unwrap();
    let server = tokio::spawn(
        Server::builder()
            .add_service(pb::engine_service_server::EngineServiceServer::new(engine))
            .serve_with_incoming(tokio_stream::wrappers::TcpListenerStream::new(listener)),
    );
    let mut client =
        pb::engine_service_client::EngineServiceClient::connect(format!("http://{addr}"))
            .await
            .unwrap();
    let mut first = client
        .watch_stops(pb::WatchStopsRequest {})
        .await
        .unwrap()
        .into_inner();
    let mut second = client
        .watch_stops(pb::WatchStopsRequest {})
        .await
        .unwrap()
        .into_inner();
    let mut watcher = Watcher::default();
    let positions = [
        position(
            "11111111-1111-4111-8111-111111111111",
            "BTCUSDT",
            "long",
            dec!(99.12345678901234567890123456),
        ),
        position(
            "22222222-2222-4222-8222-222222222222",
            "ETHUSDT",
            "short",
            dec!(200),
        ),
        position("safe", "SOLUSDT", "long", dec!(10)),
    ];
    let prices = HashMap::from([
        ("BTCUSDT".into(), dec!(99.12345678901234567890123455)),
        ("ETHUSDT".into(), dec!(200)),
        ("SOLUSDT".into(), dec!(11)),
    ]);
    let now = Instant::now();
    let observed_at = DateTime::parse_from_rfc3339("2026-11-19T12:00:00Z")
        .unwrap()
        .with_timezone(&Utc);
    watcher.scan(&positions, &prices, &hub, now, observed_at);
    for position in &positions[..2] {
        let one = tokio::time::timeout(Duration::from_secs(2), first.message())
            .await
            .unwrap()
            .unwrap()
            .unwrap();
        let two = tokio::time::timeout(Duration::from_secs(2), second.message())
            .await
            .unwrap()
            .unwrap()
            .unwrap();
        assert_eq!(one, two);
        assert_eq!(one.position_id, position.id);
        assert_eq!(one.symbol, position.symbol);
        assert_eq!(one.stop, position.stop.normalize().to_string());
        assert_eq!(one.price, prices[&position.symbol].normalize().to_string());
        assert_eq!(one.observed_at, "2026-11-19T12:00:00.000Z");
    }
    watcher.scan(&positions, &prices, &hub, now + POLL_INTERVAL, observed_at);
    assert!(
        tokio::time::timeout(Duration::from_millis(30), first.message())
            .await
            .is_err()
    );
    drop(first);
    let mut reconnected = client
        .watch_stops(pb::WatchStopsRequest {})
        .await
        .unwrap()
        .into_inner();
    watcher.scan(&positions, &prices, &hub, now + RESEND_AFTER, observed_at);
    assert_eq!(
        tokio::time::timeout(Duration::from_secs(2), reconnected.message())
            .await
            .unwrap()
            .unwrap()
            .unwrap()
            .position_id,
        positions[0].id
    );
    server.abort();
    let _ = server.await;
}

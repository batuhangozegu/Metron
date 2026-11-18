use std::sync::Arc;

use chrono::Utc;
use rust_decimal::dec;
use tonic::{Code, Request, Status};

use super::grpc::{MarketSource, Service, Snapshot};
use crate::market::classify::CoinClass;
use crate::pb::{
    self,
    risk_service_server::{RiskService, RiskServiceServer},
};

struct FakeSource(Result<Option<Snapshot>, Status>);

#[tonic::async_trait]
impl MarketSource for FakeSource {
    async fn snapshot(&self, symbol: &str) -> Result<Option<Snapshot>, Status> {
        assert_eq!(symbol, "BTCUSDT");
        self.0.clone()
    }
}

fn request() -> pb::CheckRequest {
    pb::CheckRequest {
        symbol: "BTCUSDT".into(),
        side: "long".into(),
        entry_price: "100".into(),
        manual_stop: None,
        account_equity: "10000.12345678".into(),
        risk_profile: None,
        ..Default::default()
    }
}

fn snapshot() -> Snapshot {
    Snapshot {
        class: CoinClass::Core,
        atr: Some(dec!(10)),
        updated_at: Utc::now(),
    }
}

#[tokio::test]
async fn risk_over_the_wire_uses_latest_class_and_atr() {
    let mut snapshot = snapshot();
    snapshot.class = CoinClass::Small;
    let service = Service {
        source: Arc::new(FakeSource(Ok(Some(snapshot)))),
    };
    let listener = tokio::net::TcpListener::bind("127.0.0.1:0").await.unwrap();
    let addr = listener.local_addr().unwrap();
    let server = tokio::spawn(
        tonic::transport::Server::builder()
            .add_service(RiskServiceServer::new(service))
            .serve_with_incoming(tokio_stream::wrappers::TcpListenerStream::new(listener)),
    );
    let mut client = pb::risk_service_client::RiskServiceClient::connect(format!("http://{addr}"))
        .await
        .unwrap();
    let result = client.check(request()).await.unwrap().into_inner();
    assert_eq!(result.coin_class, "small");
    assert_eq!(result.atr, "10");
    assert_eq!(result.atr_multiplier, "2");
    assert_eq!(result.stop, "80");
    assert_eq!(result.account_equity, "10000.12345678");
    assert_eq!(result.risk_amount, "50.0006172839");
    assert_eq!(result.position_size, "2.50003086");
    assert_eq!(result.actual_risk, "50.0006172");
    assert_eq!(result.verdict, "caution");
    assert_eq!((result.passed, result.cautions, result.failed), (11, 1, 0));
    assert!(result.allowed);
    assert!(!result.market_updated_at.is_empty());
    let mut short = request();
    short.side = "short".into();
    short.manual_stop = Some("110".into());
    let result = client.check(short).await.unwrap().into_inner();
    assert!(!result.allowed);
    assert_eq!(result.verdict, "fail");
    assert_eq!(result.suggested_stop, "120");
    assert_eq!(result.stop, "110");
    server.abort();
}

#[tokio::test]
async fn market_lookup_errors_and_freshness_fail_closed() {
    let mut stale = snapshot();
    stale.updated_at -= chrono::Duration::hours(2);
    let mut future = snapshot();
    future.updated_at += chrono::Duration::minutes(5);
    let mut missing_atr = snapshot();
    missing_atr.atr = None;
    let mut zero_atr = snapshot();
    zero_atr.atr = Some(dec!(0));
    let mut stable = snapshot();
    stable.class = CoinClass::Stable;
    for (source, code, message) in [
        (Ok(None), Code::NotFound, "symbol_not_found"),
        (
            Err(Status::unavailable("market_unavailable")),
            Code::Unavailable,
            "market_unavailable",
        ),
        (Ok(Some(stale)), Code::Unavailable, "market_stale"),
        (Ok(Some(future)), Code::Unavailable, "market_stale"),
        (Ok(Some(missing_atr)), Code::Unavailable, "atr_unavailable"),
        (Ok(Some(zero_atr)), Code::Unavailable, "atr_unavailable"),
        (
            Ok(Some(stable)),
            Code::InvalidArgument,
            "stablecoin_unsupported",
        ),
    ] {
        let service = Service {
            source: Arc::new(FakeSource(source)),
        };
        let error = service.check(Request::new(request())).await.unwrap_err();
        assert_eq!(error.code(), code);
        assert_eq!(error.message(), message);
    }
}

#[tokio::test]
async fn transport_rejects_bad_decimal_side_and_stop() {
    let service = Service {
        source: Arc::new(FakeSource(Ok(Some(snapshot())))),
    };
    let mut bad_decimal = request();
    bad_decimal.entry_price = "1e2".into();
    let mut bad_side = request();
    bad_side.side = "buy".into();
    let mut wrong_stop = request();
    wrong_stop.manual_stop = Some("101".into());
    for (request, message) in [
        (bad_decimal, "invalid_decimal"),
        (bad_side, "invalid_side"),
        (wrong_stop, "wrong_stop_side"),
    ] {
        let error = service.check(Request::new(request)).await.unwrap_err();
        assert_eq!(error.code(), Code::InvalidArgument);
        assert_eq!(error.message(), message);
    }
}

#[tokio::test]
async fn exposes_stored_and_effective_atr() {
    let mut snapshot = snapshot();
    snapshot.atr = Some(dec!(1.2345678912345678912345678912));
    let service = Service {
        source: Arc::new(FakeSource(Ok(Some(snapshot)))),
    };
    let result = service
        .check(Request::new(request()))
        .await
        .unwrap()
        .into_inner();
    assert_eq!(result.source_atr, "1.2345678912345678912345678912");
    assert_eq!(result.atr, "1.2345679");
    assert_eq!(result.suggested_stop, "98.14814815");
}

#[tokio::test]
async fn futures_transport_profiles_activity_and_verdict() {
    let service = Service {
        source: Arc::new(FakeSource(Ok(Some(snapshot())))),
    };
    let mut req = request();
    req.market = "futures".into();
    req.leverage = Some(10);
    req.manual_stop = Some("99".into());
    req.risk_profile = Some(pb::RiskProfile {
        max_losing_trades_today: Some(4),
        ..Default::default()
    });
    req.activity = Some(pb::Activity {
        losing_trades_today: 3,
        ..Default::default()
    });
    let output = service
        .check(Request::new(req.clone()))
        .await
        .unwrap()
        .into_inner();
    assert_eq!(output.market, "futures");
    assert_eq!(output.leverage, 10);
    assert!(output.liquidation_price.is_some());
    assert!(output.initial_margin.is_some());
    assert!(output.max_position_percent.is_empty());
    assert!(output.allowed);
    assert_eq!(output.verdict, "caution");
    req.activity.as_mut().unwrap().losing_trades_today = 4;
    let output = service
        .check(Request::new(req.clone()))
        .await
        .unwrap()
        .into_inner();
    assert!(!output.allowed);
    assert_eq!(output.verdict, "fail");
    assert_eq!(output.failed, 1);
    assert_eq!(
        output.passed + output.cautions + output.failed,
        output.checks.len() as u32
    );
    req.leverage = None;
    assert_eq!(
        service
            .check(Request::new(req.clone()))
            .await
            .unwrap_err()
            .message(),
        "invalid_leverage"
    );
    req.market = "margin".into();
    assert_eq!(
        service
            .check(Request::new(req))
            .await
            .unwrap_err()
            .message(),
        "invalid_market"
    );
}

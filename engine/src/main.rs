//! Metron risk engine: market data ingestion and gRPC risk services.

mod market;
mod risk;
mod stops;

use std::net::SocketAddr;
use std::time::{SystemTime, UNIX_EPOCH};

use tonic::{Request, Response, Status, transport::Server};
use tracing::info;

pub mod pb {
    tonic::include_proto!("metron.engine.v1");
}

use pb::engine_service_server::{EngineService, EngineServiceServer};
use pb::{PingRequest, PingResponse, WatchStopsRequest};

/// Where the engine listens unless `ENGINE_ADDR` says otherwise.
/// Not 50051: Windows often reserves the 50000–50059 range for Hyper-V.
const DEFAULT_ADDR: &str = "127.0.0.1:9090";

#[derive(Debug, Default)]
pub struct Engine {
    stops: stops::Hub,
}

#[tonic::async_trait]
impl EngineService for Engine {
    type WatchStopsStream = stops::HitStream;

    async fn watch_stops(
        &self,
        _request: Request<WatchStopsRequest>,
    ) -> Result<Response<Self::WatchStopsStream>, Status> {
        Ok(Response::new(self.stops.subscribe()))
    }

    async fn ping(&self, request: Request<PingRequest>) -> Result<Response<PingResponse>, Status> {
        let message = request.into_inner().message;
        let server_time_unix_ms = SystemTime::now()
            .duration_since(UNIX_EPOCH)
            .map_err(|e| Status::internal(format!("system clock before Unix epoch: {e}")))?
            .as_millis() as i64;

        Ok(Response::new(PingResponse {
            message,
            engine_version: env!("CARGO_PKG_VERSION").to_string(),
            server_time_unix_ms,
        }))
    }
}

#[tokio::main]
async fn main() -> Result<(), Box<dyn std::error::Error>> {
    tracing_subscriber::fmt()
        .with_env_filter(
            tracing_subscriber::EnvFilter::try_from_default_env().unwrap_or_else(|_| "info".into()),
        )
        .init();

    let addr: SocketAddr = std::env::var("ENGINE_ADDR")
        .unwrap_or_else(|_| DEFAULT_ADDR.to_string())
        .parse()?;

    let config = market::Config::from_env()?;
    let pool = sqlx::postgres::PgPoolOptions::new()
        .max_connections(5)
        .connect(&config.database_url)
        .await?;
    let redis = redis::Client::open(config.redis_url.as_str())?
        .get_connection_manager()
        .await?;
    let rest = market::binance::Rest::new(&config.rest_url)?;
    let sync = tokio::spawn(market::sync::run(config.clone(), rest, pool.clone()));
    let ticker = tokio::spawn(market::ticker::run(config, redis.clone()));
    let engine = Engine::default();
    let stop_watcher = tokio::spawn(stops::run(pool.clone(), redis, engine.stops.clone()));

    info!(%addr, version = env!("CARGO_PKG_VERSION"), "engine listening");

    let result = Server::builder()
        .add_service(EngineServiceServer::new(engine))
        .add_service(pb::risk_service_server::RiskServiceServer::new(
            risk::grpc::Service {
                source: std::sync::Arc::new(pool.clone()),
            },
        ))
        .serve_with_shutdown(addr, shutdown_signal())
        .await;

    sync.abort();
    ticker.abort();
    stop_watcher.abort();
    let _ = tokio::join!(sync, ticker, stop_watcher);
    pool.close().await;
    result?;

    Ok(())
}

async fn shutdown_signal() {
    #[cfg(unix)]
    {
        let mut terminate =
            tokio::signal::unix::signal(tokio::signal::unix::SignalKind::terminate())
                .expect("install SIGTERM handler");
        tokio::select! {
            _ = tokio::signal::ctrl_c() => {},
            _ = terminate.recv() => {},
        }
    }
    #[cfg(not(unix))]
    let _ = tokio::signal::ctrl_c().await;
    info!("shutting down");
}

#[cfg(test)]
mod tests {
    use super::*;

    #[tokio::test]
    async fn ping_echoes_message_and_reports_version() {
        let response = Engine::default()
            .ping(Request::new(PingRequest {
                message: "hello".into(),
            }))
            .await
            .unwrap()
            .into_inner();

        assert_eq!(response.message, "hello");
        assert_eq!(response.engine_version, env!("CARGO_PKG_VERSION"));
        assert!(response.server_time_unix_ms > 0);
    }

    /// Starts a real server on a free port and calls it with the generated
    /// client, so the whole gRPC path (HTTP/2, encoding, routing) is covered.
    #[tokio::test]
    async fn ping_over_the_wire() {
        let listener = tokio::net::TcpListener::bind("127.0.0.1:0").await.unwrap();
        let addr = listener.local_addr().unwrap();
        tokio::spawn(
            Server::builder()
                .add_service(EngineServiceServer::new(Engine::default()))
                .serve_with_incoming(tokio_stream::wrappers::TcpListenerStream::new(listener)),
        );

        let mut client =
            pb::engine_service_client::EngineServiceClient::connect(format!("http://{addr}"))
                .await
                .unwrap();
        let response = client
            .ping(PingRequest {
                message: "over the wire".into(),
            })
            .await
            .unwrap()
            .into_inner();

        assert_eq!(response.message, "over the wire");
    }
}

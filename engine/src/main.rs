//! Metron risk engine. For now it only answers the `Ping` health check.

use std::net::SocketAddr;
use std::time::{SystemTime, UNIX_EPOCH};

use tonic::{Request, Response, Status, transport::Server};
use tracing::info;

pub mod pb {
    tonic::include_proto!("metron.engine.v1");
}

use pb::engine_service_server::{EngineService, EngineServiceServer};
use pb::{PingRequest, PingResponse};

/// Where the engine listens unless `ENGINE_ADDR` says otherwise.
/// Not 50051: Windows often reserves the 50000–50059 range for Hyper-V.
const DEFAULT_ADDR: &str = "127.0.0.1:9090";

#[derive(Debug, Default)]
pub struct Engine;

#[tonic::async_trait]
impl EngineService for Engine {
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

    info!(%addr, version = env!("CARGO_PKG_VERSION"), "engine listening");

    Server::builder()
        .add_service(EngineServiceServer::new(Engine))
        .serve_with_shutdown(addr, async {
            let _ = tokio::signal::ctrl_c().await;
            info!("shutting down");
        })
        .await?;

    Ok(())
}

#[cfg(test)]
mod tests {
    use super::*;

    #[tokio::test]
    async fn ping_echoes_message_and_reports_version() {
        let response = Engine
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
                .add_service(EngineServiceServer::new(Engine))
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

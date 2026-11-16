# Build context is the repo root, because the engine compiles ../proto:
#   docker build -f engine/Dockerfile .
FROM rust:1.98-slim-trixie AS build
WORKDIR /src
COPY proto proto
COPY engine engine
WORKDIR /src/engine
RUN cargo build --release --locked

FROM debian:trixie-slim
COPY --from=build /src/engine/target/release/metron-engine /usr/local/bin/metron-engine
USER 65532:65532
ENV ENGINE_ADDR=0.0.0.0:9090
EXPOSE 9090
ENTRYPOINT ["metron-engine"]

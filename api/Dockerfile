# Build context is the repo root, because code is generated from ../proto:
#   docker build -f api/Dockerfile .
FROM golang:1.27 AS build
WORKDIR /src/api
COPY api/go.mod api/go.sum ./
COPY api/tools/go.mod api/tools/go.sum tools/
RUN go mod download && go mod download -modfile=tools/go.mod
COPY proto /src/proto
COPY api ./
RUN go generate ./... && CGO_ENABLED=0 go build -trimpath -o /out/api ./cmd/api

FROM gcr.io/distroless/static-debian12:nonroot
COPY --from=build /out/api /api
ENV API_ADDR=0.0.0.0:8080
EXPOSE 8080
ENTRYPOINT ["/api"]

# Runs the goose migrations in db/ against the compose Postgres.
# Build context is the repo root:  docker build -f db/Dockerfile .
FROM golang:1.27-alpine AS build
RUN CGO_ENABLED=0 go install github.com/pressly/goose/v3/cmd/goose@v3.30.1

FROM alpine:3.22
COPY --from=build /go/bin/goose /usr/local/bin/goose
COPY db /db
COPY db/migrate.sh /usr/local/bin/migrate
RUN chmod 755 /usr/local/bin/migrate
USER 65532:65532
ENTRYPOINT ["migrate"]

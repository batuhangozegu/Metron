// Command api is Metron's Go API gateway. For now it checks the Rust engine
// on startup and serves GET /healthz.
package main

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	enginev1 "github.com/batuhangozegu/metron/api/internal/gen/metron/engine/v1"
	"github.com/batuhangozegu/metron/api/internal/health"
)

func main() {
	log := slog.New(slog.NewTextHandler(os.Stderr, nil))
	if err := run(log); err != nil {
		log.Error("api stopped", "err", err)
		os.Exit(1)
	}
}

func run(log *slog.Logger) error {
	apiAddr := envOr("API_ADDR", "127.0.0.1:8080")
	engineAddr := envOr("ENGINE_ADDR", "127.0.0.1:9090")

	// The connection is lazy: NewClient does not dial, so the API starts
	// even when the engine is down and reconnects once it is up.
	conn, err := grpc.NewClient(engineAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return err
	}
	defer conn.Close()
	engine := enginev1.NewEngineServiceClient(conn)

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	if s := health.CheckEngine(ctx, engine); s.Reachable {
		log.Info("engine reachable", "addr", engineAddr, "version", s.Version, "latency_ms", s.LatencyMS)
	} else {
		log.Warn("engine not reachable; continuing", "addr", engineAddr, "err", s.Error)
	}

	r := chi.NewRouter()
	r.Use(middleware.RequestID, middleware.Recoverer)
	r.Get("/healthz", health.Handler(engine))

	srv := &http.Server{Addr: apiAddr, Handler: r, ReadHeaderTimeout: 5 * time.Second}
	errCh := make(chan error, 1)
	go func() {
		log.Info("api listening", "addr", apiAddr)
		errCh <- srv.ListenAndServe()
	}()

	select {
	case err := <-errCh:
		return err
	case <-ctx.Done():
		log.Info("shutting down")
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := srv.Shutdown(shutdownCtx); err != nil && !errors.Is(err, http.ErrServerClosed) {
			return err
		}
		return nil
	}
}

func envOr(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

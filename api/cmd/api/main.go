// Command api is Metron's Go API for health, markets and user accounts.
package main

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	"github.com/batuhangozegu/metron/api/internal/accounts"
	"github.com/batuhangozegu/metron/api/internal/auth"
	"github.com/batuhangozegu/metron/api/internal/cryptobox"
	enginev1 "github.com/batuhangozegu/metron/api/internal/gen/metron/engine/v1"
	"github.com/batuhangozegu/metron/api/internal/health"
	"github.com/batuhangozegu/metron/api/internal/market"
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

	var symbols market.SymbolSource
	var authStore auth.Store
	var accountStore accounts.Store
	// No implicit database password. With no URL, health and the HTTP server
	// remain available while markets and authentication report 503.
	if databaseURL := os.Getenv("DATABASE_URL"); databaseURL != "" {
		// Use Compose's local port only when a loopback URL omits its port.
		// An explicitly supplied :5432 (or any remote endpoint) stays intact.
		if local, err := url.Parse(databaseURL); err == nil && local.Port() == "" && os.Getenv("PGPORT") == "" &&
			(local.Hostname() == "127.0.0.1" || local.Hostname() == "localhost") {
			local.Host += ":5433"
			databaseURL = local.String()
		}
		config, err := pgxpool.ParseConfig(databaseURL)
		if err != nil {
			log.Warn("invalid DATABASE_URL; markets and authentication will return 503")
		} else {
			if config.ConnConfig.Host == "" {
				config.ConnConfig.Host = "127.0.0.1"
			}
			pool, err := pgxpool.NewWithConfig(ctx, config)
			if err != nil {
				log.Warn("database unavailable; markets and authentication will return 503")
			} else {
				defer pool.Close()
				symbols = market.Postgres{Pool: pool}
				authStore = auth.Postgres{Pool: pool}
				accountStore = accounts.Postgres{Pool: pool}
			}
		}
	} else {
		log.Warn("DATABASE_URL unset; configure Postgres at 127.0.0.1:5433 with an explicit password; markets and authentication will return 503")
	}
	var tickers market.TickerSource
	redisOptions, err := redis.ParseURL(envOr("REDIS_URL", "redis://127.0.0.1:6379"))
	if err != nil {
		log.Warn("invalid REDIS_URL; prices will be null")
	} else {
		redisOptions.DialTimeout = time.Second
		redisOptions.ReadTimeout = time.Second
		redisOptions.WriteTimeout = time.Second
		redisOptions.ContextTimeoutEnabled = true
		client := redis.NewClient(redisOptions)
		defer client.Close()
		tickers = market.Redis{Client: client}
	}

	if s := health.CheckEngine(ctx, engine); s.Reachable {
		log.Info("engine reachable", "addr", engineAddr, "version", s.Version, "latency_ms", s.LatencyMS)
	} else {
		log.Warn("engine not reachable; continuing", "addr", engineAddr, "err", s.Error)
	}

	r := chi.NewRouter()
	r.Use(middleware.RequestID, middleware.Recoverer)
	r.Get("/healthz", health.Handler(engine))
	r.Get("/v1/markets", market.Handler(symbols, tickers))
	box, keyErr := cryptobox.New(os.Getenv("APP_ENCRYPTION_KEY"))
	if keyErr != nil {
		log.Warn("APP_ENCRYPTION_KEY missing or invalid; authentication endpoints will return 503", "err", keyErr)
	}
	authService, err := auth.New(authStore, box)
	if err != nil {
		return err
	}
	authHTTP := auth.HTTP{Service: authService, Secure: os.Getenv("COOKIE_SECURE") == "true"}
	authHTTP.Mount(r, accounts.Service{Store: accountStore}.Mount)

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

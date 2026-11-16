// Package health reports whether the API can reach the Rust engine.
package health

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	enginev1 "github.com/batuhangozegu/metron/api/internal/gen/metron/engine/v1"
)

// pingTimeout bounds how long a health check waits for the engine.
const pingTimeout = 2 * time.Second

// EngineStatus is the result of one Ping to the engine.
type EngineStatus struct {
	Reachable bool   `json:"reachable"`
	Version   string `json:"version,omitempty"`
	LatencyMS int64  `json:"latency_ms,omitempty"`
	Error     string `json:"error,omitempty"`
}

// Response is the body of GET /healthz.
type Response struct {
	Status string       `json:"status"` // "ok" or "degraded"
	Engine EngineStatus `json:"engine"`
}

// CheckEngine pings the engine once and reports the outcome.
func CheckEngine(ctx context.Context, client enginev1.EngineServiceClient) EngineStatus {
	ctx, cancel := context.WithTimeout(ctx, pingTimeout)
	defer cancel()

	start := time.Now()
	res, err := client.Ping(ctx, &enginev1.PingRequest{Message: "api health check"})
	if err != nil {
		return EngineStatus{Reachable: false, Error: err.Error()}
	}
	return EngineStatus{
		Reachable: true,
		Version:   res.GetEngineVersion(),
		LatencyMS: time.Since(start).Milliseconds(),
	}
}

// Handler serves GET /healthz: 200 when the engine answers, 503 otherwise.
func Handler(client enginev1.EngineServiceClient) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		engine := CheckEngine(r.Context(), client)
		body := Response{Status: "ok", Engine: engine}
		code := http.StatusOK
		if !engine.Reachable {
			body.Status = "degraded"
			code = http.StatusServiceUnavailable
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(code)
		_ = json.NewEncoder(w).Encode(body)
	}
}

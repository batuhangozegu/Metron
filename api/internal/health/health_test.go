package health

import (
	"context"
	"encoding/json"
	"net"
	"net/http"
	"net/http/httptest"
	"testing"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/test/bufconn"

	enginev1 "github.com/batuhangozegu/metron/api/internal/gen/metron/engine/v1"
)

// fakeEngine stands in for the Rust engine over an in-memory gRPC connection.
type fakeEngine struct {
	enginev1.UnimplementedEngineServiceServer
}

func (fakeEngine) Ping(_ context.Context, req *enginev1.PingRequest) (*enginev1.PingResponse, error) {
	return &enginev1.PingResponse{Message: req.GetMessage(), EngineVersion: "9.9.9"}, nil
}

// dial returns a client for an engine served over bufconn. With serve=false
// nothing listens, so every call fails like an unreachable engine.
func dial(t *testing.T, serve bool) enginev1.EngineServiceClient {
	t.Helper()
	lis := bufconn.Listen(1 << 20)
	if serve {
		srv := grpc.NewServer()
		enginev1.RegisterEngineServiceServer(srv, fakeEngine{})
		go func() { _ = srv.Serve(lis) }()
		t.Cleanup(srv.Stop)
	} else {
		_ = lis.Close()
	}

	conn, err := grpc.NewClient("passthrough:///bufnet",
		grpc.WithContextDialer(func(ctx context.Context, _ string) (net.Conn, error) { return lis.DialContext(ctx) }),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = conn.Close() })
	return enginev1.NewEngineServiceClient(conn)
}

func get(t *testing.T, client enginev1.EngineServiceClient) (int, Response) {
	t.Helper()
	rec := httptest.NewRecorder()
	Handler(client)(rec, httptest.NewRequest(http.MethodGet, "/healthz", nil))

	var body Response
	if err := json.NewDecoder(rec.Body).Decode(&body); err != nil {
		t.Fatalf("decode body: %v", err)
	}
	return rec.Code, body
}

func TestHealthzEngineUp(t *testing.T) {
	code, body := get(t, dial(t, true))

	if code != http.StatusOK || body.Status != "ok" {
		t.Fatalf("got %d %q, want 200 ok", code, body.Status)
	}
	if !body.Engine.Reachable || body.Engine.Version != "9.9.9" {
		t.Fatalf("engine = %+v, want reachable with version 9.9.9", body.Engine)
	}
}

func TestHealthzEngineDown(t *testing.T) {
	code, body := get(t, dial(t, false))

	if code != http.StatusServiceUnavailable || body.Status != "degraded" {
		t.Fatalf("got %d %q, want 503 degraded", code, body.Status)
	}
	if body.Engine.Reachable || body.Engine.Error == "" {
		t.Fatalf("engine = %+v, want unreachable with an error", body.Engine)
	}
}

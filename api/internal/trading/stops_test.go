package trading

import (
	"context"
	"io"
	"log/slog"
	"net"
	"sync/atomic"
	"testing"
	"time"

	enginev1 "github.com/batuhangozegu/metron/api/internal/gen/metron/engine/v1"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/status"
	"google.golang.org/grpc/test/bufconn"
)

type stopServer struct {
	enginev1.UnimplementedEngineServiceServer
	calls atomic.Int32
	hit   *enginev1.StopHit
}

func (s *stopServer) WatchStops(_ *enginev1.WatchStopsRequest, stream grpc.ServerStreamingServer[enginev1.StopHit]) error {
	if s.calls.Add(1) == 1 {
		return status.Error(codes.Unavailable, "transient engine outage")
	}
	for range 2 {
		if err := stream.Send(s.hit); err != nil {
			return err
		}
	}
	<-stream.Context().Done()
	return stream.Context().Err()
}

func TestStopConsumerReconnectsAndRetriesSettlement(t *testing.T) {
	s, store, _, _, _ := fixture(t)
	opened := mustOpen(t, s)
	store.failBegin = 1
	store.onCommit = make(chan struct{}, 2)
	listener := bufconn.Listen(1024 * 1024)
	server := grpc.NewServer()
	producer := &stopServer{hit: &enginev1.StopHit{PositionId: opened.Position.ID, Symbol: "BTCUSDT", Stop: "95", Price: "94"}}
	enginev1.RegisterEngineServiceServer(server, producer)
	go func() { _ = server.Serve(listener) }()
	defer server.Stop()
	defer listener.Close()
	conn, err := grpc.NewClient("passthrough:///stops", grpc.WithTransportCredentials(insecure.NewCredentials()), grpc.WithContextDialer(func(ctx context.Context, _ string) (net.Conn, error) { return listener.DialContext(ctx) }))
	if err != nil {
		t.Fatal(err)
	}
	defer conn.Close()
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	done := make(chan struct{})
	go func() {
		defer close(done)
		s.WatchStops(ctx, enginev1.NewEngineServiceClient(conn), slog.New(slog.NewTextHandler(io.Discard, nil)))
	}()
	select {
	case <-store.onCommit:
	case <-ctx.Done():
		t.Fatal("stop did not settle after stream/database recovery")
	}
	cancel()
	select {
	case <-done:
	case <-time.After(time.Second):
		t.Fatal("consumer did not shut down")
	}
	if producer.calls.Load() < 2 || store.commits != 2 || len(store.orders) != 2 || *store.positions[opened.Position.ID].RealizedPNL != "-10.39" {
		t.Fatalf("unexpected stop settlement: calls=%d store=%+v", producer.calls.Load(), store)
	}
}

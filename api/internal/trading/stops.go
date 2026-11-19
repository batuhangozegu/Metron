package trading

import (
	"context"
	"log/slog"
	"time"

	enginev1 "github.com/batuhangozegu/metron/api/internal/gen/metron/engine/v1"
	"google.golang.org/grpc"
)

type StopSource interface {
	WatchStops(context.Context, *enginev1.WatchStopsRequest, ...grpc.CallOption) (grpc.ServerStreamingClient[enginev1.StopHit], error)
}

func pause(ctx context.Context, delay time.Duration) bool {
	timer := time.NewTimer(delay)
	defer timer.Stop()
	select {
	case <-ctx.Done():
		return false
	case <-timer.C:
		return true
	}
}

// WatchStops reconnects on stream failures. A received hit is retried until it
// settles, so a transient database outage cannot silently discard the close.
func (s Service) WatchStops(ctx context.Context, engine StopSource, log *slog.Logger) {
	backoff := 500 * time.Millisecond
	for ctx.Err() == nil {
		stream, err := engine.WatchStops(ctx, &enginev1.WatchStopsRequest{})
		if err == nil {
			for ctx.Err() == nil {
				hit, recvErr := stream.Recv()
				if recvErr != nil {
					err = recvErr
					break
				}
				retry := 500 * time.Millisecond
				for ctx.Err() == nil {
					attempt, cancel := context.WithTimeout(ctx, 4*time.Second)
					closeErr := s.Stop(attempt, hit)
					cancel()
					if closeErr == nil {
						break
					}
					log.Warn("stop settlement failed; retrying", "err", closeErr)
					if !pause(ctx, retry) {
						return
					}
					retry = min(retry*2, 30*time.Second)
				}
				backoff = 500 * time.Millisecond
			}
		}
		if ctx.Err() != nil {
			return
		}
		log.Warn("stop stream interrupted; reconnecting", "err", err, "backoff", backoff)
		if !pause(ctx, backoff) {
			return
		}
		backoff = min(backoff*2, 30*time.Second)
	}
}

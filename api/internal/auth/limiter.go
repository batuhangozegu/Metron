package auth

import (
	"sync"
	"time"
)

type bucket struct {
	start time.Time
	count int
}

// Limiter is local to one API process. It ignores forwarded headers: RemoteAddr
// must be the client address, or a trusted proxy must supply that separately.
type Limiter struct {
	mu      sync.Mutex
	entries map[string]bucket
}

func (l *Limiter) Allow(now time.Time, keys ...string) bool {
	l.mu.Lock()
	defer l.mu.Unlock()
	if l.entries == nil {
		l.entries = make(map[string]bucket)
	}
	for key, b := range l.entries {
		if !now.Before(b.start.Add(15 * time.Minute)) {
			delete(l.entries, key)
		}
	}
	allowed := true
	for _, key := range keys {
		b, ok := l.entries[key]
		if !ok {
			if len(l.entries) >= 10000 {
				allowed = false
				continue
			}
			b.start = now
		}
		if b.count >= 10 {
			allowed = false
		} else {
			b.count++
		}
		l.entries[key] = b
	}
	return allowed
}

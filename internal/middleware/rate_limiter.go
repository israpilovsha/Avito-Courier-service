package middleware

import (
	"net/http"
	"sync"
	"time"

	"github.com/Avito-courses/course-go-avito-israpilovsha/internal/metrics"
	"go.uber.org/zap"
)

type TokenBucket struct {
	mu         sync.Mutex
	tokens     float64
	capacity   float64
	refillRate float64
	lastRefill time.Time
}

func NewTokenBucket(rps int) *TokenBucket {
	now := time.Now()
	return &TokenBucket{
		tokens:     float64(rps),
		capacity:   float64(rps),
		refillRate: float64(rps),
		lastRefill: now,
	}
}

func (b *TokenBucket) Allow() bool {
	b.mu.Lock()
	defer b.mu.Unlock()

	now := time.Now()
	elapsed := now.Sub(b.lastRefill).Seconds()
	if elapsed > 0 {
		b.tokens += elapsed * b.refillRate
		if b.tokens > b.capacity {
			b.tokens = b.capacity
		}
		b.lastRefill = now
	}

	if b.tokens >= 1 {
		b.tokens -= 1
		return true
	}

	return false
}

func RateLimit(bucket *TokenBucket, log *zap.SugaredLogger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if bucket.Allow() {
				next.ServeHTTP(w, r)
				return
			}

			metrics.RateLimitExceededTotal.Inc()
			log.Warnw("rate limit exceeded",
				"method", r.Method,
				"path", r.URL.Path,
				"remote_addr", r.RemoteAddr,
			)

			w.Header().Set("Retry-After", "1")
			http.Error(w, http.StatusText(http.StatusTooManyRequests), http.StatusTooManyRequests)
		})
	}
}

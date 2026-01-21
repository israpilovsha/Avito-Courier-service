package middleware

import (
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Avito-courses/course-go-avito-israpilovsha/internal/metrics"
	"github.com/prometheus/client_golang/prometheus/testutil"
	"go.uber.org/zap"
)

func TestRateLimit_AllowsFirstRequest_ThenRejectsSecond_AndIncrementsMetric(t *testing.T) {
	log := zap.NewNop().Sugar()
	bucket := NewTokenBucket(1)

	before := testutil.ToFloat64(metrics.RateLimitExceededTotal)

	h := RateLimit(bucket, log)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
	}))

	srv := httptest.NewServer(h)
	defer srv.Close()

	resp1, err := http.Get(srv.URL + "/test")
	if err != nil {
		t.Fatalf("request #1 failed: %v", err)
	}
	_, _ = io.Copy(io.Discard, resp1.Body)
	_ = resp1.Body.Close()
	if resp1.StatusCode != http.StatusOK {
		t.Fatalf("expected 200 for first request, got %d", resp1.StatusCode)
	}

	resp2, err := http.Get(srv.URL + "/test")
	if err != nil {
		t.Fatalf("request #2 failed: %v", err)
	}
	_, _ = io.Copy(io.Discard, resp2.Body)
	_ = resp2.Body.Close()

	if resp2.StatusCode != http.StatusTooManyRequests {
		t.Fatalf("expected 429 for second request, got %d", resp2.StatusCode)
	}
	if ra := resp2.Header.Get("Retry-After"); ra != "1" {
		t.Fatalf("expected Retry-After=1, got %q", ra)
	}

	after := testutil.ToFloat64(metrics.RateLimitExceededTotal)
	if after-before != 1 {
		t.Fatalf("expected metric delta=1, got %v", after-before)
	}
}

func TestRateLimit_DoesNotIncrementMetric_WhenUnderLimit(t *testing.T) {
	log := zap.NewNop().Sugar()
	bucket := NewTokenBucket(5)

	before := testutil.ToFloat64(metrics.RateLimitExceededTotal)

	h := RateLimit(bucket, log)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	srv := httptest.NewServer(h)
	defer srv.Close()

	resp, err := http.Get(srv.URL + "/ok")
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	_, _ = io.Copy(io.Discard, resp.Body)
	_ = resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}

	after := testutil.ToFloat64(metrics.RateLimitExceededTotal)
	if after-before != 0 {
		t.Fatalf("expected metric delta=0, got %v", after-before)
	}
}

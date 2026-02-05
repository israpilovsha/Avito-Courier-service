package order

import (
	"context"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
	"time"

	"github.com/Avito-courses/course-go-avito-israpilovsha/internal/metrics"
	"github.com/prometheus/client_golang/prometheus/testutil"
)

type testLogger struct {
	calls int64
}

func (l *testLogger) Warnw(msg string, keysAndValues ...any) {
	atomic.AddInt64(&l.calls, 1)
}

func TestGateway_RetryOn429_IncrementsMetricAndEventuallySucceeds(t *testing.T) {

	var hits int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(&hits, 1)

		if atomic.LoadInt32(&hits) <= 2 {
			w.WriteHeader(http.StatusTooManyRequests)
			_, _ = w.Write([]byte(`{"error":"too many requests"}`))
			return
		}

		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"order_id":"x","status":"created"}`))
	}))
	defer srv.Close()

	gw := NewHTTPGateway(srv.URL)

	hg := gw.(*httpGateway)
	hg.maxRetries = 5
	hg.baseDelay = 1 * time.Millisecond
	log := &testLogger{}
	WithLogger(gw, log)

	ctx := context.Background()

	status, err := gw.GetStatus(ctx, "x")
	if err != nil {
		t.Fatalf("expected success, got err: %v", err)
	}
	if status != "created" {
		t.Fatalf("expected status=created, got %s", status)
	}

	retries := testutil.ToFloat64(metrics.GatewayRetriesTotal)
	if retries < 2 {
		t.Fatalf("expected retries >= 2, got %v", retries)
	}
	if atomic.LoadInt64(&log.calls) < 2 {
		t.Fatalf("expected warn logs >=2, got %d", log.calls)
	}
}

func TestGateway_NoRetryOn400(t *testing.T) {
	before := testutil.ToFloat64(metrics.GatewayRetriesTotal)

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte(`{"error":"bad request"}`))
	}))
	defer srv.Close()

	gw := NewHTTPGateway(srv.URL)
	hg := gw.(*httpGateway)
	hg.maxRetries = 5
	hg.baseDelay = 1 * time.Millisecond

	_, err := gw.GetStatus(context.Background(), "x")
	if err == nil {
		t.Fatal("expected error, got nil")
	}

	after := testutil.ToFloat64(metrics.GatewayRetriesTotal)
	if after-before != 0 {
		t.Fatalf("expected retries delta=0, got %v", after-before)
	}
}

func TestGateway_RetryOnNetworkError(t *testing.T) {

	gw := NewHTTPGateway("http://127.0.0.1:1")
	hg := gw.(*httpGateway)
	hg.maxRetries = 3
	hg.baseDelay = 1 * time.Millisecond

	_, err := gw.GetStatus(context.Background(), "x")
	if err == nil {
		t.Fatal("expected error, got nil")
	}

	retries := testutil.ToFloat64(metrics.GatewayRetriesTotal)
	if retries <= 0 {
		t.Fatalf("expected retries > 0, got %v (err=%v)", retries, err)
	}
}

func TestGateway_FetchOrders_RetryAndDecode(t *testing.T) {

	var hits int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/public/api/v1/orders" {
			w.WriteHeader(http.StatusNotFound)
			return
		}

		atomic.AddInt32(&hits, 1)
		if atomic.LoadInt32(&hits) == 1 {
			w.WriteHeader(http.StatusServiceUnavailable)
			_, _ = w.Write([]byte(`{"error":"temporary"}`))
			return
		}

		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`[
			{"id":"o1","status":"created","created_at":"2026-01-01T00:00:00Z"},
			{"id":"o2","status":"pending","created_at":"2026-01-01T00:00:00Z"}
		]`))
	}))
	defer srv.Close()

	gw := NewHTTPGateway(srv.URL)
	hg := gw.(*httpGateway)
	hg.maxRetries = 5
	hg.baseDelay = 1 * time.Millisecond

	orders, err := gw.FetchOrders(context.Background(), time.Now())
	if err != nil {
		t.Fatalf("expected success, got err: %v", err)
	}
	if len(orders) != 2 {
		t.Fatalf("expected 2 orders, got %d", len(orders))
	}

	retries := testutil.ToFloat64(metrics.GatewayRetriesTotal)
	if retries < 1 {
		t.Fatalf("expected retries >=1, got %v", retries)
	}
}

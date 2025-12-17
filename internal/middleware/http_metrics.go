package middleware

import (
	"net/http"
	"time"

	"go.uber.org/zap"

	"github.com/Avito-courses/course-go-avito-israpilovsha/internal/metrics"
)

type responseWriter struct {
	http.ResponseWriter
	status int
}

func (w *responseWriter) WriteHeader(code int) {
	w.status = code
	w.ResponseWriter.WriteHeader(code)
}

func MetricsAndLogging(log *zap.SugaredLogger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()

			rw := &responseWriter{ResponseWriter: w, status: 200}
			next.ServeHTTP(rw, r)

			duration := time.Since(start)
			status := rw.status

			metrics.HttpRequestsTotal.WithLabelValues(
				r.Method,
				r.URL.Path,
				http.StatusText(status),
			).Inc()

			metrics.HttpRequestDuration.WithLabelValues(
				r.Method,
				r.URL.Path,
				http.StatusText(status),
			).Observe(duration.Seconds())

			log.Infow("http request",
				"timestamp", time.Now().Format(time.RFC3339),
				"method", r.Method,
				"path", r.URL.Path,
				"status", status,
				"duration", duration.String(),
			)
		})
	}
}

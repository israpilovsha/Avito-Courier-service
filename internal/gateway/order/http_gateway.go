package order

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"os"
	"syscall"
	"time"

	"github.com/Avito-courses/course-go-avito-israpilovsha/internal/metrics"
)

var errRetryableHTTP = errors.New("retryable http status")

type Logger interface {
	Warnw(msg string, keysAndValues ...any)
}

type noopLogger struct{}

func (noopLogger) Warnw(string, ...any) {}

type httpGateway struct {
	baseURL string
	client  *http.Client
	log     Logger

	maxRetries int
	baseDelay  time.Duration
}

func NewHTTPGateway(baseURL string) Gateway {
	return &httpGateway{
		baseURL:    baseURL,
		client:     &http.Client{Timeout: 5 * time.Second},
		log:        noopLogger{},
		maxRetries: 4,
		baseDelay:  100 * time.Millisecond,
	}
}

func WithLogger(g Gateway, log Logger) Gateway {
	hg, ok := g.(*httpGateway)
	if !ok {
		return g
	}
	hg.log = log
	return hg
}

func (g *httpGateway) FetchOrders(ctx context.Context, from time.Time) ([]Order, error) {
	u, err := url.Parse(g.baseURL)
	if err != nil {
		return nil, err
	}

	u.Path = "/public/api/v1/orders"
	q := u.Query()
	q.Set("from", from.UTC().Format(time.RFC3339Nano))
	u.RawQuery = q.Encode()

	var orders []Order

	err = g.doWithRetry(ctx, func() (*http.Response, error) {
		req, err := http.NewRequestWithContext(ctx, http.MethodGet, u.String(), nil)
		if err != nil {
			return nil, err
		}
		req.Header.Set("Accept", "application/json")
		req.Header.Set("X-Bypass-Auth", "true")
		return g.client.Do(req)
	}, func(resp *http.Response) error {
		defer func() { _ = resp.Body.Close() }()

		if resp.StatusCode == http.StatusTooManyRequests || resp.StatusCode >= 500 {
			return fmt.Errorf("%w: status %s", errRetryableHTTP, resp.Status)
		}

		if resp.StatusCode < 200 || resp.StatusCode >= 300 {
			return fmt.Errorf("order gateway: status %s", resp.Status)
		}

		return json.NewDecoder(resp.Body).Decode(&orders)
	})

	if err != nil {
		return nil, err
	}

	return orders, nil
}

func (g *httpGateway) GetStatus(ctx context.Context, id string) (string, error) {
	u, err := url.Parse(g.baseURL)
	g.log.Warnw("gateway GetStatus called", "order_id", id, "url", u.String())
	if err != nil {
		return "", err
	}

	u.Path = fmt.Sprintf("/public/api/v1/order/%s/status", id)

	var status string

	err = g.doWithRetry(ctx, func() (*http.Response, error) {
		req, err := http.NewRequestWithContext(ctx, http.MethodGet, u.String(), nil)
		if err != nil {
			return nil, err
		}
		req.Header.Set("Accept", "application/json")
		req.Header.Set("X-Bypass-Auth", "true")
		return g.client.Do(req)
	}, func(resp *http.Response) error {
		defer func() { _ = resp.Body.Close() }()

		if resp.StatusCode == http.StatusTooManyRequests || resp.StatusCode >= 500 {
			return fmt.Errorf("%w: status %s", errRetryableHTTP, resp.Status)
		}

		if resp.StatusCode != http.StatusOK {
			return fmt.Errorf("order gateway: status %s", resp.Status)
		}

		var res struct {
			OrderID string `json:"order_id"`
			Status  string `json:"status"`
		}
		if err := json.NewDecoder(resp.Body).Decode(&res); err != nil {
			return err
		}
		status = res.Status
		return nil
	})

	if err != nil {
		return "", err
	}

	return status, nil
}

func (g *httpGateway) doWithRetry(
	ctx context.Context,
	do func() (*http.Response, error),
	handle func(resp *http.Response) error,
) error {
	delay := g.baseDelay

	var lastErr error

	for attempt := 0; attempt < g.maxRetries; attempt++ {
		resp, err := do()
		if err == nil {
			if hErr := handle(resp); hErr == nil {
				return nil
			} else {
				lastErr = hErr
				if !errors.Is(hErr, errRetryableHTTP) {
					return hErr
				}
			}
		} else {
			lastErr = err
			if !isRetryableNetErr(err) {
				return err
			}
		}

		metrics.GatewayRetriesTotal.Inc()
		g.log.Warnw("gateway retry",
			"attempt", attempt+1,
			"max", g.maxRetries,
			"err", lastErr,
			"delay", delay.String(),
		)

		timer := time.NewTimer(delay)
		select {
		case <-ctx.Done():
			timer.Stop()
			return ctx.Err()
		case <-timer.C:
		}

		delay *= 2
	}

	return lastErr
}

func isRetryableNetErr(err error) bool {
	if err == nil {
		return false
	}

	if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
		return false
	}

	var nerr net.Error
	if errors.As(err, &nerr) {
		if nerr.Timeout() {
			return true
		}
	}

	if errors.Is(err, io.EOF) {
		return true
	}

	if errors.Is(err, syscall.ECONNREFUSED) ||
		errors.Is(err, syscall.ECONNRESET) ||
		errors.Is(err, syscall.EPIPE) ||
		errors.Is(err, syscall.ETIMEDOUT) {
		return true
	}

	var se *os.SyscallError
	if errors.As(err, &se) {
		if errors.Is(se, syscall.ECONNREFUSED) ||
			errors.Is(se, syscall.ECONNRESET) ||
			errors.Is(se, syscall.EPIPE) ||
			errors.Is(se, syscall.ETIMEDOUT) {
			return true
		}
	}

	return false
}

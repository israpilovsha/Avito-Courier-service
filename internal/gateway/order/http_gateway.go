package order

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"time"
)

type httpGateway struct {
	baseURL string
	client  *http.Client
}

func NewHTTPGateway(baseURL string) Gateway {
	return &httpGateway{
		baseURL: baseURL,
		client:  &http.Client{Timeout: 5 * time.Second},
	}
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

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u.String(), nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Accept", "application/json")

	resp, err := g.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("order gateway: status %s", resp.Status)
	}

	var orders []Order
	if err := json.NewDecoder(resp.Body).Decode(&orders); err != nil {
		return nil, err
	}

	return orders, nil
}

func (g *httpGateway) GetStatus(ctx context.Context, id string) (string, error) {
	u, err := url.Parse(g.baseURL)
	if err != nil {
		return "", err
	}

	u.Path = fmt.Sprintf("/public/api/v1/orders/%s/status", id)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u.String(), nil)
	if err != nil {
		return "", err
	}

	resp, err := g.client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("order gateway: status %s", resp.Status)
	}

	var res struct {
		Status string `json:"status"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&res); err != nil {
		return "", err
	}

	return res.Status, nil
}

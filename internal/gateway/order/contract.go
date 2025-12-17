package order

import (
	"context"
	"time"
)

type Order struct {
	ID        string    `json:"id"`
	CreatedAt time.Time `json:"created_at"`
}

type Gateway interface {
	FetchOrders(ctx context.Context, from time.Time) ([]Order, error)
}

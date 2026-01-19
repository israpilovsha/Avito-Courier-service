package model

import "strings"

type OrderStatus string

const (
	OrderStatusCreated   OrderStatus = "created"
	OrderStatusCancelled OrderStatus = "cancelled"
	OrderStatusCompleted OrderStatus = "completed"
)

func ParseOrderStatus(raw string) OrderStatus {
	switch strings.ToLower(raw) {
	case "created":
		return OrderStatusCreated
	case "cancelled", "canceled":
		return OrderStatusCancelled
	case "completed":
		return OrderStatusCompleted
	default:
		return OrderStatus(raw)
	}
}

func (s OrderStatus) IsTerminal() bool {
	return s == OrderStatusCancelled || s == OrderStatusCompleted
}

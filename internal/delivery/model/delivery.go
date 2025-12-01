package model

import "time"

type Delivery struct {
	ID         int64     `json:"id"`
	CourierID  int64     `json:"courier_id"`
	OrderID    string    `json:"order_id"`
	AssignedAt time.Time `json:"assigned_at"`
	Deadline   time.Time `json:"deadline"`
}

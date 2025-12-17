package worker

type OrderEvent struct {
	OrderID string `json:"order_id"`
	Status  string `json:"status"`
}

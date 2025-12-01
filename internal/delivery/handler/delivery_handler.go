package handler

import (
	"context"
	"encoding/json"
	"net/http"

	courierModel "github.com/Avito-courses/course-go-avito-israpilovsha/internal/courier/model"
	deliveryModel "github.com/Avito-courses/course-go-avito-israpilovsha/internal/delivery/model"
	"go.uber.org/zap"
)

type DeliveryService interface {
	Assign(ctx context.Context, orderID string) (*deliveryModel.Delivery, *courierModel.Courier, error)
	Unassign(ctx context.Context, orderID string) (*deliveryModel.Delivery, error)
}

type Handler struct {
	svc DeliveryService
	log *zap.SugaredLogger
}

func NewHandler(s DeliveryService, log *zap.SugaredLogger) *Handler {
	return &Handler{svc: s, log: log}
}

type assignReq struct {
	OrderID string `json:"order_id"`
}

type unassignReq struct {
	OrderID string `json:"order_id"`
}

func respond(w http.ResponseWriter, code int, body any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	if body != nil {
		_ = json.NewEncoder(w).Encode(body)
	}
}

func (h *Handler) Assign(w http.ResponseWriter, r *http.Request) {
	var req assignReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.OrderID == "" {
		respond(w, http.StatusBadRequest, map[string]string{"error": "invalid request body"})
		return
	}

	delivery, courier, err := h.svc.Assign(r.Context(), req.OrderID)
	if err != nil {
		h.log.Warnf("Assign failed: %v", err)
		respond(w, http.StatusConflict, map[string]string{"error": err.Error()})
		return
	}

	resp := map[string]any{
		"courier_id":        courier.ID,
		"order_id":          delivery.OrderID,
		"transport_type":    courier.TransportType,
		"delivery_deadline": delivery.Deadline,
	}

	respond(w, http.StatusOK, resp)
}

func (h *Handler) Unassign(w http.ResponseWriter, r *http.Request) {
	var req unassignReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.OrderID == "" {
		respond(w, http.StatusBadRequest, map[string]string{"error": "invalid request body"})
		return
	}

	delivery, err := h.svc.Unassign(r.Context(), req.OrderID)
	if err != nil {
		h.log.Warnf("Unassign failed: %v", err)
		respond(w, http.StatusNotFound, map[string]string{"error": err.Error()})
		return
	}

	resp := map[string]any{
		"order_id":   req.OrderID,
		"status":     "unassigned",
		"courier_id": delivery.CourierID,
	}

	respond(w, http.StatusOK, resp)
}

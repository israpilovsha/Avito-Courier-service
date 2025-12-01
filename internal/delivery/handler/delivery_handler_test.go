package handler_test

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	courierModel "github.com/Avito-courses/course-go-avito-israpilovsha/internal/courier/model"
	"github.com/Avito-courses/course-go-avito-israpilovsha/internal/delivery/handler"
	deliveryModel "github.com/Avito-courses/course-go-avito-israpilovsha/internal/delivery/model"

	"go.uber.org/zap"
)

type mockDeliveryService struct {
	AssignFn   func(orderID string) (*deliveryModel.Delivery, *courierModel.Courier, error)
	UnassignFn func(orderID string) (*deliveryModel.Delivery, error)
}

func (m *mockDeliveryService) Assign(_ context.Context, orderID string) (*deliveryModel.Delivery, *courierModel.Courier, error) {
	return m.AssignFn(orderID)
}
func (m *mockDeliveryService) Unassign(_ context.Context, orderID string) (*deliveryModel.Delivery, error) {
	return m.UnassignFn(orderID)
}

// TestAssignHandlerSuccess - успешное назначение курьера
func TestAssignHandlerSuccess(t *testing.T) {
	t.Parallel()
	svc := &mockDeliveryService{
		AssignFn: func(orderID string) (*deliveryModel.Delivery, *courierModel.Courier, error) {
			return &deliveryModel.Delivery{
					ID:        1,
					OrderID:   orderID,
					CourierID: 10,
				},
				&courierModel.Courier{ID: 10, TransportType: "scooter"},
				nil
		},
	}

	log := zap.NewExample().Sugar()
	h := handler.NewHandler(svc, log)

	body := []byte(`{"order_id":"abc"}`)
	req := httptest.NewRequest(http.MethodPost, "/delivery/assign", bytes.NewBuffer(body))
	w := httptest.NewRecorder()

	h.Assign(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}

	var resp map[string]interface{}
	_ = json.Unmarshal(w.Body.Bytes(), &resp)

	if resp["courier_id"] != float64(10) {
		t.Fatalf("expected courier 10, got %v", resp["courier_id"])
	}
}

// TestAssignHandlerBadRequest - неправильный JSON
func TestAssignHandlerBadRequest(t *testing.T) {
	t.Parallel()
	svc := &mockDeliveryService{}

	log := zap.NewExample().Sugar()
	h := handler.NewHandler(svc, log)

	req := httptest.NewRequest("POST", "/delivery/assign", bytes.NewBuffer([]byte(`{bad json}`)))
	w := httptest.NewRecorder()

	h.Assign(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", w.Code)
	}
}

// TestAssignHandlerConflict -
func TestAssignHandlerConflict(t *testing.T) {
	t.Parallel()
	svc := &mockDeliveryService{
		AssignFn: func(orderID string) (*deliveryModel.Delivery, *courierModel.Courier, error) {
			return nil, nil, errors.New("no couriers")
		},
	}

	log := zap.NewExample().Sugar()
	h := handler.NewHandler(svc, log)

	req := httptest.NewRequest("POST", "/delivery/assign", bytes.NewBuffer([]byte(`{"order_id":"x"}`)))
	w := httptest.NewRecorder()

	h.Assign(w, req)

	if w.Code != http.StatusConflict {
		t.Fatalf("expected 409, got %d", w.Code)
	}
}

// - TestUnassignHandlerSuccess - успешное снятие курьера
func TestUnassignHandlerSuccess(t *testing.T) {
	t.Parallel()

	svc := &mockDeliveryService{
		UnassignFn: func(orderID string) (*deliveryModel.Delivery, error) {
			return &deliveryModel.Delivery{
				ID:        1,
				OrderID:   orderID,
				CourierID: 42,
			}, nil
		},
	}

	log := zap.NewExample().Sugar()
	h := handler.NewHandler(svc, log)

	req := httptest.NewRequest(http.MethodPost, "/delivery/unassign",
		bytes.NewBuffer([]byte(`{"order_id":"xyz"}`)))
	w := httptest.NewRecorder()

	h.Unassign(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}

	var resp map[string]any
	_ = json.Unmarshal(w.Body.Bytes(), &resp)

	if resp["courier_id"] != float64(42) {
		t.Fatalf("expected courier 42, got %v", resp["courier_id"])
	}
	if resp["status"] != "unassigned" {
		t.Fatalf("expected status unassigned, got %v", resp["status"])
	}
}

// TestUnassignHandlerBadRequest - некорректный JSON
func TestUnassignHandlerBadRequest(t *testing.T) {
	t.Parallel()

	svc := &mockDeliveryService{}
	log := zap.NewExample().Sugar()
	h := handler.NewHandler(svc, log)

	req := httptest.NewRequest(http.MethodPost, "/delivery/unassign",
		bytes.NewBuffer([]byte(`{invalid json}`)))
	w := httptest.NewRecorder()

	h.Unassign(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", w.Code)
	}
}

// TestUnassignHandlerNotFound - запись не найдена
func TestUnassignHandlerNotFound(t *testing.T) {
	t.Parallel()

	svc := &mockDeliveryService{
		UnassignFn: func(orderID string) (*deliveryModel.Delivery, error) {
			return nil, errors.New("not found")
		},
	}

	log := zap.NewExample().Sugar()
	h := handler.NewHandler(svc, log)

	req := httptest.NewRequest(http.MethodPost, "/delivery/unassign",
		bytes.NewBuffer([]byte(`{"order_id":"123"}`)))
	w := httptest.NewRecorder()

	h.Unassign(w, req)

	if w.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", w.Code)
	}
}

// TestAssignHandlerMissingField - поле order_id отсутствует
func TestAssignHandlerMissingField(t *testing.T) {
	t.Parallel()

	svc := &mockDeliveryService{}
	log := zap.NewExample().Sugar()
	h := handler.NewHandler(svc, log)

	req := httptest.NewRequest(http.MethodPost, "/delivery/assign",
		bytes.NewBuffer([]byte(`{"wrong":"field"}`)))
	w := httptest.NewRecorder()

	h.Assign(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", w.Code)
	}
}

// TestAssignHandlerEmptyBody - пустое тело запроса
func TestAssignHandlerEmptyBody(t *testing.T) {
	t.Parallel()

	svc := &mockDeliveryService{}
	log := zap.NewExample().Sugar()
	h := handler.NewHandler(svc, log)

	req := httptest.NewRequest(http.MethodPost, "/delivery/assign", nil)
	w := httptest.NewRecorder()

	h.Assign(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", w.Code)
	}
}

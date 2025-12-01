package handler_test

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Avito-courses/course-go-avito-israpilovsha/internal/courier/handler"
	"github.com/Avito-courses/course-go-avito-israpilovsha/internal/courier/model"
	"github.com/Avito-courses/course-go-avito-israpilovsha/internal/courier/repository"
	"github.com/gorilla/mux"
	"go.uber.org/zap"
)

type mockCourierService struct {
	CreateFn  func(ctx context.Context, c *model.Courier) error
	GetByIDFn func(ctx context.Context, id int64) (*model.Courier, error)
	GetAllFn  func(ctx context.Context) ([]*model.Courier, error)
	UpdateFn  func(ctx context.Context, c *model.Courier) error
}

func (m *mockCourierService) Create(ctx context.Context, c *model.Courier) error {
	return m.CreateFn(ctx, c)
}
func (m *mockCourierService) GetByID(ctx context.Context, id int64) (*model.Courier, error) {
	return m.GetByIDFn(ctx, id)
}
func (m *mockCourierService) GetAll(ctx context.Context) ([]*model.Courier, error) {
	return m.GetAllFn(ctx)
}
func (m *mockCourierService) Update(ctx context.Context, c *model.Courier) error {
	return m.UpdateFn(ctx, c)
}

func muxSetVar(req *http.Request, key, value string) *http.Request {
	ctx := context.WithValue(req.Context(), "mux.vars", map[string]string{
		key: value,
	})
	return req.WithContext(ctx)
}

func TestGetByID_Success(t *testing.T) {
	t.Parallel()
	svc := &mockCourierService{
		GetByIDFn: func(ctx context.Context, id int64) (*model.Courier, error) {
			return &model.Courier{ID: id, Name: "Test"}, nil
		},
	}

	log := zap.NewExample().Sugar()
	h := handler.NewHandler(svc, log)

	req := httptest.NewRequest("GET", "/courier/5", nil)
	req = mux.SetURLVars(req, map[string]string{"id": "5"}) // ВАЖНО!!!

	w := httptest.NewRecorder()
	h.GetByID(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}

	var resp model.Courier
	_ = json.Unmarshal(w.Body.Bytes(), &resp)

	if resp.ID != 5 {
		t.Fatalf("expected ID=5, got %d", resp.ID)
	}
}

func TestGetByID_NotFound(t *testing.T) {
	t.Parallel()
	svc := &mockCourierService{
		GetByIDFn: func(ctx context.Context, id int64) (*model.Courier, error) {
			return nil, errors.New("not found")
		},
	}

	log := zap.NewExample().Sugar()
	h := handler.NewHandler(svc, log)

	req := httptest.NewRequest("GET", "/courier/10", nil)
	req = mux.SetURLVars(req, map[string]string{"id": "10"})

	w := httptest.NewRecorder()
	h.GetByID(w, req)

	if w.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", w.Code)
	}
}

func TestGetAll_Success(t *testing.T) {
	t.Parallel()
	svc := &mockCourierService{
		GetAllFn: func(ctx context.Context) ([]*model.Courier, error) {
			return []*model.Courier{
				{ID: 1, Name: "A"},
				{ID: 2, Name: "B"},
			}, nil
		},
	}

	log := zap.NewExample().Sugar()
	h := handler.NewHandler(svc, log)

	req := httptest.NewRequest("GET", "/couriers", nil)
	w := httptest.NewRecorder()

	h.GetAll(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}

	var resp []model.Courier
	_ = json.Unmarshal(w.Body.Bytes(), &resp)

	if len(resp) != 2 {
		t.Fatalf("expected 2 couriers, got %d", len(resp))
	}
}

func TestCreate_Success(t *testing.T) {
	t.Parallel()
	svc := &mockCourierService{
		CreateFn: func(ctx context.Context, c *model.Courier) error {
			c.ID = 100
			return nil
		},
	}

	log := zap.NewExample().Sugar()
	h := handler.NewHandler(svc, log)

	body := []byte(`{"name":"Ivan","phone":"123","status":"available","transport_type":"car"}`)

	req := httptest.NewRequest("POST", "/courier", bytes.NewBuffer(body))
	w := httptest.NewRecorder()

	h.Create(w, req)

	if w.Code != http.StatusCreated {
		t.Fatalf("expected 200, got %d", w.Code)
	}

	var resp model.Courier
	_ = json.Unmarshal(w.Body.Bytes(), &resp)

	if resp.ID != 100 {
		t.Fatalf("expected ID=100, got %d", resp.ID)
	}
}

func TestCreate_BadRequest(t *testing.T) {
	t.Parallel()
	svc := &mockCourierService{}

	log := zap.NewExample().Sugar()
	h := handler.NewHandler(svc, log)

	req := httptest.NewRequest("POST", "/courier", bytes.NewBuffer([]byte(`{bad json}`)))
	w := httptest.NewRecorder()

	h.Create(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", w.Code)
	}
}

func TestUpdate_Success(t *testing.T) {
	t.Parallel()
	svc := &mockCourierService{
		UpdateFn: func(ctx context.Context, c *model.Courier) error { return nil },
	}

	log := zap.NewExample().Sugar()
	h := handler.NewHandler(svc, log)

	body := []byte(`{"id":7,"name":"Updated"}`)

	req := httptest.NewRequest("PUT", "/courier", bytes.NewBuffer(body))
	w := httptest.NewRecorder()

	h.Update(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
}

func TestUpdate_NotFound(t *testing.T) {
	t.Parallel()
	svc := &mockCourierService{
		UpdateFn: func(ctx context.Context, c *model.Courier) error {
			return errors.New("not found")
		},
	}

	log := zap.NewExample().Sugar()
	h := handler.NewHandler(svc, log)

	body := []byte(`{"id":7}`)

	req := httptest.NewRequest("PUT", "/courier", bytes.NewBuffer(body))
	w := httptest.NewRecorder()

	h.Update(w, req)

	if w.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", w.Code)
	}
}

func TestPing(t *testing.T) {
	t.Parallel()
	svc := &mockCourierService{}
	log := zap.NewExample().Sugar()
	h := handler.NewHandler(svc, log)

	req := httptest.NewRequest("GET", "/ping", nil)
	w := httptest.NewRecorder()

	h.Ping(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
}

func TestHealthCheck(t *testing.T) {
	t.Parallel()
	svc := &mockCourierService{}
	log := zap.NewExample().Sugar()
	h := handler.NewHandler(svc, log)

	req := httptest.NewRequest("HEAD", "/healthcheck", nil)
	w := httptest.NewRecorder()

	h.HealthCheck(w, req)

	if w.Code != http.StatusNoContent {
		t.Fatalf("expected 204, got %d", w.Code)
	}
}

func TestGetByID_InvalidID(t *testing.T) {
	t.Parallel()

	svc := &mockCourierService{}
	log := zap.NewExample().Sugar()
	h := handler.NewHandler(svc, log)

	req := httptest.NewRequest("GET", "/courier/xxx", nil)
	req = mux.SetURLVars(req, map[string]string{"id": "xxx"})
	w := httptest.NewRecorder()

	h.GetByID(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", w.Code)
	}
}

func TestGetAll_Error(t *testing.T) {
	t.Parallel()

	svc := &mockCourierService{
		GetAllFn: func(ctx context.Context) ([]*model.Courier, error) {
			return nil, errors.New("db error")
		},
	}

	log := zap.NewExample().Sugar()
	h := handler.NewHandler(svc, log)

	req := httptest.NewRequest("GET", "/couriers", nil)
	w := httptest.NewRecorder()

	h.GetAll(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Fatalf("expected 500, got %d", w.Code)
	}
}

func TestUpdate_BadRequest(t *testing.T) {
	t.Parallel()

	svc := &mockCourierService{}
	log := zap.NewExample().Sugar()
	h := handler.NewHandler(svc, log)

	req := httptest.NewRequest("PUT", "/courier", bytes.NewBuffer([]byte(`{bad json}`)))
	w := httptest.NewRecorder()

	h.Update(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", w.Code)
	}
}

func TestCreate_Conflict(t *testing.T) {
	t.Parallel()

	svc := &mockCourierService{
		CreateFn: func(ctx context.Context, c *model.Courier) error {
			return repository.ErrConflict // ВАЖНО!!!
		},
	}

	log := zap.NewExample().Sugar()
	h := handler.NewHandler(svc, log)

	req := httptest.NewRequest("POST", "/courier",
		bytes.NewBufferString(`{"name":"A","phone":"123","status":"available","transport_type":"car"}`))
	w := httptest.NewRecorder()

	h.Create(w, req)

	if w.Code != http.StatusConflict {
		t.Fatalf("expected 409, got %d", w.Code)
	}
}

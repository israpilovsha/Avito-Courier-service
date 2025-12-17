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

func TestCourierHandler_GetByID(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		idParam    string
		prepareSvc func() *mockCourierService
		wantStatus int
	}{
		{
			name:    "success",
			idParam: "5",
			prepareSvc: func() *mockCourierService {
				return &mockCourierService{
					GetByIDFn: func(ctx context.Context, id int64) (*model.Courier, error) {
						return &model.Courier{ID: id, Name: "Test"}, nil
					},
				}
			},
			wantStatus: http.StatusOK,
		},
		{
			name:    "not found",
			idParam: "10",
			prepareSvc: func() *mockCourierService {
				return &mockCourierService{
					GetByIDFn: func(ctx context.Context, id int64) (*model.Courier, error) {
						return nil, errors.New("not found")
					},
				}
			},
			wantStatus: http.StatusNotFound,
		},
		{
			name:       "invalid id",
			idParam:    "xxx",
			prepareSvc: func() *mockCourierService { return &mockCourierService{} },
			wantStatus: http.StatusBadRequest,
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			log := zap.NewExample().Sugar()
			h := handler.NewHandler(tc.prepareSvc(), log)

			req := httptest.NewRequest("GET", "/courier/"+tc.idParam, nil)
			req = mux.SetURLVars(req, map[string]string{"id": tc.idParam})
			w := httptest.NewRecorder()

			h.GetByID(w, req)

			if w.Code != tc.wantStatus {
				t.Fatalf("expected %d, got %d", tc.wantStatus, w.Code)
			}
		})
	}
}

func TestCourierHandler_Create(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		body       string
		prepareSvc func() *mockCourierService
		wantStatus int
	}{
		{
			name: "success",
			body: `{"name":"Ivan","phone":"123","status":"available","transport_type":"car"}`,
			prepareSvc: func() *mockCourierService {
				return &mockCourierService{
					CreateFn: func(ctx context.Context, c *model.Courier) error {
						c.ID = 100
						return nil
					},
				}
			},
			wantStatus: http.StatusCreated,
		},
		{
			name:       "bad request",
			body:       `{bad json}`,
			prepareSvc: func() *mockCourierService { return &mockCourierService{} },
			wantStatus: http.StatusBadRequest,
		},
		{
			name: "conflict",
			body: `{"name":"A","phone":"123","status":"available","transport_type":"car"}`,
			prepareSvc: func() *mockCourierService {
				return &mockCourierService{
					CreateFn: func(ctx context.Context, c *model.Courier) error {
						return repository.ErrConflict
					},
				}
			},
			wantStatus: http.StatusConflict,
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			log := zap.NewExample().Sugar()
			h := handler.NewHandler(tc.prepareSvc(), log)

			req := httptest.NewRequest("POST", "/courier", bytes.NewBufferString(tc.body))
			w := httptest.NewRecorder()

			h.Create(w, req)

			if w.Code != tc.wantStatus {
				t.Fatalf("expected %d, got %d", tc.wantStatus, w.Code)
			}
		})
	}
}

func TestCourierHandler_GetAll(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		prepareSvc func() *mockCourierService
		wantStatus int
		wantLen    int
	}{
		{
			name: "success",
			prepareSvc: func() *mockCourierService {
				return &mockCourierService{
					GetAllFn: func(ctx context.Context) ([]*model.Courier, error) {
						return []*model.Courier{
							{ID: 1, Name: "A"},
							{ID: 2, Name: "B"},
						}, nil
					},
				}
			},
			wantStatus: http.StatusOK,
			wantLen:    2,
		},
		{
			name: "internal error",
			prepareSvc: func() *mockCourierService {
				return &mockCourierService{
					GetAllFn: func(ctx context.Context) ([]*model.Courier, error) {
						return nil, errors.New("db error")
					},
				}
			},
			wantStatus: http.StatusInternalServerError,
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			h := handler.NewHandler(tc.prepareSvc(), zap.NewExample().Sugar())
			req := httptest.NewRequest("GET", "/couriers", nil)
			w := httptest.NewRecorder()

			h.GetAll(w, req)

			if w.Code != tc.wantStatus {
				t.Fatalf("expected %d, got %d", tc.wantStatus, w.Code)
			}

			if tc.wantLen > 0 {
				var resp []model.Courier
				_ = json.Unmarshal(w.Body.Bytes(), &resp)
				if len(resp) != tc.wantLen {
					t.Fatalf("expected %d, got %d", tc.wantLen, len(resp))
				}
			}
		})
	}
}

func TestCourierHandler_Update(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		body       string
		prepareSvc func() *mockCourierService
		wantStatus int
	}{
		{
			name: "success",
			body: `{"id":7,"name":"Updated"}`,
			prepareSvc: func() *mockCourierService {
				return &mockCourierService{
					UpdateFn: func(ctx context.Context, c *model.Courier) error {
						return nil
					},
				}
			},
			wantStatus: http.StatusOK,
		},
		{
			name: "not found",
			body: `{"id":7}`,
			prepareSvc: func() *mockCourierService {
				return &mockCourierService{
					UpdateFn: func(ctx context.Context, c *model.Courier) error {
						return errors.New("not found")
					},
				}
			},
			wantStatus: http.StatusNotFound,
		},
		{
			name:       "bad request",
			body:       `{bad json}`,
			prepareSvc: func() *mockCourierService { return &mockCourierService{} },
			wantStatus: http.StatusBadRequest,
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			h := handler.NewHandler(tc.prepareSvc(), zap.NewExample().Sugar())
			req := httptest.NewRequest("PUT", "/courier", bytes.NewBufferString(tc.body))
			w := httptest.NewRecorder()

			h.Update(w, req)

			if w.Code != tc.wantStatus {
				t.Fatalf("expected %d, got %d", tc.wantStatus, w.Code)
			}
		})
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

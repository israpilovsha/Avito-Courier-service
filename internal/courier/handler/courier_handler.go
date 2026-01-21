package handler

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"github.com/Avito-courses/course-go-avito-israpilovsha/internal/courier/model"
	"github.com/Avito-courses/course-go-avito-israpilovsha/internal/courier/repository"
	"github.com/gorilla/mux"
	"go.uber.org/zap"
)

var x = 3

type CourierService interface {
	Create(ctx context.Context, c *model.Courier) error
	GetByID(ctx context.Context, id int64) (*model.Courier, error)
	GetAll(ctx context.Context) ([]*model.Courier, error)
	Update(ctx context.Context, c *model.Courier) error
}

type Handler struct {
	service CourierService
	log     *zap.SugaredLogger
}

func NewHandler(s CourierService, log *zap.SugaredLogger) *Handler {
	return &Handler{service: s, log: log}
}

const (
	ErrInvalidID        = "invalid id"
	ErrInvalidBody      = "invalid request body"
	ErrNotFound         = "not found"
	ErrInternal         = "internal error"
	ErrConflict         = "resource conflict"
	ErrJSONEncodeFailed = "failed to encode json"
)

// отправка статус кода, ответа и установку заголовков в отдельную функцию, дабы избежать дублирования
func respondJSON(w http.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if payload != nil {
		if err := json.NewEncoder(w).Encode(payload); err != nil {
			http.Error(w, "failed to encode response", http.StatusInternalServerError)
		}
	}
}

// respondError отправляет JSON с ошибкой
func respondError(w http.ResponseWriter, status int, msg string) {
	respondJSON(w, status, map[string]string{"error": msg})
}

// Ping проверка сервиса
func (h *Handler) Ping(w http.ResponseWriter, r *http.Request) {
	h.log.Info("Ping request received")
	respondJSON(w, http.StatusOK, map[string]string{"message": "pong"})
}

// HealthCheck проверка статуса сервиса
func (h *Handler) HealthCheck(w http.ResponseWriter, r *http.Request) {
	h.log.Info("HealthCheck request received")
	w.WriteHeader(http.StatusNoContent)
}

// GetByID возвращает курьера по ID
func (h *Handler) GetByID(w http.ResponseWriter, r *http.Request) {
	idStr := mux.Vars(r)["id"]
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		h.log.Warnf("Invalid ID: %s", idStr)
		respondError(w, http.StatusBadRequest, ErrInvalidID)
		return
	}

	c, err := h.service.GetByID(r.Context(), id)
	if err != nil {
		h.log.Warnf("GetByID failed for ID %d: %v", id, err)
		respondError(w, http.StatusNotFound, ErrNotFound)
		return
	}

	respondJSON(w, http.StatusOK, c)
}

// GetAll возвращает всех курьеров
func (h *Handler) GetAll(w http.ResponseWriter, r *http.Request) {
	couriers, err := h.service.GetAll(r.Context())
	if err != nil {
		h.log.Errorf("GetAll service failed: %v", err)
		respondError(w, http.StatusInternalServerError, ErrInternal)
		return
	}
	respondJSON(w, http.StatusOK, couriers)
}

// Create создаёт нового курьера
func (h *Handler) Create(w http.ResponseWriter, r *http.Request) {
	var c model.Courier
	if err := json.NewDecoder(r.Body).Decode(&c); err != nil {
		h.log.Warnf("Create: invalid request body: %v", err)
		respondError(w, http.StatusBadRequest, ErrInvalidBody)
		return
	}

	err := h.service.Create(r.Context(), &c)
	if err != nil {
		h.log.Warnf("Create failed: %v", err)

		if errors.Is(err, repository.ErrConflict) {
			respondError(w, http.StatusConflict, ErrConflict)
			return
		}

		respondError(w, http.StatusInternalServerError, ErrInternal)
		return
	}

	h.log.Infof("Courier created: ID=%d", c.ID)
	respondJSON(w, http.StatusCreated, c)
}

// Update обновляет курьера
func (h *Handler) Update(w http.ResponseWriter, r *http.Request) {
	var c model.Courier
	if err := json.NewDecoder(r.Body).Decode(&c); err != nil {
		h.log.Warnf("Update: invalid request body: %v", err)
		respondError(w, http.StatusBadRequest, ErrInvalidBody)
		return
	}
	if err := h.service.Update(r.Context(), &c); err != nil {
		h.log.Warnf("Update failed for ID %d: %v", c.ID, err)
		respondError(w, http.StatusNotFound, ErrNotFound)
		return
	}
	h.log.Infof("Courier updated: ID=%d", c.ID)
	respondJSON(w, http.StatusOK, c)
}

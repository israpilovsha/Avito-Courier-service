package repository

//go:generate mockgen -source=delivery_repository.go -destination=mock_delivery_repository.go -package=repository

import (
	"context"

	"github.com/Avito-courses/course-go-avito-israpilovsha/internal/delivery/model"
)

type DeliveryRepository interface {
	WithTx(ctx context.Context, fn func(txCtx context.Context) error) error

	Create(ctx context.Context, d *model.Delivery) error
	DeleteByOrderID(ctx context.Context, orderID string) (*model.Delivery, error)
	GetByOrderID(ctx context.Context, orderID string) (*model.Delivery, error)
}

var (
	ErrNotFound = errorNew("delivery not found")
)

type customError struct{ msg string }

func (e *customError) Error() string { return e.msg }

func errorNew(msg string) error { return &customError{msg} }

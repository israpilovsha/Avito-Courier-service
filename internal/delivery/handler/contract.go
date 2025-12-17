package handler

import (
	"context"

	courierModel "github.com/Avito-courses/course-go-avito-israpilovsha/internal/courier/model"
	deliveryModel "github.com/Avito-courses/course-go-avito-israpilovsha/internal/delivery/model"
)

type deliveryService interface {
	Assign(ctx context.Context, orderID string) (*deliveryModel.Delivery, *courierModel.Courier, error)
	Unassign(ctx context.Context, orderID string) (*deliveryModel.Delivery, error)
}

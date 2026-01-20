package worker

import (
	"context"
	"errors"
	"strings"

	"github.com/Avito-courses/course-go-avito-israpilovsha/internal/delivery/model"
	deliveryRepo "github.com/Avito-courses/course-go-avito-israpilovsha/internal/delivery/repository"
	deliveryUsecase "github.com/Avito-courses/course-go-avito-israpilovsha/internal/delivery/usecase"
	"github.com/Avito-courses/course-go-avito-israpilovsha/internal/gateway/order"
)

type OrderEventProcessor struct {
	assign   *deliveryUsecase.DeliveryService
	unassign *deliveryUsecase.DeliveryService
	complete *deliveryUsecase.CompleteService
	orders   order.Gateway
}

func NewOrderEventProcessor(
	assign *deliveryUsecase.DeliveryService,
	unassign *deliveryUsecase.DeliveryService,
	complete *deliveryUsecase.CompleteService,
	orders order.Gateway,
) *OrderEventProcessor {
	return &OrderEventProcessor{
		assign:   assign,
		unassign: unassign,
		complete: complete,
		orders:   orders,
	}
}

func (p *OrderEventProcessor) Process(ctx context.Context, ev OrderEvent) error {
	rawStatus, err := p.orders.GetStatus(ctx, ev.OrderID)
	if err != nil {
		// order-service: заказ может отсутствовать (дубли, out-of-order, in-memory, удалён)
		if isOrderNotFound(err) {
			return nil
		}
		return err
	}

	status := model.ParseOrderStatus(rawStatus)

	switch status {
	case model.OrderStatusCreated:
		_, _, err := p.assign.Assign(ctx, ev.OrderID)
		return err

	case model.OrderStatusCancelled:
		_, err := p.unassign.Unassign(ctx, ev.OrderID)
		if err != nil && errors.Is(err, deliveryRepo.ErrNotFound) {
			return nil
		}
		return err

	case model.OrderStatusCompleted:
		err := p.complete.Complete(ctx, ev.OrderID)
		if err != nil && errors.Is(err, deliveryRepo.ErrNotFound) {
			return nil
		}
		return err

	default:
		return nil
	}
}

func isOrderNotFound(err error) bool {
	if err == nil {
		return false
	}
	s := err.Error()
	return strings.Contains(s, "status 404") || strings.Contains(s, "not_found")
}

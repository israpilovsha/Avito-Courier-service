package worker

import (
	"context"

	"github.com/Avito-courses/course-go-avito-israpilovsha/internal/delivery/model"
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
		return err
	}

	status := model.ParseOrderStatus(rawStatus)

	switch status {

	case model.OrderStatusCreated:
		_, _, err := p.assign.Assign(ctx, ev.OrderID)
		return err

	case model.OrderStatusCancelled:
		_, err := p.unassign.Unassign(ctx, ev.OrderID)
		return err

	case model.OrderStatusCompleted:
		return p.complete.Complete(ctx, ev.OrderID)

	default:
		// событие устарело или статус не интересен
		return nil
	}
}

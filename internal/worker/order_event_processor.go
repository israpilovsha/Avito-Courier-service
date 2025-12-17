package worker

import (
	"context"
	"strings"

	deliveryUsecase "github.com/Avito-courses/course-go-avito-israpilovsha/internal/delivery/usecase"
)

type OrderEventProcessor struct {
	assign   *deliveryUsecase.DeliveryService
	unassign *deliveryUsecase.DeliveryService
	complete *deliveryUsecase.CompleteService
}

func NewOrderEventProcessor(
	assign *deliveryUsecase.DeliveryService,
	complete *deliveryUsecase.CompleteService,
) *OrderEventProcessor {
	return &OrderEventProcessor{
		assign:   assign,
		unassign: assign,
		complete: complete,
	}
}

func (p *OrderEventProcessor) Process(ctx context.Context, ev OrderEvent) error {
	switch strings.ToLower(ev.Status) {
	case "created":
		_, _, err := p.assign.Assign(ctx, ev.OrderID)
		return err

	case "cancelled":
		_, err := p.unassign.Unassign(ctx, ev.OrderID)
		return err

	case "completed":
		return p.complete.Complete(ctx, ev.OrderID)

	default:

		return nil
	}
}

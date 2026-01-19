package worker

import (
	"context"
	"time"

	"go.uber.org/zap"

	courierModel "github.com/Avito-courses/course-go-avito-israpilovsha/internal/courier/model"
	deliveryModel "github.com/Avito-courses/course-go-avito-israpilovsha/internal/delivery/model"
	orderGateway "github.com/Avito-courses/course-go-avito-israpilovsha/internal/gateway/order"
)

type deliveryAssigner interface {
	Assign(
		ctx context.Context,
		orderID string,
	) (
		*deliveryModel.Delivery,
		*courierModel.Courier,
		error,
	)
}

type OrderFetcher struct {
	gw     orderGateway.Gateway
	svc    deliveryAssigner
	log    *zap.SugaredLogger
	period time.Duration
}

func NewOrderFetcher(
	gw orderGateway.Gateway,
	svc deliveryAssigner,
	log *zap.SugaredLogger,
) *OrderFetcher {
	return &OrderFetcher{
		gw:     gw,
		svc:    svc,
		log:    log,
		period: 5 * time.Second,
	}
}

func (p *OrderFetcher) Run(ctx context.Context) {
	ticker := time.NewTicker(p.period)
	defer ticker.Stop()

	cursor := time.Now().Add(-p.period).UTC()

	for {
		select {
		case <-ctx.Done():
			p.log.Infow("order poller stopped", "reason", ctx.Err())
			return

		case <-ticker.C:
			orders, err := p.gw.FetchOrders(ctx, cursor)
			if err != nil {
				p.log.Warnw("fetch orders failed", "err", err)
				continue
			}

			if len(orders) == 0 {
				cursor = time.Now().Add(-p.period).UTC()
				continue
			}

			maxCreated := cursor
			for _, o := range orders {
				if o.CreatedAt.After(maxCreated) {
					maxCreated = o.CreatedAt
				}

				if _, _, err := p.svc.Assign(ctx, o.ID); err != nil {
					p.log.Warnw("assign failed", "order_id", o.ID, "err", err)
				}
			}

			cursor = maxCreated.Add(time.Nanosecond)
		}
	}
}

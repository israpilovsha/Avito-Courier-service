package usecase

import (
	"context"
	"time"

	courierModel "github.com/Avito-courses/course-go-avito-israpilovsha/internal/courier/model"
	courierRepo "github.com/Avito-courses/course-go-avito-israpilovsha/internal/courier/repository"
	deliveryModel "github.com/Avito-courses/course-go-avito-israpilovsha/internal/delivery/model"
	deliveryRepo "github.com/Avito-courses/course-go-avito-israpilovsha/internal/delivery/repository"
)

type DeliveryService struct {
	courierRepo  courierRepo.CourierRepository
	deliveryRepo deliveryRepo.DeliveryRepository
	nowFunc      func() time.Time
}

func NewDeliveryService(c courierRepo.CourierRepository, d deliveryRepo.DeliveryRepository) *DeliveryService {
	return &DeliveryService{
		courierRepo:  c,
		deliveryRepo: d,
		nowFunc:      time.Now,
	}
}

func (s *DeliveryService) Assign(ctx context.Context, orderID string) (*deliveryModel.Delivery, *courierModel.Courier, error) {
	c, err := s.courierRepo.FindAvailable(ctx)
	if err != nil {
		return nil, nil, err
	}
	if c == nil {
		return nil, nil, courierRepo.ErrNotFound
	}

	var delivery *deliveryModel.Delivery

	err = s.deliveryRepo.WithTx(ctx, func(txCtx context.Context) error {

		now := s.nowFunc()

		delivery = &deliveryModel.Delivery{
			CourierID:  c.ID,
			OrderID:    orderID,
			AssignedAt: now,
			Deadline:   CalculateDeadline(c.TransportType, now),
		}

		if err := s.deliveryRepo.Create(txCtx, delivery); err != nil {
			return err
		}

		if err := s.courierRepo.UpdateStatus(txCtx, c.ID, "busy"); err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		return nil, nil, err
	}

	return delivery, c, nil
}

func (s *DeliveryService) Unassign(ctx context.Context, orderID string) (*deliveryModel.Delivery, error) {
	var result *deliveryModel.Delivery

	err := s.deliveryRepo.WithTx(ctx, func(txCtx context.Context) error {
		d, err := s.deliveryRepo.DeleteByOrderID(txCtx, orderID)
		if err != nil {
			return err
		}
		result = d

		if err := s.courierRepo.UpdateStatus(txCtx, d.CourierID, "available"); err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	return result, nil
}

// ReleaseExpired проверяет просроченные заказы и освобождает курьеров
func (s *DeliveryService) ReleaseExpired(ctx context.Context) error {
	now := s.nowFunc()
	_, err := s.courierRepo.ReleaseExpired(ctx, now)
	return err
}

// StartAutoRelease — фоновая задача
func (s *DeliveryService) StartAutoRelease(ctx context.Context, interval time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			_ = s.ReleaseExpired(ctx)
		}
	}
}

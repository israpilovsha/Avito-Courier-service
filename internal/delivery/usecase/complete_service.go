package usecase

import (
	"context"

	courierRepo "github.com/Avito-courses/course-go-avito-israpilovsha/internal/courier/repository"
	deliveryRepo "github.com/Avito-courses/course-go-avito-israpilovsha/internal/delivery/repository"
)

type CompleteService struct {
	deliveryRepo deliveryRepo.DeliveryRepository
	courierRepo  courierRepo.CourierRepository
}

func NewCompleteService(
	d deliveryRepo.DeliveryRepository,
	c courierRepo.CourierRepository,
) *CompleteService {
	return &CompleteService{
		deliveryRepo: d,
		courierRepo:  c,
	}
}

func (s *CompleteService) Complete(ctx context.Context, orderID string) error {
	return s.deliveryRepo.WithTx(ctx, func(txCtx context.Context) error {
		d, err := s.deliveryRepo.GetByOrderID(txCtx, orderID)
		if err != nil {
			return err
		}
		return s.courierRepo.UpdateStatus(txCtx, d.CourierID, "available")
	})
}

package service

import (
	"context"

	"github.com/Avito-courses/course-go-avito-israpilovsha/internal/courier/model"
	"github.com/Avito-courses/course-go-avito-israpilovsha/internal/courier/repository"
)

type courierService struct {
	repo repository.CourierRepository
}

func NewCourierService(repo repository.CourierRepository) *courierService {
	return &courierService{repo: repo}
}

func (s *courierService) Create(ctx context.Context, c *model.Courier) error {
	return s.repo.Create(ctx, c)
}

func (s *courierService) GetByID(ctx context.Context, id int64) (*model.Courier, error) {
	return s.repo.GetByID(ctx, id)
}

func (s *courierService) GetAll(ctx context.Context) ([]*model.Courier, error) {
	return s.repo.GetAll(ctx)
}

func (s *courierService) Update(ctx context.Context, c *model.Courier) error {
	return s.repo.Update(ctx, c)
}

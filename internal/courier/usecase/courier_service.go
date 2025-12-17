package usecase

import (
	"context"

	"github.com/Avito-courses/course-go-avito-israpilovsha/internal/courier/model"
	"github.com/Avito-courses/course-go-avito-israpilovsha/internal/courier/repository"
)

type CourierService struct {
	repo repository.CourierRepository
}

func NewCourierService(repo repository.CourierRepository) *CourierService {
	return &CourierService{repo: repo}
}

func (s *CourierService) Create(ctx context.Context, c *model.Courier) error {
	return s.repo.Create(ctx, c)
}

func (s *CourierService) GetByID(ctx context.Context, id int64) (*model.Courier, error) {
	return s.repo.GetByID(ctx, id)
}

func (s *CourierService) GetAll(ctx context.Context) ([]*model.Courier, error) {
	return s.repo.GetAll(ctx)
}

func (s *CourierService) Update(ctx context.Context, c *model.Courier) error {
	return s.repo.Update(ctx, c)
}

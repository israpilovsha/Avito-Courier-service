package usecase_test

import (
	"context"
	"errors"
	"testing"

	"github.com/golang/mock/gomock"

	model "github.com/Avito-courses/course-go-avito-israpilovsha/internal/courier/model"
	repoMock "github.com/Avito-courses/course-go-avito-israpilovsha/internal/courier/repository"
	"github.com/Avito-courses/course-go-avito-israpilovsha/internal/courier/usecase"
)

func TestCreate_Success(t *testing.T) {
	t.Parallel()
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	repo := repoMock.NewMockCourierRepository(ctrl)

	c := &model.Courier{Name: "Ivan"}

	repo.EXPECT().
		Create(gomock.Any(), c).
		Return(nil)

	svc := usecase.NewCourierService(repo)

	if err := svc.Create(context.Background(), c); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestCreate_Error(t *testing.T) {
	t.Parallel()
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	repo := repoMock.NewMockCourierRepository(ctrl)

	c := &model.Courier{}

	repo.EXPECT().
		Create(gomock.Any(), c).
		Return(errors.New("db error"))

	svc := usecase.NewCourierService(repo)

	if err := svc.Create(context.Background(), c); err == nil {
		t.Fatalf("expected error, got nil")
	}
}

func TestGetByID_Success(t *testing.T) {
	t.Parallel()
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	repo := repoMock.NewMockCourierRepository(ctrl)

	repo.EXPECT().
		GetByID(gomock.Any(), int64(10)).
		Return(&model.Courier{ID: 10, Name: "Test"}, nil)

	svc := usecase.NewCourierService(repo)

	c, err := svc.GetByID(context.Background(), 10)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if c.ID != 10 {
		t.Fatalf("expected ID=10, got %d", c.ID)
	}
}

func TestGetByID_NotFound(t *testing.T) {
	t.Parallel()
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	repo := repoMock.NewMockCourierRepository(ctrl)

	repo.EXPECT().
		GetByID(gomock.Any(), int64(10)).
		Return(nil, errors.New("not found"))

	svc := usecase.NewCourierService(repo)

	_, err := svc.GetByID(context.Background(), 10)
	if err == nil {
		t.Fatalf("expected error, got nil")
	}
}

func TestGetAll_Success(t *testing.T) {
	t.Parallel()
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	repo := repoMock.NewMockCourierRepository(ctrl)

	expected := []*model.Courier{
		{ID: 1, Name: "A"},
		{ID: 2, Name: "B"},
	}

	repo.EXPECT().
		GetAll(gomock.Any()).
		Return(expected, nil)

	svc := usecase.NewCourierService(repo)

	list, err := svc.GetAll(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(list) != 2 {
		t.Fatalf("expected 2 couriers, got %d", len(list))
	}
}

func TestUpdate_Success(t *testing.T) {
	t.Parallel()
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	repo := repoMock.NewMockCourierRepository(ctrl)

	c := &model.Courier{ID: 5, Name: "Updated"}

	repo.EXPECT().
		Update(gomock.Any(), c).
		Return(nil)

	svc := usecase.NewCourierService(repo)

	if err := svc.Update(context.Background(), c); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestUpdate_Error(t *testing.T) {
	t.Parallel()
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	repo := repoMock.NewMockCourierRepository(ctrl)

	c := &model.Courier{ID: 5}

	repo.EXPECT().
		Update(gomock.Any(), c).
		Return(errors.New("not found"))

	svc := usecase.NewCourierService(repo)

	if err := svc.Update(context.Background(), c); err == nil {
		t.Fatalf("expected error, got nil")
	}
}

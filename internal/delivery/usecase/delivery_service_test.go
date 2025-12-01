package usecase_test

import (
	"context"
	"errors"
	"testing"
	"time"

	courierModel "github.com/Avito-courses/course-go-avito-israpilovsha/internal/courier/model"
	courierMock "github.com/Avito-courses/course-go-avito-israpilovsha/internal/courier/repository"
	deliveryModel "github.com/Avito-courses/course-go-avito-israpilovsha/internal/delivery/model"
	deliveryMock "github.com/Avito-courses/course-go-avito-israpilovsha/internal/delivery/repository"
	"github.com/Avito-courses/course-go-avito-israpilovsha/internal/delivery/usecase"
	"github.com/golang/mock/gomock"
)

func TestAssignSuccess(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	cRepo := courierMock.NewMockCourierRepository(ctrl)
	dRepo := deliveryMock.NewMockDeliveryRepository(ctrl)

	svc := usecase.NewDeliveryService(cRepo, dRepo)

	orderID := "test-order"

	c := &courierModel.Courier{
		ID:            1,
		TransportType: "car",
	}

	cRepo.EXPECT().FindAvailable(gomock.Any()).Return(c, nil)
	dRepo.EXPECT().WithTx(gomock.Any(), gomock.Any()).DoAndReturn(func(_ context.Context, fn func(context.Context) error) error {
		return fn(context.Background())
	})

	dRepo.EXPECT().Create(gomock.Any(), gomock.Any()).Return(nil)
	cRepo.EXPECT().UpdateStatus(gomock.Any(), int64(1), "busy").Return(nil)

	_, courier, err := svc.Assign(context.Background(), orderID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if courier.ID != 1 {
		t.Fatalf("wrong courier ID, expected 1 got %d", courier.ID)
	}
}

// TestAssignNoCouriers - нет доступных курьеров
func TestAssignNoCouriers(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	cRepo := courierMock.NewMockCourierRepository(ctrl)
	dRepo := deliveryMock.NewMockDeliveryRepository(ctrl)

	svc := usecase.NewDeliveryService(cRepo, dRepo)

	cRepo.EXPECT().FindAvailable(gomock.Any()).Return(nil, nil)

	_, _, err := svc.Assign(context.Background(), "x")
	if err == nil {
		t.Fatalf("expected error, got nil")
	}
}

// TestAssignRepoError - ошибка репозитория при Create
func TestAssignRepoError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	cRepo := courierMock.NewMockCourierRepository(ctrl)
	dRepo := deliveryMock.NewMockDeliveryRepository(ctrl)

	svc := usecase.NewDeliveryService(cRepo, dRepo)

	c := &courierModel.Courier{ID: 1, TransportType: "car"}

	cRepo.EXPECT().FindAvailable(gomock.Any()).Return(c, nil)
	dRepo.EXPECT().WithTx(gomock.Any(), gomock.Any()).DoAndReturn(func(_ context.Context, fn func(context.Context) error) error {
		return fn(context.Background())
	})

	dRepo.EXPECT().Create(gomock.Any(), gomock.Any()).Return(errors.New("db error"))

	_, _, err := svc.Assign(context.Background(), "x")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestUnassignSuccess(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	cRepo := courierMock.NewMockCourierRepository(ctrl)
	dRepo := deliveryMock.NewMockDeliveryRepository(ctrl)
	svc := usecase.NewDeliveryService(cRepo, dRepo)

	d := &deliveryModel.Delivery{ID: 1, CourierID: 10, OrderID: "abc"}

	dRepo.EXPECT().WithTx(gomock.Any(), gomock.Any()).
		DoAndReturn(func(_ context.Context, fn func(context.Context) error) error {
			return fn(context.Background())
		})

	dRepo.EXPECT().DeleteByOrderID(gomock.Any(), "abc").Return(d, nil)
	cRepo.EXPECT().UpdateStatus(gomock.Any(), int64(10), "available").Return(nil)

	out, err := svc.Unassign(context.Background(), "abc")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if out.CourierID != 10 {
		t.Fatalf("expected courierID=10, got %d", out.CourierID)
	}
}

func TestUnassignNotFound(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	cRepo := courierMock.NewMockCourierRepository(ctrl)
	dRepo := deliveryMock.NewMockDeliveryRepository(ctrl)
	svc := usecase.NewDeliveryService(cRepo, dRepo)

	dRepo.EXPECT().WithTx(gomock.Any(), gomock.Any()).
		DoAndReturn(func(_ context.Context, fn func(context.Context) error) error {
			return fn(context.Background())
		})

	dRepo.EXPECT().DeleteByOrderID(gomock.Any(), "x").Return(nil, errors.New("not found"))

	_, err := svc.Unassign(context.Background(), "x")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestUnassignUpdateStatusFails(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	cRepo := courierMock.NewMockCourierRepository(ctrl)
	dRepo := deliveryMock.NewMockDeliveryRepository(ctrl)
	svc := usecase.NewDeliveryService(cRepo, dRepo)

	d := &deliveryModel.Delivery{ID: 1, CourierID: 10, OrderID: "abc"}

	dRepo.EXPECT().WithTx(gomock.Any(), gomock.Any()).
		DoAndReturn(func(_ context.Context, fn func(context.Context) error) error {
			return fn(context.Background())
		})

	dRepo.EXPECT().DeleteByOrderID(gomock.Any(), "abc").Return(d, nil)
	cRepo.EXPECT().UpdateStatus(gomock.Any(), int64(10), "available").Return(errors.New("fail"))

	_, err := svc.Unassign(context.Background(), "abc")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestReleaseExpiredSuccess(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	cRepo := courierMock.NewMockCourierRepository(ctrl)
	dRepo := deliveryMock.NewMockDeliveryRepository(ctrl)
	svc := usecase.NewDeliveryService(cRepo, dRepo)

	cRepo.EXPECT().
		ReleaseExpired(gomock.Any(), gomock.Any()).
		Return(int64(5), nil)

	if err := svc.ReleaseExpired(context.Background()); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestReleaseExpiredError(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	cRepo := courierMock.NewMockCourierRepository(ctrl)
	dRepo := deliveryMock.NewMockDeliveryRepository(ctrl)
	svc := usecase.NewDeliveryService(cRepo, dRepo)

	cRepo.EXPECT().
		ReleaseExpired(gomock.Any(), gomock.Any()).
		Return(int64(0), errors.New("db error"))

	if err := svc.ReleaseExpired(context.Background()); err == nil {
		t.Fatalf("expected error, got nil")
	}
}

func TestStartAutoRelease(t *testing.T) {

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	cRepo := courierMock.NewMockCourierRepository(ctrl)
	dRepo := deliveryMock.NewMockDeliveryRepository(ctrl)
	svc := usecase.NewDeliveryService(cRepo, dRepo)

	// ReleaseExpired внутри будет вызывать courierRepo.ReleaseExpired
	cRepo.EXPECT().
		ReleaseExpired(gomock.Any(), gomock.Any()).
		AnyTimes().
		Return(int64(0), nil)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go svc.StartAutoRelease(ctx, 10*time.Millisecond)

	time.Sleep(25 * time.Millisecond)
}

func TestCalculateDeadline(t *testing.T) {
	t.Parallel()

	now := time.Now()

	tests := []struct {
		name      string
		transport string
		add       time.Duration
	}{
		{"on_foot", "on_foot", 30 * time.Minute},
		{"scooter", "scooter", 15 * time.Minute},
		{"car", "car", 5 * time.Minute},
		{"default", "unknown", 30 * time.Minute},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got := usecase.CalculateDeadline(tt.transport, now)
			want := now.Add(tt.add)

			if diff := got.Sub(want); diff > time.Second || diff < -time.Second {
				t.Fatalf("transport=%s, want %v, got %v (diff %v)", tt.transport, want, got, diff)
			}
		})
	}
}

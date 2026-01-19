//go:build integration
// +build integration

package repository_test

import (
	"context"
	"testing"

	"github.com/Avito-courses/course-go-avito-israpilovsha/internal/courier/model"
	"github.com/Avito-courses/course-go-avito-israpilovsha/internal/courier/repository"
	"github.com/Avito-courses/course-go-avito-israpilovsha/internal/db"
	"go.uber.org/zap"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
)

func TestCourierRepositoryIntegration(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	pgContainer, err := postgres.RunContainer(ctx,
		testcontainers.WithImage("postgres:15"),
		postgres.WithDatabase("test_db"),
		postgres.WithUsername("user"),
		postgres.WithPassword("password"),
	)
	if err != nil {
		t.Fatalf("failed to start container: %v", err)
	}

	defer func() {
		_ = pgContainer.Terminate(ctx)
	}()

	dsn, err := pgContainer.ConnectionString(ctx, "sslmode=disable")
	if err != nil {
		t.Fatalf("failed to get DSN: %v", err)
	}

	logger := zap.NewExample().Sugar()
	database := db.New(dsn, logger)

	schema := `
		CREATE TABLE couriers (
			id BIGSERIAL PRIMARY KEY,
			name TEXT NOT NULL,
			phone TEXT NOT NULL,
			status TEXT NOT NULL DEFAULT 'available',
			transport_type TEXT NOT NULL DEFAULT 'on_foot',
			created_at TIMESTAMP DEFAULT NOW(),
			updated_at TIMESTAMP DEFAULT NOW()
		);
	`

	_, err = database.Pool.Exec(ctx, schema)
	if err != nil {
		t.Fatalf("failed to apply schema: %v", err)
	}

	repo := repository.NewCourierRepository(database)

	c := &model.Courier{
		Name:          "Ivan",
		Phone:         "123456",
		Status:        "available",
		TransportType: "scooter",
	}

	err = repo.Create(ctx, c)
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}
	if c.ID == 0 {
		t.Fatalf("expected ID to be assigned, got 0")
	}

	got, err := repo.GetByID(ctx, c.ID)
	if err != nil {
		t.Fatalf("GetByID failed: %v", err)
	}
	if got.Name != "Ivan" {
		t.Fatalf("expected name Ivan, got %s", got.Name)
	}

	c.Status = "busy"
	err = repo.Update(ctx, c)
	if err != nil {
		t.Fatalf("Update failed: %v", err)
	}

	got, err = repo.GetByID(ctx, c.ID)
	if err != nil {
		t.Fatalf("GetByID failed: %v", err)
	}
	if got.Status != "busy" {
		t.Fatalf("expected status busy, got %s", got.Status)
	}

	available, err := repo.FindAvailable(ctx)
	if err != nil {
		t.Fatalf("FindAvailable failed: %v", err)
	}
	if available != nil {
		t.Fatalf("expected no available couriers, got %+v", available)
	}

	err = repo.UpdateStatus(ctx, c.ID, "available")
	if err != nil {
		t.Fatalf("UpdateStatus failed: %v", err)
	}

	got, _ = repo.GetByID(ctx, c.ID)
	if got.Status != "available" {
		t.Fatalf("expected available, got %s", got.Status)
	}
}

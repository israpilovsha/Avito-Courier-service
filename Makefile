include .env
export

DATABASE_URL := postgres://$(POSTGRES_USER):$(POSTGRES_PASSWORD)@$(POSTGRES_HOST):$(POSTGRES_PORT)/$(POSTGRES_DB)?sslmode=disable

.PHONY: migrate migrate-down migrate-status run build

migrate:
	@echo "Running migrations..."
	goose -dir ./migrations postgres "$(DATABASE_URL)" up

migrate-down:
	@echo "Rolling back last migration..."
	goose -dir ./migrations postgres "$(DATABASE_URL)" down

migrate-status:
	@echo "Migration status..."
	goose -dir ./migrations postgres "$(DATABASE_URL)" status

run:
	go run ./cmd/myapp/main.go

build:
	go build -o myapp ./cmd/myapp	

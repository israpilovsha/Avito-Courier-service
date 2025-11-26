# Makefile

# Загружаем переменные из .env
include .env
export

# DATABASE_URL
DATABASE_URL=postgres://${POSTGRES_USER}:${POSTGRES_PASSWORD}@${POSTGRES_HOST}:${POSTGRES_PORT}/${POSTGRES_DB}?sslmode=disable

.PHONY: migrate migrate-down migrate-status run build

# Применить все миграции
migrate:
	@echo "Running migrations..."
	goose -dir ./migrations postgres "$(DATABASE_URL)" up

# Откатить последнюю миграцию
migrate-down:
	@echo "Rolling back last migration..."
	goose -dir ./migrations postgres "$(DATABASE_URL)" down

# Проверка статуса миграций
migrate-status:
	@echo "Migration status..."
	goose -dir ./migrations postgres "$(DATABASE_URL)" status

# Запуск приложения
run:
	go run ./cmd/myapp/main.go

# Сборка бинарника
build:
	go build -o myapp ./cmd/myapp

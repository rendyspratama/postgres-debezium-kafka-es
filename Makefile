include .env
export

.PHONY: run build test clean docker-up docker-down logs lint fmt \
	migrate-up migrate-down migrate-create migrate-status migrate-backup \
	migrate-monitor migrate-test seed-help seed-create seed-apply seed-remove seed-list \
	migrate-verify help

# Build and Run
build-all: build-api build-sync

build-api:
	go build -o bin/api ./api

build-sync:
	go build -o bin/sync ./sync

run-api:
	go run ./api

run-sync:
	go run ./sync

test:
	go test ./...

test-api:
	go test ./api/...

test-sync:
	go test ./sync/...

clean:
	rm -rf bin/

# Docker commands
docker-up:
	docker-compose up -d

docker-down:
	docker-compose down

logs:
	docker-compose logs -f

# Code quality
lint:
	golangci-lint run

fmt:
	go fmt ./...

# Database migrations
migrate-up:
	docker-compose run --rm migrate -path=/migrations -database="postgres://${POSTGRES_USER}:${POSTGRES_PASSWORD}@postgres:5432/${POSTGRES_DB}?sslmode=disable" up

migrate-down:
	docker-compose run --rm migrate -path=/migrations -database="postgres://${POSTGRES_USER}:${POSTGRES_PASSWORD}@postgres:5432/${POSTGRES_DB}?sslmode=disable" down

migrate-create:
	@read -p "Enter migration name: " name; \
	docker-compose run --rm migrate create -ext sql -dir /migrations -seq $$name

migrate-status:
	docker-compose run --rm migrate -path=/migrations -database="postgres://${POSTGRES_USER}:${POSTGRES_PASSWORD}@postgres:5432/${POSTGRES_DB}?sslmode=disable" version
	@echo "\nPending migrations:"
	docker-compose run --rm migrate -path=/migrations -database="postgres://${POSTGRES_USER}:${POSTGRES_PASSWORD}@postgres:5432/${POSTGRES_DB}?sslmode=disable" status

migrate-backup:
	@timestamp=$$(date +%Y%m%d_%H%M%S); \
	docker-compose exec postgres pg_dump -U ${POSTGRES_USER} ${POSTGRES_DB} > "./scripts/migrations/backups/backup_$$timestamp.sql"

# Migration utilities
migrate-monitor:
	@./scripts/migrations/monitor.sh

migrate-test:
	@./scripts/migrations/test.sh

# Seed data management
seed-help:
	@echo "Seed data commands:"
	@echo "  make seed-create <name>  - Create a new seed file"
	@echo "  make seed-apply <file>   - Apply seed data"
	@echo "  make seed-remove <file>  - Remove seed data"
	@echo "  make seed-list          - List available seeds"

seed-create:
	@read -p "Enter seed name: " name; \
	./scripts/migrations/seed.sh create $$name

seed-apply:
	@read -p "Enter seed file name: " file; \
	./scripts/migrations/seed.sh apply "$$file"

seed-remove:
	@read -p "Enter seed file name: " file; \
	./scripts/migrations/seed.sh remove $$file

seed-list:
	@./scripts/migrations/seed.sh list

# Add the verify command
migrate-verify:
	@echo "Verifying migrations..."
	docker-compose run --rm migrate -path=/migrations -database="postgres://user:password@postgres:5432/digital_discovery?sslmode=disable" validate

# Add help target for documentation
help:
	@awk 'BEGIN {FS = ":.*##"; printf "\nUsage:\n  make \033[36m<target>\033[0m\n\nTargets:\n"} /^[a-zA-Z_-]+:.*?##/ { printf "  \033[36m%-15s\033[0m %s\n", $$1, $$2 }' $(MAKEFILE_LIST)

.DEFAULT_GOAL := help

include .env

APP_NAME=pr-reviewer-service
DOCKER_COMPOSE=docker-compose
GOLINT=golangci-lint
GO=go
CMD_DIR=./cmd/app

MIGRATIONS_DIR=./migrations

DB_HOST=localhost
DB_PORT=5436

# PostgreSQL connection variables (can be overridden via environment)
POSTGRES_USER ?= $(POSTGRES_USER)
POSTGRES_PASSWORD ?= $(POSTGRES_PASSWORD)
POSTGRES_DB ?= $(POSTGRES_DB)

help:
	@echo "Available commands:"
	@echo "  make build           - Build application"
	@echo "  make run             - Run application"
	@echo "  make docker-build    - Build Docker image"
	@echo "  make docker-up       - Start services"
	@echo "  make docker-down     - Stop services"
	@echo "  make migrate-up      - Apply migrations"
	@echo "  make docker-logs     - View logs"
	@echo "  make migrate-up      - Apply migrations"
	@echo "  make migrate-down    - Rollback migrations"
	@echo "  make lint            - Run linter"
	@echo "  make load-test       - Run load test"


load-test:
	k6 run tests/load_test.js

build:
	@$(GO) build -o bin/$(APP_NAME) $(CMD_DIR)/main.go

run:
	@$(GO) run $(CMD_DIR)/main.go

docker-build:
	@$(DOCKER_COMPOSE) build

docker-up:
	@$(DOCKER_COMPOSE) up -d

docker-down:
	@$(DOCKER_COMPOSE) down

docker-logs:
	@$(DOCKER_COMPOSE) logs -f

migrate-up: 
	@echo "Applying migrations to database..."
	@for migration in $$(ls $(MIGRATIONS_DIR)/*.up.sql | sort); do \
		echo "Applying: $$(basename $$migration)"; \
		PGPASSWORD=$(POSTGRES_PASSWORD) psql -h $(DB_HOST) -p $(DB_PORT) -U $(POSTGRES_USER) -d $(POSTGRES_DB) -f $$migration || exit 1; \
		echo "✓ Successfully applied: $$(basename $$migration)"; \
	done
	@echo "✓ All migrations successfully applied!"

migrate-down: 
	@echo "Rolling back migrations..."
	@for migration in $$(ls $(MIGRATIONS_DIR)/*.down.sql | sort -r); do \
		echo "Rolling back: $$(basename $$migration)"; \
		PGPASSWORD=$(POSTGRES_PASSWORD) psql -h $(DB_HOST) -p $(DB_PORT) -U $(POSTGRES_USER) -d $(POSTGRES_DB) -f $$migration || exit 1; \
		echo "✓ Successfully rolled back: $$(basename $$migration)"; \
	done
	@echo "✓ All migrations successfully rolled back!"

lint:
	@$(GOLINT) run ./...

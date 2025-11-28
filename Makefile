include .env

APP_NAME=pr-reviewer-service
DOCKER_COMPOSE=docker-compose
GOLINT=golangci-lint
GO=go
CMD_DIR=./cmd/app

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
	${DOCKER_COMPOSE} exec app migrate -path /app/migrations -database "${DB_URL}" up

migrate-down:
	${DOCKER_COMPOSE} exec app migrate -path /app/migrations -database "${DB_URL}" down
lint:
	@$(GOLINT) run ./...

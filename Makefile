APP_NAME := food-ordering-app-backend
DOCKER_COMPOSE ?= docker-compose
GO ?= go

.PHONY: help deps test run build migrate clean \
	up up-build down restart ps logs logs-api logs-db logs-redis \
	health promo runapp

help: ## Show available make targets
	@echo ""
	@echo "Local development targets:"
	@awk 'BEGIN {FS = ":.*##"; printf "  %-18s %s\n", "Target", "Description"} /^[a-zA-Z0-9_.-]+:.*##/ { printf "  %-18s %s\n", $$1, $$2 }' Makefile
	@echo ""

deps: ## Download Go dependencies
	$(GO) mod download

test: ## Run all Go tests
	$(GO) test ./...

run: ## Run API locally (expects local postgres/redis config)
	$(GO) run ./cmd/server

runapp: ## Start postgres/redis via compose, then run API locally
	$(DOCKER_COMPOSE) up -d postgres redis
	@echo "Waiting for database and redis to initialize..."
	@sleep 3
	$(GO) run ./cmd/server

build: ## Build API binary at ./bin/server
	mkdir -p bin
	CGO_ENABLED=0 $(GO) build -o ./bin/server ./cmd/server

migrate: ## Run migrations once
	$(GO) run ./cmd/migrate

clean: ## Remove local build artifacts
	rm -rf ./bin

up: ## Start full stack in background (api + postgres + redis)
	$(DOCKER_COMPOSE) up -d

up-build: ## Rebuild image and start full stack
	$(DOCKER_COMPOSE) up -d --build

down: ## Stop and remove containers (keep volumes)
	$(DOCKER_COMPOSE) down

restart: ## Restart API container
	$(DOCKER_COMPOSE) restart api

ps: ## Show stack status
	$(DOCKER_COMPOSE) ps

logs: ## Follow all service logs
	$(DOCKER_COMPOSE) logs -f

logs-api: ## Follow API logs
	$(DOCKER_COMPOSE) logs -f api

logs-db: ## Follow Postgres logs
	$(DOCKER_COMPOSE) logs -f postgres

logs-redis: ## Follow Redis logs
	$(DOCKER_COMPOSE) logs -f redis

health: ## Check local health endpoint
	curl -fsS http://localhost:8080/health || true

promo: ## Check promo code (usage: make promo CODE=FIFTYOFF)
	@code="$${CODE:-FIFTYOFF}"; \
	curl -fsS "http://localhost:8080/checkpromo?code=$$code" || true


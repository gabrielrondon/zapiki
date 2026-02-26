.PHONY: help build run test clean docker-up docker-down db-migrate check-openapi check-openapi-contract check-coverage validate-secrets rotation-smoke sync-openapi-frontend verify-openapi-sync smoke-proof-flow check-slo-alerts load-evidence design-partner-report

help: ## Show this help message
	@echo 'Usage: make [target]'
	@echo ''
	@echo 'Available targets:'
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "  %-15s %s\n", $$1, $$2}' $(MAKEFILE_LIST)

build: ## Build the API server
	@echo "Building Zapiki API server..."
	@go build -o bin/zapiki-api cmd/api/main.go

build-worker: ## Build the worker
	@echo "Building Zapiki worker..."
	@go build -o bin/zapiki-worker cmd/worker/main.go

build-all: build build-worker ## Build all binaries

run: ## Run the API server
	@echo "Running Zapiki API server..."
	@go run cmd/api/main.go

run-worker: ## Run the worker
	@echo "Running Zapiki worker..."
	@go run cmd/worker/main.go

test: ## Run tests
	@echo "Running tests..."
	@go test -v ./...

test-coverage: ## Run tests with coverage
	@echo "Running tests with coverage..."
	@go test -v -coverprofile=coverage.out ./...
	@go tool cover -html=coverage.out -o coverage.html

clean: ## Clean build artifacts
	@echo "Cleaning..."
	@rm -rf bin/
	@rm -f coverage.out coverage.html

docker-up: ## Start Docker services
	@echo "Starting Docker services..."
	@cd deployments/docker && docker compose up -d
	@echo "Waiting for services to be ready..."
	@sleep 5

docker-down: ## Stop Docker services
	@echo "Stopping Docker services..."
	@cd deployments/docker && docker compose down

docker-logs: ## Show Docker logs
	@cd deployments/docker && docker compose logs -f

db-reset: ## Reset database (WARNING: destructive)
	@echo "Resetting database..."
	@cd deployments/docker && docker compose down -v
	@cd deployments/docker && docker compose up -d postgres
	@sleep 5
	@echo "Database reset complete"

deps: ## Download dependencies
	@echo "Downloading dependencies..."
	@go mod download
	@go mod tidy

fmt: ## Format code
	@echo "Formatting code..."
	@go fmt ./...

lint: ## Run linter
	@echo "Running linter..."
	@golangci-lint run

check-openapi: ## Validate OpenAPI covers implemented routes
	@echo "Checking OpenAPI route coverage..."
	@./scripts/check-openapi-routes.sh

check-openapi-contract: ## Validate OpenAPI semantic contract for critical endpoints
	@echo "Checking OpenAPI semantic contract..."
	@./scripts/check-openapi-contract.sh

sync-openapi-frontend: ## Sync backend openapi.yaml to sibling frontend/public/openapi.yaml
	@echo "Syncing OpenAPI to frontend..."
	@./scripts/sync-frontend-openapi.sh

verify-openapi-sync: ## Verify backend and frontend OpenAPI files are in sync
	@echo "Verifying backend/frontend OpenAPI sync..."
	@./scripts/sync-frontend-openapi.sh --check

check-coverage: ## Validate minimum test coverage
	@echo "Checking coverage threshold..."
	@./scripts/check-coverage.sh coverage.out

validate-secrets: ## Validate required production secrets/env vars
	@echo "Validating production secrets..."
	@./scripts/validate-production-secrets.sh

rotation-smoke: ## Run post-rotation smoke test (requires ZAPIKI_BASE_URL and ZAPIKI_API_KEY)
	@echo "Running secret rotation smoke test..."
	@./scripts/secret-rotation-smoke.sh

smoke-proof-flow: ## Run end-to-end proof smoke (generate -> get -> verify)
	@echo "Running proof flow smoke test..."
	@./scripts/smoke-proof-flow.sh

check-slo-alerts: ## Validate required SLO alert rules exist
	@echo "Checking SLO alert rules..."
	@./scripts/check-slo-alert-rules.sh

load-evidence: ## Run commitment load test and generate markdown evidence report
	@echo "Running load evidence script..."
	@./scripts/load-evidence.sh

design-partner-report: ## Generate design-partner pipeline report and gate status
	@echo "Generating design-partner report..."
	@./scripts/design-partner-report.sh

dev: docker-up run ## Start development environment

.DEFAULT_GOAL := help

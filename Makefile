.PHONY: help build run test clean docker-up docker-down db-migrate

help: ## Show this help message
	@echo 'Usage: make [target]'
	@echo ''
	@echo 'Available targets:'
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "  %-15s %s\n", $$1, $$2}' $(MAKEFILE_LIST)

build: ## Build the API server
	@echo "Building Zapiki API server..."
	@go build -o bin/zapiki-api cmd/api/main.go

run: ## Run the API server
	@echo "Running Zapiki API server..."
	@go run cmd/api/main.go

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
	@cd deployments/docker && docker-compose up -d
	@echo "Waiting for services to be ready..."
	@sleep 5

docker-down: ## Stop Docker services
	@echo "Stopping Docker services..."
	@cd deployments/docker && docker-compose down

docker-logs: ## Show Docker logs
	@cd deployments/docker && docker-compose logs -f

db-reset: ## Reset database (WARNING: destructive)
	@echo "Resetting database..."
	@cd deployments/docker && docker-compose down -v
	@cd deployments/docker && docker-compose up -d postgres
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
	@golangci-lint run || true

dev: docker-up run ## Start development environment

.DEFAULT_GOAL := help

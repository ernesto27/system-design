# Makefile for Twitter Service Development

.PHONY: help run build test clean deps start-db stop-db

help: ## Show available commands
	@echo "Available commands:"
	@echo "  run        - Run the application locally"
	@echo "  build      - Build the application"
	@echo "  test       - Run tests"
	@echo "  clean      - Clean build artifacts"
	@echo "  deps       - Download dependencies"
	@echo "  start-db   - Start PostgreSQL database"
	@echo "  stop-db    - Stop PostgreSQL database"

run: ## Run the application locally
	@echo "Starting Twitter Service..."
	go run cmd/server/main.go

build: ## Build the application
	@echo "Building Twitter Service..."
	go build -o bin/twitter-service cmd/server/main.go

test: ## Run tests
	@echo "Running tests..."
	go test -v ./...

clean: ## Clean build artifacts
	@echo "Cleaning..."
	rm -rf bin/
	go clean

deps: ## Download dependencies
	@echo "Downloading dependencies..."
	go mod tidy
	go mod download

start-db: ## Start PostgreSQL database
	@echo "Starting PostgreSQL database..."
	docker-compose up -d postgres

stop-db: ## Stop PostgreSQL database
	@echo "Stopping PostgreSQL database..."
	docker-compose down

dev: ## Start development environment (database + server)
	@echo "Starting development environment..."
	make start-db
	@echo "Waiting for database to be ready..."
	@sleep 5
	@echo "Starting server..."
	make run

docker-build: ## Build Docker image
	@echo "Building Docker image..."
	docker build -t twitter-service .

docker-run: ## Run Docker container
	@echo "Running Docker container..."
	docker run -p 8080:8080 --env-file .env twitter-service

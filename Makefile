# Makefile for Auth Microservice

.PHONY: docs docs-serve build run test clean docker-build docker-up docker-down

# Variables
SERVICE_NAME := auth-service
DOCKER_COMPOSE := docker-compose

# Documentation
docs:
	@echo "🔄 Generating Swagger documentation..."
	swag init -g cmd/auth-service/main.go -o docs/
	@echo "✅ Documentation generated successfully!"
	@echo "📖 Access at: http://localhost:8080/docs/"

docs-serve: docs
	@echo "🚀 Starting service with updated docs..."
	go run cmd/auth-service/main.go

# Build
build:
	@echo "🔨 Building $(SERVICE_NAME)..."
	go build -o bin/$(SERVICE_NAME) cmd/auth-service/main.go
	@echo "✅ Build completed: bin/$(SERVICE_NAME)"

# Run
run:
	@echo "🚀 Running $(SERVICE_NAME)..."
	go run cmd/auth-service/main.go

# Test
test:
	@echo "🧪 Running tests..."
	go test -v ./...

# Clean
clean:
	@echo "🧹 Cleaning build artifacts..."
	rm -rf bin/
	@echo "✅ Clean completed"

# Docker commands
docker-build:
	@echo "🐳 Building Docker image..."
	$(DOCKER_COMPOSE) build

docker-up:
	@echo "🐳 Starting Docker services..."
	$(DOCKER_COMPOSE) up -d --build
	@echo "✅ Services started!"
	@echo "📖 Swagger UI: http://localhost:8080/docs/"
	@echo "📊 Prometheus: http://localhost:9090"
	@echo "📈 Grafana: http://localhost:3000"
	@echo "🗄️ PgAdmin: http://localhost:5050"
	@echo "🔍 Jaeger: http://localhost:16686"

docker-down:
	@echo "🐳 Stopping Docker services..."
	$(DOCKER_COMPOSE) down

docker-logs:
	@echo "📋 Showing Docker logs..."
	$(DOCKER_COMPOSE) logs -f

# Development helpers
dev-setup:
	@echo "🔧 Setting up development environment..."
	cp .env.example .env
	@echo "✅ Please edit .env file with your configurations"

# Install swag if not present
install-swag:
	@which swag > /dev/null || (echo "📦 Installing swag..." && go install github.com/swaggo/swag/cmd/swag@latest)

# Complete setup
setup: install-swag dev-setup docs
	@echo "🎉 Setup completed! Run 'make docker-up' to start all services"

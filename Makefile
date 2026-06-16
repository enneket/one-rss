.PHONY: dev build test lint clean docker

# Development
dev:
	@echo "Starting development server..."
	@cd frontend && npm run dev &
	@go run main.go

# Build frontend
frontend-build:
	@echo "Building frontend..."
	@cd frontend && npm run build

# Build Go binary
build: frontend-build
	@echo "Building Go binary..."
	@CGO_ENABLED=0 go build -o bin/one-rss main.go

# Run tests
test:
	@echo "Running Go tests..."
	@go test ./...
	@echo "Running frontend tests..."
	@cd frontend && npm test

# Lint code
lint:
	@echo "Linting Go code..."
	@golangci-lint run
	@echo "Linting frontend code..."
	@cd frontend && npm run lint

# Clean build artifacts
clean:
	@echo "Cleaning..."
	@rm -rf bin/
	@rm -rf frontend/dist/
	@rm -rf frontend/node_modules/

# Install dependencies
install:
	@echo "Installing Go dependencies..."
	@go mod download
	@echo "Installing frontend dependencies..."
	@cd frontend && npm install

# Docker build
docker:
	@echo "Building Docker image..."
	@docker build -t one-rss:latest .

# Docker run
docker-run:
	@echo "Running Docker container..."
	@docker run -p 6011:6011 -v one-rss-data:/data one-rss:latest

# Help
help:
	@echo "Available commands:"
	@echo "  make dev          - Start development server"
	@echo "  make build        - Build production binary"
	@echo "  make test         - Run tests"
	@echo "  make lint         - Lint code"
	@echo "  make clean        - Clean build artifacts"
	@echo "  make install      - Install dependencies"
	@echo "  make docker       - Build Docker image"
	@echo "  make docker-run   - Run Docker container"

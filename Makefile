.PHONY: help build run test clean deps lint fmt install-tools

# Variables
BINARY_NAME=smanzy
BINARY_PATH=./bin/$(BINARY_NAME)
MAIN_PATH=./cmd/api/main.go

help:
	@echo "Available targets:"
	@echo "  make build         - Build the application"
	@echo "  make run           - Run the application"
	@echo "  make test          - Run tests"
	@echo "  make clean         - Clean build artifacts"
	@echo "  make deps          - Download dependencies"
	@echo "  make fmt           - Format code"
	@echo "  make lint          - Lint code (requires golangci-lint)"
	@echo "  make dev           - Run in development mode"

# Install dependencies
deps:
	@echo "Downloading dependencies..."
	go mod download
	go mod tidy

# Build the application
build: clean
	@echo "Building application..."
	mkdir -p bin
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o $(BINARY_PATH) $(MAIN_PATH)
	@echo "Built $(BINARY_PATH)"

# Run the application
run:
	@echo "Running application..."
	go run $(MAIN_PATH)

# Development mode with hot reload (requires air)
dev:
	@command -v -migrate air >/dev/null 2>&1 || go install github.com/air-verse/air@latest
	@PATH=$$(go env GOPATH)/bin:$$PATH air

# Run tests
test:
	@echo "Running tests..."
	go test -v -cover ./...

# Format code
fmt:
	@echo "Formatting code..."
	go fmt ./...

# Lint code (requires golangci-lint)
lint:
	@command -v golangci-lint >/dev/null 2>&1 || (echo "Installing golangci-lint..."; go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest)
	@echo "Linting code..."
	golangci-lint run ./...

# Install development tools
install-tools:
	@echo "Installing development tools..."
	go install github.com/air-verse/air@latest
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest

# Clean build artifacts
clean:
	@echo "Cleaning build artifacts..."
	rm -rf bin/
	go clean

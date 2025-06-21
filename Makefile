# GW2 MCP Server Makefile

.PHONY: build clean test lint format deps run help

# Default target
all: format lint test build

# Build the server
build:
	@echo "Building GW2 MCP Server..."
	go build -o bin/gw2-mcp -ldflags "-s -w" ./

# Clean build artifacts
clean:
	@echo "Cleaning build artifacts..."
	rm -rf bin/
	go clean

# Run tests
test:
	@echo "Running tests..."
	go test -v -race -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html

# Run linter
lint:
	@echo "Running linter..."
	golangci-lint run .

# Format code
format:
	@echo "Formatting code..."
	gofumpt -w .
	go mod tidy

# Install/update dependencies
deps:
	@echo "Installing dependencies..."
	go mod download
	go mod verify

# Run the server
run: build
	@echo "Starting GW2 MCP Server..."
	./bin/gw2-mcp

# Development run (with race detection)
dev:
	@echo "Starting GW2 MCP Server in development mode..."
	go run -race ./

# Install development tools
tools:
	@echo "Installing development tools..."
	go install mvdan.cc/gofumpt@latest
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest

# Check for security vulnerabilities
security:
	@echo "Checking for security vulnerabilities..."
	go list -json -deps ./... | nancy sleuth

# Generate documentation
docs:
	@echo "Generating documentation..."
	go doc -all ./... > docs/api.txt

# Release build (optimized)
release:
	@echo "Building release version..."
	CGO_ENABLED=0 go build -a -installsuffix cgo -ldflags "-s -w -X main.version=$(shell git describe --tags --always)" -o bin/gw2-mcp ./

# Docker build
docker:
	@echo "Building Docker image..."
	docker build -t gw2-mcp:latest .

# Show help
help:
	@echo "Available targets:"
	@echo "  build     - Build the server binary"
	@echo "  clean     - Clean build artifacts"
	@echo "  test      - Run tests with coverage"
	@echo "  lint      - Run linter"
	@echo "  format    - Format code and tidy modules"
	@echo "  deps      - Install/update dependencies"
	@echo "  run       - Build and run the server"
	@echo "  dev       - Run in development mode with race detection"
	@echo "  tools     - Install development tools"
	@echo "  security  - Check for security vulnerabilities"
	@echo "  docs      - Generate documentation"
	@echo "  release   - Build optimized release version"
	@echo "  docker    - Build Docker image"
	@echo "  help      - Show this help message"

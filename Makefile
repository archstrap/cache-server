# Makefile for Redis Cache Server
# Go project build, run, and maintenance commands

# Variables
BINARY_NAME=redis-server
CMD_PATH=./cmd/redis_server
BUILD_DIR=./bin
BUILD_OUTPUT=$(BUILD_DIR)/$(BINARY_NAME)
GO_FILES=$(shell find . -name '*.go' -not -path './vendor/*')

# Default target
.DEFAULT_GOAL := help

# Build the application
.PHONY: build
build:
	@echo "Building $(BINARY_NAME)..."
	@mkdir -p $(BUILD_DIR)
	@go build -o $(BUILD_OUTPUT) $(CMD_PATH)
	@echo "Build complete: $(BUILD_OUTPUT)"

# Run the application
.PHONY: run
run:
	@echo "Running $(BINARY_NAME)..."
	@go run $(CMD_PATH)/*.go

# Format Go code
.PHONY: format
format:
	@echo "Formatting Go code..."
	@go fmt ./...
	@echo "Formatting complete"

# Run tests
.PHONY: test
test:
	@echo "Running tests..."
	@go test -v ./...

# Run tests with coverage
.PHONY: test-coverage
test-coverage:
	@echo "Running tests with coverage..."
	@go test -v -coverprofile=coverage.out ./...
	@go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"

# Lint the code (requires golangci-lint)
.PHONY: lint
lint:
	@echo "Linting code..."
	@if command -v golangci-lint > /dev/null; then \
		golangci-lint run ./...; \
	else \
		echo "golangci-lint not installed. Install with: go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest"; \
	fi

# Clean build artifacts
.PHONY: clean
clean:
	@echo "Cleaning build artifacts..."
	@rm -rf $(BUILD_DIR)
	@rm -f coverage.out coverage.html
	@go clean
	@echo "Clean complete"

# Install dependencies
.PHONY: deps
deps:
	@echo "Downloading dependencies..."
	@go mod download
	@go mod tidy
	@echo "Dependencies updated"

# Verify code (format, vet, test)
.PHONY: verify
verify: format vet test
	@echo "Verification complete"

# Run go vet
.PHONY: vet
vet:
	@echo "Running go vet..."
	@go vet ./...

# Show help
.PHONY: help
help:
	@echo "Available targets:"
	@echo "  make build         - Build the application"
	@echo "  make run           - Run the application"
	@echo "  make format        - Format Go code"
	@echo "  make test          - Run tests"
	@echo "  make test-coverage - Run tests with coverage report"
	@echo "  make lint          - Lint the code (requires golangci-lint)"
	@echo "  make vet           - Run go vet"
	@echo "  make verify        - Format, vet, and test"
	@echo "  make deps           - Download and tidy dependencies"
	@echo "  make clean         - Clean build artifacts"
	@echo "  make help          - Show this help message"


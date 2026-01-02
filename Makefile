.PHONY: build test clean run-demo install

# Build the MeritChain binary
build:
	@echo "Building MeritChain..."
	@go build -o meritchain ./cmd/meritchain
	@echo "✓ Build complete: ./meritchain"

# Run all tests
test:
	@echo "Running tests..."
	@go test ./pkg/... -v

# Run tests with coverage
test-coverage:
	@echo "Running tests with coverage..."
	@go test ./pkg/... -coverprofile=coverage.out
	@go tool cover -html=coverage.out -o coverage.html
	@echo "✓ Coverage report generated: coverage.html"

# Clean build artifacts
clean:
	@echo "Cleaning build artifacts..."
	@rm -f meritchain coverage.out coverage.html
	@echo "✓ Clean complete"

# Run the demo
run-demo: build
	@./meritchain demo

# Install dependencies
install:
	@echo "Installing dependencies..."
	@go mod download
	@go mod tidy
	@echo "✓ Dependencies installed"

# Build and run
run: build
	@./meritchain

# Format code
fmt:
	@echo "Formatting code..."
	@go fmt ./...
	@echo "✓ Code formatted"

# Run linter (requires golangci-lint)
lint:
	@echo "Running linter..."
	@golangci-lint run ./...
	@echo "✓ Linting complete"

# Help
help:
	@echo "MeritChain Makefile"
	@echo ""
	@echo "Available targets:"
	@echo "  build         - Build the MeritChain binary"
	@echo "  test          - Run all tests"
	@echo "  test-coverage - Run tests with coverage report"
	@echo "  clean         - Clean build artifacts"
	@echo "  run-demo      - Build and run demo"
	@echo "  install       - Install dependencies"
	@echo "  run           - Build and run CLI"
	@echo "  fmt           - Format code"
	@echo "  lint          - Run linter (requires golangci-lint)"
	@echo "  help          - Show this help message"

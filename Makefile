.PHONY: help build test test-verbose test-race test-coverage bench lint clean install run

# Default target
help:
	@echo "Subzy - Subdomain Takeover Detection Tool"
	@echo ""
	@echo "Available targets:"
	@echo "  build          - Build the binary"
	@echo "  test           - Run tests"
	@echo "  test-verbose   - Run tests with verbose output"
	@echo "  test-race      - Run tests with race detector"
	@echo "  test-coverage  - Run tests with coverage report"
	@echo "  bench          - Run benchmark tests"
	@echo "  lint           - Run golangci-lint"
	@echo "  clean          - Remove build artifacts"
	@echo "  install        - Install the binary to GOPATH/bin"
	@echo "  run            - Build and run with example targets"

# Build the application
build:
	@echo "Building subzy..."
	go build -o subzy main.go
	@echo "Build complete: ./subzy"

# Run all tests
test:
	@echo "Running tests..."
	go test ./...

# Run tests with verbose output
test-verbose:
	@echo "Running tests (verbose)..."
	go test -v ./...

# Run tests with race detector
test-race:
	@echo "Running tests with race detector..."
	go test -race ./...

# Run tests with coverage
test-coverage:
	@echo "Running tests with coverage..."
	go test -cover ./...
	@echo ""
	@echo "Generating detailed coverage report..."
	go test -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"

# Run benchmark tests
bench:
	@echo "Running benchmarks..."
	go test -bench=. -benchmem ./runner

# Run linter
lint:
	@echo "Running golangci-lint..."
	@which golangci-lint > /dev/null || (echo "golangci-lint not installed. Install from https://golangci-lint.run/usage/install/" && exit 1)
	golangci-lint run

# Clean build artifacts
clean:
	@echo "Cleaning build artifacts..."
	rm -f subzy
	rm -f coverage.out coverage.html
	@echo "Clean complete"

# Install to GOPATH/bin
install:
	@echo "Installing subzy..."
	go install -v

# Example run
run: build
	@echo "Running subzy with example targets..."
	@echo "Note: Create list.txt with test domains first"
	./subzy run --help

# Development helpers
fmt:
	@echo "Formatting code..."
	go fmt ./...

vet:
	@echo "Running go vet..."
	go vet ./...

# Run all checks (fmt, vet, test, lint)
check: fmt vet test lint
	@echo "All checks passed!"

# Update dependencies
deps-update:
	@echo "Updating dependencies..."
	go get -u ./...
	go mod tidy

# Show dependency graph
deps-graph:
	@echo "Dependency graph:"
	go mod graph

# Check for security vulnerabilities
security:
	@echo "Checking for security vulnerabilities..."
	@which govulncheck > /dev/null || go install golang.org/x/vuln/cmd/govulncheck@latest
	govulncheck ./...

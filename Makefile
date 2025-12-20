.PHONY: help test coverage lint build clean examples

help:
	@echo "go-agentic Development Tasks"
	@echo "=============================="
	@echo "make test              - Run all tests"
	@echo "make test-verbose      - Run tests with verbose output"
	@echo "make test-race         - Run tests with race detector"
	@echo "make coverage          - Run tests and generate coverage report"
	@echo "make coverage-html     - Generate HTML coverage report"
	@echo "make lint              - Run linter checks"
	@echo "make build             - Build library"
	@echo "make examples          - Build all examples"
	@echo "make clean             - Clean build artifacts"
	@echo "make benchmark         - Run performance benchmarks"

test:
	@echo "Running tests on $(shell go env GOOS)..."
	go test -v ./...

test-verbose:
	@echo "Running tests with verbose output..."
	go test -v -race ./...

test-race:
	@echo "Running tests with race detector..."
	go test -race ./...

coverage:
	@echo "Running tests with coverage..."
	go test -coverprofile=coverage.out ./...
	@echo "Coverage summary:"
	@go tool cover -func=coverage.out | grep total

coverage-html: coverage
	@echo "Generating HTML coverage report..."
	go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"

lint:
	@echo "Running linter..."
	golangci-lint run ./...

build:
	@echo "Building library..."
	go build -v ./go-agentic/...

examples: build
	@echo "Building examples..."
	go build -o ./examples/it-support/it-support-example ./examples/it-support
	go build -o ./examples/research-assistant/research-assistant-example ./examples/research-assistant
	go build -o ./examples/customer-service/customer-service-example ./examples/customer-service
	go build -o ./examples/data-analysis/data-analysis-example ./examples/data-analysis

clean:
	@echo "Cleaning build artifacts..."
	go clean -v ./...
	rm -f coverage.out coverage.html
	rm -f ./examples/*/example-*

benchmark:
	@echo "Running benchmarks..."
	go test -bench=. -benchmem ./...

# CI/CD target - runs before commit
ci: test lint coverage
	@echo "CI checks passed!"

# Development target - quick tests during development
dev: test-race
	@echo "Development tests passed!"

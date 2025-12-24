.PHONY: run run-verbose build clean help

help:
	@echo "Multi-Team Demo - Makefile Commands"
	@echo "===================================="
	@echo "make run          - Run the multi-team demo"
	@echo "make run-verbose  - Run with verbose output"
	@echo "make build        - Build the binary"
	@echo "make clean        - Clean up"

run:
	@go run ./cmd/

run-verbose:
	@go run ./cmd/ -v

build:
	@go build -o multi-team ./cmd/
	@echo "Built: ./multi-team"

clean:
	@rm -f multi-team
	@go clean
	@echo "Cleaned up"

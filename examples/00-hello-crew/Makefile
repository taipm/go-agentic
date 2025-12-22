.PHONY: run run-server build clean help

help:
	@echo "Hello Crew - Makefile Commands"
	@echo "==============================="
	@echo "make run          - Run in CLI mode"
	@echo "make run-server   - Run in server mode on port 8081"
	@echo "make build        - Build the binary"
	@echo "make clean        - Clean up compiled files"

run:
	@if [ ! -f .env ]; then \
		echo "Creating .env from .env.example..."; \
		cp .env.example .env; \
		echo "Note: Ollama is the default. Ensure 'ollama serve' is running."; \
		echo "To use OpenAI instead, edit .env and set OPENAI_API_KEY"; \
	fi
	@set -a; . ./.env; set +a; go run cmd/main.go

run-server:
	@if [ ! -f .env ]; then \
		echo "Creating .env from .env.example..."; \
		cp .env.example .env; \
		echo "Note: Ollama is the default. Ensure 'ollama serve' is running."; \
		echo "To use OpenAI instead, edit .env and set OPENAI_API_KEY"; \
	fi
	@set -a; . ./.env; set +a; go run cmd/main.go -server -port 8081

build:
	@go build -o hello-crew cmd/main.go
	@echo "Built: ./hello-crew"

clean:
	@rm -f hello-crew
	@go clean
	@echo "Cleaned up"

.PHONY: build run clean test help

# Build the application
build:
	@echo "Building hello-crew-tools..."
	go build -o hello-crew-tools ./cmd/main.go
	@echo "✅ Build complete: ./hello-crew-tools"

# Run in CLI mode
run:
	@echo "Starting Hello Crew with Tools (CLI mode)..."
	@echo "Make sure Ollama is running on http://localhost:11434"
	go run ./cmd/main.go

# Run in server mode
server:
	@echo "Starting Hello Crew with Tools (Server mode on port 8082)..."
	go run ./cmd/main.go -server -port 8082

# Run tests
test:
	@echo "Running tests..."
	go test -v ./...

# Clean build artifacts
clean:
	rm -f hello-crew-tools
	go clean

# Show help
help:
	@echo "Hello Crew with Tools - Memory Validation Example"
	@echo ""
	@echo "Available targets:"
	@echo "  make build   - Build the application"
	@echo "  make run     - Run in interactive CLI mode"
	@echo "  make server  - Run in server mode (port 8082)"
	@echo "  make test    - Run tests"
	@echo "  make clean   - Clean build artifacts"
	@echo "  make help    - Show this help message"
	@echo ""
	@echo "Testing Memory with Tools:"
	@echo ""
	@echo "1. Start Ollama:"
	@echo "   ollama run deepseek-r1:1.5b"
	@echo ""
	@echo "2. In another terminal, run:"
	@echo "   make run"
	@echo ""
	@echo "3. Try these conversations:"
	@echo "   > Tôi tên gì?"
	@echo "   > Tôi là John Doe"
	@echo "   > Tôi tên gì?"
	@echo "   > Tôi đã hỏi mấy câu?"
	@echo "   > Bạn nhớ tôi nói gì lần đầu?"
	@echo ""
	@echo "Expected Results:"
	@echo "   - Agent should use tools to count messages"
	@echo "   - Agent should retrieve facts from conversation"
	@echo "   - With tools: Agent CAN count (unlike hello-crew)"
	@echo "   - Without tools: Agent CANNOT count correctly"

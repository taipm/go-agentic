.PHONY: build run clean test

# Build the application
build:
	go build -o quiz-exam ./cmd/main.go

# Run the quiz exam
run:
	go run ./cmd/main.go

# Run with verbose output
run-verbose:
	go run ./cmd/main.go -verbose

# Clean build artifacts
clean:
	rm -f quiz-exam

# Update dependencies
deps:
	go mod tidy

# Test the example compiles
test:
	go build ./...

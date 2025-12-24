# Deployment Guide

## Production Checklist

- [ ] Choose and configure LLM provider (Ollama or OpenAI)
- [ ] Set environment variables securely
- [ ] Configure agents with appropriate models
- [ ] Set safety limits (MaxRounds, MaxHandoffs)
- [ ] Test with multiple inputs
- [ ] Monitor token usage and costs
- [ ] Set up error logging
- [ ] Configure HTTP server (if using)
- [ ] Set up reverse proxy (if needed)
- [ ] Test context cancellation
- [ ] Load test the system

## Configuration for Production

### OpenAI Configuration

```yaml
# config/agents/my-agent.yaml
model: gpt-4o  # Use gpt-4o, not gpt-4o-mini, for production
provider: openai
temperature: 0.3  # Lower temperature for consistency
max_tokens: 2000  # Control output length
```

Set API key securely:

```bash
# Option 1: Environment variable (recommended)
export OPENAI_API_KEY="sk-xxx..."

# Option 2: .env file (for development only)
# Create .env file with OPENAI_API_KEY=sk-xxx...
# Load with: source .env
```

### Ollama Configuration (Self-Hosted)

```yaml
# config/agents/my-agent.yaml
model: deepseek-r1:7b  # Larger model for production
provider: ollama
provider_url: http://ollama-server:11434
temperature: 0.3
```

Start Ollama server:

```bash
# Docker
docker run -d \
  --name ollama \
  -p 11434:11434 \
  -v ollama:/root/.ollama \
  ollama/ollama

# Then pull model
curl http://ollama-server:11434/api/pull -d '{"name":"deepseek-r1:7b"}'
```

## Safety Limits

Configure for your use case:

```yaml
# config/crew.yaml

# For simple tasks (2-3 tool uses)
max_rounds: 5
max_handoffs: 2

# For complex tasks (multiple investigations)
max_rounds: 15
max_handoffs: 4

# For very complex multi-step processes
max_rounds: 20
max_handoffs: 5
```

## HTTP Server Deployment

### Basic Setup

```go
package main

import (
	"log"
	"net/http"
	"os"
	
	crewai "github.com/taipm/go-crewai"
)

func main() {
	// Load crew configuration
	crew, err := crewai.LoadCrewConfig("config/crew.yaml")
	if err != nil {
		log.Fatal(err)
	}

	// Create executor
	executor := crewai.NewCrewExecutor(crew, os.Getenv("OPENAI_API_KEY"))

	// Set up HTTP handlers
	http.HandleFunc("/execute", func(w http.ResponseWriter, r *http.Request) {
		// Handle crew execution requests
		crewai.HandleCrewRequest(executor, w, r)
	})

	// Start server
	port := ":8080"
	log.Printf("Server starting on %s\n", port)
	log.Fatal(http.ListenAndServe(port, nil))
}
```

### Reverse Proxy (nginx)

```nginx
upstream go_agentic {
    server localhost:8080;
}

server {
    listen 80;
    server_name api.example.com;

    # Rate limiting
    limit_req_zone $binary_remote_addr zone=api:10m rate=10r/s;
    limit_req zone=api burst=20;

    location /execute {
        # Authentication
        auth_basic "Restricted";
        auth_basic_user_file /etc/nginx/.htpasswd;

        # Proxy settings
        proxy_pass http://go_agentic;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        
        # Timeouts for long-running requests
        proxy_connect_timeout 60s;
        proxy_send_timeout 300s;
        proxy_read_timeout 300s;
    }

    # Health check
    location /health {
        proxy_pass http://go_agentic;
    }
}
```

### Docker Deployment

Create `Dockerfile`:

```dockerfile
FROM golang:1.25-alpine AS builder

WORKDIR /app
COPY . .

# Build
RUN go build -o app ./cmd/main.go

# Runtime stage
FROM alpine:latest

RUN apk --no-cache add ca-certificates

WORKDIR /app
COPY --from=builder /app/app .
COPY config ./config

EXPOSE 8080
ENV OPENAI_API_KEY=""

CMD ["./app", "--server", "--port", "8080"]
```

Build and run:

```bash
# Build
docker build -t go-agentic:latest .

# Run
docker run -d \
  -p 8080:8080 \
  -e OPENAI_API_KEY="sk-xxx..." \
  --name go-agentic \
  go-agentic:latest
```

### Docker Compose

Create `docker-compose.yml`:

```yaml
version: '3.8'

services:
  app:
    build: .
    ports:
      - "8080:8080"
    environment:
      OPENAI_API_KEY: ${OPENAI_API_KEY}
    volumes:
      - ./config:/app/config
    restart: unless-stopped

  ollama:
    image: ollama/ollama
    ports:
      - "11434:11434"
    volumes:
      - ollama:/root/.ollama
    environment:
      OLLAMA_MODELS: deepseek-r1:7b

volumes:
  ollama:
```

Run:

```bash
docker-compose up -d
```

## Monitoring and Logging

### Error Logging

```go
import "log"

// Structured logging
log.Printf("Agent: %s, Status: %s, Duration: %dms", 
    agent.Name, status, duration)
```

### Token Usage Tracking

Monitor costs with OpenAI:

```bash
# Check your usage
curl -H "Authorization: Bearer $OPENAI_API_KEY" \
  https://api.openai.com/v1/usage/tokens
```

### Health Checks

Add a health endpoint:

```go
http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"status":"healthy"}`))
})
```

## Scaling

### Horizontal Scaling

For multiple replicas:

```bash
# Load balancer configuration
upstream go_agentic {
    server instance1:8080;
    server instance2:8080;
    server instance3:8080;
}
```

### Performance Tuning

```go
// Adjust based on your machine
crew := &crewai.Crew{
	Agents:      agents,
	MaxRounds:   10,        // Adjust based on complexity
	MaxHandoffs: 5,         // Prevent endless routing
}
```

## Troubleshooting

### High Latency

- Check model size (use smaller models for faster responses)
- Monitor network latency to LLM provider
- Check MaxRounds - might be excessive
- Profile CPU and memory usage

### High Cost

- Switch from OpenAI to local Ollama for development
- Reduce MaxRounds to limit tool calls
- Use smaller models (gpt-4o-mini instead of gpt-4o)
- Monitor token usage

### Rate Limiting

Configure reasonable limits:

```yaml
# In nginx or application
rate_limit: 100/minute  # Requests per minute
burst: 20               # Allow burst up to 20
```

## Backup and Recovery

### Configuration Backups

```bash
# Backup agent configurations
tar -czf agent-backup-$(date +%Y%m%d).tar.gz config/

# Version control
git commit -m "Backup agent configuration"
git push
```

### State Management

For stateful applications:

```go
// Save conversation history
type ConversationLog struct {
	ID        string
	Messages  []Message
	CreatedAt time.Time
}

// Load and continue conversation
log, _ := LoadConversation(conversationID)
result, _ := executor.Execute(ctx, newMessage, WithHistory(log))
```

# Examples

## IT Support System

A complete multi-agent IT troubleshooting system with 3 agents and 13 tools.

**Location**: `examples/it-support/`

### Agents

1. **Orchestrator (My)** - Entry point and router
   - Analyzes user input
   - Decides between clarification or execution
   - Routes to appropriate agent

2. **Clarifier (Ngân)** - Information gatherer
   - Asks clarifying questions
   - Gathers system information
   - Routes to executor with context

3. **Executor (Trang)** - Technical expert (terminal)
   - Executes diagnostic tools
   - Analyzes results
   - Provides recommendations

### Tools Available

- `GetSystemInfo()` - OS, hostname, uptime
- `GetCPUUsage()` - CPU usage percentage
- `GetMemoryUsage()` - Memory utilization
- `GetDiskSpace(path)` - Disk space by path
- `GetRunningProcesses(count)` - Top processes
- `PingHost(host, count)` - Network connectivity
- `CheckServiceStatus(service)` - Service status
- `ResolveDNS(hostname)` - DNS resolution
- And more...

### Running

```bash
cd examples/it-support
go run ./cmd/main.go

# Try queries:
# "Máy chậm lắm" (Machine is slow)
# "Kiểm tra internet" (Check internet)
# "Dịch vụ nào đang chạy?" (What services are running?)
```

### Directory Structure

```
examples/it-support/
├── cmd/
│   └── main.go                # Entry point
├── internal/
│   ├── agents/                # Agent definitions
│   ├── tools/                 # Tool implementations
│   └── crew/                  # Crew orchestration
├── config/
│   ├── agents/                # Agent YAML configs
│   │   ├── orchestrator.yaml
│   │   ├── clarifier.yaml
│   │   └── executor.yaml
│   └── crew.yaml              # Crew configuration
└── README.md
```

## Building Your Own Example

### Step 1: Define Agents

Create `config/agents/my-agent.yaml`:

```yaml
id: my-agent
name: My Agent
role: Specialized role
backstory: Background and expertise
model: gpt-4o-mini
provider: openai
temperature: 0.7
tools:
  - tool1
  - tool2
```

### Step 2: Define Tools

Create tools in `internal/tools/`:

```go
package tools

import (
	"context"
	crewai "github.com/taipm/go-crewai"
)

func CreateMyTool() *crewai.Tool {
	return &crewai.Tool{
		Name:        "MyTool",
		Description: "Does something useful",
		Parameters: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"param": map[string]interface{}{
					"type":        "string",
					"description": "Input parameter",
				},
			},
			"required": []string{"param"},
		},
		Handler: func(ctx context.Context, args map[string]interface{}) (string, error) {
			// Implementation
			return "Result", nil
		},
	}
}
```

### Step 3: Create Crew

Create crew configuration in `config/crew.yaml`:

```yaml
agents:
  - my-agent

tasks:
  - id: main-task
    description: "Main task"

max_rounds: 10
max_handoffs: 5

routing:
  signals:
    my-agent:
      - signal: "[ROUTE_OTHER]"
        target: other-agent
```

### Step 4: Implement Main

Create `cmd/main.go`:

```go
package main

import (
	"context"
	"fmt"
	"log"
	"os"
	
	crewai "github.com/taipm/go-crewai"
)

func main() {
	// Load configuration
	crew, err := crewai.LoadCrewConfig("config/crew.yaml")
	if err != nil {
		log.Fatal(err)
	}

	// Create executor
	apiKey := os.Getenv("OPENAI_API_KEY")
	executor := crewai.NewCrewExecutor(crew, apiKey)

	// Execute task
	ctx := context.Background()
	result, err := executor.Execute(ctx, "Your task here")
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Result: %s\n", result.Content)
}
```

## Streaming Example

For real-time response streaming:

```go
responseChan := make(chan *crewai.Response)

go func() {
	if err := executor.ExecuteStream(ctx, "task", responseChan); err != nil {
		log.Fatal(err)
	}
}()

for response := range responseChan {
	if response.Content != "" {
		fmt.Printf("[%s]: %s\n", response.AgentName, response.Content)
	}
	if response.Signal != "" {
		fmt.Printf("  → Signal: %s\n", response.Signal)
	}
}
```

## Provider Configuration Examples

### Using Ollama Locally

In `config/agents/my-agent.yaml`:

```yaml
model: deepseek-r1:1.5b
provider: ollama
provider_url: http://localhost:11434
```

### Using OpenAI

In `config/agents/my-agent.yaml`:

```yaml
model: gpt-4o-mini
provider: openai
```

Requires `OPENAI_API_KEY` environment variable.

## Testing Your Example

```bash
# Build
go build ./cmd

# Test with input
./main < test_input.txt

# Run with streaming
go run ./cmd/main.go --stream

# HTTP server mode
go run ./cmd/main.go --server --port 8080
```

## More Examples Coming

- Customer Service Agent
- Research Agent
- Data Analysis Agent
- Content Generation Agent

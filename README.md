# Go-Agentic: Multi-Agent Framework in Go

A production-ready framework for building multi-agent AI systems in Go, featuring signal-based agent routing, streaming support, and comprehensive tooling.

## Project Structure

```plaintext
go-agentic/
├── core/                      # Core library
│   ├── types.go               # Core data types (Agent, Crew, Tool, etc.)
│   ├── agent.go               # Single agent execution
│   ├── crew.go                # Multi-agent orchestration
│   ├── config.go              # YAML configuration loading
│   ├── http.go                # HTTP API server
│   ├── streaming.go           # Server-Sent Events streaming
│   ├── html_client.go         # Web UI base template
│   ├── report.go              # HTML report generation
│   ├── tests.go               # Testing utilities
│   └── go.mod                 # Core library module
├── examples/                  # Example applications
│   ├── it-support/            # IT Support multi-agent system
│   │   ├── cmd/main.go
│   │   ├── internal/
│   │   ├── config/
│   │   └── README.md
│   └── README.md
└── README.md
```

## Overview

### Core Library (core)

A minimal but complete multi-agent framework providing:

- **Agent System**: Define agents with roles, tools, and models
- **Crew Orchestration**: Coordinate multiple agents with signal-based routing
- **Tool System**: Build extensible tool sets with context-aware execution
- **Configuration**: YAML-based agent and crew definitions
- **Streaming**: Real-time Server-Sent Events for agent interactions
- **HTTP API**: RESTful API for crew execution
- **Web UI**: HTML client interface for interaction

**Key Files:**
- `types.go` (84 lines): Core data structures
- `agent.go` (234 lines): Agent execution engine
- `crew.go` (398 lines): Orchestration and routing
- `config.go` (169 lines): Configuration loading
- `http.go` (187 lines): HTTP server
- `streaming.go` (54 lines): SSE streaming
- `html_client.go` (252 lines): Web UI
- `report.go` (696 lines): Report generation
- `tests.go` (316 lines): Testing utilities

**Total: 2,384 lines of pure framework code**

### Example Applications (examples)

#### IT Support System ✅ Complete

A fully functional multi-agent IT troubleshooting system with:

- **3 Specialized Agents**:
  - Orchestrator (My): Entry point and router
  - Clarifier (Ngân): Information gatherer
  - Executor (Trang): Technical expert

- **13 Diagnostic Tools**:
  - System info, CPU, Memory, Disk
  - Network (Ping, DNS, Network status)
  - Services, Processes
  - Advanced diagnostics and command execution

- **Signal-Based Routing**: Agents communicate via signals
- **Vietnamese Language**: Full Vietnamese support
- **YAML Configuration**: Agent and crew definitions

## Key Features

### 1. Signal-Based Agent Routing

Agents emit signals that determine handoff targets:

```go
Orchestrator: "[ROUTE_EXECUTOR]" → Executor
           or "[ROUTE_CLARIFIER]" → Clarifier
Clarifier:   "[KẾT THÚC]" → Executor
Executor:    Terminal (no handoff)
```

### 2. Extensible Tool System

Define tools with parameters and handlers:

```go
tool := &crewai.Tool{
    Name:        "CheckStatus",
    Description: "Check system status",
    Parameters:  map[string]interface{}{...},
    Handler: func(ctx context.Context, args map[string]interface{}) (string, error) {
        // Implementation
        return result, nil
    },
}
```

### 3. YAML Configuration

Configure agents and crews without recompiling:

```yaml
agents:
  - orchestrator
  - clarifier
  - executor

routing:
  signals:
    orchestrator:
      - signal: "[ROUTE_EXECUTOR]"
        target: executor
```

### 4. Real-Time Streaming

Server-Sent Events for live agent interactions:

```go
executor.ExecuteStream(ctx, task, streamChan)
```

### 5. Safety-First Design

- Context-aware execution (supports cancellation)
- Dangerous command blocking
- Parameter validation
- Type-safe APIs

## Getting Started

### Prerequisites

- Go 1.25.2 or later
- OpenAI API key

### Install Core Library

```bash
go get github.com/taipm/go-crewai
```

### Run IT Support Example

```bash
cd go-agentic-examples/it-support
export OPENAI_API_KEY=your_key_here
go run ./cmd/main.go
```

## Usage Example

### Creating a Crew

```go
package main

import (
    "github.com/taipm/go-crewai"
)

func main() {
    // Define agents
    agent := &crewai.Agent{
        ID:          "expert",
        Name:        "Expert",
        Role:        "Problem Solver",
        Backstory:   "An experienced problem solver",
        Model:       "gpt-4o-mini",
        Tools:       tools,
        Temperature: 0.7,
        IsTerminal:  true,
    }

    // Create crew
    crew := &crewai.Crew{
        Agents:    []*crewai.Agent{agent},
        MaxRounds: 10,
    }

    // Execute
    executor := crewai.NewCrewExecutor(crew, os.Getenv("OPENAI_API_KEY"))
    result, _ := executor.Execute(ctx, "Solve this problem...")
    fmt.Println(result.Content)
}
```

### Defining Tools

```go
tools := []*crewai.Tool{
    {
        Name:        "Calculator",
        Description: "Perform mathematical calculations",
        Parameters: map[string]interface{}{
            "type": "object",
            "properties": map[string]interface{}{
                "expression": map[string]interface{}{
                    "type": "string",
                    "description": "Math expression to evaluate",
                },
            },
            "required": []string{"expression"},
        },
        Handler: func(ctx context.Context, args map[string]interface{}) (string, error) {
            expr := args["expression"].(string)
            // Evaluate expression
            return result, nil
        },
    },
}
```

## Architecture

### Agent Execution Flow

1. User input → Crew
2. Crew routes to entry agent
3. Agent processes with tools
4. Agent emits signal
5. Crew routes based on signal
6. Next agent receives context
7. Continue until terminal agent
8. Return final response

### Tool Execution

1. Agent requests tool execution
2. Tool handler is called with context
3. Context-aware execution (supports cancellation)
4. Result returned to agent
5. Agent processes result
6. Agent continues or hands off

## Development

### Adding New Example

1. Create subdirectory in `go-agentic-examples/`
2. Create agent definitions
3. Create tool implementations
4. Create crew configuration
5. Create entry point (cmd/main.go)
6. Document in README.md

### Contributing

- Follow Go best practices
- Add tests for new features
- Update documentation
- Ensure backward compatibility

## Performance

- Minimal overhead for agent orchestration
- Efficient streaming support
- Context-aware execution with cancellation
- Support for concurrent requests
- Cross-platform support (Linux, macOS, Windows)

## Language Support

- **Core**: English
- **Examples**: Vietnamese (IT Support), extensible to other languages

## Error Handling

- Missing environment variables
- Network timeouts and failures
- Invalid parameters
- Tool execution errors
- Dangerous command blocking

## Testing

### Build Tests

```bash
# Test core library
cd go-crewai
go build ./...

# Test IT Support example
cd go-agentic-examples/it-support
go build ./cmd/main.go
```

### Run Tests

```bash
go test ./...
```

## Documentation

- [Core Library](./go-crewai/README.md)
- [Examples](./go-agentic-examples/README.md)
- [IT Support Example](./go-agentic-examples/it-support/README.md)

## Project Status

- ✅ **PHASE 1**: Backup & Prepare
- ✅ **PHASE 2**: Remove IT Code from Core
- ✅ **PHASE 3**: Create Examples Package
- ✅ **PHASE 4**: Move IT Support Code
- ✅ **PHASE 5**: Update go.mod Files
- ✅ **PHASE 6**: Test & Verify
- ✅ **PHASE 7**: Documentation
- ⏳ **PHASE 8**: Final Commit

## License

Part of the go-agentic project.

## Next Steps

- Implement additional examples (Customer Service, Research, Data Analysis)
- Add gRPC support for distributed systems
- Create CLI tools for agent management
- Build monitoring and logging systems
- Add advanced scheduling capabilities

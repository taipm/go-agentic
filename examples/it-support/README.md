# IT Support Example

A complete example of using go-agentic for IT support and system diagnostics.

## Overview

This example demonstrates a multi-agent system that handles IT support requests:
- **Orchestrator**: Analyzes incoming requests
- **Clarifier**: Gathers additional information if needed
- **Executor**: Performs system diagnostics and troubleshooting

## Features

- 12+ system diagnostic tools
- Multi-round tool execution
- Interactive CLI interface
- Real-time streaming responses

## Quick Start

### 1. Setup

```bash
cd examples/it-support
cp ../.env.example .env
# Edit .env and add your OPENAI_API_KEY
```

### 2. Run

```bash
go run main.go
# Or: go run main.go --server --port 8081
```

### 3. Try It

Example requests:
- "Check my system health"
- "Why is my CPU high?"
- "Is nginx running?"
- "What's my disk space?"

## Available Tools

- **GetCPUUsage** - Current CPU percentage
- **GetMemoryUsage** - Memory utilization
- **GetDiskSpace** - Disk usage by path
- **GetSystemInfo** - OS, hostname, uptime
- **GetRunningProcesses** - Top processes
- **PingHost** - Network connectivity
- **CheckServiceStatus** - Service health
- **ResolveDNS** - DNS resolution
- **ExecuteCommand** - Run shell commands
- **GetSystemDiagnostics** - Full diagnostics

## Configuration

Edit `config/crew.yaml` to customize:
- Agent roles and behaviors
- Routing logic
- Tool availability
- Max rounds and handoffs

## Web Interface

Run with server mode to use the web UI:

```bash
go run main.go --server --port 8081
```

Then open: http://localhost:8081

## Testing

```bash
go run main.go test
```

This runs the built-in test suite against the IT support agents.

## Customization

To add new tools:

1. Implement a tool handler function
2. Add it to `createITSupportTools()`
3. Update agent configuration
4. Test with the CLI

Example:

```go
func customToolHandler(ctx context.Context, args map[string]interface{}) (string, error) {
    // Your tool logic here
    return "result", nil
}
```

## Architecture

```
IT Support Crew
├── Orchestrator (entry point)
├── Clarifier (info gathering)
└── Executor (with tools)
    ├── System Info Tools
    ├── Network Tools
    ├── Service Tools
    └── Diagnostic Tools
```

## Learn More

- See parent directory README for general go-agentic information
- Check `config/crew.yaml` for routing configuration
- Review `example_it_support.go` for tool implementations

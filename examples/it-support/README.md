# IT Support System Example

A complete multi-agent IT support system demonstrating the go-crewai library. This example implements an intelligent IT troubleshooting system with three specialized agents working together.

## System Overview

The IT Support System uses a three-agent architecture with signal-based routing:

1. **Orchestrator (My)** - Entry point agent that analyzes IT issues and routes them appropriately
2. **Clarifier (Ngân)** - Information gatherer that asks clarifying questions when details are unclear
3. **Executor (Trang)** - Technical expert that diagnoses issues using system diagnostic tools

## Architecture

```
User Input
    ↓
Orchestrator (Entry Point)
    ├─ [ROUTE_EXECUTOR] → Executor (if sufficient info)
    └─ [ROUTE_CLARIFIER] → Clarifier (if more info needed)
         ↓
      Clarifier (gathers info)
         ↓
      [KẾT THÚC] → Executor
         ↓
      Executor (runs diagnostics) → Terminal Response
```

## Features

### 13 Diagnostic Tools

The Executor agent has access to comprehensive system diagnostic tools:

**System Information:**
- `GetSystemInfo()` - OS, hostname, uptime
- `GetCPUUsage()` - Current CPU usage percentage
- `GetMemoryUsage()` - Memory consumption
- `GetDiskSpace(path)` - Disk usage for specific paths

**Advanced Diagnostics:**
- `CheckMemoryStatus()` - Detailed memory information
- `CheckDiskStatus(path)` - Detailed disk usage with percentages
- `GetRunningProcesses(count)` - Top running processes
- `GetSystemDiagnostics()` - Comprehensive system diagnostics report

**Network Tools:**
- `PingHost(host, count)` - Test connectivity
- `ResolveDNS(hostname)` - Hostname to IP resolution
- `CheckNetworkStatus(host, count)` - Network connectivity verification

**Service Management:**
- `CheckServiceStatus(service)` - Service status check
- `ExecuteCommand(command)` - Execute arbitrary shell commands (with safety checks)

### Safety Features

- Dangerous command blocking (prevents `rm -rf`, `mkfs`, `dd if=`, etc.)
- Context-aware execution (supports cancellation)
- Parameter validation
- Cross-platform support (Linux/macOS)

## Project Structure

```
it-support/
├── cmd/
│   └── main.go                 # Entry point
├── internal/
│   ├── crew.go                 # Crew definition and agent setup
│   └── tools.go                # Tool implementations (13 tools)
├── config/
│   ├── crew.yaml               # Crew configuration with routing
│   └── agents/
│       ├── orchestrator.yaml    # Orchestrator agent config
│       ├── clarifier.yaml       # Clarifier agent config
│       └── executor.yaml        # Executor agent config
├── go.mod                       # Module definition
├── .env.example                 # Environment variables template
└── README.md                    # This file
```

## Getting Started

### Prerequisites

- Go 1.25.2 or later
- OpenAI API key

### Setup

1. Clone the repository and navigate to the IT support example:
```bash
cd go-agentic-examples/it-support
```

2. Copy and configure environment variables:
```bash
cp .env.example .env
# Edit .env and add your OPENAI_API_KEY
```

3. Install dependencies:
```bash
go mod tidy
```

### Running the Application

#### CLI Mode (Interactive)

```bash
export OPENAI_API_KEY=your_key_here
go run ./cmd/main.go
```

You'll see:
```
=== IT Support System (CLI) ===
Describe your IT issue:
```

Enter your IT issue description. Examples:
- "Bạn tự lấy thông tin máy hiện tại" (Check my machine)
- "Check localhost" (Diagnose local system)
- "Server 192.168.1.50 không ping được" (Server not responding)
- "CPU cao" (High CPU usage)

#### Web Server Mode with Real-time Streaming

```bash
export OPENAI_API_KEY=your_key_here
go run ./cmd/main.go --server --port 8081
```

Then open browser to:

- **Web Client**: [http://localhost:8081](http://localhost:8081)
- **SSE Endpoint**: [http://localhost:8081/api/crew/stream](http://localhost:8081/api/crew/stream)

The web client includes:

- **Real-time SSE streaming** of agent interactions
- **Preset scenarios** for quick testing (Máy chậm, Không Internet, Network lỗi, Disk đầy, RAM cao)
- **Conversation history** management
- **Live event visualization** with color-coded event types

**Flags:**

- `--server`: Enable HTTP server mode (default: CLI mode)
- `--port N`: Set server port (default: 8081)

## Usage Examples

### Example 1: Auto-diagnose Local Machine
```
User: "Bạn tự lấy thông tin máy hiện tại"
Orchestrator: Routes to Executor immediately (detected localhost keyword)
Executor: Runs GetSystemDiagnostics() and provides analysis
```

### Example 2: Remote Server Issue
```
User: "Server 192.168.1.50 không ping được"
Orchestrator: Routes to Executor (has specific IP and issue)
Executor: Pings host and provides diagnostic results
```

### Example 3: Vague Issue
```
User: "Máy tính của tôi chậm"
Orchestrator: Routes to Clarifier (needs more details)
Clarifier: Asks clarifying questions about the issue
User: Provides more details
Clarifier: Routes to Executor after gathering sufficient info
Executor: Runs diagnostics and recommends solutions
```

## Configuration

### Agent Configuration Files

Each agent is configured via YAML:
- **Orchestrator** (`orchestrator.yaml`): Routing rules and pattern matching
- **Clarifier** (`clarifier.yaml`): Information gathering protocol
- **Executor** (`executor.yaml`): Tool access and diagnostics procedures

### Crew Configuration

The `crew.yaml` defines:
- Entry point (orchestrator)
- Maximum handoffs between agents (5)
- Maximum conversation rounds (10)
- Signal-based routing rules
- Agent-specific behaviors

## Development

### Adding New Tools

To add a new diagnostic tool:

1. Implement the tool function in `internal/tools.go`:
```go
func myNewTool(ctx context.Context, args map[string]interface{}) (string, error) {
    // Implementation
    return result, nil
}
```

2. Add the tool to `createITSupportTools()`:
```go
{
    Name:        "MyNewTool",
    Description: "Tool description",
    Parameters: map[string]interface{}{
        "type":       "object",
        "properties": map[string]interface{}{
            // Define parameters
        },
    },
    Handler: myNewTool,
}
```

3. Update `executor.yaml` to include the new tool in the tools list

### Modifying Agent Behavior

Edit the corresponding YAML file:
- Change agent prompts in `system_prompt` field
- Adjust temperature for different response styles
- Update routing rules in `crew.yaml`

## Testing

Build the example:
```bash
go build ./cmd/main.go
```

Run tests:
```bash
go test ./...
```

## Language

The IT Support System is configured to work primarily in **Vietnamese (Tiếng Việt)**. All agent prompts, system instructions, and example dialogues use Vietnamese.

## Error Handling

The system handles:
- Missing OPENAI_API_KEY environment variable
- Network timeouts and failures
- Invalid command parameters
- Dangerous command attempts
- Tool execution errors

## Performance

- Minimal latency with streaming support
- Efficient tool execution with context cancellation
- Scalable multi-agent orchestration
- Support for concurrent requests

## License

Part of the go-agentic project. See LICENSE for details.

## Related Files

- Core Library: [../../core/README.md](../../core/README.md)
- Examples Overview: [../README.md](../README.md)
- Main README: [../../README.md](../../README.md)

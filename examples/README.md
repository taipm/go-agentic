# 🚀 Go-Agentic Examples

This package contains complete example applications demonstrating how to use the go-agentic core library for different domains.

## Examples

### 1. IT Support System ✅ Complete

Multi-agent system for IT troubleshooting and system diagnostics with intelligent routing.
- **Location**: `it-support/`
- **Agents**:
  - Orchestrator (My) - Entry point and router
  - Clarifier (Ngân) - Information gatherer
  - Executor (Trang) - Technical expert with diagnostic tools
- **Tools**: 13 diagnostic tools (CPU, Memory, Disk, Network, etc.)
- **Language**: Vietnamese
- **How to Run**:

  ```bash
  cd it-support
  export OPENAI_API_KEY=your_key_here
  go run ./cmd/main.go
  ```

- **Documentation**: [it-support/README.md](it-support/README.md)

### 2. Customer Service System (Coming Soon)

Multi-agent system for customer support ticket management.

- **Location**: `customer-service/`
- **Agents**: Intake, Knowledge Base, Resolution
- **Tools**: CRM, Ticket System, FAQ Search

### 3. Research Assistant System (Coming Soon)

Multi-agent system for research and information synthesis.

- **Location**: `research-assistant/`
- **Agents**: Researcher, Analyst, Writer
- **Tools**: Web Search, Paper Analysis, Citation

### 4. Data Analysis System (Coming Soon)

Multi-agent system for data processing and visualization.

- **Location**: `data-analysis/`
- **Agents**: Loader, Analyzer, Visualizer
- **Tools**: Data Processing, Analysis, Chart Generation

## Project Structure

```plaintext
examples/
├── it-support/
│   ├── cmd/
│   ├── internal/
│   ├── config/
│   ├── go.mod
│   └── README.md
├── go.mod
└── README.md (this file)
```

## Getting Started

### Prerequisites

- Go 1.25.2 or later
- OPENAI_API_KEY environment variable set

### Setup

1. Navigate to an example directory:

   ```bash
   cd it-support
   ```

2. Install dependencies:

   ```bash
   go mod tidy
   ```

3. Set your OpenAI API key:

   ```bash
   export OPENAI_API_KEY=your_key_here
   ```

4. Run the example:

   ```bash
   go run ./cmd/main.go
   ```

## Documentation

- **IT Support Example**: [it-support/README.md](it-support/README.md)
- **Core Library**: [../core/README.md](../core/README.md)
- **Main Project**: [../README.md](../README.md)

## Key Features of Examples

- **Multi-Agent Orchestration**: Agents work together with signal-based routing
- **Domain-Specific Tools**: Each example comes with specialized tools
- **Type Safety**: Built with Go's strong type system
- **Streaming Support**: Real-time event streaming for agent interactions
- **Error Handling**: Comprehensive error handling and validation
- **Configuration**: YAML-based agent and crew configuration

## Development Tips

1. Each example demonstrates best practices for multi-agent systems
2. Tools have built-in safety checks to prevent dangerous operations
3. Agents use specialized system prompts for their roles
4. Configuration files (YAML) can be modified without recompiling

## Testing

Each example can be tested by running:

```bash
go build ./cmd/main.go
```

Or run with specific inputs to verify functionality.

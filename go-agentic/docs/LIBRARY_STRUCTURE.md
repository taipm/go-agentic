# go-agentic Library Structure

Professional, clean architecture for the go-agentic multi-agent orchestration library.

## Directory Layout

```
go-agentic/
│
├── Core Library Files (Package Root)
│   ├── types.go              # Core data structures
│   │   └── Types: Agent, Crew, Tool, ToolCall, Message, StreamEvent
│   │              AgentResponse, CrewResponse, CrewConfig, AgentConfig
│   │
│   ├── agent.go              # Agent execution logic
│   │   └── Agent: ID, Name, Role, Model, Tools, Temperature, IsTerminal
│   │           Functions: Execute(), ExecuteWithTools()
│   │
│   ├── crew.go               # Crew orchestration engine
│   │   └── Crew: Agents, MaxRounds, MaxHandoffs, Routing
│   │            Functions: orchestrate agent collaboration
│   │
│   ├── config.go             # Configuration management
│   │   └── Load YAML configs for crew and agents
│   │            CreateAgentFromConfig(), LoadCrewConfig()
│   │
│   ├── tests.go              # Test utilities and scenarios
│   │   └── GetTestScenarios(), RunTestScenario(), HTMLReport
│   │
│   ├── report.go             # Test report generation
│   │   └── HTMLReport: Generate HTML test reports
│   │
│   ├── http.go               # HTTP server implementation
│   │   └── StartHTTPServer(), API endpoints, request handling
│   │
│   ├── html_client.go        # Web UI client
│   │   └── HTML/CSS for interactive testing
│   │
│   ├── streaming.go          # SSE streaming utilities
│   │   └── FormatStreamEvent(), SendStreamEvent()
│   │           NewStreamEvent(), NewStreamEventWithMetadata()
│   │
│   ├── go.mod                # Go module definition
│   │   └── module: github.com/taipm/go-agentic
│   │       go: 1.25.2
│   │       requires: openai-go/v3, yaml.v3
│   │
│   └── go.sum                # Dependency checksums
│
├── docs/                     # Documentation
│   ├── ARCHITECTURE.md       # System design and architecture
│   ├── LIBRARY_USAGE.md      # Library usage guide
│   ├── LIBRARY_INTRO.md      # Introduction for developers
│   ├── DEPLOYMENT_CHECKLIST.md
│   ├── MIGRATION_GUIDE.md    # Migration from previous versions
│   ├── DEMO_EXAMPLES.md
│   ├── DEMO_QUICK_START.md
│   ├── DEMO_README.md
│   └── (other documentation)
│
├── _old_files/               # Legacy artifacts (for reference)
│   ├── Binary builds (crewai, crewai-example, etc.)
│   ├── Test reports (.html files)
│   ├── Old shell scripts (demo.sh, test_streaming.sh)
│   └── .env examples
│
├── README.md                 # Library README
└── LIBRARY_STRUCTURE.md      # This file

```

## File Organization Guide

### Core Package (Root Level)

Files in the root directory of `go-agentic/` form the main package `agentic` and are directly importable:

```go
import "github.com/taipm/go-agentic"

// Usage:
crew := &agentic.Crew{ ... }
agent := &agentic.Agent{ ... }
tool := &agentic.Tool{ ... }
```

### Public API Surface

#### Types (types.go)
- `Agent` - Represents an AI agent with tools and personality
- `Crew` - Orchestrates multiple agents
- `Tool` - Represents a callable tool/function
- `ToolCall` - Represents a tool invocation
- `Message` - Agent communication message
- `AgentResponse` - Agent response to a task
- `CrewResponse` - Final crew response
- `StreamEvent` - Streaming event for SSE
- Configuration types: `CrewConfig`, `AgentConfig`

#### Core Functions
- `NewCrewExecutor()` - Create executor for crew
- `(c *Crew) Execute()` - Execute crew on a task
- `(a *Agent) Execute()` - Execute single agent
- `(t *Tool) Call()` - Invoke a tool

#### Configuration
- `LoadCrewConfig()` - Load crew from YAML
- `LoadAgentConfigs()` - Load agent configs from directory
- `CreateAgentFromConfig()` - Create agent from config struct

#### HTTP Server
- `StartHTTPServer()` - Start HTTP server with SSE streaming
- RESTful API for crew execution with real-time streaming

#### Testing
- `GetTestScenarios()` - Get predefined test scenarios
- `RunTestScenario()` - Execute a test scenario
- `NewHTMLReport()` - Generate test report

#### Streaming
- `FormatStreamEvent()` - Format event for SSE
- `SendStreamEvent()` - Send event over HTTP
- `NewStreamEvent()` - Create new event
- `NewStreamEventWithMetadata()` - Create event with data

## Design Principles

### 1. Clean Architecture
- **Separation of Concerns**: Core orchestration, execution, streaming, and HTTP layers are separate
- **Interface-Based Design**: Tools are functions, not structs; agents delegate to tools
- **Composition Over Inheritance**: Crew contains Agents, Agents contain Tools

### 2. Single Responsibility
- **types.go** - Data structures only
- **agent.go** - Agent execution logic
- **crew.go** - Crew orchestration
- **http.go** - HTTP server concerns
- **streaming.go** - SSE implementation

### 3. Minimal External Dependencies
- OpenAI Go SDK (required for LLM)
- yaml.v3 (for configuration)
- Standard Go libraries

### 4. Backward Compatibility
- All core types and functions remain in package root
- No breaking changes to import paths
- Documentation location change doesn't affect code

## Building and Testing

### Building Examples
```bash
cd ../examples/[example-name]
go build -o [example-name]-example ./main.go ./example_*.go
```

### Library Tests
```bash
# From library root
go test ./...
```

### Import Paths
```go
// All imports use the same path
import "github.com/taipm/go-agentic"

// Types, functions, etc. are all in the agentic package
agent := &agentic.Agent{}
crew := &agentic.Crew{}
tool := &agentic.Tool{}
```

## Documentation Files

| File | Purpose |
|------|---------|
| `ARCHITECTURE.md` | System design, interaction patterns, decision rationale |
| `LIBRARY_USAGE.md` | Comprehensive usage guide with examples |
| `LIBRARY_INTRO.md` | Introduction for new developers |
| `DEPLOYMENT_CHECKLIST.md` | Production deployment checklist |
| `MIGRATION_GUIDE.md` | Migration from previous versions |
| Root `README.md` | Quick start, overview, examples |

## Best Practices for Library Users

### 1. Configuration-Driven Setup
```go
crewConfig := agentic.LoadCrewConfig("config/crew.yaml")
agentConfigs := agentic.LoadAgentConfigs("config/agents")
```

### 2. Tool Design
- Keep tools focused and single-purpose
- Tools should be stateless when possible
- Use context for cancellation and timeouts

### 3. Agent Configuration
```go
agent := &agentic.Agent{
    ID: "specialist",
    Name: "Specialist Agent",
    Role: "Perform specialized tasks",
    Model: "gpt-4o",
    Tools: tools,
    Temperature: 0.7,
    IsTerminal: false, // Set true for final agents
}
```

### 4. Crew Orchestration
```go
crew := &agentic.Crew{
    Agents: []*agentic.Agent{orchestrator, worker, responder},
    MaxRounds: 10,
    MaxHandoffs: 5,
}
executor := agentic.NewCrewExecutor(crew, apiKey)
response, err := executor.Execute(ctx, userRequest)
```

## Future Organization

As the library grows, consider:

1. **Internal Packages** (if private implementations become complex):
   ```
   internal/
   ├── executor/    # Core execution engine
   ├── llm/         # LLM provider abstraction
   └── router/      # Agent routing logic
   ```

2. **Sub-packages** (if new major features added):
   ```
   plugins/         # Optional extensions
   ├── slack/       # Slack integration
   ├── postgres/    # Database backend
   └── langchain/   # LangChain integration
   ```

3. **Examples** (keep at root level):
   ```
   examples/        # Reference implementations
   ├── it-support/
   ├── customer-service/
   └── ...
   ```

## Clean Code Principles Applied

✅ **Minimal Files** - Only essential files in root
✅ **Clear Names** - File names match their primary type/concern
✅ **Documentation** - docs/ directory for comprehensive guides
✅ **No Clutter** - Old artifacts moved to _old_files/
✅ **Single Import** - All public API via one package root
✅ **Explicit Dependencies** - Clear go.mod with minimal deps
✅ **Professional Layout** - Follows Go community standards

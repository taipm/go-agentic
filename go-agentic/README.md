# go-agentic Library

Professional Go library for orchestrating multiple AI agents to solve complex problems collaboratively.

## Package Import

```go
import "github.com/taipm/go-agentic"
```

## Quick Start

```go
// Create agents
agent := &agentic.Agent{
    ID: "agent-1",
    Name: "Agent One",
    Role: "Primary agent",
    Model: "gpt-4o",
    Tools: tools,
}

// Create crew
crew := &agentic.Crew{
    Agents: []*agentic.Agent{agent},
    MaxRounds: 10,
}

// Execute
executor := agentic.NewCrewExecutor(crew, apiKey)
response, err := executor.Execute(context.Background(), "Your task here")
```

## Core Components

- **Agent**: Represents an AI agent with tools and personality
- **Crew**: Orchestrates multiple agents
- **Tool**: Represents a callable function/tool
- **Message**: Agent communication
- **StreamEvent**: Real-time event streaming

## Documentation

See the [docs/](docs/) directory for comprehensive guides:

- [LIBRARY_README.md](docs/LIBRARY_README.md) - Library overview
- [LIBRARY_STRUCTURE.md](docs/LIBRARY_STRUCTURE.md) - Architecture and design
- [ARCHITECTURE.txt](docs/ARCHITECTURE.txt) - Visual diagrams and flows

## Key Features

✅ Multi-agent orchestration  
✅ Real-time SSE streaming  
✅ Intelligent agent routing  
✅ Configuration-driven setup  
✅ Comprehensive testing framework  
✅ Production-ready error handling

## Dependencies

- `github.com/openai/openai-go/v3` - OpenAI SDK
- `gopkg.in/yaml.v3` - YAML configuration

## License

See LICENSE file in repository root.

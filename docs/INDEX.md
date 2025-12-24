# Documentation Index

## Quick Navigation

- **New to go-agentic?** → [Getting Started](01-GETTING_STARTED.md)
- **Want to understand how it works?** → [Core Concepts](02-CORE_CONCEPTS.md)
- **Building your own system?** → [API Reference](03-API_REFERENCE.md) + [Examples](04-EXAMPLES.md)
- **Deploying to production?** → [Deployment Guide](05-DEPLOYMENT.md)
- **Using different LLM providers?** → [Provider Guide](PROVIDER_GUIDE.md)

## Documentation Structure

### 📚 Essential Guides

| Document | Purpose | Audience |
|----------|---------|----------|
| [Getting Started](01-GETTING_STARTED.md) | 5-minute quick start | Everyone |
| [Core Concepts](02-CORE_CONCEPTS.md) | Understanding architecture | Developers |
| [API Reference](03-API_REFERENCE.md) | Complete API documentation | Developers |
| [Examples](04-EXAMPLES.md) | Working code examples | Developers |
| [Deployment Guide](05-DEPLOYMENT.md) | Production deployment | DevOps/Ops |
| [Provider Guide](PROVIDER_GUIDE.md) | LLM provider setup | Everyone |

### 🚀 What You Can Do

#### For Development
1. Start with [Getting Started](01-GETTING_STARTED.md)
2. Understand [Core Concepts](02-CORE_CONCEPTS.md)
3. Use [API Reference](03-API_REFERENCE.md) as you code
4. Check [Examples](04-EXAMPLES.md) for patterns

#### For Production
1. Read [Deployment Guide](05-DEPLOYMENT.md)
2. Configure providers using [Provider Guide](PROVIDER_GUIDE.md)
3. Set up security and monitoring
4. Load test your application

#### For Contributing
1. Read all essential guides
2. Follow Go best practices
3. Add tests for new features
4. Update documentation

## Directory Structure

```
docs/
├── INDEX.md                    # This file
├── 01-GETTING_STARTED.md      # Quick start guide
├── 02-CORE_CONCEPTS.md        # Architecture and concepts
├── 03-API_REFERENCE.md        # Complete API documentation
├── 04-EXAMPLES.md             # Working examples
├── 05-DEPLOYMENT.md           # Production deployment
└── PROVIDER_GUIDE.md          # LLM provider setup
```

## Core Library Structure

```
core/
├── types.go                    # Core data types (Agent, Crew, Tool, Response)
├── agent.go                    # Single agent execution
├── crew.go                     # Multi-agent orchestration
├── crew_routing.go            # Signal-based routing
├── crew_tools.go              # Tool execution
├── config.go                  # YAML configuration loading
├── http.go                    # HTTP API server
├── streaming.go               # Server-Sent Events
├── html_client.go             # Web UI
├── report.go                  # HTML report generation
├── defaults.go                # Default configurations
├── metadata_logging.go        # Request metadata tracking
├── request_tracking.go        # Request lifecycle tracking
├── shutdown.go                # Graceful shutdown
├── validation.go              # Input validation
├── providers/                 # LLM provider implementations
│   ├── openai.go
│   └── ollama.go
├── tools/                     # Built-in tools
└── tests.go                   # Testing utilities
```

## Examples Structure

```
examples/
└── it-support/                # Complete IT support multi-agent system
    ├── cmd/main.go           # Entry point
    ├── internal/             # Internal implementation
    │   ├── agents/          # Agent definitions
    │   ├── tools/           # Tool implementations
    │   └── crew/            # Crew setup
    ├── config/              # Configuration files
    │   ├── agents/          # Agent configs (YAML)
    │   └── crew.yaml        # Crew configuration
    └── README.md
```

## Key Concepts at a Glance

### Agent
An autonomous entity with a role, tools, and decision-making capability.

### Crew
A collection of agents working together, coordinated by the executor.

### Tool
A capability agents can use - with parameters, description, and handler.

### Signal
A keyword or pattern that triggers agent-to-agent handoffs.

### Executor
The orchestration engine that manages agent execution and routing.

## Common Tasks

### Task: Create a New Agent
→ [API Reference: Creating Agents Programmatically](03-API_REFERENCE.md#creating-agents-programmatically)

### Task: Create a Custom Tool
→ [API Reference: Creating Tools](03-API_REFERENCE.md#creating-tools)

### Task: Deploy to Production
→ [Deployment Guide](05-DEPLOYMENT.md)

### Task: Use Different LLM Provider
→ [Provider Guide](PROVIDER_GUIDE.md)

### Task: Stream Responses
→ [API Reference: Streaming Results](03-API_REFERENCE.md#streaming-results)

### Task: Build Complete System
→ [Examples: Building Your Own Example](04-EXAMPLES.md#building-your-own-example)

## Support

- **Questions?** Check the [API Reference](03-API_REFERENCE.md)
- **Need examples?** See [Examples](04-EXAMPLES.md)
- **Production issues?** Check [Deployment Guide](05-DEPLOYMENT.md)
- **Code contributions?** Follow [Core Concepts](02-CORE_CONCEPTS.md) first

## Version Information

- **Go:** 1.25.2+
- **Latest Release:** See [README.md](../README.md)
- **Status:** Production Ready

---

**Last Updated**: 2025-12-23

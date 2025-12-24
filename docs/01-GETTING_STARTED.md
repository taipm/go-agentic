# Getting Started with go-agentic

## Prerequisites

- **Go 1.25.2** or later
- One of these LLM providers:
  - **Ollama** (local, free) - [Download](https://ollama.com)
  - **OpenAI API key** (cloud, paid) - [Get Key](https://platform.openai.com/api-keys)

## Quick Start (5 Minutes)

### Option 1: Using Ollama (Recommended for Development)

```bash
# 1. Install and start Ollama server (in one terminal)
# Download from https://ollama.com
ollama serve

# 2. Pull a model (in another terminal)
ollama pull deepseek-r1:1.5b

# 3. Run IT Support Example
cd examples/it-support
go run ./cmd/main.go
```

### Option 2: Using OpenAI

```bash
# 1. Set your API key
export OPENAI_API_KEY=sk-xxx...

# 2. Update example config to use OpenAI
# Edit examples/it-support/config/agents/*.yaml
# Change: provider: ollama to provider: openai
# Change: model: deepseek-r1:1.5b to model: gpt-4o-mini

# 3. Run IT Support Example
cd examples/it-support
go run ./cmd/main.go
```

## Install Core Library

```bash
go get github.com/taipm/go-crewai
```

## Basic Usage

```go
package main

import (
	"context"
	"fmt"
	"github.com/taipm/go-crewai"
)

func main() {
	// Define your agent
	agent := &crewai.Agent{
		ID:        "expert",
		Name:      "Expert",
		Role:      "Problem Solver",
		Backstory: "An experienced problem solver",
		Model:     "gpt-4o-mini",
		Tools:     []*crewai.Tool{}, // Add your tools here
	}

	// Create crew
	crew := &crewai.Crew{
		Agents:    []*crewai.Agent{agent},
		MaxRounds: 10,
	}

	// Execute
	executor := crewai.NewCrewExecutor(crew, "your-api-key")
	result, err := executor.Execute(context.Background(), "Solve this problem...")
	if err != nil {
		panic(err)
	}

	fmt.Printf("Result: %s\n", result.Content)
}
```

## Next Steps

- Read [Core Concepts](02-CORE_CONCEPTS.md)
- Check [API Reference](03-API_REFERENCE.md)
- Explore [Examples](04-EXAMPLES.md)
- Review [Deployment Guide](05-DEPLOYMENT.md)

# Core Concepts

## Architecture Overview

```
User Input
   ↓
Crew Executor
   ↓
Agent Router (signal-based)
   ↓
Entry Agent
   ├─ Process with tools
   ├─ Emit signal
   ├─ Receive context
   └─ Hand off to next agent
   ↓
Continues until terminal agent
   ↓
Final Response
```

## Key Components

### 1. Agent

An agent is an autonomous entity that:
- Has a role and expertise (defined in `Agent.Role` and `Agent.Backstory`)
- Processes tasks using available tools
- Emits signals to determine next agent
- Is either terminal (last in chain) or non-terminal (continues routing)

**Example**:
```
Orchestrator: Routes to appropriate specialist
Clarifier: Gathers information
Executor: Performs technical work (terminal)
```

### 2. Crew

A crew is a collection of agents that:
- Work together to solve complex problems
- Are routed through signal-based handoffs
- Have safety limits (MaxRounds, MaxHandoffs)
- Share context across interactions

**Key settings**:
- `MaxRounds`: Maximum tool execution iterations (default: 10)
- `MaxHandoffs`: Maximum agent-to-agent handoffs (default: 5)

### 3. Tool

A tool is a capability that agents can use:
- Has name, description, and parameter schema
- Executes via a handler function
- Returns results to the agent
- Supports context-aware execution (cancellation, timeouts)

**Example**:
```go
tool := &Tool{
	Name:        "GetCPU",
	Description: "Get CPU usage",
	Parameters:  map[string]interface{}{...},
	Handler: func(ctx context.Context, args map[string]interface{}) (string, error) {
		return "CPU: 45%", nil
	},
}
```

## Execution Flow

### Single Round

1. Agent receives task + context
2. Agent decides to use tool or respond
3. Tool executes with parameters
4. Result returned to agent
5. Agent processes result

### Multi-Round (Complete Feedback Loop)

Agents see tool results and can:
- Use another tool
- Refine previous requests
- Continue analysis
- Eventually respond

This continues until MaxRounds or agent decides to stop.

### Handoff

When agent emits signal matching routing rule:
1. Crew captures signal
2. Looks up target agent
3. Creates new context with previous agent's output
4. Routes to target agent
5. Target agent processes with full context

## Signal-Based Routing

Agents emit signals (keywords or patterns) that trigger routing:

```yaml
routing:
  signals:
    orchestrator:
      - signal: "[ROUTE_EXECUTOR]"
        target: executor
      - signal: "[ROUTE_CLARIFIER]"
        target: clarifier
```

When agent output contains signal, crew routes automatically.

## Multi-Provider LLM Support

The framework abstracts LLM provider differences:
- **Ollama**: Local, free, open-source models
- **OpenAI**: Cloud-based, production-grade models

Switch providers via YAML configuration without code changes.

## Safety Mechanisms

1. **MaxRounds**: Prevents infinite loops in tool usage
2. **MaxHandoffs**: Prevents infinite agent handoffs
3. **Terminal Agents**: Guaranteed final agents (no further handoffs)
4. **Input Validation**: Parameters validated before tool execution

## Context Preservation

Each handoff includes:
- Previous agent's response
- Original user input
- Conversation history (if applicable)
- State from previous interactions

This ensures agents have full context for decision-making.

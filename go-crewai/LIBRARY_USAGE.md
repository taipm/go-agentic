# Go-CrewAI Library Usage Guide

Go-CrewAI is a **library-quality, production-ready multi-agent framework** designed for building sophisticated AI-powered workflows. This guide demonstrates how other teams can integrate go-crewai into their own projects without being tied to the IT Support domain.

## Table of Contents

1. [Overview](#overview)
2. [Architecture](#architecture)
3. [Getting Started](#getting-started)
4. [Configuration](#configuration)
5. [Building Custom Workflows](#building-custom-workflows)
6. [Advanced Usage](#advanced-usage)
7. [Best Practices](#best-practices)
8. [Examples](#examples)

## Overview

Go-CrewAI uses a **hybrid approach** for agent orchestration:

- **Config-Driven**: Agent behavior, tools, and routing rules are defined in YAML configuration files
- **LLM-Powered**: Custom system prompts allow LLMs to make intelligent routing decisions
- **Signal-Based**: Agents use explicit signals (e.g., `[ROUTE_NEXT]`, `[COMPLETE]`) to indicate handoffs

This design makes go-crewai suitable for ANY multi-agent workflow, not just IT support.

## Architecture

### Core Components

```
CrewExecutor
├── Crew
│   ├── Agents[]
│   │   ├── Agent (Orchestrator, Clarifier, Executor, etc.)
│   │   │   ├── Tools[]
│   │   │   ├── SystemPrompt (Custom LLM instructions)
│   │   │   └── HandoffTargets[] (Routing options)
│   │   └── ...
│   └── Routing
│       ├── Signals (Define available signals per agent)
│       ├── Defaults (Fallback routing)
│       └── AgentBehaviors (Wait-for-signal, auto-route settings)
```

### Execution Flow

1. **User Input** → **Entry Agent** (typically orchestrator)
2. **Agent Response** with optional tool calls and routing signal
3. **Signal Detection** → Route to next agent if signal present
4. **Repeat** until terminal agent or no routing signal

## Getting Started

### 1. Import the Library

```go
import "github.com/taipm/go-crewai"
```

### 2. Create Your Configuration Files

Create a `config/` directory with:
- `crew.yaml` - Main configuration with routing rules
- `agents/` subdirectory with individual agent YAML files
  - `agent1.yaml`
  - `agent2.yaml`
  - etc.

### 3. Basic Usage

```go
// Load configuration
crewConfig, _ := crewai.LoadCrewConfig("config/crew.yaml")
agentConfigs, _ := crewai.LoadAgentConfigs("config/agents")

// Get your custom tools
allTools := GetYourCustomTools() // map[string]*crewai.Tool

// Create agents from config
agents := createAgentsFromConfig(crewConfig, agentConfigs, allTools)

// Create crew with routing config
crew := &crewai.Crew{
    Agents:      agents,
    MaxRounds:   crewConfig.Settings.MaxRounds,
    MaxHandoffs: crewConfig.Settings.MaxHandoffs,
    Routing:     crewConfig.Routing, // Key: Load routing configuration
}

// Execute
executor := crewai.NewCrewExecutor(crew, apiKey)
response, _ := executor.Execute(context.Background(), userInput)
```

## Configuration

### crew.yaml Structure

```yaml
version: "1.0"
description: "Your Crew Description"
entry_point: agent1  # Starting agent

agents:
  - agent1
  - agent2
  - agent3

settings:
  max_handoffs: 5
  max_rounds: 10
  timeout_seconds: 300
  language: en
  organization: YourOrg

routing:
  # Define available signals for each agent
  signals:
    agent1:
      - signal: "[NEXT_ANALYSIS]"
        target: agent2
        description: "Route to analysis agent"
      - signal: "[NEXT_DECISION]"
        target: agent3
        description: "Route to decision agent"
    agent2:
      - signal: "[COMPLETE]"
        target: null
        description: "Analysis complete"

  # Fallback routing when no signal found
  defaults:
    agent1: agent2
    agent2: agent3

  # Agent-specific behaviors
  agent_behaviors:
    agent1:
      wait_for_signal: true
      auto_route: false
      description: "Waits for explicit signal"
    agent2:
      wait_for_signal: false
      auto_route: true
      description: "Automatically routes if no signal"
```

### Agent YAML Structure

```yaml
id: your_agent_id
name: "Agent Display Name"
description: "Agent description"
role: "Agent's role"

backstory: |
  Your agent's background and personality.
  This helps the LLM understand the agent's perspective.

model: gpt-4o-mini  # or any OpenAI model
temperature: 0.7

is_terminal: false  # true for final agents
tools:
  - tool_name_1
  - tool_name_2

handoff_targets:
  - next_agent_id
  - another_agent_id

system_prompt: |
  {{name}}: Your custom system prompt here.

  Template variables available:
  - {{name}}: Agent name
  - {{role}}: Agent role
  - {{description}}: Agent description
  - {{backstory}}: Agent backstory

  Important instructions for your agent:
  - Behavior guidelines
  - Decision criteria
  - Routing signals to emit
```

## Building Custom Workflows

### Example: Customer Support Workflow

Let's build a multi-tier customer support system:

**crew.yaml:**
```yaml
version: "1.0"
description: "Multi-tier Customer Support System"
entry_point: router

agents:
  - router
  - tier1
  - tier2
  - escalator

settings:
  max_handoffs: 10
  max_rounds: 15
  timeout_seconds: 300

routing:
  signals:
    router:
      - signal: "[ROUTE_TIER1]"
        target: tier1
      - signal: "[ROUTE_ESCALATE]"
        target: escalator
    tier1:
      - signal: "[SOLVED]"
        target: null
      - signal: "[ESCALATE]"
        target: tier2
    tier2:
      - signal: "[SOLVED]"
        target: null
      - signal: "[ESCALATE]"
        target: escalator
    escalator:
      - signal: "[COMPLETE]"
        target: null

  defaults:
    router: tier1
    tier1: tier2
    tier2: escalator

  agent_behaviors:
    router:
      wait_for_signal: true
    tier1:
      wait_for_signal: true
    tier2:
      wait_for_signal: true
    escalator:
      is_terminal: true
```

**agents/router.yaml:**
```yaml
id: router
name: "Router"
description: "Route customer requests to appropriate tier"
role: "Request Router"

backstory: |
  You are the first point of contact for customer support.
  Your job is to quickly categorize issues and route them appropriately.

model: gpt-4o-mini
temperature: 0.3
is_terminal: false

tools: []
handoff_targets:
  - tier1
  - escalator

system_prompt: |
  You are {{name}}.

  Analyze the customer's issue and determine if it's:
  1. A simple FAQ that tier1 can handle → [ROUTE_TIER1]
  2. A complex issue needing escalation → [ROUTE_ESCALATE]

  Always include exactly one of these signals.
```

### Creating Custom Tools

```go
// Define your tools
func GetCustomTools() []*crewai.Tool {
    return []*crewai.Tool{
        {
            Name:        "SearchKnowledgeBase",
            Description: "Search company knowledge base for articles",
            Parameters: map[string]interface{}{
                "type": "object",
                "properties": map[string]interface{}{
                    "query": map[string]string{"type": "string"},
                },
            },
            Handler: func(ctx context.Context, args map[string]interface{}) (string, error) {
                query := args["query"].(string)
                // Implement your search logic
                return searchKnowledgeBase(query)
            },
        },
        {
            Name:        "CreateTicket",
            Description: "Create a support ticket in the system",
            Parameters: map[string]interface{}{
                "type": "object",
                "properties": map[string]interface{}{
                    "subject": map[string]string{"type": "string"},
                    "priority": map[string]string{"type": "string"},
                },
            },
            Handler: func(ctx context.Context, args map[string]interface{}) (string, error) {
                subject := args["subject"].(string)
                priority := args["priority"].(string)
                // Implement your ticket creation logic
                return createTicket(subject, priority)
            },
        },
    }
}
```

## Advanced Usage

### Custom Agent Behaviors

Implement different routing strategies by setting `AgentBehaviors`:

```yaml
agent_behaviors:
  decision_maker:
    wait_for_signal: true   # Must have explicit signal
    auto_route: false       # Won't auto-route
    is_terminal: false

  fallback_analyzer:
    wait_for_signal: false  # Can auto-route even without signal
    auto_route: true        # Uses default routing if no signal
    is_terminal: false

  final_executor:
    is_terminal: true       # No handoff possible
```

### Multi-Signal Patterns

Define multiple signal paths for complex workflows:

```yaml
signals:
  analyzer:
    - signal: "[NEEDS_APPROVAL]"
      target: approver
    - signal: "[NEEDS_IMPLEMENTATION]"
      target: implementer
    - signal: "[NEEDS_REVIEW]"
      target: reviewer
    - signal: "[COMPLETE]"
      target: null
```

### Custom Signals with Metadata

Signals can include additional context (in future versions):

```go
// Agents can include structured data in responses
// Example: "Analysis complete. [ROUTE_NEXT:priority=high]"
// Framework can parse and use this metadata
```

## Best Practices

### 1. Clear Signal Semantics

Use descriptive, unambiguous signal names:

```
✓ [ROUTE_TECHNICAL_TEAM]      # Clear intent
✓ [NEEDS_MANAGER_APPROVAL]    # Explicit action
✗ [NEXT]                       # Ambiguous
✗ [CONTINUE]                   # Not specific enough
```

### 2. Sensible Defaults

Configure fallback routing for safety:

```yaml
defaults:
  entry_agent: clarifier_agent
  clarifier_agent: executor_agent
  # Ensures workflow completes even without explicit signals
```

### 3. Temperature Tuning

Use lower temperature for decision agents, higher for creative agents:

```yaml
# Decision/Routing Agent
temperature: 0.3  # More deterministic

# Creative/Analytical Agent
temperature: 0.7  # More varied responses

# Brainstorming Agent
temperature: 0.9  # Maximum creativity
```

### 4. Tool Organization

Group related tools logically:

```go
// Good: Organized by function
GetAnalyticsTools()
GetApprovalTools()
GetImplementationTools()
GetReportingTools()

// Bad: Random grouping
GetToolsSet1()
GetToolsSet2()
```

### 5. Error Handling

Always handle tool execution errors gracefully:

```go
Handler: func(ctx context.Context, args map[string]interface{}) (string, error) {
    // Validate inputs
    if query, ok := args["query"].(string); !ok {
        return "", fmt.Errorf("invalid query parameter")
    }

    // Execute with error handling
    result, err := performOperation(query)
    if err != nil {
        return fmt.Sprintf("Error: %v", err), nil // Return error as string
    }

    return result, nil
}
```

## Examples

### Example 1: Content Review Workflow

```go
// agents/content_reviewer.yaml
system_prompt: |
  You review content for quality and compliance.

  Check for:
  1. Grammar and clarity
  2. Brand compliance
  3. Technical accuracy

  Emit signals:
  - [APPROVED] if content passes all checks
  - [NEEDS_REVISION] if issues found
  - [NEEDS_LEGAL] if legal implications detected
```

### Example 2: Data Processing Pipeline

```go
// agents/data_processor.yaml
system_prompt: |
  You process raw data input.

  Steps:
  1. Validate input format
  2. Transform to standard format
  3. Check for duplicates

  Emit:
  - [TRANSFORM_COMPLETE] when ready for analysis
  - [INVALID_DATA] if validation fails
```

### Example 3: Multi-Language Support

```yaml
# crew.yaml with language-aware routing
agents:
  - language_detector
  - english_processor
  - spanish_processor
  - chinese_processor

signals:
  language_detector:
    - signal: "[ENGLISH]"
      target: english_processor
    - signal: "[SPANISH]"
      target: spanish_processor
    - signal: "[CHINESE]"
      target: chinese_processor
```

## Testing Your Crew

```go
func TestYourCrew(t *testing.T) {
    // Load config
    crewConfig, _ := crewai.LoadCrewConfig("config/crew.yaml")
    agentConfigs, _ := crewai.LoadAgentConfigs("config/agents")

    // Create crew
    allTools := GetYourCustomTools()
    agents := createAgents(crewConfig, agentConfigs, allTools)
    crew := &crewai.Crew{
        Agents:      agents,
        MaxHandoffs: crewConfig.Settings.MaxHandoffs,
        MaxRounds:   crewConfig.Settings.MaxRounds,
        Routing:     crewConfig.Routing,
    }

    // Test execution
    executor := crewai.NewCrewExecutor(crew, "test-api-key")
    response, err := executor.Execute(context.Background(), "test input")

    // Assert expectations
    if err != nil {
        t.Fatalf("Unexpected error: %v", err)
    }
    if response.AgentID != "expected_agent" {
        t.Errorf("Expected agent_x, got %s", response.AgentID)
    }
}
```

## Troubleshooting

### Issue: Agent not routing correctly

**Solution**: Check that:
1. Signal is spelled correctly in agent response
2. Signal is defined in crew.yaml routing config
3. Target agent exists in crew.yaml agents list

### Issue: Tools not executing

**Solution**: Verify:
1. Tool is registered in agent config `tools:` section
2. Tool name matches exactly (case-sensitive)
3. Tool handler is implemented correctly

### Issue: Infinite loops or excessive handoffs

**Solution**:
1. Set appropriate `max_handoffs` limit
2. Ensure agents emit signals to reach terminal agent
3. Use `wait_for_signal: true` to prevent auto-routing

## Support

For issues, questions, or feature requests, please refer to the main project documentation or contact your team lead.

## License

Go-CrewAI is licensed under the same license as the parent project.

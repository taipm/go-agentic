# Go-CrewAI Architecture & Design Decisions

## Overview

Go-CrewAI implements a **Hybrid Multi-Agent Orchestration Framework** that combines:
- **Config-Driven Architecture**: Behavior defined in YAML, not hardcoded
- **LLM-Powered Decision Making**: Intelligent routing through custom prompts
- **Signal-Based Handoffs**: Explicit routing signals for deterministic workflows
- **Reusable Library Design**: Zero domain-specific hardcoding

## Design Philosophy

### 1. Configuration Over Code

**Principle**: Application behavior should be configurable, not hardcoded.

**Implementation**:
- All agent properties defined in YAML files
- All routing rules in crew.yaml
- All signals defined in routing config
- Framework code contains zero hardcoded agent IDs or signal strings

**Benefits**:
- Teams can modify behavior without recompiling
- Different use cases can share same code with different configs
- Non-technical stakeholders can adjust parameters

### 2. Explicit Over Implicit

**Principle**: Routing decisions should be explicit, not inferred.

**Before (Problem)**:
```go
// Bad: Hidden logic about when to route
if strings.Contains(response.Content, "[ROUTE_EXECUTOR]") {
    // Route to executor - but what if orchestrator ID changes?
    // What if we add more agents? This breaks!
}
```

**After (Solution)**:
```go
// Good: Signals defined in config, framework looks them up
signals, exists := ce.crew.Routing.Signals[currentAgent.ID]
for _, sig := range signals {
    if strings.Contains(responseContent, sig.Signal) {
        return ce.findAgentByID(sig.Target)  // Use config, not hardcoded
    }
}
```

**Benefits**:
- Easy to add/modify signals without code changes
- Framework doesn't need to know agent names
- Same code works with any agent workflow

### 3. Hybrid Approach: LLM + Structure

**Principle**: Use LLM intelligence within structured constraints.

**How it works**:
1. **LLM Makes Decisions**: Custom system prompts guide LLM behavior
2. **Explicit Signals**: LLM must emit explicit signals for routing
3. **Framework Routes**: Framework reads signals from config and routes

**Why not pure LLM routing?**
- Unpredictable behavior (could hallucinate routing targets)
- Hard to audit and test (non-deterministic)
- Expensive (every routing decision costs API call)

**Why not pure config routing?**
- Can't adapt to context variations
- Limited to explicit rules
- Can't leverage LLM intelligence

**Hybrid Benefits**:
- LLM understands nuance and context
- Framework ensures routing is valid (defined in config)
- Deterministic and auditable
- Cost-effective (LLM only for decisions, not routing)

## Architecture Components

### 1. CrewConfig (crew.yaml)

**Responsibility**: Define complete crew structure and routing rules

```yaml
entry_point: orchestrator           # Starting agent
agents: [orchestrator, clarifier, executor]  # All agents in crew
routing:                            # Routing configuration
  signals:                          # Available signals per agent
  defaults:                         # Fallback routing
  agent_behaviors:                  # Agent-specific behaviors
```

**Why separate from agents?**
- Routing is a crew-level concern, not agent-level
- Different agent sets can share routing logic
- Team deploying crew can modify routing without agent config

### 2. AgentConfig (agents/*.yaml)

**Responsibility**: Define individual agent behavior

```yaml
id: orchestrator
name: "Orchestrator"
model: gpt-4o-mini
temperature: 0.7
tools: [...]
system_prompt: |
  {{name}}: Your role and instructions...
```

**Separation of Concerns**:
- Agent config: "What is this agent's identity and capabilities?"
- Crew config: "How do agents interact with each other?"

### 3. RoutingConfig (part of crew.yaml)

**Responsibility**: Define how agents communicate

```yaml
routing:
  signals:
    orchestrator:
      - signal: "[ROUTE_EXECUTOR]"
        target: executor
  defaults:
    orchestrator: clarifier
  agent_behaviors:
    orchestrator:
      wait_for_signal: true
```

**Key Concept: Signals are First-Class**

Instead of hardcoding signal detection:
```go
// Bad: Framework knows about specific signals
if strings.Contains(response, "[ROUTE_EXECUTOR]") {
    route_to_executor()
}

// Good: Signals defined in config
for _, sig := range config.Signals[agentID] {
    if contains(response, sig.Signal) {
        route_to(sig.Target)
    }
}
```

### 4. CrewExecutor (crew.go)

**Responsibility**: Orchestrate agent execution with config-driven routing

**Key Methods**:

- `Execute()`: Main execution loop
  - Load agent
  - Execute agent
  - Check for tool calls
  - Detect routing signal
  - Handoff to next agent

- `findNextAgentBySignal()`: **Config-driven routing**
  ```go
  func (ce *CrewExecutor) findNextAgentBySignal(current *Agent, response string) *Agent {
      // Look up signals in config
      signals := ce.crew.Routing.Signals[current.ID]
      // Check which signal is in response
      for _, sig := range signals {
          if strings.Contains(response, sig.Signal) {
              // Get target from config
              return ce.findAgentByID(sig.Target)
          }
      }
      return nil
  }
  ```

- `getAgentBehavior()`: Get behavior config for agent
  ```go
  // Check if agent waits for signal before auto-routing
  behavior := ce.getAgentBehavior(current.ID)
  if behavior.WaitForSignal {
      return response // Wait for signal instead of auto-routing
  }
  ```

**No Hardcoding**:
- Never checks `if agentID == "executor"`
- Never looks for hardcoded signals like `[ROUTE_EXECUTOR]`
- All logic reads from config

## Data Flow

```
User Input
    ↓
LoadCrewConfig() → CrewConfig (routes, signals)
LoadAgentConfigs() → AgentConfig[] (agent properties)
GetYourTools() → Tool[] (tool implementations)
    ↓
CreateAgentsFromConfig() → Agent[]
    ↓
Crew{Agents, Routing}
    ↓
CrewExecutor.Execute()
    ├─ ExecuteAgent(currentAgent)
    │   └─ LLM processes custom system prompt
    │       └─ Returns response with tools + signal
    ├─ Execute tool calls (if any)
    │   └─ Add results to history
    ├─ findNextAgentBySignal()
    │   ├─ Look up signals in crew.Routing.Signals
    │   ├─ Check if signal in response
    │   └─ Get target from config
    ├─ getAgentBehavior()
    │   └─ Check wait_for_signal flag
    └─ Continue loop or return response
```

## Type System

### RoutingSignal
```go
type RoutingSignal struct {
    Signal      string  // e.g., "[ROUTE_EXECUTOR]"
    Target      string  // e.g., "executor"
    Description string  // e.g., "Route to executor for diagnosis"
}
```

### AgentBehavior
```go
type AgentBehavior struct {
    WaitForSignal bool   // If true, must have explicit signal
    AutoRoute     bool   // If true, auto-route on no signal
    IsTerminal    bool   // If true, no handoff possible
}
```

### RoutingConfig
```go
type RoutingConfig struct {
    Signals        map[string][]RoutingSignal  // Per-agent signals
    Defaults       map[string]string           // Fallback targets
    AgentBehaviors map[string]AgentBehavior    // Per-agent behaviors
}
```

## Execution Modes

### Mode 1: Explicit Signal Routing (Recommended)

**Configuration**:
```yaml
agent_behaviors:
  orchestrator:
    wait_for_signal: true
    auto_route: false
```

**Behavior**:
- Agent MUST include routing signal in response
- If no signal → return response (wait for next input)
- Framework enforces valid routing (config-defined targets only)

**When to use**:
- Critical decision points
- Policy compliance scenarios
- Multi-step approval workflows

### Mode 2: Implicit Default Routing

**Configuration**:
```yaml
agent_behaviors:
  worker:
    auto_route: true
defaults:
  worker: next_worker
```

**Behavior**:
- If agent doesn't include signal → use default routing
- No explicit signal needed
- Framework uses config default

**When to use**:
- Linear processing pipelines
- Simple pass-through agents
- Fast routing without decision overhead

### Mode 3: Terminal Agent

**Configuration**:
```yaml
agent_behaviors:
  executor:
    is_terminal: true
```

**Behavior**:
- No routing possible
- Response returned immediately
- Workflow ends

**When to use**:
- Final output generation
- Result formatting
- Workflow conclusion

## Signal Semantics

### Why Signals Matter

Signals solve the "implicit routing problem":

**Problem**: How do we know when orchestrator wants to route to executor vs clarifier?

**Option A (Keyword Matching)**: Parse agent response for keywords
- ❌ Brittle (wording changes break routing)
- ❌ Ambiguous (multiple keywords possible)
- ❌ Hallucination-prone (LLM might include keywords accidentally)

**Option B (Explicit Signals)**: Agent must include specific signal
- ✅ Unambiguous (exact match required)
- ✅ Auditable (clear intent)
- ✅ Config-driven (easy to modify)
- ✅ Testable (verify signal presence)

### Signal Format

Signals use unambiguous format: `[ACTION_TARGET]`

Examples:
```
[ROUTE_EXECUTOR]        # Clear: route to executor
[NEEDS_APPROVAL]        # Clear: needs approval
[COMPLETE]              # Clear: workflow complete
[ESCALATE_TO_MANAGER]   # Clear: escalate to manager
```

## Configuration vs. Hardcoding

### Before Refactoring (Hardcoded)

```go
// crew.go - Hardcoded agent IDs
if currentAgent.ID == "orchestrator" {
    if strings.Contains(response.Content, "[ROUTE_EXECUTOR]") {
        nextAgent := ce.findAgentByID("executor")  // Hardcoded
    }
}

// crew.go - Hardcoded signal patterns
if currentAgent.ID == "clarifier" && strings.Contains(response.Content, "[KẾT THÚC]") {
    // Hardcoded signal string
}
```

**Problems**:
- Adding new agents requires code changes
- Changing signals requires code changes
- Can't reuse code for different workflows
- Testing requires mocking framework internals

### After Refactoring (Config-Driven)

```go
// crew.go - Generic signal lookup
nextAgent := ce.findNextAgentBySignal(currentAgent, response.Content)

// findNextAgentBySignal method
func (ce *CrewExecutor) findNextAgentBySignal(current *Agent, response string) *Agent {
    signals := ce.crew.Routing.Signals[current.ID]  // From config
    for _, sig := range signals {
        if strings.Contains(response, sig.Signal) {
            return ce.findAgentByID(sig.Target)
        }
    }
    return nil
}

// crew.yaml - All signals defined
signals:
  orchestrator:
    - signal: "[ROUTE_EXECUTOR]"
      target: executor
```

**Benefits**:
- Framework doesn't know agent names
- Teams can add agents without code changes
- Same code works for any workflow
- Configuration change = deployment (not recompilation)

## Extensibility Patterns

### Pattern 1: Adding New Agents

1. Create `agents/new_agent.yaml`
2. Add to `agents:` list in `crew.yaml`
3. Define signals and routing in `crew.yaml`
4. No code changes needed

### Pattern 2: Changing Routing

1. Modify `routing:` section in `crew.yaml`
2. Change signals, defaults, or behaviors
3. No code changes needed

### Pattern 3: Custom Tools

1. Implement tool with `Handler` function
2. Register in agent config `tools:` list
3. LLM can call via standard interface
4. Framework executes independently

### Pattern 4: Domain-Specific Extensions

1. Fork framework code
2. Add domain-specific configuration
3. Keep framework-level code unchanged
4. Update tool implementations for domain

## Testing & Validation

### Test Coverage Approach

1. **Unit Tests**: Individual agent behavior
2. **Integration Tests**: Full crew workflow
3. **Scenario Tests**: Real-world use cases
4. **Routing Tests**: Signal detection and routing

### Example Test Structure

```go
scenario := &TestScenario{
    Name: "Test orchestrator routing",
    Input: "specific server problem",
    ExpectedFlow: []string{"orchestrator", "executor"},
    Assertions: []string{
        "Orchestrator emits [ROUTE_EXECUTOR]",
        "Framework routes to executor",
        "Executor receives correct input",
    },
}
```

## Performance Considerations

### Signal Matching
- O(n) where n = number of signals for agent
- Typical: 2-5 signals per agent
- Negligible overhead

### Agent Lookup
- O(m) where m = number of agents in crew
- Typical: 3-10 agents
- Could optimize with map-based lookup if needed

### LLM Calls
- Primary cost: LLM API calls
- Framework adds minimal latency
- No polling or retry logic needed (explicit signals)

## Future Enhancements

### Planned Features

1. **Signal Metadata**: Include structured data in signals
   ```
   [ROUTE_EXECUTOR:priority=high,timeout=30s]
   ```

2. **Dynamic Routing**: Conditional routing based on response content
   ```yaml
   routing:
     conditional:
       - condition: "contains(response, 'urgent')"
         target: escalator
   ```

3. **Agent Pools**: Load balancing across multiple instances
   ```yaml
   agents:
     - name: executor
       count: 3  # Run 3 executors in parallel
   ```

4. **Observability**: Built-in tracing and metrics
   ```go
   executor.OnSignal(func(signal, source, target string) {
       // Log or metrics tracking
   })
   ```

## Backward Compatibility

### Current Version (v1.0)

- Pure YAML-based configuration
- Signal-based routing
- Single-instance agents
- Synchronous execution

### Migration Path

If upgrading from previous hardcoded version:

1. Export agent configurations from code
2. Create crew.yaml with routing rules
3. Create agents/*.yaml files
4. Update initialization code to load configs
5. Test routing with test scenarios

## Conclusion

Go-CrewAI's architecture prioritizes:

1. **Flexibility**: Works with any agent workflow
2. **Maintainability**: Config changes don't require code changes
3. **Auditability**: Explicit routing is traceable
4. **Reusability**: Same framework for different domains
5. **Simplicity**: Clear separation of concerns

The hybrid approach (LLM + config) provides the best of both worlds: LLM intelligence within framework-enforced constraints.

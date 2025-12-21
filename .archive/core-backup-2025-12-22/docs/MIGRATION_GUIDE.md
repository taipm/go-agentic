# Migration Guide: From Hardcoded to Config-Driven Framework

This guide helps teams migrate from the old hardcoded agent framework to the new config-driven go-crewai library.

## Before & After Comparison

### Old Way (Hardcoded)

```go
// Creating agents - hardcoded properties
orchestrator := &Agent{
    ID:          "orchestrator",
    Name:        "Orchestrator",
    Role:        "System coordinator",
    Backstory:   "You are the entry point...",
    Model:       "gpt-4o",
    Tools:       []*Tool{},
    Temperature: 0.7,
    IsTerminal:  false,
}

// Routing - hardcoded agent IDs
if currentAgent.ID == "orchestrator" {
    if strings.Contains(response.Content, "[ROUTE_EXECUTOR]") {
        nextAgent := ce.findAgentByID("executor")  // Hardcoded string
    } else if shouldRouteToExecutor(response.Content) {
        nextAgent := ce.findAgentByID("executor")
    } else {
        nextAgent := ce.findAgentByID("clarifier")  // Hardcoded fallback
    }
}

// Adding new signals required code changes
if currentAgent.ID == "clarifier" && strings.Contains(response.Content, "[KẾT THÚC]") {
    // Change signal? Need to change code
}
```

**Problems**:
- ❌ Agent properties scattered across code
- ❌ Routing logic hardcoded
- ❌ Signal detection hardcoded
- ❌ Adding agents requires code changes
- ❌ Changing signals requires code changes
- ❌ Not reusable for other domains

### New Way (Config-Driven)

```go
// Loading agents from YAML config
crewConfig, _ := crewai.LoadCrewConfig("config/crew.yaml")
agentConfigs, _ := crewai.LoadAgentConfigs("config/agents")
agents := createAgentsFromConfig(crewConfig, agentConfigs, allTools)

// Routing - config-driven
nextAgent := ce.findNextAgentBySignal(currentAgent, response.Content)

// Adding new signals: just update crew.yaml - no code changes!
```

**Benefits**:
- ✅ Agent properties in YAML
- ✅ Routing from configuration
- ✅ Signal detection from configuration
- ✅ Add agents without code changes
- ✅ Change signals without code changes
- ✅ Reusable framework

## Migration Steps

### Step 1: Create Configuration Files

#### 1.1 Create crew.yaml

```bash
mkdir -p config/agents
touch config/crew.yaml
```

**crew.yaml** (minimal example):
```yaml
version: "1.0"
description: "Your Crew Description"
entry_point: orchestrator

agents:
  - orchestrator
  - clarifier
  - executor

settings:
  max_handoffs: 5
  max_rounds: 10
  timeout_seconds: 300
  language: en
  organization: YourOrganization

routing:
  signals:
    orchestrator:
      - signal: "[ROUTE_EXECUTOR]"
        target: executor
      - signal: "[ROUTE_CLARIFIER]"
        target: clarifier
    clarifier:
      - signal: "[KẾT THÚC]"
        target: executor
    executor:
      - signal: "[COMPLETE]"
        target: null

  defaults:
    orchestrator: clarifier
    clarifier: executor
    executor: null

  agent_behaviors:
    orchestrator:
      wait_for_signal: true
      auto_route: false
    clarifier:
      wait_for_signal: true
      auto_route: false
    executor:
      is_terminal: true
```

#### 1.2 Create Agent Configuration Files

For each agent, create `config/agents/agent_id.yaml`:

**config/agents/orchestrator.yaml**:
```yaml
id: orchestrator
name: "Orchestrator"
description: "Entry point for requests"
role: "Request Router"

backstory: |
  You are the entry point for all requests.
  Your job is to analyze and route appropriately.

model: gpt-4o-mini
temperature: 0.7
is_terminal: false

tools: []
handoff_targets:
  - clarifier
  - executor

system_prompt: |
  You are {{name}}.
  Vai trò: {{role}}

  Analyze the request and emit exactly one signal:
  - [ROUTE_EXECUTOR] if ready for execution
  - [ROUTE_CLARIFIER] if need more information
```

**config/agents/clarifier.yaml**:
```yaml
id: clarifier
name: "Clarifier"
description: "Information gatherer"
role: "Request Clarifier"

backstory: |
  You gather detailed information about requests.
  Ask clarifying questions to understand better.

model: gpt-4o-mini
temperature: 0.7
is_terminal: false

tools: []
handoff_targets:
  - executor

system_prompt: |
  You are {{name}}.

  Ask 2-3 clarifying questions to understand the request.
  When you have enough information, emit: [KẾT THÚC]
```

**config/agents/executor.yaml**:
```yaml
id: executor
name: "Executor"
description: "Action executor"
role: "Executor"

backstory: |
  You execute actions based on clear requests.
  Use tools to accomplish the task.

model: gpt-4o-mini
temperature: 0.7
is_terminal: true

tools:
  - YourTool1
  - YourTool2

system_prompt: |
  You are {{name}}.

  Execute the requested action using available tools.
  Provide clear results and recommendations.
```

### Step 2: Update Initialization Code

#### Before:
```go
func main() {
    orchestrator := &Agent{ID: "orchestrator", Name: "Orchestrator", ...}
    clarifier := &Agent{ID: "clarifier", Name: "Clarifier", ...}
    executor := &Agent{ID: "executor", Name: "Executor", ...}

    crew := &Crew{
        Agents: []*Agent{orchestrator, clarifier, executor},
        MaxRounds: 10,
        MaxHandoffs: 5,
    }

    executor := NewCrewExecutor(crew, apiKey)
}
```

#### After:
```go
func main() {
    // Load configuration
    crewConfig, _ := crewai.LoadCrewConfig("config/crew.yaml")
    agentConfigs, _ := crewai.LoadAgentConfigs("config/agents")

    // Get your custom tools
    allTools := GetYourCustomTools()

    // Create agents from config
    agents := createAgentsFromConfig(crewConfig, agentConfigs, allTools)

    // Create crew with routing configuration
    crew := &crewai.Crew{
        Agents:      agents,
        MaxRounds:   crewConfig.Settings.MaxRounds,
        MaxHandoffs: crewConfig.Settings.MaxHandoffs,
        Routing:     crewConfig.Routing,  // Important: include routing config
    }

    executor := crewai.NewCrewExecutor(crew, apiKey)
}
```

#### Helper Function:
```go
func createAgentsFromConfig(crewConfig *crewai.CrewConfig,
    agentConfigs map[string]*crewai.AgentConfig,
    allTools map[string]*crewai.Tool) []*crewai.Agent {

    var agents []*crewai.Agent
    for _, agentID := range crewConfig.Agents {
        if config, exists := agentConfigs[agentID]; exists {
            agent := crewai.CreateAgentFromConfig(config, allTools)
            agents = append(agents, agent)
        }
    }
    return agents
}
```

### Step 3: Update System Prompts

Move system prompt logic from code to YAML configuration.

#### Before (Hardcoded):
```go
orchestrator := &Agent{
    SystemPrompt: fmt.Sprintf(`
        You are %s.
        Role: %s

        Route requests to:
        - [ROUTE_EXECUTOR] if ready
        - [ROUTE_CLARIFIER] if need info
    `, "Orchestrator", "Router"),
}
```

#### After (YAML):
```yaml
# config/agents/orchestrator.yaml
system_prompt: |
  You are {{name}}.
  Vai trò: {{role}}

  Route requests to:
  - [ROUTE_EXECUTOR] if ready
  - [ROUTE_CLARIFIER] if need info
```

The framework automatically replaces:
- `{{name}}` with agent.Name
- `{{role}}` with agent.Role
- `{{description}}` with agent.Description
- `{{backstory}}` with agent.Backstory

### Step 4: Remove Hardcoded Routing Logic

#### Code to Remove:

```go
// DELETE: Hardcoded agent ID checks
if currentAgent.ID == "orchestrator" {
    // ... DELETE this entire block
}

if currentAgent.ID == "clarifier" && strings.Contains(...) {
    // ... DELETE this entire block
}

// DELETE: Hardcoded signal detection
func shouldRouteToExecutor(response string) bool {
    // ... DELETE this entire function
}

// DELETE: Manual routing decisions
if condition {
    nextAgent := ce.findAgentByID("executor")  // DELETE
}
```

#### Code to Keep:

```go
// KEEP: Generic signal-based routing (uses config)
nextAgent := ce.findNextAgentBySignal(currentAgent, response.Content)

// KEEP: Behavior-based routing (reads from config)
behavior := ce.getAgentBehavior(currentAgent.ID)
if behavior != nil && behavior.WaitForSignal {
    return response
}
```

### Step 5: Update Signals in Configuration

#### Before (Hardcoded in code):
```go
// Hardcoded signal detection
if strings.Contains(response, "[ROUTE_EXECUTOR]") {
    // Route to executor
} else if strings.Contains(response, "[KẾT THÚC]") {
    // Route to next
}
```

#### After (Configured):
```yaml
# config/crew.yaml
routing:
  signals:
    orchestrator:
      - signal: "[ROUTE_EXECUTOR]"
        target: executor
      - signal: "[ROUTE_CLARIFIER]"
        target: clarifier
    clarifier:
      - signal: "[KẾT THÚC]"
        target: executor
```

To add a new signal:
1. Update agent system_prompt to emit signal
2. Add signal definition to crew.yaml
3. **No code changes needed!**

### Step 6: Update Tests

#### Before (Hardcoded assertions):
```go
func TestRouting(t *testing.T) {
    // Create hardcoded agents
    orchestrator := &Agent{ID: "orchestrator", ...}

    // Execute and check hardcoded routing
    response := executor.Execute(ctx, input)
    if response.AgentID != "executor" {
        t.Fail()
    }
}
```

#### After (Config-driven):
```go
func TestRouting(t *testing.T) {
    // Load from config
    crewConfig, _ := crewai.LoadCrewConfig("config/crew.yaml")
    agentConfigs, _ := crewai.LoadAgentConfigs("config/agents")

    // Create crew from config
    agents := createAgentsFromConfig(crewConfig, agentConfigs, tools)
    crew := &crewai.Crew{
        Agents:  agents,
        Routing: crewConfig.Routing,
    }

    // Execute and check routing
    executor := crewai.NewCrewExecutor(crew, apiKey)
    response, _ := executor.Execute(ctx, input)

    if response.AgentID != "executor" {
        t.Errorf("Expected executor, got %s", response.AgentID)
    }
}
```

## Common Migration Patterns

### Pattern 1: Reusing Agents Across Crews

**Before**: Couldn't easily reuse agent definitions

**After**:
```yaml
# crew1/agents/analyzer.yaml
id: analyzer
# ... config

# crew2/agents/analyzer.yaml  (same config)
id: analyzer
# ... config

# Different crews, different routing
# Same agent definition reused
```

### Pattern 2: Adding New Agents

**Before**: Required code changes + recompile

**After**:
1. Create `config/agents/new_agent.yaml`
2. Add to `agents:` list in `crew.yaml`
3. Define routing in `crew.yaml`
4. Done! No code changes

### Pattern 3: Changing Signals

**Before**: Required code changes + recompile

**After**:
1. Update signal string in agent system_prompt
2. Update signal definition in `crew.yaml`
3. Deploy new config
4. Done! No code recompile

### Pattern 4: A/B Testing Behavior

**Before**: Couldn't easily test different behaviors

**After**:
```bash
# Test crew with behavior A
go run cmd/main.go --config=config/crew_variant_a.yaml

# Test crew with behavior B
go run cmd/main.go --config=config/crew_variant_b.yaml
```

## Validation Checklist

After migration, verify:

- [ ] All agents have YAML config files
- [ ] crew.yaml includes all agents
- [ ] All signals defined in crew.yaml routing section
- [ ] All signal targets exist in agents list
- [ ] System prompts use template variables ({{name}}, etc.)
- [ ] Tools registered in agent config match available tools
- [ ] Handoff targets are valid agent IDs
- [ ] Default routing covers all agents
- [ ] Agent behaviors correctly set (wait_for_signal, etc.)
- [ ] Tests pass with new config-driven setup
- [ ] No hardcoded agent IDs in framework code
- [ ] No hardcoded signal strings in framework code

## Troubleshooting

### Issue: Agent not found after migration

**Cause**: Agent ID in config doesn't match agent definition

**Solution**:
1. Check `crew.yaml` agents list
2. Check agent YAML filename matches agent ID
3. Check agent config `id:` field matches filename

### Issue: Routing not working after migration

**Cause**: Signal not defined in routing config

**Solution**:
1. Verify agent emits signal
2. Check signal spelling exactly matches crew.yaml
3. Check target agent exists
4. Enable debug logging to see routing decisions

### Issue: Tools not executing

**Cause**: Tool not registered in agent config

**Solution**:
1. Check agent YAML includes tool in `tools:` section
2. Check tool name matches exactly (case-sensitive)
3. Verify tool is registered in allTools map

## Performance Considerations

### Migration Impact

- **Startup Time**: +10-50ms (loading config files) - negligible
- **Runtime Overhead**: None (config loaded once at startup)
- **Signal Matching**: Same speed as before (string search)

### Optimization Tips

1. Load config once, reuse for multiple executions
2. Use config file caching for rapid iterations
3. Keep signal names short for efficiency (negligible)

## Rollback Plan

If issues arise:

1. Keep old hardcoded version in separate branch
2. Run both versions in parallel for comparison
3. Gradually migrate crew members to new system
4. Monitor logs for routing issues
5. Can rollback by reverting code + config

## Support & Resources

- **LIBRARY_USAGE.md**: Detailed usage guide
- **ARCHITECTURE.md**: Design philosophy
- **Example configurations**: See config/ directory
- **Test scenarios**: See tests.go for working examples

## Migration Timeline

Recommended timeline for large teams:

1. **Week 1**: Create config files, test locally
2. **Week 2**: Update initialization code, run tests
3. **Week 3**: Deploy to staging, validate
4. **Week 4**: Deploy to production
5. **Ongoing**: Monitor and optimize

## Questions?

Refer to LIBRARY_USAGE.md for detailed examples, or check the working IT Support implementation for reference.

# UX Implementation Guide: go-agentic

**Mục tiêu**: Hướng dẫn triển khai cải thiện UX cho go-agentic
**Phạm vi**: Phase 1 (Fluent API) + Phase 2 (Unified Config)
**Thời gian**: 2-3 tuần (32 giờ)

---

## 📋 Mục Lục

1. [Phase 1: Fluent Builder API](#phase-1-fluent-builder-api-6h)
2. [Phase 2: Unified YAML Configuration](#phase-2-unified-yaml-configuration-8h)
3. [Phase 3: Declarative Routing](#phase-3-declarative-routing-6h)
4. [Examples & Migration](#examples--migration-12h)

---

## Phase 1: Fluent Builder API (6h)

### 1.1 Goal

Giảm boilerplate khi create agents, từ:
```go
agent := &agentic.Agent{ID: "...", Name: "...", ...}  // 8 fields
```

Thành:
```go
agent := agentic.NewAgent("id", "Name").
    WithRole("role").
    WithBackstory("...").
    Build()
```

### 1.2 Implementation Details

#### File: `go-agentic/builder.go` (New)

```go
package agentic

// AgentBuilder for fluent agent configuration
type AgentBuilder struct {
	agent *Agent
}

// NewAgent starts building an agent
func NewAgent(id, name string) *AgentBuilder {
	return &AgentBuilder{
		agent: &Agent{
			ID:   id,
			Name: name,
			// Set sensible defaults
			Model:       "gpt-4o-mini",
			Temperature: 0.7,
			Tools:       []*Tool{},
		},
	}
}

// WithRole sets agent role
func (ab *AgentBuilder) WithRole(role string) *AgentBuilder {
	ab.agent.Role = role
	return ab
}

// WithBackstory sets agent backstory
func (ab *AgentBuilder) WithBackstory(backstory string) *AgentBuilder {
	ab.agent.Backstory = backstory
	return ab
}

// WithModel sets the model (NOT hardcoded!)
func (ab *AgentBuilder) WithModel(model string) *AgentBuilder {
	ab.agent.Model = model
	return ab
}

// WithTemperature sets temperature
func (ab *AgentBuilder) WithTemperature(temp float64) *AgentBuilder {
	ab.agent.Temperature = temp
	return ab
}

// AddTool adds a tool to agent
func (ab *AgentBuilder) AddTool(tool *Tool) *AgentBuilder {
	ab.agent.Tools = append(ab.agent.Tools, tool)
	return ab
}

// AddTools adds multiple tools
func (ab *AgentBuilder) AddTools(tools ...*Tool) *AgentBuilder {
	ab.agent.Tools = append(ab.agent.Tools, tools...)
	return ab
}

// SetTerminal marks agent as terminal
func (ab *AgentBuilder) SetTerminal(terminal bool) *AgentBuilder {
	ab.agent.IsTerminal = terminal
	return ab
}

// WithHandoff adds handoff target
func (ab *AgentBuilder) WithHandoff(target string) *AgentBuilder {
	ab.agent.HandoffTargets = append(ab.agent.HandoffTargets, target)
	return ab
}

// WithHandoffs sets all handoff targets
func (ab *AgentBuilder) WithHandoffs(targets ...string) *AgentBuilder {
	ab.agent.HandoffTargets = targets
	return ab
}

// Build returns configured agent
func (ab *AgentBuilder) Build() *Agent {
	// Validate
	if ab.agent.ID == "" {
		panic("Agent ID required")
	}
	if ab.agent.Name == "" {
		panic("Agent Name required")
	}
	if ab.agent.Role == "" {
		panic("Agent Role required")
	}
	return ab.agent
}

// ============================================
// ToolBuilder for simpler tool definition
// ============================================

type ToolBuilder struct {
	tool *Tool
}

// NewTool starts building a tool
func NewTool(name, description string) *ToolBuilder {
	return &ToolBuilder{
		tool: &Tool{
			Name:        name,
			Description: description,
			Parameters: map[string]interface{}{
				"type":       "object",
				"properties": map[string]interface{}{},
			},
		},
	}
}

// NoParameters sets empty parameter schema
func (tb *ToolBuilder) NoParameters() *ToolBuilder {
	return tb
}

// WithParameter adds a parameter to schema
func (tb *ToolBuilder) WithParameter(name, paramType, description string) *ToolBuilder {
	props := tb.tool.Parameters["properties"].(map[string]interface{})
	props[name] = map[string]interface{}{
		"type":        paramType,
		"description": description,
	}
	return tb
}

// WithParameter(name, "string", "description")
// WithParameter(name, "number", "description")
// etc.

// Handler sets the handler function
func (tb *ToolBuilder) Handler(handler ToolHandler) *ToolBuilder {
	tb.tool.Handler = handler
	return tb
}

// Build returns configured tool
func (tb *ToolBuilder) Build() *Tool {
	if tb.tool.Name == "" {
		panic("Tool name required")
	}
	if tb.tool.Handler == nil {
		panic("Tool handler required")
	}
	return tb.tool
}
```

### 1.3 Usage Example

**Before** (Old way, still works):
```go
agent := &agentic.Agent{
    ID: "executor",
    Name: "Executor",
    Role: "Execute tasks",
    Backstory: "You are an executor...",
    Model: "gpt-4o-mini",
    Temperature: 0.7,
    IsTerminal: true,
    Tools: []*agentic.Tool{...},
}
```

**After** (New fluent way):
```go
agent := agentic.NewAgent("executor", "Executor").
    WithRole("Execute tasks").
    WithBackstory("You are an executor...").
    WithModel("gpt-4o-mini").
    WithTemperature(0.7).
    SetTerminal(true).
    AddTools(cpuTool, memoryTool, diskTool).
    Build()
```

### 1.4 Tests to Add

`go-agentic/builder_test.go`:
```go
func TestAgentBuilderCreatesAgent(t *testing.T) {
    agent := NewAgent("id", "Name").
        WithRole("role").
        WithModel("gpt-4o").
        Build()

    if agent.ID != "id" || agent.Model != "gpt-4o" {
        t.Fatal("Builder failed")
    }
}

func TestToolBuilderSimplifies(t *testing.T) {
    tool := NewTool("test", "desc").
        NoParameters().
        Handler(func(ctx context.Context, args map[string]interface{}) (string, error) {
            return "result", nil
        }).
        Build()

    if tool.Name != "test" {
        t.Fatal("Tool builder failed")
    }
}
```

### 1.5 Effort Breakdown

| Task | Time | Output |
|------|------|--------|
| AgentBuilder implementation | 1.5h | builder.go (~100 lines) |
| ToolBuilder implementation | 1.5h | builder.go (+80 lines) |
| Tests | 1h | builder_test.go (~150 lines) |
| Documentation | 1h | Examples in comments |
| Simple Chat v2 example | 1h | Updated example |

**Total**: 6 hours

---

## Phase 2: Unified YAML Configuration (8h)

### 2.1 Goal

Replace 4-5 separate config files with **single team.yaml**:
```yaml
# OLD: team.yaml + agents/*.yaml (5 files)
# NEW: single team.yaml (1 file)
```

### 2.2 New YAML Schema

`examples/IT_Support_Unified/team.yaml`:
```yaml
# Unified Team Configuration
team:
  name: "IT Support System"
  config:
    maxRounds: 10
    maxHandoffs: 5

agents:
  orchestrator:
    id: orchestrator
    name: "System Orchestrator"
    role: "Initial request analyzer"
    backstory: |
      You are the entry point for IT support requests.
      Analyze user issues and route them appropriately.
    model: gpt-4o-mini
    temperature: 0.7
    isTerminal: false
    tools: []  # No tools for router

  clarifier:
    id: clarifier
    name: "Question Asker"
    role: "Get missing information"
    backstory: |
      You ask clarifying questions to understand issues better.
    model: gpt-4o-mini
    temperature: 0.6
    isTerminal: false
    tools: []

  executor:
    id: executor
    name: "System Executor"
    role: "Execute diagnostics and solutions"
    backstory: |
      You run diagnostics and implement solutions.
    model: gpt-4o-mini
    temperature: 0.7
    isTerminal: true  # Final agent
    tools:
      - get_cpu_usage
      - get_memory_info
      - check_disk_space
      - list_processes
      - check_service_status

# Tool definitions (replaces tools.go)
tools:
  get_cpu_usage:
    name: GetCPUUsage
    description: "Get current CPU usage percentage"
    # Handler defined in code

  get_memory_info:
    name: GetMemoryInfo
    description: "Get current memory usage info"

  check_disk_space:
    name: CheckDiskSpace
    description: "Check disk space on system"

  list_processes:
    name: ListProcesses
    description: "List running processes"

  check_service_status:
    name: CheckServiceStatus
    description: "Check if a service is running"
    parameters:
      type: object
      properties:
        service:
          type: string
          description: "Service name to check"

# Routing configuration (replaces hardcoded system prompts!)
routing:
  type: "signal"
  rules:
    # Orchestrator routes based on clarity
    - from_agent: orchestrator
      trigger: "needs_more_info"
      target_agent: clarifier
      description: "User provided unclear information"

    - from_agent: orchestrator
      trigger: "ready_to_diagnose"
      target_agent: executor
      description: "Enough information to start diagnosis"

    # Clarifier always hands off to executor
    - from_agent: clarifier
      trigger: "info_complete"
      target_agent: executor
      description: "Got all needed information"

    # Executor is terminal (no routing)
    - from_agent: executor
      trigger: always
      target_agent: null
      description: "Executor is final step"
```

### 2.3 Implementation

#### File: `go-agentic/config.go` (Enhanced)

```go
package agentic

import (
	"os"
	"gopkg.in/yaml.v3"
)

// UnifiedTeamConfig represents single team.yaml
type UnifiedTeamConfig struct {
	Team UnifiedTeamMetadata `yaml:"team"`
	Agents map[string]*AgentConfigUnified `yaml:"agents"`
	Tools map[string]*ToolConfigUnified `yaml:"tools"`
	Routing *RoutingConfig `yaml:"routing"`
}

type UnifiedTeamMetadata struct {
	Name string `yaml:"name"`
	Config struct {
		MaxRounds int `yaml:"maxRounds"`
		MaxHandoffs int `yaml:"maxHandoffs"`
	} `yaml:"config"`
}

type AgentConfigUnified struct {
	ID string `yaml:"id"`
	Name string `yaml:"name"`
	Role string `yaml:"role"`
	Backstory string `yaml:"backstory"`
	Model string `yaml:"model"`
	Temperature float64 `yaml:"temperature"`
	IsTerminal bool `yaml:"isTerminal"`
	Tools []string `yaml:"tools"`  // List of tool IDs
}

type ToolConfigUnified struct {
	Name string `yaml:"name"`
	Description string `yaml:"description"`
	Parameters map[string]interface{} `yaml:"parameters"`
}

type RoutingConfig struct {
	Type string `yaml:"type"`  // "signal" or "custom"
	Rules []RoutingRule `yaml:"rules"`
}

type RoutingRule struct {
	FromAgent string `yaml:"from_agent"`
	Trigger string `yaml:"trigger"`
	TargetAgent *string `yaml:"target_agent"`  // nil = terminal
	Description string `yaml:"description"`
}

// LoadTeamFromYAML loads unified team configuration
func LoadTeamFromYAML(path string, toolHandlers map[string]ToolHandler) (*Team, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var config UnifiedTeamConfig
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, err
	}

	// Create agents from config
	agents := make([]*Agent, 0, len(config.Agents))
	for _, agentCfg := range config.Agents {
		agent := NewAgent(agentCfg.ID, agentCfg.Name).
			WithRole(agentCfg.Role).
			WithBackstory(agentCfg.Backstory).
			WithModel(agentCfg.Model).
			WithTemperature(agentCfg.Temperature).
			SetTerminal(agentCfg.IsTerminal)

		// Add tools from config
		for _, toolID := range agentCfg.Tools {
			if handler, ok := toolHandlers[toolID]; ok {
				toolCfg := config.Tools[toolID]
				tool := NewTool(toolCfg.Name, toolCfg.Description).
					Handler(handler).
					Build()
				agent.AddTool(tool)
			}
		}

		agents = append(agents, agent.Build())
	}

	// Create team
	team := &Team{
		Agents: agents,
		MaxRounds: config.Team.Config.MaxRounds,
		MaxHandoffs: config.Team.Config.MaxHandoffs,
	}

	// Set routing if provided
	if config.Routing != nil {
		team.RoutingConfig = config.Routing
	}

	return team, nil
}
```

### 2.4 Usage Example

**Before** (5 files):
```
team.yaml
agents/orchestrator.yaml
agents/clarifier.yaml
agents/executor.yaml
```

**After** (1 file):
```yaml
# team.yaml (single file, complete config)
team:
  name: "IT Support"
  config:
    maxRounds: 10
agents:
  orchestrator: {...}
  clarifier: {...}
  executor: {...}
tools:
  get_cpu_usage: {...}
routing:
  rules: [...]
```

**Code**:
```go
// Define tool handlers in code
toolHandlers := map[string]agentic.ToolHandler{
    "get_cpu_usage": getCPUUsageHandler,
    "get_memory_info": getMemoryHandler,
    // ...
}

// Load everything from single file
team, err := agentic.LoadTeamFromYAML("team.yaml", toolHandlers)
executor := agentic.NewTeamExecutor(team, apiKey)
```

### 2.5 Routing Configuration Details

**Signal-based Routing** (new):
```yaml
routing:
  type: "signal"
  rules:
    - from_agent: orchestrator
      trigger: "needs_info"  # LLM generates this signal
      target_agent: clarifier
```

**How it works**:
1. Orchestrator system prompt includes: "If you need more info, output [TRIGGER_NEEDS_INFO]"
2. After response, we parse for trigger signal
3. Match against routing rules
4. Route to next agent

**Old way** (text matching):
```
If response contains "[ROUTE_CLARIFIER]" → route to clarifier
```

**New way** (signal-based):
```yaml
trigger: "needs_info"  # Intent, not text
```

### 2.6 Tests to Add

`go-agentic/config_unified_test.go`:
```go
func TestLoadTeamFromYAML(t *testing.T) {
    handlers := map[string]agentic.ToolHandler{
        "test_tool": func(ctx context.Context, args map[string]interface{}) (string, error) {
            return "result", nil
        },
    }

    team, err := agentic.LoadTeamFromYAML("test_team.yaml", handlers)
    if err != nil || len(team.Agents) == 0 {
        t.Fatal("LoadTeamFromYAML failed")
    }
}
```

### 2.7 Effort Breakdown

| Task | Time | Output |
|------|------|--------|
| Schema design (UnifiedTeamConfig types) | 1h | types.go updates |
| LoadTeamFromYAML function | 2h | config.go (~80 lines) |
| Tests | 1h | config_test.go (~100 lines) |
| Example IT Support unified config | 1.5h | team.yaml (~80 lines) |
| Documentation | 1.5h | README updates |
| Verify backward compatibility | 1h | Testing both formats |

**Total**: 8 hours

---

## Phase 3: Declarative Routing (6h)

### 3.1 Goal

Remove hardcoded routing from system prompts. Instead use declarative routing rules.

### 3.2 Problem with Current Approach

Current system prompt (orchestrator.yaml, 145 lines):
```
Jika [ROUTE_CLARIFIER] thì output "[ROUTE_CLARIFIER]"
Jika [ROUTE_EXECUTOR] thì output "[ROUTE_EXECUTOR]"
[200+ lines of routing logic in text...]
```

**Issues**:
- ❌ Fragile text matching
- ❌ Not reusable
- ❌ Difficult to debug
- ❌ Hardcoded signal names

### 3.3 Solution: Router & Trigger System

#### File: `go-agentic/routing.go` (New)

```go
package agentic

// Router manages agent routing logic
type Router struct {
	rules map[string][]*RoutingRule  // from_agent → rules
}

// NewRouter creates new router
func NewRouter() *Router {
	return &Router{
		rules: make(map[string][]*RoutingRule),
	}
}

// From begins routing from agent
func (r *Router) From(agentID string) *RouterFrom {
	return &RouterFrom{
		router: r,
		agentID: agentID,
	}
}

// RouterFrom continues building routing
type RouterFrom struct {
	router *Router
	agentID string
	agent *Agent  // Set later
}

// OnTrigger adds routing rule
func (rf *RouterFrom) OnTrigger(trigger string) *RouterTrigger {
	return &RouterTrigger{
		router: rf.router,
		fromAgent: rf.agentID,
		trigger: trigger,
	}
}

// RouterTrigger specifies routing target
type RouterTrigger struct {
	router *Router
	fromAgent string
	trigger string
}

// To routes to target agent
func (rt *RouterTrigger) To(targetAgent string) *RouterFrom {
	rule := &RoutingRule{
		FromAgent: rt.fromAgent,
		Trigger: rt.trigger,
		TargetAgent: &targetAgent,
	}
	rt.router.rules[rt.fromAgent] = append(
		rt.router.rules[rt.fromAgent],
		rule,
	)
	return &RouterFrom{
		router: rt.router,
		agentID: rt.fromAgent,
	}
}

// Route determines next agent based on trigger
func (r *Router) Route(fromAgent, trigger string) (string, error) {
	rules, ok := r.rules[fromAgent]
	if !ok {
		return "", fmt.Errorf("no routing rules for %s", fromAgent)
	}

	for _, rule := range rules {
		if rule.Trigger == trigger || rule.Trigger == "always" {
			if rule.TargetAgent == nil {
				return "", nil  // Terminal
			}
			return *rule.TargetAgent, nil
		}
	}

	return "", fmt.Errorf("no matching route for %s:%s", fromAgent, trigger)
}

// TriggerDetector detects trigger from agent response
type TriggerDetector interface {
	Detect(agentResponse string) string
}

// SimpleTriggerDetector looks for [TRIGGER_*] patterns
type SimpleTriggerDetector struct {
	triggerPrefix string
}

func (std *SimpleTriggerDetector) Detect(response string) string {
	// Parse [TRIGGER_SOMETHING] from response
	re := regexp.MustCompile(`\[TRIGGER_([A-Z_]+)\]`)
	matches := re.FindStringSubmatch(response)
	if len(matches) > 1 {
		return strings.ToLower(matches[1])
	}
	return "no_trigger"
}
```

### 3.4 Usage

**Builder pattern**:
```go
router := agentic.NewRouter().
    From("orchestrator").
        OnTrigger("needs_info").To("clarifier").
        OnTrigger("ready").To("executor").
    From("clarifier").
        OnTrigger("complete").To("executor").
    From("executor").
        OnTrigger("always").To("")  // Terminal

team.Router = router
```

**From YAML**:
```yaml
routing:
  rules:
    - from_agent: orchestrator
      trigger: needs_info
      target_agent: clarifier
```

### 3.5 System Prompt Impact

**Before** (145 lines with routing logic):
```
Jika [ROUTE_CLARIFIER] thì output "[ROUTE_CLARIFIER]"
Jika [ROUTE_EXECUTOR] thì output "[ROUTE_EXECUTOR]"
...
```

**After** (simplified):
```
Nếu bạn cần thêm thông tin, output [TRIGGER_NEEDS_INFO]
Nếu bạn sẵn sàng, output [TRIGGER_READY]
Nếu hoàn thành, output [TRIGGER_COMPLETE]
```

**System prompt generation becomes**:
```go
func buildSystemPrompt(agent *Agent, router *Router) string {
    // Base prompt
    prompt := fmt.Sprintf(
        "You are %s. Role: %s.\n%s\n\n",
        agent.Name, agent.Role, agent.Backstory,
    )

    // Add trigger instructions from router
    if rules, ok := router.rules[agent.ID]; ok {
        prompt += "When you need to:\n"
        for _, rule := range rules {
            prompt += fmt.Sprintf(
                "- %s: output [TRIGGER_%s]\n",
                rule.Trigger,
                strings.ToUpper(rule.Trigger),
            )
        }
    }

    return prompt
}
```

### 3.6 Effort Breakdown

| Task | Time | Output |
|------|------|--------|
| Router implementation | 1.5h | routing.go (~150 lines) |
| TriggerDetector | 1h | routing.go (+60 lines) |
| System prompt refactoring | 1.5h | agent.go changes |
| Tests | 1h | routing_test.go |
| IT Support v3 with routing | 1h | Updated example |

**Total**: 6 hours

---

## Examples & Migration (12h)

### 4.1 Example Updates

Create updated examples showing new patterns:

**examples/simple-chat-v2/** (With Fluent API)
```go
team := agentic.NewTeam().
    AddAgent(agentic.NewAgent("bot1", "Enthusiast").
        WithRole("Ask questions").
        WithBackstory("You are curious...").
        Build()).
    AddAgent(agentic.NewAgent("bot2", "Expert").
        WithRole("Answer questions").
        WithBackstory("You are knowledgeable...").
        SetTerminal(true).
        Build()).
    Build()
```

**examples/it-support-unified/** (With Unified YAML)
```
team.yaml (single file)
```

```go
toolHandlers := map[string]agentic.ToolHandler{
    "get_cpu_usage": getCPUHandler,
    // ...
}

team, _ := agentic.LoadTeamFromYAML("team.yaml", toolHandlers)
```

### 4.2 Migration Guide

`MIGRATION_UX_V1_V2.md`:
```markdown
# Migration Guide: go-agentic v0.0.1 → v0.0.2 UX Improvements

## What's New

### 1. Fluent Agent Builder (Recommended)
Instead of:
```go
agent := &agentic.Agent{ID: "id", Name: "Name", ...}
```

Use:
```go
agent := agentic.NewAgent("id", "Name").
    WithRole("role").
    Build()
```

## Backward Compatibility

✅ Old code still works!
✅ Can mix old and new patterns
✅ Gradual migration possible
```

### 4.3 Documentation

Create comprehensive docs:
- `FLUENT_API_GUIDE.md` - How to use builders
- `UNIFIED_CONFIG_GUIDE.md` - How to use single YAML
- `ROUTING_GUIDE.md` - How to configure routing

---

## 📊 Total Implementation Summary

| Phase | Duration | Impact | Files |
|-------|----------|--------|-------|
| 1: Fluent API | 6h | 40% boilerplate reduction | builder.go |
| 2: Unified Config | 8h | Single config file | config_unified.go |
| 3: Routing DSL | 6h | 80% system prompt reduction | routing.go |
| Examples & Docs | 12h | Clear patterns | 5 examples + 3 guides |
| **Total** | **32h** | **Significant UX improvement** | **8 new files** |

---

## 🎯 Success Criteria

After implementation:

✅ **Simplicity**: 50% less code for basic setup
✅ **Clarity**: One way to do each thing
✅ **Configuration**: Single file for team setup
✅ **Routing**: Declarative, not text-based
✅ **Documentation**: Clear examples for each pattern
✅ **Compatibility**: Zero breaking changes

---

**Tác giả**: Implementation Planning
**Ngày**: 20 tháng 12 năm 2025
**Tình trạng**: Ready for development

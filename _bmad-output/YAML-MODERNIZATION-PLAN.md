# YAML Modernization Implementation Plan

**Created:** 2025-12-22
**Phase:** Design & Planning
**Scope:** Standardize and modernize Agent + Crew YAML configuration system
**Priority:** FOUNDATION-LEVEL (Must fix before scaling)

---

## Overview

Before implementing the Message History Limit fix, we should establish a solid, modern YAML configuration foundation. This document outlines a phased approach to modernize the configuration system while maintaining backward compatibility.

### Why This Matters First

1. **Foundation Stability:** Message history fix needs clear configuration patterns
2. **Scaling Support:** Modern system can handle complex multi-agent workflows
3. **Production Readiness:** Explicit validation prevents configuration errors
4. **Team Efficiency:** Clear standards reduce debugging time
5. **Future Flexibility:** Schema v2.0 enables advanced features without rewrites

---

## Phase 1: Signal Schema Validation (CRITICAL - Week 1)

### Objective
Implement explicit signal validation to prevent runtime signal mismatches.

### Changes

#### 1.1 Add Signal Schema to Crew YAML

```yaml
# In crew.yaml - add this section
signal_schema:
  validation_enabled: true
  
  patterns:
    routing:
      pattern: "^\\[ROUTE_[A-Z_]+\\]$"
      description: "Agent routing signals"
      required: true
    
    status:
      pattern: "^\\[COMPLETE|ERROR|RETRY|KẾT_THÚC\\]$"
      description: "Status signals"
      required: false
    
    custom:
      pattern: "^\\[[A-Z_][A-Z_0-9]*\\]$"
      description: "Custom signals"
      required: false
```

#### 1.2 Add Signal Validation to Agent Config

```yaml
# In agent/*.yaml - add this field
allowed_signals:
  - "[ROUTE_EXECUTOR]"        # Must match signal_schema pattern
  - "[ROUTE_CLARIFIER]"
  - "[ERROR_CONDITION]"

signal_instructions: |
  Always end response with one of these signals:
  - [ROUTE_EXECUTOR] when ready for diagnosis
  - [ROUTE_CLARIFIER] when more info needed
  - [ERROR_CONDITION] if error occurred
```

#### 1.3 Implement Validation Logic

**File:** `core/validation.go`

```go
type SignalValidator struct {
    Pattern     *regexp.Regexp
    Required    bool
    Description string
}

type CrewSignalConfig struct {
    ValidationEnabled bool
    Patterns          map[string]*SignalValidator
}

func (v *ConfigValidator) ValidateSignals(
    crew *Crew,
    agents map[string]*Agent,
) error {
    // For each agent
    for _, agentID := range crew.Agents {
        agent := agents[agentID]
        
        // Check allowed_signals against schema
        for _, signal := range agent.AllowedSignals {
            if !v.matchesAnyPattern(signal, crew.SignalSchema) {
                return fmt.Errorf(
                    "Agent %s: signal %s doesn't match any schema pattern",
                    agentID, signal,
                )
            }
        }
        
        // Check system_prompt contains required signals
        requiredSignals := v.extractSignals(agent.SystemPrompt)
        for _, signal := range requiredSignals {
            if !contains(agent.AllowedSignals, signal) {
                return fmt.Errorf(
                    "Agent %s: system_prompt uses signal %s not in allowed_signals",
                    agentID, signal,
                )
            }
        }
    }
    return nil
}

func (v *ConfigValidator) extractSignals(text string) []string {
    // Regex to find all [SIGNAL] patterns
    re := regexp.MustCompile(`\[([A-Z_][A-Z_0-9]*)\]`)
    matches := re.FindAllString(text, -1)
    return matches
}
```

#### 1.4 Update Config Loaders

**File:** `core/config.go`

```go
func LoadCrewConfig(path string) (*Crew, error) {
    // ... existing load logic ...
    
    // NEW: Validate signals
    validator := NewConfigValidator()
    if err := validator.ValidateSignals(crew); err != nil {
        return nil, fmt.Errorf("signal validation failed: %w", err)
    }
    
    return crew, nil
}

func LoadAgentConfigs(dir string) (map[string]*AgentConfig, error) {
    configs := make(map[string]*AgentConfig)
    
    files, _ := filepath.Glob(filepath.Join(dir, "*.yaml"))
    for _, file := range files {
        config, err := loadAgentFile(file)
        if err != nil {
            return nil, err
        }
        
        // NEW: Validate allowed_signals format
        if err := validateAllowedSignals(config); err != nil {
            return nil, fmt.Errorf("%s: %w", filepath.Base(file), err)
        }
        
        configs[config.ID] = config
    }
    
    return configs, nil
}

func validateAllowedSignals(config *AgentConfig) error {
    re := regexp.MustCompile(`^\[[A-Z_][A-Z_0-9]*\]$`)
    for _, signal := range config.AllowedSignals {
        if !re.MatchString(signal) {
            return fmt.Errorf(
                "invalid signal format '%s' (must be [SIGNAL_NAME])",
                signal,
            )
        }
    }
    return nil
}
```

#### 1.5 Update Types

**File:** `core/types.go`

```go
type Agent struct {
    // ... existing fields ...
    AllowedSignals   []string  // NEW: Signals this agent can emit
    SignalInstructions string   // NEW: Instructions for signal usage
}

type Crew struct {
    // ... existing fields ...
    SignalSchema     *CrewSignalConfig  // NEW: Signal validation schema
}

type CrewSignalConfig struct {
    ValidationEnabled bool
    Patterns          map[string]*SignalValidator
}

type SignalValidator struct {
    Pattern     string  // Regex pattern
    Description string
    Required    bool
    Examples    []string
}
```

### Deliverables

- ✅ Updated types.go with signal fields
- ✅ Enhanced validation.go with signal validation logic
- ✅ Updated config.go to load and validate signals
- ✅ IT Support example updated with signal schema
- ✅ Unit tests for signal validation
- ✅ Documentation: `docs/YAML-SIGNALS.md`

### Testing

```go
func TestSignalSchemaValidation(t *testing.T) {
    // Test: Valid signals pass
    // Test: Invalid signal format rejected
    // Test: Missing required signals detected
    // Test: Undefined signals in system_prompt caught
    // Test: Signal schema patterns work correctly
}
```

---

## Phase 2: Tool Parameter Specification (HIGH - Week 1-2)

### Objective
Enable per-agent tool parameter customization and validation.

### Changes

#### 2.1 Add Tool Specs to Agent YAML

```yaml
# In agent/executor.yaml
tools:
  - name: GetCPUUsage
    parameters:
      threshold: 80                    # Override default
      include_per_core: true
    behavior:
      max_calls: 5                     # Max uses per request
      timeout: 10s
      cache: true
      retry_on_failure: true
    expected_output:
      format: json
      schema:
        usage: {type: number, description: "CPU usage %"}
        cores: {type: array}
  
  - name: GetMemoryUsage
    parameters:
      units: gb
    behavior:
      max_calls: 3
      timeout: 10s
```

#### 2.2 Implement Tool Spec Validation

**File:** `core/validation.go`

```go
type ToolSpec struct {
    Name             string
    Parameters       map[string]interface{}
    Behavior         ToolBehavior
    ExpectedOutput   OutputSchema
}

type ToolBehavior struct {
    MaxCalls         int
    Timeout          time.Duration
    Cache            bool
    RetryOnFailure   bool
}

func (v *ConfigValidator) ValidateToolSpec(
    spec *ToolSpec,
    agent *Agent,
    tools map[string]*Tool,
) error {
    // Check tool exists
    tool, ok := tools[spec.Name]
    if !ok {
        return fmt.Errorf("tool %s not found in registry", spec.Name)
    }
    
    // Validate parameters against tool schema
    for param, value := range spec.Parameters {
        // Check parameter exists in tool definition
        // Validate parameter type
        // Validate parameter value range
    }
    
    // Validate behavior settings
    if spec.Behavior.MaxCalls <= 0 {
        return fmt.Errorf("max_calls must be > 0")
    }
    if spec.Behavior.Timeout <= 0 {
        return fmt.Errorf("timeout must be > 0")
    }
    
    return nil
}
```

#### 2.3 Update Execution Logic

**File:** `core/crew.go`

```go
// Before executing tool, apply per-agent specs
func (ce *CrewExecutor) executeToolWithSpec(
    ctx context.Context,
    tool *Tool,
    agent *Agent,
    spec *ToolSpec,
) (string, error) {
    // Apply per-agent timeout
    if spec.Behavior.Timeout > 0 {
        var cancel context.CancelFunc
        ctx, cancel = context.WithTimeout(ctx, spec.Behavior.Timeout)
        defer cancel()
    }
    
    // Apply per-agent parameters
    args := mergeParameters(tool.Parameters, spec.Parameters)
    
    // Execute tool
    return safeExecuteTool(ctx, tool, args)
}

func mergeParameters(
    defaults map[string]interface{},
    overrides map[string]interface{},
) map[string]interface{} {
    // Start with defaults
    merged := make(map[string]interface{})
    for k, v := range defaults {
        merged[k] = v
    }
    
    // Override with agent-specific values
    for k, v := range overrides {
        merged[k] = v
    }
    
    return merged
}
```

### Deliverables

- ✅ Updated types.go with ToolSpec
- ✅ Enhanced validation.go for tool parameter validation
- ✅ Updated crew.go execution logic to use tool specs
- ✅ IT Support example updated with tool specs
- ✅ Unit tests for tool spec validation
- ✅ Documentation: `docs/YAML-TOOLS.md`

---

## Phase 3: Error Handling Policies (HIGH - Week 2)

### Objective
Add structured error handling to improve resilience.

### Changes

#### 3.1 Add Error Policies to YAML

```yaml
# In crew.yaml
error_policies:
  default:
    on_timeout:
      action: retry
      max_attempts: 3
      backoff: exponential
    
    on_tool_failure:
      action: continue
      log_level: warning
    
    on_validation_error:
      action: escalate

# In agent/executor.yaml - per-agent overrides
error_handling:
  on_timeout:
    action: retry
    max_attempts: 5
    backoff: exponential
```

#### 3.2 Implement Error Policy Handling

**File:** `core/crew.go`

```go
type ErrorPolicy struct {
    OnTimeout      TimeoutPolicy
    OnToolFailure  ToolFailurePolicy
    OnValidation   ValidationPolicy
}

type TimeoutPolicy struct {
    Action      string          // retry, skip, escalate
    MaxAttempts int
    Backoff     string          // linear, exponential
}

func (ce *CrewExecutor) executeToolWithPolicy(
    ctx context.Context,
    tool *Tool,
    agent *Agent,
) (string, error) {
    policy := ce.getErrorPolicy(agent.ID)
    
    for attempt := 0; attempt < policy.OnTimeout.MaxAttempts; attempt++ {
        output, err := safeExecuteTool(ctx, tool, args)
        
        if err == nil {
            return output, nil
        }
        
        // Check error type
        if errors.Is(err, context.DeadlineExceeded) {
            switch policy.OnTimeout.Action {
            case "retry":
                backoff := calculateBackoff(attempt, policy.OnTimeout.Backoff)
                time.Sleep(backoff)
                continue
            case "skip":
                return "", nil
            case "escalate":
                return "", err
            }
        }
        
        // Handle other errors...
    }
    
    return "", fmt.Errorf("exhausted retry attempts")
}
```

### Deliverables

- ✅ Error policy types in types.go
- ✅ Error handling logic in crew.go
- ✅ Policy configuration loading
- ✅ Unit tests for error policies
- ✅ Documentation: `docs/YAML-ERROR-HANDLING.md`

---

## Phase 4: Template Variables Extension (HIGH - Week 3)

### Objective
Extend template variables for dynamic system prompt generation.

### Changes

#### 4.1 Extend Template Context

**File:** `core/config.go`

```go
type TemplateContext struct {
    // Standard variables
    Name            string
    Role            string
    Description     string
    Backstory       string
    
    // NEW: Extended variables
    Tools           []string              // Available tools
    ToolDescriptions map[string]string    // Tool descriptions
    HandoffTargets  []string              // Routing destinations
    AllowedSignals  []string              // Signals to emit
    Constraints     []string              // Operating constraints
    Examples        []map[string]string   // Input/output examples
    
    // Computed
    ToolsList       string                // Formatted list
    SignalList      string                // Formatted signals
    ConstraintsList string                // Formatted constraints
}

func (ctx *TemplateContext) Format(template string) (string, error) {
    // Create formatter with all variables
    funcMap := template.FuncMap{
        "tools_bullet_list": func() string {
            return ctx.formatBulletList(ctx.Tools)
        },
        "signals_comma_separated": func() string {
            return ctx.formatCommaSeparated(ctx.AllowedSignals)
        },
        // ... more formatters ...
    }
    
    // Parse and execute template
    tmpl, err := template.New("system_prompt").
        Funcs(funcMap).
        Parse(template)
    if err != nil {
        return "", err
    }
    
    var buf bytes.Buffer
    if err := tmpl.Execute(&buf, ctx); err != nil {
        return "", err
    }
    
    return buf.String(), nil
}
```

#### 4.2 Build Template Context

**File:** `core/crew.go`

```go
func (ce *CrewExecutor) buildTemplateContext(agent *Agent) *TemplateContext {
    // Gather tools for this agent
    tools := make([]string, len(agent.Tools))
    for i, tool := range agent.Tools {
        tools[i] = tool.Name
    }
    
    // Get handoff targets
    targets := agent.HandoffTargets
    
    // Get allowed signals
    signals := agent.AllowedSignals
    
    // Build context
    return &TemplateContext{
        Name:            agent.Name,
        Role:            agent.Role,
        Description:     agent.Description,
        Backstory:       agent.Backstory,
        Tools:           tools,
        HandoffTargets:  targets,
        AllowedSignals:  signals,
        ToolsList:       formatToolsList(tools),
        SignalList:      formatSignalList(signals),
    }
}
```

### Deliverables

- ✅ Extended TemplateContext type
- ✅ Template formatting utilities
- ✅ System prompt template rendering
- ✅ Unit tests for template expansion
- ✅ Documentation: `docs/YAML-TEMPLATES.md`

---

## Phase 5: Execution Graph Support (MEDIUM - Week 3-4)

### Objective
Explicit dependency management and parallelization.

### Changes

#### 5.1 Add Execution Graph to Crew YAML

```yaml
execution_graph:
  phases:
    phase_1:
      agents: [orchestrator]
      parallel: false
      depends_on: []
      timeout: 60s
    
    phase_2:
      agents: [clarifier]
      parallel: false
      depends_on: [phase_1]
      timeout: 120s
    
    phase_3:
      agents: [executor]
      parallel: false
      depends_on: [phase_2]
      timeout: 120s
```

#### 5.2 Implement Graph Execution

**File:** `core/crew.go`

```go
type ExecutionPhase struct {
    Name       string
    Agents     []string
    Parallel   bool
    Timeout    time.Duration
    DependsOn  []string
}

func (ce *CrewExecutor) ExecuteWithGraph(
    ctx context.Context,
    input string,
    graph *ExecutionGraph,
) (*CrewResponse, error) {
    for _, phase := range graph.Phases {
        // Execute phase
        if phase.Parallel {
            results, err := ce.ExecuteParallel(ctx, input, agents)
        } else {
            results, err := ce.executeSequential(ctx, input, agents)
        }
    }
}
```

### Deliverables

- ✅ ExecutionGraph and ExecutionPhase types
- ✅ Graph execution logic
- ✅ Phase dependency validation
- ✅ Unit tests for graph execution
- ✅ Documentation: `docs/YAML-EXECUTION-GRAPH.md`

---

## Phase 6: Behavior Configuration Enhancement (MEDIUM - Week 4)

### Objective
Extend behavior customization options.

### Changes

Add to agent config:

```yaml
behavior:
  priority: high
  cache_results: true
  context_injection:
    prefix: "Previous context..."
    max_history_depth: 5
  retry_policy:
    max_attempts: 3
    backoff: exponential
```

---

## Implementation Timeline

```
Week 1:
  ├─ Phase 1: Signal Schema Validation
  └─ Phase 2: Tool Parameter Specification (start)

Week 2:
  ├─ Phase 2: Tool Parameter Specification (finish)
  └─ Phase 3: Error Handling Policies

Week 3:
  ├─ Phase 4: Template Variables Extension
  └─ Phase 5: Execution Graph Support (start)

Week 4:
  ├─ Phase 5: Execution Graph Support (finish)
  └─ Phase 6: Behavior Configuration Enhancement

Post-Implementation:
  ├─ Comprehensive testing
  ├─ Documentation finalization
  ├─ Migration guide for v1.0 → v2.0
  └─ Team training
```

---

## Success Criteria

### Phase 1 (Signal Validation)
- [ ] All signal patterns validated at load time
- [ ] Mismatched signals detected and reported
- [ ] IT Support example passes validation
- [ ] Tests cover all signal patterns

### Phase 2 (Tool Parameters)
- [ ] Per-agent tool customization working
- [ ] Parameters validated against schema
- [ ] Tool specs applied during execution
- [ ] Tests verify parameter merging

### Phase 3 (Error Policies)
- [ ] Error handling policies applied
- [ ] Retry logic working with backoff
- [ ] Error escalation functioning
- [ ] Tests cover all error scenarios

### Phase 4 (Template Variables)
- [ ] All template variables available
- [ ] System prompts dynamically generated
- [ ] Template formatting working
- [ ] Tests verify template expansion

### Phase 5 (Execution Graph)
- [ ] Graph execution working
- [ ] Dependencies validated
- [ ] Parallel and sequential modes working
- [ ] Tests cover graph scenarios

### Phase 6 (Behavior Config)
- [ ] Enhanced behavior options working
- [ ] Caching implemented
- [ ] Context injection working
- [ ] Tests verify behavior options

---

## Benefits Upon Completion

### Immediate
✅ Explicit signal validation prevents runtime errors
✅ Tool customization enables flexible agent configuration
✅ Error policies improve resilience

### Short-term (1-2 months)
✅ Template variables reduce configuration boilerplate
✅ Execution graphs enable complex workflows
✅ Enhanced behavior options support diverse use cases

### Long-term (3-6 months)
✅ Foundation for multi-team agent systems
✅ Support for enterprise-scale workflows
✅ Configuration as code best practices established

---

## Backward Compatibility

All changes maintain backward compatibility with v1.0 configs:

```go
func LoadCrewConfig(path string) (*Crew, error) {
    data := unmarshalYAML(path)
    
    // Auto-detect version
    version := detectVersion(data)
    
    switch version {
    case "1.0":
        return migrateFromV1(data)  // Apply compat layer
    case "2.0":
        return parseV2(data)         // New schema
    }
}
```

---

## Next Steps

1. **Review this plan** with team for feedback
2. **Prioritize phases** based on needs
3. **Create tracking issues** for each phase
4. **Begin Phase 1** implementation
5. **Document as we go** to build migration guide

---

## Conclusion

This YAML modernization plan establishes a solid foundation for go-agentic to scale from IT Support example to enterprise multi-agent systems. The phased approach allows incremental progress while maintaining stability and backward compatibility.

**Recommended:** Complete Phases 1-3 (foundation) before scaling to complex workflows. Phases 4-6 can be done based on use case requirements.


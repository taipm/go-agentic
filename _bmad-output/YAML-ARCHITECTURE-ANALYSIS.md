# YAML Architecture Analysis & Standardization Plan

**Created:** 2025-12-22
**Status:** Analysis Complete - Ready for Design Phase
**Focus:** Modernize Agent + Crew YAML structure for production use

---

## Executive Summary

Based on comprehensive analysis of go-agentic codebase, the current YAML configuration system uses **signal-based explicit routing** (IT Support example as canonical). However, there are several opportunities for standardization and modernization:

### Key Findings

✅ **Current Strengths:**
- Signal-based routing prevents ambiguous handoffs
- Comprehensive validation catches errors early
- Clear agent type separation (router, gatherer, executor)
- Flexible tool assignment per agent
- Thread-safe concurrent execution support

⚠️ **Gaps Identified:**
- Signal patterns hardcoded in system prompts (not schema-driven)
- Limited behavior customization beyond routing
- Tool parameter validation missing at config level
- Parallel execution not documented in examples
- Template variables limited to 4 fields
- No inline tool definitions or parameter specs
- Agent error handling policies missing
- Dependency declaration not supported

---

## Part 1: Current YAML Structures

### Agent YAML (Current Production Schema)

**File:** `examples/it-support/config/agents/executor.yaml`

```yaml
# REQUIRED FIELDS
id: executor                               # Unique identifier
name: System Diagnostics Agent             # Display name
role: Diagnostic execution and analysis    # Agent responsibility

# OPTIONAL FIELDS
description: Performs system diagnostics   # Brief purpose
backstory: |                              # Personality/context
  You are an expert system administrator...
  With deep knowledge of Linux/Windows...

# LLM CONFIGURATION
model: gpt-4o-mini                        # OpenAI model
temperature: 0.7                          # 0-1 range (validated)

# WORKFLOW CONFIGURATION
is_terminal: true                         # Final agent in pipeline
tools:                                    # Available tool names
  - GetCPUUsage
  - GetMemoryUsage
  - GetDiskSpace
  - ... (15+ tools)
handoff_targets: []                       # No further routing

# CUSTOM BEHAVIOR (OPTIONAL)
system_prompt: |                          # Override default system prompt
  Custom instructions...
```

**Validation Rules:**
```go
REQUIRED: id, name, role (all non-empty)
OPTIONAL: description, backstory, system_prompt
CONSTRAINTS:
  - temperature: 0 ≤ temp ≤ 1 (error if > 1)
  - model: warn if not in approved list
  - tools: all must exist in tool registry
  - handoff_targets: all must be valid agent IDs
```

### Crew YAML (Current Production Schema)

**File:** `examples/it-support/config/crew.yaml`

```yaml
# REQUIRED FIELDS
version: "1.0"                            # Config schema version
entry_point: orchestrator                 # Starting agent ID
agents:                                   # Agent IDs to load
  - orchestrator
  - clarifier
  - executor

# SETTINGS
settings:
  max_handoffs: 5                        # Max routing transitions
  max_rounds: 10                         # Max conversation rounds
  timeout_seconds: 300                   # Execution timeout

# ROUTING CONFIGURATION (NEW PATTERN)
routing:
  signals:                               # Signal-based routing
    orchestrator:
      - signal: "[ROUTE_EXECUTOR]"      # Signal pattern
        target: executor                 # Target agent ID
        description: "Route to executor"
  
  defaults:                              # Fallback routing
    orchestrator: clarifier              # Default target
  
  agent_behaviors:                       # Per-agent behavior
    orchestrator:
      wait_for_signal: true
      auto_route: false
      is_terminal: false
    
    executor:
      is_terminal: true
```

**Validation Rules:**
```go
REQUIRED: version, entry_point, agents
CONSTRAINTS:
  - entry_point must exist in agents list
  - agents list must not be empty
  - All signal targets must be valid agent IDs
  - max_handoffs ≥ 1
  - No circular routing patterns
  - WARNING: unreachable agents from entry_point
```

---

## Part 2: Issues with Current Structure

### Issue 1: Signal Hardcoding

**Problem:**
```yaml
# Signals embedded in system prompts (not schema-validated)
system_prompt: |
  Always end with [ROUTE_EXECUTOR] or [ROUTE_CLARIFIER]
```

**Risk:**
- Agent can emit wrong signal format
- No validation that agent uses correct signals
- Signal patterns not discoverable from config alone
- Difficult to maintain consistency across agents

**Solution Needed:**
```yaml
# Schema-driven signals with validation
allowed_signals:
  - pattern: "^\\[ROUTE_[A-Z_]+\\]$"
    required: true
    examples: ["[ROUTE_EXECUTOR]", "[ROUTE_CLARIFIER]"]
```

### Issue 2: Limited Template Variables

**Problem:**
```yaml
# Only 4 variables available
system_prompt: |
  You are {{name}} with role {{role}}.
  {{backstory}}
  {{description}}
```

**Missing:**
- `{{tools}}` - Auto-list available tools
- `{{handoff_targets}}` - Auto-list routing options
- `{{context}}` - Workflow context injection
- `{{examples}}` - Example inputs/outputs
- `{{constraints}}` - Operating constraints

**Solution Needed:**
```yaml
system_prompt: |
  You are {{name}} with role {{role}}.
  {{backstory}}
  
  Available tools: {{tools:comma-separated}}
  Can route to: {{handoff_targets:list-format}}
  
  Constraints: {{constraints}}
  Examples: {{examples:json}}
```

### Issue 3: Tool Definition Limitations

**Problem:**
```yaml
# Tools only by name - no parameter spec or per-agent validation
tools:
  - GetCPUUsage
  - GetMemoryUsage
```

**Missing:**
- Tool parameters per agent
- Tool usage constraints (max calls, etc.)
- Tool success/failure handling
- Tool output format expectations

**Solution Needed:**
```yaml
tools:
  - name: GetCPUUsage
    parameters:
      threshold: 80        # Optional param override for this agent
      format: percentage   # Expected output format
    max_calls: 5          # Max uses per request
    required: true        # Must use this tool
  
  - name: GetMemoryUsage
    parameters:
      units: gb           # bytes, mb, gb
    max_calls: 3
```

### Issue 4: Error Handling Not Specified

**Problem:**
```go
// No policy defined for agent errors
if err != nil {
    return nil, fmt.Errorf("agent %s failed: %w", agentID, err)  // Hard fail
}
```

**Missing:**
- Retry policies per agent
- Fallback behavior
- Error recovery strategies
- Timeout handling policies

**Solution Needed:**
```yaml
agents:
  executor:
    error_handling:
      on_timeout:
        action: retry                  # retry, skip, escalate
        max_attempts: 3
        backoff: exponential
      
      on_tool_failure:
        action: continue              # continue, stop, escalate
        fallback_tool: alternate_name
      
      on_validation_error:
        action: escalate              # escalate to human
```

### Issue 5: Dependency Declaration Missing

**Problem:**
```yaml
# No way to specify agent dependencies
agents:
  - research-lead
  - research-analyst    # Depends on research-lead output
  - qa-reviewer         # Depends on both above
```

**Missing:**
- Explicit dependency declarations
- Parallel execution grouping
- Sequential execution requirements
- Context passing between agents

**Solution Needed:**
```yaml
execution_graph:
  research-lead:
    parallel_with: []
    depends_on: []
  
  research-analyst:
    parallel_with: []
    depends_on: [research-lead]
  
  qa-reviewer:
    parallel_with: []
    depends_on: [research-lead, research-analyst]
    aggregation: merge_results
```

### Issue 6: Behavior Customization Limited

**Problem:**
```yaml
# Only 3 behavior options
agent_behaviors:
  orchestrator:
    wait_for_signal: true
    auto_route: false
    is_terminal: false
```

**Missing:**
- Retry policies
- Fallback rules
- Timeout overrides
- Context injection
- Priority levels

**Solution Needed:**
```yaml
agent_behaviors:
  orchestrator:
    # Routing
    wait_for_signal: true
    auto_route: false
    is_terminal: false
    
    # Execution
    retry_policy:
      max_attempts: 3
      backoff: exponential
      timeout_override: 30s
    
    # Behavior
    priority: high
    fallback_agent: clarifier
    context_injection:
      prefix: "Previous decisions: {{history}}"
```

---

## Part 3: Standardization Recommendations

### Recommendation 1: Signal Schema Definition

**Implement explicit signal validation:**

```yaml
# In crew.yaml
signal_schema:
  patterns:
    routing:
      pattern: "^\\[ROUTE_[A-Z_]+\\]$"
      examples: ["[ROUTE_EXECUTOR]", "[ROUTE_CLARIFIER]"]
      description: "Explicit agent routing signals"
    
    status:
      pattern: "^\\[COMPLETE|ERROR|RETRY\\]$"
      examples: ["[COMPLETE]", "[ERROR]"]
      description: "Status signals for terminal agents"
    
    localized:
      pattern: "^\\[[A-Z_]+\\]$"
      examples: ["[KẾT_THÚC]", "[HOÀN_THÀNH]"]
      description: "Localized status signals"

# In agent config (orchestrator.yaml)
allowed_signals:
  - "[ROUTE_EXECUTOR]"
  - "[ROUTE_CLARIFIER]"
  - "[ERROR_CONDITION]"
```

**Validation Logic:**
```go
func (v *ConfigValidator) ValidateSignals(agent *Agent, crew *Crew) error {
    // Extract signals agent should emit (from system_prompt)
    expectedSignals := detectSignalsInPrompt(agent.SystemPrompt)
    
    // Compare against allowed_signals
    allowedSignals := crew.SignalSchema[agent.ID]
    
    for _, signal := range expectedSignals {
        if !isAllowed(signal, allowedSignals) {
            return fmt.Errorf(
                "Agent %s uses signal %s not in allowed_signals",
                agent.ID, signal,
            )
        }
    }
}
```

### Recommendation 2: Enhanced Template Variables

**Implement dynamic template variable system:**

```go
type TemplateContext struct {
    // Standard variables
    Name            string
    Role            string
    Description     string
    Backstory       string
    
    // New variables
    Tools           []string              // Available tools
    ToolsFormatted  string               // Formatted list
    HandoffTargets  []string             // Routing destinations
    HandoffNames    string               // Formatted names
    Context         map[string]interface{} // Custom context
    Constraints     []string             // Operating constraints
    Examples        []map[string]string  // Input/output examples
    
    // Computed variables
    SignalList      []string             // Required signals
    RetryPolicy     string               // Retry behavior doc
    TimeoutValue    string               // Timeout in human form
}
```

**Usage in YAML:**
```yaml
system_prompt: |
  你好, 我是 {{name}}
  
  我的角色: {{role}}
  {{description}}
  {{backstory}}
  
  ## 可用工具
  {{tools:formatted-as-bullet-list}}
  
  ## 路由
  我可以转交给以下代理:
  {{handoff_targets:formatted-as-links}}
  
  ## 信号指示
  我必须发出这些信号之一:
  {{signal_list:formatted-as-bullets}}
  
  ## 约束
  {{constraints:formatted-as-bullets}}
  
  ## 示例
  Input: {{examples[0].input}}
  Output: {{examples[0].output}}
```

**Implementation:**
```go
func (c *ConfigLoader) RenderSystemPrompt(
    agent *Agent,
    context TemplateContext,
) (string, error) {
    tmpl, err := template.New("system_prompt").Parse(agent.SystemPrompt)
    if err != nil {
        return "", err
    }
    
    var buf bytes.Buffer
    err = tmpl.Execute(&buf, context)
    return buf.String(), err
}
```

### Recommendation 3: Tool Parameter Specification

**Add tool specifications at config level:**

```yaml
# agent/executor.yaml
tools:
  - name: GetCPUUsage
    parameters:
      threshold: 80                    # Optional threshold
      include_per_core: true
    behavior:
      max_calls: 5                     # Max uses in this request
      timeout: 10s                     # Tool-specific timeout
      cache: true                      # Cache results
      retry_on_failure: true
    expected_output:
      format: json
      schema:
        type: object
        properties:
          usage: {type: number}
          cores: {type: array}
  
  - name: GetMemoryUsage
    parameters:
      units: gb
    behavior:
      max_calls: 3
      timeout: 10s
```

**Validation:**
```go
type ToolSpec struct {
    Name              string
    Parameters        map[string]interface{}    // Override params for this agent
    MaxCalls          int                       // Rate limiting
    Timeout           time.Duration             // Tool-specific timeout
    CacheResults      bool                      // Cache tool output
    RetryOnFailure    bool
    ExpectedOutput    OutputSchema
}

func (v *ConfigValidator) ValidateToolSpec(spec ToolSpec, tool *Tool) error {
    // Validate tool exists in registry
    // Validate parameters match tool schema
    // Validate timeout is reasonable
    // Validate output schema matches tool's actual output
}
```

### Recommendation 4: Error Handling Policies

**Add structured error handling:**

```yaml
# crew.yaml
error_policies:
  default:
    on_timeout:
      action: retry
      max_attempts: 3
      backoff_strategy: exponential  # linear, exponential, fixed
      initial_backoff: 1s
      max_backoff: 30s
    
    on_tool_failure:
      action: continue               # continue, skip, escalate, retry
      log_level: warning
      fallback_tool: null
    
    on_validation_error:
      action: escalate               # escalate to user
      notify: true
    
    on_max_handoffs:
      action: terminate              # terminate, escalate
      reason: "Max routing limit exceeded"

# agent/executor.yaml
error_handling:
  # Override defaults for this agent
  on_timeout:
    action: retry
    max_attempts: 5
    backoff_strategy: exponential
  
  on_tool_failure:
    action: escalate
    fallback_tool: GeneralDiagnostics
```

**Implementation:**
```go
type ErrorPolicy struct {
    OnTimeout      TimeoutPolicy
    OnToolFailure  ToolFailurePolicy
    OnValidation   ValidationPolicy
    OnMaxHandoffs  HandoffPolicy
}

type TimeoutPolicy struct {
    Action          string              // retry, skip, escalate
    MaxAttempts     int
    BackoffStrategy string              // linear, exponential
    InitialBackoff  time.Duration
    MaxBackoff      time.Duration
}

func (ce *CrewExecutor) HandleError(err error, policy *ErrorPolicy) (string, error) {
    switch policy.OnTimeout.Action {
    case "retry":
        return ce.retryWithBackoff(err, policy)
    case "skip":
        return "", nil  // Continue with next step
    case "escalate":
        return "", fmt.Errorf("escalated: %w", err)
    }
}
```

### Recommendation 5: Explicit Dependency Graph

**Add execution graph support:**

```yaml
# crew.yaml
execution_graph:
  # Define agent dependencies and parallelization
  phases:
    phase_1:
      agents: [research-lead]          # Run first
      parallel: false
      timeout: 120s
      depends_on: []
    
    phase_2:
      agents: [research-analyst, alternative-analyst]  # Can run in parallel
      parallel: true
      timeout: 120s
      depends_on: [phase_1]
      aggregation: merge_results       # How to combine outputs
    
    phase_3:
      agents: [qa-reviewer]            # Runs last
      parallel: false
      timeout: 60s
      depends_on: [phase_2]

# Alternative: DAG-based definition
dag:
  research-lead:
    outputs: [initial_research]
    outputs_to: [research-analyst]
  
  research-analyst:
    inputs_from: [research-lead]
    outputs: [detailed_analysis]
    outputs_to: [qa-reviewer]
  
  qa-reviewer:
    inputs_from: [research-analyst]
    outputs: [final_report]
```

**Implementation:**
```go
type ExecutionGraph struct {
    Phases []ExecutionPhase
    // OR
    DAG    map[string]*AgentNode
}

type ExecutionPhase struct {
    Name          string
    Agents        []string
    Parallel      bool
    Timeout       time.Duration
    DependsOn     []string
    Aggregation   string  // merge_results, select_best, concatenate
}

func (ce *CrewExecutor) ExecuteWithGraph(ctx context.Context, graph *ExecutionGraph) (*CrewResponse, error) {
    for _, phase := range graph.Phases {
        if phase.Parallel {
            results, err := ce.ExecuteParallel(ctx, input, agents)
        } else {
            results, err := ce.ExecuteSequential(ctx, input, agents)
        }
    }
}
```

### Recommendation 6: Enhanced Behavior Configuration

**Extend agent behavior options:**

```yaml
# crew.yaml
agent_behaviors:
  orchestrator:
    # ROUTING
    wait_for_signal: true
    auto_route: false
    is_terminal: false
    
    # EXECUTION
    retry_policy:
      max_attempts: 3
      backoff: exponential
      timeout_override: 30s
    
    # CONTEXT MANAGEMENT
    context_injection:
      prefix: "Previous analysis: {{history}}"
      max_history_depth: 5
      include_metadata: true
    
    # PRIORITY & PERFORMANCE
    priority: high                    # high, normal, low
    estimated_tokens: 5000           # For context planning
    cache_results: true
    
    # DEBUGGING
    verbose_logging: false
    trace_execution: true
    
    # SAFETY
    max_retries: 3
    timeout_seconds: 60
    fallback_agent: clarifier
    on_failure: escalate
```

---

## Part 4: Modernized Schema (Proposal)

### New Agent YAML Schema (v2.0)

```yaml
# examples/[service]/config/agents/[agent-id].yaml

# ============================================================================
# IDENTITY (REQUIRED)
# ============================================================================
id: orchestrator                                    # Unique identifier
name: System Orchestrator                          # Display name
role: Routing and decision making                  # Responsibility
type: router                                       # router|gatherer|executor|analyzer|reviewer|synthesizer

description: |                                     # Purpose statement
  Routes customer issues to appropriate specialists
  based on initial diagnosis

# ============================================================================
# PERSONALITY & CONTEXT (RECOMMENDED)
# ============================================================================
backstory: |
  You are an experienced IT system administrator
  with 15 years of experience managing complex systems.
  
  Your strengths:
  - Rapid problem diagnosis
  - Understanding of root causes
  - Ability to delegate to specialists

# ============================================================================
# LLM CONFIGURATION
# ============================================================================
model: gpt-4o-mini                                 # gpt-4o, gpt-4-turbo, gpt-4, gpt-3.5-turbo
temperature: 0.5                                   # 0-2 range (0=deterministic, 2=creative)

# ============================================================================
# TOOL CONFIGURATION
# ============================================================================
tools:
  - name: GetCPUUsage
    parameters:
      threshold: 80
      include_per_core: true
    behavior:
      max_calls: 5
      timeout: 10s
      cache: true
      retry_on_failure: true
    expected_output:
      format: json
      schema:
        usage: number
        cores: array

  - name: GetMemoryUsage
    parameters:
      units: gb
    behavior:
      max_calls: 3
      timeout: 10s

# ============================================================================
# ROUTING CONFIGURATION
# ============================================================================
workflow:
  is_terminal: false                              # Terminal agent?
  handoff_targets:                                # Can route to:
    - executor
    - clarifier
  
  # New: explicit signal definition
  allowed_signals:
    - pattern: "^\\[ROUTE_[A-Z_]+\\]$"
      examples: ["[ROUTE_EXECUTOR]", "[ROUTE_CLARIFIER]"]
      required: true                              # Agent MUST emit signal
  
  # New: error handling
  error_handling:
    on_timeout:
      action: retry
      max_attempts: 3
      backoff: exponential
    
    on_tool_failure:
      action: continue
      fallback_tool: null
    
    on_validation_error:
      action: escalate

# ============================================================================
# EXECUTION BEHAVIOR
# ============================================================================
behavior:
  priority: high                                  # high|normal|low
  timeout_seconds: 60                            # Agent-specific timeout
  
  # Context management
  context:
    max_history_depth: 10
    include_metadata: true
    injection_prefix: "Previous context: {{history}}"
  
  # Caching
  cache_results: true
  cache_ttl: 3600s
  
  # Debugging
  verbose_logging: false
  trace_execution: true

# ============================================================================
# CUSTOM SYSTEM PROMPT (OPTIONAL)
# ============================================================================
system_prompt: |
  你好，我是 {{name}}
  
  ## 角色
  {{role}}
  {{description}}
  
  ## 背景
  {{backstory}}
  
  ## 可用工具
  {{tools:formatted-as-bullet-list}}
  
  ## 路由指示
  我可以转交给:
  {{handoff_targets:formatted-as-list}}
  
  当我需要转交时，我会发出这些信号之一:
  {{allowed_signals:formatted-as-list}}
  
  ## 约束
  - 最多尝试 3 次
  - 如果超时，重试
  - 如果工具失败，继续
  
  ## 示例
  [Examples section]

# ============================================================================
# METADATA & DOCUMENTATION
# ============================================================================
metadata:
  version: "2.0"
  created_date: "2025-12-22"
  last_modified: "2025-12-22"
  author: "Taipm"
  status: production
  
  # For API documentation
  tags:
    - routing
    - orchestration
    - decision-making
  
  # For team
  owner: "IT-Support-Team"
  slack_channel: "#it-support"
  
  # Monitoring
  expected_tokens: 5000
  estimated_latency: 30s
  success_rate_target: 0.95
```

### New Crew YAML Schema (v2.0)

```yaml
# examples/[service]/config/crew.yaml

version: "2.0"                                    # Config schema version
name: IT-Support-Crew                            # Crew name
description: |                                  # Crew purpose
  Multi-agent system for IT problem diagnosis
  and resolution with expert routing

# ============================================================================
# CORE CONFIGURATION
# ============================================================================
entry_point: orchestrator                       # Starting agent
agents:                                         # Agents to load
  - orchestrator
  - clarifier
  - executor

# ============================================================================
# EXECUTION SETTINGS
# ============================================================================
settings:
  max_handoffs: 5                              # Max routing transitions
  max_rounds: 10                               # Max conversation rounds
  max_total_time: 300s                         # Total execution timeout
  language: zh                                 # Primary language
  organization: IT-Support-Team                # Organization

# ============================================================================
# SIGNAL SCHEMA (NEW)
# ============================================================================
signal_schema:
  patterns:
    routing:
      pattern: "^\\[ROUTE_[A-Z_]+\\]$"
      description: "Agent routing signals"
      examples: ["[ROUTE_EXECUTOR]", "[ROUTE_CLARIFIER]"]
    
    status:
      pattern: "^\\[[A-Z_]+\\]$"
      description: "Status/completion signals"
      examples: ["[COMPLETE]", "[ERROR]", "[RETRY]"]

# ============================================================================
# EXECUTION GRAPH (NEW)
# ============================================================================
execution_graph:
  phases:
    initial:
      agents: [orchestrator]
      parallel: false
      timeout: 60s
      depends_on: []
    
    analysis:
      agents: [clarifier]
      parallel: false
      timeout: 120s
      depends_on: [initial]
    
    execution:
      agents: [executor]
      parallel: false
      timeout: 120s
      depends_on: [analysis]

# ============================================================================
# ROUTING CONFIGURATION
# ============================================================================
routing:
  # Signal-based routing definitions
  signals:
    orchestrator:
      - signal: "[ROUTE_EXECUTOR]"
        target: executor
        description: "Route directly to executor"
      
      - signal: "[ROUTE_CLARIFIER]"
        target: clarifier
        description: "Route to clarifier for more info"
    
    clarifier:
      - signal: "[KẾT_THÚC]"
        target: executor
        description: "Clarification complete, proceed to diagnosis"
    
    executor:
      # Terminal agent - no further routing
  
  # Default routing when no explicit signal
  defaults:
    orchestrator: clarifier                    # Safe default
    clarifier: executor
    executor: null                             # Terminal
  
  # Per-agent behavior overrides
  agent_behaviors:
    orchestrator:
      wait_for_signal: true                   # Must emit routing signal
      auto_route: false                       # Don't auto-route
      is_terminal: false
      retry_policy:
        max_attempts: 3
        backoff: exponential
      timeout_override: 60s
    
    clarifier:
      wait_for_signal: true
      auto_route: false
      is_terminal: false
      timeout_override: 120s
    
    executor:
      is_terminal: true
      timeout_override: 120s
      retry_policy:
        max_attempts: 5                       # More aggressive retries
        backoff: exponential

# ============================================================================
# GLOBAL ERROR POLICIES
# ============================================================================
error_policies:
  default:
    on_timeout:
      action: retry                           # retry|skip|escalate
      max_attempts: 3
      backoff_strategy: exponential
      initial_backoff: 1s
      max_backoff: 30s
    
    on_tool_failure:
      action: continue                        # continue|skip|escalate
      log_level: warning
    
    on_validation_error:
      action: escalate
    
    on_max_handoffs:
      action: terminate

# ============================================================================
# PARALLEL GROUPS (IF NEEDED)
# ============================================================================
parallel_groups:
  analysis_group:
    agents: [clarifier, alternative-analyzer]
    wait_for_all: true
    timeout: 120s
    aggregation: merge_results                # merge_results|select_best|concatenate
    next_agent: executor

# ============================================================================
# CONTEXT INJECTION (NEW)
# ============================================================================
context_injection:
  global:
    system_info: "Customer environment details"
    incident_history: "Relevant previous incidents"
  
  agent_specific:
    orchestrator:
      prefix: "You are routing this incident..."
    clarifier:
      prefix: "You are gathering more information..."
    executor:
      prefix: "You are now ready to diagnose..."

# ============================================================================
# METADATA & MONITORING
# ============================================================================
metadata:
  version: "2.0"
  created_date: "2025-12-22"
  last_modified: "2025-12-22"
  author: "Taipm"
  status: production
  
  tags:
    - it-support
    - diagnostics
    - multi-agent
  
  # Monitoring targets
  monitoring:
    success_rate_target: 0.95
    average_latency_target: 45s
    error_rate_threshold: 0.05
  
  # Team
  owner_team: "IT-Support-Team"
  slack_channel: "#it-support"
  pagerduty_team: "it-support-oncall"
```

---

## Part 5: Migration Path

### Phase 1: Backward Compatibility (v1.0 → v2.0)

**Maintain full support for v1.0 configs:**
- Auto-detect schema version
- Apply compatibility layer for missing fields
- Warn on deprecated patterns
- Provide migration guidance

```go
func LoadAgentConfig(path string) (*Agent, error) {
    data := unmarshalYAML(path)
    
    // Detect version
    version := detectVersion(data)
    
    switch version {
    case "1.0":
        return migrateFromV1(data)  // Apply compatibility layer
    case "2.0":
        return parseV2(data)         // New schema
    default:
        return nil, fmt.Errorf("unknown schema version")
    }
}

func migrateFromV1(data map[string]interface{}) (*Agent, error) {
    // Apply defaults for missing v2.0 fields
    // Log warnings for deprecated patterns
    // Convert old patterns to new patterns
}
```

### Phase 2: Documentation & Examples

1. Create migration guide: `docs/YAML-MIGRATION-v1-to-v2.md`
2. Create standardization guide: `docs/YAML-STANDARDS.md`
3. Update IT Support example to v2.0
4. Create template generator tool

### Phase 3: Tooling

1. YAML validator with detailed error messages
2. Config generator from templates
3. Schema documentation generator
4. Migration script for v1.0 → v2.0

---

## Part 6: Recommendations Priority

### CRITICAL (Implement First)

1. ✅ **Signal Schema Validation** - Prevents runtime signal mismatches
2. ✅ **Tool Parameter Specification** - Enable per-agent tool customization
3. ✅ **Error Handling Policies** - Graceful failure modes

### HIGH (Implement Next)

4. ⭐ **Template Variables Extension** - Auto-populate system prompts
5. ⭐ **Behavior Configuration** - Per-agent customization
6. ⭐ **Execution Graph** - Explicit dependency management

### MEDIUM (Nice to Have)

7. 📋 **Dependency Declaration** - Support complex workflows
8. 📋 **Context Injection** - Global context sharing
9. 📋 **Metadata & Monitoring** - Observability integration

### LOW (Future)

10. 🔮 **Parallel Group Enhancements** - Advanced parallelization
11. 🔮 **Config Validation Tooling** - DevOps automation
12. 🔮 **Schema Documentation Generator** - API docs

---

## Summary

The current YAML structure is production-ready but would benefit from modernization in 6 key areas. The proposed v2.0 schema maintains backward compatibility while adding powerful new features for:

- **Explicit signal validation** - No more guessing about signals
- **Tool customization** - Per-agent tool parameter specification
- **Error resilience** - Structured error handling policies
- **Template flexibility** - Dynamic system prompt generation
- **Execution clarity** - Explicit dependency graphs
- **Observable behavior** - Enhanced debugging and monitoring

**Recommended next step:** Create comprehensive standardization document + migration guide before implementing message history limit fix.


# Phase 5 Priorities - Detailed Implementation Plan

**Document Date:** 2025-12-25
**Phase:** 5 (Post-Architecture Refactoring)
**Status:** Planning

---

## Executive Summary

Phase 4 completed a successful architecture refactoring, splitting the `executor/` module into three specialized modules:
- `routing/` - Agent routing decision logic
- `state-management/` - Execution state and history tracking
- `execution/` - Workflow execution orchestration

Phase 5 focuses on **completing the routing implementation**, improving **cost tracking**, and enhancing **developer experience** with better error handling and logging.

---

## 1. CRITICAL BLOCKERS - Routing Implementation

### 1.1 ❌ CRITICAL: Agent Handoff Not Executing

**Priority:** 🔴 CRITICAL
**File:** `core/workflow/execution.go` (Lines 193, 218)
**Status:** TODO - Incomplete

**Current Implementation:**
```go
if routingDecision != nil && routingDecision.NextAgentID != "" {
    execCtx.emitSignal(signal.SignalHandoff, ...)

    // TODO: Look up next agent by ID and continue execution
    // For now, return (will be implemented in crew.go integration)
    return response, nil  // ❌ RETURNS INSTEAD OF CONTINUING
}
```

**Problem:**
- Routing decision is made but execution stops
- Next agent is never looked up or executed
- **Blocks ALL multi-agent workflows**

**Impact:** HIGH
- Cannot execute workflows with multiple agents
- Signal routing is non-functional in practice

**Solution Design:**

```go
// Phase 5 Implementation Plan:
if routingDecision != nil && routingDecision.NextAgentID != "" {
    execCtx.emitSignal(signal.SignalHandoff, map[string]interface{}{
        "from_agent": currentAgent.ID,
        "to_agent": routingDecision.NextAgentID,
        "reason": routingDecision.Reason,
    })

    // NEW: Look up next agent
    nextAgent, exists := agentsMap[routingDecision.NextAgentID]
    if !exists {
        return nil, fmt.Errorf("target agent not found: %s", routingDecision.NextAgentID)
    }

    // NEW: Update execution context
    execCtx.CurrentAgent = nextAgent
    execCtx.HandoffCount++

    // NEW: Recursively execute next agent
    return executeAgent(ctx, execCtx, userInput, apiKey)
}
```

**Testing Plan:**
- Unit test: Single handoff (Agent A → Agent B)
- Unit test: Multi-step handoff (A → B → C → terminal)
- Integration test: Signal-based routing with handoff
- Integration test: Terminal agent detection stops execution

**Effort:** Medium (3-4 hours)
**Dependencies:** None

---

### 1.2 ❌ HIGH: Behavior-Based Routing is a Stub

**Priority:** 🟠 HIGH
**File:** `core/routing/signal.go` (Lines 36-51)
**Status:** Not Implemented

**Current Code:**
```go
func RouteByBehavior(behavior string, routing *common.RoutingConfig) (string, error) {
    // Placeholder implementation for behavior-based routing
    // In a full implementation, would lookup behavior in routing.AgentBehaviors map
    // and return the target agent ID

    return "", fmt.Errorf("behavior '%s' not found in routing configuration", behavior)
}
```

**Problem:**
- Function always returns error
- `routing.AgentBehaviors` map defined but never used
- No behavior-based routing capability

**Config Already Defined:**
```go
type AgentBehavior struct {
    Name        string   `yaml:"name"`
    Description string   `yaml:"description"`
    Triggers    []string `yaml:"triggers"`      // When to trigger this behavior
    TargetAgent string   `yaml:"target_agent"`  // Where to route to
}

// In RoutingConfig:
AgentBehaviors map[string]*AgentBehavior `yaml:"agent_behaviors"`
```

**Solution:**

```go
func RouteByBehavior(behavior string, routing *common.RoutingConfig) (string, error) {
    if routing == nil {
        return "", fmt.Errorf("routing configuration is nil")
    }

    if behavior == "" {
        return "", fmt.Errorf("behavior name is empty")
    }

    if routing.AgentBehaviors == nil {
        return "", fmt.Errorf("no behaviors configured in routing")
    }

    // Look up behavior by name
    for behaviorName, behaviorConfig := range routing.AgentBehaviors {
        if behaviorName == behavior {
            if behaviorConfig == nil {
                return "", fmt.Errorf("behavior '%s' configuration is nil", behavior)
            }
            if behaviorConfig.TargetAgent == "" {
                return "", fmt.Errorf("behavior '%s' has no target agent configured", behavior)
            }
            return behaviorConfig.TargetAgent, nil
        }
    }

    return "", fmt.Errorf("behavior '%s' not found in routing configuration", behavior)
}
```

**Testing Plan:**
- Unit test: Valid behavior lookup
- Unit test: Non-existent behavior
- Unit test: Behavior with empty target
- Integration test: Behavior routing in workflow

**Effort:** Low (1-2 hours)
**Dependencies:** None

---

### 1.3 ❌ HIGH: Parallel Groups Defined But Not Executed

**Priority:** 🟠 HIGH
**Files:**
- `core/common/types.go` (Lines 134-141) - Type definition
- Need to create: `core/routing/parallel.go` - Implementation

**Status:** Not Implemented

**Current Type Definition:**
```go
type ParallelGroupConfig struct {
    Agents         []string `yaml:"agents"`           // ✓ Defined
    WaitForAll     bool     `yaml:"wait_for_all"`     // ✓ Defined
    TimeoutSeconds int      `yaml:"timeout_seconds"`  // ✓ Defined
    NextAgent      string   `yaml:"next_agent"`       // ✓ Defined
    Description    string   `yaml:"description"`      // ✓ Defined
}
```

**Problem:**
- Type is parsed from YAML but **zero execution logic exists**
- Parallel groups referenced in routing config but never executed
- No concurrent agent execution capability

**Solution Design:**

```go
// New file: core/routing/parallel.go

type ParallelExecutionResult struct {
    AgentID      string
    Response     *common.AgentResponse
    Error        error
    Duration     time.Duration
}

type ParallelGroupExecutor struct {
    config   *common.ParallelGroupConfig
    agents   map[string]*common.Agent
    timeout  time.Duration
}

func (pge *ParallelGroupExecutor) Execute(
    ctx context.Context,
    input string,
    history []common.Message,
    apiKey string,
) ([]ParallelExecutionResult, error) {

    // Create channels for results
    resultChan := make(chan ParallelExecutionResult, len(pge.config.Agents))

    // Apply timeout
    ctx, cancel := context.WithTimeout(ctx, pge.timeout)
    defer cancel()

    // Launch concurrent agent execution
    for _, agentID := range pge.config.Agents {
        go pge.executeAgentConcurrently(ctx, agentID, input, history, apiKey, resultChan)
    }

    // Collect results
    results := []ParallelExecutionResult{}

    if pge.config.WaitForAll {
        // Wait for all agents to complete
        for range pge.config.Agents {
            result := <-resultChan
            results = append(results, result)
        }
    } else {
        // Wait for first successful result
        for range pge.config.Agents {
            select {
            case result := <-resultChan:
                if result.Error == nil {
                    return []ParallelExecutionResult{result}, nil
                }
                results = append(results, result)
            case <-ctx.Done():
                return results, ctx.Err()
            }
        }
    }

    return results, nil
}
```

**Testing Plan:**
- Unit test: All agents execute successfully
- Unit test: Some agents fail, WaitForAll=true
- Unit test: First-success routing (WaitForAll=false)
- Unit test: Timeout handling
- Integration test: Parallel group with 3 agents

**Effort:** Medium-High (4-6 hours)
**Dependencies:** Agent execution infrastructure (already exists)

---

## 2. HIGH PRIORITY - Cost Tracking & Metrics

### 2.1 🟠 HIGH: Cost Calculation Uses Hardcoded Pricing

**Priority:** 🟠 HIGH
**File:** `core/common/types.go` (Lines 663-677)
**Impact:** Cost tracking completely inaccurate

**Current Implementation:**
```go
func (a *Agent) CalculateCost(tokenCount int) float64 {
    // Default pricing approximation (varies by model)
    // GPT-4: ~$30/1M input tokens, ~$60/1M output tokens
    // Assume 60% input, 40% output = average $42/1M
    averagePricePerMillion := 0.042  // ❌ HARDCODED!

    costPerToken := averagePricePerMillion / 1_000_000.0
    return float64(tokenCount) * costPerToken
}
```

**Problems:**
1. Uses fixed $0.042/token for ALL models
2. Ignores configured pricing in `CostLimitsConfig`
3. Doesn't differentiate input vs output tokens
4. No provider-specific pricing (OpenAI vs Ollama)

**Config Defined But Ignored:**
```go
type CostLimitsConfig struct {
    MaxTokensPerCall           int     // Used
    MaxTokensPerDay            int     // Used
    MaxCostPerDayUSD           float64 // Used
    AlertThreshold             float64 // Used
    BlockOnCostExceed          bool    // Used
    InputTokenPricePerMillion  float64 // ❌ IGNORED
    OutputTokenPricePerMillion float64 // ❌ IGNORED
}
```

**Solution:**

```go
// Modified Agent.CalculateCost to use configured pricing
func (a *Agent) CalculateCost(inputTokens, outputTokens int) float64 {
    if a == nil || a.Quota == nil {
        return 0
    }

    // Get configured pricing (falls back to defaults)
    inputPrice := a.Quota.InputTokenPricePerMillion
    if inputPrice == 0 {
        inputPrice = 0.03 // Default GPT-3.5 input pricing
    }

    outputPrice := a.Quota.OutputTokenPricePerMillion
    if outputPrice == 0 {
        outputPrice = 0.06 // Default GPT-3.5 output pricing
    }

    // Calculate cost with input/output differentiation
    inputCost := float64(inputTokens) * (inputPrice / 1_000_000.0)
    outputCost := float64(outputTokens) * (outputPrice / 1_000_000.0)

    return inputCost + outputCost
}

// Update signature throughout codebase:
// OLD: CalculateCost(tokenCount int)
// NEW: CalculateCost(inputTokens, outputTokens int)
```

**Provider-Specific Pricing:**
```go
// Add pricing defaults for each provider in providers/provider.go
const (
    // OpenAI GPT-4
    OpenAI_GPT4_InputPrice  = 0.03   // $30/1M tokens
    OpenAI_GPT4_OutputPrice = 0.06   // $60/1M tokens

    // OpenAI GPT-3.5
    OpenAI_GPT35_InputPrice  = 0.0015  // $1.5/1M tokens
    OpenAI_GPT35_OutputPrice = 0.002   // $2/1M tokens

    // Ollama (local, free)
    Ollama_Price = 0.0
)
```

**Testing Plan:**
- Unit test: Input vs output cost calculation
- Unit test: Uses configured pricing
- Unit test: Falls back to defaults
- Unit test: Provider-specific pricing
- Integration test: Cost calculation in workflow

**Effort:** Low (2-3 hours)
**Dependencies:** Need to track input vs output tokens (currently just total)

---

### 2.2 🟠 HIGH: Per-Tool Cost Tracking Missing

**Priority:** 🟠 HIGH
**Files:** `core/tools/executor.go`, `core/metrics.go`
**Status:** Not Implemented

**Current:** Only agent-level metrics
**Need:** Tool-level cost breakdown

**Solution:**

```go
// New struct in metrics.go
type ToolMetrics struct {
    ToolName          string
    ExecutionCount    int64
    SuccessfulCount   int64
    FailedCount       int64
    TotalDuration     time.Duration
    AverageDuration   time.Duration
    TotalCost         float64        // NEW
    MaxDuration       time.Duration
    MinDuration       time.Duration
    AvgTokensPerCall  float64        // NEW
}

// Update AgentMetrics
type AgentMetrics struct {
    // ... existing fields
    ToolMetrics map[string]*ToolMetrics  // NEW: Per-tool breakdown
    TotalToolCost float64                // NEW: Sum of all tool costs
}

// Update ExecuteTool to record metrics
func ExecuteTool(ctx context.Context, name string, tool *Tool, args map[string]interface{}, agentID string) (string, error) {
    startTime := time.Now()

    // ... existing execution logic

    // Record metrics
    metricsCollector.RecordToolExecution(
        toolName: name,
        agentID: agentID,
        duration: time.Since(startTime),
        success: err == nil,
        cost: estimateToolCost(name, result),  // NEW
    )

    return result, err
}
```

**Effort:** Medium (3-4 hours)
**Dependencies:** ExecuteTool refactoring

---

## 3. MEDIUM PRIORITY - Developer Experience

### 3.1 🟡 MEDIUM: Custom Error Types Underutilized

**Priority:** 🟡 MEDIUM
**Files:** `core/common/errors.go`, `core/workflow/execution.go`, `core/routing/*.go`
**Impact:** Poor error debugging in production

**Defined Error Types (Not Being Used):**
```go
- ErrorType enum with 7 types (Timeout, Panic, Validation, Network, etc.)
- ValidationError with Field + Message
- ExecutionError with AgentID + TaskID + ErrorType
- TimeoutError with Operation + Duration
- QuotaExceededError with QuotaType + Limits
- ConfigurationError with Field + Message
- ProviderError with ProviderName + Message
- ToolExecutionError with ToolName + Message
```

**Problem - Current Usage:**
```go
// ❌ Generic errors everywhere
if execCtx.RoundCount >= execCtx.MaxRounds {
    return nil, fmt.Errorf("max rounds (%d) exceeded", execCtx.MaxRounds)
}

if currentAgent == nil {
    return nil, fmt.Errorf("current agent cannot be nil")
}

if len(currentAgent.HandoffTargets) == 0 {
    return nil, fmt.Errorf("no handoff targets configured")
}
```

**Solution:**

```go
// ✅ Use typed errors
if execCtx.RoundCount >= execCtx.MaxRounds {
    return nil, &common.ExecutionError{
        AgentID: currentAgent.ID,
        TaskID: "workflow_execution",
        ErrorType: common.ErrorTypeTimeout,
        Err: fmt.Errorf("max rounds (%d) exceeded", execCtx.MaxRounds),
    }
}

if currentAgent == nil {
    return nil, &common.ExecutionError{
        ErrorType: common.ErrorTypeValidation,
        Err: fmt.Errorf("current agent cannot be nil"),
    }
}

if len(currentAgent.HandoffTargets) == 0 {
    return nil, &common.ConfigurationError{
        Field: fmt.Sprintf("agents[%s].handoff_targets", currentAgent.ID),
        Err: fmt.Errorf("no handoff targets configured"),
    }
}
```

**Audit & Migration Plan:**
1. List all `fmt.Errorf` calls in workflow, routing, execution packages
2. For each error, determine appropriate typed error
3. Replace with typed error
4. Update error handling code to use `.Type()` or `.Is()`
5. Add tests for error types

**Effort:** Medium (4-5 hours - systematic replacement)
**Dependencies:** None

---

### 3.2 🟡 MEDIUM: Unstructured Logging

**Priority:** 🟡 MEDIUM
**Files:** `core/http.go`, `core/workflow/execution.go`, `core/executor/executor.go`
**Status:** printf-based logging only

**Current Problems:**
```go
// ❌ Unstructured
log.Printf("[INPUT ERROR] Invalid query: %v", err)
log.Printf("[EXECUTOR] Execution completed: %s", response.Content)
log.Printf("[WORKFLOW] Tool execution had errors: %v", toolErr)

// Issues:
// 1. No structured fields
// 2. No correlation IDs
// 3. No log levels (all mixed)
// 4. Hard to parse
// 5. No context preservation
```

**Solution:**

```go
// Use structured logger (Go 1.21+ slog or zerolog)
import "log/slog"

// In HTTP handler
requestID := generateRequestID()  // NEW: UUID
logger := slog.With("request_id", requestID)

logger.Info("workflow_started",
    "agent_id", agent.ID,
    "input_length", len(input),
    "history_length", len(history),
)

// In workflow execution
logger.Info("agent_execution",
    "request_id", requestID,
    "agent_id", currentAgent.ID,
    "round", execCtx.RoundCount,
)

if err != nil {
    logger.Error("agent_execution_failed",
        "request_id", requestID,
        "agent_id", currentAgent.ID,
        "error_type", fmt.Sprintf("%T", err),
        "error", err.Error(),
    )
}

// Tool execution
logger.Info("tool_execution",
    "request_id", requestID,
    "agent_id", currentAgent.ID,
    "tool_name", toolName,
    "duration_ms", duration.Milliseconds(),
)
```

**Implementation Plan:**
1. Add slog initialization in main/http
2. Create logger context in HTTP handler
3. Pass request ID through execution context
4. Replace all log.Printf with logger.Info/Error/Warn
5. Configure JSON output for production

**Effort:** Medium (5-6 hours)
**Dependencies:** Go 1.21+ (or add zerolog dependency)

---

### 3.3 🟡 MEDIUM: Validation Error Messages Lack Context

**Priority:** 🟡 MEDIUM
**Files:** `core/validation/*.go`, `core/config/loader.go`
**Impact:** Users struggle to fix configuration errors

**Current - Poor Context:**
```go
// ❌ Where is this error from?
return &common.ValidationError{
    Field:   "signals",
    Message: "invalid signal format",
}
// User: "Which signals? In which file? Which agent?"
```

**Better - With File Context:**
```go
// ✅ Full context
return &common.ValidationError{
    Field:   "agents[student].signals[0]",
    Message: "signal 'SUBMIT_EXAM' must be in [SIGNAL_NAME] format (uppercase with underscores)",
    FilePath: "config/agents/student.yaml",
    LineNumber: 42,
}
```

**Solution:**

```go
// Update ValidationError struct
type ValidationError struct {
    Field       string      // Field path: agents[id].signals[0]
    Message     string      // What's wrong
    FilePath    string      // crew.yaml, agents/agent1.yaml, etc.
    LineNumber  int         // Line where error occurs
    Suggestion  string      // How to fix it
    Err         error       // Wrapped error
}

func (e *ValidationError) Error() string {
    if e.FilePath != "" {
        return fmt.Sprintf("%s:%d - %s: %s (suggestion: %s)",
            e.FilePath, e.LineNumber, e.Field, e.Message, e.Suggestion)
    }
    return fmt.Sprintf("%s: %s", e.Field, e.Message)
}

// Example usage
return &common.ValidationError{
    Field:      "agents[teacher].signals",
    Message:    "signal '[UNKNOWN]' not recognized",
    FilePath:   "config/agents/teacher.yaml",
    LineNumber: 28,
    Suggestion: "Valid signals: [NEXT_AGENT], [DELEGATE], [ESCALATE]",
}
```

**Validation audit points:**
- Agent configuration loading
- Routing signal validation
- Tool parameter validation
- Quota configuration validation

**Effort:** Low-Medium (3-4 hours)
**Dependencies:** None

---

## 4. DISTRIBUTED EXECUTION GROUNDWORK

### 4.1 🟡 MEDIUM: Context Propagation for Distribution

**Priority:** 🟡 MEDIUM
**File:** `core/workflow/execution.go`
**Status:** Partially implemented

**Current Issue:**
```go
// History passed by reference, modifications affect all branches
execCtx.History = append(execCtx.History, responseMsg)
// In distributed setting: which node owns this history slice?
```

**Solution for Phase 5:**

```go
// Deep copy history at critical points
type ExecutionContext struct {
    // ... existing fields
    History []common.Message  // Will be deep-copied at handoffs
}

func (ec *ExecutionContext) CopyForHandoff() *ExecutionContext {
    newCtx := &ExecutionContext{
        CurrentAgent: ec.CurrentAgent,
        History:      make([]common.Message, len(ec.History)),  // NEW: Deep copy
        // ... other fields
    }

    // Deep copy history
    for i, msg := range ec.History {
        newCtx.History[i] = common.Message{
            Role:    msg.Role,
            Content: msg.Content,
        }
    }

    return newCtx
}

// Use in handoff
nextCtx := currentCtx.CopyForHandoff()
nextCtx.CurrentAgent = nextAgent
response, err := executeAgent(ctx, nextCtx, input, apiKey)
```

**Effort:** Low (2-3 hours)
**Dependencies:** None (can do now)

---

## Implementation Roadmap

### Sprint 1 (Week 1) - Critical Routing Fixes
- [ ] Implement agent handoff execution (Lines 193, 218 in execution.go)
- [ ] Implement behavior-based routing
- [ ] Tests for both features
- **Effort:** 8 hours
- **Owner:** Dev Team

### Sprint 2 (Week 2) - Cost Tracking
- [ ] Fix cost calculation with configured pricing
- [ ] Separate input/output token tracking
- [ ] Add per-tool cost metrics
- **Effort:** 10 hours

### Sprint 3 (Week 3) - DX Improvements
- [ ] Replace generic errors with typed errors
- [ ] Add structured logging
- [ ] Improve validation error messages
- **Effort:** 12 hours

### Sprint 4+ (Future) - Parallel & Distributed
- [ ] Parallel group execution
- [ ] Distributed coordination groundwork
- [ ] Remote agent invocation

---

## Success Criteria

- ✅ Multi-agent workflows execute end-to-end
- ✅ Behavior-based routing works
- ✅ Parallel groups execute concurrently
- ✅ Cost tracking is accurate per agent/tool
- ✅ Error messages include file path and line number
- ✅ Structured logs with correlation IDs
- ✅ 95%+ test coverage for routing module

---

## Related Documents

- [Phase 4 Module Splitting Report](./PHASE_4_COMPLETION_REPORT.md)
- [Architecture Overview](./ARCHITECTURE.md)
- [Routing Module Documentation](../core/routing/README.md) *(to be created)*

---

**Last Updated:** 2025-12-25
**Next Review:** 2025-12-29

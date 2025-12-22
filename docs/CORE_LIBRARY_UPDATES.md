# Core Library Updates & New Features

**Version**: 1.0+
**Last Updated**: 2025-12-22
**Status**: Production Ready

---

## Tổng Quan (Overview)

Core library go-agentic đã được nâng cấp với nhiều tính năng mới và cải tiến. Tài liệu này mô tả các thay đổi chính, tính năng mới, và cách sử dụng chúng.

---

## 1. Model Fallback System (Primary & Backup Models)

### ✅ Fix Issue #1: Hardcoded Model Bug

**Vấn đề Cũ**:
- Model được hardcode thay vì đọc từ config
- Không có xác nhập provider

**Giải Pháp**:
- Primary & Backup model configuration
- Automatic fallback khi primary model fails
- Explicit provider validation

### Cấu Trúc Mới trong agent.yaml

```yaml
id: my-agent
name: My Agent
role: Specialist
description: Agent description

backstory: My backstory
temperature: 0.7

# Primary model (required)
model: gpt-4-turbo
provider: openai

# Optional: Backup model for failover
primary:
  model: gpt-4-turbo
  provider: openai
  provider_url: https://api.openai.com

backup:
  model: gpt-4o-mini
  provider: openai
  provider_url: https://api.openai.com

# OR for Ollama
model: deepseek-r1:1.5b
provider: ollama
provider_url: http://localhost:11434

# Backup Ollama
backup:
  model: gemma3:1b
  provider: ollama
  provider_url: http://localhost:11434
```

### API Changes

**Old Structure** (Still Supported - Backward Compatible):
```go
agent.Model       = "gpt-4-turbo"
agent.Provider    = "openai"
agent.ProviderURL = "https://api.openai.com"
```

**New Structure** (Recommended):
```go
agent.Primary = &ModelConfig{
    Model:       "gpt-4-turbo",
    Provider:    "openai",
    ProviderURL: "https://api.openai.com",
}

agent.Backup = &ModelConfig{
    Model:       "gpt-4o-mini",
    Provider:    "openai",
    ProviderURL: "https://api.openai.com",
}
```

### Behavior

```go
// ExecuteAgent automatically handles fallback
response, err := ExecuteAgent(ctx, agent, input, history, apiKey)

// Logic:
// 1. Try primary model
// 2. If primary fails AND backup exists, try backup
// 3. If backup succeeds, return response
// 4. If both fail, return detailed error
```

### Benefits

✅ High availability - automatic failover
✅ Cost optimization - use cheaper backup model
✅ Provider flexibility - mix OpenAI + Ollama
✅ Backward compatible - old format still works

---

## 2. Configuration Validation System

### ✅ Fix Issue #6: YAML Validation

**New Component**: `validation.go`

**Validation Pipeline**:

```
Stage 1: Structure Validation
├─ entry_point exists?
├─ agents list not empty?
└─ version specified?

Stage 2: Field Validation
├─ temperature range (0.0-1.0)?
├─ required fields present?
└─ data types correct?

Stage 3: Agent Validation
├─ All referenced agents exist?
├─ Agent IDs unique?
└─ Providers valid?

Stage 4: Routing Validation
├─ Signal targets exist?
├─ Defaults valid?
└─ No circular routes?

Stage 5: Graph Validation
├─ Reachability check
├─ Cycle detection
└─ Terminal node analysis
```

### Usage

```go
// Load and validate configuration
executor, err := NewCrewExecutorFromConfig(apiKey, "config", nil)
if err != nil {
    // Error includes detailed validation report
    fmt.Fprintf(os.Stderr, "Configuration error:\n%s\n", err)
}

// Manual validation
validator := NewConfigValidator(crewConfig, agentConfigs)
err := validator.ValidateAll()
if err != nil {
    report := validator.GenerateErrorReport()
    fmt.Println(report)
}
```

### Validation Errors

Setiap error mencakup:
```
File:     crew.yaml
Section:  entry_point
Field:    entry_point
Message:  entry_point must reference an agent in the agents list
Severity: error
Fix:      entry_point: orchestrator (where orchestrator is in agents list)
Line:     4
```

---

## 3. Graceful Shutdown Management

### ✅ Fix Issue #18: Graceful Shutdown

**New Component**: `shutdown.go`

**GracefulShutdownManager**:
- Graceful shutdown of HTTP servers
- Tracks active requests/streams
- Ensures all work completes before stopping
- Prevents data loss during restart

### Usage

```go
// Create shutdown manager
shutdownMgr := NewGracefulShutdownManager()
shutdownMgr.GracefulTimeout = 30 * time.Second

// Start monitoring signals (SIGTERM, SIGINT)
go shutdownMgr.Start()

// Setup HTTP server
http.HandleFunc("/execute", handleExecute)
server := &http.Server{Addr: ":8080"}

// Register server with shutdown manager
shutdownMgr.RegisterServer(server)

// Start server
if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
    log.Fatal(err)
}

// Cleanup
shutdownMgr.Shutdown()
```

### Lifecycle

```
Running
  ↓ [SIGTERM/SIGINT received]
Shutting Down
  ├─ Stop accepting new requests
  ├─ Wait for active requests (timeout: 30s)
  ├─ Cancel active streams
  └─ Close server
Stopped
```

---

## 4. Request ID Tracking System

### ✅ Fix Issue #17: Request ID Tracking

**New Component**: `request_tracking.go`

**Purpose**: Correlation of logs/events across all components

### API

```go
// Generate request ID
requestID := GenerateRequestID()           // Full UUID
shortID := GenerateShortRequestID()        // "req-" + 12 chars

// Store in context
ctx = context.WithValue(ctx, RequestIDKey, requestID)

// Retrieve from context
requestID := GetRequestID(ctx)             // Returns "unknown" if not found

// Get or create
requestID, ctx := GetOrCreateRequestID(ctx)

// Use in logs
log.Printf("[%s] Processing request...", GetRequestID(ctx))
```

### Request Lifecycle Tracking

```go
// Track events during execution
eventTracker := NewRequestEventTracker(requestID)

eventTracker.TrackEvent(Event{
    Type:      "agent_thinking",
    Agent:     "orchestrator",
    Timestamp: time.Now(),
    Details: map[string]interface{}{
        "input": userInput,
    },
})

eventTracker.TrackEvent(Event{
    Type:      "tool_call",
    Agent:     "executor",
    Tool:      "GetCPUUsage",
    Timestamp: time.Now(),
})

// Get full timeline
timeline := eventTracker.GetTimeline()
```

---

## 5. Metrics & Observability System

### ✅ Fix Issue #14: Metrics & Observability

**New Component**: `metrics.go`

**MetricsCollector** tracks:

```
SystemMetrics
├─ Total requests / successful / failed
├─ Total execution time
├─ Average request time
├─ Memory usage (current & peak)
├─ Cache hit rate
└─ Per-agent breakdown
    ├─ AgentMetrics
    │  ├─ Execution count
    │  ├─ Success/Error/Timeout count
    │  ├─ Duration statistics
    │  └─ Per-tool breakdown
    │      ├─ ToolMetrics
    │      │  ├─ Execution count
    │      │  ├─ Duration stats
    │      │  └─ Error tracking
```

### Usage

```go
// Create and enable metrics collector
metricsCollector := NewMetricsCollector()
metricsCollector.Enable()

// Record execution
metricsCollector.RecordAgentExecution("orchestrator", true, 150*time.Millisecond)
metricsCollector.RecordToolExecution("orchestrator", "GetCPUUsage", true, 50*time.Millisecond)

// Get metrics
metrics := metricsCollector.GetSystemMetrics()

// Export as JSON
jsonMetrics, _ := json.Marshal(metrics)
fmt.Println(string(jsonMetrics))
```

### Metrics Structure

```json
{
  "start_time": "2025-12-22T10:00:00Z",
  "total_requests": 100,
  "successful_requests": 98,
  "failed_requests": 2,
  "average_request_time": 250000000,
  "cache_hit_rate": 0.92,
  "agent_metrics": {
    "orchestrator": {
      "agent_id": "orchestrator",
      "execution_count": 100,
      "success_count": 98,
      "error_count": 2,
      "total_duration": 25000000000,
      "average_duration": 250000000,
      "tool_metrics": {
        "GetCPUUsage": {
          "execution_count": 50,
          "success_count": 50,
          "average_duration": 50000000
        }
      }
    }
  }
}
```

---

## 6. Tool Validation System

### ✅ Fix Issue #25: Tool Execution Validation

**New Component**: `crew.go` (validateToolArguments)

**Validation Levels**:

```
1. Required fields check
   └─ All required parameters present?

2. Type validation
   └─ Parameter types match schema?

3. Argument validation
   └─ Custom tool validation (if defined)
```

### Tool Definition

```go
tool := &Tool{
    Name:        "GetCPUUsage",
    Description: "Get CPU usage percentage",
    Parameters: map[string]interface{}{
        "type": "object",
        "properties": map[string]interface{}{
            "timeout": map[string]interface{}{
                "type": "number",
            },
        },
        "required": []interface{}{"timeout"},
    },
    Handler: func(ctx context.Context, args map[string]interface{}) (string, error) {
        timeout := args["timeout"].(float64)
        // Execute tool
        return "50.2%", nil
    },
}
```

### Automatic Validation

```go
// Validation happens automatically before tool execution
response, err := ExecuteAgent(ctx, agent, input, history, apiKey)

// If tool arguments are invalid, tool won't execute
// Error: tool 'GetCPUUsage': required parameter 'timeout' is missing
```

---

## 7. Enhanced Types & Structures

### Agent Struct Enhancements

```go
type Agent struct {
    ID              string          // Unique agent ID
    Name            string          // Display name
    Role            string          // Agent role
    Backstory       string          // Agent background

    // ✅ New: Primary & Backup models
    Primary         *ModelConfig    // Primary LLM config
    Backup          *ModelConfig    // Fallback LLM config

    // ⚠️  Deprecated (use Primary instead)
    Model           string
    Provider        string
    ProviderURL     string

    SystemPrompt    string          // Custom system prompt
    Tools           []*Tool         // Available tools
    Temperature     float64         // Creativity level (0.0-1.0)
    IsTerminal      bool            // Execution endpoint?
    HandoffTargets  []string        // Can route to these agents
}
```

### Crew Struct Enhancements

```go
type Crew struct {
    Agents                  []*Agent
    Tasks                   []*Task
    MaxRounds               int
    MaxHandoffs             int

    // ✅ New: Configurable timeouts (Issue #4)
    ParallelAgentTimeout    time.Duration   // Default: 60s

    // ✅ New: Tool output limit (Issue #5)
    MaxToolOutputChars      int             // Default: 2000

    // ✅ New: Routing configuration
    Routing                 *RoutingConfig
}
```

### RoutingConfig Enhancements

```go
type RoutingConfig struct {
    // Signal-based routing
    Signals        map[string][]RoutingSignal

    // Default routing when no signal
    Defaults       map[string]string

    // Agent-specific behaviors
    AgentBehaviors map[string]AgentBehavior

    // ✅ New: Parallel execution (Issue #9)
    ParallelGroups map[string]ParallelGroupConfig
}

// ✅ New: Parallel execution configuration
type ParallelGroupConfig struct {
    Agents         []string    // Agents to run in parallel
    WaitForAll     bool        // Wait for all to complete?
    TimeoutSeconds int         // Max time for parallel execution
    NextAgent      string      // After parallel, route to this agent
    Description    string      // What is this parallel group?
}
```

---

## 8. Configuration Enhancements

### crew.yaml Enhancements

```yaml
version: "1.0"
name: my-crew
description: Crew description
entry_point: orchestrator

agents:
  - orchestrator
  - executor

settings:
  max_handoffs: 5
  max_rounds: 10
  timeout_seconds: 300
  language: en
  organization: My-Org

# ✅ New: Detailed routing configuration
routing:
  signals:
    orchestrator:
      - signal: "[ROUTE_EXECUTOR]"
        target: executor
        description: "Route to executor"

  defaults:
    orchestrator: executor
    executor: null

  # ✅ New: Agent behavior specification
  agent_behaviors:
    orchestrator:
      wait_for_signal: true
      auto_route: false
      is_terminal: false
      description: "Orchestrator waits for explicit signals"
    executor:
      is_terminal: true
      description: "Executor is the endpoint"

  # ✅ New: Parallel execution groups
  parallel_groups:
    diagnostics:
      agents: ["cpu-monitor", "memory-monitor", "disk-monitor"]
      wait_for_all: true
      timeout_seconds: 30
      next_agent: analyzer
      description: "Parallel system diagnostics"
```

### agent.yaml Enhancements

```yaml
id: executor
name: Executor
role: Technical Expert
description: Executes technical tasks
backstory: I am a technical expert...

# ✅ New: Primary & Backup models
primary:
  model: gpt-4-turbo
  provider: openai
  provider_url: https://api.openai.com

backup:
  model: gpt-4o-mini
  provider: openai
  provider_url: https://api.openai.com

# Optional: Legacy format still supported
model: gpt-4-turbo
provider: openai
provider_url: https://api.openai.com

temperature: 0.5
is_terminal: true

tools:
  - GetCPUUsage
  - GetMemoryUsage
  - PingHost

handoff_targets: []

system_prompt: |
  You are {{name}}.
  Be precise and helpful.
```

---

## 9. Migration Guide

### From Old Format to New Format

**Old Agent Structure**:
```yaml
id: executor
model: gpt-4-turbo
provider: openai
provider_url: https://api.openai.com
```

**New Agent Structure** (Recommended):
```yaml
id: executor
primary:
  model: gpt-4-turbo
  provider: openai
  provider_url: https://api.openai.com
backup:
  model: gpt-4o-mini
  provider: openai
  provider_url: https://api.openai.com
```

**Backward Compatibility**: Old format still works, automatically converted to new format

### Code Migration

**Old Code**:
```go
agent := &Agent{
    ID:          "executor",
    Model:       "gpt-4-turbo",
    Provider:    "openai",
    ProviderURL: "https://api.openai.com",
}
```

**New Code** (Recommended):
```go
agent := &Agent{
    ID: "executor",
    Primary: &ModelConfig{
        Model:       "gpt-4-turbo",
        Provider:    "openai",
        ProviderURL: "https://api.openai.com",
    },
    Backup: &ModelConfig{
        Model:       "gpt-4o-mini",
        Provider:    "openai",
        ProviderURL: "https://api.openai.com",
    },
}
```

---

## 10. Performance Improvements

### ✅ Fix Issue #4: Parallel Agent Timeout

**Configurability**: `ParallelAgentTimeout` now configurable per crew

```go
crew := &Crew{
    Agents:                 agents,
    MaxRounds:             10,
    MaxHandoffs:           5,
    ParallelAgentTimeout:  120 * time.Second,  // ✅ Configurable
}
```

### ✅ Fix Issue #5: Max Tool Output

**Configurability**: `MaxToolOutputChars` now configurable per crew

```go
crew := &Crew{
    Agents:             agents,
    MaxToolOutputChars: 5000,  // ✅ Configurable (default: 2000)
}
```

### Benefits
- Prevent memory bloat from large tool outputs
- Balance verbosity vs. cost
- Per-crew customization

---

## 11. Testing & Validation

### New Test Files

- `agent_test.go` - Agent execution tests
- `crew_test.go` - Crew coordination tests
- `config_test.go` - Configuration loading tests
- `validation_test.go` - Validation system tests
- `http_test.go` - HTTP handler tests
- `shutdown_test.go` - Graceful shutdown tests
- `request_tracking_test.go` - Request tracking tests

### Running Tests

```bash
# Run all tests
go test ./...

# Run with coverage
go test -cover ./...

# Run specific test
go test -run TestValidation ./...
```

---

## 12. Summary of Changes

| Feature | Issue | Status | File | Notes |
|---------|-------|--------|------|-------|
| Model Fallback | #1 | ✅ Complete | agent.go | Primary + Backup |
| YAML Validation | #6 | ✅ Complete | validation.go | 5-stage pipeline |
| Graceful Shutdown | #18 | ✅ Complete | shutdown.go | Signal handling |
| Request Tracking | #17 | ✅ Complete | request_tracking.go | UUID-based |
| Metrics Collection | #14 | ✅ Complete | metrics.go | SystemMetrics |
| Tool Validation | #25 | ✅ Complete | crew.go | Parameter check |
| Parallel Timeout | #4 | ✅ Complete | crew.go | Configurable |
| Tool Output Limit | #5 | ✅ Complete | crew.go | Configurable |
| Provider Factory | - | ✅ Complete | agent.go | Multi-provider |
| Routing Config | - | ✅ Complete | config.go | Signal-based |

---

## 13. Backward Compatibility

✅ **Fully Backward Compatible**

- Old agent YAML format still works
- Old code using Model/Provider still works
- Configuration loading handles both old & new formats
- Graceful migration path to new features

---

## 14. Next Steps

1. **Update your crew.yaml** to use new routing configuration
2. **Add primary model** to agent YAML files
3. **Consider adding backup model** for high-availability setups
4. **Enable metrics collection** for production monitoring
5. **Use request ID tracking** for debugging

---

## 15. References

- [CONFIG_SPECIFICATION.md](CONFIG_SPECIFICATION.md) - Updated with new fields
- [Core Library API](#) - Full API reference
- [Examples](#) - Updated examples with new features
- [Testing Guide](#) - How to test configurations

---

**Questions or Issues?** Check the troubleshooting section or review test files for usage examples.

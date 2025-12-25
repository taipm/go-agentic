# Chi Tiết Dependency Map & Implementation Checklist

---

## 1. CURRENT ARCHITECTURE - DEPENDENCY TREE

### 1.1 Current File Relationships

```
LAYER 4: Application
├── examples/
├── cmd/
└── main.go
    ↓
LAYER 3: Orchestrator (HIGH COUPLING - 85/100)
├── crew.go
│   ├── Depends: types, config_types, agent_types, signal, metrics, validation
│   ├── validation.go (900 lines)
│   │   ├── Depends: config_types, agent_types
│   │   ├── Uses: ConfigValidator class
│   │   └── Size: 34 functions, 2402 lines (including tests)
│   │
│   ├── config_loader.go (546 lines)
│   │   ├── Depends: validation, defaults, types
│   │   ├── Uses: LoadCrewConfig, LoadAgentConfigs, CreateAgentFromConfig, ConfigToHardcodedDefaults
│   │   └── Problem: mixing concerns (load + validate + convert)
│   │
│   ├── defaults.go (526 lines)
│   │   ├── Depends: types, time
│   │   ├── Uses: DefaultHardcodedDefaults, ConfigToHardcodedDefaults
│   │   └── Purpose: Default values for everything
│   │
│   ├── agent_execution.go (631 lines)
│   │   ├── Depends: providers/openai, providers/ollama, agent_cost
│   │   ├── Uses: ExecuteAgent, ExecuteAgentStream, buildSystemPrompt
│   │   └── Functions: 19
│   │
│   ├── agent_cost.go (195 lines)
│   │   ├── Depends: types, agent_types
│   │   ├── Uses: EstimateTokens, CalculateCost, CalculateOutputCost, ResetDailyMetricsIfNeeded
│   │   └── Functions: 8
│   │
│   ├── tool_execution.go (392 lines)
│   │   ├── Depends: types, internal
│   │   ├── Uses: ExecuteCalls, safeExecuteTool, executeCallsWithTimeout
│   │   └── Functions: 12
│   │
│   ├── team_execution.go (588 lines) ⚠️ HIGH COMPLEXITY
│   │   ├── Depends: types, context, log
│   │   ├── Uses: OutputHandler, executionContext, executeWorkflow, executeAgentWithMetrics
│   │   ├── Nested Logic: 4-5 levels deep (callback hell)
│   │   └── Functions: 33
│   │
│   ├── team_routing.go (284 lines)
│   │   ├── Depends: types, signal, log
│   │   ├── Uses: executeSignalBasedRouting, determineNextAgent
│   │   └── Functions: 8
│   │
│   ├── team_parallel.go (155 lines) ⚠️ COMPLEX
│   │   ├── Depends: types, context, sync, golang.org/x/sync/errgroup
│   │   ├── Uses: ExecuteParallelStream (155 lines single function)
│   │   └── Functions: 2
│   │
│   ├── team_history.go (284 lines)
│   │   ├── Depends: types, sync, log, context
│   │   ├── Uses: HistoryManager struct, message management
│   │   └── Thread-safe: yes (RWMutex)
│   │
│   ├── team_tools.go (102 lines)
│   │   ├── Depends: types, strings, fmt
│   │   ├── Uses: formatToolResults, getToolParameterNames
│   │   └── Functions: 4
│   │
│   ├── metrics.go (500+ lines)
│   │   ├── Depends: types, sync, time, json, encoding/json
│   │   ├── Uses: MetricsCollector, RecordAgentExecution, ExportJSON, ExportPrometheus
│   │   └── Functions: 21
│   │
│   ├── types.go (base types)
│   │   ├── Crew
│   │   ├── Agent
│   │   ├── Task
│   │   ├── Signal
│   │   └── No dependencies (layer 0)
│   │
│   ├── config_types.go
│   │   ├── CrewConfig
│   │   ├── AgentConfig
│   │   └── Depends: types.go only
│   │
│   ├── agent_types.go
│   │   ├── AgentMetadata
│   │   ├── AgentQuotas
│   │   └── Depends: types.go only
│   │
│   └── execution_constants.go
│       ├── DefaultMaxParallelAgents
│       ├── DefaultHistoryTrimThreshold
│       └── No dependencies
│
└── signal/
    ├── types.go
    ├── registry.go
    ├── validator.go
    └── Used by: team_routing.go only

LAYER 2: Providers
├── providers/openai/
│   └── Used by: agent_execution.go only
├── providers/ollama/
│   └── Used by: agent_execution.go only
└── providers/interface.go
    └── Provider interface

LAYER 1: Supporting
├── internal/
│   ├── errors.go
│   ├── timeout.go
│   └── validation.go
└── tools/
    ├── errors.go
    └── timeout.go
```

### 1.2 Detailed Dependency Matrix

```
                  types config agent team   exec  metric signal provider internal
crew.go           ✓     ✓      ✓     ✓      ✓     ✓      ✓      ✓        ✓
team_execution    ✓     -      -     -      -     ✓      -      -        -
team_routing      ✓     -      -     -      -     -      ✓      -        -
agent_execution   ✓     ✓      ✓     -      -     ✓      -      ✓        ✓
tool_execution    ✓     -      -     -      -     -      -      -        ✓
validation        ✓     ✓      ✓     -      -     -      -      -        -
config_loader     ✓     ✓      ✓     -      -     -      -      -        -
metrics           ✓     -      -     -      -     -      -      -        -
agent_cost        ✓     ✓      -     -      -     -      -      -        -

Total imports (crew.go): 15
Average per file: 4-5 imports
```

---

## 2. NEW ARCHITECTURE - PROPOSED STRUCTURE

### 2.1 New Package Hierarchy

```
LAYER 0: Common (No dependencies except stdlib)
core/common/
├── types.go
│   ├── Agent
│   ├── Crew
│   ├── Task
│   ├── Message
│   ├── CrewConfig
│   ├── AgentConfig
│   ├── AgentMetadata
│   ├── AgentQuotas
│   └── Others
├── constants.go
│   ├── DefaultMaxParallelAgents
│   ├── DefaultHistoryTrimThreshold
│   ├── DefaultMaxTokens
│   └── All execution constants
├── errors.go
│   ├── ValidationError
│   ├── ExecutionError
│   ├── TimeoutError
│   ├── QuotaExceededError
│   └── Custom error types
└── helpers.go
    ├── Utility functions
    ├── Common algorithms
    └── Non-business logic helpers

Dependencies: stdlib only ✓
Used by: ALL other packages
Imports to: NOTHING
```

### 2.2 Configuration Layer

```
core/config/
├── types.go (PURE TYPES - no logic)
│   ├── CrewConfig struct
│   ├── AgentConfig struct
│   ├── ConfigMetadata struct
│   └── Constants for config (KEY_AGENT_ID, etc.)
│
├── loader.go (LOADING ONLY)
│   ├── LoadCrewConfig(path string) (*CrewConfig, error)
│   ├── LoadAgentConfigs(basePath string) (map[string]*AgentConfig, error)
│   ├── loadFromYAML(path string, v interface{}) error
│   └── expandEnvVars(value string) string
│
├── converter.go (CONVERSION ONLY)
│   ├── ConfigToHardcodedDefaults(cfg *AgentConfig) *HardcodedDefaults
│   ├── BuildAgentMetadata(cfg *AgentConfig) *AgentMetadata
│   ├── BuildAgentQuotas(cfg *AgentConfig) *AgentQuotas
│   └── convertProviderConfig(cfg interface{}) (*ProviderConfig, error)
│
└── defaults.go (OPTIONAL - if needed)
    ├── DefaultCrewConfig() *CrewConfig
    ├── DefaultAgentConfig() *AgentConfig
    └── DefaultHardcodedDefaults() *HardcodedDefaults

Dependencies: common/types, stdlib, gopkg.in/yaml.v3
Used by: executor, validation
Imports to: common only
```

### 2.3 Validation Layer

```
core/validation/
├── crew.go
│   ├── ValidateCrewConfig(cfg *CrewConfig) error
│   ├── validateCrewRequiredFields(cfg *CrewConfig) error
│   ├── validateEntryPointAndBuildMap(cfg *CrewConfig, agents map[string]*AgentConfig) error
│   └── validateCrewSettings(cfg *CrewConfig) error
│
├── agent.go
│   ├── ValidateAgentConfig(cfg *AgentConfig) error
│   ├── validateAgentBasicConstraints(cfg *AgentConfig) error
│   ├── validateAgentTemperature(temp float64) error
│   ├── validateAgentModel(model string) error
│   └── validateAgentProvider(provider string) error
│
├── routing.go
│   ├── ValidateSignals(cfg *CrewConfig, signals map[string]interface{}) error
│   ├── ValidateParallelGroups(cfg *CrewConfig, groups map[string]interface{}) error
│   ├── DetectCircularReferences(cfg *CrewConfig) error
│   ├── CheckReachability(cfg *CrewConfig) error
│   └── validateSignalFormat(signal interface{}) error
│
├── helpers.go
│   ├── validateDuration(d time.Duration) error
│   ├── validateInt(val int, min, max int) error
│   ├── validateFloatRange(val float64, min, max float64) error
│   ├── contains(arr []string, s string) bool
│   ├── ValidateRequiredFields(obj interface{}, fields []string) error
│   └── formatValidationError(err error) string
│
└── graph.go (OPTIONAL - for signal graph validation)
    ├── BuildAgentGraph(cfg *CrewConfig) *AgentGraph
    ├── DetectCycles(graph *AgentGraph) [][]string
    └── ValidateGraphProperties(graph *AgentGraph) error

Dependencies: common/types, common/errors, stdlib
Used by: config/loader, executor
Imports to: common only
```

### 2.4 Agent Layer

```
core/agent/
├── execution.go
│   ├── (ce *CrewExecutor) Execute(ctx context.Context, input interface{}) (interface{}, error)
│   ├── (ce *CrewExecutor) ExecuteStream(ctx context.Context, input interface{}) (chan interface{}, error)
│   ├── executeWithModelConfig(ctx context.Context, agent *Agent, ...) (*Response, error)
│   ├── validateAndCheckCostLimits(agent *Agent, estimatedTokens int) error
│   └── executeProviderCall(ctx context.Context, agent *Agent, req *ProviderRequest) (*Response, error)
│
├── provisioning.go
│   ├── CreateAgentFromConfig(cfg *config.AgentConfig) (*Agent, error)
│   ├── setupAgentMetadata(agent *Agent, cfg *config.AgentConfig) error
│   ├── setupAgentQuotas(agent *Agent, cfg *config.AgentConfig) error
│   └── setupAgentTools(agent *Agent, tools []interface{}) error
│
├── cost.go (from agent_cost.go)
│   ├── (a *Agent) EstimateTokens(text string) int
│   ├── (a *Agent) CalculateCost(tokens int) float64
│   ├── (a *Agent) CalculateOutputCost(tokens int) float64
│   ├── (a *Agent) CalculateTotalCost(input, output int) float64
│   ├── (a *Agent) CheckCostLimits() error
│   └── (a *Agent) ResetDailyMetricsIfNeeded()
│
├── messaging.go
│   ├── (a *Agent) BuildSystemPrompt() string
│   ├── (a *Agent) BuildGenericPrompt(userMsg string) string
│   ├── (a *Agent) ConvertToProviderMessages(msgs []interface{}) interface{}
│   ├── (a *Agent) ConvertToolsToProvider(tools []interface{}) interface{}
│   ├── (a *Agent) ConvertToolCallsFromProvider(calls interface{}) []interface{}
│   └── (a *Agent) InvalidateSystemPromptCache()
│
└── interface.go (OPTIONAL - for dependency injection)
    ├── AgentExecutor interface
    ├── AgentProvisioner interface
    └── AgentMessenger interface

Dependencies: common/types, config/types, provider/interface, metrics, stdlib
Used by: executor, workflow
Imports to: common, config, provider
```

### 2.5 Tool Layer

```
core/tool/
├── execution.go
│   ├── (ce *CrewExecutor) ExecuteCalls(ctx context.Context, ...) (*ToolExecutionResult, error)
│   ├── executeCallsWithTimeout(ctx context.Context, ...) error
│   ├── executeSingleCall(ctx context.Context, tool Tool, ...) interface{}
│   ├── safeExecuteTool(ctx context.Context, tool Tool, ...) (result interface{}, err error)
│   ├── executeToolWithPanic(tool Tool, args map[string]interface{}) (result interface{}, recovered error)
│   └── handleToolTimeout(tool Tool, timeoutDuration time.Duration) interface{}
│
├── formatting.go
│   ├── formatToolResults(results interface{}, format string) string
│   ├── FormatAsJSON(results interface{}) string
│   ├── FormatAsText(results interface{}) string
│   ├── getToolParameterNames(tool Tool) []string
│   └── parseToolArguments(argsStr string, tool Tool) (map[string]interface{}, error)
│
└── interface.go (OPTIONAL)
    ├── Tool interface
    ├── ToolExecutor interface
    └── ToolFormatter interface

Dependencies: common/types, metrics, stdlib
Used by: executor, workflow
Imports to: common, metrics
```

### 2.6 Workflow Layer

```
core/workflow/
├── handler.go
│   ├── OutputHandler interface
│   │   ├── SendStreamEvent(data interface{})
│   │   ├── HandleError(err error)
│   │   └── HandleSuccess(result interface{})
│   │
│   ├── SyncHandler struct
│   │   └── implementations
│   │
│   └── StreamHandler struct
│       └── implementations
│
├── execution.go
│   ├── (ce *CrewExecutor) executeAgentWithMetrics(ctx context.Context, ...) (*Agent, interface{}, error)
│   ├── (ce *CrewExecutor) handleAgentError(agent *Agent, err error) (*Agent, error)
│   ├── updateAgentMetrics(agent *Agent, result *ExecutionResult)
│   ├── recordAgentExecution(agent *Agent, result *ExecutionResult)
│   ├── determineNextTransition(lastOutput interface{}, signal *Signal) *Agent
│   └── executeAgentWithContext(ctx context.Context, agent *Agent, ...) error
│
├── routing.go (from team_routing.go)
│   ├── (ce *CrewExecutor) executeSignalBasedRouting(lastOutput interface{}, ...) (*Agent, error)
│   ├── evaluateSignalCondition(signal *Signal, output interface{}) bool
│   ├── routeByBehavior(behavior *Behavior, agents map[string]*Agent) *Agent
│   └── routeBySignal(signal *Signal, agents map[string]*Agent) *Agent
│
└── parallel.go (from team_parallel.go)
    ├── (ce *CrewExecutor) executeParallelStream(ctx context.Context, ...) (*ParallelExecutionResult, error)
    ├── executeParallelGroupWithFallback(ctx context.Context, group []Agent, ...) error
    └── collectParallelResults(results chan *AgentResult, timeout time.Duration) []*AgentResult

Dependencies: common/types, agent/execution, tool/execution, metrics, signal, stdlib
Used by: executor
Imports to: common, agent, tool, metrics, signal
```

### 2.7 Executor Layer (Top-level Orchestrator)

```
core/executor/
├── executor.go
│   ├── CrewExecutor struct (MAIN - moved from crew.go)
│   │   ├── Crew *Crew
│   │   ├── Agents map[string]*Agent
│   │   ├── History *HistoryManager
│   │   ├── Metrics *metrics.Collector
│   │   ├── SignalRegistry signal.Registry
│   │   ├── Config *ExecutorConfig
│   │   └── (removed: validation, config loading, tool formatting)
│   │
│   ├── NewCrewExecutor(crew *Crew) (*CrewExecutor, error)
│   ├── NewCrewExecutorFromConfig(configPath string) (*CrewExecutor, error)
│   ├── (ce *CrewExecutor) Execute(ctx context.Context, input interface{}) (interface{}, error)
│   ├── (ce *CrewExecutor) ExecuteStream(ctx context.Context, input interface{}) (chan interface{}, error)
│   ├── (ce *CrewExecutor) SetSignalRegistry(registry signal.Registry)
│   ├── (ce *CrewExecutor) SetVerbose(verbose bool)
│   └── (ce *CrewExecutor) Close() error
│
├── workflow.go
│   ├── (ce *CrewExecutor) executeWorkflow(ctx context.Context, input interface{}) error
│   ├── (ce *CrewExecutor) initializeWorkflow(input interface{}) error
│   ├── (ce *CrewExecutor) runMainLoop(ctx context.Context, handler workflow.OutputHandler) error
│   └── (ce *CrewExecutor) finalizeWorkflow() (interface{}, error)
│
└── history.go
    ├── HistoryManager struct
    ├── (ce *CrewExecutor) appendMessage(msg interface{})
    ├── (ce *CrewExecutor) getHistoryCopy() []interface{}
    ├── (ce *CrewExecutor) GetHistory() []interface{}
    ├── (ce *CrewExecutor) ClearHistory()
    ├── (ce *CrewExecutor) trimHistoryIfNeeded()
    ├── (ce *CrewExecutor) estimateHistoryTokens() int
    ├── (ce *CrewExecutor) copyHistory(msgs []interface{}) []interface{}
    └── (ce *CrewExecutor) handleConcurrentAccess(fn func() error) error

Dependencies: common/types, config/types, agent/execution, workflow/*, tool/execution, metrics, signal, stdlib
Used by: Application/CLI
Imports to: ALL modules (OK - it's the orchestrator)
```

### 2.8 Metrics Layer (Optional separation)

```
core/metrics/
├── collector.go
│   ├── MetricsCollector struct
│   ├── NewMetricsCollector() *MetricsCollector
│   ├── (mc *MetricsCollector) RecordAgentExecution(...)
│   ├── (mc *MetricsCollector) RecordToolExecution(...)
│   ├── (mc *MetricsCollector) RecordTokenUsage(...)
│   ├── (mc *MetricsCollector) GetMetrics() *Metrics
│   └── (mc *MetricsCollector) Reset()
│
├── exporters.go
│   ├── ExportJSON(mc *MetricsCollector) ([]byte, error)
│   ├── ExportPrometheus(mc *MetricsCollector) (string, error)
│   └── ExportCSV(mc *MetricsCollector) (string, error)
│
└── logging.go (OPTIONAL)
    ├── LogAgentMetadata(agent *Agent) string
    ├── LogExecutionMetrics(metrics *Metrics) string
    └── LogPerformanceStats(stats *PerformanceStats) string

Dependencies: common/types, stdlib (json, fmt, etc.)
Used by: executor, workflow, agent
Imports to: common only
```

### 2.9 Provider Layer (Unchanged)

```
core/provider/
├── interface.go
│   ├── Provider interface
│   ├── ProviderConfig interface
│   ├── ProviderRequest struct
│   └── ProviderResponse struct
│
├── registry.go
│   ├── RegisterProvider(name string, provider Provider)
│   ├── GetProvider(name string) (Provider, error)
│   └── ListProviders() []string
│
├── openai/
│   ├── provider.go
│   └── (specific OpenAI implementation)
│
└── ollama/
    ├── provider.go
    └── (specific Ollama implementation)

Dependencies: common/types, stdlib, external APIs
Used by: agent/execution only
Imports to: common only
```

---

## 3. DEPENDENCY REDUCTION ANALYSIS

### 3.1 Before: crew.go Imports

```
crew.go (1500+ lines)
├── github.com/taipm/go-agentic/core/types          ✓
├── github.com/taipm/go-agentic/core/config_types   ✓
├── github.com/taipm/go-agentic/core/agent_types    ✓
├── github.com/taipm/go-agentic/core/signal         ✓
├── github.com/taipm/go-agentic/core/metrics        ✓
├── github.com/taipm/go-agentic/core/validation     ✓
├── github.com/taipm/go-agentic/core/config_loader  ✓
├── github.com/taipm/go-agentic/core/defaults       ✓
├── github.com/taipm/go-agentic/core/agent_cost     ✓
├── github.com/taipm/go-agentic/core/agent_execution ✓
├── github.com/taipm/go-agentic/core/tool_execution ✓
├── github.com/taipm/go-agentic/core/team_execution ✓
├── github.com/taipm/go-agentic/core/team_routing   ✓
├── github.com/taipm/go-agentic/core/team_parallel  ✓
└── github.com/taipm/go-agentic/core/team_history   ✓

Total: 15 imports
Coupling Score: 85/100
```

### 3.2 After: executor/executor.go Imports

```
executor/executor.go (400-500 lines)
├── github.com/taipm/go-agentic/core/common         ✓
├── github.com/taipm/go-agentic/core/config         ✓
├── github.com/taipm/go-agentic/core/signal         ✓
├── github.com/taipm/go-agentic/core/executor/history ✓
├── github.com/taipm/go-agentic/core/metrics        ✓
├── github.com/taipm/go-agentic/core/workflow       ✓
└── (stdlib only)                                    ✓

Total: 6 imports
Coupling Score: 50/100
Reduction: 60% ✅
```

### 3.3 After: Other Key Files

```
config/loader.go
├── github.com/taipm/go-agentic/core/common
├── github.com/taipm/go-agentic/core/config
├── github.com/taipm/go-agentic/core/validation
└── (stdlib + gopkg.in/yaml.v3)
Total: 3 imports

validation/crew.go
├── github.com/taipm/go-agentic/core/common
├── github.com/taipm/go-agentic/core/config
└── (stdlib)
Total: 2 imports

agent/execution.go
├── github.com/taipm/go-agentic/core/common
├── github.com/taipm/go-agentic/core/config
├── github.com/taipm/go-agentic/core/provider
├── github.com/taipm/go-agentic/core/metrics
└── (stdlib)
Total: 4 imports

workflow/execution.go
├── github.com/taipm/go-agentic/core/common
├── github.com/taipm/go-agentic/core/agent
├── github.com/taipm/go-agentic/core/tool
├── github.com/taipm/go-agentic/core/metrics
└── (stdlib)
Total: 4 imports
```

### 3.4 Coupling Scores Summary

| File | Before | After | Reduction |
|------|--------|-------|-----------|
| crew.go → executor | 85 | 50 | 41% ↓ |
| (none) → config/loader | - | 40 | New optimized |
| validation.go | 75 | 45 | 40% ↓ |
| agent_execution.go | 65 | 50 | 23% ↓ |
| team_execution.go | 55 | 50 | 9% ↓ |
| **AVERAGE** | **68** | **47** | **31% ↓** |

---

## 4. IMPLEMENTATION CHECKLIST - DETAILED

### Phase 1: Foundation (Week 1 - Estimated 40 hours)

#### Step 1.1: Create /core/common package
- [ ] Create directory: `mkdir -p core/common`
- [ ] Create `core/common/types.go` (consolidate types)
  - [ ] Move all types from `types.go`
  - [ ] Move all types from `config_types.go`
  - [ ] Move all types from `agent_types.go`
  - [ ] Verify no circular imports
  - [ ] Add comments for each type
- [ ] Create `core/common/constants.go`
  - [ ] Move all constants from `execution_constants.go`
  - [ ] Add new constants for validation
  - [ ] Document constant usage
- [ ] Create `core/common/errors.go`
  - [ ] Move error types from `tools/errors.go`
  - [ ] Add new error types (ValidationError, ExecutionError, etc.)
  - [ ] Implement error interfaces (Error(), Unwrap())
- [ ] Create `core/common/helpers.go`
  - [ ] Move utilities from `tools/timeout.go`
  - [ ] Add common utility functions
- [ ] Create `core/common/common_test.go`
  - [ ] Test type definitions
  - [ ] Test error types
  - [ ] 80%+ coverage

#### Step 1.2: Create /core/config package
- [ ] Create directory: `mkdir -p core/config`
- [ ] Create `core/config/types.go`
  - [ ] Create pure config type definitions
  - [ ] NO logic, only struct + constants
  - [ ] Add JSON/YAML struct tags
- [ ] Copy to `core/config/loader.go` (minimal modification)
  - [ ] Remove validation calls (will be separate)
  - [ ] Remove converter calls (will be separate)
  - [ ] Keep: LoadCrewConfig, LoadAgentConfigs
  - [ ] Keep: loadFromYAML, expandEnvVars
- [ ] Create `core/config/converter.go`
  - [ ] Move: ConfigToHardcodedDefaults
  - [ ] Move: BuildAgentMetadata
  - [ ] Move: BuildAgentQuotas
  - [ ] Move: convertProviderConfig
- [ ] Create `core/config/defaults.go` (optional, keep if useful)
  - [ ] Move: DefaultHardcodedDefaults
  - [ ] Move: default value factories
- [ ] Update imports in copied files
  - [ ] Replace `config_types` imports with `config/types`
  - [ ] Replace `agent_types` imports with `common/types`
  - [ ] Replace `types` imports with `common/types`

#### Step 1.3: Create /core/validation package
- [ ] Create directory: `mkdir -p core/validation`
- [ ] Create `core/validation/crew.go`
  - [ ] Move: ValidateCrewConfig
  - [ ] Move: validateCrewRequiredFields
  - [ ] Move: validateEntryPointAndBuildMap
  - [ ] Move: validateCrewSettings
  - [ ] Update imports
- [ ] Create `core/validation/agent.go`
  - [ ] Move: ValidateAgentConfig
  - [ ] Move: validateAgentBasicConstraints
  - [ ] Move: validateAgentTemperature
  - [ ] Move: validateAgentModel
  - [ ] Move: validateAgentProvider
  - [ ] Update imports
- [ ] Create `core/validation/routing.go`
  - [ ] Move: ValidateSignals
  - [ ] Move: ValidateParallelGroups
  - [ ] Move: DetectCircularReferences
  - [ ] Move: CheckReachability
  - [ ] Move: validateSignalFormat
  - [ ] Update imports
- [ ] Create `core/validation/helpers.go`
  - [ ] Move: validateDuration
  - [ ] Move: validateInt
  - [ ] Move: validateFloatRange
  - [ ] Move: contains
  - [ ] Move: ValidateRequiredFields
  - [ ] Move: formatValidationError
- [ ] Create `core/validation/validation_test.go`
  - [ ] Move all validation tests
  - [ ] Update test imports
  - [ ] Ensure all tests pass
- [ ] Update imports in new files
  - [ ] Replace `types` with `common/types`
  - [ ] Replace `config_types` with `config/types`

#### Step 1.4: Update all imports project-wide
- [ ] Create migration script: `scripts/migrate_imports.sh`
  ```bash
  # Replace types imports
  find core -name "*.go" -type f -exec sed -i '' \
    's|github.com/taipm/go-agentic/core/types|github.com/taipm/go-agentic/core/common|g' {} \;

  # Replace config_types imports
  find core -name "*.go" -type f -exec sed -i '' \
    's|github.com/taipm/go-agentic/core/config_types|github.com/taipm/go-agentic/core/config|g' {} \;

  # Replace agent_types imports
  find core -name "*.go" -type f -exec sed -i '' \
    's|github.com/taipm/go-agentic/core/agent_types|github.com/taipm/go-agentic/core/common|g' {} \;
  ```
- [ ] Run script: `bash scripts/migrate_imports.sh`
- [ ] Manually verify critical files:
  - [ ] crew.go
  - [ ] agent_execution.go
  - [ ] config_loader.go
  - [ ] team_execution.go
- [ ] Run go mod tidy
- [ ] Run tests: `go test ./core/...`

#### Step 1.5: Verify Phase 1
- [ ] All tests pass: `go test ./core/...`
- [ ] No circular dependencies: `go mod graph | grep "circular"`
- [ ] Build succeeds: `go build ./...`
- [ ] Code coverage maintained: `go test -cover ./core/...`
- [ ] Commit: "refactor: Phase 1 - Create common, config, validation packages"

---

### Phase 2: Extract Validation (Week 2 - Estimated 30 hours)

#### Step 2.1: Decouple validation from config_loader
- [ ] Update `core/config/loader.go`
  - Remove: validation logic
  - Add: call to `validation.ValidateCrewConfig`
  - Add: call to `validation.ValidateAgentConfig`
  - [ ] Test with actual config files
  - [ ] Ensure error messages are clear
- [ ] Update test file: `core/config/config_test.go`
  - [ ] Update validation test imports
  - [ ] Move validation-only tests to `core/validation/*_test.go`
  - [ ] Keep loader tests in config_test.go
- [ ] Run tests: `go test ./core/config/... ./core/validation/...`
- [ ] Verify validation error messages

#### Step 2.2: Consolidate validation helpers
- [ ] Review `validation/helpers.go`
  - [ ] Ensure all validators are there
  - [ ] Add missing validators (e.g., for new fields)
  - [ ] Document each validator
- [ ] Create validation builder (optional - for complex validations)
  ```go
  type ValidationBuilder struct {
      cfg *config.CrewConfig
      errs []error
  }

  func (vb *ValidationBuilder) ValidateAll() error { ... }
  ```

#### Step 2.3: Add graph validation (optional but recommended)
- [ ] Create `core/validation/graph.go`
  - [ ] BuildAgentGraph function
  - [ ] DetectCycles function
  - [ ] ValidateGraphProperties function
  - [ ] Add tests

#### Step 2.4: Verify Phase 2
- [ ] All validation tests pass
- [ ] Config loading works with new validation
- [ ] Error messages are helpful
- [ ] Commit: "refactor: Phase 2 - Extract validation layer"

---

### Phase 3: Agent & Tool Modules (Week 3 - Estimated 40 hours)

#### Step 3.1: Create /core/agent package
- [ ] Create directory: `mkdir -p core/agent`
- [ ] Create `core/agent/execution.go`
  - [ ] Move entire `agent_execution.go` content
  - [ ] Update imports (types → common/types, config_types → config/types)
  - [ ] Update method receivers to work with agent package
  - [ ] Test all methods individually
- [ ] Create `core/agent/provisioning.go`
  - [ ] Move: CreateAgentFromConfig
  - [ ] Move: setupAgentMetadata
  - [ ] Move: setupAgentQuotas
  - [ ] Move: setupAgentTools
  - [ ] Add tests
- [ ] Create `core/agent/cost.go`
  - [ ] Move all functions from `agent_cost.go`
  - [ ] Keep as methods on Agent struct
  - [ ] Update imports
  - [ ] Add/update tests
- [ ] Create `core/agent/messaging.go`
  - [ ] Move: BuildSystemPrompt
  - [ ] Move: BuildGenericPrompt
  - [ ] Move: ConvertToProviderMessages
  - [ ] Move: ConvertToolsToProvider
  - [ ] Move: ConvertToolCallsFromProvider
  - [ ] Move: InvalidateSystemPromptCache
  - [ ] Add tests
- [ ] Create `core/agent/agent_test.go`
  - [ ] Move all agent tests
  - [ ] Update test imports
  - [ ] Ensure 80%+ coverage

#### Step 3.2: Create /core/tool package
- [ ] Create directory: `mkdir -p core/tool`
- [ ] Create `core/tool/execution.go`
  - [ ] Move: ExecuteCalls
  - [ ] Move: executeCallsWithTimeout
  - [ ] Move: executeSingleCall
  - [ ] Move: safeExecuteTool
  - [ ] Move: executeToolWithPanic
  - [ ] Move: handleToolTimeout
  - [ ] Update imports
  - [ ] Update test structure
- [ ] Create `core/tool/formatting.go`
  - [ ] Move: formatToolResults (from team_tools.go)
  - [ ] Move: FormatAsJSON, FormatAsText
  - [ ] Move: getToolParameterNames (from team_tools.go)
  - [ ] Move: parseToolArguments (from team_tools.go)
  - [ ] Add tests
- [ ] Create `core/tool/tool_test.go`
  - [ ] Move all tool tests
  - [ ] Update imports

#### Step 3.3: Update executor references
- [ ] Update `crew.go` to import from new packages
  - [ ] agent execution → `core/agent`
  - [ ] tool execution → `core/tool`
  - [ ] Verify all function calls work
- [ ] Update tests that import these modules

#### Step 3.4: Verify Phase 3
- [ ] All tests pass: `go test ./core/agent/... ./core/tool/...`
- [ ] No circular dependencies
- [ ] Build succeeds
- [ ] agent/execution tests pass with new imports
- [ ] Commit: "refactor: Phase 3 - Extract agent and tool packages"

---

### Phase 4: Workflow & Executor (Week 4 - Estimated 50 hours)

#### Step 4.1: Create /core/workflow package
- [ ] Create directory: `mkdir -p core/workflow`
- [ ] Create `core/workflow/handler.go`
  - [ ] Move: OutputHandler interface (from team_execution.go)
  - [ ] Move: SyncHandler implementation
  - [ ] Move: StreamHandler implementation
  - [ ] Add tests
  - [ ] Document handler responsibilities
- [ ] Create `core/workflow/execution.go`
  - [ ] Move: executeAgentWithMetrics (from team_execution.go)
  - [ ] Move: handleAgentError
  - [ ] Move: updateAgentMetrics
  - [ ] Move: recordAgentExecution
  - [ ] Move: determineNextTransition
  - [ ] Move: executeAgentWithContext
  - [ ] Refactor nested logic into smaller functions
  - [ ] Add tests
- [ ] Create `core/workflow/routing.go`
  - [ ] Move: executeSignalBasedRouting (from team_routing.go)
  - [ ] Move: evaluateSignalCondition
  - [ ] Move: routeByBehavior
  - [ ] Move: routeBySignal
  - [ ] Add tests
  - [ ] Verify signal routing logic
- [ ] Create `core/workflow/parallel.go`
  - [ ] Move: ExecuteParallelStream (from team_parallel.go)
  - [ ] Extract nested logic into helpers:
    - [ ] executeParallelGroupWithFallback
    - [ ] collectParallelResults
  - [ ] Refactor 155-line function into manageable pieces
  - [ ] Add tests for each part
- [ ] Create `core/workflow/workflow_test.go`
  - [ ] Move all workflow tests
  - [ ] Update imports
  - [ ] Ensure tests cover refactored code

#### Step 4.2: Create /core/executor package
- [ ] Create directory: `mkdir -p core/executor`
- [ ] Create `core/executor/executor.go`
  - [ ] Move: CrewExecutor struct definition
  - [ ] Move: NewCrewExecutor
  - [ ] Move: NewCrewExecutorFromConfig
  - [ ] Move: Execute (main entry point)
  - [ ] Move: ExecuteStream
  - [ ] Move: SetSignalRegistry
  - [ ] Move: SetVerbose
  - [ ] Move: ClearResumeAgent
  - [ ] Move: GetResumeAgentID
  - [ ] Update imports (remove validation, config_loader dependencies)
  - [ ] Verify function signatures
- [ ] Create `core/executor/workflow.go`
  - [ ] Move: executeWorkflow (from crew.go)
  - [ ] Extract key operations:
    - [ ] initializeWorkflow
    - [ ] runMainLoop
    - [ ] finalizeWorkflow
  - [ ] Reduce nesting
  - [ ] Add tests
- [ ] Create `core/executor/history.go`
  - [ ] Move: HistoryManager struct (from team_history.go)
  - [ ] Move: appendMessage
  - [ ] Move: getHistoryCopy
  - [ ] Move: GetHistory
  - [ ] Move: ClearHistory
  - [ ] Move: trimHistoryIfNeeded
  - [ ] Move: estimateHistoryTokens
  - [ ] Move: copyHistory
  - [ ] Move: handleConcurrentAccess
  - [ ] Add tests
  - [ ] Verify thread-safety
- [ ] Create `core/executor/executor_test.go`
  - [ ] Move all executor tests
  - [ ] Update imports
  - [ ] Ensure all tests pass

#### Step 4.3: Refactor for reduced coupling
- [ ] Remove validation logic from executor.go
  - [ ] Move validation to initialization step
  - [ ] Call `validation.ValidateCrewConfig` before creation
- [ ] Remove config loading from executor.go
  - [ ] Keep: NewCrewExecutorFromConfig as thin wrapper
  - [ ] Call: `config.LoadCrewConfig`
  - [ ] Call: `validation.ValidateCrewConfig`
  - [ ] Call: `NewCrewExecutor`
- [ ] Update dependencies in executor.go
  - [ ] Remove: validation import
  - [ ] Remove: config_loader import
  - [ ] Remove: defaults import
  - [ ] Keep: only what's needed for execution
  - [ ] Target: ≤7 imports

#### Step 4.4: Verify Phase 4
- [ ] All tests pass: `go test ./core/executor/... ./core/workflow/...`
- [ ] crew.go still works (unchanged)
- [ ] No circular dependencies
- [ ] executor.go coupling reduced to ~50/100
- [ ] Build succeeds with all new packages
- [ ] Commit: "refactor: Phase 4 - Extract workflow and executor packages"

---

### Phase 5: Cleanup (Week 5 - Estimated 20 hours)

#### Step 5.1: Create backwards compatibility (optional)
- [ ] If keeping old crew.go, create deprecation wrapper
  ```go
  // crew.go - DEPRECATED
  // Use core/executor/executor.go instead

  // Re-export for backwards compatibility
  type CrewExecutor = executor.CrewExecutor
  var NewCrewExecutor = executor.NewCrewExecutor
  ```
- [ ] Or: Delete old crew.go and update all imports

#### Step 5.2: Update imports in old files
- [ ] Find all files importing old modules:
  ```bash
  grep -r "import.*core/crew" . --include="*.go"
  grep -r "import.*core/team_" . --include="*.go"
  grep -r "import.*core/validation" . --include="*.go"
  ```
- [ ] Update imports to new packages
- [ ] Examples:
  - `"github.com/taipm/go-agentic/core"` → `"github.com/taipm/go-agentic/core/executor"`
  - `"github.com/taipm/go-agentic/core/team_routing"` → `"github.com/taipm/go-agentic/core/workflow"`

#### Step 5.3: Delete old files (if hard break)
- [ ] Delete old files (after ensuring all imports updated):
  - [ ] `rm core/crew.go`
  - [ ] `rm core/team_*.go`
  - [ ] `rm core/types.go`
  - [ ] `rm core/config_types.go`
  - [ ] `rm core/agent_types.go`
  - [ ] `rm core/validation.go`
  - [ ] `rm core/config_loader.go`
  - [ ] `rm core/defaults.go`
  - [ ] `rm core/agent_execution.go`
  - [ ] `rm core/tool_execution.go`
  - [ ] `rm core/agent_cost.go`

#### Step 5.4: Update examples & documentation
- [ ] Update all examples to use new imports
  - [ ] `examples/*/*.go`
  - [ ] Update import statements
  - [ ] Test each example compiles
- [ ] Update README.md with new package structure
- [ ] Update architecture documentation
- [ ] Add migration guide for external users

#### Step 5.5: Final verification
- [ ] Run full test suite: `go test ./...`
- [ ] Build all examples: `go build ./examples/...`
- [ ] Run benchmarks
- [ ] Check code coverage: `go test -cover ./core/...`
- [ ] Final commit: "refactor: Phase 5 - Cleanup and finalization"

---

## 5. TESTING STRATEGY FOR MIGRATION

### Unit Tests (per Phase)
- [ ] After each phase, run: `go test ./core/... -v`
- [ ] Verify coverage: `go test -cover ./core/...`
- [ ] Target: ≥80% coverage

### Integration Tests
- [ ] Test full workflow after Phase 4
  ```go
  func TestIntegration_LoadConfigAndExecute(t *testing.T) {
      cfg, err := config.LoadCrewConfig("config.yaml")
      require.NoError(t, err)

      err = validation.ValidateCrewConfig(cfg)
      require.NoError(t, err)

      executor, err := executor.NewCrewExecutor(cfg)
      require.NoError(t, err)

      result, err := executor.Execute(ctx, input)
      require.NoError(t, err)
      assert.NotNil(t, result)
  }
  ```

### Circular Dependency Tests
```bash
# After each phase, verify no circular imports
go mod graph | grep -i circular
# Should return nothing
```

### Build Tests
```bash
# After each phase
go build ./...
# Should succeed with no warnings
```

---

## 6. DOCUMENTATION UPDATES

### Update Files
- [ ] `README.md` - Add new architecture section
- [ ] `docs/architecture.md` - Detailed architecture docs
- [ ] `docs/migration-guide.md` - For external users
- [ ] `CONTRIBUTING.md` - Update development guidelines
- [ ] Package comments - Add comments to each package explaining purpose

### Code Comments
- [ ] Add package-level comments
  ```go
  // Package config provides configuration loading and type definitions.
  // Use LoadCrewConfig to load configuration from YAML files.
  package config
  ```
- [ ] Add function comments for public APIs
- [ ] Document public interfaces

---

## 7. RISK MITIGATION

| Risk | Probability | Impact | Mitigation |
|------|-------------|--------|-----------|
| Circular dependencies | Medium | High | Use dependency analysis tools after each phase |
| Test failures | Low | Medium | Keep comprehensive test suite, update imports immediately |
| Performance regression | Low | Medium | Benchmark before/after, profile memory usage |
| Breaking changes | Medium | High | Provide migration guide, versioning, backwards compatibility |
| Developer confusion | High | Medium | Document thoroughly, train team, pair programming |

---

**Next Steps**:
1. Review this detailed plan
2. Get team approval
3. Create feature branch: `git checkout -b refactor/architecture-v2`
4. Start Phase 1 (Foundation)
5. Follow checklist for each phase

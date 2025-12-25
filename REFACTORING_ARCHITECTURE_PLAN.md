# Phương Án Tổ Chức Lại Package & Giảm Phụ Thuộc

**Ngày**: 2025-12-25
**Trạng thái**: Draft - Chờ Approval
**Objective**: Giảm coupling từ 85/100 → 60/100, tăng testability & maintainability

---

## 1. TÌNH HÌNH HIỆN TẠI

### 1.1 Cấu Trúc Thư Mục
```
/core
├── agents/              # LLM providers (OpenAI, Ollama)
├── signal/              # Signal-based routing
├── tools/               # Tool utilities
├── internal/            # Internal helpers
└── *.go files (39 files)
    ├── crew.go          # Monolithic orchestrator (1500+ lines)
    ├── team_*.go        # Extracted modules (4 files)
    ├── config_*.go      # Configuration (3 files)
    ├── validation.go    # Validator (900+ lines)
    ├── agent_*.go       # Agent logic (2 files)
    ├── metrics.go       # Metrics
    └── others
```

### 1.2 Coupling Analysis
| Component | Coupling Score | Phụ Thuộc Vào | Được Phụ Thuộc Bởi |
|-----------|----------------|---|---|
| crew.go | 85/100 | 15+ modules | Everything |
| validation.go | 75/100 | 5 modules | 3 modules |
| config_loader.go | 70/100 | 4 modules | 2 modules |
| agent_execution.go | 65/100 | 5 modules | 2 modules |
| team_execution.go | 55/100 | 3 modules | crew.go |

### 1.3 Vấn Đề Chính
1. **crew.go quá lớn** (1500+ lines) → khó test, khó maintain
2. **validation.go phức tạp** (900+ lines) → nhiều responsibilities
3. **config_loader.go lẫn lộn** → loader + validator + converter
4. **team_execution.go nested logic** → callback hell
5. **Agent execution phức tạp** → nhiều step (cost, validate, execute, metrics)

---

## 2. KIẾN TRÚC MỚI ĐỀ XUẤT

### 2.1 Package Structure (New)
```
/core
├── /executor              # NEW: Core execution orchestration
│   ├── executor.go        # CrewExecutor main struct (từ crew.go)
│   ├── workflow.go        # executeWorkflow logic
│   ├── history.go         # History management
│   └── executor_test.go   # Tests
│
├── /agent                 # NEW: Agent execution
│   ├── execution.go       # ExecuteAgent, ExecuteAgentStream
│   ├── provisioning.go    # CreateAgentFromConfig, agent setup
│   ├── cost.go            # Cost tracking (từ agent_cost.go)
│   ├── messaging.go       # Message conversion & prompts
│   └── agent_test.go      # Tests
│
├── /workflow              # NEW: Workflow management
│   ├── handler.go         # OutputHandler interface + impls
│   ├── execution.go       # executeAgentWithMetrics, transitions
│   ├── routing.go         # Signal routing logic
│   ├── parallel.go        # Parallel execution
│   └── workflow_test.go   # Tests
│
├── /config                # NEW: Configuration management
│   ├── types.go           # Config struct definitions
│   ├── loader.go          # LoadCrewConfig, LoadAgentConfigs
│   ├── converter.go       # ConfigToHardcodedDefaults
│   └── config_test.go     # Tests
│
├── /validation            # NEW: Validation layer
│   ├── crew.go            # ValidateCrewConfig
│   ├── agent.go           # ValidateAgentConfig
│   ├── routing.go         # Signal/routing validation
│   ├── helpers.go         # Common validation helpers
│   └── validation_test.go # Tests
│
├── /tool                  # RENAMED from tools/ (singular)
│   ├── execution.go       # Tool execution
│   ├── formatting.go      # Tool formatting
│   └── tool_test.go
│
├── /signal                # UNCHANGED
│   └── (existing signal routing)
│
├── /provider              # RENAMED from agents/ (generic name)
│   ├── openai/
│   ├── ollama/
│   └── interface.go       # Provider interface
│
├── /common                # NEW: Shared utilities
│   ├── types.go           # Core types (types.go + config_types.go + agent_types.go)
│   ├── constants.go       # Constants (execution_constants.go)
│   ├── errors.go          # Error definitions
│   ├── helpers.go         # Utility functions
│   └── common_test.go
│
├── /metrics               # OPTIONAL: Metrics sub-package
│   ├── collector.go       # MetricsCollector
│   ├── exporters.go       # JSON, Prometheus exporters
│   └── metrics_test.go
│
├── /internal              # UNCHANGED
│   └── (existing internal helpers)
│
└── types.go               # LEGACY (deprecated, use /common/types.go)
```

### 2.2 Dependency Diagram (New)

```
                    /common/types.go
                           |
            +------+--------+--------+------+
            |      |        |        |      |
        /config  /agent  /workflow /tool /signal
            |      |        |        |      |
    +-------+------+--------+--------+------+
    |                                      |
  /validation                         /executor
    |                                      |
    +------+----------+-------------------+
           |
        Application/CLI
```

**Kejelasan:**
- `/common` adalah base layer (no dependencies on anything else)
- `/config`, `/validation`, `/agent`, `/tool` adalah independent modules (low coupling)
- `/workflow` coordinated handlers (medium coupling)
- `/executor` adalah top-level orchestrator (high coupling - by design)

---

## 3. MIGRATION STRATEGY

### Phase 1: Create New Package Structure (Week 1)
**Goal**: Set up new directories, create minimal files, zero breaking changes

**Step 1.1**: Create new directories
```bash
mkdir -p core/executor core/agent core/workflow core/config core/validation core/tool core/common core/metrics
```

**Step 1.2**: Move/create base types
```go
// core/common/types.go
// Consolidate: types.go + config_types.go + agent_types.go
type Agent struct { ... }
type CrewConfig struct { ... }
type Task struct { ... }
// etc
```

**Step 1.3**: Move constants & errors
```go
// core/common/constants.go
const DefaultMaxParallelAgents = 10
const DefaultHistoryTrimThreshold = 10000

// core/common/errors.go
type ValidationError struct { ... }
type ExecutionError struct { ... }
```

### Phase 2: Extract Validation Layer (Week 2)
**Goal**: Decouple validation from config loading

**Step 2.1**: Create validation package
```go
// core/validation/crew.go
func ValidateCrewConfig(cfg *config.CrewConfig) error { ... }

// core/validation/agent.go
func ValidateAgentConfig(cfg *config.AgentConfig) error { ... }

// core/validation/routing.go
func ValidateSignals(signals map[string]interface{}) error { ... }

// core/validation/helpers.go
func validateDuration(d time.Duration) error { ... }
func contains(arr []string, s string) bool { ... }
```

**Step 2.2**: Update config_loader.go to use validation package
```go
// core/config/loader.go
func LoadCrewConfig(path string) (*CrewConfig, error) {
    cfg := &CrewConfig{}
    // load YAML...
    if err := validation.ValidateCrewConfig(cfg); err != nil {
        return nil, err
    }
    return cfg, nil
}
```

**Before**: validation.go (900 lines) → **After**: 4 focused files

### Phase 3: Modularize Config Loading (Week 2)
**Goal**: Separate loader, converter, and types

**Step 3.1**: Reorganize config package
```go
// core/config/types.go - PURE TYPES (no logic)
type CrewConfig struct { ... }
type AgentConfig struct { ... }

// core/config/loader.go - LOADING ONLY
func LoadCrewConfig(path string) (*CrewConfig, error) { ... }
func LoadAgentConfigs(agents map[string]*AgentConfig) error { ... }

// core/config/converter.go - CONVERSION ONLY
func ConfigToHardcodedDefaults(cfg *AgentConfig) *HardcodedDefaults { ... }
func BuildAgentMetadata(cfg *AgentConfig) *AgentMetadata { ... }
```

### Phase 4: Refactor Agent Execution (Week 3)
**Goal**: Split agent execution into focused modules

**Step 4.1**: Create agent package
```go
// core/agent/execution.go (FROM agent_execution.go)
func (a *Agent) Execute(ctx context.Context, msg interface{}) (interface{}, error) { ... }
func (a *Agent) ExecuteStream(ctx context.Context, msg interface{}) (chan interface{}, error) { ... }

// core/agent/provisioning.go
func CreateAgentFromConfig(cfg *config.AgentConfig) (*Agent, error) { ... }
func (a *Agent) SetupTooling(tools []interface{}) error { ... }

// core/agent/cost.go (FROM agent_cost.go)
func (a *Agent) EstimateTokens(msg string) int { ... }
func (a *Agent) CalculateCost(tokens int) float64 { ... }

// core/agent/messaging.go (NEW)
func (a *Agent) BuildSystemPrompt() string { ... }
func (a *Agent) ConvertToProviderMessages(msgs []interface{}) interface{} { ... }
```

### Phase 5: Reorganize Workflow Handlers (Week 3)
**Goal**: Reduce team_execution.go complexity

**Step 5.1**: Create workflow package
```go
// core/workflow/handler.go
type OutputHandler interface {
    SendStreamEvent(data interface{})
    HandleError(err error)
}

type SyncHandler struct { ... }
type StreamHandler struct { ... }

// core/workflow/execution.go
func executeAgentWithMetrics(ctx context.Context, agent *Agent, ...) error { ... }
func handleToolExecution(ctx context.Context, ...) error { ... }
func determineNextTransition(lastOutput interface{}, ...) *Agent { ... }

// core/workflow/routing.go (FROM team_routing.go)
func (ce *CrewExecutor) executeSignalBasedRouting(...) (*Agent, error) { ... }

// core/workflow/parallel.go (FROM team_parallel.go)
func (ce *CrewExecutor) executeParallelStream(...) error { ... }
```

### Phase 6: Reorganize Tool Execution (Week 4)
**Goal**: Clean tool handling

**Step 6.1**: Create tool package
```go
// core/tool/execution.go (FROM tool_execution.go)
func (ce *CrewExecutor) ExecuteCalls(ctx context.Context, ...) error { ... }
func safeExecuteTool(ctx context.Context, tool Tool, ...) interface{} { ... }

// core/tool/formatting.go
func formatToolResults(results interface{}, format string) string { ... }
func getToolParameterNames(tool Tool) []string { ... }
```

### Phase 7: Extract CrewExecutor (Week 4)
**Goal**: Split crew.go into focused modules

**Step 7.1**: Create executor package
```go
// core/executor/executor.go
type CrewExecutor struct {
    Crew           *Crew
    Agents         map[string]*Agent
    SignalRegistry signal.Registry
    Metrics        *metrics.Collector
    History        *HistoryManager
    // Remove: tool logic, validation logic, config logic
}

func NewCrewExecutor(crew *Crew) (*CrewExecutor, error) { ... }
func (ce *CrewExecutor) Execute(ctx context.Context, input interface{}) (interface{}, error) { ... }
func (ce *CrewExecutor) ExecuteStream(ctx context.Context, input interface{}) (chan interface{}, error) { ... }

// core/executor/workflow.go
func (ce *CrewExecutor) executeWorkflow(ctx context.Context, ...) error { ... }
func (ce *CrewExecutor) executeAgentWithContext(ctx context.Context, ...) error { ... }

// core/executor/history.go
func (ce *CrewExecutor) appendMessage(msg interface{}) { ... }
func (ce *CrewExecutor) trimHistoryIfNeeded() { ... }
func (ce *CrewExecutor) GetHistory() []interface{} { ... }
```

---

## 4. FILE MAPPING: OLD → NEW

### Types & Constants
- `types.go` → `common/types.go`
- `config_types.go` → `config/types.go` (merged with common/types.go)
- `agent_types.go` → `common/types.go` (merged)
- `execution_constants.go` → `common/constants.go`
- `tools/errors.go` → `common/errors.go`
- `tools/timeout.go` → `common/helpers.go`

### Configuration
- `config_loader.go` → `config/loader.go` (pure loading logic only)
- `defaults.go` → `config/defaults.go` or `config/converter.go`

### Validation
- `validation.go` → `validation/crew.go`, `validation/agent.go`, `validation/routing.go`, `validation/helpers.go`

### Agent Execution
- `agent_execution.go` → `agent/execution.go` + `agent/messaging.go`
- `agent_cost.go` → `agent/cost.go`

### Tool Execution
- `tool_execution.go` → `tool/execution.go`
- Team tool formatting → `tool/formatting.go`

### Workflow Management
- `team_execution.go` → `workflow/handler.go` + `workflow/execution.go`
- `team_routing.go` → `workflow/routing.go`
- `team_parallel.go` → `workflow/parallel.go`
- `team_history.go` → `executor/history.go`
- `team_tools.go` → `tool/formatting.go`

### Executor
- `crew.go` → `executor/executor.go` + `executor/workflow.go` + `executor/history.go`

### Metrics
- `metrics.go` → `metrics/collector.go`
- Metadata logging → `metrics/logging.go` (optional)

---

## 5. DEPENDENCY REDUCTION TARGET

### Current State (Before)
```
crew.go imports:
  - types ✓
  - config_types ✓
  - agent_types ✓
  - signal ✓
  - metrics ✓
  - tool execution ✓
  - agent execution ✓
  - team_execution ✓
  - team_routing ✓
  - team_parallel ✓
  - team_history ✓
  - team_tools ✓
  - validation ✓
  - config_loader ✓
  - defaults ✓
  Total: 15 imports → Coupling: 85/100
```

### Target State (After)
```
executor/executor.go imports:
  - common/types ✓
  - config/types ✓
  - signal ✓
  - executor/history ✓
  - metrics/collector ✓
  - workflow/handler ✓
  - workflow/execution ✓
  Total: 7 imports → Coupling: 50/100

executor/workflow.go imports:
  - common/types ✓
  - agent/Agent interface ✓
  - workflow/handler ✓
  - metrics/collector ✓
  Total: 4 imports

agent/execution.go imports:
  - common/types ✓
  - config/types ✓
  - provider (interface) ✓
  - agent/cost ✓
  - metrics/collector ✓
  Total: 5 imports
```

**Reduction**: 15 → 7 (52% reduction in crew executor coupling)

---

## 6. IMPLEMENTATION MILESTONES

### Milestone 1: Foundation (Week 1)
- [ ] Create /core/common package
- [ ] Create /core/config package with types
- [ ] Create /core/validation package
- [ ] Move types.go → common/types.go
- [ ] Update all imports (use sed/find & replace)
- [ ] Run tests (should all pass - no logic changes)

### Milestone 2: Configuration & Validation (Week 2)
- [ ] Extract validation logic into validation/* files
- [ ] Refactor config_loader.go (loader only)
- [ ] Create config/converter.go (ConfigToHardcodedDefaults)
- [ ] Run tests (validation tests should work)

### Milestone 3: Agent & Tool Modules (Week 3)
- [ ] Create /core/agent package with execution logic
- [ ] Create /core/tool package with execution logic
- [ ] Extract agent_execution.go → agent/execution.go
- [ ] Extract tool_execution.go → tool/execution.go
- [ ] Run tests

### Milestone 4: Workflow & Executor (Week 4)
- [ ] Create /core/workflow package
- [ ] Create /core/executor package
- [ ] Refactor team_execution.go logic
- [ ] Refactor crew.go → executor/executor.go
- [ ] Run all tests
- [ ] Performance check (should be same)

### Milestone 5: Cleanup & Optimization (Week 5)
- [ ] Remove old files (after migration)
- [ ] Update examples to use new imports
- [ ] Update documentation
- [ ] Run full test suite
- [ ] Benchmark tests

---

## 7. RISKS & MITIGATION

### Risk 1: Breaking Changes
**Risk**: Old import paths stop working
**Mitigation**:
- Use go mod replace during migration
- Keep backwards compatibility shims for 1-2 releases
- Update all internal tests before external migration

### Risk 2: Circular Dependencies
**Risk**: New structure introduces cycles
**Mitigation**:
- Use dependency analysis tools (go mod graph)
- Test each phase independently
- Validate: no cycles allowed

### Risk 3: Performance Regression
**Risk**: More files/packages = slower compilation/execution
**Mitigation**:
- Benchmark before & after
- Use build caching
- Profile memory usage

### Risk 4: Test Failures
**Risk**: Import changes break tests
**Mitigation**:
- Update test imports immediately
- Keep test coverage >80%
- Run test suite after each phase

### Risk 5: Team Productivity
**Risk**: Developers confused by new structure
**Mitigation**:
- Document new structure clearly
- Provide import migration guide
- Pair programming during migration
- Train on new patterns

---

## 8. SUCCESS CRITERIA

- ✅ All tests pass
- ✅ No circular dependencies
- ✅ crew.go coupling reduced from 85 → 60
- ✅ Each package has single responsibility
- ✅ Documentation updated
- ✅ Build time same or faster
- ✅ No performance regression
- ✅ Code coverage maintained >80%

---

## 9. BACKWARDS COMPATIBILITY

### Option A: Hard Break (Recommended)
- Remove old files completely
- Update all imports in codebase
- Update examples + docs
- **Pros**: Clean, no legacy code
- **Cons**: Requires coordinated update

### Option B: Soft Break (Gradual)
- Keep old files as stubs calling new packages
- Old imports still work but deprecated
- Plan removal in v2.0
- **Pros**: Existing code works
- **Cons**: Technical debt, confusion

### Recommendation
Use **Option A** if this is internal refactoring.
Use **Option B** if this is a public library with external users.

---

## 10. NEXT STEPS

### Immediate (Today)
- [ ] Review this plan
- [ ] Validate proposed structure with team
- [ ] Identify blockers

### Short Term (This Week)
- [ ] Create new package structure (empty files)
- [ ] Prepare migration script for imports
- [ ] Setup CI/CD to validate new structure

### Medium Term (Next 2-4 Weeks)
- [ ] Execute Phase 1-5 migration
- [ ] Update documentation
- [ ] Run performance tests

### Long Term
- [ ] Monitor metrics
- [ ] Gather feedback
- [ ] Plan Phase 2 improvements (microservices, plugins, etc.)

---

## 11. COMPARISON: OLD vs NEW ARCHITECTURE

| Aspect | OLD | NEW |
|--------|-----|-----|
| # of top-level files in core | 39 | 9 (organized in packages) |
| crew.go size | 1500+ lines | 400-500 lines |
| validation.go size | 900+ lines | 4 files × 200-250 lines |
| Coupling (crew.go) | 85/100 | 50-60/100 |
| Testability | Hard (many dependencies) | Easy (isolated modules) |
| Navigation | Flat structure | Organized by feature |
| Reusability | Low (tight coupling) | High (independent modules) |
| Time to understand | High (many interdependencies) | Low (clear module boundaries) |

---

## 12. ADDITIONAL IMPROVEMENTS (Post-Migration)

After architectural refactoring, consider:

1. **Interface-Based Design**
   - Define interfaces for each package's public API
   - Use dependency injection
   - Easier mocking for tests

2. **Plugin System**
   - Make provider/llm pluggable (already done)
   - Make validation rules pluggable
   - Make metrics exporters pluggable

3. **Middleware Pattern**
   - Add middleware chain for execution hooks
   - Pre/post execution handlers
   - Logging, metrics, security plugins

4. **Configuration Management**
   - Support multiple config formats (YAML, JSON, ENV)
   - Configuration profiles (dev, staging, prod)
   - Hot reload support

---

**Created**: 2025-12-25
**Last Updated**: 2025-12-25
**Status**: Ready for Review & Approval

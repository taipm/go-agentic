# 🎯 DETAILED ACTION PLAN: Dead Code & Duplicate Code Cleanup

**Project:** go-agentic
**Scope:** ./core/ directory (~22,614 LOC, 54 files)
**Objective:** Eliminate dead code and consolidate duplicate code patterns
**Target Effort:** 5-7 substantial refactoring tasks

---

## 📊 EXECUTIVE SUMMARY

| Category | Issues Found | Severity | Est. Impact |
|----------|-------------|----------|------------|
| **Duplicate Tool Parsing** | 1 major | 🔴 HIGH | 50+ LOC elimination |
| **Executor Overlap** | 1 significant | 🟡 MEDIUM | Architecture clarity |
| **Unimplemented TODOs** | 2 blockers | 🟡 MEDIUM | Feature completeness |
| **Large Files** | 2 refactoring targets | 🟡 MEDIUM | Code maintainability |
| **Legacy Code** | 3 items | 🟢 LOW | Cleanup opportunity |

---

## 🔴 PHASE 1: HIGH PRIORITY - DUPLICATE TOOL PARSING

### Issue 1.1: Consolidated Tool Argument Parsing

**Severity:** HIGH
**Files Affected:**
- `core/providers/ollama/provider.go` (lines 367-420) ❌ DUPLICATE
- `core/tools/arguments.go` ✅ SOURCE OF TRUTH
- `core/providers/openai/provider.go` (lines 446-447) ✅ CORRECT APPROACH

**Current State:**
```
ollama/provider.go has FULL implementation (54 lines):
  parseToolArguments()        [50+ lines of logic]
  splitArguments()            [delegator - delegates to tools]
  isAlphanumeric()            [delegator - delegates to tools]

tools/arguments.go has FULL implementation (90 lines):
  ParseArguments()            [shared, correct]
  SplitArguments()            [shared, correct]
  IsAlphanumeric()            [shared, correct]

openai/provider.go CORRECTLY uses:
  parseToolArguments() → tools.ParseArguments()
```

**Problem:**
- Ollama parseToolArguments (367-420) is 54 lines of duplicated logic
- Different from tools.ParseArguments() in behavior (ollama has key=value parsing)
- openai/provider.go correctly delegates to tools.ParseArguments()
- Creates maintenance burden if one changes but not the other

**Solution:**
1. **Compare implementations** - Determine which approach is correct
2. **Unify** - Both providers should use tools.ParseArguments()
3. **Enhance tools.ParseArguments()** if needed to support both formats
4. **Remove duplicate** from ollama/provider.go

**Action Steps:**

```bash
# Step 1: Analyze behavioral differences
# Current: tools.ParseArguments() uses simple JSON + positional parsing
# Current: ollama parseToolArguments() uses JSON + key=value + positional
# Decision: Ollama has MORE features, so enhance tools.ParseArguments()

# Step 2: Update tools/arguments.go to support key=value parsing
# - Enhance ParseArguments() to detect and handle "key=value" format
# - Keep JSON parsing first
# - Then try key=value
# - Then fallback to positional

# Step 3: Update ollama/provider.go
# - Replace parseToolArguments() with delegation to tools.ParseArguments()
# - Delete lines 367-420 (54 lines)
# - Add: tools.ParseArguments(argsStr)

# Step 4: Verify openai/provider.go already correct
# - Confirm it uses tools.ParseArguments()
# - No changes needed
```

**Files to Modify:**
- `core/tools/arguments.go` - ADD key=value parsing capability
- `core/providers/ollama/provider.go` - REMOVE lines 367-420, delegate to tools

**Testing:**
- Verify both providers handle:
  - JSON arguments: `{key: value}`
  - Key=value arguments: `question_number=1, question="Q"`
  - Positional arguments: `arg1, arg2, arg3`
  - Mixed types: numbers, strings, booleans

**Expected Outcome:**
- ✅ Remove 54 lines of duplicate code from ollama
- ✅ Single source of truth for argument parsing
- ✅ Both providers use consistent logic

---

### Issue 1.2: Tool Call Extraction Methods

**Severity:** HIGH
**Files Affected:**
- `core/providers/ollama/provider.go` - Has extractToolCallsFromText()
- `core/providers/openai/provider.go` - Has extractToolCallsFromResponse() + extractToolCallsFromText()

**Current State:**
```
ollama/provider.go (lines 268-365):
  extractToolCallsFromText() [98 lines]
    └─ parseToolArguments()   [delegates to shared]
    └─ splitArguments()       [delegates to shared]
    └─ isAlphanumeric()       [delegates to shared]

openai/provider.go:
  extractToolCallsFromResponse() [specific to OpenAI format]
  extractToolCallsFromText() [similar to ollama but different format handling]
```

**Problem:**
- Each provider has custom extraction logic, hard to maintain
- Some delegation to tools package, some not
- Different parsing strategies between ollama and openai
- Makes adding new providers costly

**Solution:**
1. **Extract common pattern** to tools package
2. **Create abstract tool extraction** that both providers can use
3. **Keep provider-specific logic** for format differences only

**Action Steps:**

```bash
# Step 1: Create tools/extraction.go
# - New file for tool extraction utilities
# - Implement ExtractToolCalls(text, format) function
# - Support both "tool:" prefix format (ollama) and other formats

# Step 2: Refactor ollama/provider.go
# - Replace extractToolCallsFromText() with call to tools.ExtractToolCalls()
# - Keep provider-specific format handling if needed

# Step 3: Refactor openai/provider.go
# - Use tools.ExtractToolCalls() for text extraction
# - Keep extractToolCallsFromResponse() for OpenAI-specific format

# Step 4: Document extraction format
# - Define supported tool call formats
# - Document provider-specific variants
```

**Files to Modify:**
- `core/tools/extraction.go` - NEW FILE
- `core/providers/ollama/provider.go` - REFACTOR extractToolCallsFromText()
- `core/providers/openai/provider.go` - REFACTOR to use common utilities

**Expected Outcome:**
- ✅ Unified tool extraction logic
- ✅ Easier to add new providers
- ✅ Clearer separation of concerns

---

## 🟡 PHASE 2: MEDIUM PRIORITY - UNIMPLEMENTED TODOS

### Issue 2.1: Missing Agent Routing in Signal-based Execution

**Severity:** MEDIUM
**Files Affected:**
- `core/workflow/execution.go` (lines 193-195, 218-220) ⚠️ TODO MARKER

**Current State:**
```go
// Lines 193-195
// TODO: Look up next agent by ID and continue execution
// For now, return (will be implemented in crew.go integration)
return response, nil

// Lines 218-220
// TODO: Look up next agent by ID and continue execution
// For now, return response
return response, nil
```

**Problem:**
- Signal-based routing returns early instead of continuing execution
- Agent handoff doesn't work with signal routing
- Feature incomplete - signals detected but not acted upon
- Blocks: Signal routing, agent handoffs, multi-agent workflows

**Solution:**
1. **Implement agent lookup** by ID
2. **Continue execution** with next agent
3. **Update history** with handoff event

**Action Steps:**

```bash
# Step 1: Add agent lookup helper to workflow/execution.go
# Function signature:
# func (e *ExecutionContext) getAgentByID(agentID string) (*Agent, error)
# - Search crew.Agents for matching ID
# - Return error if not found
# - Log lookup failures

# Step 2: Implement signal-based handoff (lines 193-195)
# - Look up nextAgent using getAgentByID()
# - Append routing signal to history
# - Update CurrentAgent
# - Continue loop (don't return, continue iteration)

# Step 3: Implement target-based handoff (lines 218-220)
# - Same as above
# - Use HandoffTargets[0].ID as nextAgentID

# Step 4: Handle errors
# - If agent not found: emit error signal, return error
# - If handoff count exceeded: emit signal, terminate
# - Log all handoff transitions

# Step 5: Test signal-based routing
# - Create test case with signal-routed workflow
# - Verify agent transitions happen correctly
# - Verify history captures all handoffs
```

**Files to Modify:**
- `core/workflow/execution.go` - IMPLEMENT agent lookup and routing

**Testing:**
- Signal-based agent routing
- Manual agent handoff via HandoffTargets
- Error handling for missing agents
- Handoff count limits

**Expected Outcome:**
- ✅ Signal-based routing fully implemented
- ✅ Agents properly hand off to next agents
- ✅ Complete multi-agent conversation flows

---

### Issue 2.2: Missing Tool Conversion in Agent Execution

**Severity:** MEDIUM
**Files Affected:**
- `core/agent/execution.go` (line 131) ⚠️ TODO MARKER

**Current State:**
```go
// Line 131
Tools: nil, // TODO: Implement proper tool conversion from agent.Tools
```

**Problem:**
- Agent tools not passed to provider
- Tools defined in agent config not available during execution
- Provider receives nil tools, can't execute tool calls
- Limits agent capability to only built-in provider features

**Solution:**
1. **Convert agent tools** to provider format
2. **Pass tools** to provider execution
3. **Handle tool resolution** at runtime

**Action Steps:**

```bash
# Step 1: Analyze tool format differences
# - Agent.Tools format (from config)
# - Provider tool format expectations
# - Create conversion function if needed

# Step 2: Implement tool conversion
# Function in agent/execution.go:
# func convertAgentTools(agentTools []ToolDefinition) []ProviderTool {
#     // Convert from agent config format to provider format
#     // Handle tool parameters, descriptions, etc.
# }

# Step 3: Update ExecuteWithProvider()
# - Call convertAgentTools(agent.Tools)
# - Pass converted tools to provider.Execute()
# - Ensure nil check: only pass if tools exist

# Step 4: Test tool execution
# - Agent with tools
# - Provider receives and can call tools
# - Tool responses integrated into chat flow

# Step 5: Validate tool parameters
# - Ensure required fields present
# - Type checking on parameters
# - Error messages for invalid tools
```

**Files to Modify:**
- `core/agent/execution.go` - IMPLEMENT tool conversion

**Testing:**
- Agent execution with tools
- Tool calling from provider
- Tool result integration
- Error handling for invalid tools

**Expected Outcome:**
- ✅ Agent tools properly passed to providers
- ✅ Tools available during execution
- ✅ Tool calls work end-to-end

---

## 🟡 PHASE 3: MEDIUM PRIORITY - LARGE FILE REFACTORING

### Issue 3.1: Break Down crew.go (602 LOC)

**Severity:** MEDIUM
**Files Affected:**
- `core/crew.go` (602 lines) - Too large for single file

**Current State:**
```
crew.go contains:
  - Type definition (CrewExecutor struct)
  - Constructors (2)
  - Configuration methods (4)
  - Signal management (2)
  - History management (8)
  - Metrics & logging (4)
  - Execution pipeline (3)
  - Helper functions (10+)
```

**Problem:**
- 602 lines in single file makes navigation difficult
- Mix of public API, internal logic, and helpers
- Harder to understand flow with everything in one file
- Harder to test individual concerns

**Solution:**
Extract into focused modules while keeping crew.go as orchestrator

**Action Steps:**

```bash
# Step 1: Create crew_signal_handlers.go (40 lines)
# Move from crew.go lines 200-260:
#   - ValidateSignals()
#   - RegisterSignalHandlers()
# Benefits:
#   - Self-contained signal concerns
#   - Reusable in other contexts
#   - Clear signal management API

# Step 2: Create crew_history.go (70 lines)
# Move from crew.go lines 264-330:
#   - appendMessage()
#   - getHistoryCopy()
#   - GetHistory()
#   - estimateHistoryTokens()
#   - trimHistoryIfNeeded()
#   - ClearHistory()
#   - addUserMessageToHistory()
#   - addAssistantMessageToHistory()
# Benefits:
#   - All history operations grouped
#   - Easier to maintain history logic
#   - Can extend history manager pattern

# Step 3: Create crew_metrics.go (50 lines)
# Move from crew.go lines 336-423:
#   - sendStreamEvent()
#   - handleAgentError()
#   - updateAgentMetrics()
#   - recordAgentExecution()
# Benefits:
#   - Separate observability concerns
#   - Could become observer pattern
#   - Cleaner error handling

# Step 4: Extract message conversion helpers
# Create crew_message.go (20 lines):
#   - messageToCommon(m Message) common.Message
#   - commonToMessage(m common.Message) Message
# Benefits:
#   - Consistent type conversions
#   - Single place to modify conversion logic
#   - Reusable in other packages

# Step 5: Keep crew.go as orchestrator (250 lines)
# - CrewExecutor struct definition
# - Constructors
# - Configuration methods
# - Public API (Execute, ExecuteStream)
# - executeWorkflow() core logic
# - References to extracted modules

# Step 6: Document module responsibilities
# - crew_signal_handlers.go - Signal routing validation
# - crew_history.go - Conversation history management
# - crew_metrics.go - Metrics collection and error handling
# - crew_message.go - Type conversion utilities
# - crew.go - Main orchestration and execution

# Step 7: Update imports in crew.go
# - Keep internal references
# - All refactored methods still accessible
# - Package-level visibility unchanged
```

**Files to Create:**
- `core/crew_signal_handlers.go` - Signal management
- `core/crew_history.go` - History operations
- `core/crew_metrics.go` - Metrics and error handling
- `core/crew_message.go` - Message conversion

**Files to Modify:**
- `core/crew.go` - Remove extracted code, reference extracted modules

**Expected Outcome:**
- ✅ crew.go reduced from 602 to ~250 lines
- ✅ Clear module boundaries
- ✅ Easier to navigate and maintain
- ✅ Better code organization

---

### Issue 3.2: Refactor NewCrewExecutorFromConfig (104 lines)

**Severity:** MEDIUM
**Files Affected:**
- `core/crew.go` (lines 71-174) - Large constructor

**Current State:**
```go
// Lines 71-174 (104 lines)
func NewCrewExecutorFromConfig(...) (*CrewExecutor, error) {
    // 1. Load crew config
    // 2. Load agent configs
    // 3. Build agent objects (12+ field assignments each)
    // 4. Build routing signals
    // 5. Create executor
    // 6. Validate signals
    // 7. Register handlers
}
```

**Problem:**
- 104 lines in single constructor is difficult to follow
- Nested field conversions hard to understand
- Logic could be reused elsewhere
- Makes testing harder

**Solution:**
Extract agent creation logic into separate functions

**Action Steps:**

```bash
# Step 1: Extract agent creation function
# File: core/crew_config.go (NEW)
# Function: createAgentFromConfig(agentObj *config.Agent, defaults ...) (*Agent, error)
# - Handles all field assignments
# - Type conversions
# - Default value application
# - Validation

# Step 2: Extract signal routing setup
# File: core/crew_config.go
# Function: setupSignalRouting(crew *Crew, config *config.CrewConfig) error
# - Parses signal definitions
# - Builds routing table
# - Validates routing targets

# Step 3: Refactor NewCrewExecutorFromConfig
# Simplify to:
#   1. Load crew config
#   2. Load agent configs
#   3. Create agents using createAgentFromConfig()
#   4. Setup routing using setupSignalRouting()
#   5. Create executor
#   6. Validate and register signals
# Result: ~30 lines of clear, orchestration logic

# Step 4: Extract validation logic
# File: core/crew_validation.go (NEW)
# Functions:
#   - validateCrewConfig(config *config.CrewConfig) error
#   - validateAgentConfig(agent *config.Agent) error
# - Centralized validation
# - Reusable validation logic

# Step 5: Unit test extracted functions
# - Test agent creation with various configs
# - Test signal routing setup
# - Test validation logic
```

**Files to Create:**
- `core/crew_config.go` - Agent and routing creation
- `core/crew_validation.go` - Configuration validation

**Files to Modify:**
- `core/crew.go` - Simplify NewCrewExecutorFromConfig

**Expected Outcome:**
- ✅ Constructor reduced from 104 to ~30 lines
- ✅ Reusable agent creation logic
- ✅ Better testability
- ✅ Clearer intent

---

## 🟢 PHASE 4: LOW PRIORITY - LEGACY CODE CLEANUP

### Issue 4.1: Remove Type Alias Backward Compatibility Wrapper

**Severity:** LOW
**Files Affected:**
- `core/types.go` - Entire file

**Current State:**
```go
// core/types.go - Full of type aliases and re-exports
type CrewExecutor = executor.CrewExecutor
type HistoryManager = executor.HistoryManager
type SignalRegistry = signal.SignalRegistry
// ... 20+ more aliases
```

**Problem:**
- Exists only for backward compatibility
- Adds indirection: types.go → executor package
- Makes code harder to follow (unclear where types come from)
- Not adding value if no external packages depend on it

**Decision:**
- **If no external packages use it:** Remove completely, use direct imports
- **If external packages depend on it:** Keep for compatibility

**Action Steps:**

```bash
# Step 1: Check external dependencies
# Search all Go files outside /core/ directory:
#   grep -r "types\." --include="*.go" | grep -v core/types.go | head -20
# If results: Keep types.go, it's providing important API
# If empty: It's internal only, can remove

# Step 2A: If keeping types.go
# - Add comment explaining backward compatibility purpose
# - Update CLAUDE.md to document this pattern
# - Mark as deprecated in comments if being phased out

# Step 2B: If removing types.go
# Update imports across core/:
#   - From: import "github.com/taipm/go-agentic/core"
#   - To:   import (
#             "github.com/taipm/go-agentic/core/executor"
#             "github.com/taipm/go-agentic/core/signal"
#           )
# - Remove core/types.go file
# - Update any code relying on these aliases

# Step 3: Verify compilation
# go build ./...
# go test ./...
```

**Files to Modify (if removing):**
- `core/types.go` - DELETE

**Files to Check:**
- All core/*.go files for internal usage of types.go
- All external packages for usage of core/types.go

**Expected Outcome:**
- ✅ Remove unnecessary indirection OR
- ✅ Document backward compatibility layer

---

### Issue 4.2: Consolidate Token Calculation Functions

**Severity:** LOW
**Files Affected:**
- `core/crew.go` (line 394-396) - calculateMessageTokens()
- `core/common/types.go` (lines 526-540) - Agent.estimateTokens() and Agent.EstimateTokens()

**Current State:**
```go
// crew.go lines 394-396
func calculateMessageTokens(msg common.Message) int {
    return 1 + (len(msg.Content) + 50) / 4
}

// common/types.go lines 526-534
func (a *Agent) estimateTokens(input string) int {
    return len(input) / 4
}

func (a *Agent) EstimateTokens(input string) int {
    return a.estimateTokens(input)
}
```

**Problem:**
- Three similar functions doing token estimation
- Different calculation methods (inconsistent)
- Formula unclear (magic numbers: 1, 50, 4)
- Potential source of bugs if calculation changes

**Solution:**
1. **Centralize token calculation** in tools or common package
2. **Use consistent formula** everywhere
3. **Document assumptions**

**Action Steps:**

```bash
# Step 1: Create tools/tokens.go (NEW FILE)
# Define constant and function:
#   const TokenEstimationDivisor = 4
#   const TokenEstimationBaseOverhead = 1
#   const TokenEstimationPadding = 50
#
# func EstimateMessageTokens(content string) int {
#     return TokenEstimationBaseOverhead + (len(content) + TokenEstimationPadding) / TokenEstimationDivisor
# }
#
# func EstimateTokens(input string) int {
#     return len(input) / TokenEstimationDivisor
# }

# Step 2: Document token estimation strategy
# - Add comment explaining assumptions
# - Link to how tokens are actually counted
# - Document error margin

# Step 3: Update crew.go
# Replace calculateMessageTokens() call with tools.EstimateMessageTokens()

# Step 4: Update common/types.go
# Replace Agent.estimateTokens() to call tools.EstimateTokens()

# Step 5: Remove duplicates
# - Delete calculateMessageTokens() from crew.go
# - Keep Agent.EstimateTokens() as public API but delegate

# Step 6: Add test cases
# - Test token calculation consistency
# - Document token estimation accuracy
```

**Files to Create:**
- `core/tools/tokens.go` - Token estimation utilities

**Files to Modify:**
- `core/crew.go` - Use tools.EstimateMessageTokens()
- `core/common/types.go` - Delegate Agent.estimateTokens() to tools

**Expected Outcome:**
- ✅ Single source of truth for token calculation
- ✅ Consistent results across codebase
- ✅ Easier to adjust formula in one place

---

### Issue 4.3: Remove or Document Deprecated Fields

**Severity:** LOW
**Files Affected:**
- `core/common/types.go` (lines 287, 293-294, 448-450)

**Current State:**
```go
// CrewConfig fields (lines 287, 293-294)
Model string           // Deprecated: Use Primary instead
Provider string        // Deprecated: Use Primary.Provider instead
ProviderURL string     // Deprecated: Use Primary.ProviderURL instead

// Agent fields (lines 448-450)
Model string           // Deprecated: Use LLMProvider.Model instead
Provider string        // Deprecated: Use LLMProvider.Provider instead
ProviderURL string     // Deprecated: Use LLMProvider.ProviderURL instead
```

**Problem:**
- Deprecated fields still in struct but shouldn't be used
- No migration path documented
- Code using old fields unclear if safe
- Add confusion for new developers

**Solution:**
Document deprecation path or remove fields entirely

**Action Steps:**

```bash
# Step 1: Assess impact
# Find all usages of deprecated fields:
#   grep -r "\.Model" --include="*.go" core/ | grep -v "LLMProvider"
#   grep -r "\.Provider" --include="*.go" core/ | grep -v "LLMProvider\|Primary"
#   grep -r "\.ProviderURL" --include="*.go" core/ | grep -v "LLMProvider\|Primary"

# Step 2A: If no usages found
# Option 1: Remove deprecated fields entirely
#   - Delete lines with deprecated fields
#   - Update json tags
#   - Update documentation
#
# Option 2: Keep but add migration guide
#   - Add comment with deprecation notice
#   - Document how to migrate to new fields
#   - Link to migration PR/docs

# Step 2B: If usages found
# Create migration plan:
#   1. Update code to use new fields
#   2. Add compatibility layer if needed
#   3. Mark for removal in next major version
#   4. Document removal timeline

# Step 3: Update documentation
# - CHANGELOG.md: Document deprecation
# - migration guide: How to update existing configs
# - code comments: Why field was deprecated

# Step 4: Consider YAML config compatibility
# - Ensure old YAML configs still load
# - Auto-convert old format to new format
# - Log deprecation warnings when old format used

# Step 5: Tag version
# - Mark deprecated in code with build tag
# - Plan removal for next major version
```

**Files to Modify:**
- `core/common/types.go` - Remove or document deprecated fields
- Documentation - Add migration guide

**Expected Outcome:**
- ✅ Clear deprecation path
- ✅ No confusion about field usage
- ✅ Documented migration strategy

---

## 📈 IMPLEMENTATION ROADMAP

### Week 1: HIGH PRIORITY (Phase 1)
**Goal:** Eliminate duplicate code in tool parsing

| Day | Task | Est. Time |
|-----|------|-----------|
| Mon | Analyze parseToolArguments() differences | 2h |
| Tue | Enhance tools.ParseArguments() with key=value support | 3h |
| Wed | Update ollama/provider.go to use enhanced ParseArguments() | 2h |
| Wed | Test both providers (ollama + openai) | 3h |
| Thu | Extract tool extraction to tools package | 4h |
| Fri | Test tool call extraction across providers | 3h |

**Deliverables:**
- ✅ Unified tool argument parsing
- ✅ Common tool extraction utilities
- ✅ All tests passing
- ✅ ~100 LOC removed

---

### Week 2: MEDIUM PRIORITY - Part A (Phase 2)
**Goal:** Implement missing TODO functionality

| Day | Task | Est. Time |
|-----|------|-----------|
| Mon | Implement agent routing in workflow/execution.go | 4h |
| Tue | Implement tool conversion in agent/execution.go | 3h |
| Wed-Fri | Test signal-based routing and tool execution | 6h |

**Deliverables:**
- ✅ Signal-based agent routing working
- ✅ Tool conversion implemented
- ✅ Integration tests passing
- ✅ Documentation updated

---

### Week 3: MEDIUM PRIORITY - Part B (Phase 3)
**Goal:** Refactor large files into focused modules

| Day | Task | Est. Time |
|-----|------|-----------|
| Mon | Extract signal handlers (crew_signal_handlers.go) | 2h |
| Tue | Extract history operations (crew_history.go) | 2h |
| Wed | Extract metrics and error handling (crew_metrics.go) | 2h |
| Thu | Extract message utilities (crew_message.go) | 1h |
| Fri | Verify refactored crew.go and test | 3h |

**Deliverables:**
- ✅ crew.go refactored into 5 focused files
- ✅ All internal tests passing
- ✅ API unchanged (backward compatible)
- ✅ Code easier to navigate

---

### Week 4: MEDIUM-LOW PRIORITY (Phase 3 cont. & Phase 4)
**Goal:** Clean up configuration and legacy code

| Day | Task | Est. Time |
|-----|------|-----------|
| Mon | Extract agent creation (crew_config.go) | 3h |
| Tue | Simplify NewCrewExecutorFromConfig() | 2h |
| Wed | Consolidate token calculation functions | 2h |
| Thu | Document/remove deprecated fields | 1h |
| Fri | Full integration testing and cleanup | 3h |

**Deliverables:**
- ✅ Configuration logic extracted and reusable
- ✅ Token calculation centralized
- ✅ Deprecated fields handled
- ✅ All tests passing

---

## 🧪 TESTING STRATEGY

### Unit Tests Required
```bash
# Tools package
✓ tools/arguments_test.go - ParseArguments() with all formats
✓ tools/extraction_test.go - Tool extraction across formats
✓ tools/tokens_test.go - Token calculation consistency

# Providers
✓ providers/ollama/provider_test.go - Updated to use common tools
✓ providers/openai/provider_test.go - Verify parsing still works

# Workflow
✓ workflow/execution_test.go - Signal-based routing
✓ workflow/signal_test.go - Agent handoff

# Agent
✓ agent/execution_test.go - Tool conversion

# Crew
✓ crew_test.go - All crew.go methods
✓ crew_signal_handlers_test.go - Signal validation
✓ crew_history_test.go - History operations
✓ crew_metrics_test.go - Metrics collection
✓ crew_config_test.go - Configuration setup
```

### Integration Tests Required
```bash
✓ Full multi-agent workflow with signals
✓ Agent handoff with tools
✓ History tracking across agents
✓ Metrics collection across workflow
✓ Error handling and recovery
✓ YAML config loading and validation
```

### Regression Tests
```bash
✓ Existing test suite continues to pass
✓ Public API unchanged
✓ Backward compatibility preserved
✓ Performance characteristics maintained
```

---

## ✅ SUCCESS CRITERIA

### Code Quality
- [ ] All duplicated code removed or consolidated
- [ ] No dead code remains
- [ ] No unimplemented TODOs with actual logic
- [ ] All files < 400 LOC (except common/types.go which is by design)
- [ ] Clear module responsibilities

### Test Coverage
- [ ] All new functions tested
- [ ] Integration tests for signal-based routing
- [ ] All existing tests still pass
- [ ] Code coverage >= 80%

### Documentation
- [ ] Public API documented
- [ ] Token calculation strategy documented
- [ ] Deprecation path documented
- [ ] Migration guide for any breaking changes

### Performance
- [ ] No performance regression
- [ ] Token calculation consistent
- [ ] History trimming still works
- [ ] Signal routing doesn't add significant overhead

---

## 🎯 RISK MITIGATION

### Risk 1: Breaking Changes in Public API
**Mitigation:**
- Keep crew.go public API unchanged
- Extract private methods only
- Document any changes in CHANGELOG
- Maintain backward compatibility layer for types.go

### Risk 2: Test Coverage Gaps
**Mitigation:**
- Run full test suite before and after each phase
- Add integration tests for signal routing
- Profile tool argument parsing across formats
- Test with real YAML configs

### Risk 3: Performance Regression
**Mitigation:**
- Benchmark token calculation before/after
- Profile signal routing overhead
- Monitor history trimming performance
- Load test with large conversation histories

### Risk 4: Tool Extraction Complexity
**Mitigation:**
- Start with simple extraction utilities
- Incrementally enhance ParseArguments()
- Keep provider-specific logic where needed
- Add comprehensive tests for edge cases

---

## 📝 APPENDIX: FILE CHECKLIST

### New Files to Create
- [ ] `core/tools/extraction.go` - Tool extraction utilities
- [ ] `core/tools/tokens.go` - Token estimation
- [ ] `core/crew_signal_handlers.go` - Signal management
- [ ] `core/crew_history.go` - History operations
- [ ] `core/crew_metrics.go` - Metrics and error handling
- [ ] `core/crew_message.go` - Message conversion
- [ ] `core/crew_config.go` - Configuration setup
- [ ] `core/crew_validation.go` - Configuration validation

### Files to Modify
- [ ] `core/tools/arguments.go` - Add key=value parsing
- [ ] `core/providers/ollama/provider.go` - Remove 54 LOC duplicate
- [ ] `core/providers/openai/provider.go` - Verify correct usage
- [ ] `core/workflow/execution.go` - Implement agent routing (TODO)
- [ ] `core/agent/execution.go` - Implement tool conversion (TODO)
- [ ] `core/crew.go` - Extract and refactor
- [ ] `core/common/types.go` - Consolidate token functions, handle deprecation
- [ ] Test files - Add new tests for extracted functions

### Files to Delete (Conditional)
- [ ] `core/types.go` - Only if no external dependencies

---

## 📞 QUESTIONS FOR TEAM

1. **Tool parsing:** Should both ollama and openai support key=value format?
2. **Agent routing:** How should signal routing integrate with crew execution?
3. **External dependencies:** Does anything outside /core depend on core/types.go?
4. **Token formula:** Is the 1 + (len + 50) / 4 formula correct? Where does it come from?
5. **Deprecation timeline:** When should deprecated fields be removed?
6. **Breaking changes:** Is it acceptable to refactor crew.go structure?

---

**Document Version:** 1.0
**Created:** 2025-12-25
**Status:** Ready for Implementation
**Next Step:** Review & Team Approval

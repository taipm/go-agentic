# DX Improvement Roadmap: go-agentic

**Goal:** Improve Developer Experience from **6.5/10 → 8.5+/10**

Based on analysis of Anthropic SDK, LangChain, FastAPI, and best practices in tool definition frameworks.

---

## Executive Summary

### Current State (6.5/10)
- Tool registration scattered across multiple steps
- Manual type assertions & validation boilerplate
- Silent failures on tool registration mismatch
- Error handling not propagated to LLM
- 40+ LOC per tool in multi-tool examples
- Confusing dual routing system (signals + tools)

### Target State (8.5+/10)
- Single, clear tool registration method
- Auto-generated schemas from Go structs
- Fail-fast validation at load time
- Clear error messages
- 10-15 LOC per tool
- Unified tool-based routing

### Key Changes
```
BEFORE: RegisterTool() + Tool struct + JSON schema + YAML reference
AFTER:  Struct params + Function + ToolRegistry.Add() + Framework handles everything
```

---

## Phase 1: Struct-Based Parameters (Week 1-2)

**Goal:** Enable type-safe parameters instead of `map[string]interface{}`

### Task 1.1: Create Parameter Schema Generator

**File:** `core/tools/schema_generator.go` (NEW)

```go
package tools

import (
    "reflect"
    "encoding/json"
)

// ParameterSchema represents JSON schema for tool parameters
type ParameterSchema struct {
    Type       string                 `json:"type"`
    Properties map[string]interface{} `json:"properties"`
    Required   []string               `json:"required"`
}

// GenerateSchemaFromStruct creates JSON schema from Go struct
// Leverages struct tags: json, description, enum, minimum, maximum, etc.
func GenerateSchemaFromStruct(v interface{}) (*ParameterSchema, error)

// Example usage:
// type GetStatusParams struct {
//     ID string `json:"id" description:"Status ID" minLength:"1"`
// }
// schema, _ := GenerateSchemaFromStruct(GetStatusParams{})
```

**Acceptance Criteria:**
- [ ] Generate schema from struct with json tags
- [ ] Support field descriptions from struct tags
- [ ] Support validation rules (minLength, maxLength, minimum, maximum, enum)
- [ ] Handle nested structs
- [ ] Unit tests pass

**Effort:** 8-12 hours

---

### Task 1.2: Create Structured Tool Wrapper

**File:** `core/tools/structured_tool.go` (NEW)

```go
package tools

import (
    "context"
    "reflect"
)

// StructuredTool wraps a function with struct-based parameters
type StructuredTool struct {
    name        string
    description string
    fn          interface{}
    paramType   reflect.Type
    schema      *ParameterSchema
}

// NewStructuredTool creates tool from typed function
// func GetStatus(ctx context.Context, params GetStatusParams) (string, error)
func NewStructuredTool(name, description string, fn interface{}) (*StructuredTool, error)

// Execute validates params then calls function
func (t *StructuredTool) Execute(ctx context.Context, rawParams interface{}) (string, error)

// GetSchema returns JSON schema for parameters
func (t *StructuredTool) GetSchema() *ParameterSchema
```

**Acceptance Criteria:**
- [ ] Accept typed functions
- [ ] Validate parameters against schema before calling
- [ ] Convert raw JSON to typed struct
- [ ] Clear error messages on validation failure
- [ ] Unit tests with multiple parameter types

**Effort:** 12-16 hours

---

### Task 1.3: Update Tool Registry

**File:** `core/tools/registry.go` (MODIFY)

```go
package tools

// ToolRegistry manages tool registration
type ToolRegistry struct {
    tools map[string]Tool
    // ...
}

// Add registers a tool (can be structured or legacy)
func (r *ToolRegistry) Add(tool interface{}) error

// Example:
// registry.Add(GetStatus)  // Typed function
// registry.Add(GetWeatherTool())  // Legacy Tool struct
```

**Acceptance Criteria:**
- [ ] Support both StructuredTool and legacy Tool
- [ ] Validate tool names are unique
- [ ] Return clear error if duplicate
- [ ] Unit tests

**Effort:** 4-6 hours

---

## Phase 2: Auto-Generated Schemas (Week 2-3)

**Goal:** Remove hand-written JSON schemas entirely

### Task 2.1: Tool Object Refactoring

**File:** `core/common/types.go` (MODIFY)

```go
// OLD:
type Tool struct {
    ID          string
    Name        string
    Description string
    Func        interface{}
    Input       interface{}      // ❌ Unused
    Parameters  interface{}      // ❌ Could be JSON schema or map
    Output      interface{}      // ❌ Unused
}

// NEW:
type Tool struct {
    Name        string
    Description string
    Func        interface{}
    // Schema is automatically generated - no manual definition needed
    // Input/Output removed - confusing and unused
}
```

**Acceptance Criteria:**
- [ ] Remove unused Input/Output fields
- [ ] Update all tool creation code
- [ ] Update tests
- [ ] Create migration guide for users

**Effort:** 6-8 hours

---

### Task 2.2: Provider Tool Conversion

**File:** `core/providers/provider.go` (MODIFY)

```go
// When converting Tool to provider format:
// 1. Extract Name, Description
// 2. Generate schema from function signature
// 3. Schema used by LLM, function stored separately

type ProviderTool struct {
    Name       string
    Description string
    Parameters *ParameterSchema  // Auto-generated, not hand-written
    // Note: Function/implementation NOT sent to LLM
}
```

**Acceptance Criteria:**
- [ ] Schema auto-generated from struct tags
- [ ] No hand-written JSON schemas needed
- [ ] OpenAI and Ollama provider support it
- [ ] Tests pass

**Effort:** 8-10 hours

---

## Phase 3: Fail-Fast Validation (Week 3-4)

**Goal:** Catch configuration errors at load time, not runtime

### Task 3.1: Tool Registration Validation

**File:** `core/config/validation.go` (MODIFY)

```go
// When loading crew config:
// 1. Parse crew.yaml + agents/*.yaml
// 2. Collect all tool references from agents
// 3. Check each tool exists in ToolRegistry
// 4. Return clear error if mismatch

func ValidateToolsConfiguration(agents []*AgentConfig, registry *ToolRegistry) error {
    // For each agent, for each tool reference:
    for _, agent := range agents {
        for _, toolName := range agent.Tools {
            if _, exists := registry.Get(toolName); !exists {
                return fmt.Errorf(
                    "Tool '%s' referenced in agent '%s' (config: %s) but not registered\n"+
                    "Registered tools: %v",
                    toolName, agent.Name, agent.ConfigFile, registry.ListNames(),
                )
            }
        }
    }
    return nil
}
```

**Acceptance Criteria:**
- [ ] Validate at NewCrewExecutorFromConfig() call
- [ ] Clear error messages listing available tools
- [ ] Unit tests with missing tool scenarios
- [ ] Integration tests with quiz example

**Effort:** 6-8 hours

---

### Task 3.2: Route Validation

**File:** `core/routing/validation.go` (NEW)

```go
// Validate signal-based routing configuration
type RouteValidator struct {
    agents map[string]*Agent
}

func (rv *RouteValidator) Validate() error {
    // Check:
    // 1. All target agents exist
    // 2. No circular routes
    // 3. Terminal conditions properly defined
    return nil
}
```

**Acceptance Criteria:**
- [ ] Detect non-existent target agents
- [ ] Detect circular routing
- [ ] Clear error messages
- [ ] Unit tests

**Effort:** 4-6 hours

---

## Phase 4: Error Propagation (Week 4)

**Goal:** Send tool errors back to LLM so it can retry

### Task 4.1: Error Response Formatting

**File:** `core/tools/error_handling.go` (NEW)

```go
// When tool fails:
// 1. Check if error is retriable
// 2. Format error message clearly
// 3. Add to history for LLM to see
// 4. LLM can retry with different parameters

type ToolError struct {
    Type       string // "validation", "execution", "timeout"
    Message    string
    Retriable  bool
}

func (t *StructuredTool) Execute(ctx context.Context, params json.RawMessage) (string, error) {
    // 1. Validate params against schema
    if err := validate(params, t.schema); err != nil {
        // Return validation error - LLM should retry
        return "", ToolError{
            Type: "validation",
            Message: fmt.Sprintf("Parameter validation failed: %v", err),
            Retriable: true,
        }
    }

    // 2. Call function
    result, err := t.fn(ctx, params)
    if err != nil {
        // Return execution error
        return "", ToolError{
            Type: "execution",
            Message: err.Error(),
            Retriable: isRetriable(err),
        }
    }

    return result, nil
}
```

**Acceptance Criteria:**
- [ ] Tool errors formatted with metadata
- [ ] Errors sent to LLM in history
- [ ] LLM can see and respond to errors
- [ ] Integration tests with error scenarios

**Effort:** 8-10 hours

---

### Task 4.2: Workflow Error Integration

**File:** `core/execution/flow.go` (MODIFY)

```go
// In workflow execution loop:
// Instead of logging errors and continuing silently,
// add errors to history so LLM sees them

func ExecuteWorkflowStep(/*...*/) {
    response, err := agent.Execute(/*...*/)

    toolResults, toolErrors := tools.ExecuteToolCalls(response.ToolCalls)

    // ✅ NEW: Send tool errors to LLM
    if len(toolErrors) > 0 {
        errorMsg := formatToolErrors(toolErrors)
        history.Add(Message{
            Role: "system",
            Content: errorMsg,
        })
    }

    // Continue - LLM will see errors in next turn and retry
}
```

**Acceptance Criteria:**
- [ ] Tool errors in history for LLM
- [ ] LLM can see and understand errors
- [ ] Integration tests showing LLM retries
- [ ] Quiz example uses error handling

**Effort:** 6-8 hours

---

## Phase 5: Documentation & Examples (Week 4-5)

**Goal:** Make tool definition obvious and easy

### Task 5.1: Tool Definition Guide

**File:** `docs/tool-definition-guide.md` (NEW)

```markdown
# Tool Definition Guide

## Quick Start: 3 Steps

### Step 1: Define Parameters (Go struct)
```go
type GetStatusParams struct {
    ID string `json:"id" description:"The ID to check"`
}
```

### Step 2: Define Function
```go
func GetStatus(ctx context.Context, params GetStatusParams) (string, error) {
    // params.ID is typed and validated automatically
    return status, nil
}
```

### Step 3: Register Tool
```go
registry := core.NewToolRegistry()
registry.Add(GetStatus)
```

That's it! Framework handles:
- ✅ Schema generation from struct
- ✅ Parameter validation
- ✅ Error propagation
- ✅ Type safety

## Detailed Examples

### Example 1: Simple Tool
[code example]

### Example 2: Multiple Parameters
[code example]

### Example 3: Validation (minLength, enum, etc)
[code example]

### Example 4: Error Handling
[code example]
```

**Acceptance Criteria:**
- [ ] Clear 3-step process documented
- [ ] Multiple examples
- [ ] Common mistakes section
- [ ] Migration from old pattern

**Effort:** 4-6 hours

---

### Task 5.2: Update Examples

**File:** `examples/00-hello-crew/internal/tools.go` (REFACTOR)

```go
// OLD (~40 LOC):
func Greet(ctx context.Context, args map[string]interface{}) (string, error) {
    name, ok := args["name"].(string)
    if !ok {
        return "", fmt.Errorf("name must be string")
    }
    // ...
}
toolsMap["greet"] = &Tool{...} // manual JSON schema

// NEW (~20 LOC):
type GreetParams struct {
    Name string `json:"name" description:"Person's name"`
}

func Greet(ctx context.Context, params GreetParams) (string, error) {
    // params.Name already validated & typed
    return fmt.Sprintf("Hello %s!", params.Name), nil
}

// In main.go:
registry.Add(Greet)  // Done!
```

**Acceptance Criteria:**
- [ ] hello-crew refactored to new pattern
- [ ] quiz-exam refactored to new pattern
- [ ] All examples updated
- [ ] Examples run successfully
- [ ] Code is simpler and clearer

**Effort:** 8-12 hours

---

## Phase 6: Unified Routing (Optional, Week 5-6)

**Goal:** Simplify signal-based routing to tool-based routing

### Task 6.1: Routing Tools

**File:** `core/routing/routing_tools.go` (NEW)

Instead of:
```yaml
signals:
  teacher:
    - signal: "[QUESTION]"
      parallel_targets: [student, reporter]
```

Use:
```go
// Routing tools are just regular tools
type RouteToStudentParams struct {
    Question string `json:"question"`
}

func RouteToStudent(ctx context.Context, params RouteToStudentParams) (string, error) {
    // Framework automatically routes to student agent
    // (framework handles this, not user)
    return fmt.Sprintf("Routed: %s", params.Question), nil
}

agent.Tools = [RouteToStudent, RouteToReporter, Terminate]
```

**Acceptance Criteria:**
- [ ] Routing tools defined and work
- [ ] Framework handles routing automatically
- [ ] All agents can use routing tools
- [ ] No need for signal config in YAML
- [ ] Unit and integration tests

**Effort:** 16-20 hours (optional, high impact)

---

## Implementation Timeline

| Phase | Week | Tasks | Hours | Complexity |
|-------|------|-------|-------|-----------|
| 1 | 1-2 | Struct params + schema gen + registry | 20-34 | Medium |
| 2 | 2-3 | Auto schemas + tool refactoring | 14-18 | Medium |
| 3 | 3-4 | Fail-fast validation | 10-14 | Medium |
| 4 | 4 | Error propagation | 14-18 | Medium |
| 5 | 4-5 | Docs + examples | 12-18 | Low |
| 6 | 5-6 | Unified routing (optional) | 16-20 | High |
| Testing | All | Integration tests + examples | 20-30 | High |
| **Total** | **6 weeks** | **All phases** | **106-152 hours** | **Medium** |

---

## Success Metrics

### Code Metrics
| Metric | Before | After | Target |
|--------|--------|-------|--------|
| LOC per tool (hello-crew) | 40 | 15 | <20 |
| LOC per tool (quiz-exam) | 45 | 20 | <25 |
| Hand-written schemas | 100% | 0% | 0% |
| Config validation at load | 0% | 100% | 100% |
| Tool registration methods | 2 | 1 | 1 |

### Developer Experience
| Metric | Before | After | Target |
|--------|--------|-------|--------|
| Onboarding time | 2-3 hours | 30 min | <45 min |
| Tool definition clarity | 6.5/10 | 8.5/10 | >8.5/10 |
| Error message clarity | 5/10 | 9/10 | >8/10 |
| Example complexity | High | Low | Low |
| First-tool success rate | 40% | 90% | >90% |

### Framework Quality
| Metric | Before | After | Target |
|--------|--------|-------|--------|
| Type safety | Low | High | High |
| Validation boilerplate | 60% of code | 0% | 0% |
| Silent failures | Many | None | None |
| Error propagation | Manual | Auto | Auto |
| Schema consistency | Manual | Auto | Auto |

---

## Risk Mitigation

### Risk 1: Breaking Changes
**Risk:** Changes to Tool struct and registration API
**Mitigation:**
- Provide migration guide
- Support legacy pattern for 1 release
- Clear deprecation warnings

### Risk 2: Schema Generation Complexity
**Risk:** Struct tag parsing might not handle all cases
**Mitigation:**
- Start with common cases (string, int, bool, enums)
- Allow manual schema override if needed
- Comprehensive unit tests

### Risk 3: Performance Impact
**Risk:** Schema generation and validation at load time
**Mitigation:**
- Cache generated schemas
- Validate only what's needed
- Benchmark against old implementation

---

## Dependencies

- **Go 1.18+** (for generic types if used)
- **reflect package** (for struct introspection)
- **No new external dependencies**

---

## Testing Strategy

### Unit Tests
```
core/tools/
├─ schema_generator_test.go
├─ structured_tool_test.go
├─ registry_test.go
├─ error_handling_test.go
└─ validation_test.go

core/config/
└─ tool_validation_test.go
```

### Integration Tests
```
examples/00-hello-crew/
├─ test_new_pattern.go
└─ test_old_pattern_compat.go

examples/01-quiz-exam/
├─ test_tool_registration.go
└─ test_tool_execution.go
```

### Compatibility Tests
```
- Old tool registration still works
- New tool registration works
- Both can be mixed if needed
```

---

## Deliverables

1. ✅ Struct-based parameter support
2. ✅ Auto-generated schemas
3. ✅ Fail-fast validation
4. ✅ Error propagation
5. ✅ Updated documentation
6. ✅ Refactored examples
7. ✅ Migration guide
8. ✅ Test suite (>85% coverage)

---

## Success Criteria (All Must Pass)

- [ ] All unit tests pass
- [ ] All integration tests pass
- [ ] Quiz example runs without errors
- [ ] hello-crew refactored and cleaner
- [ ] No regression in existing functionality
- [ ] DX score improved to 8.5+/10
- [ ] Developer can write first tool in <30 min
- [ ] All tools <20 LOC
- [ ] Error messages are clear and helpful
- [ ] Documentation is comprehensive

---

## Next Steps

1. **Review this roadmap** with team
2. **Get approval** on timeline and resources
3. **Create tracking** in issue system
4. **Start Phase 1** (struct parameters)
5. **Weekly sync** to track progress

---

## Questions & Contact

For questions about this roadmap:
- Review comparison with best-practice frameworks
- Check examples for expected outcomes
- Refer to success metrics for clarity

**Generated:** 2025-12-25
**Status:** Ready for implementation

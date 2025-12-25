# Phase 2: Critical Features - Implementation Guide

**Status:** In Progress
**Date:** 2025-12-25
**Estimated Duration:** 13 hours
**Priority:** Medium - Unblocks critical multi-agent workflows

---

## 📋 OVERVIEW

Phase 2 addresses two critical features that enable multi-agent workflows and tool execution:

- **Issue 2.1:** Signal-Based Agent Routing (Status: PARTIALLY COMPLETE)
- **Issue 2.2:** Tool Conversion in Agent Execution (Status: TODO)

---

## 🔍 CURRENT ARCHITECTURE ANALYSIS

### Issue 2.1: Signal-Based Agent Routing

#### Current State

The signal-based routing is **PARTIALLY IMPLEMENTED** across two layers:

1. **executor/workflow.go** (NEW ARCHITECTURE - COMPLETE) ✅
   - `ExecutionFlow.ExecuteWithCallbacks()` (lines 208-257)
   - `ExecutionFlow.HandleAgentResponse()` (lines 115-180)
   - Uses: `DetermineNextAgentWithSignals()` from workflow/routing.go
   - **Status:** FULLY IMPLEMENTED for multi-agent routing

2. **workflow/execution.go** (OLD ARCHITECTURE - INCOMPLETE) ⚠️
   - Lines 193-195: Signal-based routing TODO
   - Lines 218-220: Handoff target routing TODO
   - **Status:** LEGACY CODE - Not actively used by crew.go

3. **workflow/routing.go** (UTILITIES - COMPLETE) ✅
   - `DetermineNextAgentWithSignals()` (lines 38-121)
   - `DetermineNextAgent()` (lines 13-35)
   - **Status:** FULLY IMPLEMENTED

#### Key Finding

**The signal-based routing is already implemented!** The crew.go Execute() method uses `executor/workflow.go` which properly implements:

```go
// From executor/workflow.go lines 244
shouldContinue, err := ef.HandleAgentResponse(ctx, response, routing, agents)
```

This calls:
```go
// From executor/workflow.go lines 157-165
if nextAgent, exists := agentsMap[nextAgentID]; exists {
    ef.CurrentAgent = nextAgent
    ef.HandoffCount++
    ef.State.RecordHandoff()
    return true, nil // Continue with next agent
}
```

**Action Required:** Update/deprecate the legacy workflow/execution.go or update the CLEANUP_ACTION_PLAN to reflect that this is already done.

---

### Issue 2.2: Tool Conversion in Agent Execution

#### Current State - GAPS IDENTIFIED

**Problem:** Agent tools not being passed to providers

**Location:** `core/agent/execution.go` lines 131 & 181
```go
Tools: nil, // TODO: Implement proper tool conversion from agent.Tools
```

#### Tool Format Mismatch

```
Agent.Tools Format:           ProviderTool Format Required:
├── []interface{}             ├── []ProviderTool
└── Stored as generic         ├── Name: string
    (type varies)             ├── Description: string
                              └── Parameters: map[string]interface{}
```

#### Root Cause

1. Agent.Tools is `[]interface{}` (generic, could be anything)
2. Provider expects `[]ProviderTool` (structured format)
3. No conversion function exists
4. Tools are completely ignored during execution

---

## ✅ ACTION ITEMS

### Issue 2.1: Signal-Based Routing

**Status:** ✅ ALREADY IMPLEMENTED

The current implementation in `executor/workflow.go` fully supports:
- Signal extraction from agent responses
- Signal-based routing using `DetermineNextAgentWithSignals()`
- Handoff target routing
- Agent lookup and continuation

**Recommended Actions:**
1. Update documentation to clarify that routing is implemented in executor, not workflow
2. Consider deprecating workflow/execution.go in favor of executor/workflow.go
3. Add tests for complete signal-based routing flow

**Files Already Complete:**
- ✅ executor/workflow.go - Fully implements routing
- ✅ workflow/routing.go - All routing logic complete
- ✅ core/crew.go - Uses proper ExecuteWithCallbacks method

---

### Issue 2.2: Tool Conversion Implementation

**Status:** 🔴 REQUIRES IMPLEMENTATION

#### Step 1: Analyze Agent Tool Types

First, determine what types can be in Agent.Tools:

```go
// Potential agent.Tools structures:
type AgentToolDef struct {
    Name        string
    Description string
    Function    func(ctx context.Context, args map[string]interface{}) (string, error)
    Parameters  map[string]interface{}
}
// OR simple Tool struct from common.Tool
// OR string tool names to lookup from registry
```

**Action:** Inspect crew config parsing and tool definitions

#### Step 2: Implement Tool Conversion Function

Location: `core/agent/execution.go` (new function)

```go
// ConvertAgentToolsToProviderTools converts agent tools to provider format
func ConvertAgentToolsToProviderTools(agentTools []interface{}) []providers.ProviderTool {
    var providerTools []providers.ProviderTool

    for _, tool := range agentTools {
        // Type assertion and conversion
        switch t := tool.(type) {
        case *common.Tool:
            providerTools = append(providerTools, providers.ProviderTool{
                Name:        t.Name,
                Description: t.Description,
                Parameters:  t.Parameters.(map[string]interface{}),
            })
        case providers.ProviderTool:
            providerTools = append(providerTools, t)
        case string:
            // Tool name lookup - implement if tool registry exists
            // providerTools = append(providerTools, lookupTool(t))
        default:
            // Log warning for unhandled tool type
            fmt.Printf("[WARN] Unknown tool type: %T\n", tool)
        }
    }

    return providerTools
}
```

#### Step 3: Update executeWithModelConfig

Update lines 131 & 181 to use conversion:

```go
// Before (line 131)
Tools: nil, // TODO: Implement proper tool conversion from agent.Tools

// After
Tools: ConvertAgentToolsToProviderTools(agent.Tools),
```

#### Step 4: Add Tests

Create test in `core/agent/execution_test.go`:

```go
func TestConvertAgentToolsToProviderTools(t *testing.T) {
    // Test with common.Tool
    // Test with mixed types
    // Test with empty array
    // Test with invalid types (should be skipped with warning)
}

func TestExecuteAgentWithTools(t *testing.T) {
    // Test agent execution passes tools to provider
    // Verify ProviderTool format is correct
    // Test tool calls are extracted and processed
}
```

---

## 📊 DETAILED IMPLEMENTATION PLAN

### Issue 2.1: Documentation Update

**Files to Update:**
- CLEANUP_ACTION_PLAN.md - Mark as COMPLETE
- Add note that routing is in executor/workflow.go not workflow/execution.go
- Consider deprecation notice in workflow/execution.go

**New Issue:**
- Issue 2.1-Alpha: Deprecate legacy workflow/execution.go in favor of executor/workflow.go

---

### Issue 2.2: Full Implementation

**Phase 2.2a: Analysis (1 hour)**
```bash
# 1. Inspect crew config parsing
find core -name "*.go" | xargs grep -l "agent.Tools"
# 2. Check tool definitions across codebase
# 3. Understand tool registry if it exists
```

**Phase 2.2b: Implementation (3 hours)**
- Write ConvertAgentToolsToProviderTools()
- Update executeWithModelConfig() line 131
- Update executeWithModelConfigStream() line 181
- Add error handling and logging

**Phase 2.2c: Testing (2 hours)**
- Unit tests for conversion function
- Integration test with actual provider
- Edge case testing (empty tools, invalid types)

**Phase 2.2d: Documentation (1 hour)**
- Add comments explaining tool conversion
- Document tool format requirements
- Update this guide with completion notes

---

## 🔗 DEPENDENCY ANALYSIS

### Issue 2.1 Dependencies
✅ ZERO - Already implemented

### Issue 2.2 Dependencies
- Common types must be defined (Tool struct)
- Provider types must be defined (ProviderTool struct)
- Agent structure must have Tools field
- **All dependencies exist** - Ready for implementation

---

## 🎯 SUCCESS CRITERIA

### Issue 2.1
- [x] Signal-based routing working
- [x] Agents properly transition between each other
- [x] History captures all transitions
- [ ] Documentation updated to reflect implementation location

### Issue 2.2
- [ ] Tool conversion function implemented
- [ ] ExecuteWithProvider passes tools to providers
- [ ] Tools available during agent execution
- [ ] Tool calls properly extracted and processed
- [ ] Unit tests passing (100%)
- [ ] Integration tests passing

---

## 📈 COMPLETION CHECKLIST

- [ ] Issue 2.1 documented as complete
- [ ] Issue 2.2 conversion function implemented
- [ ] executeWithModelConfig() updated
- [ ] executeWithModelConfigStream() updated
- [ ] Tests added and passing
- [ ] Code review ready
- [ ] Phase 2 marked complete

---

## 📝 TECHNICAL NOTES

### Important Findings

1. **Multi-Agent Routing Works**: The executor/workflow.go properly implements the routing with callbacks. The workflow package is called correctly from crew.go.

2. **Tool Format Gap**: Agent.Tools is stored as `[]interface{}` which requires runtime type assertion. No safe conversion exists yet.

3. **Signal Extraction**: Already implemented in ExecuteAgent() via ExtractSignalsFromContent() - proper signals are extracted.

4. **Architecture Separation**:
   - executor/workflow.go = New architecture (crew.go uses this) ✅
   - workflow/execution.go = Old architecture (legacy)
   - workflow/routing.go = Shared utilities ✅

---

## 🚀 NEXT STEPS

1. Mark Issue 2.1 as complete in this document
2. Implement Issue 2.2 tool conversion
3. Write comprehensive tests
4. Prepare for Phase 3 (Large file refactoring)

---

**Document Created:** 2025-12-25
**Last Updated:** 2025-12-25
**Author:** Claude Haiku 4.5
**Status:** In Progress

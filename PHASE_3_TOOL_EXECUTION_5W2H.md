# Phase 3.1: Tool Execution Implementation - 5W2H Framework

**Status:** Ready for Implementation
**Priority:** CRITICAL - Blocks all tool functionality
**Date:** 2025-12-25
**Estimated Duration:** 4-5 hours

---

## 🎯 5W2H ANALYSIS - TOOL EXECUTION

### **WHAT (Là gì?) - Define the Problem**

**Current State:**
- ✅ Tools are defined in agent config
- ✅ Tools are converted to provider format (Phase 2)
- ✅ LLM extracts tool calls from response
- ❌ **Tool calls are never executed**
- ❌ **Results are never returned to agent**
- ❌ **Agent never sees tool results**

**What Needs to Happen:**
```
Agent Response Received
    ↓
Check for Tool Calls
    ↓
Find Each Tool
    ↓
Execute Tool with Arguments
    ↓
Collect Results
    ↓
Add Results to History
    ↓
Optionally: Re-invoke Agent with Results
```

**Required Components:**
1. Tool executor - Execute single tool
2. Tool batch executor - Execute multiple tools
3. Tool lookup - Find tool by name in agent.Tools
4. Result formatter - Format results for history
5. History integration - Add results to conversation

---

### **WHY (Tại sao?) - Justification**

**Business Impact:**
1. **Feature Enablement** - Tools completely non-functional without this
2. **User Experience** - Agents can't answer questions requiring tools
3. **Use Cases Blocked** - Any tool-dependent workflow fails
4. **Testing** - Can't verify tool functionality end-to-end

**Technical Impact:**
1. **Architecture Complete** - Fills critical gap
2. **Error Handling** - Uses existing retry + error logic
3. **Workflow Loop** - Enables agent → tool → agent cycle
4. **Integration Ready** - All pieces exist, just need orchestration

**Timeline Impact:**
- **Blocks Phase 4** - Can't clean up without this working
- **Blocks Examples** - Tool examples don't work currently
- **Blocks Testing** - Need tool execution for integration tests

---

### **WHERE (Ở đâu?) - Location & Architecture**

**New File: `/core/tools/executor.go`**

```
Purpose: Tool execution orchestration
Size: 200-250 LOC
Exports:
  ├─ ExecuteTool()        - Execute single tool
  ├─ ExecuteToolCalls()   - Execute tool batch
  └─ Helper functions:
      ├─ findToolByName()
      ├─ formatToolResults()
      └─ extractToolFunction()
```

**Modified Files:**

1. **`/core/executor/workflow.go`** (lines ~80-110)
   - After agent response received
   - Before response returned to user
   - Call ExecuteToolCalls() if tools present
   - Add results to history
   - Optional: Re-invoke agent

2. **`/core/agent/execution.go`**
   - No changes needed (tool conversion already done ✅)

3. **`/core/common/types.go`**
   - No changes needed (types already exist)

---

### **WHEN (Khi nào?) - Timeline & Dependencies**

**Timeline:**
```
Phase 3.1 Implementation: 4-5 hours
├─ 0.5h: Design & API review
├─ 1.5h: Implement ExecuteTool()
├─ 1.0h: Implement ExecuteToolCalls()
├─ 1.0h: Write tests (10+ cases)
└─ 1.0h: Integration & verification
```

**Dependencies:**
- ✅ Tool definition (exists)
- ✅ Tool extraction (exists)
- ✅ Tool conversion (exists - Phase 2)
- ✅ Error handling (exists)
- ✅ Retry logic (exists)
- ⏳ Just needs orchestration

**Blocking:**
- Can't move to Phase 4 without this
- Examples won't work without this
- Integration tests incomplete without this

---

### **WHO (Ai?) - Team & Responsibilities**

**Development:**
- **Implement executor.go** - Core team
- **Update workflow integration** - Workflow team
- **Testing** - QA team

**Review:**
- **Code review** - Architecture team
- **Test verification** - QA team
- **Integration test** - Full team

**Documentation:**
- **Implementation details** - Developer who implements
- **Usage examples** - Documentation team
- **API docs** - API documentation team

---

### **HOW (Làm thế nào?) - Implementation Steps**

### **Step 1: Design & API Definition** (30 minutes)

**Define the API:**

```go
// File: core/tools/executor.go

// ExecuteTool executes a single tool and returns result
func ExecuteTool(ctx context.Context, toolName string, tool interface{}, args map[string]interface{}) (string, error) {
    // Find tool if string name provided
    // Extract tool function
    // Call ExecuteWithRetry from errors.go
    // Return result string
}

// ExecuteToolCalls executes all tool calls from agent response
func ExecuteToolCalls(ctx context.Context, toolCalls []common.ToolCall, agentTools []interface{}) (map[string]string, error) {
    // For each tool call:
    //   - Find tool in agentTools
    //   - Execute it
    //   - Collect results
    // Return map[toolName]result or error
}

// findToolByName finds tool by name in agent.Tools
func findToolByName(agentTools []interface{}, toolName string) *common.Tool {
    // Search agentTools for matching name
    // Return tool or nil
}

// formatToolResults formats tool results for history
func formatToolResults(results map[string]string) string {
    // Format results as readable message
    // Include tool names and results
    // Return formatted string
}

// extractToolFunction extracts executable function from tool
func extractToolFunction(tool *common.Tool) (ToolHandler, error) {
    // Assert tool.Func to ToolHandler type
    // Validate function exists
    // Return function
}
```

**Design Decisions:**
- ✅ Batch execution (multiple tools at once)
- ✅ Error recovery (individual tool failure doesn't stop others)
- ✅ Result formatting (readable history messages)
- ✅ Type safety (use ToolHandler type from errors.go)

---

### **Step 2: Implement ExecuteTool()** (90 minutes)

**Core Implementation:**

```go
func ExecuteTool(ctx context.Context, toolName string, tool interface{}, args map[string]interface{}) (string, error) {
    // Type check
    if tool == nil {
        return "", fmt.Errorf("tool is nil for '%s'", toolName)
    }

    // Type assert to common.Tool
    commonTool, ok := tool.(*common.Tool)
    if !ok {
        // Try by value
        if t, ok := tool.(common.Tool); ok {
            commonTool = &t
        } else {
            return "", fmt.Errorf("tool '%s' is not common.Tool type", toolName)
        }
    }

    // Extract function
    handler, ok := commonTool.Func.(ToolHandler)
    if !ok || handler == nil {
        return "", fmt.Errorf("tool '%s' has no executable function", toolName)
    }

    // Execute with retry logic (from errors.go)
    result, err := ExecuteWithRetry(ctx, handler, args)
    if err != nil {
        return "", fmt.Errorf("tool '%s' execution failed: %w", toolName, err)
    }

    // Convert result to string
    resultStr := fmt.Sprintf("%v", result)
    return resultStr, nil
}
```

**Key Points:**
- Type assertions for both pointer and value types
- Uses existing ExecuteWithRetry() from errors.go
- Proper error wrapping with context
- Handles nil cases gracefully

---

### **Step 3: Implement ExecuteToolCalls()** (60 minutes)

**Batch Implementation:**

```go
func ExecuteToolCalls(ctx context.Context, toolCalls []common.ToolCall, agentTools []interface{}) (map[string]string, error) {
    results := make(map[string]string)
    var errors []string

    // Build tool map for fast lookup
    toolMap := buildToolMap(agentTools)

    for _, call := range toolCalls {
        // Find tool
        tool, exists := toolMap[call.ToolName]
        if !exists {
            errors = append(errors, fmt.Sprintf("tool '%s' not found", call.ToolName))
            continue
        }

        // Execute tool
        result, err := ExecuteTool(ctx, call.ToolName, tool, call.Arguments)
        if err != nil {
            errors = append(errors, fmt.Sprintf("tool '%s' failed: %v", call.ToolName, err))
            continue
        }

        // Store result
        results[call.ToolName] = result
    }

    // Return results and any collected errors
    if len(errors) > 0 {
        return results, fmt.Errorf("tool execution errors: %v", errors)
    }

    return results, nil
}

func buildToolMap(agentTools []interface{}) map[string]interface{} {
    toolMap := make(map[string]interface{})
    for _, tool := range agentTools {
        if tool == nil {
            continue
        }

        // Handle pointer
        if toolPtr, ok := tool.(*common.Tool); ok {
            toolMap[toolPtr.Name] = toolPtr
            continue
        }

        // Handle value
        if toolVal, ok := tool.(common.Tool); ok {
            toolMap[toolVal.Name] = toolVal
            continue
        }
    }
    return toolMap
}
```

**Key Features:**
- Batch execution (all tools at once)
- Partial failure handling (one tool failure doesn't stop others)
- Tool lookup map (O(1) instead of O(n))
- Error collection (all errors reported together)
- Clear error messages (tool name included)

---

### **Step 4: Integration with Workflow** (60 minutes)

**Update `/core/executor/workflow.go`** (lines ~80-110):

```go
// After agent response is received (current line ~110)
func (ef *ExecutionFlow) ExecuteWorkflowStep(ctx context.Context, handler workflow.OutputHandler, apiKey string) (*common.AgentResponse, error) {
    // ... existing code to execute agent ...

    // ADD: Tool execution if tools were called
    if response != nil && len(response.ToolCalls) > 0 {
        // Execute all tools from response
        toolResults, toolErr := tools.ExecuteToolCalls(ctx, response.ToolCalls, ef.CurrentAgent.Tools)

        // Add tool results to history
        if len(toolResults) > 0 {
            resultMsg := common.Message{
                Role:    common.RoleSystem,
                Content: tools.FormatToolResults(toolResults),
            }
            ef.History = append(ef.History, resultMsg)

            // OPTIONAL: Re-invoke agent with tool results for awareness
            // (Enables agent → tool → think → respond cycle)
            // This would require another agent execution step
            // Decision: Include if time permits, skip for MVP
        }

        // Log any tool execution errors
        if toolErr != nil {
            fmt.Printf("[WARN] Some tools failed: %v\n", toolErr)
            // Still continue - partial success is OK
        }
    }

    return response, nil
}
```

**Integration Points:**
- After agent.Complete() call
- Before response returned to caller
- Works with existing history tracking
- Optional: Re-invoke agent with results

---

### **Step 5: Testing** (60 minutes)

**Test Suite (10+ test cases):**

```go
func TestExecuteTool(t *testing.T) {
    // Test 1: Execute valid tool
    // Test 2: Tool with nil Func
    // Test 3: Tool with wrong Func type
    // Test 4: Missing tool
    // Test 5: Tool error handling
}

func TestExecuteToolCalls(t *testing.T) {
    // Test 1: Single tool execution
    // Test 2: Multiple tools execution
    // Test 3: Partial failure (one tool fails, others succeed)
    // Test 4: All tools fail
    // Test 5: No tools (empty list)
    // Test 6: Tool not found
    // Test 7: Batch execution order (doesn't matter)
}

func TestFindToolByName(t *testing.T) {
    // Test 1: Find by exact name
    // Test 2: Tool not found
    // Test 3: Nil tools
    // Test 4: Empty tools
}

func TestFormatToolResults(t *testing.T) {
    // Test 1: Single result
    // Test 2: Multiple results
    // Test 3: Empty results
    // Test 4: Special characters in result
}
```

---

### **HOW MUCH (Bao nhiêu?) - Effort & Resources**

**Time Breakdown:**
```
Analysis & Design:         0.5 hours
Code Implementation:       2.5 hours
├─ ExecuteTool()          1.0h
├─ ExecuteToolCalls()     1.0h
└─ Helpers                0.5h
Testing:                  1.0 hours
Integration:              0.5 hours
Verification:             0.5 hours
────────────────────────────────────
TOTAL:                    5.0 hours
```

**Code Changes:**
```
New Code:
  /core/tools/executor.go          +200-250 LOC

Modified Code:
  /core/executor/workflow.go       +30-50 LOC (tool execution integration)

Tests:
  /core/tools/executor_test.go     +250-300 LOC

Total New:                         +480-600 LOC
```

---

## 📋 DETAILED TASK BREAKDOWN

### **Task 3.1.1: Design & API Definition**
- [ ] Define ExecuteTool() signature
- [ ] Define ExecuteToolCalls() signature
- [ ] Define helper functions
- [ ] Review with team
- **Duration:** 30 minutes

### **Task 3.1.2: Implement executor.go**
- [ ] Create /core/tools/executor.go
- [ ] Implement ExecuteTool()
- [ ] Implement ExecuteToolCalls()
- [ ] Implement helpers
- [ ] Add comprehensive comments
- **Duration:** 90 minutes

### **Task 3.1.3: Write Tests**
- [ ] Create executor_test.go
- [ ] Write 10+ test cases
- [ ] All tests passing
- [ ] Coverage >90%
- **Duration:** 60 minutes

### **Task 3.1.4: Integration**
- [ ] Update executor/workflow.go
- [ ] Integrate tool execution
- [ ] Integrate result formatting
- [ ] Handle errors gracefully
- **Duration:** 60 minutes

### **Task 3.1.5: Verification**
- [ ] Build successful
- [ ] All tests passing
- [ ] Manual verification
- [ ] Example works end-to-end
- **Duration:** 30 minutes

---

## ✅ SUCCESS CRITERIA

**Functional:**
- ✅ Tool execution works end-to-end
- ✅ Multiple tools can be executed in one cycle
- ✅ Tool results available to agent
- ✅ Partial failure handled gracefully
- ✅ Tool not found handled with clear error

**Quality:**
- ✅ All tests passing (11+ tests)
- ✅ Code coverage >90%
- ✅ Build successful
- ✅ No breaking changes
- ✅ Comprehensive error messages

**Documentation:**
- ✅ Code comments explain logic
- ✅ Test cases document behavior
- ✅ Error messages clear
- ✅ Function contracts documented

---

## 🚀 NEXT PHASE (After 3.1)

**Phase 3.2: Delete Legacy Code**
- Remove workflow/execution.go (273 LOC)
- Remove orphaned functions (30 LOC)
- Duration: 1-2 hours

**Phase 3.3: Code Organization**
- Clean up remaining dead code
- Add deprecation notices
- Document architecture
- Duration: 1-2 hours

---

## 📚 REFERENCE IMPLEMENTATION

**Key Code Locations:**
- Tool definition: `/core/common/types.go` lines 18-26
- Tool type: `type ToolHandler func(...)`
- Error handling: `/core/tools/errors.go` - ExecuteWithRetry()
- Workflow: `/core/executor/workflow.go` lines 49-110
- Messages: `/core/common/types.go` - Message type

**Related Files:**
- `/core/agent/execution.go` - Tool conversion (already done ✅)
- `/core/providers/ollama/provider.go` - Tool extraction
- `/core/executor/workflow.go` - Workflow execution

---

**Status:** Ready for Implementation
**Assigned to:** Development Team
**Start Date:** 2025-12-25
**Target Completion:** Within 5 hours
**Next Review:** Upon Phase 3.1 completion

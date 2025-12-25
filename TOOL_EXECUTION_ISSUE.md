# Tool Execution Issue - Complete Analysis

## Problem Summary

**Status:** Tools are extracted and passed to LLM, but **NOT executed** - results are never returned to the agent

When user asks "Mấy giờ rồi?" (What time is it?):
1. ✅ Agent receives tools in system prompt
2. ✅ LLM extracts tool call: `GetCurrentTime()`
3. ✅ Agent response includes ToolCalls
4. ❌ Tool is NEVER EXECUTED
5. ❌ Result NEVER returned to agent
6. ❌ Agent gets no context about tool outcome

## Current Data Flow

```
User Input
  ↓
Agent.ExecuteAgent()
  ↓
Provider.Complete(ctx, request with Tools)
  ↓
LLM extracts: "GetCurrentTime()"
  ↓
Returns AgentResponse with ToolCalls[]
  ↓
ExecuteWorkflowStep() receives response
  ↓
response.Content = "<think>...\nGetCurrentTime()\n"
response.ToolCalls = [{ToolName: "GetCurrentTime", Arguments: {}}]
  ↓
❌ DEAD END - No tool execution happens
  ↓
Response returned to user with just thinking text
```

## Missing Implementation

### Current Tool-Related Code

1. **Tool Extraction** ✅ - `/core/providers/ollama/provider.go` extracts tool calls from LLM text
2. **Tool Passing** ✅ - `/core/agent/execution.go` line 131 now passes `ConvertAgentToolsToProviderTools(agent.Tools)` to provider
3. **Tool Execution** ❌ **MISSING** - No code executes actual tools after extraction

### Required Components

The following components exist but are **NOT USED** in the execution flow:

1. **Tool Handler Signature** - `/core/tools/errors.go`:
```go
type ToolHandler func(ctx context.Context, args map[string]interface{}) (string, error)
```

2. **Retry Logic** - `/core/tools/errors.go` function `ExecuteWithRetry()` exists but never called

3. **Error Handling** - Complete error classification and retry logic built but unused

### Where Tool Execution Should Happen

**Option 1: In Workflow (Recommended for Agent Loop)**
- File: `/core/executor/workflow.go`
- After agent response is received in `ExecuteWorkflowStep()`
- Execute all tools in `response.ToolCalls[]`
- Collect tool results
- Add tool results back to history
- Re-prompt agent with tool results
- Agent should acknowledge and continue

**Option 2: In Provider (Less Flexible)**
- File: `/core/providers/ollama/provider.go` or similar
- After LLM generates tool calls
- Execute them before returning to agent
- Include tool results in next context
- Not preferred - ties tool execution to LLM provider

**Option 3: In Example (Not Recommended)**
- Handle tool execution in example code
- Not recommended - tools should be handled by core

## Implementation Requirements

### 1. Tool Executor Component

Need function to execute a single tool:
```go
func ExecuteTool(ctx context.Context, toolName string, tool interface{}, args map[string]interface{}) (string, error) {
    // Type assert to *common.Tool
    // Extract tool.Func
    // Call ExecuteWithRetry
    // Return result string
}
```

### 2. Tool Batch Executor

Need function to execute all tools from response:
```go
func ExecuteToolCalls(ctx context.Context, toolCalls []common.ToolCall, agentTools []interface{}) (map[string]string, error) {
    // For each tool call:
    //   - Find tool in agentTools
    //   - Execute it
    //   - Collect results
    // Return map[toolName]result
}
```

### 3. Tool Result Integration

After tool execution, results must be:
```go
// In ExecuteWorkflowStep or HandleAgentResponse:
if response.ToolCalls != nil && len(response.ToolCalls) > 0 {
    toolResults, err := ExecuteToolCalls(ctx, response.ToolCalls, agent.Tools)

    // Add tool results to history for agent awareness
    toolResultMsg := common.Message{
        Role: "system",
        Content: formatToolResults(toolResults),
    }
    ef.History = append(ef.History, toolResultMsg)

    // Optionally: Re-invoke agent with tool results
    // This creates an agent loop: think -> call tools -> see results -> think again
}
```

## Tool Lookup by Name

When executing tools, need to match by name:
```go
func findToolByName(agentTools []interface{}, toolName string) *common.Tool {
    for _, tool := range agentTools {
        if t, ok := tool.(*common.Tool); ok && t.Name == toolName {
            return t
        }
    }
    return nil
}
```

## Files That Need Changes

1. **`/core/workflow/execution.go`** - Add tool execution in `ExecuteWorkflowStep()` after agent response
2. **`/core/agent/execution.go`** - May need new helper function for tool execution
3. **`/core/executor/workflow.go`** - May need tool execution integration
4. **New file or new package** - Tool executor component with proper error handling

## Testing the Fix

After implementation, example should show:
```
> Mấy giờ rồi?
[Agent thinking...] GetCurrentTime()
[Tool executing...] GetCurrentTime
[Tool result] {"timestamp": 1735137000, "datetime": "2025-12-25 14:50:00", "timezone": "UTC"}
[Agent thinking again with result...]
Response: Bây giờ là 14:50:00 (hoặc similar with actual time)
```

## Related Code Locations

- Tool definition: `/core/common/types.go` lines 18-26
- Tool extraction: `/core/providers/ollama/provider.go`
- Tool passing to LLM: `/core/agent/execution.go` line 131
- Error handling: `/core/tools/errors.go`
- Retry logic: `/core/tools/errors.go` `ExecuteWithRetry()`
- Workflow execution: `/core/executor/workflow.go` lines 49-110

## Current Status

- **Tools in LLM context**: ✅ Working (after fix to line 131)
- **LLM extracts tool calls**: ✅ Working
- **Tool execution**: ❌ **NOT IMPLEMENTED**
- **Tool result feedback**: ❌ **NOT IMPLEMENTED**

## Recommendation

This is a **Phase 2 feature** that needs complete implementation in core. The example code is correct, but core is missing the tool execution loop. Once core implements tool execution, examples will work automatically.

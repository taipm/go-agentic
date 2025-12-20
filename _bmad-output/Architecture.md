---
title: "Architecture: go-agentic Library Quality Improvements"
version: "1.0.0"
date: "2025-12-20"
project: "go-agentic"
type: "Technical Architecture"
status: "In Review"
---

# Architecture Document
## go-agentic Library Quality & Robustness Improvements

**Project:** go-agentic
**Version:** 1.0.0
**Date:** 2025-12-20
**Audience:** Development Team, Architecture Review

---

## 1. ARCHITECTURE OVERVIEW

### Current Architecture (Existing)

The go-agentic library follows a clean layered architecture:

```
┌─────────────────────────────────────────────┐
│         HTTP Server Layer                   │
│       (http.go, streaming.go)               │
└──────────────┬──────────────────────────────┘
               │
┌──────────────┴──────────────────────────────┐
│      Team Orchestration Layer               │
│  (team.go - TeamExecutor)                   │
├──────────────┬──────────────────────────────┤
│ Agent Loop │ Tool Execution │ Routing       │
└──────────────┬──────────────────────────────┘
               │
┌──────────────┴──────────────────────────────┐
│     Agent Execution Layer                   │
│  (agent.go - ExecuteAgent)                  │
├──────────────┬──────────────────────────────┤
│ OpenAI API   │ Tool Calls │ System Prompts  │
└──────────────┬──────────────────────────────┘
               │
┌──────────────┴──────────────────────────────┐
│  Configuration Layer                        │
│  (config.go, types.go)                      │
├──────────────┬──────────────────────────────┤
│ YAML Loading │ Type Defs │ Tool Handlers    │
└─────────────────────────────────────────────┘
```

### Issues in Current Architecture

1. **agent.go:24** - Model parameter ignored (hardcoded "gpt-4o-mini")
2. **agent.go:126-190** - Text parsing instead of native API
3. **Missing validation layers** - Parameters not validated
4. **Inadequate error handling** - Silent failures in tool execution
5. **Semantic mismatch** - Tool results use wrong message role

---

## 2. ARCHITECTURAL DECISIONS

### AD-1: Fix Hardcoded Model Bug (CRITICAL)

**Problem:** Line `agent.go:24` hardcodes `"gpt-4o-mini"` ignoring `agent.Model` field.

**Current Code:**
```go
params := openai.ChatCompletionNewParams{
    Model:    "gpt-4o-mini",  // ❌ HARDCODED
    Messages: messages,
}
```

**Decision:** Use `agent.Model` from configuration.

**Technical Solution:**
```go
params := openai.ChatCompletionNewParams{
    Model:    agent.Model,  // ✅ USE CONFIGURED MODEL
    Messages: messages,
}
```

**Why This Approach:**
- ✅ Respects user configuration
- ✅ Enables cost optimization (cheaper models for simple tasks)
- ✅ Enables quality optimization (stronger models for complex tasks)
- ✅ Minimal code change
- ✅ Backward compatible (existing configs still work)

**Files Affected:** `go-agentic/agent.go`
**Impact:** CRITICAL - Enables proper agent model selection

**Acceptance Criteria:**
- [ ] Agent.Model field is read and used
- [ ] Integration tests verify model used in API call
- [ ] Existing examples work unchanged

---

### AD-2: Implement Native Tool Call Parsing (CRITICAL)

**Problem:** Current text-based parsing (agent.go:126-190) is fragile:
- Doesn't handle whitespace variations
- Vulnerable to false positives
- Doesn't use OpenAI's native `tool_calls` API

**Current Code:**
```go
// Parse from response text via string matching
if strings.Contains(line, toolName+"(") {
    // Extract args manually from text
}
```

**Decision:** Use OpenAI's native `tool_calls` field in API response.

**Technical Solution:**

OpenAI's Chat Completion response includes:
```go
type Message struct {
    ToolCalls []ToolCall  // ← Native tool calls here
    Content   string      // ← Text fallback
}

type ToolCall struct {
    ID       string
    Type     string                 // "function"
    Function struct {
        Name      string
        Arguments string             // JSON string
    }
}
```

**Implementation Steps:**

1. **Update ExecuteAgent to read tool_calls:**
```go
// In agent.go, after API call:
choice := completion.Choices[0]
message := choice.Message

// Extract from native API field
var toolCalls []ToolCall
if len(message.ToolCalls) > 0 {
    for _, tc := range message.ToolCalls {
        // Parse function arguments from JSON
        var args map[string]interface{}
        json.Unmarshal([]byte(tc.Function.Arguments), &args)

        toolCalls = append(toolCalls, ToolCall{
            ID:        tc.ID,
            ToolName:  tc.Function.Name,
            Arguments: args,
        })
    }
}

// Fallback to text parsing if no native tool_calls
if len(toolCalls) == 0 {
    toolCalls = extractToolCallsFromText(message.Content, agent)
}
```

2. **Update system prompt to use function calling syntax:**
```go
// Instead of: "Write ToolName(args) on its own line"
// Use: Tell agent to use function_calls mechanism

// When building OpenAI params, enable function calling:
params := openai.ChatCompletionNewParams{
    Model: agent.Model,
    Tools: convertToolsToOpenAIFormat(agent.Tools),  // New!
    // ... rest of params
}
```

3. **Helper function to convert tools:**
```go
func convertToolsToOpenAIFormat(tools []*Tool) []openai.Tool {
    var result []openai.Tool
    for _, tool := range tools {
        // Convert Tool struct to OpenAI Tool format
        // Map Parameters (JSON Schema) to OpenAI format
    }
    return result
}
```

**Why This Approach:**
- ✅ Uses official OpenAI API mechanism
- ✅ No parsing fragility
- ✅ Better structured data
- ✅ Fallback to text parsing for safety
- ✅ More reliable agent communication

**Files Affected:** `go-agentic/agent.go`
**Impact:** CRITICAL - Robust tool execution

**Dependencies:**
- Requires encoding/json package (already imported)
- Requires conversion of Tool structs to OpenAI format

**Acceptance Criteria:**
- [ ] Native tool_calls parsed correctly
- [ ] Fallback to text parsing works
- [ ] 50+ format variations tested
- [ ] Agent responses properly converted

---

### AD-3: Add Tool Parameter Validation (HIGH)

**Problem:** JSON Schema defined but never validated; parameters passed without type checking.

**Current State:**
```go
// Tool defined with schema:
Tool{
    Parameters: map[string]interface{}{
        "properties": map[string]interface{}{
            "host": {"type": "string"},
        },
    },
}

// But never validated when called!
```

**Decision:** Validate parameters before handler execution.

**Technical Solution:**

Create validation layer in team.go:

```go
// New function
func validateToolParameters(tool *Tool, args map[string]interface{}) error {
    if tool.Parameters == nil {
        return nil
    }

    // Get schema
    schema := tool.Parameters
    props, ok := schema["properties"].(map[string]interface{})
    if !ok {
        return nil  // No properties to validate
    }

    // Get required fields
    required := []string{}
    if req, ok := schema["required"].([]interface{}); ok {
        for _, r := range req {
            required = append(required, r.(string))
        }
    }

    // Validate each argument
    for argName, argValue := range args {
        prop, exists := props[argName].(map[string]interface{})
        if !exists {
            return fmt.Errorf("unknown parameter: %s", argName)
        }

        expectedType := prop["type"].(string)
        if !validateType(argValue, expectedType) {
            return fmt.Errorf("parameter %s: expected %s, got %T",
                argName, expectedType, argValue)
        }
    }

    // Check required fields
    for _, req := range required {
        if _, exists := args[req]; !exists {
            return fmt.Errorf("required parameter missing: %s", req)
        }
    }

    return nil
}

func validateType(value interface{}, expectedType string) bool {
    switch expectedType {
    case "string":
        _, ok := value.(string)
        return ok
    case "number":
        _, ok := value.(float64)
        return ok || isNumericString(value)
    case "integer":
        _, ok := value.(int)
        return ok || isIntegerString(value)
    case "boolean":
        _, ok := value.(bool)
        return ok
    default:
        return true  // Unknown type, allow
    }
}
```

**Integration Point (team.go):**

```go
// In executeToolCall or similar:
func (te *TeamExecutor) executeToolCall(toolCall ToolCall, agent *Agent) (string, error) {
    tool := findTool(agent, toolCall.ToolName)
    if tool == nil {
        return "", fmt.Errorf("tool not found: %s", toolCall.ToolName)
    }

    // NEW: Validate parameters
    if err := validateToolParameters(tool, toolCall.Arguments); err != nil {
        return "", fmt.Errorf("parameter validation failed: %w", err)
    }

    // Execute handler
    return tool.Handler(context.Background(), toolCall.Arguments)
}
```

**Files Affected:** `go-agentic/team.go`, `go-agentic/agent.go`
**Impact:** HIGH - Prevents runtime type errors

**Acceptance Criteria:**
- [ ] Parameters validated before execution
- [ ] Clear error messages for type mismatches
- [ ] Required fields checked
- [ ] Tests with invalid parameters pass

---

### AD-4: Cross-Platform Tool Compatibility (HIGH)

**Problem:** System tools use OS-specific flags without handling all platforms.

**Example - Ping Command (example_it_support.go:353):**
```go
// ❌ Current: -c flag only works on macOS/Linux
cmd := exec.CommandContext(ctx, "ping", "-c", count, host)
```

**Windows vs Unix Differences:**
| Operation | Windows | macOS/Linux |
|-----------|---------|-------------|
| ping | `-n` count | `-c` count |
| service check | `net start \| find` | `systemctl is-active` |
| processes | `tasklist` | `ps aux` |

**Decision:** Implement OS-aware tool wrappers.

**Technical Solution:**

Create abstraction for platform-specific commands:

```go
// In example_it_support.go - New helper
func getPingCommand(ctx context.Context, host string, count string) (*exec.Cmd, error) {
    var cmd *exec.Cmd

    switch runtime.GOOS {
    case "windows":
        cmd = exec.CommandContext(ctx, "ping", "-n", count, host)
    case "darwin", "linux":
        cmd = exec.CommandContext(ctx, "ping", "-c", count, host)
    default:
        return nil, fmt.Errorf("unsupported OS: %s", runtime.GOOS)
    }

    return cmd, nil
}

// Update handlers to use abstraction
func pingHostHandler(ctx context.Context, args map[string]interface{}) (string, error) {
    host, ok := args["host"].(string)
    if !ok {
        return "", fmt.Errorf("host parameter required")
    }

    count := "4"
    if c, ok := args["count"]; ok {
        if cs, ok := c.(string); ok {
            count = cs
        }
    }

    // NEW: Use OS-aware wrapper
    cmd, err := getPingCommand(ctx, host, count)
    if err != nil {
        return "", err
    }

    output, err := cmd.CombinedOutput()
    if err != nil {
        return "", fmt.Errorf("ping failed: %w", err)
    }

    return strings.TrimSpace(string(output)), nil
}
```

**Pattern for All System Tools:**
```go
// Template pattern
func getCommand(ctx context.Context, args map[string]interface{}) (*exec.Cmd, error) {
    switch runtime.GOOS {
    case "windows":
        return getWindowsCommand(ctx, args)
    case "darwin":
        return getMacOSCommand(ctx, args)
    case "linux":
        return getLinuxCommand(ctx, args)
    default:
        return nil, fmt.Errorf("unsupported OS: %s", runtime.GOOS)
    }
}
```

**Files Affected:** `examples/it-support/example_it_support.go`
**Impact:** HIGH - Enables cross-platform deployment

**Acceptance Criteria:**
- [ ] Same test suite passes on Windows, macOS, Linux
- [ ] No hardcoded platform-specific flags
- [ ] CI/CD runs tests on all platforms
- [ ] Commands return consistent results

---

### AD-5: Comprehensive Error Handling (HIGH)

**Problem:** Errors are silently ignored with blank identifiers `_, _`.

**Current Issues:**

```go
// ❌ example_it_support.go:370
output, _ := cmd.Output()  // Error ignored!

// ❌ example_it_support.go:378
output, _ := cmd.Output()  // Error ignored!
```

**Impact:** "Service not running" vs "command failed" indistinguishable.

**Decision:** Capture and distinguish error cases.

**Technical Solution:**

```go
// Instead of:
output, _ := cmd.Output()

// Do:
output, err := cmd.Output()
if err != nil {
    // Distinguish error types
    if exitErr, ok := err.(*exec.ExitError); ok {
        // Command ran but exited with error
        return fmt.Sprintf("Service check failed (exit %d): %s",
            exitErr.ExitCode(), string(exitErr.Stderr))
    } else {
        // Command not found or other error
        return fmt.Errorf("command execution error: %w", err)
    }
}

// Process successful output
return parseServiceStatus(string(output))
```

**Error Classification Pattern:**

```go
type ToolError struct {
    Type    string  // "not_found", "permission_denied", "timeout", "not_running"
    Message string
    Cause   error
}

func (e *ToolError) Error() string {
    return fmt.Sprintf("[%s] %s: %v", e.Type, e.Message, e.Cause)
}

// Usage:
if err != nil {
    return "", &ToolError{
        Type:    "command_failed",
        Message: "Service status check failed",
        Cause:   err,
    }
}
```

**Files Affected:** `examples/it-support/example_it_support.go`, `go-agentic/agent.go`
**Impact:** HIGH - Clear error diagnostics

**Acceptance Criteria:**
- [ ] All errors captured (no `_, _`)
- [ ] Error types categorized
- [ ] Messages distinguish root causes
- [ ] Errors propagate to user

---

### AD-6: Semantic Correctness - Message Roles (MEDIUM)

**Problem:** Tool results added to history with role `"user"` (semantically wrong).

**Current Code (team.go):**
```go
// ❌ Wrong semantic
te.history = append(te.history, Message{
    Role:    "user",  // WRONG! Tool results aren't from user
    Content: resultText,
})
```

**OpenAI Message Roles:**
- `"system"` - System instructions
- `"user"` - User input
- `"assistant"` - Agent response
- `"tool"` - Tool result (in newer API)

**Decision:** Use `"system"` role for tool results (for compatibility).

**Technical Solution:**

```go
// In team.go, after tool execution:
// Create tool result message with correct role
toolResultMsg := Message{
    Role:    "system",  // ✅ Correct semantic
    Content: fmt.Sprintf("[Tool Result] %s: %s", toolCall.ToolName, resultText),
}

te.history = append(te.history, toolResultMsg)
```

**Alternative (if OpenAI API supports "tool" role):**
```go
// If using OpenAI API v3.16+
toolResultMsg := Message{
    Role:    "tool",  // ✅ Newer, more explicit
    Content: resultText,
}
```

**Files Affected:** `go-agentic/team.go`
**Impact:** MEDIUM - Semantic correctness

**Acceptance Criteria:**
- [ ] Tool results use system or tool role
- [ ] LLM context accurate
- [ ] No behavior changes (backward compatible)

---

### AD-7: Test Framework API Alignment (MEDIUM)

**Problem:** test.go references non-existent APIs.

**Current Issues:**
- Calls `agentic.RunTestScenario()` - doesn't exist
- Calls `agentic.GetTestScenarios()` - doesn't exist
- References `agentic.HTMLReport` - unclear structure

**Decision:** Either implement test APIs or update example to use library correctly.

**Option A: Implement Test APIs in Library**

```go
// In go-agentic/tests.go (already exists)

// Export test scenario runner
func RunTestScenario(ctx context.Context, scenario *TestScenario,
    executor *TeamExecutor) *TestResult {
    // Implementation
}

// Export scenario getter
func GetTestScenarios() []*TestScenario {
    return testScenarios  // Pre-defined test scenarios
}

// Export report generator
func GenerateHTMLReport(results []*TestResult) string {
    // Generate HTML report
}
```

**Option B: Update Example (Lighter Weight)**

```go
// In example_it_support.go
// Use directly available APIs instead

executor := agentic.NewTeamExecutor(crew, apiKey)

// Manually run test cases
scenarios := []struct{
    name string
    input string
}{
    {"vague_issue", "My computer is slow"},
    {"network_issue", "Can't connect to server at 192.168.1.100"},
}

for _, scenario := range scenarios {
    response, err := executor.Execute(ctx, scenario.input)
    if err != nil {
        fmt.Printf("FAIL: %s - %v\n", scenario.name, err)
    } else {
        fmt.Printf("PASS: %s - %s\n", scenario.name, response.AgentName)
    }
}
```

**Recommendation:** Option A (implement in library) for reusability.

**Files Affected:** `go-agentic/tests.go`, `examples/it-support/test.go`
**Impact:** MEDIUM - Test infrastructure completeness

---

### AD-8: Configuration Validation (LOW)

**Problem:** Temperature override logic incorrect (config.go).

```go
// ❌ Wrong: 0 is a valid value
if config.Temperature == 0 {
    config.Temperature = 0.7  // Overrides valid 0.0 setting!
}
```

**Decision:** Only set default if not explicitly configured.

**Technical Solution:**

Use pointer for optional fields:

```go
type AgentConfig struct {
    // ...
    Temperature *float64  // Pointer: nil = not set, present = explicitly set
}

// When loading:
if config.Temperature == nil {
    defaultTemp := 0.7
    config.Temperature = &defaultTemp
}

// When using:
agent.Temperature = *config.Temperature
```

Or use explicit "IsSet" flag:

```go
type AgentConfig struct {
    Temperature float64
    TemperatureSet bool  // Was Temperature explicitly configured?
}

if !config.TemperatureSet {
    config.Temperature = 0.7
}
```

**Files Affected:** `go-agentic/config.go`
**Impact:** LOW - Configuration correctness

---

## 3. API CONTRACTS

### Key APIs (No Breaking Changes)

```go
// ExecuteAgent - signature unchanged
func ExecuteAgent(ctx context.Context, agent *Agent, input string,
    history []Message, apiKey string) (*AgentResponse, error)

// TeamExecutor.Execute - signature unchanged
func (te *TeamExecutor) Execute(ctx context.Context, input string)
    (*TeamResponse, error)

// Tool.Handler - signature unchanged
type Tool struct {
    Handler func(ctx context.Context, args map[string]interface{})
        (string, error)
}
```

### New APIs (Additions, No Breaking Changes)

```go
// Validation function - NEW
func validateToolParameters(tool *Tool, args map[string]interface{}) error

// Test APIs - NEW (if implementing tests)
func RunTestScenario(ctx context.Context, scenario *TestScenario,
    executor *TeamExecutor) *TestResult
func GetTestScenarios() []*TestScenario
func GenerateHTMLReport(results []*TestResult) string

// Helper functions - NEW
func convertToolsToOpenAIFormat(tools []*Tool) []openai.Tool
```

---

## 4. DATA MODELS (No Changes)

Core data structures remain unchanged:

```go
type Agent struct {
    ID             string
    Name           string
    Role           string
    Backstory      string
    Model          string         // ← Will be used (was ignored)
    Tools          []*Tool
    Temperature    float64
    IsTerminal     bool
    HandoffTargets []string
}

type Message struct {
    Role    string  // ← Will use "system" for tool results (was "user")
    Content string
}

type ToolCall struct {
    ID        string
    ToolName  string
    Arguments map[string]interface{}  // ← Will be validated (was not)
}
```

---

## 5. INTEGRATION POINTS

### 1. Agent Model Configuration
- **Where:** agent.go:24
- **Impact:** All agents
- **Breaking:** NO (backward compatible)

### 2. Tool Call Parsing
- **Where:** agent.go:126-190 (refactor existing function)
- **Impact:** Tool execution
- **Breaking:** NO (fallback to text parsing)

### 3. Parameter Validation
- **Where:** team.go:executeTool (new validation call)
- **Impact:** All tool handlers
- **Breaking:** NO (validation before existing code)

### 4. Cross-Platform Commands
- **Where:** example_it_support.go (refactor handlers)
- **Impact:** IT Support example
- **Breaking:** NO (same external interface)

### 5. Error Handling
- **Where:** Scattered (pattern adoption)
- **Impact:** All error paths
- **Breaking:** NO (better error messages)

### 6. Message Roles
- **Where:** team.go (history management)
- **Impact:** LLM context
- **Breaking:** NO (semantic clarity only)

---

## 6. DEPLOYMENT CONSIDERATIONS

### Configuration Changes
- No changes to YAML schema
- Existing configs continue to work
- New validation happens transparently

### Testing Strategy
- Unit tests for validation
- Integration tests for model selection
- Platform-specific tests (Windows, macOS, Linux)
- Tool execution tests with invalid parameters

### Rollout Plan
1. Fix issues in feature branch
2. All tests passing locally (3 platforms)
3. CI/CD validates (GitHub Actions)
4. Code review approval
5. Merge to main
6. Release as v0.0.2-alpha.2

### Backward Compatibility
- ✅ All existing APIs work unchanged
- ✅ All existing examples work unchanged
- ✅ All existing configs work unchanged
- ✅ Better error handling is non-breaking

---

## 7. SECURITY CONSIDERATIONS

### Command Injection Prevention
Existing safety checks for dangerous patterns:
```go
dangerousPatterns := []string{"rm -rf", "mkfs", "dd if=", ":(){:|:", "fork"}
```

**Enhancement:** Add platform-specific dangerous patterns
```go
if runtime.GOOS == "windows" {
    dangerousPatterns = append(dangerousPatterns, "del /s /q", "format C:")
}
```

### Type Validation
Parameter validation prevents unexpected types reaching handlers.

### Error Information Disclosure
Error messages are detailed but avoid sensitive info (no passwords in error messages).

---

## 8. PERFORMANCE CONSIDERATIONS

### Tool Call Parsing
- Native OpenAI API: O(1) lookup
- Text fallback: O(n) where n = agent tools
- No significant performance impact

### Parameter Validation
- JSON Schema validation: O(p) where p = parameters
- Typical: 1-3 parameters per tool
- Negligible overhead

### Overall Impact
- Expected: <1ms additional overhead per tool call
- Acceptable for orchestration workload

---

## 9. MONITORING & OBSERVABILITY

### What to Monitor
- Model selection (ensure configured model used)
- Tool call success rate (catch parsing issues)
- Parameter validation errors (catch invalid inputs)
- Cross-platform execution (consistency)

### Metrics to Track
- Tool execution success rate
- Average tool call latency
- Validation error rate
- Platform-specific error patterns

### Logging Points
- Agent execution start/end (with model used)
- Tool call parsing (native vs fallback)
- Parameter validation failures
- Command execution errors (with exit codes)

---

## 10. TECHNICAL DEBT & FUTURE

### Addressed
- Hardcoded model (CRITICAL)
- Fragile tool parsing (CRITICAL)
- Missing validation (HIGH)
- Error handling gaps (HIGH)
- Cross-platform (HIGH)

### Remaining Technical Debt (Future)
- Complete OpenAI SDK v3 migration (already started)
- Performance optimization for large tool sets
- Sandboxing for untrusted tool execution
- Distributed orchestration (multi-node agents)

---

**Document Status:** Ready for Epic Design
**Next Steps:** Create epics and user stories based on this architecture


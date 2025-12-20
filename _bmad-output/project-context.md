---
project_name: 'go-agentic'
user_name: 'taipm'
date: '2025-12-20'
sections_completed: []
---

# Project Context for AI Agents

_This file contains critical rules and patterns that AI agents must follow when implementing code in this project. Focus on unobvious details that agents might otherwise miss._

---

## Technology Stack & Versions

**Language:** Go 1.25.5

**Core Dependencies:**
- `github.com/openai/openai-go/v3` v3.15.0 - OpenAI API client (CRITICAL)
- `gopkg.in/yaml.v3` v3.0.1 - YAML configuration parsing
- `github.com/tidwall/gjson` v1.18.0 - JSON path navigation
- `github.com/tidwall/sjson` v1.2.5 - JSON modification
- `github.com/tidwall/match` v1.2.0 - Pattern matching utilities
- `github.com/tidwall/pretty` v1.2.1 - JSON formatting

**Project Type:** Multi-agent orchestration library

**Package Structure:**
- Core library: `github.com/taipm/go-agentic/go-agentic`
- Examples: `examples/` directory with 4 implementations (it-support, research-assistant, customer-service, data-analysis)

---

## Critical Implementation Rules

### 1. Go Language-Specific Rules

#### Error Handling (CRITICAL - Epic 3 Focus)
- ❌ **NEVER** use `_, _ = cmd.Output()` - this ignores errors silently
- ✅ **ALWAYS** capture error returns: `output, err := cmd.Output()`
- ✅ **ALWAYS** check errors before using results
- ✅ **ALWAYS** wrap errors with context: `fmt.Errorf("operation failed: %w", err)`
- ✅ Use named error types (ToolError) for categorization

**Example - WRONG:**
```go
output, _ := cmd.Output()  // ❌ Error silently ignored!
return string(output)
```

**Example - CORRECT:**
```go
output, err := cmd.Output()  // ✅ Error captured
if err != nil {
    return "", fmt.Errorf("[COMMAND_FAILED] %w", err)
}
return string(output), nil
```

#### Context Usage (CRITICAL - Always Required)
- ✅ **ALWAYS** include `context.Context` as first parameter in functions
- ✅ **ALWAYS** pass context to child function calls
- ✅ **ALWAYS** use context timeouts for external operations (OpenAI API, system commands)
- ✅ Example: `ExecuteAgent(ctx context.Context, agent *Agent, ...)`

#### Package Structure
- All library code in `go-agentic/` package
- Examples in `examples/` with `main` package
- No subpackages - keep library flat for simplicity
- Exported functions: PascalCase (ExecuteAgent, NewTeamExecutor)
- Unexported functions: camelCase (buildSystemPrompt, extractToolCalls)

#### Type Definitions
- All types in `types.go` for consistency
- Struct fields always PascalCase and exported (Agent, Tool, Message)
- Define JSON tag for struct fields if needed: `json:"field_name"`

---

### 2. OpenAI API Integration (CRITICAL - Epic 1, 2a Focus)

#### Model Configuration (Epic 1 - Story 1.1)
- ❌ **NEVER** hardcode "gpt-4o-mini" in agent.go:24
- ✅ **ALWAYS** use `agent.Model` field from configuration
- ✅ **ALWAYS** validate model name is not empty before API call
- ✅ **ALWAYS** log which model is being used: `log.Printf("[INFO] Agent %s using model %s", agent.Name, agent.Model)`

**File Location:** go-agentic/agent.go, line 23-26 (ExecuteAgent function)

**Required Change:**
```go
// BEFORE (BROKEN):
params := openai.ChatCompletionNewParams{
    Model: "gpt-4o-mini",  // ❌ Hardcoded
}

// AFTER (CORRECT):
params := openai.ChatCompletionNewParams{
    Model: agent.Model,  // ✅ Use config
}
```

#### Temperature Configuration (Epic 1 - Story 1.2)
- ❌ **NEVER** override Temperature=0.0 to 0.7
- ✅ **ALWAYS** respect all valid temperature values (0.0-2.0)
- ✅ **ALWAYS** validate temperature is within bounds before API call
- ✅ Temperature 0.0 = deterministic (perfect for repeatability)
- ✅ Temperature 1.0 = balanced
- ✅ Temperature 2.0 = maximum creativity

**File Location:** go-agentic/config.go (temperature override logic)

#### Tool Calls Extraction (Epic 2a - Story 2a.1, 2a.2)
- ❌ **NEVER** rely only on text-based regex parsing
- ✅ **ALWAYS** try native OpenAI `tool_calls` API first (message.ToolCalls)
- ✅ **ALWAYS** provide text parsing fallback for compatibility
- ✅ **ALWAYS** validate parsed arguments are valid JSON
- ✅ **ALWAYS** update system prompts to request function calling format

**File Location:** go-agentic/agent.go, lines 44-46 (extractToolCallsFromText call)

**Required Implementation Pattern:**
```go
var toolCalls []ToolCall

// Try native API first
if len(message.ToolCalls) > 0 {
    toolCalls = parseNativeToolCalls(message.ToolCalls)
}

// Fallback to text parsing if needed
if len(toolCalls) == 0 {
    toolCalls = extractToolCallsFromText(content, agent)
}
```

---

### 3. Tool Execution & Validation (Epic 2b, 3 Focus)

#### Parameter Validation (Epic 2b - Story 2b.1, 2b.2)
- ❌ **NEVER** pass parameters to tool handlers without validation
- ✅ **ALWAYS** validate parameters against JSON Schema before handler execution
- ✅ **ALWAYS** check required fields exist
- ✅ **ALWAYS** validate parameter types match schema (string, number, integer, boolean)
- ✅ **ALWAYS** return clear error messages for validation failures

**File Location:** go-agentic/team.go (tool execution loop)

**Validation Pattern:**
```go
// Before executing tool handler:
if err := validateToolParameters(tool, args); err != nil {
    return "", fmt.Errorf("parameter validation failed: %w", err)
}
// Now safe to call handler
result, err := tool.Handler(ctx, args)
```

#### Error Categorization (Epic 3 - Story 3.1, 3.2)
- ✅ **ALWAYS** categorize errors with types:
  - `PERMISSION_DENIED` - access/privilege issues
  - `TIMEOUT` - operation exceeded time limit
  - `NOT_FOUND` - command/file/service not found
  - `COMMAND_FAILED` - non-zero exit code
  - `INVALID_INPUT` - parameter validation failed
- ✅ **ALWAYS** provide error messages with:
  - Error type: `[PERMISSION_DENIED]`
  - Description: what went wrong
  - Suggestion: how to fix it

**Error Message Format:**
```
[ERROR_TYPE] High-level problem
Context: Where it happened
Reason: Why it happened
Action: How to fix it
```

#### Message Role Semantics (Epic 3 - Story 3.3)
- ❌ **NEVER** use role="user" for tool results
- ✅ **ALWAYS** use role="system" for tool results
- ✅ **ALWAYS** use role="assistant" for agent responses
- ✅ **ALWAYS** use role="user" for user messages only

**File Location:** go-agentic/team.go (message history management)

**Correct Pattern:**
```go
// Tool result goes back as system message:
history = append(history, Message{
    Role:    "system",  // ✅ NOT "user"!
    Content: fmt.Sprintf("[Tool Result] %s returned: %s", toolName, result),
})
```

---

### 4. Cross-Platform Compatibility (Epic 4 Focus)

#### OS Detection (Critical - Story 4.1, 4.2)
- ✅ **ALWAYS** use `runtime.GOOS` to detect operating system
- ✅ **ALWAYS** handle Windows, macOS, and Linux separately when needed
- ✅ **ALWAYS** test on all 3 platforms (Windows-latest, macos-latest, ubuntu-latest in CI/CD)

#### Platform-Specific Commands

**Ping Command (Currently broken on Windows):**
- ❌ **NEVER** hardcode `-c` flag (Unix only)
- ✅ **ALWAYS** use `-n` on Windows, `-c` on macOS/Linux

**File Location:** examples/it-support/example_it_support.go, pingHostHandler

**Correct Pattern:**
```go
func getPingCommand(ctx context.Context, host string, count string) (*exec.Cmd, error) {
    switch runtime.GOOS {
    case "windows":
        return exec.CommandContext(ctx, "ping", "-n", count, host), nil
    case "darwin", "linux":
        return exec.CommandContext(ctx, "ping", "-c", count, host), nil
    default:
        return nil, fmt.Errorf("unsupported OS: %s", runtime.GOOS)
    }
}
```

**Service Status Checking:**
- Windows: Use `net start` or `Get-Service` PowerShell
- macOS: Use `launchctl list | grep service`
- Linux: Use `systemctl is-active service`

**Process Listing:**
- Windows: `tasklist` command
- macOS/Linux: `ps` command with different flags

---

### 5. Testing Rules (Epic 5, 6 Focus)

#### Test Scenario Design
- ✅ **ALWAYS** define TestScenario with ID, Name, UserInput, ExpectedFlow
- ✅ **ALWAYS** test complete workflows (agent flow, tool calls, results)
- ✅ **ALWAYS** verify correct agent models are used in execution
- ✅ **ALWAYS** verify tool calls are extracted and executed
- ✅ **ALWAYS** verify error messages are clear and actionable

#### Test File Organization
- Test functions: `func Test*` pattern (TestModelConfiguration, TestToolParsing)
- Test scenarios: Defined as package-level variables or constants
- Unit tests: For individual functions (validateParameters, parseToolCalls)
- Integration tests: For complete workflows (ExecuteAgent → tool execution → response)
- Platform tests: Separate test suites for Windows, macOS, Linux

#### Coverage Requirements
- ✅ **Target:** >90% code coverage across all packages
- ✅ **ALWAYS** test both success and error paths
- ✅ **ALWAYS** test edge cases (empty strings, null values, type mismatches)
- ✅ **ALWAYS** test cross-platform compatibility

---

### 6. Code Quality & Documentation Rules

#### Naming Conventions
- Files: `lowercase_underscore.go` (agent.go, team.go, config.go)
- Functions: `PascalCase` for exported (ExecuteAgent, NewTeamExecutor), `camelCase` for unexported
- Variables: `camelCase` local variables, `PascalCase` for struct fields
- Constants: `UPPERCASE_UNDERSCORE` for exported constants
- Agents: Named by role (Orchestrator, Clarifier, Executor)
- Tools: Named as verbs (GetCPUUsage, PingHost, CheckServiceStatus)

#### Comments & Documentation
- ✅ **ALWAYS** add comments for exported functions/types
- ✅ **ALWAYS** explain WHY code does something, not just WHAT
- ✅ **ALWAYS** document parameters and return values
- ✅ **ALWAYS** document error conditions and edge cases

**Comment Pattern:**
```go
// ExecuteAgent runs an agent with the given input and history,
// respecting the agent's configured model and tools.
// Returns the agent's response with any tool calls it made.
func ExecuteAgent(ctx context.Context, agent *Agent, ...) (*AgentResponse, error)
```

#### Code Organization
- Core library in `go-agentic/` package
- All types in single `types.go` file
- Separate files by responsibility: agent.go, team.go, config.go, etc.
- Keep functions focused and small (<100 lines when possible)
- Group related functions together

---

### 7. Backward Compatibility Rules (Critical)

- ✅ **ALWAYS** maintain existing API signatures (don't break function names/parameters)
- ✅ **ALWAYS** support v0.0.1-alpha.1 examples with v0.0.2 library
- ✅ **ALWAYS** add new fields as optional (with sensible defaults)
- ✅ **ALWAYS** test that existing examples still work unchanged
- ✅ **ALWAYS** add new APIs without removing old ones

**Example - CORRECT Approach:**
```go
// Old API still works:
type Agent struct {
    Model string  // ← Previously ignored, now honored
}

// New optional fields have defaults:
type Agent struct {
    Model string
    // ... existing fields ...
    MaxRetries int  // ← New optional field (default 0)
}
```

---

### 8. Configuration & Loading Rules

#### YAML Configuration Parsing
- ✅ **ALWAYS** use `gopkg.in/yaml.v3` for parsing
- ✅ **ALWAYS** validate configuration after loading
- ✅ **ALWAYS** provide clear error messages for invalid config
- ✅ **ALWAYS** use sensible defaults for optional fields

#### Agent Configuration Validation
- ✅ Agent.ID must not be empty
- ✅ Agent.Name must not be empty
- ✅ Agent.Model must be a valid OpenAI model string (not empty)
- ✅ Agent.Temperature must be between 0.0 and 2.0
- ✅ Tool names must be unique within agent

---

## Critical Don't-Miss Rules (Anti-Patterns)

### ❌ DO NOT Commit These Mistakes:

1. **Silent Error Handling**
   ```go
   // ❌ WRONG - Error ignored:
   _, _ = cmd.Output()

   // ✅ CORRECT:
   output, err := cmd.Output()
   if err != nil { /* handle */ }
   ```

2. **Hardcoded Configuration**
   ```go
   // ❌ WRONG - Hardcoded model:
   Model: "gpt-4o-mini",

   // ✅ CORRECT:
   Model: agent.Model,
   ```

3. **Platform-Specific Code Without Detection**
   ```go
   // ❌ WRONG - Unix-only flag:
   cmd := exec.Command("ping", "-c", host)

   // ✅ CORRECT:
   args := []string{"ping"}
   if runtime.GOOS == "windows" {
       args = append(args, "-n")
   } else {
       args = append(args, "-c")
   }
   ```

4. **Missing Parameter Validation**
   ```go
   // ❌ WRONG - No validation:
   result, _ := tool.Handler(ctx, args)

   // ✅ CORRECT:
   if err := validateToolParameters(tool, args); err != nil {
       return "", err
   }
   result, _ := tool.Handler(ctx, args)
   ```

5. **Wrong Message Role Semantics**
   ```go
   // ❌ WRONG - Tool results as user:
   Message{Role: "user", Content: result}

   // ✅ CORRECT:
   Message{Role: "system", Content: result}
   ```

6. **Type Assertion Without Checking**
   ```go
   // ❌ WRONG - Will panic:
   host := args["host"].(string)

   // ✅ CORRECT:
   host, ok := args["host"].(string)
   if !ok {
       return "", fmt.Errorf("parameter 'host': expected string")
   }
   ```

---

## Epic-Specific Constraints

### Epic 1: Configuration Integrity & Trust
- **Constraint:** Agent.Model must be used, never hardcoded
- **Constraint:** Temperature must allow 0.0 values
- **Constraint:** All config changes must be backward compatible
- **Test:** Verify config is respected in actual OpenAI API calls

### Epic 2a: Native Tool Call Parsing
- **Constraint:** Use native API first, text parsing only as fallback
- **Constraint:** Fallback parser must handle 50+ format variations
- **Constraint:** Performance impact <100ms overhead

### Epic 2b: Parameter Validation
- **Constraint:** Validate before handler execution (not inside)
- **Constraint:** Use JSON Schema from Tool.Parameters
- **Constraint:** Clear error messages for each type of validation failure

### Epic 3: Clear & Actionable Error Handling
- **Constraint:** No silent error ignoring (`_, _` patterns)
- **Constraint:** All errors must have type, reason, and action
- **Constraint:** Tool results must use "system" role

### Epic 4: Cross-Platform Compatibility
- **Constraint:** Same test suite must pass on all 3 platforms
- **Constraint:** No platform-specific logic without `runtime.GOOS` check
- **Constraint:** CI/CD must run tests on Windows, macOS, Linux

### Epic 5: Production-Ready Testing Framework
- **Constraint:** Test APIs must be exported (RunTestScenario, GetTestScenarios)
- **Constraint:** Coverage must reach >90%
- **Constraint:** HTML reports must be human-readable

---

## Enhanced Constraints from Party Mode Review

### Epic 1 Enhanced: Configuration Integrity & Trust
- **Temperature Validation:** Must be validated at config load time (0.0 ≤ temp ≤ 2.0)
- **Model Validation:** Agent.Model must be validated against OpenAI valid models list
- **Validation Location:** config.go LoadConfig() function
- **Test Cases Required:**
  - Boundary tests: Temperature 0.0, 1.0, 2.0 (valid)
  - Invalid tests: Temperature 2.1, -1.0 (invalid)
  - Invalid model string handling and error messages

### Epic 2a Enhanced: Native Tool Call Parsing
- **OpenAI Library Integration:** Handle `openai.ChatCompletionMessageToolCall` struct conversion
- **Tool Call Format Variations to Support (50+ including):**
  - `ToolName()` - no arguments
  - `ToolName(arg1)` - single argument
  - `ToolName("string", 123)` - multiple types
  - `ToolName(key=value)` - named arguments
  - Whitespace variations: `ToolName ( arg )`, `ToolName( arg)`, etc.
  - Create comprehensive test suite with documented variations
- **System Prompt Update:** Must request `function_calls` format from OpenAI API
- **Performance Benchmark:** Must measure <100ms overhead with `go test -bench`

### Epic 2b Enhanced: Parameter Validation
- **Validation Approach:** Custom implementation (no external JSON Schema library)
- **Validation Function Signature:** `validateToolParameters(tool *Tool, args map[string]interface{}) error`
- **Type Validation:** Must support schema types: string, number, integer, boolean
- **Required Fields:** Must check existence of required fields from schema
- **Error Messages:** Clear messages with field name, expected type, received type

### Epic 3 Enhanced: Clear & Actionable Error Handling
- **ToolError Type Definition:** Must have fields: Type, Message, Cause, Suggestion
- **Error Types to Implement:**
  - `PERMISSION_DENIED` - access/privilege issues
  - `TIMEOUT` - operation exceeded time limit
  - `NOT_FOUND` - command/file/service not found
  - `COMMAND_FAILED` - non-zero exit code
  - `INVALID_INPUT` - parameter validation failed
- **Error Message Format:** `[ERROR_TYPE] Description\nSuggestion: how to fix it`
- **Test Scenarios:**
  - Service check when permission denied → [PERMISSION_DENIED]
  - Ping timeout (if possible in test) → [TIMEOUT]
  - Command not found → [NOT_FOUND]
  - Exit code non-zero → [COMMAND_FAILED]
  - Parameter type mismatch → [INVALID_INPUT]

### Epic 4 Enhanced: Cross-Platform Compatibility
- **CI/CD Platforms:** Must test on `windows-latest`, `macos-latest`, `ubuntu-latest`
- **Command Variants (Centralized Wrappers):**
  - Ping: `-n` count (Windows) vs `-c` count (macOS/Linux)
  - Service Status: `net start` (Windows) vs `systemctl is-active` (Linux) vs `launchctl list` (macOS)
  - Process List: `tasklist` (Windows) vs `ps aux` (macOS/Linux)
- **OS Detection Pattern:** Always use `runtime.GOOS` for platform decisions
- **Test Strategy:** Platform-specific test suites (not skipped/xfail, but run natively)

### Epic 5 Enhanced: Production-Ready Testing Framework
- **Test Scenario Coverage:** Minimum 1 per epic, ideally 2-3 for comprehensive coverage
- **Exported APIs:** RunTestScenario, GetTestScenarios, GenerateHTMLReport
- **HTML Report Contents:** Summary, per-test results with status, overall coverage %
- **Coverage Measurement:** `go test -cover` tool for metrics
- **Test File Organization:**
  - `tests.go` - Test APIs and scenario definitions
  - `*_test.go` - Unit tests for each component
  - Separate platform-specific test files as needed

---

## Implementation Reference Patterns

### Error Handling Pattern (REQUIRED - Epic 3)
```go
// Correct pattern for command execution:
output, err := cmd.Output()  // Don't use _, _
if err != nil {
    if exitErr, ok := err.(*exec.ExitError); ok {
        // Handle specific exit code error
        return "", &ToolError{
            Type:       "COMMAND_FAILED",
            Message:    "Command exited with status " + strconv.Itoa(exitErr.ExitCode()),
            Cause:      err,
            Suggestion: "Check command syntax or service status",
        }
    }
    // Handle other errors
    return "", fmt.Errorf("[ERROR_TYPE] %w", err)
}
return string(output), nil
```

### OS-Aware Command Wrapper Pattern (REQUIRED - Epic 4)
```go
// Create helper function that returns proper command per OS:
func getPingCommand(ctx context.Context, host string, count string) (*exec.Cmd, error) {
    switch runtime.GOOS {
    case "windows":
        return exec.CommandContext(ctx, "ping", "-n", count, host), nil
    case "darwin", "linux":
        return exec.CommandContext(ctx, "ping", "-c", count, host), nil
    default:
        return nil, fmt.Errorf("[UNSUPPORTED_OS] OS %s not supported", runtime.GOOS)
    }
}
```

### Parameter Validation Pattern (REQUIRED - Epic 2b)
```go
// Validation function should:
// 1. Check required fields exist
// 2. Check types match schema
// 3. Return clear error with field name and expected type
func validateToolParameters(tool *Tool, args map[string]interface{}) error {
    // Check each required parameter
    // Validate types match Tool.Parameters schema
    // Return descriptive errors with field name and expected type
    // Example error: "parameter 'host': expected string, got int"
}
```

### Message Role Semantics Pattern (REQUIRED - Epic 3)
```go
// CORRECT - Tool results as system message:
history = append(history, Message{
    Role:    "system",  // ✅ Not "user"!
    Content: fmt.Sprintf("[Tool Result] %s returned: %s", toolName, result),
})

// WRONG - This breaks LLM context:
history = append(history, Message{
    Role:    "user",  // ❌ WRONG!
    Content: result,
})
```

### Config Validation Pattern (REQUIRED - Epic 1)
```go
// In config.go LoadConfig() function:
// 1. Validate Agent.Model is not empty
// 2. Validate Agent.Temperature is 0.0-2.0
// 3. Return clear validation error if invalid
// 4. Log which agents were loaded with which models
```

---

## Development Workflow

### Pre-Implementation Checklist
- [ ] Read relevant epic in epics.md for detailed story requirements
- [ ] Review corresponding project-context constraints for this epic
- [ ] Check if story depends on other stories (in-epic dependencies)
- [ ] Identify all files that will be modified
- [ ] Plan test cases for each acceptance criterion
- [ ] Review code examples in project-context for patterns

### Code Review Checklist
- [ ] All error handling uses proper error wrapping (`%w` format)
- [ ] No `_, _` error ignoring patterns remain
- [ ] Platform-specific code uses `runtime.GOOS` detection
- [ ] Configuration values are never hardcoded (use struct fields)
- [ ] Tool results use correct message role ("system", not "user")
- [ ] Parameters are validated BEFORE handler execution
- [ ] Test coverage >90% for modified code
- [ ] Tests pass on all platforms (or marked as platform-specific)
- [ ] Documentation and comments updated for new APIs
- [ ] Backward compatibility verified (existing code still works)

### Release Validation Checklist (Before v0.0.2-alpha.2)
- [ ] All 7 epics completed
- [ ] All 28 stories implemented and tests passing
- [ ] >90% overall code coverage achieved
- [ ] All tests pass on Windows, macOS, Linux
- [ ] Backward compatibility verified (v0.0.1 examples still work)
- [ ] Performance benchmarks meet <100ms targets (Epic 2a)
- [ ] Error messages reviewed for clarity and usefulness
- [ ] All exported APIs tested and documented
- [ ] Documentation updated with fixes and migration notes

---

## Library Integration Notes

**OpenAI API (v3.15.0):**
- Use `openai.NewClient()` to create API client
- Use `client.Chat.Completions.New()` for chat API calls
- Handle `openai.ChatCompletionMessageToolCall` struct in native parsing
- Requires API key via environment or option

**YAML Configuration (v3.0.1):**
- Use `yaml.Unmarshal()` for parsing YAML files
- Validate after unmarshaling before using values

**Context (Go standard):**
- Always pass context to API calls
- Use `context.WithTimeout()` for operations that might hang
- Check `ctx.Done()` for cancellation

---


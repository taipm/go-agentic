# Implementation Plan: Tool Declaration & Usage Improvements

**Date:** 2025-12-25
**Objective:** Reduce boilerplate, eliminate type coercion bugs, and improve developer experience when declaring and using tools
**Timeline:** 2-3 days (12-18 hours total work)
**Risk:** LOW (backward compatible, no breaking changes)

---

## Overview: What We're Building

| Component | Purpose | Files | Effort |
|-----------|---------|-------|--------|
| **Quick Win #1** | Reusable type coercion utilities | `coercion.go` | 30 min |
| **Quick Win #2** | Tool schema validation | `validation.go` | 45 min |
| **Quick Win #3** | Per-tool timeout support | Update `types.go`, `executor.go` | 30 min |
| **Opportunity #1** | Tool builder pattern | `builder.go` | 2-3 hours |
| **Opportunity #2** | Schema auto-generation | `struct_schema.go` | 2-3 hours |
| **Testing & Docs** | Unit tests + examples | `*_test.go`, `examples/` | 4-6 hours |

---

# PHASE 1: QUICK WINS (2-3 Hours)

## Quick Win #1: Type Coercion Utility (30 minutes)

### Goal
Replace repetitive type assertions across tools with reusable utility functions.

### Why This Matters
- Current code in `examples/01-quiz-exam/internal/tools.go:370-380` has manual type switches
- Same pattern repeated 5+ times in project
- Error-prone (easy to miss edge cases like `float32`, `bool`)
- Boilerplate that obscures actual logic

### Step 1.1: Create `core/tools/coercion.go`

**File:** `/Users/taipm/GitHub/go-agentic/core/tools/coercion.go`

```go
package tools

import (
	"fmt"
	"strconv"
	"strings"
)

// CoerceToString converts various types to string safely.
// Handles JSON number types (float64), integers, booleans, and nil.
func CoerceToString(v interface{}) (string, error) {
	switch val := v.(type) {
	case string:
		return val, nil
	case float64:
		// JSON numbers come as float64
		// Check if it's actually an integer
		if val == float64(int64(val)) {
			return fmt.Sprintf("%d", int64(val)), nil
		}
		return fmt.Sprintf("%v", val), nil
	case int64:
		return fmt.Sprintf("%d", val), nil
	case int:
		return fmt.Sprintf("%d", val), nil
	case int32:
		return fmt.Sprintf("%d", val), nil
	case bool:
		return fmt.Sprintf("%v", val), nil
	case nil:
		return "", fmt.Errorf("cannot coerce nil to string")
	default:
		// Fallback: convert using fmt
		return fmt.Sprintf("%v", val), nil
	}
}

// CoerceToInt converts various types to int safely.
// Handles JSON numbers (float64), string integers, and other integer types.
func CoerceToInt(v interface{}) (int, error) {
	switch val := v.(type) {
	case float64:
		return int(val), nil
	case int64:
		return int(val), nil
	case int32:
		return int(val), nil
	case int:
		return val, nil
	case string:
		// Try to parse string as integer
		var i int
		_, err := fmt.Sscanf(val, "%d", &i)
		if err != nil {
			return 0, fmt.Errorf("cannot parse %q as integer: %w", val, err)
		}
		return i, nil
	case nil:
		return 0, fmt.Errorf("cannot coerce nil to int")
	default:
		return 0, fmt.Errorf("cannot coerce %T to int", v)
	}
}

// CoerceToBool converts various types to bool safely.
// Handles booleans, strings ("true"/"false"), and numeric values.
func CoerceToBool(v interface{}) (bool, error) {
	switch val := v.(type) {
	case bool:
		return val, nil
	case string:
		// Accept common boolean string representations
		switch strings.ToLower(strings.TrimSpace(val)) {
		case "true", "yes", "1", "on", "enabled":
			return true, nil
		case "false", "no", "0", "off", "disabled":
			return false, nil
		default:
			return false, fmt.Errorf("cannot parse %q as boolean", val)
		}
	case float64:
		return val != 0, nil
	case int, int32, int64:
		return val != 0, nil
	case nil:
		return false, fmt.Errorf("cannot coerce nil to bool")
	default:
		return false, fmt.Errorf("cannot coerce %T to bool", v)
	}
}

// CoerceToFloat converts various types to float64 safely.
func CoerceToFloat(v interface{}) (float64, error) {
	switch val := v.(type) {
	case float64:
		return val, nil
	case float32:
		return float64(val), nil
	case int, int32, int64:
		return float64(val.(int64)), nil
	case string:
		f, err := strconv.ParseFloat(val, 64)
		if err != nil {
			return 0, fmt.Errorf("cannot parse %q as float: %w", val, err)
		}
		return f, nil
	case nil:
		return 0, fmt.Errorf("cannot coerce nil to float")
	default:
		return 0, fmt.Errorf("cannot coerce %T to float", v)
	}
}

// MustGetString extracts and coerces a required string parameter.
// Returns error if parameter missing or cannot be coerced.
func MustGetString(args map[string]interface{}, key string) (string, error) {
	val, exists := args[key]
	if !exists {
		return "", fmt.Errorf("required parameter %q missing", key)
	}

	result, err := CoerceToString(val)
	if err != nil {
		return "", fmt.Errorf("parameter %q: %w", key, err)
	}
	return result, nil
}

// MustGetInt extracts and coerces a required int parameter.
func MustGetInt(args map[string]interface{}, key string) (int, error) {
	val, exists := args[key]
	if !exists {
		return 0, fmt.Errorf("required parameter %q missing", key)
	}

	result, err := CoerceToInt(val)
	if err != nil {
		return 0, fmt.Errorf("parameter %q: %w", key, err)
	}
	return result, nil
}

// MustGetBool extracts and coerces a required bool parameter.
func MustGetBool(args map[string]interface{}, key string) (bool, error) {
	val, exists := args[key]
	if !exists {
		return false, fmt.Errorf("required parameter %q missing", key)
	}

	result, err := CoerceToBool(val)
	if err != nil {
		return false, fmt.Errorf("parameter %q: %w", key, err)
	}
	return result, nil
}

// OptionalGetString extracts an optional string parameter with default value.
// Returns default if parameter missing or cannot be coerced.
func OptionalGetString(args map[string]interface{}, key, defaultVal string) string {
	val, exists := args[key]
	if !exists {
		return defaultVal
	}

	result, err := CoerceToString(val)
	if err != nil {
		return defaultVal
	}
	return result
}

// OptionalGetInt extracts an optional int parameter with default value.
func OptionalGetInt(args map[string]interface{}, key string, defaultVal int) int {
	val, exists := args[key]
	if !exists {
		return defaultVal
	}

	result, err := CoerceToInt(val)
	if err != nil {
		return defaultVal
	}
	return result
}

// OptionalGetBool extracts an optional bool parameter with default value.
func OptionalGetBool(args map[string]interface{}, key string, defaultVal bool) bool {
	val, exists := args[key]
	if !exists {
		return defaultVal
	}

	result, err := CoerceToBool(val)
	if err != nil {
		return defaultVal
	}
	return result
}
```

### Step 1.2: Create unit tests for coercion

**File:** `/Users/taipm/GitHub/go-agentic/core/tools/coercion_test.go`

```go
package tools

import (
	"testing"
)

func TestCoerceToString(t *testing.T) {
	tests := []struct {
		name    string
		input   interface{}
		want    string
		wantErr bool
	}{
		{"string", "hello", "hello", false},
		{"float64 int", 42.0, "42", false},
		{"float64 decimal", 3.14, "3.14", false},
		{"int64", int64(100), "100", false},
		{"int", 50, "50", false},
		{"bool true", true, "true", false},
		{"bool false", false, "false", false},
		{"nil", nil, "", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := CoerceToString(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("CoerceToString() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("CoerceToString() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestCoerceToInt(t *testing.T) {
	tests := []struct {
		name    string
		input   interface{}
		want    int
		wantErr bool
	}{
		{"float64", 42.0, 42, false},
		{"int64", int64(100), 100, false},
		{"int", 50, 50, false},
		{"string", "123", 123, false},
		{"invalid string", "not a number", 0, true},
		{"nil", nil, 0, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := CoerceToInt(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("CoerceToInt() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("CoerceToInt() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestCoerceToBool(t *testing.T) {
	tests := []struct {
		name    string
		input   interface{}
		want    bool
		wantErr bool
	}{
		{"bool true", true, true, false},
		{"bool false", false, false, false},
		{"string true", "true", true, false},
		{"string false", "false", false, false},
		{"string yes", "yes", true, false},
		{"string no", "no", false, false},
		{"int 1", 1, true, false},
		{"int 0", 0, false, false},
		{"float 1.0", 1.0, true, false},
		{"float 0.0", 0.0, false, false},
		{"invalid string", "maybe", false, true},
		{"nil", nil, false, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := CoerceToBool(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("CoerceToBool() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("CoerceToBool() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestMustGetString(t *testing.T) {
	args := map[string]interface{}{"name": "John", "age": 30}

	got, err := MustGetString(args, "name")
	if err != nil {
		t.Errorf("MustGetString() error = %v, want nil", err)
	}
	if got != "John" {
		t.Errorf("MustGetString() = %v, want John", got)
	}

	_, err = MustGetString(args, "missing")
	if err == nil {
		t.Errorf("MustGetString() with missing key should error")
	}
}

func TestOptionalGetString(t *testing.T) {
	args := map[string]interface{}{"name": "John"}

	got := OptionalGetString(args, "name", "default")
	if got != "John" {
		t.Errorf("OptionalGetString() = %v, want John", got)
	}

	got = OptionalGetString(args, "missing", "default")
	if got != "default" {
		t.Errorf("OptionalGetString() = %v, want default", got)
	}
}
```

### Step 1.3: Refactor existing code to use utilities

**File:** `/Users/taipm/GitHub/go-agentic/examples/01-quiz-exam/internal/tools.go`

Find this code (around line 370-380):
```go
var studentAnswer string
switch v := args["student_answer"].(type) {
case string:
    studentAnswer = v
case float64:
    studentAnswer = fmt.Sprintf("%v", v)
case int64:
    studentAnswer = fmt.Sprintf("%d", v)
case int:
    studentAnswer = fmt.Sprintf("%d", v)
default:
    studentAnswer = fmt.Sprintf("%v", v)
}
```

Replace with:
```go
studentAnswer, err := tools.MustGetString(args, "student_answer")
if err != nil {
    return fmt.Sprintf(`{"error": "invalid student_answer: %s"}`, err.Error()), nil
}
```

### Step 1.4: Run tests to verify

```bash
cd /Users/taipm/GitHub/go-agentic
go test ./core/tools -v -run Coerce
go test ./examples/01-quiz-exam -v
```

**Expected output:** All tests pass, no regressions

---

## Quick Win #2: Tool Schema Validation (45 minutes)

### Goal
Validate tool schemas at load time to catch configuration errors early.

### Why This Matters
- Current: Tool definition errors only discovered at runtime
- Better: Catch errors immediately when executor loads
- Benefit: Clear error messages instead of silent failures

### Step 2.1: Create `core/tools/validation.go`

**File:** `/Users/taipm/GitHub/go-agentic/core/tools/validation.go`

```go
package tools

import (
	"fmt"

	agenticcore "github.com/taipm/go-agentic/core"
)

// ValidateToolSchema performs comprehensive validation of a tool definition.
// Checks:
// - Name is not empty
// - Description is not empty
// - Handler function is not nil
// - Parameters structure (if defined) is valid
func ValidateToolSchema(tool *agenticcore.Tool) error {
	if tool == nil {
		return fmt.Errorf("tool cannot be nil")
	}

	if tool.Name == "" {
		return fmt.Errorf("tool name cannot be empty")
	}

	if tool.Description == "" {
		return fmt.Errorf("tool %q: description cannot be empty", tool.Name)
	}

	if tool.Func == nil {
		return fmt.Errorf("tool %q: handler function cannot be nil", tool.Name)
	}

	// Validate Parameters structure if defined
	if tool.Parameters != nil {
		if err := validateParameters(tool.Name, tool.Parameters); err != nil {
			return err
		}
	}

	return nil
}

// validateParameters validates the structure of a tool's Parameters field.
func validateParameters(toolName string, params map[string]interface{}) error {
	// Check that Parameters has a "type" field
	paramType, exists := params["type"].(string)
	if !exists {
		return fmt.Errorf("tool %q: Parameters.type field missing (should be 'object')", toolName)
	}

	if paramType != "object" {
		return fmt.Errorf(
			"tool %q: Parameters.type must be 'object', got %q",
			toolName, paramType,
		)
	}

	// Get properties if defined
	props, hasProps := params["properties"].(map[string]interface{})
	if !hasProps {
		// No properties defined - that's OK for tools with no parameters
		return nil
	}

	// If there are required fields, validate them
	required, hasRequired := params["required"].([]interface{})
	if hasRequired {
		for _, field := range required {
			fieldName, ok := field.(string)
			if !ok {
				return fmt.Errorf(
					"tool %q: 'required' field must contain strings, got %T",
					toolName, field,
				)
			}

			// Verify required field exists in properties
			if _, exists := props[fieldName]; !exists {
				return fmt.Errorf(
					"tool %q: required parameter %q not found in properties",
					toolName, fieldName,
				)
			}
		}
	}

	return nil
}

// ValidateToolCallArgs validates that provided arguments match the tool's schema.
// Checks:
// - All required parameters are provided
// - Parameter types match schema (basic type checking)
func ValidateToolCallArgs(tool *agenticcore.Tool, args map[string]interface{}) error {
	if tool == nil || tool.Parameters == nil {
		// No schema to validate against
		return nil
	}

	// Get required fields from schema
	required, hasRequired := tool.Parameters["required"].([]interface{})
	if !hasRequired {
		// No required fields specified
		return nil
	}

	// Verify all required arguments are provided
	for _, field := range required {
		fieldName, _ := field.(string)
		if _, exists := args[fieldName]; !exists {
			return fmt.Errorf(
				"tool %q: required parameter %q is missing",
				tool.Name, fieldName,
			)
		}
	}

	return nil
}

// ValidateToolMap validates all tools in a map.
// Returns first error found, or nil if all tools are valid.
func ValidateToolMap(toolsMap map[string]*agenticcore.Tool) error {
	for name, tool := range toolsMap {
		// Check key matches tool name
		if name != tool.Name {
			return fmt.Errorf(
				"tool map key %q does not match tool.Name %q",
				name, tool.Name,
			)
		}

		// Validate the tool
		if err := ValidateToolSchema(tool); err != nil {
			return err
		}
	}

	return nil
}

// ValidateToolReferences checks that all tools referenced in YAML config
// are actually registered in the tools map.
func ValidateToolReferences(toolNames []string, toolsMap map[string]*agenticcore.Tool) error {
	for _, toolName := range toolNames {
		if _, exists := toolsMap[toolName]; !exists {
			// Build list of available tools for error message
			available := make([]string, 0, len(toolsMap))
			for name := range toolsMap {
				available = append(available, name)
			}

			return fmt.Errorf(
				"tool %q is referenced but not registered. Available tools: %v",
				toolName, available,
			)
		}
	}

	return nil
}
```

### Step 2.2: Create unit tests

**File:** `/Users/taipm/GitHub/go-agentic/core/tools/validation_test.go`

```go
package tools

import (
	"context"
	"testing"

	agenticcore "github.com/taipm/go-agentic/core"
	agentictools "github.com/taipm/go-agentic/core/tools"
)

func TestValidateToolSchema(t *testing.T) {
	tests := []struct {
		name    string
		tool    *agenticcore.Tool
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid tool",
			tool: &agenticcore.Tool{
				Name:        "TestTool",
				Description: "A test tool",
				Parameters: map[string]interface{}{
					"type": "object",
					"properties": map[string]interface{}{
						"param1": map[string]interface{}{"type": "string"},
					},
					"required": []interface{}{"param1"},
				},
				Func: agentictools.ToolHandler(func(ctx context.Context, args map[string]interface{}) (string, error) {
					return "ok", nil
				}),
			},
			wantErr: false,
		},
		{
			name: "missing name",
			tool: &agenticcore.Tool{
				Name:        "",
				Description: "A test tool",
				Func:        agentictools.ToolHandler(func(ctx context.Context, args map[string]interface{}) (string, error) { return "", nil }),
			},
			wantErr: true,
			errMsg:  "name cannot be empty",
		},
		{
			name: "missing description",
			tool: &agenticcore.Tool{
				Name: "TestTool",
				Func: agentictools.ToolHandler(func(ctx context.Context, args map[string]interface{}) (string, error) { return "", nil }),
			},
			wantErr: true,
			errMsg:  "description cannot be empty",
		},
		{
			name: "missing handler",
			tool: &agenticcore.Tool{
				Name:        "TestTool",
				Description: "A test tool",
				Func:        nil,
			},
			wantErr: true,
			errMsg:  "handler function cannot be nil",
		},
		{
			name: "invalid parameter type",
			tool: &agenticcore.Tool{
				Name:        "TestTool",
				Description: "A test tool",
				Parameters: map[string]interface{}{
					"type": "string", // Should be "object"!
				},
				Func: agentictools.ToolHandler(func(ctx context.Context, args map[string]interface{}) (string, error) { return "", nil }),
			},
			wantErr: true,
			errMsg:  "Parameters.type must be 'object'",
		},
		{
			name: "required field not in properties",
			tool: &agenticcore.Tool{
				Name:        "TestTool",
				Description: "A test tool",
				Parameters: map[string]interface{}{
					"type": "object",
					"properties": map[string]interface{}{
						"param1": map[string]interface{}{"type": "string"},
					},
					"required": []interface{}{"param2"}, // param2 doesn't exist!
				},
				Func: agentictools.ToolHandler(func(ctx context.Context, args map[string]interface{}) (string, error) { return "", nil }),
			},
			wantErr: true,
			errMsg:  "required parameter",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateToolSchema(tt.tool)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateToolSchema() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr && err != nil && tt.errMsg != "" && !contains(err.Error(), tt.errMsg) {
				t.Errorf("ValidateToolSchema() error %v does not contain %q", err, tt.errMsg)
			}
		})
	}
}

func TestValidateToolCallArgs(t *testing.T) {
	tool := &agenticcore.Tool{
		Name:        "SearchDB",
		Description: "Search database",
		Parameters: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"query":  map[string]interface{}{"type": "string"},
				"limit":  map[string]interface{}{"type": "integer"},
			},
			"required": []interface{}{"query"},
		},
	}

	tests := []struct {
		name    string
		args    map[string]interface{}
		wantErr bool
	}{
		{
			name:    "all required args present",
			args:    map[string]interface{}{"query": "test", "limit": 10},
			wantErr: false,
		},
		{
			name:    "required arg missing",
			args:    map[string]interface{}{"limit": 10},
			wantErr: true,
		},
		{
			name:    "only required arg",
			args:    map[string]interface{}{"query": "test"},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateToolCallArgs(tool, tt.args)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateToolCallArgs() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func contains(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
```

### Step 2.3: Integrate validation into executor

**File:** `/Users/taipm/GitHub/go-agentic/core/executor/executor.go`

Find the `ExecuteTool` function and add validation before execution:

```go
// Add before actual execution
if err := tools.ValidateToolCallArgs(tool, call.Arguments); err != nil {
    // Return clear error to caller
    log.Printf("Tool call validation failed for %q: %v", tool.Name, err)
    return "", fmt.Errorf("tool validation error: %w", err)
}
```

Find tool loading code and add:

```go
// After loading all tools
if err := tools.ValidateToolMap(toolsMap); err != nil {
    return nil, fmt.Errorf("tool configuration error: %w", err)
}
```

### Step 2.4: Run tests

```bash
cd /Users/taipm/GitHub/go-agentic
go test ./core/tools -v -run Validate
```

**Expected output:** All validation tests pass

---

## Quick Win #3: Per-Tool Timeout Support (30 minutes)

### Goal
Allow each tool to specify its own timeout instead of using system-wide timeout.

### Why This Matters
- Currently: All tools share same timeout (not flexible)
- Better: Fast tools don't wait for slow tool defaults
- Use case: GetWeather needs 5s, DataAnalysis needs 30s

### Step 3.1: Update Tool struct

**File:** `/Users/taipm/GitHub/go-agentic/core/types.go`

Find the `Tool` struct definition and add:

```go
type Tool struct {
    Name              string                      `json:"name"`
    Description       string                      `json:"description"`
    Parameters        map[string]interface{}      `json:"parameters,omitempty"`
    Func              ToolFunc                    `json:"-"`

    // NEW: Per-tool timeout in seconds
    // If 0 or not set, uses system default from context
    TimeoutSeconds    int                         `json:"timeout_seconds,omitempty"`
}
```

### Step 3.2: Update executor to respect per-tool timeout

**File:** `/Users/taipm/GitHub/go-agentic/core/tools/executor.go`

Find the `ExecuteWithRetry` function:

```go
func ExecuteWithRetry(ctx context.Context, tool *agenticcore.Tool, args map[string]interface{}) (string, error) {
    // BEFORE:
    // result, err := executeOnce(ctx, tool, args)

    // AFTER:
    // Apply tool-specific timeout if defined
    if tool.TimeoutSeconds > 0 {
        var cancel context.CancelFunc
        newCtx, cancel := context.WithTimeout(ctx, time.Duration(tool.TimeoutSeconds)*time.Second)
        defer cancel()
        ctx = newCtx
    }

    result, err := executeOnce(ctx, tool, args)
    // ... rest of function
}
```

### Step 3.3: Create tests for timeout

**File:** `/Users/taipm/GitHub/go-agentic/core/tools/timeout_test.go`

```go
package tools

import (
	"context"
	"fmt"
	"testing"
	"time"

	agenticcore "github.com/taipm/go-agentic/core"
	agentictools "github.com/taipm/go-agentic/core/tools"
)

func TestPerToolTimeout(t *testing.T) {
	// Create a slow tool that takes 2 seconds
	slowTool := &agenticcore.Tool{
		Name:           "SlowTool",
		Description:    "A slow tool",
		TimeoutSeconds: 1, // 1 second timeout
		Func: agentictools.ToolHandler(func(ctx context.Context, args map[string]interface{}) (string, error) {
			// Simulate slow operation
			time.Sleep(2 * time.Second)
			return "done", nil
		}),
	}

	// Execute should timeout
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err := ExecuteWithRetry(ctx, slowTool, map[string]interface{}{})
	if err == nil {
		t.Errorf("Expected timeout error, but got none")
	}
}

func TestPerToolTimeoutFastTool(t *testing.T) {
	// Create a fast tool with long timeout
	fastTool := &agenticcore.Tool{
		Name:           "FastTool",
		Description:    "A fast tool",
		TimeoutSeconds: 5, // 5 seconds timeout
		Func: agentictools.ToolHandler(func(ctx context.Context, args map[string]interface{}) (string, error) {
			return "done", nil
		}),
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	result, err := ExecuteWithRetry(ctx, fastTool, map[string]interface{}{})
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	if result != "done" {
		t.Errorf("Expected 'done', got %q", result)
	}
}
```

### Step 3.4: Update tool definitions in examples

**File:** `/Users/taipm/GitHub/go-agentic/examples/00-hello-crew-tools/cmd/main.go`

Add timeouts to tool definitions:

```go
tools["GetWeather"] = &agenticcore.Tool{
    Name:           "GetWeather",
    Description:    "Get weather for a location",
    TimeoutSeconds: 5,  // 5 second timeout for API call
    // ... rest of definition
}

tools["BookHotel"] = &agenticcore.Tool{
    Name:           "BookHotel",
    Description:    "Book a hotel room",
    TimeoutSeconds: 10, // 10 second timeout for booking
    // ... rest of definition
}
```

### Step 3.5: Run tests

```bash
cd /Users/taipm/GitHub/go-agentic
go test ./core/tools -v -run Timeout
```

---

## Summary: Quick Wins Checklist

```
Quick Win #1: Type Coercion (30 min)
  ✅ Create core/tools/coercion.go
  ✅ Create core/tools/coercion_test.go
  ✅ Refactor examples/01-quiz-exam/internal/tools.go
  ✅ Run tests: go test ./core/tools -run Coerce

Quick Win #2: Schema Validation (45 min)
  ✅ Create core/tools/validation.go
  ✅ Create core/tools/validation_test.go
  ✅ Integrate into core/executor/executor.go
  ✅ Run tests: go test ./core/tools -run Validate

Quick Win #3: Per-Tool Timeout (30 min)
  ✅ Update core/types.go - add TimeoutSeconds field
  ✅ Update core/tools/executor.go - respect per-tool timeout
  ✅ Create core/tools/timeout_test.go
  ✅ Update examples with timeout values
  ✅ Run tests: go test ./core/tools -run Timeout

TOTAL PHASE 1: ~105 minutes (1.75 hours)
```

---

# PHASE 2: MEDIUM WINS (4-5 Hours)

## Opportunity #1: Tool Builder Pattern (2-3 hours)

### Goal
Provide fluent API for building tools with minimal boilerplate.

### Why This Matters
- Current: 100+ lines to define 5 tools
- Better: 20-30 lines with builder pattern
- Benefit: Easier to read, less error-prone

### Step 1.1: Create `core/tools/builder.go`

**File:** `/Users/taipm/GitHub/go-agentic/core/tools/builder.go`

```go
package tools

import (
	"fmt"

	agenticcore "github.com/taipm/go-agentic/core"
)

// ToolBuilder provides a fluent API for building Tool instances.
type ToolBuilder struct {
	name           string
	description    string
	properties     map[string]interface{}
	required       []string
	handler        agenticcore.ToolFunc
	timeoutSeconds int
}

// NewTool starts building a new tool with the given name.
func NewTool(name string) *ToolBuilder {
	return &ToolBuilder{
		name:       name,
		properties: make(map[string]interface{}),
		required:   make([]string, 0),
	}
}

// Description sets the tool's description.
func (tb *ToolBuilder) Description(desc string) *ToolBuilder {
	tb.description = desc
	return tb
}

// StringParameter adds a required string parameter.
func (tb *ToolBuilder) StringParameter(name, description string) *ToolBuilder {
	tb.properties[name] = map[string]interface{}{
		"type":        "string",
		"description": description,
	}
	tb.required = append(tb.required, name)
	return tb
}

// StringParameterOptional adds an optional string parameter with a default value.
func (tb *ToolBuilder) StringParameterOptional(name, description, defaultVal string) *ToolBuilder {
	tb.properties[name] = map[string]interface{}{
		"type":        "string",
		"description": description,
		"default":     defaultVal,
	}
	return tb
}

// IntParameter adds a required int parameter.
func (tb *ToolBuilder) IntParameter(name, description string) *ToolBuilder {
	tb.properties[name] = map[string]interface{}{
		"type":        "integer",
		"description": description,
	}
	tb.required = append(tb.required, name)
	return tb
}

// IntParameterOptional adds an optional int parameter with a default value.
func (tb *ToolBuilder) IntParameterOptional(name, description string, defaultVal int) *ToolBuilder {
	tb.properties[name] = map[string]interface{}{
		"type":        "integer",
		"description": description,
		"default":     defaultVal,
	}
	return tb
}

// BoolParameter adds a required bool parameter.
func (tb *ToolBuilder) BoolParameter(name, description string) *ToolBuilder {
	tb.properties[name] = map[string]interface{}{
		"type":        "boolean",
		"description": description,
	}
	tb.required = append(tb.required, name)
	return tb
}

// BoolParameterOptional adds an optional bool parameter with a default value.
func (tb *ToolBuilder) BoolParameterOptional(name, description string, defaultVal bool) *ToolBuilder {
	tb.properties[name] = map[string]interface{}{
		"type":        "boolean",
		"description": description,
		"default":     defaultVal,
	}
	return tb
}

// FloatParameter adds a required float parameter.
func (tb *ToolBuilder) FloatParameter(name, description string) *ToolBuilder {
	tb.properties[name] = map[string]interface{}{
		"type":        "number",
		"description": description,
	}
	tb.required = append(tb.required, name)
	return tb
}

// FloatParameterOptional adds an optional float parameter with a default value.
func (tb *ToolBuilder) FloatParameterOptional(name, description string, defaultVal float64) *ToolBuilder {
	tb.properties[name] = map[string]interface{}{
		"type":        "number",
		"description": description,
		"default":     defaultVal,
	}
	return tb
}

// Handler sets the handler function for this tool.
func (tb *ToolBuilder) Handler(handler agenticcore.ToolFunc) *ToolBuilder {
	tb.handler = handler
	return tb
}

// Timeout sets the timeout for this tool in seconds.
// If not set or 0, uses the system default timeout from context.
func (tb *ToolBuilder) Timeout(seconds int) *ToolBuilder {
	tb.timeoutSeconds = seconds
	return tb
}

// Build creates and validates the final Tool instance.
func (tb *ToolBuilder) Build() *agenticcore.Tool {
	// Validate required fields
	if tb.name == "" {
		panic("Tool name is required")
	}
	if tb.description == "" {
		panic("Tool description is required")
	}
	if tb.handler == nil {
		panic("Tool handler is required")
	}

	// Build parameters map
	parameters := map[string]interface{}{
		"type":       "object",
		"properties": tb.properties,
	}
	if len(tb.required) > 0 {
		parameters["required"] = tb.required
	}

	tool := &agenticcore.Tool{
		Name:           tb.name,
		Description:    tb.description,
		Parameters:     parameters,
		Func:           tb.handler,
		TimeoutSeconds: tb.timeoutSeconds,
	}

	// Validate the tool before returning
	if err := ValidateToolSchema(tool); err != nil {
		panic(fmt.Sprintf("Tool validation failed: %v", err))
	}

	return tool
}

// ToolSetBuilder helps build a set of tools for easy registration.
type ToolSetBuilder struct {
	tools []*agenticcore.Tool
}

// NewToolSet starts building a set of tools.
func NewToolSet() *ToolSetBuilder {
	return &ToolSetBuilder{
		tools: make([]*agenticcore.Tool, 0),
	}
}

// Add adds a tool to the set.
func (tsb *ToolSetBuilder) Add(tool *agenticcore.Tool) *ToolSetBuilder {
	tsb.tools = append(tsb.tools, tool)
	return tsb
}

// BuildMap creates a map suitable for tool registration with the executor.
func (tsb *ToolSetBuilder) BuildMap() map[string]*agenticcore.Tool {
	result := make(map[string]*agenticcore.Tool)
	for _, tool := range tsb.tools {
		result[tool.Name] = tool
	}
	return result
}

// BuildSlice returns the tools as a slice.
func (tsb *ToolSetBuilder) BuildSlice() []*agenticcore.Tool {
	return tsb.tools
}
```

### Step 1.2: Create tests for builder

**File:** `/Users/taipm/GitHub/go-agentic/core/tools/builder_test.go`

```go
package tools

import (
	"context"
	"testing"

	agenticcore "github.com/taipm/go-agentic/core"
	agentictools "github.com/taipm/go-agentic/core/tools"
)

func TestToolBuilder(t *testing.T) {
	tool := NewTool("SearchDB").
		Description("Search database").
		StringParameter("query", "The search query").
		IntParameterOptional("limit", "Max results", 10).
		Timeout(5).
		Handler(agentictools.ToolHandler(func(ctx context.Context, args map[string]interface{}) (string, error) {
			return "results", nil
		})).
		Build()

	if tool.Name != "SearchDB" {
		t.Errorf("Expected name 'SearchDB', got %q", tool.Name)
	}
	if tool.Description != "Search database" {
		t.Errorf("Expected description 'Search database', got %q", tool.Description)
	}
	if tool.TimeoutSeconds != 5 {
		t.Errorf("Expected timeout 5, got %d", tool.TimeoutSeconds)
	}

	// Check parameters structure
	if tool.Parameters == nil {
		t.Fatal("Parameters should not be nil")
	}

	// Check required fields
	required, ok := tool.Parameters["required"].([]string)
	if !ok {
		t.Fatal("Expected 'required' field to be []string")
	}
	if len(required) != 1 || required[0] != "query" {
		t.Errorf("Expected required=['query'], got %v", required)
	}

	// Check properties
	props, ok := tool.Parameters["properties"].(map[string]interface{})
	if !ok {
		t.Fatal("Expected 'properties' field to be map")
	}
	if _, ok := props["query"]; !ok {
		t.Error("Expected 'query' in properties")
	}
	if _, ok := props["limit"]; !ok {
		t.Error("Expected 'limit' in properties")
	}
}

func TestToolSetBuilder(t *testing.T) {
	tools := NewToolSet().
		Add(NewTool("Tool1").
			Description("First tool").
			Handler(agentictools.ToolHandler(func(ctx context.Context, args map[string]interface{}) (string, error) {
				return "result1", nil
			})).
			Build()).
		Add(NewTool("Tool2").
			Description("Second tool").
			Handler(agentictools.ToolHandler(func(ctx context.Context, args map[string]interface{}) (string, error) {
				return "result2", nil
			})).
			Build()).
		BuildMap()

	if len(tools) != 2 {
		t.Errorf("Expected 2 tools, got %d", len(tools))
	}

	if _, ok := tools["Tool1"]; !ok {
		t.Error("Expected 'Tool1' in map")
	}
	if _, ok := tools["Tool2"]; !ok {
		t.Error("Expected 'Tool2' in map")
	}
}

func TestToolBuilderPanicOnMissingRequired(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Error("Expected panic when description is missing")
		}
	}()

	NewTool("TestTool").
		Handler(agentictools.ToolHandler(func(ctx context.Context, args map[string]interface{}) (string, error) {
			return "", nil
		})).
		Build()
}
```

### Step 1.3: Create example showing builder usage

**File:** `/Users/taipm/GitHub/go-agentic/examples/03-tool-builder-demo/main.go`

```go
package main

import (
	"context"
	"fmt"
	"log"

	agenticcore "github.com/taipm/go-agentic/core"
	agentictools "github.com/taipm/go-agentic/core/tools"
	"github.com/taipm/go-agentic/core/tools"
)

// Demonstrate using the ToolBuilder pattern to create tools with minimal boilerplate.
//
// This example shows:
// - Building tools with fluent API
// - Parameter definitions as method calls
// - Automatic schema generation
// - Clear, readable tool definitions

func main() {
	// Create a set of tools using the builder pattern
	toolsMap := tools.NewToolSet().
		Add(createSearchTool()).
		Add(createCalculatorTool()).
		Add(createGreeterTool()).
		BuildMap()

	// Validate all tools
	if err := tools.ValidateToolMap(toolsMap); err != nil {
		log.Fatalf("Tool validation failed: %v", err)
	}

	// Display tools
	fmt.Println("Available Tools:")
	for name, tool := range toolsMap {
		fmt.Printf("- %s: %s\n", name, tool.Description)
		if tool.TimeoutSeconds > 0 {
			fmt.Printf("  Timeout: %ds\n", tool.TimeoutSeconds)
		}
	}
}

func createSearchTool() *agenticcore.Tool {
	return tools.NewTool("SearchDatabase").
		Description("Search database for records matching criteria").
		StringParameter("query", "The search query string").
		IntParameterOptional("limit", "Maximum number of results to return", 10).
		Timeout(5).
		Handler(agentictools.ToolHandler(func(ctx context.Context, args map[string]interface{}) (string, error) {
			query, err := tools.MustGetString(args, "query")
			if err != nil {
				return "", err
			}

			limit := tools.OptionalGetInt(args, "limit", 10)

			// Simulate search
			return fmt.Sprintf("Found 5 results for '%s' (showing top %d)", query, limit), nil
		})).
		Build()
}

func createCalculatorTool() *agenticcore.Tool {
	return tools.NewTool("Calculator").
		Description("Perform mathematical calculations").
		FloatParameter("a", "First number").
		FloatParameter("b", "Second number").
		StringParameter("operation", "Operation to perform: add, subtract, multiply, divide").
		Timeout(2).
		Handler(agentictools.ToolHandler(func(ctx context.Context, args map[string]interface{}) (string, error) {
			a, err := tools.MustGetString(args, "a")
			if err != nil {
				return "", err
			}

			b, err := tools.MustGetString(args, "b")
			if err != nil {
				return "", err
			}

			op, err := tools.MustGetString(args, "operation")
			if err != nil {
				return "", err
			}

			return fmt.Sprintf("Calculation: %s %s %s", a, op, b), nil
		})).
		Build()
}

func createGreeterTool() *agenticcore.Tool {
	return tools.NewTool("Greet").
		Description("Greet someone by name").
		StringParameter("name", "Name of person to greet").
		StringParameterOptional("greeting", "Type of greeting", "Hello").
		Timeout(1).
		Handler(agentictools.ToolHandler(func(ctx context.Context, args map[string]interface{}) (string, error) {
			name := tools.OptionalGetString(args, "name", "World")
			greeting := tools.OptionalGetString(args, "greeting", "Hello")

			return fmt.Sprintf("%s, %s!", greeting, name), nil
		})).
		Build()
}
```

### Step 1.4: Run tests and example

```bash
cd /Users/taipm/GitHub/go-agentic
go test ./core/tools -v -run Builder
go run ./examples/03-tool-builder-demo/main.go
```

---

## Opportunity #2: Schema Auto-Generation (2-3 hours)

### Goal
Auto-generate tool parameters from Go struct definitions.

### Why This Matters
- Current: Schema defined manually in map (boilerplate)
- Better: Define struct once, schema generated automatically
- Benefit: Single source of truth, less code

### Step 2.1: Create `core/tools/struct_schema.go`

**File:** `/Users/taipm/GitHub/go-agentic/core/tools/struct_schema.go`

```go
package tools

import (
	"fmt"
	"reflect"
	"strconv"
	"strings"
)

// StructSchemaGenerator generates JSON schema from Go struct definitions.
type StructSchemaGenerator struct {
	model interface{}
}

// NewStructSchema creates a generator for the given struct.
func NewStructSchema(model interface{}) *StructSchemaGenerator {
	return &StructSchemaGenerator{model: model}
}

// Generate creates a JSON schema map from the struct.
// Supports struct tags:
//   - json:"name" - Field name in JSON
//   - tool:"description:...; required; default:..."
func (g *StructSchemaGenerator) Generate() (map[string]interface{}, error) {
	t := reflect.TypeOf(g.model)

	// Handle pointers
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}

	// Must be a struct
	if t.Kind() != reflect.Struct {
		return nil, fmt.Errorf("expected struct or *struct, got %s", t.Kind())
	}

	properties := make(map[string]interface{})
	required := make([]string, 0)

	// Iterate over fields
	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)

		// Skip unexported fields
		if !field.IsExported() {
			continue
		}

		// Get JSON field name
		jsonTag := field.Tag.Get("json")
		if jsonTag == "" {
			jsonTag = field.Name
		}

		// Skip if explicitly ignored
		if jsonTag == "-" {
			continue
		}

		// Parse tool metadata tags
		toolTag := field.Tag.Get("tool")
		metadata := g.parseToolTag(toolTag)

		// Determine JSON type
		jsonType := g.getJSONType(field.Type)

		// Build property schema
		property := map[string]interface{}{
			"type": jsonType,
		}

		// Add description
		if desc, ok := metadata["description"]; ok {
			property["description"] = desc
		}

		// Add optional constraints
		if def, ok := metadata["default"]; ok {
			property["default"] = def
		}
		if min, ok := metadata["minimum"]; ok {
			property["minimum"] = min
		}
		if max, ok := metadata["maximum"]; ok {
			property["maximum"] = max
		}
		if enum, ok := metadata["enum"]; ok {
			property["enum"] = strings.Split(enum, ",")
		}

		properties[jsonTag] = property

		// Mark as required if not optional
		if _, isOptional := metadata["optional"]; !isOptional && metadata["description"] != "" {
			required = append(required, jsonTag)
		}
	}

	// Build final schema
	result := map[string]interface{}{
		"type":       "object",
		"properties": properties,
	}

	if len(required) > 0 {
		result["required"] = required
	}

	return result, nil
}

// parseToolTag parses the tool metadata tag.
// Format: "description:...; required; optional; default:...; minimum:...; maximum:..."
func (g *StructSchemaGenerator) parseToolTag(tag string) map[string]string {
	result := make(map[string]string)

	if tag == "" {
		return result
	}

	// Split by semicolon
	parts := strings.Split(tag, ";")
	for _, part := range parts {
		part = strings.TrimSpace(part)

		if part == "required" {
			result["required"] = "true"
		} else if part == "optional" {
			result["optional"] = "true"
		} else if strings.Contains(part, ":") {
			// Key:value format
			kv := strings.SplitN(part, ":", 2)
			if len(kv) == 2 {
				result[strings.TrimSpace(kv[0])] = strings.TrimSpace(kv[1])
			}
		}
	}

	return result
}

// getJSONType maps Go types to JSON schema types.
func (g *StructSchemaGenerator) getJSONType(t reflect.Type) string {
	switch t.Kind() {
	case reflect.String:
		return "string"
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return "integer"
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return "integer"
	case reflect.Float32, reflect.Float64:
		return "number"
	case reflect.Bool:
		return "boolean"
	case reflect.Slice:
		return "array"
	case reflect.Map:
		return "object"
	default:
		// Default to string for unknown types
		return "string"
	}
}

// ToolBuilder extension method to use struct schema
// This allows: NewTool("...").SchemaFromStruct(MyParams{})
func (tb *ToolBuilder) SchemaFromStruct(model interface{}) *ToolBuilder {
	generator := NewStructSchema(model)
	schema, err := generator.Generate()
	if err != nil {
		panic(fmt.Sprintf("failed to generate schema from struct: %v", err))
	}

	tb.properties = schema["properties"].(map[string]interface{})

	if required, ok := schema["required"].([]string); ok {
		tb.required = required
	}

	return tb
}
```

### Step 2.2: Create tests for schema generation

**File:** `/Users/taipm/GitHub/go-agentic/core/tools/struct_schema_test.go`

```go
package tools

import (
	"encoding/json"
	"testing"
)

type SearchParams struct {
	Query    string `json:"query" tool:"description:Search query;required"`
	Limit    int    `json:"limit" tool:"description:Max results;default:10;optional"`
	Filters  string `json:"filters" tool:"description:Filters;optional"`
	MinScore float64 `json:"min_score" tool:"description:Min score;default:0.5;optional"`
}

type UserParams struct {
	Name  string `json:"name" tool:"description:User name"`
	Email string `json:"email" tool:"description:Email address"`
	Admin bool   `json:"admin" tool:"description:Is admin;optional;default:false"`
}

func TestStructSchemaGeneration(t *testing.T) {
	gen := NewStructSchema(&SearchParams{})
	schema, err := gen.Generate()
	if err != nil {
		t.Fatalf("Generate() error: %v", err)
	}

	// Check type
	if typ, ok := schema["type"].(string); !ok || typ != "object" {
		t.Errorf("Expected type 'object', got %v", schema["type"])
	}

	// Check properties exist
	props, ok := schema["properties"].(map[string]interface{})
	if !ok {
		t.Fatal("Expected properties to be map")
	}

	// Check query field
	if queryProp, ok := props["query"]; ok {
		queryMap := queryProp.(map[string]interface{})
		if queryMap["type"] != "string" {
			t.Errorf("Expected query type string, got %v", queryMap["type"])
		}
	} else {
		t.Error("Expected query in properties")
	}

	// Check limit field with default
	if limitProp, ok := props["limit"]; ok {
		limitMap := limitProp.(map[string]interface{})
		if limitMap["type"] != "integer" {
			t.Errorf("Expected limit type integer, got %v", limitMap["type"])
		}
		if limitMap["default"] != "10" {
			t.Errorf("Expected limit default 10, got %v", limitMap["default"])
		}
	} else {
		t.Error("Expected limit in properties")
	}

	// Check required fields
	required, ok := schema["required"].([]string)
	if !ok {
		t.Fatal("Expected required to be []string")
	}
	if len(required) != 1 || required[0] != "query" {
		t.Errorf("Expected required=['query'], got %v", required)
	}
}

func TestStructSchemaJSON(t *testing.T) {
	gen := NewStructSchema(&UserParams{})
	schema, err := gen.Generate()
	if err != nil {
		t.Fatalf("Generate() error: %v", err)
	}

	// Should be valid JSON
	b, err := json.MarshalIndent(schema, "", "  ")
	if err != nil {
		t.Fatalf("JSON marshal error: %v", err)
	}

	t.Logf("Generated schema:\n%s", string(b))
}

func TestStructSchemaWithPointer(t *testing.T) {
	// Should work with pointer to struct
	gen := NewStructSchema(&SearchParams{})
	schema, err := gen.Generate()
	if err != nil {
		t.Fatalf("Generate() with pointer error: %v", err)
	}

	if schema == nil {
		t.Error("Expected schema, got nil")
	}
}

func TestStructSchemaIgnoresUnexportedFields(t *testing.T) {
	type Params struct {
		Public  string `json:"public" tool:"description:Public field"`
		private string `json:"private" tool:"description:Private field"` // Should be ignored
	}

	gen := NewStructSchema(&Params{})
	schema, err := gen.Generate()
	if err != nil {
		t.Fatalf("Generate() error: %v", err)
	}

	props := schema["properties"].(map[string]interface{})
	if _, ok := props["private"]; ok {
		t.Error("Expected private field to be ignored")
	}
}
```

### Step 2.3: Update ToolBuilder to support SchemaFromStruct

The `SchemaFromStruct` method was added to `ToolBuilder` in the code above.

### Step 2.4: Create example with struct-based tools

**File:** `/Users/taipm/GitHub/go-agentic/examples/04-struct-schema-demo/main.go`

```go
package main

import (
	"context"
	"fmt"
	"log"

	agenticcore "github.com/taipm/go-agentic/core"
	agentictools "github.com/taipm/go-agentic/core/tools"
	"github.com/taipm/go-agentic/core/tools"
)

// Define tool parameters as structs - schema is auto-generated!
type SearchParams struct {
	Query   string  `json:"query" tool:"description:Search query;required"`
	Limit   int     `json:"limit" tool:"description:Max results;default:10;optional"`
	MinScore float64 `json:"min_score" tool:"description:Min score;default:0.5;optional"`
}

type CalculateParams struct {
	Operation string  `json:"operation" tool:"description:add|subtract|multiply|divide;required"`
	A         float64 `json:"a" tool:"description:First number;required"`
	B         float64 `json:"b" tool:"description:Second number;required"`
}

func main() {
	// Create tools using struct-based schema generation
	toolsMap := tools.NewToolSet().
		Add(createSearchToolWithStruct()).
		Add(createCalculatorToolWithStruct()).
		BuildMap()

	// Validate all tools
	if err := tools.ValidateToolMap(toolsMap); err != nil {
		log.Fatalf("Tool validation failed: %v", err)
	}

	// Display tools
	fmt.Println("Available Tools with Auto-Generated Schema:")
	for name, tool := range toolsMap {
		fmt.Printf("\n%s: %s\n", name, tool.Description)
		if props, ok := tool.Parameters["properties"].(map[string]interface{}); ok {
			for paramName, param := range props {
				paramMap := param.(map[string]interface{})
				fmt.Printf("  - %s (%s): %v\n",
					paramName, paramMap["type"], paramMap["description"])
			}
		}
	}
}

func createSearchToolWithStruct() *agenticcore.Tool {
	return tools.NewTool("Search").
		Description("Search database using auto-generated schema from struct").
		SchemaFromStruct(&SearchParams{}).
		Timeout(5).
		Handler(agentictools.ToolHandler(func(ctx context.Context, args map[string]interface{}) (string, error) {
			query, err := tools.MustGetString(args, "query")
			if err != nil {
				return "", err
			}

			limit := tools.OptionalGetInt(args, "limit", 10)
			minScore := tools.OptionalGetFloat(args, "min_score", 0.5)

			return fmt.Sprintf(
				"Searching for '%s' (limit: %d, min_score: %.1f)",
				query, limit, minScore,
			), nil
		})).
		Build()
}

func createCalculatorToolWithStruct() *agenticcore.Tool {
	return tools.NewTool("Calculate").
		Description("Perform calculations using auto-generated schema from struct").
		SchemaFromStruct(&CalculateParams{}).
		Timeout(2).
		Handler(agentictools.ToolHandler(func(ctx context.Context, args map[string]interface{}) (string, error) {
			op, err := tools.MustGetString(args, "operation")
			if err != nil {
				return "", err
			}

			a, err := tools.CoerceToFloat(args["a"])
			if err != nil {
				return "", err
			}

			b, err := tools.CoerceToFloat(args["b"])
			if err != nil {
				return "", err
			}

			var result float64
			switch op {
			case "add":
				result = a + b
			case "subtract":
				result = a - b
			case "multiply":
				result = a * b
			case "divide":
				if b == 0 {
					return "", fmt.Errorf("division by zero")
				}
				result = a / b
			default:
				return "", fmt.Errorf("unknown operation: %s", op)
			}

			return fmt.Sprintf("%s %.2f %s %.2f = %.2f", op, a, op, b, result), nil
		})).
		Build()
}
```

### Step 2.5: Add helper to coercion.go

Add this helper to `core/tools/coercion.go`:

```go
// OptionalGetFloat extracts an optional float parameter with default value.
func OptionalGetFloat(args map[string]interface{}, key string, defaultVal float64) float64 {
	val, exists := args[key]
	if !exists {
		return defaultVal
	}

	result, err := CoerceToFloat(val)
	if err != nil {
		return defaultVal
	}
	return result
}
```

### Step 2.6: Run tests

```bash
cd /Users/taipm/GitHub/go-agentic
go test ./core/tools -v -run StructSchema
go run ./examples/04-struct-schema-demo/main.go
```

---

## Summary: Phase 2 Checklist

```
Opportunity #1: Builder Pattern (2-3 hours)
  ✅ Create core/tools/builder.go with ToolBuilder and ToolSetBuilder
  ✅ Create core/tools/builder_test.go
  ✅ Create examples/03-tool-builder-demo/main.go
  ✅ Run tests: go test ./core/tools -run Builder
  ✅ Run example: go run ./examples/03-tool-builder-demo/main.go

Opportunity #2: Schema Auto-Gen (2-3 hours)
  ✅ Create core/tools/struct_schema.go
  ✅ Create core/tools/struct_schema_test.go
  ✅ Update ToolBuilder with SchemaFromStruct() method
  ✅ Create examples/04-struct-schema-demo/main.go
  ✅ Add OptionalGetFloat() to coercion.go
  ✅ Run tests: go test ./core/tools -run StructSchema
  ✅ Run example: go run ./examples/04-struct-schema-demo/main.go

TOTAL PHASE 2: ~4-5 hours
```

---

# PHASE 3: REFACTORING EXISTING EXAMPLES (2-3 hours)

## Goal
Show how to use new patterns in existing projects.

### Step 1: Refactor `examples/01-quiz-exam`

Replace old pattern with new utilities:

```go
// OLD
var studentAnswer string
switch v := args["student_answer"].(type) {
case string:
    studentAnswer = v
case float64:
    studentAnswer = fmt.Sprintf("%v", v)
// ... more cases ...
}

// NEW
studentAnswer, err := tools.MustGetString(args, "student_answer")
if err != nil {
    return ..., fmt.Errorf("invalid answer: %w", err)
}
```

### Step 2: Show before/after comparison

Create a document: `IMPROVEMENTS_SHOWCASE.md`

```markdown
# Tool Declaration Improvements Showcase

## Example 1: Type Coercion

### Before (10 lines of boilerplate)
```go
var studentAnswer string
switch v := args["student_answer"].(type) {
case string:
    studentAnswer = v
case float64:
    studentAnswer = fmt.Sprintf("%v", v)
case int64:
    studentAnswer = fmt.Sprintf("%d", v)
case int:
    studentAnswer = fmt.Sprintf("%d", v)
default:
    studentAnswer = fmt.Sprintf("%v", v)
}
```

### After (2 lines)
```go
studentAnswer, err := tools.MustGetString(args, "student_answer")
if err != nil {
    return "", fmt.Errorf("invalid answer: %w", err)
}
```

... more examples ...
```

---

# FINAL CHECKLIST: All 5 Improvements

```
✅ QUICK WIN #1: Type Coercion Utility (30 min)
   Files: core/tools/coercion.go, coercion_test.go
   Status: Ready to use

✅ QUICK WIN #2: Schema Validation (45 min)
   Files: core/tools/validation.go, validation_test.go
   Integrated: executor.go
   Status: Ready to use

✅ QUICK WIN #3: Per-Tool Timeout (30 min)
   Files: core/types.go, executor.go, timeout_test.go
   Status: Ready to use

✅ OPPORTUNITY #1: Tool Builder Pattern (2-3 hours)
   Files: core/tools/builder.go, builder_test.go
   Example: examples/03-tool-builder-demo/main.go
   Status: Ready to use

✅ OPPORTUNITY #2: Schema Auto-Generation (2-3 hours)
   Files: core/tools/struct_schema.go, struct_schema_test.go
   Example: examples/04-struct-schema-demo/main.go
   Status: Ready to use

✅ PHASE 3: Refactor Examples (2-3 hours)
   Update: examples/01-quiz-exam, examples/00-hello-crew-tools
   Document: IMPROVEMENTS_SHOWCASE.md
   Status: Show improvements in action

TOTAL IMPLEMENTATION: 12-18 hours
```

---

# NEXT STEPS FOR TEAM

1. **Today/Tomorrow:** Implement Quick Wins #1-3 (2-3 hours)
   - These have immediate ROI
   - Zero breaking changes
   - Can be deployed independently

2. **This Week:** Implement Opportunities #1-2 (4-5 hours)
   - Build on Quick Wins
   - Show in new examples
   - Document usage

3. **Next Week:** Refactor existing examples (2-3 hours)
   - Update all examples to use new patterns
   - Create migration guide for users
   - Close old patterns

---

**🎯 Success Criteria:**

- ✅ All tests pass
- ✅ No breaking changes to existing API
- ✅ Boilerplate reduced by 60-70%
- ✅ Type coercion bugs eliminated
- ✅ Schema divergence impossible
- ✅ Developer experience significantly improved
- ✅ Examples show new patterns
- ✅ Clear upgrade path for existing projects


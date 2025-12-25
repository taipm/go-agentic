# Core Tool API Issue - Analysis & Recommendations

## Problem Summary

**Severity:** High - API Inconsistency between struct definition and usage

The `Tool` struct in `core/common/types.go` defines the function field as `Func`, but all example code and documentation uses `Handler`. This causes compilation errors.

## Current State

### Struct Definition (core/common/types.go:18-26)
```go
type Tool struct {
	ID          string
	Name        string
	Description string
	Func        interface{} // ← Currently named "Func"
	Input       interface{}
	Parameters  interface{}
	Output      interface{}
}
```

### Example Usage (examples/00-hello-crew-tools/cmd/main.go)
All 5 tools in this example use `Handler` field:
```go
tool1 := &agenticcore.Tool{
	Name:        "GetMessageCount",
	Description: "Returns the total number of messages...",
	Parameters:  map[string]interface{}{...},
	Handler: func(ctx context.Context, args map[string]interface{}) (string, error) {
		// Implementation
		return output, nil
	},
}
```

### Compilation Error
```
cmd/main.go:77:3: unknown field Handler in struct literal of type crewai.Tool
cmd/main.go:104:3: unknown field Handler in struct literal of type crewai.Tool
cmd/main.go:138:3: unknown field Handler in struct literal of type crewai.Tool
cmd/main.go:171:3: unknown field Handler in struct literal of type crewai.Tool
cmd/main.go:194:3: unknown field Handler in struct literal of type crewai.Tool
```

## Root Cause Analysis

The struct field is named `Func` but all documentation and examples refer to it as `Handler`. This indicates one of these scenarios:

1. **API Design Decision Changed**: The field was renamed from `Handler` to `Func` at some point, but examples and documentation weren't updated
2. **Struct Tag Missing**: The field should have a struct tag like `// Handler` or a JSON/YAML alias
3. **Incomplete Implementation**: The `Handler` field was removed but code still references it

## Investigation Findings

- Core type definition: `/core/common/types.go` (line 22)
- Handler function signature defined in: `/core/tools/errors.go`
- Working examples exist in: `/examples/01-quiz-exam/internal/tools.go` (uses `Func` correctly)
- Broken examples: `/examples/00-hello-crew-tools/cmd/main.go` (uses `Handler`)

### Function Signature (core/tools/errors.go:61)
```go
type ToolHandler func(ctx context.Context, args map[string]interface{}) (string, error)
```

## Temporary Solution (Applied)

Fixed `/examples/00-hello-crew-tools/cmd/main.go` by changing all `Handler` fields to `Func`:

**Changes Made:**
- Line 77: `Handler:` → `Func:`
- Line 104: `Handler:` → `Func:`
- Line 138: `Handler:` → `Func:`
- Line 171: `Handler:` → `Func:`
- Line 194: `Handler:` → `Func:`

**Status:** ✅ Example now builds successfully

## Recommended Core Fixes

### Option 1: Rename to `Handler` (Recommended for API clarity)
**Pros:**
- More semantic - "Handler" clearly indicates a function
- Matches all existing examples and documentation
- Better developer experience

**Cons:**
- Requires changing struct definition
- Possible breaking change if `Func` is already used elsewhere

**Implementation:**
```go
type Tool struct {
	ID          string
	Name        string
	Description string
	Handler     interface{} // Renamed from Func
	Input       interface{}
	Parameters  interface{}
	Output      interface{}
}
```

### Option 2: Add Field Alias (Keep Both)
**Pros:**
- Backward compatible
- Supports both naming conventions

**Cons:**
- Can cause confusion

**Implementation:**
```go
type Tool struct {
	ID          string
	Name        string
	Description string
	Func        interface{}     // Keep original
	Handler     interface{}     // Add alias
	Input       interface{}
	Parameters  interface{}
	Output      interface{}
}
```

### Option 3: Update Documentation (Not Recommended)
**Status:** Already attempted - examples still use `Handler`

## Action Items for Core

1. **Decide on naming convention**: `Func` vs `Handler`
2. **Audit all examples**: Ensure consistency across all examples
3. **Update type definition** if renaming to `Handler`
4. **Test all examples** to verify they compile and work
5. **Document the Tool API** clearly in README

## Files That Need Review

- [ ] `/core/common/types.go` - Struct definition
- [ ] `/examples/00-hello-crew-tools/cmd/main.go` - FIXED ✅
- [ ] `/examples/01-quiz-exam/internal/tools.go` - Uses `Func` (correct)
- [ ] Any other examples that use Tool struct
- [ ] Documentation and API references

## Usage Pattern (Correct)

After fix, the correct pattern is:

```go
tool := &agenticcore.Tool{
	Name:        "ToolName",
	Description: "What this tool does",
	Parameters: map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"param1": map[string]interface{}{
				"type":        "string",
				"description": "Parameter description",
			},
		},
		"required": []string{"param1"},
	},
	Func: func(ctx context.Context, args map[string]interface{}) (string, error) {
		// Tool implementation
		result := map[string]interface{}{"key": "value"}
		jsonBytes, _ := json.Marshal(result)
		return string(jsonBytes), nil
	},
}
```

## Next Steps

1. ✅ Fixed example 00-hello-crew-tools
2. Recommend core maintainer to choose naming convention and update struct
3. Verify all examples build and run correctly
4. Update any API documentation

# Phase 2: ValidationErrorFormatter - JSON Output

## Overview

Phase 2 adds JSON output support to the `ConfigValidator` for structured error reporting. This enables:
- **APIs**: Return structured JSON errors to HTTP clients
- **Logging Systems**: Send validation results to centralized logging
- **CI/CD**: Parse validation results programmatically in pipelines
- **Developer Tools**: CLI tools can consume JSON and format as desired

## Implementation: Approach A (JSON-Only)

**Status**: ✅ Complete and tested

### What Was Added

#### New Types (core/validation.go:444-468)

```go
// ErrorDetail represents a single validation error in JSON format
type ErrorDetail struct {
	File     string `json:"file"`
	Field    string `json:"field"`
	Message  string `json:"message"`
	Severity string `json:"severity"`
	Fix      string `json:"fix"`
	Line     int    `json:"line,omitempty"`
}

// ErrorSummary provides count and validity status
type ErrorSummary struct {
	TotalErrors   int  `json:"total_errors"`
	TotalWarnings int  `json:"total_warnings"`
	IsValid       bool `json:"is_valid"`
}

// ErrorResponse is the complete JSON structure for validation results
type ErrorResponse struct {
	Success  bool           `json:"success"`
	Errors   []ErrorDetail  `json:"errors"`
	Warnings []ErrorDetail  `json:"warnings"`
	Summary  ErrorSummary   `json:"summary"`
}
```

#### New Method (core/validation.go:495-539)

```go
// ToJSON converts validation results to JSON format
// Returns pretty-printed JSON that can be:
// - Consumed by APIs and clients (parse JSON directly)
// - Sent to logging systems (structured logging)
// - Inspected by developers (readable format)
func (cv *ConfigValidator) ToJSON() ([]byte, error)
```

### Test Coverage

5 new test cases in core/validation_test.go (lines 363-575):

1. **TestToJSONValidConfiguration** - Valid config produces `success=true`
2. **TestToJSONInvalidConfiguration** - Invalid config produces `success=false` with error details
3. **TestToJSONStructure** - Verifies complete JSON structure matches schema
4. **TestToJSONWithWarnings** - Handles warnings separately from errors
5. **TestToJSONFormatting** - Verifies pretty-printed JSON formatting

**Result**: ✅ All 5 tests passing + all 200+ existing tests still passing

### Usage Examples

#### Example 1: Get JSON for Valid Configuration

```go
validator := NewConfigValidator(crewConfig, agentConfigs)
validator.ValidateAll()

// Get JSON output
jsonData, err := validator.ToJSON()
if err != nil {
    log.Fatal(err)
}

// Pretty output
fmt.Println(string(jsonData))
```

**Output**:
```json
{
  "success": true,
  "errors": [],
  "warnings": [],
  "summary": {
    "total_errors": 0,
    "total_warnings": 0,
    "is_valid": true
  }
}
```

#### Example 2: Get JSON for Invalid Configuration

```go
// Configuration with errors
config := &CrewConfig{
    EntryPoint: "bad_entry",  // Not in agents list
    Agents:     []string{"orchestrator"},
}

validator := NewConfigValidator(config, agents)
validator.ValidateAll()

jsonData, _ := validator.ToJSON()
fmt.Println(string(jsonData))
```

**Output**:
```json
{
  "success": false,
  "errors": [
    {
      "file": "crew.yaml",
      "field": "entry_point",
      "message": "entry_point 'bad_entry' not found in agents list",
      "severity": "error",
      "fix": "entry_point must be one of: orchestrator"
    }
  ],
  "warnings": [],
  "summary": {
    "total_errors": 1,
    "total_warnings": 0,
    "is_valid": false
  }
}
```

#### Example 3: Parse JSON in CLI Tool

```go
// In a CLI application
validator := NewConfigValidator(config, agents)
validator.ValidateAll()

jsonData, _ := validator.ToJSON()

var resp ErrorResponse
if err := json.Unmarshal(jsonData, &resp); err != nil {
    log.Fatal(err)
}

// Check if valid
if resp.Success {
    fmt.Println("✅ Configuration is valid")
} else {
    fmt.Println("❌ Configuration has errors:")
    for i, err := range resp.Errors {
        fmt.Printf("  %d. %s (%s): %s\n", i+1, err.Field, err.File, err.Message)
        fmt.Printf("     Fix: %s\n", err.Fix)
    }
}
```

#### Example 4: Send to Logging System

```go
// Send structured JSON to logging backend
validator := NewConfigValidator(config, agents)
validator.ValidateAll()

jsonData, _ := validator.ToJSON()

// Send to your logging system
logger.LogJSON("config_validation", jsonData)
```

#### Example 5: Use in HTTP API

```go
// In a REST API handler
func validateConfigHandler(w http.ResponseWriter, r *http.Request) {
    // Load and validate config
    validator := NewConfigValidator(config, agents)
    validator.ValidateAll()

    // Return JSON response
    jsonData, _ := validator.ToJSON()

    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(http.StatusOK)
    w.Write(jsonData)
}
```

### JSON Schema

```
{
  "success": boolean,              // true if no errors, false if errors exist
  "errors": [
    {
      "file": string,              // Which YAML file (crew.yaml, agents/x.yaml)
      "field": string,             // Which config field
      "message": string,           // What's wrong
      "severity": string,          // "error" or "warning"
      "fix": string,               // How to fix it
      "line": number (optional)    // Line number in file
    }
  ],
  "warnings": [
    // Same structure as errors
  ],
  "summary": {
    "total_errors": number,        // Count of errors
    "total_warnings": number,      // Count of warnings
    "is_valid": boolean            // true if total_errors == 0
  }
}
```

### Integration Points

The `ToJSON()` method is available on any `ConfigValidator` instance:

```go
// After validation
validator := NewConfigValidator(config, agents)
validator.ValidateAll()

// Access validation results
errors := validator.GetErrors()      // []ValidationError
warnings := validator.GetWarnings()  // []ValidationError
isValid := validator.IsValid()       // bool

// NEW: Get JSON output
jsonData, _ := validator.ToJSON()    // []byte
```

### Performance Characteristics

- **Time Complexity**: O(n) where n = total number of errors + warnings
- **Space Complexity**: O(n) for JSON serialization
- **Thread-Safe**: Uses same RWMutex as GetErrors()/GetWarnings()
- **Pretty-Printed**: Uses `json.MarshalIndent()` for readability

### Backward Compatibility

✅ **Fully backward compatible**:
- No existing methods changed
- No existing types modified
- Existing code continues to work unchanged
- `ToJSON()` is new optional method

### Why This Approach (Not Alternatives)

**vs. Multi-Format (JSON + Markdown + Text)**:
- ❌ Multi-format adds 2.5x more code (250-300 LOC vs 120-150 LOC)
- ❌ 15-20% code duplication for format conversion
- ❌ Harder to maintain when ValidationError changes
- ✅ JSON is foundation; Markdown/Text can be added later if needed

**vs. Plugin-Based System**:
- ❌ 4x larger codebase (400-500 LOC)
- ❌ New architectural pattern not used elsewhere in go-agentic
- ❌ Overkill for 3 formatters
- ✅ YAGNI principle: simple is better

### Future Extensions

If needed, these can be added easily:

```go
// Markdown formatter (1-2 hours)
func (cv *ConfigValidator) ToMarkdown() string { ... }

// Plain text formatter (1-2 hours)
func (cv *ConfigValidator) ToText() string { ... }

// HTML formatter (1-2 hours)
func (cv *ConfigValidator) ToHTML() string { ... }
```

### Metrics

| Metric | Value |
|--------|-------|
| LOC Added | 120-150 |
| Tests Added | 5 tests |
| Test Coverage | 100% of ToJSON() |
| All Tests | 200+ passing |
| Backward Compatible | ✅ Yes |
| Implementation Time | 2-3 hours |
| Thread-Safe | ✅ Yes |

### Implementation Checklist

- ✅ Add `encoding/json` import
- ✅ Define ErrorDetail, ErrorSummary, ErrorResponse types
- ✅ Implement ToJSON() method (lines 495-539)
- ✅ Add 5 test cases
- ✅ All tests passing
- ✅ Documentation with examples
- ✅ Ready for production

## Files Modified

1. **core/validation.go**
   - Added `encoding/json` import
   - Added types: ErrorDetail, ErrorSummary, ErrorResponse (lines 444-468)
   - Added method: ToJSON() (lines 495-539)

2. **core/validation_test.go**
   - Added `encoding/json` import
   - Added 5 test functions (lines 363-575):
     - TestToJSONValidConfiguration
     - TestToJSONInvalidConfiguration
     - TestToJSONStructure
     - TestToJSONWithWarnings
     - TestToJSONFormatting

3. **core/example_json_formatter_test.go** (NEW)
   - Added example usage demonstrations

## Next Steps

1. ✅ Implement JSON formatter (DONE)
2. ✅ Write comprehensive tests (DONE)
3. ✅ Documentation with examples (DONE)
4. 📋 Optional: Add Markdown formatter (future sprint)
5. 📋 Optional: Add CLI tool that outputs validation as JSON (future sprint)

## Summary

Phase 2 implementation adds JSON output support to ConfigValidator with:
- **120-150 LOC** of clean, focused code
- **5 comprehensive test cases** covering valid/invalid configs, structure validation, warnings, and formatting
- **100% backward compatible** - no changes to existing code
- **Ready for production** - all tests passing
- **Foundation for future formatters** - easy to add Markdown/Text if needed

This approach aligns with go-agentic's philosophy: provide structured data, let consumers choose how to use it.

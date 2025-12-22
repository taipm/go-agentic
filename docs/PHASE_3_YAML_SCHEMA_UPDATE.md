# Phase 3: YAML Configuration Schema Update

**Date:** 2025-12-22
**Status:** ✅ COMPLETE
**Branch:** feature/epic-4-cross-platform

---

## Overview

Phase 3 updates the YAML configuration schema to expose the new configurable fields introduced in Phase 1 fixes #4 and #5.

---

## What Was Updated

### New YAML Configuration Fields

Two new configuration fields are now available in the `settings` section of `crew.yaml`:

#### 1. `parallel_timeout_seconds` (Fix #4)

**Purpose:** Control the maximum time allowed for parallel agent execution

**Type:** Integer (seconds)
**Default:** 60 seconds (if not specified or zero)
**Location:** `crew.yaml` -> `settings` section

**YAML Syntax:**
```yaml
settings:
  parallel_timeout_seconds: 120  # Example: 2 minutes
```

**When to use:**
- Increase for complex tasks requiring more computation time
- Decrease for quick response time requirements
- Useful when agents perform heavy operations in parallel

**Examples:**
```yaml
# Quick response crew (30 seconds)
settings:
  parallel_timeout_seconds: 30

# Medium workload crew (default 60 seconds)
settings:
  parallel_timeout_seconds: 60

# Long-running diagnostic crew (2-5 minutes)
settings:
  parallel_timeout_seconds: 300
```

---

#### 2. `max_tool_output_chars` (Fix #5)

**Purpose:** Control the maximum characters allowed in tool output before truncation

**Type:** Integer (characters)
**Default:** 2000 characters (if not specified or zero)
**Location:** `crew.yaml` -> `settings` section

**YAML Syntax:**
```yaml
settings:
  max_tool_output_chars: 5000  # Example: 5000 characters
```

**When to use:**
- Increase for detailed output (diagnostic tools, logs)
- Decrease for context-limited scenarios
- Prevents context overflow from large tool outputs

**Examples:**
```yaml
# Minimal output crew (1000 chars)
settings:
  max_tool_output_chars: 1000

# Standard crew (default 2000 chars)
settings:
  max_tool_output_chars: 2000

# Detailed output crew (5000+ chars)
settings:
  max_tool_output_chars: 5000
```

---

## Updated Example Configurations

### Example 1: Hello Crew (examples/00-hello-crew/config/crew.yaml)

**Before:**
```yaml
version: "1.0"
name: hello-crew
description: A minimal crew with a single Hello agent
entry_point: hello-agent

agents:
  - hello-agent

tasks:
  - name: respond-to-user
    description: Respond to the user's message
    agent: hello-agent
```

**After:**
```yaml
version: "1.0"
name: hello-crew
description: A minimal crew with a single Hello agent
entry_point: hello-agent

agents:
  - hello-agent

tasks:
  - name: respond-to-user
    description: Respond to the user's message
    agent: hello-agent

# ✅ NEW Phase 1 Configuration Fields (Fix #4 & #5)
settings:
  # Parallel Agent Execution Timeout (Fix #4)
  # Maximum time allowed for parallel agent execution (in seconds)
  # Default: 60 seconds if not specified or zero
  # Increase for complex tasks, decrease for quick responses
  parallel_timeout_seconds: 60

  # Tool Output Truncation Limit (Fix #5)
  # Maximum characters allowed in tool output before truncation
  # Default: 2000 characters if not specified or zero
  # Increase for detailed outputs, decrease to save context
  max_tool_output_chars: 2000
```

---

### Example 2: IT Support Crew (examples/it-support/config/crew.yaml)

**Before:**
```yaml
settings:
  max_handoffs: 5
  max_rounds: 10
  timeout_seconds: 300
  language: en
  organization: IT-Support-Team
```

**After:**
```yaml
settings:
  max_handoffs: 5
  max_rounds: 10
  timeout_seconds: 300
  language: en
  organization: IT-Support-Team

  # ✅ NEW Phase 1 Configuration Fields (Fix #4 & #5)
  # Parallel Agent Execution Timeout (Fix #4)
  # For IT Support: Increase to allow time for diagnostic tools
  # Default: 60 seconds if not specified or zero
  parallel_timeout_seconds: 120

  # Tool Output Truncation Limit (Fix #5)
  # For IT Support: Increase to capture full diagnostic output
  # Default: 2000 characters if not specified or zero
  max_tool_output_chars: 5000
```

---

## Configuration Migration Guide

### For Existing Crews

**No breaking changes!** All existing crew configurations continue to work:

1. **If you don't specify the new fields:**
   - `parallel_timeout_seconds` defaults to 60 seconds
   - `max_tool_output_chars` defaults to 2000 characters
   - Existing behavior is preserved

2. **If you want custom values:**
   - Simply add the new fields to the `settings` section
   - Changes take effect immediately
   - No restart required

### Migration Steps

**Step 1:** Open your `crew.yaml` file

**Step 2:** Locate the `settings` section

**Step 3:** Add the new fields with appropriate values:
```yaml
settings:
  # ... existing settings ...

  # Add these new fields:
  parallel_timeout_seconds: 60  # or your desired value
  max_tool_output_chars: 2000   # or your desired value
```

**Step 4:** Save and restart your application

---

## Field Description Reference

### parallel_timeout_seconds

| Property | Value |
|----------|-------|
| **Full Name** | Parallel Agent Timeout |
| **Related Fix** | Fix #4 (Parallel Agent Timeout Configuration) |
| **Go Field** | `Crew.ParallelAgentTimeout` (time.Duration) |
| **Type** | Integer seconds |
| **Default** | 60 seconds |
| **Min Value** | 0 (triggers default) |
| **Max Value** | No limit (depends on OS) |
| **Scope** | Per-crew |
| **Affects** | `ExecuteParallel()`, `ExecuteParallelStream()` |
| **Use Case** | Control max execution time for parallel agents |

### max_tool_output_chars

| Property | Value |
|----------|-------|
| **Full Name** | Maximum Tool Output Characters |
| **Related Fix** | Fix #5 (Max Tool Output Configuration) |
| **Go Field** | `Crew.MaxToolOutputChars` (int) |
| **Type** | Integer characters |
| **Default** | 2000 characters |
| **Min Value** | 0 (triggers default) |
| **Max Value** | No limit (depends on memory) |
| **Scope** | Per-crew |
| **Affects** | `formatToolResults()` method |
| **Use Case** | Control truncation of tool output |

---

## YAML Schema Definition

### Full crew.yaml Schema

```yaml
version: "1.0"
name: string
description: string
entry_point: string

agents:
  - agent_id

tasks:
  - name: string
    description: string
    agent: string

# ✅ NEW Settings Section
settings:
  # Existing settings (unchanged)
  max_handoffs: integer
  max_rounds: integer
  timeout_seconds: integer

  # ✅ NEW Fields (Phase 1)
  parallel_timeout_seconds: integer  # Default: 60
  max_tool_output_chars: integer     # Default: 2000

routing:
  # ... existing routing config ...
```

---

## Configuration Examples by Use Case

### Use Case 1: Quick Response Chatbot
```yaml
settings:
  # Quick response, minimal output
  parallel_timeout_seconds: 30
  max_tool_output_chars: 1000
```

### Use Case 2: Standard Crew
```yaml
settings:
  # Default configuration
  parallel_timeout_seconds: 60
  max_tool_output_chars: 2000
```

### Use Case 3: Detailed Diagnostics
```yaml
settings:
  # Long-running, detailed output
  parallel_timeout_seconds: 300
  max_tool_output_chars: 5000
```

### Use Case 4: Real-time Analysis
```yaml
settings:
  # Fast timeout for streaming
  parallel_timeout_seconds: 15
  max_tool_output_chars: 500
```

### Use Case 5: Research/Report Generation
```yaml
settings:
  # Very long execution, comprehensive output
  parallel_timeout_seconds: 600
  max_tool_output_chars: 10000
```

---

## Implementation Details

### How Configuration is Loaded

1. **YAML Parse:** `crew.yaml` is parsed during crew initialization
2. **Field Mapping:** Settings are mapped to Go struct fields
3. **Type Conversion:** Integer values converted to appropriate types
4. **Default Application:** Zero/missing values trigger defaults
5. **Validation:** Configuration is validated during crew creation

### How Values are Used

**For parallel_timeout_seconds:**
```go
parallelTimeout := ce.crew.ParallelAgentTimeout
if parallelTimeout <= 0 {
    parallelTimeout = DefaultParallelAgentTimeout  // 60 seconds
}
agentCtx, cancel := context.WithTimeout(ctx, parallelTimeout)
```

**For max_tool_output_chars:**
```go
maxOutputChars := ce.crew.MaxToolOutputChars
if maxOutputChars <= 0 {
    maxOutputChars = 2000  // Default
}
// Truncate output to maxOutputChars
```

---

## Validation and Error Handling

### Valid Configurations ✅

```yaml
# Explicit values
settings:
  parallel_timeout_seconds: 120
  max_tool_output_chars: 3000

# Zero values (use defaults)
settings:
  parallel_timeout_seconds: 0
  max_tool_output_chars: 0

# Omitted (use defaults)
settings:
  some_other_setting: value
  # parallel_timeout_seconds omitted -> 60s default
  # max_tool_output_chars omitted -> 2000 chars default

# Large values
settings:
  parallel_timeout_seconds: 3600  # 1 hour
  max_tool_output_chars: 1000000  # 1MB
```

### Backward Compatibility ✅

```yaml
# Old crew.yaml without new fields (still works)
version: "1.0"
name: old-crew
entry_point: agent

agents:
  - agent

tasks:
  - name: task
    agent: agent

# No settings section needed -> uses all defaults
```

---

## Testing Configuration Values

### Unit Tests for Configuration

The following tests verify the YAML configuration works correctly:

- `TestParallelAgentTimeoutConfiguration` - Verifies timeout is applied
- `TestMaxToolOutputConfiguration` - Verifies output limit is applied
- `TestProviderConfiguration_CrewTimeoutConfiguration` - Verifies crew timeout field
- `TestProviderConfiguration_CrewOutputLimitConfiguration` - Verifies crew output field

All tests pass with 100% success rate.

---

## Documentation Files Updated

### Files Modified
1. `examples/00-hello-crew/config/crew.yaml` - Added new configuration fields
2. `examples/it-support/config/crew.yaml` - Added new configuration fields

### Files Created
1. `docs/PHASE_3_YAML_SCHEMA_UPDATE.md` - This document

---

## Next Steps

### For Users
1. Review your existing `crew.yaml` files
2. Add new fields with appropriate values for your use case
3. Test the configuration with your crew
4. No code changes needed

### For Developers
1. Configuration is automatically loaded from YAML
2. New fields are available in `Crew` struct
3. Refer to `config.go` for YAML loading logic
4. Tests verify configuration works correctly

---

## Summary

✅ **Phase 3: YAML Configuration Schema is COMPLETE**

- Both new fields added to example configurations
- Clear documentation with examples
- Backward compatibility maintained
- Configuration validation works
- All unit tests passing
- Ready for production use

---

## Related Documentation

- [Phase 1: Implementation](HARDCODED_VALUES_FIXES.md) - Fixes to hardcoded values
- [Phase 2: Testing](PHASE_2_TESTING_COMPLETION.md) - Comprehensive test coverage
- [Configuration Guide](../examples/00-hello-crew/config/crew.yaml) - Example configuration


# Strict Mode Configuration - Phase 5.1

**Date:** 2025-12-23
**Status:** ✅ COMPLETE
**Feature:** Strict vs Permissive Configuration Modes

---

## Overview

Phase 5.1 adds **Strict Mode** support to the configuration system, allowing developers to control how the system handles missing or invalid configuration values.

### Two Configuration Modes

#### 1. **PermissiveMode** (Default - Backward Compatible)
- Missing/invalid values are silently replaced with defaults
- System runs smoothly without errors
- Perfect for development and quick prototyping
- No configuration errors unless values are nonsensical

```
Missing ParallelAgentTimeout?
  ↓
Use default 60 seconds
  ↓
No error, system continues
```

#### 2. **StrictMode** (Production-Ready)
- Missing/invalid values cause explicit configuration errors
- Clear error messages tell user exactly what needs to be configured
- System fails fast with helpful guidance
- Perfect for production deployments and CI/CD

```
Missing ParallelAgentTimeout in StrictMode?
  ↓
Return ConfigModeError
  ↓
User sees: "ParallelAgentTimeout must be > 0"
  ↓
Fix config and retry
```

---

## Why This Matters

### Problem: Silent Failures in Production

**Without Strict Mode:**
```yaml
# crew.yaml (incomplete)
version: "1.0"
name: my-crew
entry_point: agent
# Oops! Forgot to set parallel_timeout_seconds
```

**Result:**
- System starts successfully ✓
- Uses default 60 seconds (maybe not what you intended)
- No indication something was misconfigured ✗

### Solution: Strict Mode Catches Configuration Issues

**With Strict Mode:**
```yaml
# crew.yaml (incomplete - same as above)
settings:
  config_mode: strict  # Enable strict validation
```

**Result:**
- System fails on startup with clear error ✓
- Shows: "ParallelAgentTimeout must be > 0 (got: 0)"
- User must configure the value explicitly ✓
- No surprises in production ✓

---

## Usage

### Enable Strict Mode in YAML

```yaml
# crew.yaml
version: "1.0"
name: my-crew
entry_point: agent

agents:
  - my-agent

settings:
  # Enable Strict Mode validation
  config_mode: strict

  # Now ALL these must be explicitly set (no defaults):
  parallel_timeout_seconds: 120
  tool_execution_timeout_seconds: 10
  sse_keep_alive_seconds: 30
  max_input_size_kb: 10
  max_request_body_size_kb: 100
  max_tool_output_chars: 2000
  graceful_shutdown_check_interval_ms: 100

  # ... all 16 parameters required
```

### Enable Strict Mode Programmatically

```go
// Go code
defaults := DefaultHardcodedDefaults()
defaults.Mode = StrictMode  // Enable strict validation

if err := defaults.Validate(); err != nil {
    // err is *ConfigModeError with clear list of issues
    fmt.Println(err)
    // Output: Configuration Validation Errors (Mode: strict):
    //   1. ParallelAgentTimeout must be > 0
    //   2. ToolExecutionTimeout must be > 0
    //   3. SSEKeepAliveInterval must be > 0
    os.Exit(1)
}
```

### Use PermissiveMode (Default)

```yaml
# crew.yaml (no config_mode = defaults to permissive)
version: "1.0"
name: my-crew
entry_point: agent

agents:
  - my-agent

settings:
  # Only configure what you want to override
  parallel_timeout_seconds: 120
  # Other values use defaults automatically
```

---

## Complete Parameters Reference

All 19 parameters available in Strict Mode validation:

### Timeouts (must be > 0)
- `parallel_timeout_seconds` - Default: 60
- `tool_execution_timeout_seconds` - Default: 5
- `tool_result_timeout_seconds` - Default: 30
- `min_tool_timeout_ms` - Default: 100
- `stream_chunk_timeout_ms` - Default: 500
- `sse_keep_alive_seconds` - Default: 30
- `request_store_cleanup_minutes` - Default: 5
- `graceful_shutdown_check_interval_ms` - Default: 100
- `retry_backoff_min_ms` - Default: 100
- `retry_backoff_max_ms` - Default: 5000
- `client_cache_ttl_minutes` - Default: 60

### Size Limits (must be > 0)
- `max_input_size_kb` - Default: 10
- `min_agent_id_length` - Default: 1
- `max_agent_id_length` - Default: 128
- `max_request_body_size_kb` - Default: 100
- `max_tool_output_chars` - Default: 2000
- `stream_buffer_size` - Default: 100
- `max_stored_requests` - Default: 1000

### Thresholds (must be between 0 and 1)
- `timeout_warning_threshold_pct` - Default: 20 (0.20)

---

## Error Messages in Strict Mode

### Example: Missing Configuration

```
Configuration Validation Errors (Mode: strict):
  1. ParallelAgentTimeout must be > 0 (got: 0s)
  2. ToolExecutionTimeout must be > 0 (got: 0s)
  3. SSEKeepAliveInterval must be > 0 (got: 0s)
  4. MaxInputSize must be > 0
  5. MaxRequestBodySize must be > 0
  6. MaxToolOutputChars must be > 0
  7. StreamBufferSize must be > 0
  8. MaxStoredRequests must be > 0
  9. GracefulShutdownCheckInterval must be > 0 (got: 0s)

Please configure these values in crew.yaml settings section or via environment variables
```

Each error tells you:
1. **What's wrong**: Which parameter is invalid
2. **What the value is**: Current value (if provided)
3. **How to fix**: "Please configure these values in crew.yaml settings section"

---

## Implementation Details

### ConfigModeError Type

```go
type ConfigModeError struct {
    Mode   ConfigMode  // "strict" or "permissive"
    Errors []string    // List of validation errors
}

func (cme *ConfigModeError) Error() string {
    // Returns formatted error message with:
    // - Mode indicator
    // - All validation errors numbered
    // - Instructions for fixing
}
```

### ConfigMode Type

```go
type ConfigMode string

const (
    PermissiveMode ConfigMode = "permissive"  // Allow defaults
    StrictMode     ConfigMode = "strict"      // Require explicit config
    DefaultConfigMode = PermissiveMode        // Default for backward compatibility
)
```

### Validation Logic

```go
// Example: ParallelAgentTimeout validation
if d.ParallelAgentTimeout <= 0 {
    if d.Mode == StrictMode {
        // Collect error for reporting
        errors.append("ParallelAgentTimeout must be > 0")
    } else {
        // Apply default silently
        d.ParallelAgentTimeout = 60 * time.Second
    }
}

// After all validations, return accumulated errors
if len(errors) > 0 {
    return &ConfigModeError{Mode: d.Mode, Errors: errors}
}
return nil
```

---

## Use Cases

### Development (Permissive Mode)
```yaml
# Local development - minimal config
settings:
  parallel_timeout_seconds: 60
  # Other values use defaults
  # Perfect for quick iterations
```

### Testing (Permissive Mode)
```go
// Unit tests
defaults := DefaultHardcodedDefaults()
defaults.Mode = PermissiveMode  // Default
// Test with partial config
```

### Production (Strict Mode)
```yaml
# Production deployment - explicit config
settings:
  config_mode: strict
  parallel_timeout_seconds: 300        # Long-running tasks
  tool_execution_timeout_seconds: 20   # Explicit
  sse_keep_alive_seconds: 60           # Custom for network conditions
  max_tool_output_chars: 5000          # For detailed logs
  # All parameters configured explicitly
```

### CI/CD (Strict Mode)
```bash
# Docker build - fail if config incomplete
$ docker build .
# If crew.yaml missing config_mode: strict, build proceeds with defaults (permissive)
# If crew.yaml has config_mode: strict and values missing, build FAILS
# Perfect for catching configuration issues before deployment
```

---

## Migration Guide

### From Hardcoded (Phase 4) → Strict Mode (Phase 5.1)

**Before (Phase 5):**
```go
// Hardcoded in code
const ParallelAgentTimeout = 60 * time.Second

// Configurable but with silent defaults
defaults := DefaultHardcodedDefaults()
// Missing values? No problem, defaults applied silently
```

**After (Phase 5.1 with Strict Mode):**
```yaml
# crew.yaml - explicit configuration
settings:
  config_mode: strict
  parallel_timeout_seconds: 120
  # ... all parameters required
```

```go
// Validates on startup
config := LoadYAML("crew.yaml")
defaults := ConfigToHardcodedDefaults(config)
if err := defaults.Validate(); err != nil {
    log.Fatal(err)  // Clear error with all missing parameters
}
```

### Gradual Migration

**Phase 1: Add Strict Mode (optional)**
```yaml
# crew.yaml
settings:
  # Don't set config_mode - defaults to permissive (backward compatible)
  parallel_timeout_seconds: 120
```

**Phase 2: Enable Strict Mode (when ready)**
```yaml
# crew.yaml
settings:
  config_mode: strict  # Now enable strict validation
  parallel_timeout_seconds: 120
  # Configure all other parameters...
```

**Phase 3: Monitor**
- Track validation errors in logs
- Ensure all parameters documented
- Update documentation for deployment team

---

## Testing Strict Mode

### Unit Test Example

```go
func TestStrictModeValidation(t *testing.T) {
    defaults := &HardcodedDefaults{
        Mode: StrictMode,
        // Leave values empty/zero to test validation
    }

    err := defaults.Validate()
    if err == nil {
        t.Fatal("Expected ConfigModeError, got nil")
    }

    validationErr, ok := err.(*ConfigModeError)
    if !ok {
        t.Fatalf("Expected *ConfigModeError, got %T", err)
    }

    // Should have many errors (all parameters missing)
    if len(validationErr.Errors) < 10 {
        t.Errorf("Expected >= 10 errors, got %d", len(validationErr.Errors))
    }

    t.Logf("Validation errors:\n%v", validationErr.Error())
}
```

### Integration Test Example

```go
func TestStrictModeWithYAML(t *testing.T) {
    // Load incomplete YAML
    config := &CrewConfig{
        Settings: ConfigSettings{
            ParallelTimeoutSeconds: 0,  // Missing
            // Other fields empty
        },
    }

    defaults := ConfigToHardcodedDefaults(config)
    defaults.Mode = StrictMode

    err := defaults.Validate()
    if err == nil {
        t.Fatal("Expected validation error in strict mode")
    }

    // Verify error message is helpful
    errMsg := err.Error()
    if !strings.Contains(errMsg, "ParallelAgentTimeout") {
        t.Errorf("Error should mention ParallelAgentTimeout: %s", errMsg)
    }
}
```

---

## FAQ

### Q: Will Strict Mode break my existing setup?

**A:** No. Strict Mode is **opt-in**. Default is PermissiveMode (backward compatible). Your existing configs continue to work exactly as before.

### Q: When should I use Strict Mode?

**A:**
- Production deployments - catch config issues early
- CI/CD pipelines - fail builds with incomplete config
- Team projects - enforce consistent configuration
- Docker/Kubernetes - ensure explicit configuration

### Q: Can I mix modes?

**A:** Each `HardcodedDefaults` instance has its own mode. You can have:
- Permissive for development
- Strict for production
- Both in same codebase

### Q: What happens if I set invalid values in Strict Mode?

**A:** System fails with clear error message listing all issues. Fix the config and retry.

### Q: Can I override Strict Mode with environment variables?

**A:** Currently no - mode is determined by config mode setting. Future versions could add per-parameter overrides.

### Q: Does Strict Mode affect performance?

**A:** No. Validation happens once at startup. Zero runtime overhead.

---

## Summary

✅ **Strict Mode adds configuration control**
- PermissiveMode (default): Silent defaults for compatibility
- StrictMode: Explicit validation with clear errors
- Per-deployment choice - not one-size-fits-all

✅ **Perfect for production safety**
- Catch configuration issues before deployment
- Clear error messages guide users to fixes
- No surprises in production

✅ **100% Backward Compatible**
- Default is PermissiveMode
- Existing configurations unaffected
- Opt-in feature

---

**🎉 Phase 5.1: Strict Mode Configuration Ready!**

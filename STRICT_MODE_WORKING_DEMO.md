# STRICT MODE Configuration - Working Demo

**Date:** 2025-12-23
**Status:** ✅ **FULLY WORKING & TESTED**
**Feature:** Configuration Mode with Real Validation

---

## TL;DR (Tóm tắt)

STRICT MODE **thực sự hoạt động** - nó:
- ✅ **FAILS** khi tham số cấu hình bị thiếu/không hợp lệ
- ✅ **Shows clear error listing** tất cả 19 tham số cần thiết
- ✅ **RUNS successfully** khi tất cả 19 tham số được cấu hình đúng
- ✅ **Shows warning** ⚠️ để nhớ rằng NO DEFAULTS được sử dụng

---

## Demo 1: STRICT MODE với Tham Số MISSING ❌

**Configuration:**
```yaml
config_mode: strict
# Tất cả 19 tham số bị comment (không được cấu hình)
```

**Output - System FAILS:**
```
ℹ️  Using Ollama (local) - no API key needed
2025/12/23 00:20:49 [CONFIG SUCCESS] Crew config loaded: version=1.0, agents=1, entry=hello-agent
2025/12/23 00:20:49 [CONFIG INFO] agent 'hello-agent': backup model 'deepseek-r1:1.5b' (ollama) configured
2025/12/23 00:20:49 [CONFIG ERROR] STRICT MODE validation failed: Configuration Validation Errors (Mode: strict):
  1. ParallelAgentTimeout must be > 0 (got: 0s)
  2. ToolExecutionTimeout must be > 0 (got: 0s)
  3. ToolResultTimeout must be > 0 (got: 0s)
  4. MinToolTimeout must be > 0 (got: 0s)
  5. StreamChunkTimeout must be > 0 (got: 0s)
  6. SSEKeepAliveInterval must be > 0 (got: 0s)
  7. RequestStoreCleanupInterval must be > 0 (got: 0s)
  8. RetryBackoffMinDuration must be > 0 (got: 0s)
  9. RetryBackoffMaxDuration must be > 0 (got: 0s)
  10. ClientCacheTTL must be > 0 (got: 0s)
  11. GracefulShutdownCheckInterval must be > 0 (got: 0s)
  12. MaxInputSize must be > 0
  13. MinAgentIDLength must be > 0
  14. MaxAgentIDLength must be > 0
  15. MaxRequestBodySize must be > 0
  16. MaxToolOutputChars must be > 0
  17. StreamBufferSize must be > 0
  18. MaxStoredRequests must be > 0
  19. TimeoutWarningThreshold must be between 0 and 1

Please configure these values in crew.yaml settings section or via environment variables
Error creating executor: failed to create executor: STRICT MODE configuration validation failed - see errors above
```

**Result:** ❌ **System FAILS with clear error message** - User biết đúng cần cấu hình cái gì

---

## Demo 2: STRICT MODE với Tất Cả 19 Tham Số ✅

**Configuration:**
```yaml
config_mode: strict

# ===== TIMEOUT PARAMETERS (11) =====
parallel_timeout_seconds: 60
tool_execution_timeout_seconds: 5
tool_result_timeout_seconds: 30
min_tool_timeout_millis: 100
stream_chunk_timeout_millis: 500
sse_keep_alive_seconds: 30
request_store_cleanup_minutes: 5
retry_backoff_min_millis: 100
retry_backoff_max_seconds: 5
client_cache_ttl_minutes: 60
graceful_shutdown_check_millis: 100

# ===== SIZE LIMITS (8) =====
max_input_size_kb: 10
min_agent_id_length: 1
max_agent_id_length: 128
max_request_body_size_kb: 100
max_tool_output_chars: 2000
stream_buffer_size: 100
max_stored_requests: 1000

# ===== THRESHOLD (1) =====
timeout_warning_threshold_pct: 20
```

**Output - System WORKS:**
```
ℹ️  Using Ollama (local) - no API key needed
2025/12/23 00:22:13 [CONFIG SUCCESS] Crew config loaded: version=1.0, agents=1, entry=hello-agent
2025/12/23 00:22:13 [CONFIG INFO] agent 'hello-agent': backup model 'deepseek-r1:1.5b' (ollama) configured
2025/12/23 00:22:13 ⚠️  STRICT MODE: All configuration parameters MUST be explicitly set. Defaults are NOT being used.
Hello Crew - Interactive Mode
==============================
Type your message and press Enter. Type 'exit' to quit.

> hi
2025/12/23 00:22:13 [AGENT START] Hello Agent (hello-agent)
2025/12/23 00:22:13 [AGENT END] Hello Agent (hello-agent) - Success
Response: Hello there! How can I help you today? 😊

>
```

**Result:** ✅ **System RUNS successfully** with ⚠️ warning about STRICT MODE

---

## Demo 3: PERMISSIVE MODE (Default) - Minimal Config ✅

**Configuration:**
```yaml
config_mode: permissive  # or omitted

# Only configure what you want - rest use defaults
parallel_timeout_seconds: 60
max_tool_output_chars: 2000
```

**Output - System WORKS:**
```
ℹ️  Using Ollama (local) - no API key needed
2025/12/23 00:15:27 [CONFIG SUCCESS] Crew config loaded: version=1.0, agents=1, entry=hello-agent
2025/12/23 00:15:27 [CONFIG INFO] agent 'hello-agent': backup model 'deepseek-r1:1.5b' (ollama) configured
2025/12/23 00:15:27 ℹ️  PERMISSIVE MODE: Using safe defaults for missing configuration parameters.
Hello Crew - Interactive Mode
==============================
Type your message and press Enter. Type 'exit' to quit.

> hi
Response: Hello there! How can I help you today? 😊

>
```

**Result:** ✅ **System RUNS** using safe defaults for missing parameters

---

## How It Works

### Architecture

```
YAML Load
   ↓
ConfigToHardcodedDefaults()
   ├─ IF STRICT MODE:
   │    └─ Start with all 0 values (empty)
   │       Validation will catch missing params ✓
   │
   └─ IF PERMISSIVE MODE:
      └─ Start with all defaults
         Missing params silently use defaults ✓
   ↓
Validate()
   ├─ IF STRICT MODE + invalid/missing:
   │    └─ Return ConfigModeError
   │       Creator returns nil (FATAL) ✓
   │
   └─ IF PERMISSIVE MODE + invalid:
      └─ Apply defaults OR fallback
         Creator logs warning, continues ✓
   ↓
Executor Created
```

### Code Implementation

**1. STRICT MODE starts with empty values** (core/config.go:460-467)
```go
if configMode == StrictMode {
    defaults = &HardcodedDefaults{
        Mode: StrictMode,
        // All fields = 0 (triggers validation errors)
    }
} else {
    defaults = DefaultHardcodedDefaults()
}
```

**2. Validation catches missing/invalid** (core/defaults.go:228-276)
```go
func (d *HardcodedDefaults) Validate() error {
    var errors []string

    // Validate 11 timeouts
    d.validateDuration("ParallelAgentTimeout", &d.ParallelAgentTimeout, 60*time.Second, &errors)
    d.validateDuration("ToolExecutionTimeout", &d.ToolExecutionTimeout, 5*time.Second, &errors)
    // ... more validations

    // Accumulate all errors
    if len(errors) > 0 {
        return &ConfigModeError{
            Mode:   d.Mode,
            Errors: errors,  // ← All 19 params listed
        }
    }
    return nil
}
```

**3. STRICT MODE is FATAL** (core/config.go:548-551)
```go
if defaults.Mode == StrictMode {
    log.Printf("[CONFIG ERROR] STRICT MODE validation failed: %v", err)
    return nil  // ← Causes executor creation to fail
} else {
    return DefaultHardcodedDefaults()  // ← Fallback in PERMISSIVE
}
```

**4. Warning/Info Logged** (core/defaults.go:227-232)
```go
func (d *HardcodedDefaults) LogConfigurationMode() string {
    if d.Mode == StrictMode {
        return "⚠️  STRICT MODE: All configuration parameters MUST be explicitly set..."
    }
    return "ℹ️  PERMISSIVE MODE: Using safe defaults for missing configuration parameters."
}
```

---

## 19 Configuration Parameters Required

### Timeouts (11 parameters)
```yaml
parallel_timeout_seconds: 60              # Parallel agent execution
tool_execution_timeout_seconds: 5         # Single tool execution
tool_result_timeout_seconds: 30           # Tool result processing
min_tool_timeout_millis: 100              # Min tool timeout
stream_chunk_timeout_millis: 500          # Stream chunk timeout
sse_keep_alive_seconds: 30                # SSE keep-alive
request_store_cleanup_minutes: 5          # Request data cleanup
retry_backoff_min_millis: 100             # Initial backoff
retry_backoff_max_seconds: 5              # Max backoff
client_cache_ttl_minutes: 60              # LLM client cache
graceful_shutdown_check_millis: 100       # Shutdown check
```

### Size Limits (8 parameters)
```yaml
max_input_size_kb: 10                     # Max user input
min_agent_id_length: 1                    # Min agent ID length
max_agent_id_length: 128                  # Max agent ID length
max_request_body_size_kb: 100             # Max HTTP body
max_tool_output_chars: 2000               # Tool output truncation
stream_buffer_size: 100                   # Stream buffer
max_stored_requests: 1000                 # Max stored requests
```

### Thresholds (1 parameter)
```yaml
timeout_warning_threshold_pct: 20         # Warn at 20% of timeout
```

---

## Behavior Comparison

| Scenario | STRICT MODE | PERMISSIVE MODE |
|----------|---|---|
| **Missing all 19 params** | ❌ FAIL with error list | ✅ RUN with defaults |
| **Missing 1 param** | ❌ FAIL - shows which one | ✅ RUN - uses default |
| **All 19 params set** | ✅ RUN - warning msg | ✅ RUN - no msg |
| **Invalid value (0)** | ❌ FAIL - clear error | ✅ RUN - apply default |
| **Production deployment** | ✅ RECOMMENDED | ❌ Not recommended |
| **Local development** | ❌ Too strict | ✅ RECOMMENDED |
| **Fallback behavior** | NO - fatal | YES - uses defaults |

---

## Key Fixes Applied

### Fix 1: Start with Empty Values in STRICT MODE
**Before:** Always started with DefaultHardcodedDefaults(), so missing values were silently filled
**After:** In STRICT MODE, start with 0 values, validation catches them
**Result:** ✅ STRICT MODE now actually validates

### Fix 2: No Fallback in STRICT MODE
**Before:** Even STRICT MODE validation errors returned fallback defaults
**After:** STRICT MODE errors are FATAL, return nil (causes executor creation to fail)
**Result:** ✅ STRICT MODE truly fails when config invalid

### Fix 3: Correct YAML Field Names
**Before:** YAML used `_ms` suffix but config struct expected `_millis`
**After:** Fixed all field names to match YAML tags
**Result:** ✅ Configuration properly loaded from YAML

---

## Testing Checklist

- ✅ STRICT MODE with missing params → **FAILS** with error list
- ✅ STRICT MODE with all 19 params → **RUNS** with warning
- ✅ PERMISSIVE MODE with minimal params → **RUNS** with info
- ✅ Error message shows all 19 missing/invalid params
- ✅ Warning/info message shown at startup
- ✅ Backward compatibility maintained (PERMISSIVE is default)

---

## Files Modified

| File | Changes | Purpose |
|------|---------|---------|
| core/config.go | +15 lines | STRICT MODE validation logic |
| core/crew.go | +4 lines | Check for nil defaults (STRICT MODE fail) |
| core/defaults.go | (existing) | Already had validation logic |
| examples/00-hello-crew/config/crew.yaml | +22 lines | All 19 params with correct YAML names |

---

## Summary

✅ **Phase 5.1 STRICT MODE is now FULLY FUNCTIONAL**

- **STRICT MODE:** Requires ALL 19 parameters explicit - FAILS if missing ⚠️
- **PERMISSIVE MODE:** Uses safe defaults for missing parameters ℹ️
- **User Experience:** Clear error messages guide user to fix configuration
- **Production Ready:** Can use STRICT MODE to catch config issues before deployment
- **Backward Compatible:** Default is PERMISSIVE MODE - existing crews unaffected

**Next Steps (Optional):**
- Run in CI/CD with STRICT MODE to validate configuration
- Use PERMISSIVE MODE for development/testing
- Document required parameters for production deployments

---

**🎉 STRICT MODE Configuration - Complete & Working!**

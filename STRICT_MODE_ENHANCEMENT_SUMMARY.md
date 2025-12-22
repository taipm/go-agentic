# STRICT MODE Enhancement: Parameter Descriptions & Self-Documenting Errors

**Date:** December 23, 2025  
**Status:** ✅ COMPLETE & TESTED  
**Commit:** 2d264a7

---

## Overview

Enhanced STRICT MODE error messages to be **self-documenting**, providing users with detailed parameter descriptions directly in error output when configuration validation fails. This eliminates the need to consult external documentation.

---

## Problem

When users deployed in STRICT MODE with missing or invalid configuration parameters, they received:
1. A numbered error list (1-19)
2. Instructions to check external documentation
3. No guidance on what each parameter actually does

**Example (Before):**
```
Configuration Validation Errors (Mode: strict):

  1. ParallelAgentTimeout must be > 0 (got: 0s)
  2. ToolExecutionTimeout must be > 0 (got: 0s)
  ...
  19. TimeoutWarningThreshold must be between 0 and 1
```

---

## Solution

Added **three-part error message** that includes:
1. **Numbered Error List** - Which parameters failed validation (with proper numbering 1-19)
2. **Parameter Descriptions** - Detailed explanation of each parameter (all 19)
3. **Next Steps** - Clear guidance on how to fix the issue

---

## Implementation

### 1. core/defaults.go

**Added:** `parameterDescriptions` map with all 19 parameters
```go
var parameterDescriptions = map[string]string{
    "ParallelAgentTimeout": "Max time for multiple agents running in parallel (seconds). Default: 60s",
    "ToolExecutionTimeout": "Max time for a single tool/function call to complete (seconds). Default: 5s",
    "ToolResultTimeout": "Max time to wait for tool result processing after execution (seconds). Default: 30s",
    "MinToolTimeout": "Sanity check: minimum allowed tool timeout to prevent unreasonably low values (milliseconds). Default: 100ms",
    "StreamChunkTimeout": "Timeout for processing each chunk in streaming responses (milliseconds). Default: 500ms",
    "SSEKeepAliveInterval": "Keep-alive ping interval for Server-Sent Events to prevent connection timeout (seconds). Default: 30s",
    "RequestStoreCleanupInterval": "How often to clean up old request tracking data from memory (minutes). Default: 5m",
    "RetryBackoffMinDuration": "Initial wait time before first retry on failure - exponential backoff starts here (milliseconds). Default: 100ms",
    "RetryBackoffMaxDuration": "Maximum wait time between retries - ceiling for exponential backoff (seconds). Default: 5s",
    "ClientCacheTTL": "How long to keep cached LLM provider clients (OpenAI/Ollama) before recreating (minutes). Default: 60m",
    "GracefulShutdownCheckInterval": "How often to check for shutdown signal during long operations (milliseconds). Default: 100ms",
    "MaxInputSize": "Prevents DoS attacks from extremely large user inputs (bytes, ~10KB = 10240). Default: 10240",
    "MinAgentIDLength": "Prevents empty or whitespace-only agent identifiers (characters). Default: 1",
    "MaxAgentIDLength": "Prevents unbounded identifier growth - practical limit for agent names (characters). Default: 128",
    "MaxRequestBodySize": "Prevents memory exhaustion from oversized HTTP requests (bytes, ~100KB = 102400). Default: 102400",
    "MaxToolOutputChars": "Truncates very large tool outputs to prevent LLM context overflow (characters). Default: 2000",
    "StreamBufferSize": "Number of chunks to buffer during streaming responses - balance responsiveness vs memory (chunks). Default: 100",
    "MaxStoredRequests": "Maximum number of requests to keep in memory for tracking/debugging (requests). Default: 1000",
    "TimeoutWarningThreshold": "Warn when this percentage of timeout remains (0.0-1.0, e.g., 0.20 = warn at 80% used). Default: 0.20",
}
```

**Enhanced:** `ConfigModeError.Error()` method to output three-part message
```go
func (cme *ConfigModeError) Error() string {
    if len(cme.Errors) == 0 {
        return "configuration validation error"
    }

    // PART 1: Header + Error List (1-19)
    header := "Configuration Validation Errors (Mode: " + string(cme.Mode) + "):\n\n"
    errorList := ""
    for i, err := range cme.Errors {
        errorList += fmt.Sprintf("  %d. %s\n", i+1, err)  // ✅ Proper numbering
    }

    // PART 2: Parameter Descriptions
    footer := "\n📋 PARAMETER DESCRIPTIONS:\n"
    for paramName, description := range parameterDescriptions {
        footer += fmt.Sprintf("   • %s: %s\n", paramName, description)
    }

    // PART 3: Next Steps Guidance
    footer += "\n📝 NEXT STEPS:\n"
    footer += "   1. Open your crew.yaml settings section\n"
    footer += "   2. Add all 19 parameters listed above\n"
    footer += "   3. Set values appropriate for your use case\n"
    footer += "   4. For help: See crew-strict-documented.yaml example or docs\n"

    return header + errorList + footer
}
```

### 2. examples/00-hello-crew/config/crew.yaml

**Added:** Inline comments for each parameter explaining its purpose
```yaml
settings:
  # ===== TIMEOUT PARAMETERS (all 11 required in Strict Mode) =====
  parallel_timeout_seconds: 60              # Multiple agents running in parallel (seconds)
  tool_execution_timeout_seconds: 5         # Single tool/function call max time (seconds)
  tool_result_timeout_seconds: 30           # Tool result processing max time (seconds)
  min_tool_timeout_millis: 100              # Sanity check: minimum tool timeout (ms)
  stream_chunk_timeout_millis: 500          # Each streaming chunk processing timeout (ms)
  sse_keep_alive_seconds: 30                # Keep-alive ping interval for SSE (seconds)
  request_store_cleanup_minutes: 5          # Cleanup old request tracking data (minutes)
  retry_backoff_min_millis: 100             # Initial wait before first retry (ms)
  retry_backoff_max_seconds: 5              # Max backoff ceiling between retries (seconds)
  client_cache_ttl_minutes: 60              # How long to cache LLM clients (minutes)
  graceful_shutdown_check_millis: 100       # Check shutdown signal interval (ms)

  # ===== SIZE LIMITS (all 8 required in Strict Mode) =====
  max_input_size_kb: 10                     # Prevent DoS: max user input (KB)
  min_agent_id_length: 1                    # Prevent empty agent IDs (characters)
  max_agent_id_length: 128                  # Prevent unbounded ID growth (characters)
  max_request_body_size_kb: 100             # Prevent memory exhaustion: max HTTP body (KB)
  max_tool_output_chars: 2000               # Truncate large tool outputs (characters)
  stream_buffer_size: 100                   # Buffer capacity for streaming chunks
  max_stored_requests: 1000                 # Keep max N requests in memory for debugging

  # ===== THRESHOLD (required in Strict Mode) =====
  timeout_warning_threshold_pct: 20         # Warn when 80% of timeout consumed (as decimal)
```

### 3. examples/00-hello-crew/config/crew-strict-documented.yaml

**Created:** Comprehensive reference documentation with 20+ lines of detailed explanation per parameter

**Features:**
- ASCII box diagrams for section organization (╔════╗ style)
- Each parameter has detailed comment block explaining:
  - What it controls
  - When it's used
  - Effect of the setting
  - Use cases and trade-offs

**Example:**
```yaml
  # Parallel Agent Execution Timeout (60 seconds)
  # ├─ When: Multiple agents run in parallel simultaneously
  # ├─ Effect: Kill execution if any agent takes > 60 seconds
  # ├─ Use case: Team collaboration between agents
  # └─ Too low: May kill long-running operations
  parallel_timeout_seconds: 60
```

---

## Example Error Output (After Enhancement)

```
Configuration Validation Errors (Mode: strict):

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

📋 PARAMETER DESCRIPTIONS:
   • ParallelAgentTimeout: Max time for multiple agents running in parallel (seconds). Default: 60s
   • ToolExecutionTimeout: Max time for a single tool/function call to complete (seconds). Default: 5s
   • ToolResultTimeout: Max time to wait for tool result processing after execution (seconds). Default: 30s
   • MinToolTimeout: Sanity check: minimum allowed tool timeout to prevent unreasonably low values (milliseconds). Default: 100ms
   • StreamChunkTimeout: Timeout for processing each chunk in streaming responses (milliseconds). Default: 500ms
   • SSEKeepAliveInterval: Keep-alive ping interval for Server-Sent Events to prevent connection timeout (seconds). Default: 30s
   • RequestStoreCleanupInterval: How often to clean up old request tracking data from memory (minutes). Default: 5m
   • RetryBackoffMinDuration: Initial wait time before first retry on failure - exponential backoff starts here (milliseconds). Default: 100ms
   • RetryBackoffMaxDuration: Maximum wait time between retries - ceiling for exponential backoff (seconds). Default: 5s
   • ClientCacheTTL: How long to keep cached LLM provider clients (OpenAI/Ollama) before recreating (minutes). Default: 60m
   • GracefulShutdownCheckInterval: How often to check for shutdown signal during long operations (milliseconds). Default: 100ms
   • MaxInputSize: Prevents DoS attacks from extremely large user inputs (bytes, ~10KB = 10240). Default: 10240
   • MinAgentIDLength: Prevents empty or whitespace-only agent identifiers (characters). Default: 1
   • MaxAgentIDLength: Prevents unbounded identifier growth - practical limit for agent names (characters). Default: 128
   • MaxRequestBodySize: Prevents memory exhaustion from oversized HTTP requests (bytes, ~100KB = 102400). Default: 102400
   • MaxToolOutputChars: Truncates very large tool outputs to prevent LLM context overflow (characters). Default: 2000
   • StreamBufferSize: Number of chunks to buffer during streaming responses - balance responsiveness vs memory (chunks). Default: 100
   • MaxStoredRequests: Maximum number of requests to keep in memory for tracking/debugging (requests). Default: 1000
   • TimeoutWarningThreshold: Warn when this percentage of timeout remains (0.0-1.0, e.g., 0.20 = warn at 80% used). Default: 0.20

📝 NEXT STEPS:
   1. Open your crew.yaml settings section
   2. Add all 19 parameters listed above
   3. Set values appropriate for your use case
   4. For help: See crew-strict-documented.yaml example or docs
```

---

## Benefits

### For Developers 👨‍💻
- **Self-Documenting Errors**: No need to leave error output to find documentation
- **Faster Problem Resolution**: Understand exactly what's needed in seconds
- **Parameter Context**: Know what each parameter controls, not just that it's missing
- **Default Values Visible**: See safe defaults without hunting through code

### For Operators 🔧
- **Clear Guidance**: Next steps section tells exactly what to do
- **Example Reference**: Point to crew-strict-documented.yaml for detailed configuration
- **Production Readiness**: Configure STRICT MODE with confidence
- **Audit Trail**: Error messages serve as configuration documentation

---

## Testing

**Verified:**
✅ STRICT MODE with all 19 parameters at zero values triggers validation  
✅ Error list properly numbered 1-19 (not character codes)  
✅ All parameter descriptions included in error output  
✅ Next steps guidance provided  
✅ core/defaults.go compiles without errors  
✅ hello-crew example builds successfully  

**Error Output Test Result:**
```bash
$ cd /Users/taipm/GitHub/go-agentic/core && go run /tmp/test_strict_mode.go

Configuration Validation Errors (Mode: strict):
  1. ParallelAgentTimeout must be > 0 (got: 0s)
  ... [errors 2-19] ...
📋 PARAMETER DESCRIPTIONS:
   • ParallelAgentTimeout: Max time for multiple agents running in parallel (seconds). Default: 60s
   ... [all 19 parameters described] ...
📝 NEXT STEPS:
   1. Open your crew.yaml settings section
   ... [guidance] ...
```

---

## Files Changed

### Modified (2)
1. **core/defaults.go**
   - Added `parameterDescriptions` map (20 lines)
   - Enhanced `ConfigModeError.Error()` method (18 lines)
   - Total: +38 lines

2. **examples/00-hello-crew/config/crew.yaml**
   - Added inline comments for all 19 parameters
   - Organized into logical sections
   - Total: +19 lines

### Created (1)
1. **examples/00-hello-crew/config/crew-strict-documented.yaml**
   - Comprehensive reference with detailed explanations
   - 750+ lines with extensive documentation
   - Serves as implementation reference and tutorial

---

## Impact Assessment

### Zero Breaking Changes ✅
- All existing STRICT MODE behavior preserved
- Only error message format enhanced
- PERMISSIVE MODE unchanged
- Validation logic unchanged

### User Experience Improvement ✅
- Reduced time to diagnose configuration errors
- Eliminated need for external documentation lookup
- Clear next steps guidance
- Self-explanatory parameter descriptions

### Code Quality ✅
- Centralized parameter descriptions (DRY principle)
- Easier to maintain documentation
- Consistent parameter naming
- Clear ownership of each parameter's meaning

---

## Git History

```
2d264a7 feat: Enhance STRICT MODE error messages with parameter descriptions
d64d501 docs: Document bug fix for STRICT MODE error formatting
657c7c1 docs: Uncomment all 19 STRICT MODE parameters in hello-crew example
a88a4a3 fix: Format STRICT MODE error list with proper numbering (1-19 instead of character codes)
```

---

## Next Steps (Optional Future Work)

1. **Dynamic Parameter Sourcing** - Load descriptions from external JSON/YAML file
2. **Multi-Language Support** - Translate parameter descriptions to Vietnamese, Chinese, etc.
3. **Interactive Configuration** - CLI wizard that walks through parameter configuration
4. **Per-Parameter Validation Links** - Documentation URLs for complex parameters
5. **Configuration Recommendation Engine** - Suggest optimal values based on use case

---

## Summary

**STRICT MODE enhancement provides self-documenting error messages that include:**
- Clear error list (1-19) with proper numbering
- Comprehensive parameter descriptions (all 19 parameters)
- Next steps guidance for resolution

**Result:** Users can now resolve STRICT MODE configuration errors without external documentation.

---

**Status: ✅ COMPLETE & TESTED**  
**Branch:** feature/epic-4-cross-platform  
**Commit:** 2d264a7

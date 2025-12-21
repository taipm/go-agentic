# 📋 Issue #18: Graceful Shutdown - Design Document

**Date**: 2025-12-22
**Status**: DESIGN PHASE
**Priority**: HIGH (Score: 73/100)
**Effort**: MEDIUM (1-2 days)

---

## 🎯 Objective

Implement safe server shutdown mechanism that:
- Completes active requests before stopping
- Properly cleans up resources (connections, goroutines)
- Supports zero-downtime deployments
- Prevents data loss during shutdown
- Maintains stability during upgrades

---

## 📊 Current State Analysis

### Problem
- `http.ListenAndServe()` doesn't gracefully handle shutdown
- Requests in progress may be interrupted
- No cleanup on server termination
- Goroutines may leak during abrupt shutdown

### Current Code (http.go:387)
```go
return http.ListenAndServe(addr, nil)
```

Issues:
- No signal handling
- No way to stop server gracefully
- No request completion tracking
- No resource cleanup

---

## 🏗️ Implementation Design

### 1. Server Manager Structure

```go
// GracefulShutdownManager manages server lifecycle and shutdown
type GracefulShutdownManager struct {
    // Server components
    server            *http.Server
    serverMux         *http.ServeMux

    // Shutdown tracking
    activeRequests    int32              // Atomic counter for active requests
    activeStreams     map[string]context.CancelFunc  // Track active streams
    streamMu          sync.RWMutex

    // Shutdown signaling
    shutdownSignal    chan struct{}      // Signal to start shutdown
    shutdownComplete  chan struct{}      // Signal shutdown is complete

    // Configuration
    GracefulTimeout   time.Duration      // Max wait time (default 30s)
    ShutdownCallback  func() error       // Custom cleanup callback

    // Logging
    Logger            *log.Logger
}
```

### 2. Shutdown Flow

```
┌─────────────────┐
│  SIGTERM/SIGINT │
└────────┬────────┘
         │
         ▼
┌──────────────────────────────┐
│ Signal Handler Triggered     │
│ 1. Close listen socket       │
│ 2. Stop accepting new requests
└────────┬─────────────────────┘
         │
         ▼
┌──────────────────────────────┐
│ Cancel Active Streams        │
│ 1. Send cancellation context │
│ 2. Wait for completion       │
└────────┬─────────────────────┘
         │
         ▼
┌──────────────────────────────┐
│ Wait for Request Completion  │
│ 1. Track active requests     │
│ 2. Wait up to 30s            │
│ 3. Force close if timeout    │
└────────┬─────────────────────┘
         │
         ▼
┌──────────────────────────────┐
│ Resource Cleanup             │
│ 1. Close connections         │
│ 2. Cleanup caches            │
│ 3. Close channels            │
└────────┬─────────────────────┘
         │
         ▼
┌──────────────────────────────┐
│ Server Stopped               │
│ Exit cleanly                 │
└──────────────────────────────┘
```

### 3. Component Details

#### A. Signal Handling
```go
func (gsm *GracefulShutdownManager) setupSignalHandlers() {
    // Handle SIGTERM and SIGINT
    // Start shutdown sequence
    // Log graceful shutdown message
}
```

#### B. Request Tracking
```go
func (gsm *GracefulShutdownManager) trackRequest() {
    // Increment activeRequests
}

func (gsm *GracefulShutdownManager) completeRequest() {
    // Decrement activeRequests
    // Check if shutdown can proceed
}
```

#### C. Stream Management
```go
func (gsm *GracefulShutdownManager) trackStream(id string, cancel context.CancelFunc) {
    // Add stream to activeStreams
}

func (gsm *GracefulShutdownManager) cancelAllStreams() {
    // Cancel all active streams
    // Wait for completion
}
```

#### D. Shutdown Orchestration
```go
func (gsm *GracefulShutdownManager) Shutdown() error {
    // 1. Signal all new requests to fail
    // 2. Cancel all active streams
    // 3. Wait for request completion (with timeout)
    // 4. Cleanup resources
    // 5. Return any errors
}
```

---

## 🔧 Implementation Steps

### Step 1: Create shutdown.go (150+ lines)
```go
package crewai

import (
    "context"
    "log"
    "net/http"
    "os"
    "os/signal"
    "sync"
    "sync/atomic"
    "syscall"
    "time"
)

// GracefulShutdownManager manages server lifecycle
type GracefulShutdownManager struct {
    // ... structure definition
}

// NewGracefulShutdownManager creates a new manager
func NewGracefulShutdownManager() *GracefulShutdownManager {
    // ... initialization
}

// Start begins monitoring for shutdown signals
func (gsm *GracefulShutdownManager) Start() {
    // ... signal handling
}

// Shutdown gracefully stops the server
func (gsm *GracefulShutdownManager) Shutdown(ctx context.Context) error {
    // ... shutdown logic
}

// RegisterStream tracks an active stream
func (gsm *GracefulShutdownManager) RegisterStream(id string, cancel context.CancelFunc) {
    // ... stream tracking
}

// UnregisterStream removes completed stream
func (gsm *GracefulShutdownManager) UnregisterStream(id string) {
    // ... cleanup
}

// IncrementActiveRequests tracks request start
func (gsm *GracefulShutdownManager) IncrementActiveRequests() {
    // ... request tracking
}

// DecrementActiveRequests tracks request completion
func (gsm *GracefulShutdownManager) DecrementActiveRequests() {
    // ... request tracking
}
```

### Step 2: Modify http.go (50+ lines)

Update `StartHTTPServer()` and `StartHTTPServerWithCustomUI()`:

```go
func StartHTTPServer(executor *CrewExecutor, port int) error {
    // Create graceful shutdown manager
    gsm := NewGracefulShutdownManager()
    gsm.GracefulTimeout = 30 * time.Second

    // Setup server with proper configuration
    mux := http.NewServeMux()
    handler := NewHTTPHandler(executor)

    // Register handlers
    mux.HandleFunc("/api/crew/stream", func(w http.ResponseWriter, r *http.Request) {
        gsm.IncrementActiveRequests()
        defer gsm.DecrementActiveRequests()
        handler.StreamHandler(w, r)
    })

    // ... other handlers

    server := &http.Server{
        Addr:    fmt.Sprintf(":%d", port),
        Handler: mux,
        // Timeouts
        ReadTimeout:  15 * time.Second,
        WriteTimeout: 15 * time.Second,
        IdleTimeout:  60 * time.Second,
    }

    gsm.server = server

    // Start signal handler (goroutine)
    go gsm.Start()

    // Start server
    return server.ListenAndServe()
}
```

### Step 3: Modify crew.go (30+ lines)

Add stream context tracking in `ExecuteStream()`:

```go
// In ExecuteStream, register stream with shutdown manager
if executor.ShutdownManager != nil {
    streamID := uuid.New().String()
    streamCtx, cancel := context.WithCancel(ctx)
    executor.ShutdownManager.RegisterStream(streamID, cancel)
    defer executor.ShutdownManager.UnregisterStream(streamID)

    ctx = streamCtx
}
```

### Step 4: Add fields to CrewExecutor (20+ lines)

```go
type CrewExecutor struct {
    // ... existing fields
    ShutdownManager *GracefulShutdownManager  // For graceful shutdown (Issue #18)
}
```

### Step 5: Create comprehensive tests (200+ lines)

```go
// Tests in shutdown_test.go

func TestGracefulShutdownCreation(t *testing.T)
func TestRequestTracking(t *testing.T)
func TestStreamRegistration(t *testing.T)
func TestShutdownTimeout(t *testing.T)
func TestSignalHandling(t *testing.T)
func TestCleanResourcesOnShutdown(t *testing.T)
func TestZeroDowntimeScenario(t *testing.T)
```

---

## 📋 Detailed Implementation Checklist

### Phase 1: Core Shutdown Manager
- [ ] Create `shutdown.go` file
- [ ] Define `GracefulShutdownManager` struct
- [ ] Implement initialization (`NewGracefulShutdownManager`)
- [ ] Add request counting (atomic operations)
- [ ] Add stream tracking (map + RWMutex)

### Phase 2: Signal Handling
- [ ] Setup signal handlers (SIGTERM, SIGINT)
- [ ] Log shutdown initiation
- [ ] Close listen socket
- [ ] Prevent new requests

### Phase 3: Stream Management
- [ ] Implement `RegisterStream()`
- [ ] Implement `UnregisterStream()`
- [ ] Implement `CancelAllStreams()`
- [ ] Track stream completion

### Phase 4: Request Completion
- [ ] Implement `IncrementActiveRequests()`
- [ ] Implement `DecrementActiveRequests()`
- [ ] Wait for active request completion
- [ ] Timeout protection (30s default)

### Phase 5: Resource Cleanup
- [ ] Close HTTP connections
- [ ] Cleanup goroutines
- [ ] Close metric collectors
- [ ] Clear caches

### Phase 6: HTTP Integration
- [ ] Modify `StartHTTPServer()`
- [ ] Modify `StartHTTPServerWithCustomUI()`
- [ ] Add request/stream wrapping
- [ ] Integrate with `CrewExecutor`

### Phase 7: Testing & Validation
- [ ] Create 7+ test cases
- [ ] Test signal handling
- [ ] Test request completion
- [ ] Test timeout mechanism
- [ ] Test resource cleanup
- [ ] Test concurrent scenarios

---

## ✅ Acceptance Criteria

### Functional Requirements
- ✅ SIGTERM/SIGINT handling
- ✅ Active streams complete within timeout
- ✅ No resource leaks (goroutines, connections)
- ✅ Proper logging of shutdown events
- ✅ Zero data loss during shutdown

### Performance Requirements
- ✅ < 100ms shutdown initiation
- ✅ 30s timeout for request completion
- ✅ Minimal memory overhead
- ✅ No blocking of concurrent requests

### Code Quality
- ✅ Thread-safe operations (atomic + RWMutex)
- ✅ Proper error handling
- ✅ Comprehensive logging
- ✅ 100% test coverage for new code
- ✅ Zero breaking changes

---

## 📊 Metrics & Logging

### Log Messages

```
[SHUTDOWN] Shutdown signal received (SIGTERM)
[SHUTDOWN] Stopping accepting new requests
[SHUTDOWN] 5 active requests still processing...
[SHUTDOWN] 3 active streams still running...
[SHUTDOWN] Waiting for completion (timeout: 30s)
[SHUTDOWN] All requests completed, cleaning up resources
[SHUTDOWN] Server shutdown complete
```

### Metrics to Track

- Shutdown initiation time
- Time to close listen socket
- Active request completion time
- Active stream completion time
- Total cleanup time

---

## 🧪 Testing Strategy

### Unit Tests
1. **Initialization**: Verify proper setup
2. **Request Tracking**: Increment/decrement operations
3. **Stream Management**: Register/unregister operations
4. **Timeout**: Verify timeout mechanism works
5. **Signal Handling**: Verify signals are caught
6. **Resource Cleanup**: Verify all resources cleaned

### Integration Tests
1. Shutdown with active requests
2. Shutdown with active streams
3. Shutdown with concurrent requests
4. Timeout scenario (force close)
5. Multiple signal handling

### Scenario Tests
```go
// Test: Server with 10 active requests, receives SIGTERM
// Expected: Wait for all to complete, shutdown clean

// Test: Server with long-running stream (> 30s), receives SIGTERM
// Expected: Cancel stream, timeout after 30s, force close

// Test: Multiple clients connecting during shutdown
// Expected: Reject new connections, complete existing
```

---

## 🚀 Deployment Considerations

### Zero-Downtime Update Strategy
```bash
# 1. Load balancer: Stop sending new requests to old server
# 2. Old server: Receive graceful shutdown signal
# 3. Old server: Wait for active requests to complete (max 30s)
# 4. New server: Start on same port
# 5. Load balancer: Route new requests to new server
# 6. Old server: Exit cleanly
```

### Kubernetes Integration
```yaml
# terminationGracePeriodSeconds = 40s (must be > 30s timeout)
spec:
  terminationGracePeriodSeconds: 40
  containers:
    - name: crew-api
      lifecycle:
        preStop:
          exec:
            command: ["sh", "-c", "sleep 2"]
```

---

## 📝 Files to Create/Modify

| File | Type | Lines | Changes |
|------|------|-------|---------|
| shutdown.go | NEW | 200+ | Core shutdown implementation |
| http.go | MODIFIED | +50 | Integration with StartHTTPServer |
| crew.go | MODIFIED | +30 | Stream context tracking |
| shutdown_test.go | NEW | 200+ | Comprehensive tests |

---

## 🎯 Success Criteria

### Code Quality
- ✅ All tests passing
- ✅ Race detector clean
- ✅ No deadlocks
- ✅ Proper resource cleanup

### Functional
- ✅ Graceful shutdown on SIGTERM/SIGINT
- ✅ Request completion within timeout
- ✅ Stream cancellation on shutdown
- ✅ Proper logging
- ✅ Clean exit code

### Documentation
- ✅ Design document (this file)
- ✅ Code comments
- ✅ Integration examples
- ✅ Test documentation

---

## 📈 Impact Assessment

### Positive Impact
- ✅ Zero downtime deployments
- ✅ Data loss prevention
- ✅ Resource cleanup
- ✅ Production stability

### Risk Mitigation
- Timeout protection prevents infinite hangs
- Atomic operations prevent race conditions
- Proper logging aids debugging
- Comprehensive tests ensure reliability

---

## 🔗 Related Issues

- Issue #1: Race Condition (related: synchronization)
- Issue #5: Panic Risk (related: error handling)
- Issue #11: Timeouts (related: timeout mechanism)
- Issue #14: Metrics (related: shutdown metrics)

---

## 📅 Timeline

### Day 1 (4-5 hours)
- Create `shutdown.go` with core implementation
- Add signal handling
- Implement request/stream tracking

### Day 2 (3-4 hours)
- Integrate with HTTP server
- Write 7+ test cases
- Performance testing
- Documentation finalization
- Final commit

**Total**: ~8 hours (1-2 days)

---

**Status**: DESIGN PHASE COMPLETE ✅
**Next**: Implementation can begin

---

*Design Date: 2025-12-22*
*Phase 3 Issue #18*

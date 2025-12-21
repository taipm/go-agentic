# 🔗 Issue #17: Request ID Tracking - Design Document

**Date**: 2025-12-22
**Status**: DESIGN PHASE
**Priority**: MEDIUM (Score: 60/100)
**Effort**: MEDIUM (1.5-2 days)

---

## 🎯 Objective

Implement distributed request tracking system that:
- Assigns unique ID to each request at entry point
- Propagates ID through all components (agents, tools, streams)
- Enables request correlation across components
- Supports distributed tracing and debugging
- Improves operational observability
- Simplifies request lifecycle tracking

---

## 📋 Current State Analysis

### What Exists

```go
// Execution uses context but no request IDs
func (ce *CrewExecutor) ExecuteStream(ctx context.Context, ...) error {
    // No unique request identifier
    // Can't trace request across logs
}
```

### Current Gaps

```
Missing Tracking:
❌ No unique request ID assigned
❌ No ID propagation through context
❌ Can't correlate logs for single request
❌ No request lifecycle tracking
❌ No distributed tracing support
❌ Can't group tool calls by request
❌ No request duration tracking
```

### Problem Example

```
Logs from multi-agent system:
[2025-12-22 10:30:00] Agent orchestrator thinking...
[2025-12-22 10:30:01] Agent clarifier executing...
[2025-12-22 10:30:02] Tool GetCPUUsage called
[2025-12-22 10:30:03] Agent executor thinking...

❌ Can't tell which logs are from same request!
❌ Can't trace request lifecycle!
❌ Hard to debug multi-request scenarios!
```

### Solution

```
With Request ID Tracking:
[req-abc123] Agent orchestrator thinking...
[req-abc123] Agent clarifier executing...
[req-abc123] Tool GetCPUUsage called
[req-abc123] Agent executor thinking...

✅ All logs easily grouped by request!
✅ Can trace complete request lifecycle!
✅ Easy to debug multi-request scenarios!
```

---

## 🏗️ Implementation Design

### 1. Request ID Format

```go
// UUID format for uniqueness
// Example: "550e8400-e29b-41d4-a716-446655440000"

func GenerateRequestID() string {
    id := uuid.New()  // From google/uuid
    return id.String()
}

// Alternative: Shorter format for logs
// "req-abc123def456" (18 chars)
func GenerateShortRequestID() string {
    id := uuid.New()
    return "req-" + id.String()[:12]
}
```

### 2. Context Propagation

```go
// Add request ID to context
const RequestIDKey = "request-id"

// Set request ID
ctx = context.WithValue(ctx, RequestIDKey, requestID)

// Get request ID
func GetRequestID(ctx context.Context) string {
    id, ok := ctx.Value(RequestIDKey).(string)
    if !ok {
        return "unknown"
    }
    return id
}

// Verify ID is in context at each layer
func (ce *CrewExecutor) ExecuteStream(ctx context.Context, ...) error {
    requestID := GetRequestID(ctx)
    log.Printf("[%s] Request started", requestID)
    // ...
}
```

### 3. Request Tracking Structure

```go
// Track request metadata
type RequestMetadata struct {
    ID              string        // Unique ID
    UserInput       string        // Original user input
    Model           string        // LLM model used
    StartTime       time.Time     // When request started
    EndTime         time.Time     // When request ended
    Duration        time.Duration // Total duration

    // Execution tracking
    AgentCalls      int           // How many agents called
    ToolCalls       int           // How many tools called
    RoundCount      int           // Execution rounds

    // Status tracking
    Status          string        // success, error, timeout
    ErrorMessage    string        // If failed

    // Stream events
    Events          []Event       // All events in order
}

type Event struct {
    Type      string        // agent_thinking, tool_call, etc.
    Agent     string        // Which agent
    Tool      string        // Which tool (if applicable)
    Timestamp time.Time
    Data      interface{}   // Event-specific data
}
```

### 4. Request Store (In-Memory)

```go
// Store recent requests for debugging
type RequestStore struct {
    mu       sync.RWMutex
    requests map[string]*RequestMetadata
    maxSize  int           // Keep last N requests
}

func NewRequestStore(maxSize int) *RequestStore {
    return &RequestStore{
        requests: make(map[string]*RequestMetadata),
        maxSize:  maxSize,
    }
}

func (rs *RequestStore) Add(meta *RequestMetadata) {
    rs.mu.Lock()
    defer rs.mu.Unlock()

    rs.requests[meta.ID] = meta

    // Keep only recent requests
    if len(rs.requests) > rs.maxSize {
        // Delete oldest
    }
}

func (rs *RequestStore) Get(id string) *RequestMetadata {
    rs.mu.RLock()
    defer rs.mu.RUnlock()
    return rs.requests[id]
}

func (rs *RequestStore) GetAll() map[string]*RequestMetadata {
    rs.mu.RLock()
    defer rs.mu.RUnlock()
    return rs.requests
}
```

### 5. Integration Points

```go
// 1. HTTP Handler: Generate ID
func (h *HTTPHandler) StreamHandler(w http.ResponseWriter, r *http.Request) {
    requestID := GenerateRequestID()
    ctx := context.WithValue(r.Context(), RequestIDKey, requestID)

    meta := &RequestMetadata{
        ID:        requestID,
        UserInput: input,
        Model:     model,
        StartTime: time.Now(),
    }

    // ... pass ctx to executor
}

// 2. CrewExecutor: Propagate ID
func (ce *CrewExecutor) ExecuteStream(ctx context.Context, ...) error {
    requestID := GetRequestID(ctx)
    log.Printf("[%s] Starting execution", requestID)

    // Pass ctx (with ID) to all downstream calls
    err := ce.executeAgent(ctx, agent)
    // ...
}

// 3. Agent Execution: Include ID in logs
func (ce *CrewExecutor) executeAgent(ctx context.Context, agent *Agent) error {
    requestID := GetRequestID(ctx)
    log.Printf("[%s] Executing agent: %s", requestID, agent.Name)

    // ...
}

// 4. Tool Execution: Include ID
func (ce *CrewExecutor) executeTool(ctx context.Context, tool *Tool, ...) (string, error) {
    requestID := GetRequestID(ctx)
    log.Printf("[%s] Executing tool: %s", requestID, tool.Name)

    // ...
}

// 5. SSE Streaming: Include ID in events
type StreamEvent struct {
    RequestID string      `json:"request_id"`
    Type      string      `json:"type"`
    Data      interface{} `json:"data"`
}
```

### 6. Logging Integration

```go
// Structured logging with request ID
func LogWithRequest(ctx context.Context, level string, message string, args ...interface{}) {
    requestID := GetRequestID(ctx)
    timestamp := time.Now().Format("2006-01-02 15:04:05")

    formattedMsg := fmt.Sprintf(message, args...)
    logLine := fmt.Sprintf("[%s] [%s] [%s] %s",
        timestamp, level, requestID, formattedMsg)

    fmt.Println(logLine)
}

// Usage
LogWithRequest(ctx, "INFO", "Agent %s executing", agent.Name)
// Output: [2025-12-22 10:30:00] [INFO] [req-abc123] Agent orchestrator executing
```

### 7. API Endpoint for Request History

```go
// GET /api/requests/:id
func (h *HTTPHandler) GetRequest(w http.ResponseWriter, r *http.Request) {
    id := mux.Vars(r)["id"]
    meta := h.executor.RequestStore.Get(id)

    if meta == nil {
        http.Error(w, "Request not found", 404)
        return
    }

    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(meta)
}

// Example response:
// {
//   "id": "req-abc123",
//   "user_input": "Check my system",
//   "model": "gpt-4o",
//   "start_time": "2025-12-22T10:30:00Z",
//   "end_time": "2025-12-22T10:30:05Z",
//   "duration": "5s",
//   "status": "success",
//   "agent_calls": 3,
//   "tool_calls": 5,
//   "round_count": 3,
//   "events": [...]
// }

// GET /api/requests?limit=10&status=error
func (h *HTTPHandler) ListRequests(w http.ResponseWriter, r *http.Request) {
    limit := r.URL.Query().Get("limit")
    status := r.URL.Query().Get("status")

    all := h.executor.RequestStore.GetAll()

    // Filter and limit
    // Return JSON array
}
```

---

## 📝 Implementation Steps

### Step 1: Create request_tracking.go (200+ lines)

```go
package crewai

import (
    "context"
    "github.com/google/uuid"
)

const RequestIDKey = "request-id"

type RequestMetadata struct {
    ID            string
    UserInput     string
    Model         string
    StartTime     time.Time
    EndTime       time.Time
    Duration      time.Duration
    AgentCalls    int
    ToolCalls     int
    RoundCount    int
    Status        string
    ErrorMessage  string
    Events        []Event
}

type Event struct {
    Type      string
    Agent     string
    Tool      string
    Timestamp time.Time
    Data      interface{}
}

func GenerateRequestID() string {
    return uuid.New().String()
}

func SetRequestID(ctx context.Context, id string) context.Context {
    return context.WithValue(ctx, RequestIDKey, id)
}

func GetRequestID(ctx context.Context) string {
    id, ok := ctx.Value(RequestIDKey).(string)
    if !ok {
        return "unknown"
    }
    return id
}

type RequestStore struct {
    mu       sync.RWMutex
    requests map[string]*RequestMetadata
    maxSize  int
}

func NewRequestStore(maxSize int) *RequestStore {
    return &RequestStore{
        requests: make(map[string]*RequestMetadata),
        maxSize:  maxSize,
    }
}
```

### Step 2: Modify http.go (30+ lines)

Add request ID generation and store in CrewExecutor:

```go
// In HTTPHandler.StreamHandler:
requestID := GenerateRequestID()
ctx = SetRequestID(r.Context(), requestID)

meta := &RequestMetadata{
    ID:        requestID,
    UserInput: input,
    Model:     model,
    StartTime: time.Now(),
}

h.executor.RequestStore.Add(meta)
```

### Step 3: Modify crew.go (50+ lines)

Propagate request ID through execution and update metadata:

```go
// In ExecuteStream:
requestID := GetRequestID(ctx)
meta := ce.RequestStore.Get(requestID)

// Update counts
meta.AgentCalls++
meta.ToolCalls++
meta.RoundCount++

// Log with request ID
log.Printf("[%s] Executing agent: %s", requestID, agent.Name)
```

### Step 4: Add request history API (40+ lines)

```go
// In http.go:
mux.HandleFunc("/api/requests/{id}", h.GetRequest)
mux.HandleFunc("/api/requests", h.ListRequests)

func (h *HTTPHandler) GetRequest(w http.ResponseWriter, r *http.Request) {
    // Implementation
}
```

### Step 5: Add tests (150+ lines)

```go
func TestRequestIDGeneration(t *testing.T)
func TestRequestIDPropagation(t *testing.T)
func TestRequestStoreAdd(t *testing.T)
func TestRequestStoreGet(t *testing.T)
func TestRequestMetadataUpdate(t *testing.T)
func TestEventTracking(t *testing.T)
func TestRequestHistoryAPI(t *testing.T)
```

---

## ✅ Acceptance Criteria

### Functional Requirements
- ✅ Unique request ID generated for each request
- ✅ ID propagated through all components
- ✅ Request metadata stored and retrievable
- ✅ Events logged with request ID
- ✅ Request history API endpoint works
- ✅ Can filter requests by status
- ✅ Request duration tracked
- ✅ Agent/tool call counts tracked

### Quality Requirements
- ✅ All logs include request ID
- ✅ No context leakage between requests
- ✅ Memory efficient (circular buffer for history)
- ✅ Test coverage > 90%
- ✅ Zero breaking changes

---

## 📊 Success Metrics

- ✅ All logs include request ID (100%)
- ✅ Request tracing works for multi-component execution
- ✅ Request history retrievable via API
- ✅ Can correlate logs by request ID
- ✅ Debugging time reduced by 30%+

---

## 🔗 Integration with Issue #14 (Metrics)

Request ID tracking pairs well with metrics:

```go
// Metrics with request ID correlation
type MetricEvent struct {
    RequestID  string        // Links to request
    AgentID    string
    ToolName   string
    Duration   time.Duration
    Success    bool
}

// Can analyze metrics by request
metrics.GetRequest(requestID).AgentCalls
metrics.GetRequest(requestID).ToolCalls
metrics.GetRequest(requestID).Duration
```

---

## 🎯 Implementation Checklist

- [ ] Create request_tracking.go (200+ lines)
- [ ] Implement RequestID generation
- [ ] Implement RequestStore
- [ ] Integrate into HTTP handler
- [ ] Propagate ID through crew execution
- [ ] Add ID to all logs
- [ ] Add request history API
- [ ] Write 7+ tests
- [ ] Integration tests with metrics
- [ ] Documentation

---

**Status**: DESIGN COMPLETE
**Next**: Implementation of request_tracking.go

---

*Design Date: 2025-12-22*
*Phase 3 Issue #17*

# 🏗️ Architecture Overview - go-agentic

**Status**: Production Ready
**Version**: 1.0
**Last Updated**: 2025-12-22

---

## 🎯 What is go-agentic?

go-agentic is a **production-grade multi-agent orchestration framework** that enables you to build intelligent autonomous systems where multiple specialized AI agents work together to solve complex problems.

Unlike single-agent systems, go-agentic provides:
- **Agent Collaboration**: Agents communicate and intelligently hand off work
- **Intelligent Routing**: Problems are routed to the right agent based on analysis
- **Real-time Streaming**: Watch agents work in real-time via Server-Sent Events (SSE)
- **Complete Feedback Loops**: Multi-round execution where agents see tool results
- **Production Ready**: Thread-safe, error-handled, comprehensive monitoring

---

## 🔧 Key Components

```
┌────────────────────────────────────────────────────────────┐
│                   Client Application                       │
│              (CLI, Web UI, or Custom Code)                │
└────────────────────────┬─────────────────────────────────┘
                         │
                         │ HTTP(S)
                         ▼
┌────────────────────────────────────────────────────────────┐
│           HTTP Server (Port 8081 default)                 │
│  • Request routing and validation                          │
│  • SSE streaming setup                                     │
│  • Metrics collection middleware                           │
│  • Health check endpoint                                   │
└────────────────────────┬─────────────────────────────────┘
                         │
                         ▼
┌────────────────────────────────────────────────────────────┐
│              CrewExecutor (Main Orchestrator)              │
│  • Request parsing and validation                          │
│  • Agent lifecycle management                              │
│  • Stream management for SSE                               │
│  • Error handling and recovery                             │
│  • Graceful shutdown coordination                          │
└────────────────────────┬─────────────────────────────────┘
                         │
        ┌────────────────┼────────────────┐
        ▼                ▼                ▼
    ┌────────┐      ┌────────┐      ┌────────┐
    │ Agent  │      │ Agent  │      │ Agent  │
    │   #1   │      │   #2   │      │   #3   │
    │  (LLM) │      │  (LLM) │      │  (LLM) │
    └────────┘      └────────┘      └────────┘
        │                │                │
        └────────────────┼────────────────┘
                         │
                         ▼
        ┌────────────────────────────────┐
        │   Tool Execution Engine        │
        │  • Tool lookup and validation  │
        │  • Parameter handling          │
        │  • Timeout enforcement         │
        │  • Error recovery              │
        │  • Panic prevention (defer)    │
        └────────────────────────────────┘
                         │
        ┌────────────────┴─────────────────┐
        ▼                                   ▼
    ┌──────────┐                     ┌──────────┐
    │  Custom  │                     │  Custom  │
    │  Tools   │                     │  Tools   │
    └──────────┘                     └──────────┘
```

### 1. HTTP Server (`http.go`)

**Responsibility**: Handle HTTP requests and coordinate responses

**Key Functions**:
- `StartHTTPServer()` - Starts server on specified port
- `StartHTTPServerWithCustomUI()` - Customizable UI server
- Request validation and parameter extraction
- SSE (Server-Sent Events) streaming setup
- Metrics collection at request level
- Graceful shutdown integration

**Example**: `/api/crew/stream` endpoint handles user queries

---

### 2. CrewExecutor (`crew.go`)

**Responsibility**: Orchestrate agents, manage execution flow, track state

**Key Responsibilities**:
- Load and manage agents from configuration
- Execute agents in proper sequence
- Handle agent routing (who talks to whom)
- Manage tool execution results
- Stream responses in real-time
- Track metrics for observability
- Handle errors gracefully
- Manage graceful shutdown

**Key Structures**:
```go
type CrewExecutor struct {
    crew              *Crew                          // Agent configuration
    client            *openai.Client                 // LLM client
    tools             map[string]*Tool               // Available tools
    metrics           *MetricsCollector              // Performance metrics
    shutdownManager   *GracefulShutdownManager       // Shutdown coordination
    maxConcurrentReq  int                            // Concurrency limit
}
```

**Execution Flow**:
1. Request received with user input
2. Find entry point agent (usually Orchestrator)
3. Send user input to agent LLM
4. Parse LLM response for tool calls
5. Execute requested tools
6. Send results back to agent
7. Repeat until agent signals completion
8. Return final response to client

---

### 3. Agent System (`agent.go`)

**Responsibility**: Represent and manage individual AI agents

**Key Components**:
- **Agent Definition**: Name, role, backstory, model, tools
- **Agent Execution**: Call LLM with context and tool availability
- **Tool Parsing**: Extract tool calls from LLM responses
- **Terminal Agents**: Guarantee final agent in workflow

**Agent Structure**:
```go
type Agent struct {
    ID              string             // Unique identifier
    Name            string             // Display name
    Role            string             // Agent's role/purpose
    Backstory       string             // Background and expertise
    Model           string             // LLM model to use
    Tools           []*Tool            // Available tools
    Temperature     float32            // LLM creativity (0-1)
    IsTerminal      bool               // Is last agent?
}
```

**Agent Execution Pattern**:
```
Input: User message

1. Build system prompt from role/backstory
2. Add tool availability information
3. Call OpenAI gpt-4o model
4. Parse response for tool calls [TOOL_CALL: ...]
5. Execute tools
6. Send results back to agent
7. Repeat until:
   - Agent signals completion (e.g., [DONE])
   - Max rounds reached
   - Timeout exceeded
```

---

### 4. Tool System (`types.go`, `crew.go`)

**Responsibility**: Manage executable functions that agents can call

**Tool Structure**:
```go
type Tool struct {
    Name        string                 // Tool name
    Description string                 // What it does
    Parameters  map[string]interface{} // JSON Schema
    Handler     func(...) (string, error) // Implementation
}
```

**Tool Execution Flow**:
```
1. Agent requests tool execution
2. Validate tool exists
3. Extract and validate parameters
4. Set execution timeout context
5. Call tool handler function
6. Catch panics with defer-recover
7. Return result or error
8. Stream result to client
```

**Safety Features**:
- ✅ Timeout protection (default 5s per tool)
- ✅ Panic recovery (no crash on bad tool)
- ✅ Parameter validation
- ✅ Error logging and reporting

---

### 5. Streaming System (`streaming.go`, `http.go`)

**Responsibility**: Stream agent execution in real-time via SSE

**How It Works**:

```
Client                    Server
   │                        │
   ├─ POST /api/crew/stream │
   │                        │
   │                        ├─ Start SSE stream
   │ event: start           │
   │ data: {...}    ◄───────┤
   │                        │
   │                        ├─ Call agent
   │ event: agent_thinking  │
   │ data: {...}    ◄───────┤
   │                        │
   │ event: tool_call       │
   │ data: {...}    ◄───────┤
   │                        ├─ Execute tool
   │ event: tool_result     │
   │ data: {...}    ◄───────┤
   │                        │
   │ event: agent_response  │
   │ data: {...}    ◄───────┤
   │                        │
   │ event: complete        │
   │ data: {...}    ◄───────┤
   │                        │
```

**Event Types**:
- `start` - Execution started
- `agent_thinking` - Agent processing
- `tool_call` - Tool about to execute
- `tool_result` - Tool result available
- `agent_response` - Agent final response
- `error` - Execution error
- `complete` - Execution finished

---

### 6. Metrics System (`metrics.go`)

**Responsibility**: Track performance and operational metrics

**4-Layer Metrics Architecture**:

1. **Tool Level**: Per-tool execution metrics
   - Execution count
   - Success/error rates
   - Duration stats (min/max/avg)

2. **Agent Level**: Per-agent metrics
   - Total executions
   - Success rates
   - Tool usage breakdown

3. **System Level**: Aggregate metrics
   - Total requests processed
   - Success/failure rates
   - Memory usage
   - Cache hit rates

4. **Export Formats**: JSON and Prometheus
   - JSON: Complete metrics dump
   - Prometheus: Compatible with monitoring stacks

**Example Metrics**:
```json
{
  "system_metrics": {
    "total_requests": 150,
    "successful_requests": 145,
    "failed_requests": 5,
    "average_request_time": "1.2s",
    "memory_usage": 52428800,
    "cache_hit_rate": 0.85
  }
}
```

---

### 7. Graceful Shutdown System (`shutdown.go`)

**Responsibility**: Safe server shutdown with request completion

**Shutdown Flow**:

```
User presses Ctrl+C (SIGINT)
        │
        ▼
Signal handler triggered
        │
        ├─ Mark as shutting down
        ├─ Stop accepting new requests
        │
        ▼
Cancel active streams
        │
        ▼
Wait for active requests
   (max 30 seconds)
        │
        ├─ If timeout → Force close
        │
        ▼
Run cleanup callback
        │
        ▼
Shutdown HTTP server
        │
        ▼
Exit cleanly (code 0)
```

**Key Features**:
- Atomic request counting (lock-free)
- Stream cancellation via context
- Configurable timeout (default 30s)
- Custom cleanup callback support
- Zero data loss during shutdown

---

## 📊 Data Flow Example

### Scenario: User asks "Check my system health"

```
Step 1: REQUEST ARRIVES
   Client sends:
   {
     "user_input": "Check my system health",
     "model": "gpt-4o"
   }

Step 2: ORCHESTRATOR AGENT
   • Receives user input
   • LLM analyzes: This is a system diagnostic request
   • Routes to Executor agent

Step 3: EXECUTOR AGENT
   • Receives: "Check system health"
   • Decides: Need CPU and memory info
   • Calls tools:
     - GetCPUUsage()
     - GetMemoryUsage()

Step 4: TOOL EXECUTION
   • GetCPUUsage() → Returns "85%"
   • GetMemoryUsage() → Returns "72%"
   • Send results back to agent

Step 5: AGENT ANALYSIS
   • Receives tool results
   • LLM generates response:
     "Your system health is concerning:
      - CPU: 85% (high)
      - Memory: 72% (moderate)
      Recommendation: Free up memory"

Step 6: RESPONSE STREAMING
   • Stream response to client
   • Send metrics
   • Close SSE stream

Step 7: CLIENT RECEIVES
   • Display agent response
   • Show recommendations
   • Close connection
```

---

## 🔐 Error Handling Strategy

### Panic Prevention

**Problem**: Tool execution might panic (index out of bounds, nil pointer, etc.)

**Solution**: Defer-recover pattern
```go
defer func() {
    if r := recover(); r != nil {
        log.Printf("Tool panic: %v", r)
        // Continue execution, don't crash
    }
}()
// Tool code here
```

### Timeout Protection

**Problem**: Tool might hang indefinitely

**Solution**: Context with timeout
```go
ctx, cancel := context.WithTimeout(parentCtx, 5*time.Second)
defer cancel()
result, err := tool.Handler(ctx, params)
// Automatically cancelled after 5s
```

### Request Completion

**Problem**: Server shutdown might interrupt requests

**Solution**: Graceful shutdown with request tracking
```go
gsm.IncrementActiveRequests()  // Request starts
defer gsm.DecrementActiveRequests()  // Request ends
// On shutdown: Wait for all requests to complete
```

---

## 🧵 Concurrency Model

### Thread Safety Guarantees

1. **Atomic Operations**: Lock-free request counting
   ```go
   atomic.AddInt32(&activeRequests, 1)
   ```

2. **RWMutex**: Concurrent read, exclusive write
   ```go
   mu.RLock()   // Many goroutines can read
   defer mu.RUnlock()
   ```

3. **Channels**: Safe goroutine communication
   ```go
   signals := make(chan os.Signal, 1)
   signal.Notify(signals, syscall.SIGTERM)
   ```

4. **Context**: Cancellation propagation
   ```go
   ctx, cancel := context.WithCancel(parentCtx)
   // Cancel affects all child goroutines
   ```

---

## 📈 Production Characteristics

### Performance

| Metric | Value | Notes |
|--------|-------|-------|
| Request latency | 0.5-3s | Depends on agent complexity |
| Tool execution | <5s | Default timeout |
| Stream startup | <100ms | SSE handshake |
| Memory per request | ~2MB | Typical usage |
| Concurrent requests | 100+ | System dependent |

### Reliability

| Aspect | Status | Details |
|--------|--------|---------|
| Error recovery | ✅ | Graceful degradation |
| Panic prevention | ✅ | Defer-recover on all tools |
| Data loss | ✅ | Prevented with graceful shutdown |
| Resource leaks | ✅ | Goroutines cleaned up |
| Thread safety | ✅ | Atomic + RWMutex protection |

### Observability

| Component | Metrics | Export |
|-----------|---------|--------|
| Requests | Count, latency | JSON, Prometheus |
| Agents | Executions, success rate | JSON, Prometheus |
| Tools | Executions, duration | JSON, Prometheus |
| Memory | Current, peak usage | JSON, Prometheus |
| Cache | Hits, misses, hit rate | JSON, Prometheus |

---

## 🎓 Design Principles

### 1. Configuration Over Code

**Principle**: Business logic in YAML, not Go code

**Benefit**: Non-technical users can modify agent behaviors without recompiling

**Example**:
```yaml
agents:
  orchestrator:
    name: "Smart Router"
    role: "Analyze requests and route to experts"
    model: "gpt-4o"
```

### 2. Explicit Over Implicit

**Principle**: Routing decisions made explicitly via signals

**Benefit**: Deterministic, auditable, maintainable

**Example**:
```go
// Good: Signal defined in config
if strings.Contains(response, signal.Signal) {
    // Route to target agent
}

// Bad: Hardcoded agent ID (breaks if renamed)
if strings.Contains(response, "[ROUTE_EXECUTOR]") {
    // brittle!
}
```

### 3. Safety by Default

**Principle**: Fail gracefully, never crash the server

**Examples**:
- Panic recovery on tool execution
- Timeout protection on all tools
- Graceful shutdown on signals
- Request completion guarantee

### 4. Complete Feedback Loops

**Principle**: Agents see results, not just requests

**Benefit**: More intelligent, context-aware decisions

**Example**:
```
Round 1: Agent calls GetCPUUsage() → 85%
Round 2: Agent sees "CPU: 85%" and calls GetMemory() → 72%
Round 3: Agent sees both results and provides analysis
```

---

## 📦 Deployment Architectures

### Local Development

```
Developer Machine
├─ go-agentic binary
├─ config/
│  ├─ crew.yaml
│  └─ agents/*.yaml
└─ HTTP: localhost:8081
```

### Docker Container

```
Docker Image
├─ go-agentic binary
├─ config/
├─ EXPOSE 8081
└─ Health check: GET /health
```

### Kubernetes Pod

```
Kubernetes Cluster
├─ Pod running go-agentic
├─ Service exposing port 8081
├─ Health checks (readiness/liveness)
├─ Graceful shutdown (terminationGracePeriodSeconds: 40)
└─ Resource limits
```

---

## 🔗 Component Relationships

```
┌─────────────────────┐
│   CrewExecutor      │ (Central orchestrator)
├─────────────────────┤
│ • agents[]          │ ──┐ References
│ • tools{}           │   ├─→ Agent instances
│ • metrics           │   ├─→ Tool instances
│ • shutdownManager   │   └─→ Metrics collector
└─────────────────────┘

┌─────────────────────┐
│   Agent             │
├─────────────────────┤
│ • tools[]           │ ──→ Tool references
│ • role, backstory   │
│ • model (gpt-4o)    │
└─────────────────────┘

┌─────────────────────┐
│   Tool              │
├─────────────────────┤
│ • name              │
│ • description       │
│ • parameters        │
│ • handler           │
└─────────────────────┘

┌─────────────────────┐
│  MetricsCollector   │
├─────────────────────┤
│ • System metrics    │
│ • Agent metrics     │
│ • Tool metrics      │
└─────────────────────┘

┌──────────────────────────────┐
│ GracefulShutdownManager      │
├──────────────────────────────┤
│ • activeRequests (atomic)    │
│ • activeStreams (map)        │
│ • signal handling (SIGTERM)  │
│ • shutdown coordination      │
└──────────────────────────────┘
```

---

## 🎯 Next Steps

To learn more about specific topics:

1. **Getting Started**: See [QUICK_START.md](QUICK_START.md)
2. **Configuration**: See [CONFIGURATION_GUIDE.md](CONFIGURATION_GUIDE.md)
3. **API Usage**: See [API_REFERENCE.md](API_REFERENCE.md)
4. **Operations**: See [DEPLOYMENT_GUIDE.md](DEPLOYMENT_GUIDE.md)
5. **Troubleshooting**: See [TROUBLESHOOTING_GUIDE.md](TROUBLESHOOTING_GUIDE.md)

---

**Version**: 1.0
**Last Updated**: 2025-12-22
**Status**: Production Ready ✅

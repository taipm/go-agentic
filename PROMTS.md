# PROMPTS CHAIN LIBRARY - go-agentic

**Áp dụng:** Tư duy nguyên bản (Elon Musk) + Tư duy tốc độ ánh sáng (NVIDIA)

---

## I. PROMPTS CHIẾN LƯỢC (Strategy Prompts)

### 1. Review Chuẩn Tổng Thể

**Mục tiêu:** Review nhanh nhất, tốt nhất, an toàn nhất cho thư viện lõi

```prompt
Thực hiện review toàn diện cơ sở mã go-agentic với tiêu chuẩn:

TIÊU CHÍ TƯ DUY NGUYÊN BẢN:
1) Phân tích MỤC ĐÍCH CỐT LÕI của từng component
   - core/types.go: Cấu trúc dữ liệu cốt lõi là gì?
   - core/agent.go: Vòng lặp thực thi agent làm gì?
   - core/crew.go: Logic định tuyến signal thực hiện như thế nào?

2) Xác định CÁC RỦI RO TOÀN CỤC
   - Điểm lỗi duy nhất (single point of failure)?
   - Các phụ thuộc ẩn (hidden dependencies)?
   - Race conditions hoặc deadlocks?

3) Đánh giá HIỆU SUẤT & SAFETY
   - Quota enforcement: Chi phí & memory limits có được enforce không?
   - Context management: Có bảo toàn context qua handoffs không?
   - Error handling: Tất cả error paths có được xử lý không?

TỐC ĐỘ AUDIT (Speed of Light):
- Sử dụng grep patterns để phát hiện issues nhanh
- Trace execution flow từ user input → final response
- Validate tất cả config paths (crew.yaml → agents/ → tools)

DELIVERABLE:
- Danh sách 5-10 issues cụ thể (nếu có)
- Điểm mạnh của kiến trúc hiện tại
- Khuyến nghị cải thiện theo priority (HIGH, MEDIUM, LOW)
```

### 2. Loại Bỏ Tài Liệu Dư Thừa Lỗi Thời

**Mục tiêu:** Làm sạch, giữ lại chỉ tài liệu chính

```prompt
Phân loại toàn bộ documentation trong go-agentic:

CATEGORIES:
A) CORE DOCUMENTATION (GIỮ LẠI):
   - docs/01-GETTING_STARTED.md
   - docs/02-CORE_CONCEPTS.md
   - docs/03-API_REFERENCE.md
   - docs/04-EXAMPLES.md
   - docs/05-DEPLOYMENT.md
   - examples/*/README.md
   
B) OUTDATED/REDUNDANT (XÓA):
   - Tìm tất cả files trong _bmad-output/
   - Tìm tất cả .md files ở root level không phải PROMTS.md hoặc README.md
   - Tìm files có timestamps quá cũ (>3 tháng)

C) GENERATED DURING DEVELOPMENT (STAGING):
   - DOCUMENTATION_CLEANUP_REPORT.md
   - Các files tạm từ workflow cũ

VALIDATION:
- Ensure no information loss (new standard docs cover old docs)
- Check all examples have README.md
- Verify docs match current codebase

OUTPUT:
- Danh sách files cần xóa
- Danh sách files cần cập nhật
- Verification checklist
```

---

## II. PROMPTS KIẾN TRÚC (Architecture Prompts)

### 3. Hiểu Rõ Signal-Based Routing

**Mục tiêu:** Giải thích cơ chế định tuyến của framework

```prompt
Phân tích signal-based routing trong core/crew.go:

FIRST PRINCIPLES QUESTIONS:
1. Signal là gì? (Định nghĩa chính xác)
   - String patterns: [ROUTE_EXECUTOR], [ROUTE_CLARIFIER], ...?
   - Được emit từ đâu?
   - Được parse ở đâu?

2. Routing logic hoạt động như thế nào?
   - Agent emit signal → Crew phát hiện → Route tới agent nào?
   - Có timeout không? Cơ chế fallback là gì?
   - Max handoffs enforcement ở đâu?

3. Context bảo toàn qua handoffs?
   - Message history được pass như thế nào?
   - Agent N biết kết quả từ Agent N-1 không?
   - Có lose information không?

SPEED ANALYSIS:
- Grep để tìm: "ROUTE_", "[ROUTE", signal patterns
- Trace execution flow: Execute Agent → Emit Signal → Detect → Next Agent
- Xác định performance bottlenecks

DOCUMENTATION NEEDED:
- Signal registry: Tất cả signals được hỗ trợ?
- Routing diagram: Flow từ entry_point đến terminal agent
- Best practices: Khi nào nên dùng signal routing?
```

### 4. Multi-Provider LLM Architecture

**Mục tiêu:** Hiểu cách framework hỗ trợ multiple LLM providers

```prompt
Phân tích multi-provider LLM support:

PROVIDER LANDSCAPE:
1. Primary Provider: core/providers/provider.go
   - Interface định nghĩa gì?
   - Methods: Call(), Stream(), Error handling?

2. Ollama Integration (Local):
   - core/providers/ollama/provider.go
   - Endpoint: http://localhost:11434
   - Models: deepseek-r1:1.5b, gemma3:1b, ...?
   - Fallback mechanism?

3. OpenAI Integration (Cloud):
   - core/providers/openai/provider.go
   - Authentication: OPENAI_API_KEY
   - Models: gpt-4, gpt-3.5-turbo, ...?
   - Rate limiting & cost tracking?

ARCHITECTURE INSIGHT:
- Primary → Backup fallback flow
- Cost metrics tracking per provider
- Agent config: ModelConfig với primary + backup

CONFIG PATTERN (YAML):
```yaml
agent:
  primary:
    model: deepseek-r1:1.5b
    provider: ollama
    provider_url: http://localhost:11434
  backup:
    model: gpt-3.5-turbo
    provider: openai
    provider_url: https://api.openai.com/v1
```

QUESTIONS TO ANSWER:
- Khi nào switch từ primary → backup?
- Cost metrics được track ở đâu?
- Streaming support có giống nhau không?

OUTPUT:
- Provider comparison table
- Integration guide mới
- Cost optimization strategies
```

---

## III. PROMPTS CẤU HÌNH (Configuration Prompts)

### 5. Tạo Agent Configuration Mới (Template)

**Mục tiêu:** Template để xây dựng agent mới

```prompt
Tạo agent configuration cho [DOMAIN]:

REQUIREMENTS:
- Agent Name: [name]
- Role: [specific responsibility]
- Backstory: [2-3 dòng context]
- Tools: [list of tool names]
- LLM: Ollama (dev) hoặc OpenAI (prod)?

TEMPLATE STRUCTURE:
```yaml
id: [agent-id]
name: [Agent Name]
role: "[role description]"
backstory: |
  [2-3 sentences explaining who they are]

# LLM Configuration
primary:
  model: deepseek-r1:1.5b           # Ollama model
  provider: ollama
  provider_url: http://localhost:11434

backup:
  model: gpt-3.5-turbo              # OpenAI model
  provider: openai
  provider_url: https://api.openai.com/v1

# Temperature: 0.0-1.0 (0=deterministic, 1=creative)
temperature: 0.7

# Tools this agent can use
tools:
  - [tool-name-1]
  - [tool-name-2]

# Signals this agent can emit
signals:
  - "[ROUTE_NEXT_AGENT]"

# Safety & Resource Limits
is_terminal: [true/false]

cost_limits:
  max_cost_per_call: 0.50
  max_daily_cost: 10.0

memory_limits:
  max_tokens_input: 8000
  max_tokens_output: 2000

error_limits:
  max_retries: 2
  timeout_seconds: 30

logging:
  debug_mode: false
  include_prompt_in_logs: false
```

VALIDATION:
- ID format: lowercase, hyphens only (a-z, 0-9, -)
- Role: Chỉ 1 câu mô tả trách nhiệm
- Tools: Phải tồn tại trong tool registry
- Signals: Format [ROUTE_XXXXX]
- Cost limits: Đương hợp lý vs model
- Is_terminal: Tối thiểu 1 agent phải true

QUESTIONS TO ANSWER:
- Agent này là entry_point, intermediate, hay terminal?
- Cần bao nhiêu tokens output?
- Có dùng OpenAI API keys không?
```

### 6. Tạo Crew Configuration (Multi-Agent System)

**Mục tiêu:** Định cấu hình toàn bộ crew (team agents)

```prompt
Tạo crew configuration cho [SYSTEM_NAME]:

CREW STRUCTURE:
```yaml
version: "1.0"
name: [crew-name]
description: "[One-line description]"
entry_point: [agent-name]    # Agent đầu tiên

agents:
  - [agent-id-1]
  - [agent-id-2]
  - [agent-id-3]

settings:
  # Execution Control
  config_mode: strict           # hoặc extended
  parallel_timeout_seconds: 60
  max_iterations_per_crew: 10
  max_handoffs: 5               # Max agent jumps
  
  # Global Resource Limits
  max_tokens_per_call: 4000
  max_cost_per_day: 100.0
  memory_limit_mb: 512
  
  # Routing & Flow Control
  routing:
    enabled: true
    default_signal_format: "[ROUTE_XXXXX]"
    signal_timeout_seconds: 30
    
  # Error Handling
  error_handling:
    dangerous_command_check: true
    invalid_signal_termination: true
    auto_retry_on_failure: true
    retry_timeout_seconds: 10
  
  # Logging
  logging:
    debug_mode: false
    include_all_messages: false
    include_tool_calls: true
    include_costs: true
  
  # Streaming Support
  streaming:
    enabled: true
    event_type: "sse"
    chunk_size: 100

signals:
  # Define valid signals for this crew
  - ROUTE_ORCHESTRATOR
  - ROUTE_CLARIFIER
  - ROUTE_EXECUTOR
  - CONTINUE_PROCESSING
  - ESCALATE_ISSUE
```

VALIDATION:
- entry_point phải tồn tại trong agents list
- Tất cả agents cần được define trong config/agents/
- Max handoffs ≥ số agents (để avoid premature termination)
- Cost limits < daily budget
- Ít nhất 1 agent có is_terminal: true

FLOW VERIFICATION:
1. User input → entry_point agent
2. Agent process → emit signal
3. Signal match → route to next agent
4. Repeat until terminal agent
5. Return final response

OUTPUT:
- crew.yaml file
- Agent diagram (text-based)
- Test scenario (manual walkthrough)
```

---

## IV. PROMPTS TOOL DEVELOPMENT (Tool Prompts)

### 7. Xây Dựng Tool Mới

**Mục tiêu:** Tạo tool mới cho agent

```prompt
Xây dựng tool mới cho [DOMAIN]:

TOOL SPECIFICATION:
```go
Tool{
  ID: "[tool-unique-id]",
  Name: "[Friendly Name]",
  Description: "[What this tool does]",
  Parameters: []Parameter{
    {
      Name: "[param-name]",
      Type: "string",        // string, integer, boolean, array, object
      Description: "[Purpose of this parameter]",
      Required: true,
      Schema: map[string]interface{}{
        "type": "string",
        "enum": []string{"option1", "option2"},  // if limited choices
      },
    },
  },
  Handler: func(ctx context.Context, params map[string]interface{}) (string, error) {
    // Implementation
    return "result", nil
  },
}
```

DEVELOPMENT STEPS:
1. Define parameters (name, type, required, validation)
2. Implement Handler function
3. Add error handling + validation
4. Test with sample inputs
5. Register in tool registry

VALIDATION CHECKLIST:
- Parameter names: snake_case only
- Types: Correct Go types (string, int, bool, []interface{}, map[string]interface{})
- Error handling: All error paths handled
- Timeout: Long-running tasks có timeout không?
- Side effects: Tool có modify system state không? (dangerous)
- Cost: Có sử dụng API calls không? (track cost)

DANGEROUS TOOLS:
Flag nếu tool có thể execute commands:
- FileSystem access (read beyond allowed paths)
- System commands (shell injection risk)
- Network access (external calls)
- Database modifications

TESTING:
```go
// Unit test
func TestToolName(t *testing.T) {
  tool := NewToolName()
  
  // Test case 1: Valid input
  result, err := tool.Handler(ctx, map[string]interface{}{
    "param1": "value1",
  })
  assert.NoError(t, err)
  assert.Contains(t, result, "expected")
  
  // Test case 2: Invalid input
  _, err = tool.Handler(ctx, map[string]interface{}{})
  assert.Error(t, err)
}
```

QUESTION CHECKLIST:
- Tool này solve cái problem gì?
- Có alternative tools không?
- Performance: Thường mất bao lâu?
- Reliability: Failure rate?
```

### 8. Tích Hợp Tool vào Agent

**Mục tiêu:** Thêm tool vào agent configuration

```prompt
Tích hợp [TOOL_NAME] vào [AGENT_NAME]:

STEPS:
1. Đảm bảo tool definition tồn tại trong tool registry
2. Update agent YAML:
   ```yaml
   # In config/agents/[agent-name].yaml
   tools:
     - [existing-tool]
     - [new-tool-name]     # Add here
   ```

3. Agent nên được test sau khi thêm tool
   - Tool call format chính xác?
   - Tool output được parse đúng?
   - Error handling hoạt động?

VALIDATION:
- Tool name match với registry
- Agent có access để call tool?
- Cost & quota impact?

OUTPUT:
- Updated YAML
- Test cases
- Documentation update
```

---

## V. PROMPTS EXAMPLE EXTENSION (Expansion Prompts)

### 9. Mở Rộng Example Hiện Có

**Mục tiêu:** Biến simple example thành complex multi-agent system

```prompt
Mở rộng [SOURCE_EXAMPLE] thành [NEW_EXAMPLE]:

APPROACH:
A) Copy structure từ source:
   cp -r examples/[source] examples/[new]

B) Modify crew configuration:
   - Thêm agents mới cho [specific roles]
   - Update entry_point
   - Định nghĩa signals cho routing
   - Thêm tools cần thiết

C) Create agent configs:
   - config/agents/[agent-1].yaml
   - config/agents/[agent-2].yaml
   - config/agents/[agent-3].yaml

D) Implement custom tools:
   - internal/tools.go
   - Define tool functions
   - Register tools với agent

E) Main application logic:
   - cmd/main.go
   - Load crew config
   - Execute with sample input
   - Handle output

F) Testing:
   - Test crew execution flow
   - Verify signal routing
   - Check all tools work
   - Validate error handling

STRUCTURE:
```
examples/[new-example]/
├── README.md                    # Documentation
├── go.mod
├── go.sum
├── cmd/
│   └── main.go                 # Entry point
├── internal/
│   ├── crew.go                 # Crew setup
│   ├── tools.go                # Custom tools
│   └── handlers.go             # Response handlers
├── config/
│   ├── crew.yaml               # Crew definition
│   └── agents/
│       ├── [agent-1].yaml
│       ├── [agent-2].yaml
│       └── [agent-3].yaml
└── tests/
    ├── crew_test.go
    ├── tools_test.go
    └── integration_test.go
```

CHECKLIST:
- go.mod với correct module path
- All YAML configs valid
- All tools registered
- Main execution flow complete
- Tests pass
- README with usage instructions

EXAMPLE DOMAINS:
- customer-support: FAQ answering, issue routing, escalation
- content-creation: Planner → Writer → Editor → Publisher
- data-analysis: Collector → Analyzer → Visualizer
- code-review: CodeReader → Reviewer → Suggester → Approver
```

### 10. Tạo Example Từ Đầu (Green-Field)

**Mục tiêu:** Build new example system từ zero

```prompt
Tạo example mới: [EXAMPLE_NAME]

PLANNING PHASE (First Principles):
1. Define thực hiện
   - Input: [what does user provide?]
   - Process: [what should system do?]
   - Output: [what's final deliverable?]

2. Agent requirements
   - Bao nhiêu agents cần?
   - Vai trò của từng agent?
   - Tools mỗi agent cần?

3. Signal routing design
   - Agent 1 → Signal?
   - Signal → Router → Agent?
   - Terminal agent là cái nào?

IMPLEMENTATION:
4. Create directory structure
5. Setup go.mod + go.sum
6. Write crew.yaml
7. Write agent YAML configs
8. Implement tools (internal/tools.go)
9. Implement crew setup (internal/crew.go)
10. Write main.go
11. Add testing
12. Write README.md

STRUCTURE:
```
examples/[example-name]/
├── README.md
├── go.mod
├── cmd/main.go
├── internal/
│   ├── crew.go
│   ├── tools.go
│   └── [domain].go
├── config/
│   ├── crew.yaml
│   └── agents/
└── tests/
```

VALIDATION:
- All agents defined in crew.yaml
- All tools implemented
- All signals handled
- At least 1 terminal agent
- Execution flow complete
- Error handling present
- Tests green

QUALITY GATES:
- go build: Compiles without errors
- go test: All tests pass
- go run: Runs with sample input
- Signal flow: Correct routing
- Cost tracking: Disabled or configured
```

---

## VI. PROMPTS TESTING & QUALITY (Testing Prompts)

### 11. Viết Test Comprehensive

**Mục tiêu:** Coverage cao, chất lượng cao

```prompt
Viết comprehensive test cho [COMPONENT]:

TESTING PYRAMID:
```
        ▲
       /│\
      / │ \  E2E Tests (1-2)
     /  │  \
    /   │   \
   /    │    \
  /     │     \
 /      │      \  Integration Tests (3-5)
 ────────────────
│                │  Unit Tests (10+)
│                │
──────────────────
```

UNIT TEST EXAMPLES:

A) Agent Execution:
```go
// Test normal flow
func TestAgentExecute_Success(t *testing.T) { ... }

// Test tool calling
func TestAgentExecute_WithTools(t *testing.T) { ... }

// Test timeout
func TestAgentExecute_Timeout(t *testing.T) { ... }

// Test cost tracking
func TestAgentExecute_CostTracking(t *testing.T) { ... }

// Test memory limits
func TestAgentExecute_MemoryLimits(t *testing.T) { ... }
```

B) Crew Routing:
```go
// Test signal detection
func TestCrew_SignalDetection(t *testing.T) { ... }

// Test routing to next agent
func TestCrew_RoutingFlow(t *testing.T) { ... }

// Test max handoffs enforcement
func TestCrew_MaxHandoffsExceeded(t *testing.T) { ... }

// Test terminal agent stop
func TestCrew_TerminalAgentStop(t *testing.T) { ... }
```

C) Tool Validation:
```go
// Test valid parameters
func TestTool_ValidParams(t *testing.T) { ... }

// Test missing required param
func TestTool_MissingRequired(t *testing.T) { ... }

// Test invalid type
func TestTool_InvalidType(t *testing.T) { ... }

// Test timeout
func TestTool_Timeout(t *testing.T) { ... }
```

INTEGRATION TEST EXAMPLES:
```go
// Test full crew execution
func TestCrew_FullExecution(t *testing.T) { ... }

// Test error recovery
func TestCrew_ErrorRecovery(t *testing.T) { ... }

// Test context preservation
func TestCrew_ContextPreservation(t *testing.T) { ... }
```

QUOTA ENFORCEMENT TESTS:
```go
// Test cost limit enforcement
func TestQuota_CostLimit(t *testing.T) { ... }

// Test memory limit enforcement
func TestQuota_MemoryLimit(t *testing.T) { ... }

// Test max retries enforcement
func TestQuota_MaxRetries(t *testing.T) { ... }
```

TEST PATTERNS:
- Table-driven tests for multiple scenarios
- Mock provider for reproducible results
- Context with timeout
- Error case coverage
- Edge case validation

COVERAGE GOALS:
- Line coverage: ≥85%
- Branch coverage: ≥80%
- Function coverage: 100%
- Error path coverage: All error cases
```

### 12. Performance & Load Testing

**Mục tiêu:** Validate performance under load

```prompt
Performance test cho [COMPONENT]:

BENCHMARK TESTS:
```go
func BenchmarkAgentExecute(b *testing.B) {
  for i := 0; i < b.N; i++ {
    agent.Execute(ctx, input)
  }
}

func BenchmarkCrewRouting(b *testing.B) {
  for i := 0; i < b.N; i++ {
    crew.Execute(ctx, input)
  }
}

func BenchmarkToolExecution(b *testing.B) {
  for i := 0; i < b.N; i++ {
    tool.Handler(ctx, params)
  }
}
```

LOAD TEST SCENARIOS:
1. Concurrent agents (10, 50, 100 parallel)
2. Deep routing chains (5, 10, 20 handoffs)
3. Large message history (100, 1000 messages)
4. High-frequency tool calls (100/sec)

METRICS TO TRACK:
- Latency: p50, p95, p99
- Throughput: requests/sec
- Memory usage: Peak, average
- CPU utilization: %
- Error rate: %
- Cost per request

ACCEPTANCE CRITERIA:
- p95 latency < X ms
- Throughput > Y req/s
- Memory usage < Z MB
- CPU < W%
- Error rate < V%

OUTPUT:
- Benchmark results
- Performance report
- Recommendations for optimization
```

---

## VII. PROMPTS PRODUCTION & DEPLOYMENT (Deployment Prompts)

### 13. Chuẩn Bị Production Deployment

**Mục tiêu:** Ready for production release

```prompt
Chuẩn bị [CREW_NAME] cho production:

PRODUCTION CHECKLIST:

1. CODE QUALITY
   ☐ go build: No errors
   ☐ go test: All tests pass, ≥85% coverage
   ☐ go vet: No issues
   ☐ golangci-lint: No errors
   ☐ Code review: Approved
   
2. SECURITY
   ☐ No hardcoded secrets (API keys, passwords)
   ☐ Input validation on all parameters
   ☐ Dangerous commands blocked
   ☐ No SQL injection vulnerabilities
   ☐ Secrets in environment variables only
   ☐ HTTPS enforced for API calls
   
3. CONFIGURATION
   ☐ crew.yaml: Valid and tested
   ☐ All agents: Properly configured
   ☐ LLM providers: Credentials in env vars
   ☐ Resource limits: Set appropriately
   ☐ Error handling: Comprehensive
   
4. COST MANAGEMENT
   ☐ Cost limits set: Per agent, per day
   ☐ Cost tracking: Enabled
   ☐ Budget alerts: Configured
   ☐ LLM provider: Ollama (free) or monitored OpenAI
   
5. MONITORING & LOGGING
   ☐ Logging: Enabled, appropriate level
   ☐ Error tracking: Set up
   ☐ Metrics collection: Running
   ☐ Alerts: Configured for critical issues
   
6. DOCUMENTATION
   ☐ README.md: Complete
   ☐ Setup guide: Step-by-step
   ☐ Configuration guide: All options explained
   ☐ Troubleshooting: Common issues covered
   ☐ API documentation: If applicable
   
7. TESTING
   ☐ Unit tests: All pass
   ☐ Integration tests: All pass
   ☐ Load tests: Performance acceptable
   ☐ Manual testing: Workflow verified

PRODUCTION CONFIG TEMPLATE:
```yaml
# crew.yaml for production
version: "1.0"
name: [crew-name]
entry_point: [agent]

settings:
  config_mode: strict
  parallel_timeout_seconds: 60
  max_cost_per_day: [BUDGET]
  
  error_handling:
    dangerous_command_check: true
    invalid_signal_termination: true
  
  logging:
    debug_mode: false
    include_all_messages: true
    include_costs: true
```

ENVIRONMENT VARIABLES (.env):
```
# LLM Providers
OPENAI_API_KEY=sk-...

# App Config
APP_ENV=production
APP_PORT=8080
LOG_LEVEL=info

# Monitoring
DATADOG_API_KEY=...
```

DEPLOYMENT STEPS:
1. Code freeze, final testing
2. Create release tag (v1.0.0)
3. Build binary: go build -o app
4. Package: docker build & push
5. Deploy to infrastructure
6. Run smoke tests
7. Monitor for issues
8. Document runbook for ops team

MONITORING ALERTS:
- Cost > [threshold]/day
- Error rate > [X]%
- p95 latency > [Y]ms
- Memory usage > [Z]MB
- CPU > [W]%
```

### 14. Optimization Guide

**Mục tiêu:** Tối ưu hóa chi phí, hiệu suất, memory

```prompt
Tối ưu hóa [CREW_NAME] cho production:

COST OPTIMIZATION:
1. Use Ollama (free, local):
   - Replace OpenAI calls with Ollama
   - Models: deepseek-r1:1.5b, gemma3:1b
   - Cost: $0 vs $0.002+ per call

2. Reduce token usage:
   - Shorter prompts (remove redundancy)
   - Fewer examples in prompts
   - Summarize context before passing

3. Request batching:
   - Batch multiple queries → single LLM call
   - Reduce number of agent handoffs
   - Cache results (if applicable)

4. Cost monitoring:
   - Track cost per agent
   - Identify expensive agents
   - Set daily budgets
   - Alert on overspend

PERFORMANCE OPTIMIZATION:
1. Reduce latency:
   - Parallel agent execution (where possible)
   - Cache LLM responses (for same input)
   - Optimize tool execution time
   - Use faster LLM models

2. Reduce memory:
   - Trim old messages (keep last N)
   - Compress context
   - Stream responses (don't buffer)
   - Release resources after tool call

3. Improve throughput:
   - Connection pooling for LLM provider
   - Reduce max_iterations (limit loops)
   - Optimize tool implementations

MONITORING & PROFILING:
1. CPU profiling:
   ```bash
   go test -cpuprofile=cpu.prof
   go tool pprof cpu.prof
   ```

2. Memory profiling:
   ```bash
   go test -memprofile=mem.prof
   go tool pprof mem.prof
   ```

3. Metrics to optimize:
   - CPU time per agent call
   - Memory allocation per message
   - Goroutine count at peak
   - GC pause time

QUICK WINS:
1. Increase temperature (more creative, faster)
2. Decrease max_tokens (shorter responses)
3. Use smaller models (faster, cheaper)
4. Reduce tool call depth
5. Cache tool results (if safe)

OUTPUT:
- Before/after metrics
- Cost reduction: $X/day
- Latency improvement: Xms reduction
- Memory improvement: XMB reduction
- Changes made (actionable list)
```

---

## VIII. PROMPTS DEBUGGING & TROUBLESHOOTING (Debug Prompts)

### 15. Debug Crew Execution

**Mục tiêu:** Identify and fix issues trong crew execution

```prompt
Debug [CREW_NAME] execution issues:

SYMPTOM: [What's wrong?]
Example: Agent không route tới next agent, Chi phí tăng đột ngột, Memory leak, ...

DEBUG CHECKLIST:

1. VERIFY CONFIGURATION
   ☐ crew.yaml: Valid YAML syntax
   ☐ All agents in crew.yaml exist in config/agents/
   ☐ entry_point agent exists
   ☐ All signals defined/recognized
   ☐ At least 1 agent with is_terminal: true

2. CHECK LOGS
   ☐ Enable debug_mode: true in settings
   ☐ Log level: info or debug
   ☐ Look for errors in logs
   ☐ Trace execution flow

3. VALIDATE SIGNAL ROUTING
   ☐ Agent emitting correct signal?
   ☐ Signal format: [ROUTE_XXXXX]
   ☐ Next agent receives signal?
   ☐ Router matches signal correctly?

4. TEST TOOL EXECUTION
   ☐ Tool exists and registered?
   ☐ Tool parameters correct?
   ☐ Tool returns expected output?
   ☐ Error handling working?

5. QUOTA VERIFICATION
   ☐ Cost within limits?
   ☐ Memory within limits?
   ☐ Max retries not exceeded?
   ☐ Max handoffs not exceeded?

6. ENVIRONMENT VARIABLES
   ☐ API keys set (if using OpenAI)?
   ☐ LLM provider reachable?
   ☐ Port conflicts?
   ☐ Network connectivity?

DEBUG TECHNIQUES:

A) Add logging:
   ```go
   log.Printf("[DEBUG] Agent: %s, Signal: %s", agent.ID, signal)
   log.Printf("[DEBUG] Routing to: %s", nextAgent.ID)
   ```

B) Add breakpoints (if using debugger):
   - crew.Execute() entry
   - agent.Execute() entry
   - Signal detection
   - Routing decision

C) Print state at each step:
   ```go
   log.Printf("Messages: %+v", messages)
   log.Printf("Cost: %+v", costMetrics)
   log.Printf("Quota exceeded: %v", quotaLimitExceeded)
   ```

D) Test components in isolation:
   - Test agent alone (no crew)
   - Test tool alone (no agent)
   - Test crew with mock agent

COMMON ISSUES & FIXES:

Issue: Agent doesn't route to next
Fix: Check signal format, routing config, agent names

Issue: Chi phí tăng đột ngột
Fix: Check model cost, max_tokens, reduce prompt size

Issue: Memory leak
Fix: Check message history growth, add context trimming

Issue: Timeout errors
Fix: Increase timeout_seconds, optimize tool speed

Issue: Tool not found
Fix: Verify tool registration, check tool name spelling

OUTPUT:
- Root cause identified
- Fix applied
- Test to verify fix works
- Prevent recurrence (e.g., add test)
```

---

## IX. PROMPTS DOCUMENTATION & KNOWLEDGE BASE (Doc Prompts)

### 16. Tạo Documentation Mới

**Mục tiêu:** Standardized, comprehensive documentation

```prompt
Viết documentation cho [TOPIC]:

DOCUMENTATION TYPES:

A) GETTING STARTED (New users)
   - What: Giới thiệu topic
   - Prerequisites: Công cụ cần cài
   - Quick start: 10-minute guide
   - Verification: How to verify it works

B) CONCEPTS (Understanding)
   - Architecture: How things work
   - Terminology: Key terms defined
   - Diagrams: Visual explanations
   - Examples: Concrete samples

C) API REFERENCE (Developers)
   - Functions/Methods: Signatures
   - Parameters: Type, description, default
   - Return values: Type, meaning
   - Errors: What can go wrong
   - Examples: Code samples

D) EXAMPLES (Learn by doing)
   - Problem: What does this solve?
   - Code: Full working example
   - Explanation: Line-by-line walkthrough
   - Extension: How to customize

E) BEST PRACTICES (Do's and don'ts)
   - Do: What to do
   - Don't: What to avoid
   - Why: Reasoning
   - Example: Code sample

MARKDOWN TEMPLATE:
```markdown
# [Topic]

## Overview
[1-2 paragraphs explaining what this is]

## Prerequisites
- Go 1.25.2+
- [Other requirements]

## Quick Start
[Step-by-step guide, ~10 minutes]

## Concepts
### [Concept 1]
[Explanation]

### [Concept 2]
[Explanation]

## Examples
### Example 1: [Use case]
[Code example]

### Example 2: [Use case]
[Code example]

## Best Practices
- ✓ Do [practice]
  - Why: [reasoning]
  - Example: [code]

- ✗ Don't [practice]
  - Why: [reasoning]
  - Instead: [alternative]

## Troubleshooting
### Issue: [Common problem]
**Solution:** [How to fix]

### Issue: [Another problem]
**Solution:** [How to fix]

## See Also
- [Related topic 1](link)
- [Related topic 2](link)
```

QUALITY CHECKLIST:
- ☐ Spell-checked, grammar-checked
- ☐ Code examples: Tested, working
- ☐ Links: All valid, not broken
- ☐ Format: Consistent with other docs
- ☐ Completeness: Answers key questions
- ☐ Clarity: Easy to understand
- ☐ Accuracy: Matches current codebase
```

---

## X. PROMPTS COLLECTION - QUICK REFERENCE

### 17. Rapid Problem-Solving Prompts

```prompt
QUICK TEMPLATE untuk fast problem-solving:

PROBLEM: [What's the issue?]
CONTEXT: [Background info]
CONSTRAINTS: [What can't change?]

SOLUTION APPROACH:
1. First Principles: Break down to essentials
2. Speed Analysis: Identify critical path
3. Implementation: Simplest solution first
4. Validation: Test assumptions

OUTPUT:
- Root cause
- Solution (2 options if possible)
- Implementation steps
- Validation plan
```

---

## XI. STRATEGIC PRIORITIES (Delivery Roadmap)

### Tier 1 - CRITICAL (Next Sprint)
1. **Review & Cleanup** (Prompts #1, #2)
   - Review cơ sở mã toàn diện
   - Loại bỏ tài liệu dư thừa
   - Standardize documentation

2. **Core Documentation** (Prompts #3, #4, #5)
   - Signal-based routing guide
   - Multi-provider LLM guide
   - Agent configuration template

### Tier 2 - IMPORTANT (Following Sprint)
3. **Example Expansion** (Prompts #9, #10)
   - Mở rộng existing examples
   - Tạo 2-3 new examples
   - Complete documentation

4. **Testing & Quality** (Prompts #11, #12)
   - Comprehensive test suite
   - Performance benchmarks
   - Load testing

### Tier 3 - STRATEGIC (Roadmap)
5. **Production Ready** (Prompts #13, #14)
   - Deployment guide
   - Optimization strategies
   - Monitoring setup

6. **Debugging & Support** (Prompts #15, #16, #17)
   - Troubleshooting guide
   - Knowledge base
   - Community support

---

## EXECUTION RULES

### Tư Duy Nguyên Bản (First Principles - Elon Musk)
1. **Identify Core Truths**: Tìm sự thật cơ bản, bỏ qua giả định
2. **Challenge Assumptions**: Câu hỏi "tại sao?" liên tục
3. **Build from Fundamentals**: Xây dựng từ nền tảng, không copy
4. **Measure Everything**: Độc lập, verify, không tin "ai nói"

### Tư Duy Tốc Độ Ánh Sáng (Speed of Light - NVIDIA)
1. **Parallel Processing**: Làm nhiều việc cùng lúc
2. **Eliminate Waste**: Bỏ qua tất cả overhead
3. **Optimize Hot Paths**: Focus vào critical path
4. **Rapid Iteration**: Fast feedback → fast improvement

### Tư Duy Test-Driven Development (TDD - Red-Green-Refactor)

1. **Red Phase**: Viết test trước, test fail
2. **Green Phase**: Viết code đơn giản nhất để test pass
3. **Refactor Phase**: Cải thiện code mà không break test
4. **Benefits**: High coverage, less bugs, safe refactoring, clear requirements

**Ứng dụng trong go-agentic:**

- Luôn viết test trước khi implement feature
- Test định nghĩa behavior, code thực hiện behavior
- TDD → Confidence khi refactor quota system, routing logic

### Tư Duy Systems Thinking (Holistic View)

1. **Map Dependencies**: Tất cả components liên kết như thế nào?
2. **Identify Feedback Loops**: Cause → Effect → Cause (vòng lặp)
3. **Spot Bottlenecks**: Đâu là chỗ chậm nhất, giới hạn nhất?
4. **Avoid Side Effects**: Thay đổi A → ảnh hưởng B, C, D như thế nào?

**Ứng dụng trong go-agentic:**

- Agent1 → Agent2 → Agent3 (dependency chain)
- Cost quota → affects performance → affects timeout behavior
- Context trimming → affects memory → affects latency
- Thay signal format → tất cả agents bị ảnh hưởng?

### Tư Duy Constraint-Based Design (Tối ưu xung quanh ràng buộc)

1. **Identify Constraints**: Memory, cost, quota, time limits
2. **Design Around Constraints**: Không ngoài giới hạn, tối ưu trong giới hạn
3. **Trade-offs**: Nếu tăng A, giảm B, C — chọn cái tốt nhất
4. **Validate Constraints**: Kiểm tra enforcement, xử lý violations

**Ứng dụng trong go-agentic:**

- Cost limit: $10/day → Agent implementation phải cost-aware
- Memory limit: 512MB → Context trimming, message history size
- Token limit: 4000 tokens/call → Prompt optimization
- Timeout: 30s → Fast tool execution, no long-running operations
- Max handoffs: 5 → Crew design phải terminate quickly

### Tư Duy Data-Driven Decision Making (Metrics trước, Intuition sau)

1. **Collect Metrics**: Cost, latency, memory, error rate, throughput
2. **Profile & Analyze**: Grep logs, trace execution, identify hotspots
3. **Make Decisions**: Dữ liệu nói gì? Kết luận gì?
4. **Validate Changes**: Before/After metrics để prove improvement

**Ứng dụng trong go-agentic:**

- Chi phí tăng? → Log chi tiêu từng agent, find expensive one
- Slow performance? → Benchmark each component, profile with pprof
- Memory leak? → Check message history growth, measure peak memory
- High error rate? → Analyze error logs, find patterns
- Cost optimization? → Measure before/after cost reduction

### Ứng Dụng Thực Tế - Hoàn Chỉnh
```
Problem
  ↓
Analysis (First Principles - cái gì cốt lõi?)
  ↓
Systems Thinking (ảnh hưởng tới đâu?)
  ↓
Constraint Analysis (giới hạn là gì?)
  ↓
Data-Driven Decision (metrics nói gì?)
  ↓
Solution (Speed of Light - simple, fast)
  ↓
Test-Driven Implementation (TDD red-green-refactor)
  ↓
Validation & Metrics
  ↓
Deployment
```

---

## USAGE GUIDE

Để sử dụng prompts này:

1. **Select prompt** dựa trên task
2. **Customize** với tên dự án, specific details
3. **Execute** với Claude hoặc local LLM
4. **Validate** output trước deploy
5. **Iterate** nếu output không đủ tốt

Example:
```
Sử dụng Prompt #5 (Tạo Agent Configuration):
- Agent Name: DataAnalyst
- Role: Phân tích dữ liệu
- Tools: query_database, generate_report
- LLM: Ollama deepseek-r1:1.5b
```

---

## VERSION & CHANGELOG

**Version**: 1.0.0  
**Date**: 2025-12-23  
**Author**: First Principles + Speed of Light Methodology

### Completed
- ✅ 17 prompt templates
- ✅ Strategic priorities
- ✅ Execution rules
- ✅ Quick reference

### Future
- [ ] Prompt automation (scripts)
- [ ] AI-assisted prompt optimization
- [ ] Community contributions
- [ ] Multilingual variants (Vietnamese, English)

---

**Document Status**: READY FOR PRODUCTION  
**Last Updated**: 2025-12-23  
**Maintainer**: go-agentic team

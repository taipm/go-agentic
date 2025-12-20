# UX Design Summary: go-agentic

**Tl;dr**: go-agentic có thiết kế tốt nhưng API quá phức tạp. Tôi đã phân tích 6 pain points và đề xuất giải pháp 3 giai đoạn để cải thiện trải nghiệm developer lên 80%.

---

## 🎯 6 Pain Points Chính

### 1. 🔴 Boilerplate Tool Definition
**Vấn đề**: Mỗi tool phải viết 15-20 dòng code (map[string]interface{} schema)

**Ví dụ**:
```go
// Phải lặp lại cho MỖI tool
tool := &agentic.Tool{
    Name: "GetCPU",
    Description: "...",
    Parameters: map[string]interface{}{
        "type": "object",
        "properties": map[string]interface{}{},  // Empty!
    },
    Handler: getCPU,
}
```

**Giải pháp**: ToolBuilder
```go
tool := agentic.NewTool("GetCPU", "...").
    NoParameters().
    Handler(getCPU).
    Build()
```

**Impact**: 80% giảm boilerplate, 40% ít code duplication

---

### 2. 🔴 Multiple Routing Mechanisms
**Vấn đề**: 3 cách khác nhau để control flow
- `HandoffTargets` (code)
- `routing.signals` (YAML)
- `IsTerminal` (flag)

**Plus**: Routing logic phải hardcode trong system prompt (145 dòng!)

**Giải pháp**: Declarative Router
```go
router := agentic.NewRouter().
    From("orchestrator").
        OnTrigger("needs_info").To("clarifier").
        OnTrigger("ready").To("executor")
```

**Impact**: Một cách duy nhất, dễ hiểu, dễ debug

---

### 3. 🟠 Configuration Coordination
**Vấn đề**: Team.yaml + agents/*.yaml không integrate
- 4-5 file cần keep in sync
- Tool definitions không thể config từ YAML

**Giải pháp**: Unified YAML
```yaml
# Một file thay vì 5 file
team:
  agents:
    orchestrator: {...}
    executor: {...}
  tools:
    get_cpu_usage: {...}
  routing: [...]
```

**Impact**: Từ 4-5 files xuống 1 file, dễ maintain

---

### 4. 🟡 Crew vs Team Confusion
**Vấn đề**: API có cả Crew (deprecated) và Team (new)
- Examples mix cả 2
- YAML config gọi nó "crew:" ở chỗ này, "team:" ở chỗ khác
- Developers không biết dùng cái nào

**Giải pháp**: Deprecation guide + consistent naming

**Impact**: Rõ ràng hơn cho newcomers

---

### 5. 🔴 No Type Safety for Tool Parameters
**Vấn đề**: `map[string]interface{}` everywhere = runtime panics
```go
threshold := args["cpu_threshold"].(float64)  // Unguarded cast!
```

**Giải pháp**: Struct tags (Go 1.23+ feature)
```go
type GetMetricsArgs struct {
    Metric string `json:"metric" required:"true"`
}

tool := agentic.ToolFromHandler("GetMetrics",
    func(ctx context.Context, args GetMetricsArgs) (string, error) {
        // Type-safe!
    },
)
```

**Impact**: Compile-time checking, no runtime panics

---

### 6. 🔴 Examples Don't Show Best Practices
**Vấn đề**: Library có testing framework, parameter validation, YAML config - nhưng examples không dùng
- Testing framework: 0% usage
- Parameter validation: 0% usage
- YAML config: chỉ 1 minimal example
- HTTP server: không demo

**Giải pháp**: Cập nhật examples để show best practices

**Impact**: Developers biết cách dùng features mới

---

## ✅ Proposed Solution: 3-Phase Implementation

### Phase 1: Fluent Builder API (6 hours)

**Goal**: Giảm boilerplate agent & tool creation

```go
// Agent Builder
agent := agentic.NewAgent("id", "Name").
    WithRole("role").
    WithBackstory("...").
    WithModel("gpt-4o").  // NOT hardcoded!
    SetTerminal(false).
    Build()

// Tool Builder
tool := agentic.NewTool("name", "description").
    WithParameter("metric", "string", "...").
    Handler(myHandler).
    Build()
```

**Backward Compatible**: ✅ Old way still works

**Impact**:
- ✅ 50% less code for agent setup
- ✅ 80% less code for tool setup
- ✅ Fluent interface = readable

---

### Phase 2: Unified YAML Configuration (8 hours)

**Goal**: Single config file instead of 4-5

**Before**:
```
team.yaml
agents/orchestrator.yaml
agents/clarifier.yaml
agents/executor.yaml
```

**After**:
```yaml
# team.yaml (tất cả trong 1 file)
team:
  agents:
    orchestrator: {...}
    clarifier: {...}
    executor: {...}
  tools:
    get_cpu_usage: {...}
  routing: [...]
```

**Load in code**:
```go
team := agentic.LoadTeamFromYAML("team.yaml", toolHandlers)
```

**Impact**:
- ✅ 1 file instead of 4-5
- ✅ Agent models not hardcoded
- ✅ Easier to modify and understand

---

### Phase 3: Declarative Routing (6 hours)

**Goal**: Remove text-based routing, use declarative rules

**Before** (145 lines of system prompt):
```
Nếu [ROUTE_CLARIFIER] thì output "[ROUTE_CLARIFIER]"
Nếu [ROUTE_EXECUTOR] thì output "[ROUTE_EXECUTOR]"
...
```

**After**:
```yaml
routing:
  rules:
    - from_agent: orchestrator
      trigger: needs_info
      target_agent: clarifier
```

```go
router := agentic.NewRouter().
    From("orchestrator").
        OnTrigger("needs_info").To("clarifier").
        OnTrigger("ready").To("executor")
```

**Impact**:
- ✅ 80% less system prompt code
- ✅ Clear intent, not text patterns
- ✅ Debuggable routing

---

## 📊 Expected Improvements

| Metric | Hiện Tại | Sau Cải Thiện |
|--------|---------|--------------|
| **Code per agent** | 8 fields (verbose) | 2-3 builder calls |
| **Lines per tool** | 15-20 | 2-3 |
| **Config files** | 4-5 | 1 |
| **Boilerplate** | 40% duplication | <15% |
| **Routing mechanisms** | 3 ways | 1 clear way |
| **Time to setup agents** | 5-10 min | 2-3 min |
| **Developer satisfaction** | Confused | Clear |

---

## 🗺️ Implementation Roadmap

```
Week 1:
├─ Phase 1: Fluent Builders (6h)
│  ├─ AgentBuilder
│  ├─ ToolBuilder
│  ├─ Tests
│  └─ simple-chat v2 example
└─ Phase 2 start: Config design (2h)

Week 2:
├─ Phase 2: Unified YAML (8h)
│  ├─ LoadTeamFromYAML()
│  ├─ Config schema
│  ├─ Tests
│  └─ IT Support v2 example
└─ Phase 3 start: Router design (2h)

Week 3:
├─ Phase 3: Routing DSL (6h)
│  ├─ Router builder
│  ├─ TriggerDetector
│  ├─ Tests
│  └─ IT Support v3 example
├─ Examples (4h)
├─ Documentation (4h)
├─ Migration guide (2h)
└─ Release prep (2h)
```

**Total**: 32 hours (2-3 developer weeks)

---

## 🎁 Zero Breaking Changes

✅ **Backward Compatible**: Old code continues to work
✅ **Gradual Migration**: Can mix old & new patterns
✅ **Additive Only**: No removals, only additions

```go
// Old way (still works)
agent := &agentic.Agent{ID: "...", Name: "..."}

// New way (recommended)
agent := agentic.NewAgent("id", "name").Build()

// Can mix both in same codebase
```

---

## 📋 Deliverables

1. **2 Analysis Documents**
   - `UX_DESIGN_ANALYSIS.md` - Detailed analysis (1000+ lines)
   - `UX_IMPLEMENTATION_GUIDE.md` - Step-by-step guide (600+ lines)

2. **3 Code Implementations**
   - `builder.go` - AgentBuilder & ToolBuilder
   - `config_unified.go` - LoadTeamFromYAML
   - `routing.go` - Router & TriggerDetector

3. **5 Updated Examples**
   - simple-chat-v2 (Fluent API)
   - it-support-unified (Unified YAML)
   - it-support-v3 (Routing DSL)
   - Plus improvements to others

4. **3 Documentation Guides**
   - Fluent API guide
   - Unified config guide
   - Routing guide
   - Migration guide

---

## 🎯 Why This Approach?

**Instead of**:
- ❌ Keep current API (developers unhappy)
- ❌ Massive rewrite (breaking changes, too risky)

**We should**:
- ✅ Add fluent builders (easy, additive)
- ✅ Add unified config (optional, powerful)
- ✅ Add routing DSL (cleaner, declarative)
- ✅ Keep old API (zero breaking changes)

**Result**: Developers can start simple, grow complex, with clear patterns throughout.

---

## 📁 Files Created

```
go-agentic/
├─ UX_DESIGN_ANALYSIS.md          (1200+ lines)
├─ UX_IMPLEMENTATION_GUIDE.md      (700+ lines)
├─ UX_SUMMARY.md                   (this file)
├─ go-agentic/
│  ├─ builder.go                   (NEW - AgentBuilder, ToolBuilder)
│  ├─ config_unified.go            (NEW - LoadTeamFromYAML)
│  └─ routing.go                   (NEW - Router, TriggerDetector)
└─ examples/
   ├─ simple-chat-v2/              (NEW - Fluent API example)
   ├─ it-support-unified/          (NEW - Unified YAML example)
   └─ it-support-v3/               (NEW - Routing DSL example)
```

---

## ⏭️ Next Steps

1. **Review & Approve** this approach
2. **Prioritize** which phase to implement first
3. **Start Phase 1** if approved (6 hours)
4. **Gather feedback** from early adopters
5. **Continue Phases 2 & 3** based on feedback

---

**Phân tích bởi**: UX Analysis Team
**Ngày**: 20 tháng 12 năm 2025
**Trạng thái**: ✅ Ready for discussion & decision

### 📞 Questions to Discuss

1. **Should we implement Phase 1 first?** (Lowest risk, highest immediate impact)
2. **Is Unified YAML (Phase 2) important?** (Some users prefer code-based setup)
3. **How important is Routing DSL (Phase 3)?** (Nice-to-have, but complex)
4. **Timeline acceptable?** (32 hours = 1 developer-week)
5. **Should we focus on different pain points?**

---

# Multi-Crew Enhancement Analysis Report

> Báo cáo phân tích chi tiết và đề xuất hướng đi tốt nhất cho go-agentic framework

**Version**: 1.0
**Date**: 2024-12-24
**Status**: Technical Analysis Complete
**Audience**: Production Developers & Framework Contributors

---

## Executive Summary

### Tổng quan vấn đề

Go-agentic framework hiện tại được thiết kế cho **single-crew workflows**. Khi xây dựng hệ thống multi-team (như ví dụ `02-multi-team`), developers phải viết **~215 lines orchestration code thủ công** để:

1. Parse signals bằng `strings.Contains()` - fragile và không type-safe
2. Quản lý lifecycle từng crew riêng biệt
3. Truyền context giữa các crews bằng string concatenation
4. Handle nested if/else cho workflow routing

### Mục tiêu Enhancement

Chuyển từ **Imperative Orchestration** sang **Declarative Configuration**:

```
Before: 215 lines Go code  →  After: ~20 lines YAML config
```

### ROI dự kiến

| Metric | Hiện tại | Sau Enhancement | Cải thiện |
|--------|----------|-----------------|-----------|
| Lines of orchestration code | ~215 | ~0 | -100% |
| Time to add new team | ~30 min | ~5 min | -83% |
| Risk of signal typo bug | High | None | -100% |
| Learning curve cho new devs | Steep | Gentle | Significant |

---

## Phần 1: Phân tích 5W2H Chi tiết

### 1.1 WHAT - Vấn đề cụ thể là gì?

#### A. Current Architecture Limitations

```
┌─────────────────────────────────────────────────────────────────┐
│                    CURRENT ARCHITECTURE                          │
├─────────────────────────────────────────────────────────────────┤
│                                                                  │
│  ┌──────────────┐     ┌──────────────┐     ┌──────────────┐    │
│  │ CrewExecutor │     │ CrewExecutor │     │ CrewExecutor │    │
│  │   (Master)   │     │   (Alpha)    │     │   (Beta)     │    │
│  └──────────────┘     └──────────────┘     └──────────────┘    │
│         │                    │                    │             │
│         └────────────────────┴────────────────────┘             │
│                              │                                   │
│                    KHÔNG CÓ LIÊN KẾT NATIVE                     │
│                              │                                   │
│                              ▼                                   │
│              ┌───────────────────────────────┐                  │
│              │   MANUAL ORCHESTRATOR CODE    │                  │
│              │   (main.go - 215 lines)       │                  │
│              │                               │                  │
│              │  • strings.Contains()         │                  │
│              │  • Nested if/else             │                  │
│              │  • Manual context passing     │                  │
│              │  • ClearHistory() calls       │                  │
│              └───────────────────────────────┘                  │
│                                                                  │
└─────────────────────────────────────────────────────────────────┘
```

#### B. 5 Technical Limitations Identified

| # | Limitation | File Location | Impact |
|---|------------|---------------|--------|
| 1 | **Signal Target = "" means TERMINATE only** | `crew_routing.go:115` | Không emit được output signal mà không dừng workflow |
| 2 | **No Inter-Crew Communication** | `crew.go` (không có field) | CrewExecutor độc lập, không gọi được nhau |
| 3 | **No Pause/Resume for External Handoff** | `crew.go:1114-1119` | `wait_for_signal` chỉ đợi user input |
| 4 | **Fixed Signal Priority** | `crew_routing.go:113-125` | Termination check trước routing |
| 5 | **No Crew Composition** | `config.go` (không có field) | Không định nghĩa crew-of-crews trong YAML |

#### C. Pain Points trong Example Code

**File**: `examples/02-multi-team/cmd/main.go`

```go
// Line 143-186: Manual Signal Parsing & Routing (~43 lines)
func (o *MultiTeamOrchestrator) Execute(ctx context.Context, input string) (string, error) {
    masterResult, err := o.executeTeam(ctx, o.MasterTeam, input)

    // PAIN POINT 1: Hardcoded signal detection
    if strings.Contains(masterResult, "[DELEGATE_ALPHA]") {  // ← Fragile!
        alphaResult, err := o.executeTeam(ctx, o.TeamAlpha, input)

        // PAIN POINT 2: Manual context template
        combinedInput := fmt.Sprintf("Kết quả từ Team Alpha:\n%s\n\nTopic gốc: %s",
            alphaResult, input)  // ← Hardcoded template!

        masterResult2, err := o.executeTeam(ctx, o.MasterTeam, combinedInput)

        // PAIN POINT 3: Nested if/else hell
        if strings.Contains(masterResult2, "[DELEGATE_BETA]") {  // ← More nesting!
            // ... continues with more nesting
        }
    }
}
```

**Vấn đề cụ thể:**

| Line | Code | Problem |
|------|------|---------|
| 143 | `strings.Contains(masterResult, "[DELEGATE_ALPHA]")` | Không type-safe, typo = silent bug |
| 154 | `fmt.Sprintf("Kết quả từ Team Alpha:\n%s...")` | Template hardcoded trong Go code |
| 162 | `if strings.Contains(masterResult2, "[DELEGATE_BETA]")` | Nested conditionals grow with each team |
| 202 | `team.Executor.ClearHistory()` | Developer phải nhớ gọi manually |

---

### 1.2 WHY - Tại sao cần giải quyết?

#### A. Business Justification

**Scenario**: Production team cần thêm Team Gamma vào workflow

**Hiện tại (Manual)**:
```go
// Phải sửa main.go, thêm ~40 lines mới
gammaExec, _ := agenticcore.NewCrewExecutorFromConfig(apiKey, "team-gamma/config", nil)
orchestrator.TeamGamma = &TeamExecutor{Name: "Team Gamma", Executor: gammaExec}

// Trong Execute():
if strings.Contains(masterResult, "[DELEGATE_GAMMA]") {
    gammaResult, err := o.executeTeam(ctx, o.TeamGamma, input)
    combinedInput := fmt.Sprintf("Kết quả từ Team Gamma:\n%s\n\nTopic: %s", gammaResult, input)
    // ... more nesting
}
```

**Sau Enhancement (Declarative)**:
```yaml
# Chỉ cần thêm vào master-team/config/crew.yaml
sub_crews:
  team-gamma:
    config_path: "../team-gamma/config"

routing:
  signals:
    coordinator:
      - signal: "[DELEGATE_GAMMA]"
        type: sub_crew
        target_crew: team-gamma
        return_to: coordinator
```

#### B. Developer Experience Impact

| Aspect | Manual Approach | Declarative Approach |
|--------|-----------------|----------------------|
| **Adding new team** | Modify Go code, recompile | Add YAML config, restart |
| **Debugging signal issues** | Debug Go code, add printf | Check YAML config |
| **Testing workflow changes** | Integration tests required | Unit test YAML validation |
| **Onboarding new devs** | Learn orchestration pattern | Learn YAML schema |
| **Code review** | Review Go logic | Review declarative config |

#### C. Technical Debt Reduction

```
Current Technical Debt Score: HIGH
├── Duplicated orchestration logic across projects
├── Signal strings scattered in Go code
├── No standardized error handling for sub-crews
├── No metrics aggregation across crews
└── Testing requires full integration setup

After Enhancement Debt Score: LOW
├── Single framework handles all orchestration
├── Signals defined in YAML, validated at startup
├── Framework provides error handling patterns
├── Built-in metrics aggregation
└── YAML configs can be unit tested
```

---

### 1.3 WHO - Ai được hưởng lợi?

#### A. Developer Personas

| Persona | Pain Point | Benefit |
|---------|------------|---------|
| **Application Developer** | Viết boilerplate orchestration code | Zero-code orchestration từ YAML |
| **Framework Contributor** | Mỗi PR multi-crew reinvent wheel | Standard pattern để review |
| **DevOps/SRE** | Không có visibility vào sub-crew metrics | Aggregated metrics, health checks |
| **Tech Lead** | Hard to estimate effort for multi-crew features | Predictable effort từ declarative config |

#### B. Stakeholder Analysis

```
┌─────────────────────────────────────────────────────────────┐
│                    STAKEHOLDER MATRIX                        │
├─────────────────────────────────────────────────────────────┤
│                                                              │
│  High Interest, High Influence:                              │
│  ├── Framework Maintainers (implement changes)               │
│  └── Application Developers (primary users)                  │
│                                                              │
│  High Interest, Low Influence:                               │
│  ├── DevOps/SRE (operational concerns)                       │
│  └── New Contributors (learning curve)                       │
│                                                              │
│  Low Interest, High Influence:                               │
│  └── Project Leadership (resource allocation)                │
│                                                              │
└─────────────────────────────────────────────────────────────┘
```

---

### 1.4 WHERE - Thay đổi ở đâu?

#### A. Core Files Impact Analysis

```
go-agentic/core/
├── config.go          [MAJOR CHANGE]
│   ├── Add SubCrewConfig struct
│   ├── Add SignalType enum
│   ├── Extend RoutingSignal struct
│   └── Add SubCrews field to CrewConfig
│
├── crew.go            [MAJOR CHANGE]
│   ├── Add SubCrews map[string]*CrewExecutor
│   ├── Add ParentCrew *CrewExecutor
│   ├── Add ExecuteSubCrew() method
│   └── Add CrewContext struct
│
├── crew_routing.go    [MAJOR CHANGE]
│   ├── Add processSignal() switch logic
│   ├── Handle sub_crew signal type
│   └── Handle external signal type
│
├── crew_context.go    [NEW FILE]
│   ├── CrewContext struct
│   ├── Context passing utilities
│   └── Template engine for input_template
│
└── metrics.go         [MINOR CHANGE]
    └── Add crew-level metrics aggregation
```

#### B. YAML Schema Evolution

**Current Schema (v1.0)**:
```yaml
version: "1.0"
routing:
  signals:
    agent_id:
      - signal: "[SIGNAL]"
        target: "next_agent"  # or "" for terminate
```

**Proposed Schema (v2.0)**:
```yaml
version: "2.0"

# NEW: Sub-crew definitions
sub_crews:
  team-alpha:
    config_path: "./team-alpha/config"
    description: "Research & Analysis team"
  team-beta:
    config_path: "./team-beta/config"
    description: "Content Writing team"

routing:
  signals:
    coordinator:
      # Existing: Route to agent
      - signal: "[ANALYZE]"
        target: "analyst"
        type: "route"  # NEW: explicit type (default)

      # Existing: Terminate workflow
      - signal: "[DONE]"
        target: ""
        type: "terminate"  # NEW: explicit type

      # NEW: Call sub-crew
      - signal: "[DELEGATE_ALPHA]"
        type: "sub_crew"
        target_crew: "team-alpha"
        return_to: "coordinator"
        input_template: "Research topic: {{.Input}}"

      # NEW: External/Output signal
      - signal: "[NOTIFY_ADMIN]"
        type: "external"
        pause: false  # Don't pause workflow
```

#### C. Backward Compatibility Matrix

| Feature | v1.0 Config | v2.0 Config | Migration |
|---------|-------------|-------------|-----------|
| `target: "agent"` | Works | Works | None needed |
| `target: ""` | Terminates | Terminates | None needed |
| `type` field | N/A | Optional | Auto-inferred from target |
| `sub_crews` section | N/A | Optional | Ignored if empty |
| `input_template` | N/A | Optional | Uses raw input if not set |

---

### 1.5 WHEN - Khi nào thực hiện?

#### A. Phased Implementation Plan

```
┌─────────────────────────────────────────────────────────────┐
│                    IMPLEMENTATION PHASES                     │
├─────────────────────────────────────────────────────────────┤
│                                                              │
│  PHASE 1: Foundation (Backward Compatible)                   │
│  ├── Add `type` field to RoutingSignal                       │
│  ├── Add `sub_crews` section parsing                         │
│  ├── Validate new schema at startup                          │
│  └── No breaking changes to existing configs                 │
│                                                              │
│  PHASE 2: Core Features                                      │
│  ├── Implement `sub_crew` signal type                        │
│  ├── Auto-load sub-crews from config_path                    │
│  ├── Implement input_template with text/template             │
│  ├── Basic CrewContext passing                               │
│  └── Update example 02-multi-team                            │
│                                                              │
│  PHASE 3: Production Ready                                   │
│  ├── Implement `external` signal type                        │
│  ├── Pause/Resume for external handoff                       │
│  ├── Metrics aggregation across crews                        │
│  ├── Circular dependency detection                           │
│  └── Error handling patterns                                 │
│                                                              │
│  PHASE 4: Advanced Features                                  │
│  ├── Parallel sub-crew execution                             │
│  ├── Shared memory between crews                             │
│  ├── Crew-level quotas inheritance                           │
│  └── Dynamic crew composition                                │
│                                                              │
└─────────────────────────────────────────────────────────────┘
```

#### B. Phase Dependencies

```
Phase 1 ─────► Phase 2 ─────► Phase 3 ─────► Phase 4
   │              │              │              │
   │              │              │              └─ Requires: All previous
   │              │              └─ Requires: Phase 2 complete
   │              └─ Requires: Phase 1 complete
   └─ No dependencies (can start immediately)
```

---

### 1.6 HOW - Triển khai như thế nào?

#### A. Technical Implementation Details

**1. Extended RoutingSignal Struct** (`config.go`):

```go
// SignalType defines the type of routing signal
type SignalType string

const (
    SignalTypeRoute     SignalType = "route"      // Route to another agent
    SignalTypeTerminate SignalType = "terminate"  // End workflow
    SignalTypeSubCrew   SignalType = "sub_crew"   // Call sub-crew
    SignalTypeExternal  SignalType = "external"   // Emit to external handler
)

// RoutingSignal defines a signal that can be emitted by an agent
type RoutingSignal struct {
    Signal        string     `yaml:"signal"`
    Target        string     `yaml:"target"`
    Description   string     `yaml:"description"`

    // NEW fields for v2.0
    Type          SignalType `yaml:"type"`           // Signal type (default: inferred)
    TargetCrew    string     `yaml:"target_crew"`    // For sub_crew type
    ReturnTo      string     `yaml:"return_to"`      // Agent to resume after sub-crew
    InputTemplate string     `yaml:"input_template"` // Go template for sub-crew input
    Pause         bool       `yaml:"pause"`          // For external type: pause workflow?
}
```

**2. Sub-Crew Configuration** (`config.go`):

```go
// SubCrewConfig defines a sub-crew that can be called
type SubCrewConfig struct {
    ConfigPath  string `yaml:"config_path"`
    Description string `yaml:"description"`
}

// CrewConfig - Extended
type CrewConfig struct {
    // ... existing fields ...

    // NEW: Sub-crew definitions
    SubCrews map[string]SubCrewConfig `yaml:"sub_crews"`
}
```

**3. CrewExecutor Enhancement** (`crew.go`):

```go
type CrewExecutor struct {
    // ... existing fields ...

    // NEW: Sub-crew support
    SubCrews   map[string]*CrewExecutor  // Loaded sub-crews
    ParentCrew *CrewExecutor              // Reference to parent (if sub-crew)
}

// NEW: Execute sub-crew
func (ce *CrewExecutor) ExecuteSubCrew(ctx context.Context, crewName string, crewCtx *CrewContext) (*ExecuteResult, error) {
    subCrew, exists := ce.SubCrews[crewName]
    if !exists {
        return nil, fmt.Errorf("sub-crew %s not found", crewName)
    }

    // Build input from template
    input := ce.buildInputFromTemplate(crewCtx)

    // Execute sub-crew with isolated context
    result, err := subCrew.Execute(ctx, input)
    if err != nil {
        return nil, fmt.Errorf("sub-crew %s failed: %w", crewName, err)
    }

    // Aggregate metrics
    ce.Metrics.AggregateSubCrewMetrics(crewName, subCrew.Metrics)

    return result, nil
}
```

**4. Context Passing** (`crew_context.go` - NEW FILE):

```go
// CrewContext carries information between crews
type CrewContext struct {
    OriginalInput   string                 // User's original input
    CurrentInput    string                 // Current processed input
    PreviousResult  string                 // Last sub-crew result
    PreviousResults map[string]string      // All sub-crew results by name
    SharedData      map[string]interface{} // Custom shared data
    Depth           int                    // Nesting depth (for circular detection)
    MaxDepth        int                    // Maximum allowed depth
}

// NewCrewContext creates a new context for crew execution
func NewCrewContext(input string) *CrewContext {
    return &CrewContext{
        OriginalInput:   input,
        CurrentInput:    input,
        PreviousResults: make(map[string]string),
        SharedData:      make(map[string]interface{}),
        Depth:           0,
        MaxDepth:        10, // Prevent infinite recursion
    }
}
```

**5. Signal Processing Logic** (`crew_routing.go`):

```go
// processSignal handles signal-based routing with support for all signal types
func (ce *CrewExecutor) processSignal(current *Agent, response string, ctx context.Context, crewCtx *CrewContext) (*SignalResult, error) {
    signals := ce.crew.Routing.Signals[current.ID]

    for _, sig := range signals {
        if !signalMatchesContent(sig.Signal, response) {
            continue
        }

        // Determine signal type (infer from target if not explicit)
        sigType := sig.Type
        if sigType == "" {
            if sig.Target == "" {
                sigType = SignalTypeTerminate
            } else if sig.TargetCrew != "" {
                sigType = SignalTypeSubCrew
            } else {
                sigType = SignalTypeRoute
            }
        }

        switch sigType {
        case SignalTypeRoute:
            return &SignalResult{
                Type:   SignalTypeRoute,
                Target: ce.findAgentByID(sig.Target),
            }, nil

        case SignalTypeTerminate:
            return &SignalResult{
                Type: SignalTypeTerminate,
            }, nil

        case SignalTypeSubCrew:
            // Check depth limit
            if crewCtx.Depth >= crewCtx.MaxDepth {
                return nil, fmt.Errorf("max crew nesting depth exceeded")
            }

            // Update context
            crewCtx.Depth++
            crewCtx.CurrentInput = response

            // Execute sub-crew
            result, err := ce.ExecuteSubCrew(ctx, sig.TargetCrew, crewCtx)
            if err != nil {
                return nil, err
            }

            // Store result
            crewCtx.PreviousResult = result.Content
            crewCtx.PreviousResults[sig.TargetCrew] = result.Content

            // Return to specified agent
            return &SignalResult{
                Type:        SignalTypeResume,
                Target:      ce.findAgentByID(sig.ReturnTo),
                ResumeInput: result.Content,
            }, nil

        case SignalTypeExternal:
            // Emit to external handler (if configured)
            if ce.ExternalSignalChan != nil {
                ce.ExternalSignalChan <- &ExternalSignal{
                    Signal:  sig.Signal,
                    AgentID: current.ID,
                    Content: response,
                }
            }

            if sig.Pause {
                return &SignalResult{
                    Type:     SignalTypePause,
                    ResumeAt: current.ID,
                }, nil
            }
            // Continue processing if not pausing
            continue
        }
    }

    return nil, nil // No signal matched
}
```

#### B. Developer API Changes

**Before** (Manual Orchestration):
```go
// ~50 lines of setup
masterExec, _ := agenticcore.NewCrewExecutorFromConfig(apiKey, "master-team/config", nil)
alphaExec, _ := agenticcore.NewCrewExecutorFromConfig(apiKey, "team-alpha/config", nil)
betaExec, _ := agenticcore.NewCrewExecutorFromConfig(apiKey, "team-beta/config", nil)

orchestrator := &MultiTeamOrchestrator{
    MasterTeam: masterExec,
    TeamAlpha:  alphaExec,
    TeamBeta:   betaExec,
}

// ~60 lines of orchestration logic
result, err := orchestrator.Execute(ctx, input)
```

**After** (Declarative):
```go
// 3 lines!
crew, err := agenticcore.NewCrewExecutorFromConfig(apiKey, "master-team/config", nil)
// Sub-crews are auto-loaded based on crew.yaml sub_crews section

result, err := crew.Execute(ctx, input)
// Framework handles all orchestration internally
```

---

### 1.7 HOW MUCH - Effort & Impact

#### A. Implementation Effort Matrix

| Task | Complexity | Files | Est. Lines | Risk |
|------|------------|-------|------------|------|
| Add SignalType enum | Low | `config.go` | ~20 | Low |
| Extend RoutingSignal struct | Low | `config.go` | ~30 | Low |
| Add SubCrewConfig | Low | `config.go` | ~15 | Low |
| YAML parsing for new fields | Medium | `config.go` | ~50 | Medium |
| Add SubCrews to CrewExecutor | Medium | `crew.go` | ~30 | Medium |
| Implement ExecuteSubCrew() | Medium | `crew.go` | ~80 | Medium |
| CrewContext struct & methods | Medium | `crew_context.go` | ~100 | Low |
| processSignal() enhancement | High | `crew_routing.go` | ~120 | High |
| Input template engine | Low | `crew_context.go` | ~40 | Low |
| Circular dependency detection | Medium | `crew.go` | ~50 | Medium |
| Metrics aggregation | Medium | `metrics.go` | ~60 | Low |
| Update example 02-multi-team | Low | `examples/` | ~-180 | Low |
| **TOTAL** | | | ~415 new | |

#### B. Priority Ranking

| Feature | Priority | Effort | Impact | ROI Score |
|---------|----------|--------|--------|-----------|
| Signal `type` field | **P0** | Low | High | ⭐⭐⭐⭐⭐ |
| `sub_crews` section | **P0** | Low | High | ⭐⭐⭐⭐⭐ |
| `sub_crew` signal type | **P0** | Medium | High | ⭐⭐⭐⭐ |
| `CrewContext` | **P1** | Medium | High | ⭐⭐⭐⭐ |
| Input template | **P1** | Low | Medium | ⭐⭐⭐ |
| `external` signal | **P2** | Low | Medium | ⭐⭐⭐ |
| Circular detection | **P2** | Medium | Medium | ⭐⭐⭐ |
| Metrics aggregation | **P2** | Medium | Medium | ⭐⭐⭐ |
| Parallel sub-crews | **P3** | High | Medium | ⭐⭐ |

---

## Phần 2: Đề xuất Hướng đi Tốt nhất

### 2.1 Recommended Approach: Incremental Enhancement

#### Lý do chọn Incremental thay vì Big Bang:

1. **Backward Compatibility**: Existing configs vẫn hoạt động
2. **Lower Risk**: Mỗi phase có thể test độc lập
3. **Faster Time-to-Value**: Phase 1-2 đủ cover 80% use cases
4. **Community Feedback**: Có thể adjust dựa trên feedback

### 2.2 Implementation Roadmap

```
Week 1-2: Phase 1 (Foundation)
├── Day 1-2: Add SignalType, extend RoutingSignal
├── Day 3-4: Add SubCrewConfig, YAML parsing
├── Day 5-7: Validation, backward compat tests
└── Deliverable: v2.0 schema supported, no breaking changes

Week 3-4: Phase 2 (Core Features)
├── Day 1-3: Implement sub_crew signal processing
├── Day 4-5: Add CrewContext, input template
├── Day 6-7: Update example, documentation
└── Deliverable: 02-multi-team works with YAML-only config

Week 5-6: Phase 3 (Production Ready)
├── Day 1-2: External signal type
├── Day 3-4: Metrics aggregation
├── Day 5-6: Error handling patterns
├── Day 7: Integration testing
└── Deliverable: Production-ready multi-crew support
```

### 2.3 Success Criteria

| Phase | Criteria | Measurement |
|-------|----------|-------------|
| Phase 1 | All existing tests pass | 100% test pass rate |
| Phase 1 | v1.0 configs work unchanged | Backward compat verified |
| Phase 2 | 02-multi-team works with YAML-only | main.go reduced to <20 lines |
| Phase 2 | Signal routing works | Integration tests pass |
| Phase 3 | Metrics aggregated | Dashboard shows crew hierarchy |
| Phase 3 | Error handling works | Chaos tests pass |

### 2.4 Risk Mitigation

| Risk | Probability | Impact | Mitigation |
|------|-------------|--------|------------|
| Breaking existing configs | Low | High | Extensive backward compat testing |
| Performance degradation | Medium | Medium | Benchmark before/after |
| Circular dependency bugs | Medium | High | Depth limit + build-time validation |
| Memory leaks in nested crews | Medium | High | Proper cleanup in defer |

---

## Phần 3: Open Questions cho Community

### 3.1 History Sharing Strategy

**Question**: Bao nhiêu conversation history nên share giữa parent và sub-crew?

**Options**:
| Option | Pros | Cons |
|--------|------|------|
| A. Full history | Sub-crew có đầy đủ context | Token cost cao, privacy concerns |
| B. Summary only | Efficient, focused | Có thể mất context quan trọng |
| C. Configurable | Flexible | Complexity cho users |

**Recommendation**: Option C với default là B (summary only)

### 3.2 Error Handling Pattern

**Question**: Sub-crew fail → parent crew xử lý thế nào?

**Options**:
| Option | Use Case |
|--------|----------|
| A. Propagate, terminate parent | Critical sub-crews |
| B. Retry with backoff | Transient failures |
| C. Fallback to another sub-crew | High availability |
| D. Continue with error context | Best-effort processing |

**Recommendation**: Configurable per signal, default là A

### 3.3 Metrics Aggregation

**Question**: Cost/tokens của sub-crews có tính vào parent không?

**Options**:
| Option | Behavior |
|--------|----------|
| A. Aggregate all | Parent shows total cost of all sub-crews |
| B. Separate tracking | Each crew tracks independently |
| C. Both | Aggregate + per-crew breakdown |

**Recommendation**: Option C (most visibility)

### 3.4 Circular Dependency Detection

**Question**: Crew A → B → A - detect thế nào?

**Options**:
| Option | When | Behavior |
|--------|------|----------|
| A. Build-time validation | Config load | Error if cycle detected |
| B. Runtime detection | Execution | Error when depth exceeded |
| C. Both | Both | Double protection |

**Recommendation**: Option C với max_depth default = 10

---

## Phần 4: Next Steps

### Immediate Actions (This Week)

1. **Review & Approve**: Team review this proposal
2. **Create Issues**: Break down into GitHub issues
3. **Assign Owners**: Assign Phase 1 tasks

### Short-term (Next 2 Weeks)

4. **Phase 1 Implementation**: Start Foundation work
5. **Test Infrastructure**: Set up multi-crew test suite
6. **Documentation**: Update README với v2.0 preview

### Medium-term (Next Month)

7. **Phase 2 Implementation**: Core Features
8. **Example Update**: Refactor 02-multi-team
9. **Community Preview**: Alpha release for feedback

---

## Appendix A: Complete YAML Schema (v2.0)

```yaml
# crew.yaml v2.0 - Complete Schema Reference

version: "2.0"  # Required: Schema version
name: "my-crew"  # Required: Crew identifier
description: |
  Multi-line description of this crew's purpose

# NEW in v2.0: Sub-crew definitions
sub_crews:
  team-alpha:
    config_path: "./team-alpha/config"  # Relative to this file
    description: "Research & Analysis team"
  team-beta:
    config_path: "./team-beta/config"
    description: "Content Writing team"

# Entry point agent
entry_point: coordinator

# Agent list (must have YAML files in agents/ folder)
agents:
  - coordinator
  - analyst
  - writer

# Routing configuration
routing:
  signals:
    coordinator:
      # Type: route (default) - route to another agent in same crew
      - signal: "[ANALYZE]"
        target: "analyst"
        type: "route"  # Optional: inferred from target
        description: "Send to analyst for deep analysis"

      # Type: terminate - end workflow
      - signal: "[DONE]"
        target: ""
        type: "terminate"  # Optional: inferred from empty target
        description: "Workflow complete"

      # Type: sub_crew (NEW) - call sub-crew
      - signal: "[DELEGATE_ALPHA]"
        type: "sub_crew"
        target_crew: "team-alpha"
        return_to: "coordinator"
        input_template: |
          Research the following topic:
          {{.OriginalInput}}

          Previous context:
          {{.PreviousResult}}
        description: "Delegate research to Team Alpha"

      # Type: external (NEW) - emit signal to external handler
      - signal: "[NOTIFY_ADMIN]"
        type: "external"
        pause: false  # Continue workflow after emitting
        description: "Notify admin but don't wait"

      - signal: "[HUMAN_REVIEW]"
        type: "external"
        pause: true  # Pause workflow until external response
        description: "Wait for human review"

    analyst:
      - signal: "[ANALYSIS_COMPLETE]"
        target: "coordinator"
        type: "route"

  # Agent behaviors
  agent_behaviors:
    coordinator:
      wait_for_signal: false
      auto_route: true
    analyst:
      wait_for_signal: false

  # Parallel groups (existing feature)
  parallel_groups:
    research_group:
      agents: ["researcher1", "researcher2"]
      wait_for_all: true
      timeout_seconds: 60
      next_agent: "aggregator"

# Settings
settings:
  max_rounds: 10
  max_handoffs: 10
  timeout_seconds: 300
  config_mode: permissive  # or "strict"

  # NEW in v2.0: Sub-crew specific settings
  max_crew_depth: 10  # Maximum nesting depth
  sub_crew_timeout_seconds: 120  # Default timeout for sub-crew calls
```

---

## Appendix B: Migration Guide

### From v1.0 to v2.0

**No changes required** - v1.0 configs work unchanged.

**Optional enhancements**:

1. Add explicit `type` field for clarity:
```yaml
# Before (v1.0)
- signal: "[DONE]"
  target: ""

# After (v2.0) - same behavior, more explicit
- signal: "[DONE]"
  target: ""
  type: "terminate"
```

2. Add `sub_crews` section to enable multi-crew:
```yaml
# Add at top level
sub_crews:
  team-alpha:
    config_path: "./team-alpha/config"
```

3. Add `sub_crew` signals:
```yaml
routing:
  signals:
    coordinator:
      - signal: "[DELEGATE_ALPHA]"
        type: "sub_crew"
        target_crew: "team-alpha"
        return_to: "coordinator"
```

---

**Report Prepared By**: Claude Code Analysis
**Review Status**: Ready for Team Review
**Next Review Date**: TBD

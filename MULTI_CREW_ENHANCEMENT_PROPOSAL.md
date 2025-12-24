# Multi-Crew Enhancement Proposal

> Đề xuất cải tiến go-agentic framework để hỗ trợ multi-crew architecture

**Author**: Claude Code
**Date**: 2024-12-24
**Status**: Draft
**Related Example**: `examples/02-multi-team`

---

## 1. Executive Summary

Qua quá trình xây dựng ví dụ multi-team coordination, phát hiện framework hiện tại được thiết kế cho **single-crew workflows**. Để hỗ trợ các use cases phức tạp hơn (hierarchical teams, crew composition, inter-crew communication), cần bổ sung một số tính năng mới.

---

## 2. Current Limitations

### 2.1. Signal Target Limitation

**Vấn đề**: `target: ""` chỉ có nghĩa là TERMINATE workflow.

```yaml
# Hiện tại
signals:
  coordinator:
    - signal: "[DELEGATE_ALPHA]"
      target: ""  # ← Luôn terminate, không có cách emit "output signal"
```

**Impact**: Không thể có agent emit signal cho external handler mà không dừng workflow.

### 2.2. No Inter-Crew Communication

**Vấn đề**: Mỗi CrewExecutor độc lập, không có cách để:
- Crew A gọi Crew B như sub-workflow
- Truyền context/history giữa các crews
- Định nghĩa crew hierarchy trong YAML

**Impact**: Phải viết orchestrator code thủ công cho mọi multi-crew scenario.

### 2.3. No Pause/Resume for External Handoff

**Vấn đề**: `wait_for_signal: true` chỉ đợi user input, không hỗ trợ:
- Pause workflow, gọi external crew
- Resume với kết quả từ external crew

### 2.4. Signal Priority Fixed

**Vấn đề**: Termination signals được check trước routing signals.

```go
// crew_routing.go - line ~110
for _, sig := range signals {
    if sig.Target == "" {  // Termination check FIRST
        if signalMatchesContent(sig.Signal, response) {
            return TERMINATE
        }
    }
}
```

**Impact**: Nếu response chứa nhiều signals, thứ tự check cố định có thể gây unexpected behavior.

### 2.5. No Crew Composition

**Vấn đề**: Không có cách định nghĩa crew-of-crews trong YAML.

---

## 3. Proposed Solutions

### 3.1. New Signal Types

Thêm field `type` cho signals:

```yaml
# Proposed crew.yaml v2.0
routing:
  signals:
    coordinator:
      # Type 1: Route to agent (existing)
      - signal: "[ANALYZE]"
        target: "analyst"
        type: "route"  # default

      # Type 2: Terminate workflow (existing, explicit)
      - signal: "[DONE]"
        target: ""
        type: "terminate"

      # Type 3: NEW - External/Output signal (không terminate)
      - signal: "[DELEGATE_ALPHA]"
        type: "external"  # Emit signal, workflow continues/pauses
        pause: true       # Optional: pause waiting for external result

      # Type 4: NEW - Call sub-crew
      - signal: "[RESEARCH]"
        type: "sub_crew"
        target_crew: "team-alpha"
        return_to: "coordinator"  # Resume here after sub-crew completes
```

### 3.2. Sub-Crew Definition

Cho phép định nghĩa sub-crews trong crew.yaml:

```yaml
# crew.yaml v2.0
version: "2.0"
name: master-crew

# NEW: Sub-crew definitions
sub_crews:
  team-alpha:
    config_path: "./team-alpha/config"
    description: "Research & Analysis team"

  team-beta:
    config_path: "./team-beta/config"
    description: "Content Writing team"

entry_point: coordinator

agents:
  - coordinator

routing:
  signals:
    coordinator:
      - signal: "[DELEGATE_ALPHA]"
        type: "sub_crew"
        target_crew: "team-alpha"
        input_template: "Research topic: {{.Input}}"
        return_to: "coordinator"

      - signal: "[DELEGATE_BETA]"
        type: "sub_crew"
        target_crew: "team-beta"
        input_template: "Write about: {{.PreviousResult}}"
        return_to: "coordinator"
```

### 3.3. CrewExecutor Enhancement

```go
// core/crew.go - Proposed additions

type CrewExecutor struct {
    // ... existing fields ...

    // NEW: Sub-crew support
    SubCrews    map[string]*CrewExecutor  // Loaded sub-crews
    ParentCrew  *CrewExecutor              // Reference to parent (if sub-crew)

    // NEW: External signal handling
    ExternalSignalChan chan *ExternalSignal  // For external handlers
}

// NEW: External signal structure
type ExternalSignal struct {
    Signal      string
    AgentID     string
    Content     string
    RequiresResponse bool
}

// NEW: Method to execute sub-crew
func (ce *CrewExecutor) ExecuteSubCrew(ctx context.Context, crewName string, input string) (*ExecuteResult, error) {
    subCrew, exists := ce.SubCrews[crewName]
    if !exists {
        return nil, fmt.Errorf("sub-crew %s not found", crewName)
    }

    // Execute sub-crew with isolated history
    result, err := subCrew.Execute(ctx, input)
    if err != nil {
        return nil, err
    }

    // Optionally merge relevant history back
    return result, nil
}
```

### 3.4. New Routing Logic

```go
// core/crew_routing.go - Proposed changes

type SignalType string

const (
    SignalTypeRoute     SignalType = "route"
    SignalTypeTerminate SignalType = "terminate"
    SignalTypeExternal  SignalType = "external"
    SignalTypeSubCrew   SignalType = "sub_crew"
)

type RoutingSignal struct {
    Signal       string     `yaml:"signal"`
    Target       string     `yaml:"target"`
    Type         SignalType `yaml:"type"`         // NEW
    TargetCrew   string     `yaml:"target_crew"`  // NEW: for sub_crew type
    ReturnTo     string     `yaml:"return_to"`    // NEW: resume agent
    Pause        bool       `yaml:"pause"`        // NEW: for external type
    InputTemplate string    `yaml:"input_template"` // NEW: template for sub-crew input
}

// Enhanced signal processing
func (ce *CrewExecutor) processSignal(current *Agent, response string) (*SignalResult, error) {
    signals := ce.crew.Routing.Signals[current.ID]

    for _, sig := range signals {
        if !signalMatchesContent(sig.Signal, response) {
            continue
        }

        switch sig.Type {
        case SignalTypeRoute:
            return &SignalResult{
                Type:   SignalTypeRoute,
                Target: ce.findAgentByID(sig.Target),
            }, nil

        case SignalTypeTerminate:
            return &SignalResult{
                Type: SignalTypeTerminate,
            }, nil

        case SignalTypeExternal:
            // Emit to external handler
            if ce.ExternalSignalChan != nil {
                ce.ExternalSignalChan <- &ExternalSignal{
                    Signal:  sig.Signal,
                    AgentID: current.ID,
                    Content: response,
                    RequiresResponse: sig.Pause,
                }
            }
            if sig.Pause {
                return &SignalResult{
                    Type:     SignalTypePause,
                    ResumeAt: current.ID,
                }, nil
            }
            continue // Don't terminate, continue processing

        case SignalTypeSubCrew:
            // Execute sub-crew
            result, err := ce.ExecuteSubCrew(ctx, sig.TargetCrew, input)
            if err != nil {
                return nil, err
            }
            // Resume at specified agent with sub-crew result
            return &SignalResult{
                Type:      SignalTypeResume,
                ResumeAt:  sig.ReturnTo,
                ResumeInput: result.Content,
            }, nil
        }
    }

    return nil, nil // No signal matched
}
```

### 3.5. Context Passing Between Crews

```go
// NEW: Crew context for passing between crews
type CrewContext struct {
    OriginalInput   string
    PreviousResults map[string]string  // crew_name -> result
    SharedMemory    map[string]interface{}
    ParentHistory   []Message  // Optional: relevant history from parent
}

func (ce *CrewExecutor) ExecuteWithContext(ctx context.Context, input string, crewCtx *CrewContext) (*ExecuteResult, error) {
    // Inject context into system prompt or initial message
    enhancedInput := ce.buildContextualInput(input, crewCtx)
    return ce.Execute(ctx, enhancedInput)
}
```

---

## 4. Migration Path

### Phase 1: Backward Compatible Additions
- Add `type` field to signals (default: infer from `target`)
- Add `sub_crews` section (optional)
- No breaking changes to existing configs

### Phase 2: Enhanced Features
- Implement `external` signal type
- Implement `sub_crew` signal type
- Add `CrewContext` support

### Phase 3: Advanced Features
- Parallel sub-crew execution
- Shared memory between crews
- Crew-level quotas and metrics aggregation

---

## 5. Example: Multi-Team with New Features

```yaml
# examples/02-multi-team/config/crew.yaml (v2.0)
version: "2.0"
name: multi-team-orchestrator

sub_crews:
  team-alpha:
    config_path: "./team-alpha/config"
  team-beta:
    config_path: "./team-beta/config"

entry_point: coordinator

agents:
  - coordinator

routing:
  signals:
    coordinator:
      - signal: "[RESEARCH]"
        type: "sub_crew"
        target_crew: "team-alpha"
        return_to: "coordinator"

      - signal: "[WRITE]"
        type: "sub_crew"
        target_crew: "team-beta"
        input_template: |
          Dựa trên kết quả nghiên cứu sau:
          {{.PreviousResult}}

          Hãy viết content về: {{.OriginalInput}}
        return_to: "coordinator"

      - signal: "[FINAL]"
        type: "terminate"
```

---

## 6. Implementation Priority

| Feature | Priority | Effort | Impact |
|---------|----------|--------|--------|
| Signal `type` field | High | Low | High |
| `sub_crew` signal type | High | Medium | High |
| `external` signal type | Medium | Low | Medium |
| `CrewContext` | Medium | Medium | High |
| Parallel sub-crews | Low | High | Medium |

---

## 7. Open Questions

1. **History sharing**: Bao nhiêu history nên share giữa parent và sub-crew?
2. **Error handling**: Sub-crew fail thì parent crew xử lý thế nào?
3. **Metrics aggregation**: Cost/tokens của sub-crews có tính vào parent không?
4. **Circular dependencies**: Crew A gọi B, B gọi A - detect và prevent?

---

## 8. References

- Current implementation: `core/crew.go`, `core/crew_routing.go`
- Example exposing limitations: `examples/02-multi-team/`
- Similar patterns: CrewAI's hierarchical crews, LangGraph's sub-graphs

---

## Appendix A: Current Workaround

Với framework hiện tại, multi-crew phải implement thủ công:

```go
// examples/02-multi-team/cmd/main.go
type MultiTeamOrchestrator struct {
    MasterTeam *CrewExecutor
    TeamAlpha  *CrewExecutor
    TeamBeta   *CrewExecutor
}

func (o *MultiTeamOrchestrator) Execute(ctx context.Context, input string) (string, error) {
    // 1. Call master team
    result1, _ := o.MasterTeam.Execute(ctx, input)

    // 2. Parse signals manually
    if strings.Contains(result1.Content, "[DELEGATE_ALPHA]") {
        result2, _ := o.TeamAlpha.Execute(ctx, input)
        // 3. Feed back to master...
    }
    // ... manual orchestration logic
}
```

Với đề xuất mới, tất cả logic này sẽ được handle bởi framework thông qua YAML config.

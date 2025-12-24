# Multi-Team Coordination Demo

Demo điều phối nhiều teams độc lập với go-agentic framework.

## Cấu trúc

```
02-multi-team/
├── cmd/main.go              # Orchestrator điều phối các teams
├── master-team/             # Master Team (Coordinator)
│   └── config/
│       ├── crew.yaml
│       └── agents/
│           └── coordinator.yaml
├── team-alpha/              # Team Alpha (Research & Analysis)
│   └── config/
│       ├── crew.yaml
│       └── agents/
│           ├── researcher.yaml
│           └── analyst.yaml
└── team-beta/               # Team Beta (Writing & Review)
    └── config/
        ├── crew.yaml
        └── agents/
            ├── writer.yaml
            └── reviewer.yaml
```

## Workflow

```
┌─────────────────────────────────────────────────────────────┐
│                      ORCHESTRATOR                           │
│  (cmd/main.go - điều phối các teams)                       │
└─────────────────────────────────────────────────────────────┘
                           │
           ┌───────────────┼───────────────┐
           ▼               ▼               ▼
   ┌───────────────┐ ┌───────────────┐ ┌───────────────┐
   │  MASTER TEAM  │ │  TEAM ALPHA   │ │  TEAM BETA    │
   │  (1 agent)    │ │  (2 agents)   │ │  (2 agents)   │
   │               │ │               │ │               │
   │  Coordinator  │ │  Researcher   │ │  Writer       │
   │               │ │      ↓        │ │      ↓        │
   │               │ │  Analyst      │ │  Reviewer     │
   └───────────────┘ └───────────────┘ └───────────────┘
```

**Workflow chi tiết:**

1. **User** nhập topic
2. **Orchestrator** gửi cho Master Team
3. **Master Team (Coordinator)** phân tích → emit `[DELEGATE_ALPHA]`
4. **Orchestrator** gọi Team Alpha
5. **Team Alpha**: Researcher → `[RESEARCH_DONE]` → Analyst → `[ALPHA_COMPLETE]`
6. **Orchestrator** gửi kết quả về Master Team
7. **Master Team** emit `[DELEGATE_BETA]`
8. **Orchestrator** gọi Team Beta
9. **Team Beta**: Writer → `[DRAFT_DONE]` → Reviewer → `[BETA_COMPLETE]`
10. **Orchestrator** gửi kết quả về Master Team
11. **Master Team** emit `[FINAL_REPORT]`
12. **Orchestrator** trả kết quả cho User

## Cách chạy

```bash
cd examples/02-multi-team

# Chạy demo
make run

# Chạy với verbose
make run-verbose

# Build binary
make build
./multi-team
```

## Yêu cầu

- Go 1.23+
- Ollama với model `qwen3:1.7b`

```bash
ollama pull qwen3:1.7b
ollama serve
```

## Điểm khác biệt với single-crew

| Aspect | Single Crew | Multi-Team |
|--------|-------------|------------|
| Cấu trúc | 1 crew.yaml | Mỗi team có crew.yaml riêng |
| Routing | Signal-based trong 1 crew | Orchestrator điều phối giữa crews |
| History | Shared | Mỗi team có history riêng |
| Isolation | Tất cả agents chung context | Mỗi team isolated |
| Scalability | Limited | Có thể thêm teams dễ dàng |

## Mở rộng

Để thêm team mới:

1. Tạo thư mục `team-xyz/config/`
2. Tạo `crew.yaml` và `agents/*.yaml`
3. Load trong `main.go`:
   ```go
   xyzExec, _ := agenticcore.NewCrewExecutorFromConfig(apiKey, "team-xyz/config", nil)
   ```
4. Thêm logic điều phối trong `Execute()`

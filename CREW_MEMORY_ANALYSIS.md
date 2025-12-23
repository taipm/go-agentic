# 🧠 Crew & Agent Memory Analysis

**Status:** Analysis Complete
**Date:** Dec 23, 2025

---

## 📋 Câu Hỏi Của Bạn

**Kịch bản:**
1. Bạn nói: "Tôi tên Tài"
2. Sau đó hỏi: "Tôi tên gì?"
3. **Câu hỏi:** Crew sẽ trả lời thế nào? Crew có thể nhớ được rằng tên bạn là "Tài" không?

---

## ✅ Câu Trả Lời: CÓ, CREW CÓ THỂ NHỚ!

Crew **CÓ** khả năng nhớ thông tin từ cuộc trò chuyện trước đó. Đây là cách nó hoạt động:

### Cơ Chế Hoạt Động

```go
// Trong CrewExecutor (core/crew.go)
type CrewExecutor struct {
    crew          *Crew
    apiKey        string
    history       []Message      // ✅ LƯU TOÀN BỘ CUỘC TRỌ CHUYỆN
    entryAgent    *Agent
}
```

**`history` là một mảng `[]Message` mà:**
1. ✅ **LƯU TOÀN BỘ CUỘC TRỌ CHUYỆN** từ đầu đến cuối
2. ✅ **ĐƯỢC TRUYỀN ĐẢY ĐỦ cho mỗi agent** khi thực thi
3. ✅ **ĐƯỢC CẬP NHẬT SAU MỖI LẦN** agent hoặc người dùng nói gì

---

## 🔄 Quy Trình Chi Tiết

### Bước 1: Người Dùng Nói "Tôi Tên Tài"

```go
// Execute() function - line 715
func (ce *CrewExecutor) Execute(ctx context.Context, input string) (*CrewResponse, error) {
    // Bước 1: Thêm input vào history
    ce.history = append(ce.history, Message{
        Role:    "user",
        Content: input,  // "Tôi tên Tài"
    })

    // history bây giờ là:
    // [
    //   {Role: "user", Content: "Tôi tên Tài"}
    // ]
}
```

### Bước 2: Agent Xử Lý & Trả Lời

```go
// core/agent.go - line 22
func ExecuteAgent(ctx context.Context, agent *Agent, input string, history []Message, apiKey string) (*AgentResponse, error) {
    // Agent NHẬN history từ crew
    // history = [{Role: "user", Content: "Tôi tên Tài"}]

    // Agent GỬI history + input ĐẾN LLM
    messages := convertToProviderMessages(history)
    // LLM nhận context: "User said: Tôi tên Tài"

    // LLM trả lời: "Rất vui được biết bạn tên Tài!"
}
```

### Bước 3: Response Được Thêm Vào History

```go
// core/crew.go - line 578-580
// Add agent response to history
ce.history = append(ce.history, Message{
    Role:    "assistant",
    Content: response.Content,  // "Rất vui được biết bạn tên Tài!"
})

// history bây giờ là:
// [
//   {Role: "user", Content: "Tôi tên Tài"},
//   {Role: "assistant", Content: "Rất vui được biết bạn tên Tài!"}
// ]
```

### Bước 4: Người Dùng Hỏi "Tôi Tên Gì?"

```go
// Lần 2: User input
ce.history = append(ce.history, Message{
    Role:    "user",
    Content: "Tôi tên gì?",  // ← CÂUHỎI MỚI
})

// history bây giờ là:
// [
//   {Role: "user", Content: "Tôi tên Tài"},
//   {Role: "assistant", Content: "Rất vui được biết bạn tên Tài!"},
//   {Role: "user", Content: "Tôi tên gì?"}  // ← ĐƯỢC THÊM VÀO
// ]
```

### Bước 5: Agent Nhận History ĐẦYCỦ & Trả Lời

```go
// Agent lại được gọi với FULL history
ExecuteAgent(ctx, agent, "Tôi tên gì?", history, apiKey)

// history mà agent nhận:
// [
//   {Role: "user", Content: "Tôi tên Tài"},
//   {Role: "assistant", Content: "Rất vui được biết bạn tên Tài!"},
//   {Role: "user", Content: "Tôi tên gì?"}
// ]

// LLM đọc TOÀN BỘ cuộc trọ chuyện này
// LLM thấy: "User said: Tôi tên Tài"
// LLM thấy: "Earlier I said: Rất vui được biết bạn tên Tài!"
// LLM thấy: "User now asks: Tôi tên gì?"

// LLM trả lời: "Tên bạn là Tài! Bạn vừa mới nói lúc nãy mà."
```

---

## 📊 Cấu Trúc Message

```go
type Message struct {
    Role    string  // "user", "assistant", "system"
    Content string  // Nội dung tin nhắn
}
```

**Ví dụ History Thực Tế:**

```json
[
  {
    "role": "user",
    "content": "Tôi tên Tài"
  },
  {
    "role": "assistant",
    "content": "Rất vui được biết bạn tên Tài!"
  },
  {
    "role": "user",
    "content": "Tôi tên gì?"
  },
  {
    "role": "assistant",
    "content": "Tên bạn là Tài"
  },
  {
    "role": "user",
    "content": "Tôi có bao nhiêu tuổi?"
  },
  {
    "role": "assistant",
    "content": "Tôi không có thông tin về tuổi của bạn"
  }
]
```

---

## 🎯 Key Points About Crew Memory

| Tính Năng | Mô Tả |
|-----------|-------|
| **Trí Nhớ** | ✅ Crew LƯU TOÀN BỘ cuộc trọ chuyện trong `history []Message` |
| **Phạm Vi** | ✅ Từ **đầu lần Execute** cho đến **cuối cùng** |
| **Truyền Cho Agent** | ✅ Agent nhận **TOÀN BỘ history** mỗi khi được gọi |
| **Có Giới Hạn?** | ✅ **CÓ** - Token context của LLM (thường 128k, 200k...) |
| **Context Window** | ✅ Được tracking trong `AgentMemoryMetrics.CurrentContextSize` |
| **Persistent?** | ❌ **KHÔNG** - History chỉ lưu trong RAM, reboot thì mất |

---

## ⚠️ Giới Hạn Context Window

### Problem

```
Nếu history quá dài (hàng chục ngàn tokens), LLM sẽ lỗi!

Ví dụ:
- OpenAI GPT-4: 8k/32k/128k tokens tùy version
- Ollama Mistral: 32k tokens
- Nếu history > 32k tokens → LLM báo lỗi "context exceeded"
```

### Current Solution (WEEK 2)

```go
// Type AgentMemoryMetrics tracks:
type AgentMemoryMetrics struct {
    CurrentContextSize int  // Token hiện tại trong history
    MaxContextWindow   int  // Max tokens (default: 32000)
    ContextTrimPercent float64 // Trim % nếu vượt (20%)
}
```

**Cơ chế trim context:**
- Nếu `CurrentContextSize > MaxContextWindow`
- Xoá 20% cuộc trọ chuyện cũ nhất
- Giữ lại 80% gần nhất

---

## 💭 Memory Architecture

```
┌─────────────────────────────────────────────────┐
│           CrewExecutor                          │
│  ┌───────────────────────────────────────────┐  │
│  │ history []Message                         │  │
│  │ ┌─────────────────────────────────────┐  │  │
│  │ │ 1. User: "Tôi tên Tài"              │  │  │
│  │ │ 2. Assistant: "Vui lòng biết..."    │  │  │
│  │ │ 3. User: "Tôi tên gì?"              │  │  │
│  │ │ 4. Assistant: "Tên bạn là Tài"      │  │  │
│  │ │ 5. User: "Làm nào vậy?"             │  │  │
│  │ │ ...                                 │  │  │
│  │ └─────────────────────────────────────┘  │  │
│  └───────────────────────────────────────────┘  │
│                      │                          │
│                      ├─→ ExecuteAgent(          │
│                      │     agent,              │
│                      │     input,              │
│                      │     history← FULL       │
│                      │   )                     │
│                      │                         │
│                      ├─→ LLM nhận FULL        │
│                      │   cuộc trọ chuyện      │
│                      │                         │
│                      └─→ LLM trả lời dựa      │
│                          trên context         │
└─────────────────────────────────────────────────┘
```

---

## 🔍 Kiểm Tra Code Thực Tế

### Execute() - How History Is Managed

```go
// core/crew.go - line 713-800
func (ce *CrewExecutor) Execute(ctx context.Context, input string) (*CrewResponse, error) {
    // BƯỚC 1: Thêm user input vào history
    ce.history = append(ce.history, Message{
        Role:    "user",
        Content: input,  // "Tôi tên Tài" hoặc "Tôi tên gì?"
    })

    // BƯỚC 2: Loop cho đến khi có terminal response
    for round := 0; round < ce.crew.MaxRounds; round++ {
        // Lấy agent hiện tại (first agent)
        currentAgent := ce.entryAgent

        // BƯỚC 3: TRUYỀN FULL HISTORY cho agent
        response, err := ExecuteAgent(
            ctx,
            currentAgent,
            input,
            ce.history,  // ← TOÀN BỘ lịch sử cuộc trọ chuyện
            ce.apiKey,
        )

        // BƯỚC 4: Thêm response của agent vào history
        ce.history = append(ce.history, Message{
            Role:    "assistant",
            Content: response.Content,
        })

        // BƯỚC 5: Nếu có tool calls, execute tools & thêm results vào history
        if len(response.ToolCalls) > 0 {
            for _, toolCall := range response.ToolCalls {
                // Execute tool
                toolResult := safeExecuteTool(ctx, tool, toolCall.Arguments)

                // Thêm tool result vào history
                ce.history = append(ce.history, Message{
                    Role:    "tool",
                    Content: toolResult,
                })
            }
        }
    }

    return &CrewResponse{...}, nil
}
```

---

## 📝 Ví Dụ Cụ Thể: "Tôi Tên Tài" → "Tôi Tên Gì?"

### Cuộc Trò Chuyện #1

```
USER: "Tôi tên Tài"

// CrewExecutor history:
[
  {Role: "user", Content: "Tôi tên Tài"}
]

// Agent receives:
history = [{Role: "user", Content: "Tôi tên Tài"}]

// LLM response:
"Rất vui được biết bạn tên Tài! Tôi là một AI assistant. Tên bạn là Tài, phải không?"

// History updated:
[
  {Role: "user", Content: "Tôi tên Tài"},
  {Role: "assistant", Content: "Rất vui được biết bạn tên Tài! ..."}
]
```

### Cuộc Trò Chuyện #2

```
USER: "Tôi tên gì?"

// CrewExecutor history (TOÀN BỘ):
[
  {Role: "user", Content: "Tôi tên Tài"},
  {Role: "assistant", Content: "Rất vui được biết bạn tên Tài! ..."},
  {Role: "user", Content: "Tôi tên gì?"}  ← ĐƯỢC THÊM VÀO
]

// Agent receives:
history = [
  {Role: "user", Content: "Tôi tên Tài"},
  {Role: "assistant", Content: "Rất vui được biết bạn tên Tài! ..."},
  {Role: "user", Content: "Tôi tên gì?"}
]

// LLM đọc FULL context:
// - User nói: "Tôi tên Tài"
// - Assistant đã nói: "Rất vui được biết bạn tên Tài"
// - User bây giờ hỏi: "Tôi tên gì?"
// → LLM suy luận: "User đã nói tên là Tài"

// LLM response:
"Bạn đã nói rồi mà - tên bạn là Tài!"
```

---

## ⚙️ How It's Implemented

### Storage

```go
// core/crew.go - line 396
type CrewExecutor struct {
    history []Message  // ← Lưu toàn bộ cuộc trọ chuyện
}
```

### Truyền Cho Agent

```go
// core/agent.go - line 22
func ExecuteAgent(
    ctx context.Context,
    agent *Agent,
    input string,
    history []Message,  // ← Agent nhận history
    apiKey string,
) (*AgentResponse, error) {
    // Chuyển history thành định dạng cho LLM
    messages := convertToProviderMessages(history)

    // Gửi history + input đến LLM
    request := &providers.CompletionRequest{
        Messages: messages,  // ← TOÀN BỘ HISTORY
        ...
    }
}
```

### Cập Nhật History

```go
// core/crew.go - line 578
ce.history = append(ce.history, Message{
    Role:    "assistant",
    Content: response.Content,  // ← Thêm response vào
})
```

---

## 🎁 Features Related to Memory

| Feature | Location | WEEK |
|---------|----------|------|
| **History Tracking** | crew.go | Built-in |
| **Context Window Tracking** | AgentMemoryMetrics | WEEK 2/3 |
| **Context Trim Logic** | (TODO) | Future |
| **Memory Metrics** | memory_performance.go | WEEK 3 |
| **Token Counting** | (Estimated) | WEEK 1/2 |
| **Conversation Limits** | (Configurable) | Future |

---

## 🚀 Kết Luận

### ✅ CÓ THỂ NHỚ

```
USER: "Tôi tên Tài"
AGENT: "OK, bạn tên Tài"

USER: "Tôi tên gì?"
AGENT: "Bạn tên Tài - bạn vừa nói lúc nãy"
```

**TẠI SAO:**
- Crew lưu **TOÀN BỘ** cuộc trọ chuyện trong `history []Message`
- Agent nhận **FULL history** mỗi khi thực thi
- LLM đọc **TOÀN BỘ context** khi sinh response

### ⚠️ CÓ GIỚI HẠN

```
Nếu cuộc trọ chuyện quá dài (vượt token limit):
- OpenAI GPT-4: 8k/32k/128k tokens
- LLM sẽ lỗi hoặc response chất lượng thấp
- Giải pháp: Trim context cũ hoặc summarize
```

### 💾 KHÔNG PERSISTENT

```
Nếu restart crew executor:
- History sẽ bị reset thành []Message{}
- Cuộc trọ chuyện trước bị quên
- Giải pháp: Lưu history vào database nếu cần
```

---

## 📚 Related Files

- **core/crew.go** - CrewExecutor & history management (line 396, 518, 578, 715...)
- **core/agent.go** - ExecuteAgent receives history (line 22)
- **core/types.go** - Message struct (line 168)
- **core/memory_performance.go** - Context window tracking (WEEK 3)

---

**Generated:** Dec 23, 2025
**Status:** ✅ Complete Analysis

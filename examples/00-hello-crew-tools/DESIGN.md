# hello-crew-tools: LLM Tool Capability Validation

## 🎯 Objective

Create a new example (`hello-crew-tools`) to validate whether the memory problem in `hello-crew` is due to:
1. **Architecture limitation** (no persistence) - OR
2. **LLM limitation** (model ignores memory instructions)

By providing tools for conversation analysis, we can test if the LLM can use structured data to count messages and remember facts.

---

## 📋 Current Problem Analysis

From hello-crew logs:
```
Query 1: "Tôi tên gì?" → "Phan Minh Tài" ✓
Query 2: "Tôi là Lê Văn Phương Trang" → Updated ✓
Query 3: "Tôi tên gì?" → "Lê Văn Phương Trang" ✓
Query 4: "Tôi đã hỏi mấy câu?" → "start fresh" ✗ CANNOT COUNT
```

**Hypothesis:** The LLM doesn't understand "how many messages" because:
- Without tools: LLM must count from raw text = unreliable
- With tools: LLM can call GetMessageCount() = should be accurate

---

## 🛠️ Tool Design

### Tool 1: GetMessageCount()
```go
Name: "get_message_count"
Description: "Returns the total number of messages in the conversation history"
Parameters: None
Returns: {
    "count": 4,
    "role_breakdown": {
        "user": 2,
        "assistant": 2
    }
}
```

**Test:**
- User: "Tôi đã hỏi mấy câu?"
- Expected: Agent calls GetMessageCount() → "Bạn đã hỏi 2 câu"

---

### Tool 2: GetConversationSummary()
```go
Name: "get_conversation_summary"
Description: "Returns a summary of the conversation including key facts"
Parameters: None
Returns: {
    "total_messages": 4,
    "messages": [
        {"role": "user", "content": "Tôi tên gì?", "index": 0},
        {"role": "assistant", "content": "Phan Minh Tài", "index": 1},
        ...
    ],
    "extracted_facts": {
        "user_name": "Lê Văn Phương Trang",
        "key_topics": ["personal_info"]
    }
}
```

**Test:**
- User: "Tôi là ai?"
- Expected: Agent calls GetConversationSummary() → "Bạn là Lê Văn Phương Trang"

---

### Tool 3: SearchMessages()
```go
Name: "search_messages"
Description: "Search for specific keywords or patterns in conversation history"
Parameters: {
    "query": string (required) - What to search for
    "limit": int (optional, default: 10) - Max results
}
Returns: {
    "query": "tên",
    "results": [
        {
            "index": 0,
            "role": "user",
            "content": "Tôi tên gì?",
            "relevance": 0.95
        }
    ]
}
```

**Test:**
- User: "Bạn nhớ tôi nói gì lần đầu?"
- Expected: Agent calls SearchMessages() → "Bạn hỏi 'Tôi tên gì?'"

---

### Tool 4: CountMessagesBy()
```go
Name: "count_messages_by"
Description: "Count messages filtered by role, content keywords, or time range"
Parameters: {
    "filter_by": string - "role" | "keyword" | "all"
    "filter_value": string - "user" | keyword | null
}
Returns: {
    "filter": {"by": "role", "value": "user"},
    "count": 2,
    "details": {...}
}
```

**Test:**
- User: "Tôi đã nói bao nhiêu lần?"
- Expected: Agent calls CountMessagesBy(role=user) → "Bạn đã nói 2 lần"

---

## 📁 Directory Structure

```
examples/00-hello-crew-tools/
├── cmd/
│   └── main.go                    # Entry point
├── config/
│   ├── crew.yaml                  # Crew configuration with tools enabled
│   ├── agents/
│   │   └── hello-agent-tools.yaml # Agent config with tool definitions
│   └── tools/
│       ├── message_counter.yaml   # Tool definitions for LLM
│       └── conversation_tools.yaml
├── internal/
│   ├── tools.go                   # Tool implementations
│   ├── message_analyzer.go        # Message analysis logic
│   └── tool_registry.go           # Tool registration
├── Makefile                        # Build commands
├── README.md                       # Documentation
├── go.mod
└── go.sum
```

---

## 🔧 Implementation Plan

### Phase 1: Tool Implementation (Go Backend)
```go
// internal/tools.go

type MessageAnalyzerTools struct {
    executor *CrewExecutor  // Access to history
}

func (mat *MessageAnalyzerTools) GetMessageCount() map[string]interface{} {
    history := mat.executor.GetHistory()
    return map[string]interface{}{
        "count": len(history),
        "role_breakdown": map[string]int{
            "user": countByRole(history, "user"),
            "assistant": countByRole(history, "assistant"),
        },
    }
}

func (mat *MessageAnalyzerTools) GetConversationSummary() map[string]interface{} {
    history := mat.executor.GetHistory()
    return map[string]interface{}{
        "total_messages": len(history),
        "messages": history,
        "extracted_facts": extractFacts(history),
    }
}
```

### Phase 2: Agent Configuration
```yaml
# config/agents/hello-agent-tools.yaml
id: hello-agent-tools
name: Hello Agent with Tools
tools:
  - name: get_message_count
    description: "Count total messages in conversation"
    parameters: {}
  - name: get_conversation_summary
    description: "Get conversation summary with facts"
    parameters: {}
  - name: search_messages
    description: "Search messages by keyword"
    parameters:
      query: { type: "string" }
      limit: { type: "integer", default: 10 }
  - name: count_messages_by
    description: "Count messages by filter"
    parameters:
      filter_by: { type: "string", enum: ["role", "keyword"] }
      filter_value: { type: "string" }
```

### Phase 3: Main Entry Point
```go
// cmd/main.go - Similar to hello-crew but with tool registration
func main() {
    executor, err := createExecutor(apiKey)

    // Register tools
    tools := internal.NewMessageAnalyzerTools(executor)
    executor.RegisterTools(tools)

    runCLI(executor)
}
```

### Phase 4: Testing Scenarios
```go
// Test cases in internal/tools_test.go

func TestGetMessageCount(t *testing.T) {
    // Simulate conversation:
    // User: "Tôi tên gì?"
    // Agent: "Phan Minh Tài"
    // User: "Tôi là Lê Văn Phương Trang"
    // Agent: "Ok, your name is..."

    count := tools.GetMessageCount()
    assert.Equal(t, 4, count["count"])
}

func TestAgentCallsTool(t *testing.T) {
    // Ask: "Tôi đã hỏi mấy câu?"
    // Expected: Agent calls get_message_count() tool
    // Expected: Agent returns "Bạn đã hỏi 2 câu"
}
```

---

## 📊 Expected Outcomes

### If Tools Improve Memory Behavior ✅
- Agent accurately counts: "Bạn đã hỏi 2 câu"
- Agent retrieves facts: "Tên bạn là Lê Văn Phương Trang"
- **Conclusion:** Memory problem = Architecture (lack of structured data)
- **Next Step:** Implement Simple Path with fact extraction

### If Tools Don't Help ❌
- Agent still says: "Tôi không biết"
- Agent ignores tool results
- **Conclusion:** Memory problem = LLM limitation (model can't follow instructions)
- **Next Step:** Switch to better models or add explicit system prompts

---

## 🧪 Validation Checklist

- [ ] Tools are correctly registered with LLM
- [ ] Tool definitions match what LLM expects
- [ ] Agent successfully calls tools during conversation
- [ ] Tool results are accurate (message counts, fact extraction)
- [ ] Agent incorporates tool results into responses
- [ ] System prompt includes tool usage instructions
- [ ] Error handling for tool execution failures
- [ ] Tool results are logged for debugging

---

## 📈 Success Metrics

| Metric | Target | Validation |
|--------|--------|-----------|
| Tool Call Rate | 80%+ | Agent calls tool when asked about history |
| Answer Accuracy | 100% | Tool results match actual history |
| Response Quality | Improves | Agent uses tool data in response |
| User Satisfaction | High | Can now reference conversation history |

---

## 🚀 Next Steps After Validation

1. **If tools work:**
   - Implement hello-crew-persistence (Phase 1: Simple Path)
   - Add automatic fact extraction
   - Build session management layer

2. **If tools fail:**
   - Test with better models (Claude, GPT-4, etc.)
   - Add more explicit system prompts
   - Consider alternative approaches (no-instruction memory)

3. **Regardless:**
   - Document findings in memory architecture
   - Use insights to guide Simple Path implementation
   - Plan Phase 2+ enhancements

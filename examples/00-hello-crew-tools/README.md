# hello-crew-tools: LLM Tool Capability Validation

## 🚀 Quick Start

### Setup

```bash
# Copy .env.example to .env and add your OpenAI API key
cp .env.example .env
# Edit .env and paste your API key from https://platform.openai.com/api-keys

# Build and run
make build
make run
```

### Test Tool Execution

```bash
./test_time.sh
```

---

## 🎯 Purpose

This example validates whether the memory problem in `hello-crew` is due to:
1. **Architecture limitation** (no persistence) - OR
2. **LLM limitation** (model ignores memory instructions)

By providing tools for conversation analysis, we test if the LLM can use structured data to accurately count messages and remember facts.

---

## 📊 Problem Context

From `hello-crew` testing, we observed:

```
Query 1: "Tôi tên gì?"
→ "Phan Minh Tài" ✓ Correct

Query 2: "Tôi là Lê Văn Phương Trang"
→ Updated name ✓

Query 3: "Tôi tên gì?"
→ "Lê Văn Phương Trang" ✓ Remembers update

Query 4: "Tôi đã hỏi mấy câu?"
→ "start fresh" ✗ CANNOT COUNT
```

**Why Query 4 failed:**
- Without tools: Agent must parse raw conversation text to count = unreliable
- With tools: Agent calls GetMessageCount() function = accurate

---

## 🛠️ Available Tools

### 1. `get_message_count()`
**Purpose:** Returns total number of messages in conversation

**Usage:**
```
User: "Tôi đã hỏi mấy câu?"
→ Agent calls: get_message_count()
→ Tool returns: {"count": 4, "role_breakdown": {"user": 2, "assistant": 2}}
→ Agent responds: "Bạn đã hỏi 2 câu"
```

**Expected Improvement:**
- Before: "Tôi không biết" ❌
- After: "Bạn đã hỏi 2 câu" ✅

---

### 2. `get_conversation_summary()`
**Purpose:** Returns all messages and extracted facts

**Usage:**
```
User: "Tôi là ai?"
→ Agent calls: get_conversation_summary()
→ Tool returns: {
    "total_messages": 4,
    "messages": [...],
    "extracted_facts": {"user_name": "Lê Văn Phương Trang", ...}
}
→ Agent responds: "Bạn là Lê Văn Phương Trang"
```

---

### 3. `search_messages(query, limit)`
**Purpose:** Search for keywords in conversation

**Usage:**
```
User: "Bạn nhớ tôi nói gì lần đầu?"
→ Agent calls: search_messages(query="tên")
→ Tool returns: results with matching messages
→ Agent responds: "Bạn hỏi 'Tôi tên gì?'"
```

---

### 4. `count_messages_by(filter_by, filter_value)`
**Purpose:** Count messages filtered by role or keyword

**Usage:**
```
User: "Tôi đã nói bao nhiêu lần?"
→ Agent calls: count_messages_by(filter_by="role", filter_value="user")
→ Tool returns: {"count": 2}
→ Agent responds: "Bạn đã nói 2 lần"
```

---

## 🚀 Quick Start

### Prerequisites
- Go 1.20+
- Ollama running locally with `deepseek-r1:1.5b` or `gemma3:1b`

### Step 1: Start Ollama
```bash
ollama run deepseek-r1:1.5b
```

### Step 2: Run the Example
```bash
cd examples/00-hello-crew-tools
make run
```

### Step 3: Test Conversations

Try these conversations in order:

**Conversation 1: Basic Greeting**
```
> Xin chào!
Agent: Hello! How can I help you today?

> Tôi là John Doe
Agent: Got it, your name is John Doe! 😊
```

**Conversation 2: Memory Test (Name Recall)**
```
> Tôi tên gì?
Agent: Your name is John Doe

Expected: ✅ Agent remembers the name
```

**Conversation 3: Tool Test (Message Counting)**
```
> Tôi đã hỏi mấy câu?
Expected: Agent calls get_message_count()
Expected Result:
  - With tools working: "Bạn đã hỏi 3 câu"  ✅
  - If tools fail: "I don't know" or random answer ❌
```

**Conversation 4: Tool Test (Message Search)**
```
> Bạn nhớ tôi nói gì lần đầu?
Expected: Agent calls search_messages()
Expected Result: "Bạn nói 'Xin chào!'"
```

**Conversation 5: Tool Test (Fact Extraction)**
```
> Bạn biết tôi là ai?
Expected: Agent calls get_conversation_summary()
Expected Result: "Bạn là John Doe"
```

---

## 📈 Success Criteria

### If Tools Improve Memory ✅
```
Metric                  | Before    | After
------------------------|-----------|----------
Can count messages      | NO ❌     | YES ✅
Can recall facts        | NO ❌     | YES ✅
Can search history      | NO ❌     | YES ✅
Overall accuracy        | 0%        | 95%+
```

**Conclusion:** Memory problem = **Architecture** (lack of structure)
**Next Step:** Implement Simple Path with persistence

---

### If Tools Don't Help ❌
```
Metric                  | Expected  | Actual
------------------------|-----------|----------
Agent calls tools       | 100%      | < 50%
Tool results used       | 100%      | < 20%
Answer accuracy         | 95%+      | < 30%
Model follows prompts   | YES       | NO
```

**Conclusion:** Memory problem = **LLM limitation** (model can't follow instructions)
**Next Step:** Switch to better models (Claude, GPT-4) or use different strategy

---

## 🔍 What's Different from hello-crew?

| Aspect | hello-crew | hello-crew-tools |
|--------|-----------|------------------|
| **Tools** | None | 4 conversation analysis tools |
| **Data Access** | Raw history only | Structured tool results |
| **Message Counting** | No | Yes (via tool) |
| **Fact Extraction** | No | Yes (via tool) |
| **Search Capability** | No | Yes (via tool) |
| **Expected Accuracy** | Low | High (if tools work) |

---

## 🧪 Testing Scenarios

### Scenario A: Tools Are Effective
```
Conversation Flow:
1. "Tôi tên gì?" → Agent stores name
2. "Tôi là John" → Agent updates name
3. "Tôi tên gì?" → Agent recalls "John" ✓
4. "Tôi đã hỏi mấy câu?" → Agent uses get_message_count() tool → "3" ✓
5. "Bạn nhớ lần đầu tôi nói gì?" → Agent uses search_messages() → "Tôi tên gì?" ✓

Observation: Tools enable accurate conversation understanding
```

### Scenario B: Tools Don't Work
```
Conversation Flow:
1-3. Same as above (name works)
4. "Tôi đã hỏi mấy câu?" → Agent ignores tool → "I don't know" ✗
5. "Bạn nhớ lần đầu tôi nói gì?" → Agent guesses → Random answer ✗

Observation: LLM doesn't call or use tools properly
```

---

## 📊 Logging & Debugging

The example logs tool execution details:

```
[TOOL CALL] Tool: get_message_count
[TOOL RESULT] {"count": 4, "role_breakdown": {"user": 2, "assistant": 2}}
[AGENT RESPONSE] Based on that data...
```

You can see:
- Which tools the agent called
- What data was returned
- How the agent used the results

---

## 🔧 Server Mode (API Testing)

Run in server mode for API testing:

```bash
make server
```

This starts the server on `http://localhost:8082`

### Endpoints

**POST /execute** - Execute with agent
```bash
curl -X POST http://localhost:8082/execute \
  -H "Content-Type: application/json" \
  -d '{"input": "Tôi đã hỏi mấy câu?"}'
```

**POST /execute-tool** - Execute a specific tool directly
```bash
curl -X POST http://localhost:8082/execute-tool \
  -H "Content-Type: application/json" \
  -d '{
    "tool": "get_message_count",
    "params": {}
  }'
```

**GET /tools** - List available tools
```bash
curl http://localhost:8082/tools
```

**GET /health** - Health check
```bash
curl http://localhost:8082/health
```

---

## 🎓 What This Teaches Us

### If Tools Work (Likely Outcome)
1. **LLM can use tools** when properly configured
2. **Architecture matters** - structured data > raw text
3. **Next step:** Implement persistent storage + tool-based retrieval
4. **Simple Path is feasible** - tools can provide structure

### If Tools Don't Work (Unlikely)
1. **Model limitation** - Ollama is too weak
2. **System prompt issue** - Tool instructions aren't clear enough
3. **API issue** - Tool calling mechanism isn't working
4. **Next step:** Debug with Claude/GPT-4 or fix tool definition

---

## 🚦 Troubleshooting

### Agent doesn't call tools
**Check:**
1. Tool definitions in `hello-agent-tools.yaml` are valid
2. System prompt includes tool usage instructions
3. Ollama model supports function calling (deepseek-r1, gemma3 should work)

### Tools return empty results
**Check:**
1. Conversation history is being accumulated
2. Tool parameters are correctly passed
3. Message content is not empty

### Ollama connection fails
**Check:**
1. Ollama is running: `ollama serve`
2. Model is downloaded: `ollama pull deepseek-r1:1.5b`
3. Local port is 11434 (default)

---

## 📝 Files Structure

```
hello-crew-tools/
├── DESIGN.md                              # Detailed design document
├── README.md                              # This file
├── Makefile                               # Build commands
├── cmd/
│   └── main.go                            # Entry point with tool integration
├── config/
│   ├── crew.yaml                          # Crew configuration
│   └── agents/
│       └── hello-agent-tools.yaml         # Agent with tool definitions
├── internal/
│   └── tools.go                           # Tool implementations
├── go.mod
└── go.sum
```

---

## 🎯 Key Takeaways

1. **This example validates a hypothesis** about the root cause of memory problems
2. **Tools provide structure** that raw history lacks
3. **The outcome will guide** the architecture design for Simple Path
4. **Results will inform** whether to focus on persistence vs. LLM upgrade

---

## 📚 Related Examples

- `hello-crew` - Original example with memory issues
- `hello-crew-persistence` (coming) - File-based session storage
- `hello-crew-semantic` (coming) - Vector embeddings + semantic search

---

## 🤝 Contributing

To add more tools:

1. Add tool definition to `hello-agent-tools.yaml`
2. Implement tool method in `internal/tools.go`
3. Add test case
4. Update this README

---

## 📞 Questions?

Check DESIGN.md for architectural details and implementation notes.

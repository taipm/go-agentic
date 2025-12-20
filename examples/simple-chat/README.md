# Simple Chat Example - Ví Dụ Thảo Luận Đơn Giản

This is the **simplest example** of using go-agentic with **2 agents** that automatically have a conversation with each other **in Vietnamese**.

Đây là ví dụ **đơn giản nhất** để sử dụng go-agentic với **2 agent** tự động nói chuyện với nhau **bằng tiếng Việt**.

## ✨ Features / Đặc Điểm

- 🤖 **2 Simple Agents (2 Agent Đơn Giản)**
  - Người Tò Mò (Enthusiast) - asks questions
  - Chuyên Gia (Expert) - provides answers
- 💬 **Automatic Vietnamese Conversation (Nói Chuyện Tiếng Việt Tự Động)**
- 📋 **YAML Configuration (Cấu Hình YAML)** - Easy to customize
- 🎯 **Minimal Setup (Thiết Lập Tối Thiểu)** - No tools needed
- 📚 **Easy to Understand (Dễ Hiểu)** - Perfect for learning
- 🔄 **Phase 3: Declarative Routing DSL** - Automatic agent routing based on trigger detection

## 🏗️ Architecture with Phase 3 Routing

```text
Initial Topic / Chủ Đề Ban Đầu
        ↓
┌─────────────────────────┐
│  Người Tò Mò            │
│  (Enthusiast)           │
│  Đặt câu hỏi thâm sâu   │
└──────────┬──────────────┘
           ↓
    [TriggerDetector]
    Detect: "?", "hỏi", "gì", "như thế nào"
    Matches response? → YES
           ↓
┌─────────────────────────┐
│  Chuyên Gia             │
│  (Expert)               │
│  Cung cấp câu trả lời   │
│  (isTerminal: true)     │
└──────────┬──────────────┘
           ↓
    Response Ready
```

**How Routing Works:**

1. Enthusiast asks a question (containing "?", "hỏi", etc.)
2. TriggerDetector (Phase 3) detects question keywords
3. Router automatically routes to Expert
4. Expert provides final answer (isTerminal = true)
5. Conversation ends

## 🎭 Agents / Các Agent

### 1. Người Tò Mò (Enthusiast)

- **Role / Vai Trò**: Curious learner who asks insightful questions
- **Behavior / Hành Động**: Explores ideas in Vietnamese, engages in meaningful discussion
- **IsTerminal**: `false` (can pass to next agent)
- **Temperature**: 0.8 (more creative and varied responses)

### 2. Chuyên Gia (Expert)

- **Role / Vai Trò**: Subject matter expert with deep knowledge
- **Behavior / Hành Động**: Provides comprehensive answers in Vietnamese, shares expertise
- **IsTerminal**: `true` (final response ends the conversation)
- **Temperature**: 0.7 (balanced, consistent responses)

## 🔄 Phase 3: Declarative Routing DSL

This example demonstrates **Phase 3** of go-agentic's UX improvements: **Declarative Routing DSL** with automatic trigger detection.

### Routing Configuration

```go
// Build routing with Phase 3 declarative API
routingConfig, _ := agentic.NewRouter().
    RegisterAgents("enthusiast", "expert").              // Register valid agents
    FromAgent("enthusiast").                             // Start route from enthusiast
    To("expert",                                         // Route to expert
       agentic.NewKeywordDetector(                       // When response contains...
           []string{"?", "hỏi", "gì", "như thế nào"},  // Vietnamese question keywords
           false,                                        // Case-insensitive
       ),
    ).
    Done().                                              // Complete this route
    Build()                                              // Compile and validate
```

### How It Works

| Component | Role | Example |
| --- | --- | --- |
| **RouterBuilder** | Fluent API for routing | `NewRouter()` |
| **RegisterAgents()** | Validate agent IDs | Ensures "enthusiast" and "expert" exist |
| **FromAgent()** | Start route definition | Routes originating from "enthusiast" |
| **KeywordDetector** | Trigger detector (Phase 3) | Detects "?", "hỏi", etc. in responses |
| **To()** | Define target agent | Routes to "expert" when trigger matches |
| **Done()** | Complete this route | Returns to builder for more routes |
| **Build()** | Compile rules | Validates and creates RoutingConfig |

### Trigger Detectors Available (Phase 3)

```go
// Detect keywords in response
agentic.NewKeywordDetector([]string{"error", "bug"}, false)

// Detect using regex patterns
agentic.NewPatternDetector(`\[ERROR:\s*\d+\]`)

// Detect explicit signals [SIGNAL: name]
agentic.NewSignalDetector("resolved")

// Detect line prefixes
agentic.NewPrefixDetector([]string{"ACTION:", "INFO:"}, false)

// Combine detectors (OR logic)
agentic.NewAnyDetector(detector1, detector2, ...)

// Require all conditions (AND logic)
agentic.NewAllDetector(detector1, detector2, ...)

// Default route (always matches)
agentic.NewAlwaysDetector()

// Disabled route (never matches)
agentic.NewNeverDetector()
```

## 📋 YAML Configuration

The crew is configured using `crew.yaml` - easy to customize:

```yaml
crew:
  maxRounds: 4        # Maximum rounds of conversation
  maxHandoffs: 3      # Maximum handoffs between agents

agents:
  - id: "enthusiast"
    name: "Người Tò Mò"
    role: "Người học hỏi đầy tò mò"
    backstory: "..."
    model: "gpt-4o-mini"
    temperature: 0.8
    isTerminal: false

  - id: "expert"
    name: "Chuyên Gia"
    role: "Chuyên gia có kiến thức sâu"
    backstory: "..."
    model: "gpt-4o-mini"
    temperature: 0.7
    isTerminal: true

topics:
  - "Những thực hành tốt nhất khi viết code Go là gì?"
  - "Làm thế nào mà các AI agent có thể cải thiện phát triển phần mềm?"
  - "..."
```

## 🚀 Quick Start / Bắt Đầu Nhanh

### Step 1: Setup API Key / Thiết Lập API Key

```bash
cp .env.example .env
# Edit .env and add your OpenAI API key
# Chỉnh sửa .env và thêm OpenAI API key của bạn
```

### Step 2: Run / Chạy

```bash
go run main.go
```

### Expected Output / Kết Quả Mong Đợi

```text
🤖 Hệ Thống Thảo Luận Multi-Agent Đơn Giản
==================================================

📌 Chủ đề 1: Những thực hành tốt nhất khi viết code Go là gì?
--------------------------------------------------
[Người Tò Mò]: Khi viết code Go, các thực hành tốt nhất là gì?

[Chuyên Gia]: Có rất nhiều thực hành tốt nhất...

[Người Tò Mò]: Bạn có thể giải thích thêm về...

[Chuyên Gia]: Tất nhiên! Chi tiết hơn về...

✅ Kết Quả Cuối Cùng:
[Final comprehensive response in Vietnamese]
```

## 🔧 Customization / Tùy Chỉnh

### Modify Topics / Thay Đổi Chủ Đề

Edit `crew.yaml`:

```yaml
topics:
  - "Chủ đề của bạn ở đây"
  - "Một chủ đề khác"
  - "Và thêm nữa..."
```

### Change Agent Personality / Thay Đổi Tính Cách Agent

Edit `crew.yaml`:

```yaml
agents:
  - id: "expert"
    name: "Tên mới"
    role: "Vai trò mới"
    backstory: "Câu chuyện nền mới bằng tiếng Việt"
    temperature: 0.9  # Higher = more creative
```

### Adjust Conversation Length / Điều Chỉnh Độ Dài Cuộc Trò Chuyện

Edit `crew.yaml`:

```yaml
crew:
  maxRounds: 6      # More rounds = longer conversation
  maxHandoffs: 4    # More handoffs = more back-and-forth
```

### Use Different Model / Sử Dụng Model Khác

Edit `crew.yaml`:

```yaml
agents:
  - id: "expert"
    model: "gpt-4o"        # More capable
    # or
    model: "gpt-3.5-turbo" # Cheaper
```

## 📁 File Structure / Cấu Trúc File

```text
simple-chat/
├── main.go              # Application logic (Ngôn ngữ lập trình)
├── crew.yaml            # Configuration file (File cấu hình)
├── .env.example         # API key template (Mẫu API key)
├── go.mod & go.sum      # Dependencies (Phụ thuộc)
├── README.md            # Documentation (Tài liệu)
└── GETTING_STARTED.md   # Quick start guide (Hướng dẫn bắt đầu)
```

## 🔍 Understanding the Code

### main.go - Ultra-Minimal Design

The code is just **2 functions** totaling **58 lines**:

**1. main()** - The Core Application (Lines 12-47)

```go
// Load API key
apiKey := getEnvVar("OPENAI_API_KEY")

// Load team from YAML using library function
team, _ := agentic.LoadTeamFromYAML("team.yaml", agentic.ToolHandlerRegistry{})

// Create executor and run
executor := agentic.NewTeamExecutor(team, apiKey)

// Execute sample topics
for i, topic := range topics {
    resp, _ := executor.Execute(context.Background(), topic)
    // Print results
}
```

That's it! No custom structs, no helper functions needed.

**2. getEnvVar()** - Environment Helper (Lines 49-57)

- Reads `.env` file
- Extracts API key
- Simple string parsing

### How the Library Handles Everything

| What | Before | Now |
| --- | --- | --- |
| **YAML Loading** | `loadConfig()` (4 lines) | `agentic.LoadTeamFromYAML()` |
| **Agent Creation** | `buildTeam()` (30 lines) | `LoadTeamFromYAML()` handles it |
| **Routing Setup** | Manual Phase 3 DSL | `LoadTeamFromYAML()` builds it |
| **Config Struct** | Custom `Config` struct | Not needed |
| **Total Code** | 110 lines | 58 lines |

### The Philosophy

> **The library should do the work, not the user.**

Before: Users wrote `loadConfig()` and `buildTeam()`
Now: Users just call `LoadTeamFromYAML()` and run

### Key Takeaway

Users only need to:

1. Edit `team.yaml` - define agents, topics, routing
2. Set `OPENAI_API_KEY` in `.env`
3. Run `go run main.go`

The library handles everything else!

## 🇻🇳 Vietnamese Features / Đặc Điểm Tiếng Việt

All messages and prompts are in Vietnamese:

- Agent names: Người Tò Mò, Chuyên Gia
- Agent roles and backstories in Vietnamese
- Output messages in Vietnamese
- Conversation between agents in Vietnamese
- Error messages in Vietnamese

## ✅ Security / Bảo Mật

⚠️ **Never commit your actual API keys!**

- `.env` file is ignored by git
- Always use `.env.example` as template
- For more security guidelines, see `/SECURITY.md`

## 🚀 Next Steps / Bước Tiếp Theo

After understanding this simple example:

1. **Customize the crew**
   - Modify topics in `crew.yaml`
   - Change agent personalities
   - Adjust conversation parameters

2. **Add more agents**
   - Add additional agents in `crew.yaml`
   - Create more complex workflows

3. **Explore other examples**
   - `customer-service` - Real-world use case
   - `it-support` - IT help desk automation
   - `research-assistant` - Multi-step research

4. **Build your own**
   - Create custom YAML configurations
   - Design your own multi-agent systems
   - Add specialized tools and handlers

## 📚 Learning Resources / Tài Liệu Học

- **GETTING_STARTED.md** - Detailed setup and troubleshooting
- **go-agentic documentation** - Full API reference
- **crew.yaml** - Configuration examples
- **main.go** - Code implementation

## 🤔 FAQ / Câu Hỏi Thường Gặp

**Q: Why YAML instead of code?**
A: YAML configuration makes it easy to customize without recompiling code. Non-developers can modify topics and agent behavior.

**Q: Why Vietnamese?**
A: It demonstrates that go-agentic works with any language. Agents can converse in any language supported by the OpenAI models.

**Q: Can I add more agents?**
A: Yes! Just add more entries to the `agents` section in `crew.yaml`.

**Q: How do I make longer conversations?**
A: Increase `maxRounds` and `maxHandoffs` in `crew.yaml`.

## 🆘 Troubleshooting / Khắc Phục Sự Cố

### Problem: "OPENAI_API_KEY environment variable not set"

**Solution**: Create `.env` file with your API key

```bash
cp .env.example .env
# Edit .env and add your key
```

### Problem: "cannot read file crew.yaml"

**Solution**: Make sure `crew.yaml` is in the same directory as `main.go`

```bash
# Verify file exists
ls crew.yaml
```

### Problem: Agents speaking in English instead of Vietnamese

**Solution**: The agents' backstory instructs them to speak Vietnamese. If they're not:

- Check your `crew.yaml` has proper Vietnamese instructions
- Try rephrasing the topic in Vietnamese

## 📞 Support / Hỗ Trợ

- Read **GETTING_STARTED.md** for detailed setup
- Check `/SECURITY.md` for security best practices
- Review main go-agentic documentation
- Check error messages carefully

## 📝 Summary

This example demonstrates:

✅ Loading configuration from YAML
✅ Creating agents dynamically from config
✅ Multi-language support (Vietnamese)
✅ Easy customization without code changes
✅ Automatic agent-to-agent conversation
✅ Clean, understandable code structure
✅ **Phase 3: Declarative Routing DSL** with TriggerDetector
✅ Automatic routing based on keyword detection

---

**Ready to run?**

```bash
cp .env.example .env
# Add your API key to .env
go run main.go
```

Sẵn sàng chạy? 🚀

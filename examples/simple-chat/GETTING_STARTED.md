# Getting Started with Simple Chat Example / Bắt Đầu với Ví Dụ Chat Đơn Giản

This guide will help you run the simplest go-agentic example with **YAML configuration** and **Vietnamese agents** in just 5 minutes.

Hướng dẫn này sẽ giúp bạn chạy ví dụ go-agentic đơn giản nhất với **cấu hình YAML** và **các agent nói tiếng Việt** chỉ trong 5 phút.

## Prerequisites / Yêu Cầu

- Go 1.25 or later
- OpenAI API key (from https://platform.openai.com/account/api-keys)

## Quick Start (5 minutes) / Bắt Đầu Nhanh (5 phút)

### Step 1: Get Your API Key / Lấy API Key

1. Go to https://platform.openai.com/account/api-keys
2. Create a new API key (or use an existing one)
3. Copy it to clipboard

### Step 2: Create .env File / Tạo File .env

```bash
cp .env.example .env
```

Open `.env` and add your API key:

```env
OPENAI_API_KEY=sk-proj-your-actual-key-here
```

### Step 3: Install Dependencies / Cài Đặt Phụ Thuộc

```bash
go mod download
```

### Step 4: Run the Example / Chạy Ví Dụ

```bash
go run main.go
```

## Expected Output / Kết Quả Mong Đợi

```
🤖 Hệ Thống Thảo Luận Multi-Agent Đơn Giản
==================================================

📌 Chủ đề 1: Những thực hành tốt nhất khi viết code Go là gì?
--------------------------------------------------

[Người Tò Mò]: Khi viết code Go, những thực hành tốt nhất là gì?

[Chuyên Gia]: Có rất nhiều thực hành tốt nhất mà bạn nên biết...

✅ Kết Quả Cuối Cùng:
[Final response from Expert in Vietnamese]

...
```

## Understanding the Flow / Hiểu Cách Hoạt Động

The example creates a crew with 2 Vietnamese-speaking agents:

1. **Người Tò Mò** (Enthusiast)
   - Asks insightful questions
   - Explores ideas
   - Can pass to Expert

2. **Chuyên Gia** (Expert)
   - Provides knowledgeable answers
   - Gives final response
   - Terminal agent (stops conversation)

**Conversation Flow:**

```
Người Tò Mò: "Hãy cho tôi biết về..."
    ↓
Chuyên Gia: "Tất nhiên! Đây là những điều bạn cần biết..."
    ↓
Người Tò Mò: "Bạn có thể giải thích thêm về..."
    ↓
Chuyên Gia: "Chắc chắn! Chi tiết như sau..."
    ↓ (End - Expert is terminal)
```

## YAML Configuration / Cấu Hình YAML

The crew is configured in `crew.yaml`:

```yaml
crew:
  maxRounds: 4          # Max conversation rounds
  maxHandoffs: 3        # Max handoffs between agents

agents:
  - id: "enthusiast"
    name: "Người Tò Mò"
    role: "Người học hỏi đầy tò mò"
    backstory: |
      Bạn là một người yêu thích khám phá những ý tưởng mới...
    model: "gpt-4o-mini"
    temperature: 0.8
    isTerminal: false

  - id: "expert"
    name: "Chuyên Gia"
    role: "Chuyên gia có kiến thức sâu"
    backstory: |
      Bạn là một chuyên gia thông thái...
    model: "gpt-4o-mini"
    temperature: 0.7
    isTerminal: true

topics:
  - "Những thực hành tốt nhất khi viết code Go là gì?"
  - "Làm thế nào mà các AI agent có thể cải thiện phát triển phần mềm?"
  - "..."
```

## Key Features of YAML Config / Các Tính Năng Chính của Cấu Hình YAML

✅ **No Recompiling Required** - Change config without rebuilding
✅ **Easy to Customize** - Non-developers can modify topics
✅ **Flexible** - Add agents, change parameters easily
✅ **Clear Structure** - All config in one readable file
✅ **Vietnamese Support** - Full UTF-8 support for Vietnamese text

## Customization / Tùy Chỉnh

### Add More Topics / Thêm Chủ Đề

Edit `crew.yaml`:

```yaml
topics:
  - "Chủ đề của bạn ở đây"
  - "Một chủ đề khác"
  - "Và thêm nữa..."
```

### Change Agent Names and Personalities / Thay Đổi Tên và Tính Cách

Edit `crew.yaml`:

```yaml
agents:
  - id: "expert"
    name: "Tiến Sĩ Thông Minh"
    role: "Một chuyên gia về công nghệ"
    backstory: "Bạn là một tiến sĩ với kinh nghiệm 20 năm..."
```

### Longer Conversations / Cuộc Trò Chuyện Dài Hơn

Edit `crew.yaml`:

```yaml
crew:
  maxRounds: 6        # More rounds
  maxHandoffs: 4      # More handoffs
```

### Different Models / Model Khác

Edit `crew.yaml`:

```yaml
agents:
  - id: "expert"
    model: "gpt-4o"           # More capable
    # or
    model: "gpt-3.5-turbo"    # Cheaper
```

### More Creative Responses / Phản Hồi Sáng Tạo Hơn

Edit `crew.yaml`:

```yaml
agents:
  - id: "enthusiast"
    temperature: 0.9    # Higher = more creative (0.0-1.0)
```

## How the Code Works / Cách Code Hoạt Động

### main.go Structure:

**Step 1: Load Environment**
```go
loadEnvFile()  // Reads .env file
apiKey := os.Getenv("OPENAI_API_KEY")
```

**Step 2: Load Configuration**
```go
config, err := loadConfig("crew.yaml")  // Parse YAML
```

**Step 3: Create Crew from Config**
```go
crew := createCrewFromConfig(config)  // Convert YAML to Agent objects
```

**Step 4: Run Conversations**
```go
executor := agentic.NewTeamExecutor(crew, apiKey)
response, err := executor.Execute(ctx, topic)  // Each topic
```

## Troubleshooting / Khắc Phục Sự Cố

### Problem: "OPENAI_API_KEY environment variable not set"

**Solution:**
```bash
cp .env.example .env
nano .env  # Add your API key
```

### Problem: "cannot unmarshal"

**Cause:** `crew.yaml` has incorrect YAML syntax

**Solution:** Check YAML formatting:
- Indentation must be spaces (not tabs)
- No trailing colons
- Quotes around multiline strings

### Problem: "Agents speaking in English"

**Cause:** The backstory instructs agents to speak Vietnamese

**Solution:** Make sure `crew.yaml` has proper Vietnamese instructions in backstory fields

### Problem: "module not found"

**Solution:**
```bash
go mod download
go mod tidy
```

### Problem: "file crew.yaml not found"

**Solution:** Make sure file is in same directory as main.go

```bash
# Verify
ls crew.yaml  # Should show crew.yaml

# If not, you're in wrong directory
pwd  # Check current directory
```

## File Structure Explained / Giải Thích Cấu Trúc File

```
simple-chat/
├── main.go              # Application code (~140 lines)
│   ├── Type definitions (Config, AgentConfig)
│   ├── main() - Load env, config, create crew
│   ├── loadConfig() - Parse crew.yaml
│   ├── createCrewFromConfig() - Build agents
│   └── loadEnvFile() - Load .env
│
├── crew.yaml            # Configuration file (~70 lines)
│   ├── crew settings (maxRounds, maxHandoffs)
│   ├── agents definitions (Người Tò Mò, Chuyên Gia)
│   └── topics for discussion
│
├── .env.example         # Template
│   └── OPENAI_API_KEY=sk-proj-...
│
├── go.mod & go.sum      # Dependencies
├── README.md            # Full documentation
└── GETTING_STARTED.md   # This file
```

## Code Examples / Ví Dụ Code

### Load and Parse YAML

```go
// loadConfig reads and parses crew.yaml
func loadConfig(filename string) (*Config, error) {
    data, err := os.ReadFile(filename)
    if err != nil {
        return nil, err
    }
    var config Config
    yaml.Unmarshal(data, &config)
    return &config, nil
}
```

### Create Crew from Config

```go
// Convert YAML config to Agent objects
func createCrewFromConfig(config *Config) *agentic.Crew {
    agents := make([]*agentic.Agent, len(config.Agents))
    
    for i, agentCfg := range config.Agents {
        agents[i] = &agentic.Agent{
            ID:          agentCfg.ID,
            Name:        agentCfg.Name,
            Role:        agentCfg.Role,
            Backstory:   agentCfg.Backstory,
            Model:       agentCfg.Model,
            Temperature: agentCfg.Temperature,
            IsTerminal:  agentCfg.IsTerminal,
        }
    }
    
    return &agentic.Crew{
        Agents:      agents,
        MaxRounds:   config.Crew.MaxRounds,
        MaxHandoffs: config.Crew.MaxHandoffs,
    }
}
```

## Vietnamese Language Features / Tính Năng Tiếng Việt

All text is in Vietnamese:
- 🤖 Agent names: Người Tò Mò, Chuyên Gia
- 💬 Conversation between agents
- 📝 Agent roles and backstories
- ✅ Error messages
- 📋 Output formatting

## Security Best Practices / Thực Hành Bảo Mật

✅ **Do / Nên làm:**
- Use `.env.example` as template
- Never commit `.env` file
- Rotate API keys regularly
- Check `.gitignore` excludes `.env`

❌ **Don't / Không nên:**
- Hardcode API keys in code
- Commit `.env` files
- Share API keys via email
- Reuse keys across projects

## Next Steps / Bước Tiếp Theo

1. **Customize the Topics / Tùy Chỉnh Chủ Đề**
   - Edit the topics list in `crew.yaml`
   - Try your own questions

2. **Modify Agent Behavior / Thay Đổi Hành Vi Agent**
   - Change Temperature values
   - Edit backstories
   - Change agent names

3. **Add More Agents / Thêm Nhiều Agent Hơn**
   - Add new entries to `agents` section
   - Define their roles and responsibilities

4. **Explore Other Examples / Khám Phá Các Ví Dụ Khác**
   - customer-service (3 agents, with tools)
   - it-support (real-world IT workflow)
   - research-assistant (multi-step process)

5. **Build Your Own / Xây Dựng Của Riêng Bạn**
   - Create custom YAML configurations
   - Design your own multi-agent systems
   - Add specialized tools

## Tips for Success / Mẹo Thành Công

1. ✅ Start with default config - understand it first
2. ✅ Make one change at a time - see the effect
3. ✅ Use descriptive agent names in Vietnamese
4. ✅ Write clear backstories - guides agent behavior
5. ✅ Test with different topics
6. ✅ Monitor costs - keep an eye on API usage

## Getting Help / Nhận Trợ Giúp

- **README.md** - Full documentation
- **crew.yaml** - Configuration examples
- **main.go** - Code implementation
- **/SECURITY.md** - Security guidelines
- go-agentic main documentation

## Key Advantages of YAML Config / Lợi Thế Của Cấu Hình YAML

| Feature | Benefit |
|---------|---------|
| **No Code Changes** | Modify crew without recompiling |
| **Non-Developer Friendly** | Business users can customize |
| **Easy to Version Control** | Configuration changes are tracked |
| **Multi-Language Support** | Full UTF-8 for any language |
| **Readable Format** | Easy to understand and modify |
| **Flexible** | Add agents, topics without coding |

## Summary / Tóm Tắt

This example demonstrates:

✅ Loading configuration from YAML
✅ Creating agents dynamically from config
✅ Multi-language support (Vietnamese)
✅ Easy customization without code changes
✅ Clean, understandable code structure
✅ Professional error handling
✅ Best practices for configuration management

---

**Ready to run? / Sẵn sàng chạy?**

```bash
cp .env.example .env
# Edit .env with your API key
# Chỉnh sửa .env với API key của bạn

go run main.go
```

Good luck! 🚀
Chúc bạn thành công! 🚀

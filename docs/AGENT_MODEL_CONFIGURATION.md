# Agent Model Configuration Guide

**Document Version**: 1.0
**Last Updated**: 2025-12-22

---

## Tổng Quan (Overview)

Tài liệu này mô tả cách cấu hình LLM models cho agents trong go-agentic, bao gồm primary models, backup models, và fallback mechanisms.

---

## 1. Configuration Formats

Go-agentic hỗ trợ hai format cấu hình model (cả hai hoàn toàn hỗ trợ):

### 1.1 Legacy Format (Vẫn Hỗ Trợ)

```yaml
id: my-agent
name: My Agent
role: Agent Role
description: Agent description
backstory: Agent backstory

# Legacy format (deprecated but still works)
model: gpt-4-turbo
provider: openai
provider_url: https://api.openai.com

temperature: 0.7
is_terminal: true
```

**Backward Compatible**: Yes, completely supported
**Recommended**: No, use new format

### 1.2 New Format (Recommended ⭐)

```yaml
id: my-agent
name: My Agent
role: Agent Role
description: Agent description
backstory: Agent backstory

# New format with Primary & Backup models
primary:
  model: gpt-4-turbo
  provider: openai
  provider_url: https://api.openai.com

# Optional: Backup model for fallback
backup:
  model: gpt-4o-mini
  provider: openai
  provider_url: https://api.openai.com

temperature: 0.7
is_terminal: true
```

**Advantages**:
- ✅ High availability with automatic failover
- ✅ Cost optimization (cheap backup)
- ✅ Provider flexibility
- ✅ Clearer configuration structure

---

## 2. Primary Model Configuration

**Definition**: The main LLM model that will be used for agent execution

### Structure

```yaml
primary:
  model: model-name
  provider: provider-name
  provider_url: provider-url
```

### Fields

#### **model** (String, Required)

**Definition**: Name of the LLM model

**For OpenAI**:
```yaml
primary:
  model: gpt-4-turbo        # Recommended for complex tasks
  model: gpt-4o             # Latest general model
  model: gpt-4o-mini        # Faster, cheaper
  model: gpt-3.5-turbo      # Legacy, cheap
```

**For Ollama**:
```yaml
primary:
  model: deepseek-r1:1.5b   # Reasoning model (recommended)
  model: gemma3:1b          # Light, fast
  model: mistral:latest     # General purpose
  model: llama2:70b         # Large model
```

#### **provider** (String, Required)

**Definition**: LLM service provider

**Supported Values**:
```yaml
primary:
  provider: openai          # OpenAI API
  provider: ollama          # Local Ollama
```

#### **provider_url** (String, Required for Ollama)

**Definition**: URL endpoint of LLM provider

**For OpenAI**:
```yaml
primary:
  provider: openai
  provider_url: https://api.openai.com  # Usually not needed (uses API key)
```

**For Ollama**:
```yaml
primary:
  provider: ollama
  provider_url: http://localhost:11434           # Local machine
  provider_url: http://192.168.1.100:11434      # Remote server
  provider_url: http://ollama.example.com:11434 # Custom domain
```

### Examples

#### OpenAI Model

```yaml
id: gpt4-agent
name: GPT-4 Agent
primary:
  model: gpt-4-turbo
  provider: openai
  provider_url: https://api.openai.com
```

#### Ollama Model

```yaml
id: ollama-agent
name: Ollama Agent
primary:
  model: deepseek-r1:1.5b
  provider: ollama
  provider_url: http://localhost:11434
```

---

## 3. Backup Model Configuration

**Definition**: Fallback model used when primary model fails

### When to Use Backup Model

✅ **Use backup model when**:
- You need high availability
- Primary model has rate limits
- Primary model is expensive
- You want automatic cost optimization
- You mix different providers (OpenAI + Ollama)

❌ **Don't need backup model if**:
- Single-model setup is acceptable
- Cost not a concern
- Development/testing environment

### Structure

```yaml
backup:
  model: model-name
  provider: provider-name
  provider_url: provider-url
```

**Same fields as primary model**

### Fallback Logic

```
ExecuteAgent() called
    ↓
Try PRIMARY model
    ├─ Success? → Return response
    └─ Fail? → Check if backup exists
        ├─ Backup exists? → Try BACKUP model
        │   ├─ Success? → Return response
        │   └─ Fail? → Return error (both failed)
        └─ No backup? → Return error (primary failed)
```

### Examples

#### OpenAI Primary + Backup

```yaml
id: resilient-agent
name: Resilient Agent
primary:
  model: gpt-4-turbo
  provider: openai

backup:
  model: gpt-4o-mini
  provider: openai
```

**Fallback Path**:
1. Try gpt-4-turbo (premium, fast)
2. If fails, try gpt-4o-mini (cheaper, slower)

#### Ollama Primary + OpenAI Backup

```yaml
id: mixed-provider-agent
name: Mixed Provider Agent
primary:
  model: deepseek-r1:1.5b
  provider: ollama
  provider_url: http://localhost:11434

backup:
  model: gpt-4o-mini
  provider: openai
```

**Fallback Path**:
1. Try local Ollama (free, fast, private)
2. If Ollama unavailable, fallback to OpenAI (costs money but reliable)

#### Cost Optimization

```yaml
id: cost-optimized-agent
name: Cost Optimized Agent
primary:
  model: gpt-4o-mini
  provider: openai

backup:
  model: gemma3:1b
  provider: ollama
  provider_url: http://localhost:11434
```

**Fallback Path**:
1. Try cheap OpenAI model (small cost)
2. If OpenAI fails, try free local model (no cost)

---

## 4. Temperature Configuration

**Definition**: Controls creativity level of model responses

### Value Range

```
0.0 ──────────────────────────────── 1.0
├─ Deterministic ──────────────── Creative
└─ Exact answers          ────  Random output
```

### Recommended Values by Task

#### Deterministic Tasks (0.1 - 0.3)

```yaml
temperature: 0.3
```

**Use for**:
- Data classification
- Math/Logic problems
- Code generation
- Fact extraction
- Routing decisions

**Example**:
```yaml
id: classifier-agent
temperature: 0.2  # Very deterministic
```

#### Balanced Tasks (0.5 - 0.7)

```yaml
temperature: 0.6
```

**Use for**:
- General Q&A
- Analysis reports
- Recommendations
- Content analysis
- Most common use case ⭐

**Example**:
```yaml
id: analyzer-agent
temperature: 0.7  # Balanced
```

#### Creative Tasks (0.8 - 1.0)

```yaml
temperature: 0.9
```

**Use for**:
- Creative writing
- Brainstorming
- Ideation
- Content generation
- Creative problem-solving

**Example**:
```yaml
id: writer-agent
temperature: 0.9  # Creative
```

### Full Configuration Example

```yaml
id: multi-task-agent
name: Multi-Task Agent
primary:
  model: gpt-4-turbo
  provider: openai

# Temperature for this agent
temperature: 0.6  # Balanced

# Different agents can have different temperatures
```

---

## 5. Model Selection Guide

### Decision Tree

```
What's your use case?

├─ Local, Private, Free?
│  └─ Use Ollama
│     ├─ deepseek-r1:1.5b (best reasoning)
│     └─ gemma3:1b (fast, lightweight)
│
├─ Complex reasoning required?
│  └─ Use OpenAI
│     ├─ gpt-4-turbo (best for reasoning)
│     ├─ gpt-4o (balanced, good for most)
│     └─ gpt-4o-mini (fast, cheap)
│
├─ Production, High Availability?
│  └─ Primary: gpt-4-turbo
│     Backup: gpt-4o-mini
│
├─ Development / Testing?
│  └─ Use Ollama (free, fast)
│     └─ gemma3:1b or deepseek-r1:1.5b
│
└─ Cost-Sensitive?
   └─ Primary: gpt-4o-mini or Ollama
      Backup: Free option (Ollama or cheaper model)
```

### Comparison Table

| Model | Provider | Speed | Quality | Cost | Best For |
|-------|----------|-------|---------|------|----------|
| deepseek-r1:1.5b | Ollama | Fast | Good | Free | Local development |
| gemma3:1b | Ollama | Very Fast | Fair | Free | Quick responses |
| mistral:latest | Ollama | Fast | Good | Free | General purpose |
| gpt-3.5-turbo | OpenAI | Fast | Fair | $ | Budget option |
| gpt-4o-mini | OpenAI | Medium | Good | $$ | Best value |
| gpt-4o | OpenAI | Medium | Excellent | $$$ | Production |
| gpt-4-turbo | OpenAI | Slow | Excellent | $$$ | Complex reasoning |

---

## 6. Provider Setup

### OpenAI Setup

1. **Get API Key**:
   - Sign up at https://platform.openai.com
   - Create API key in settings
   - Save securely (never commit to repo)

2. **Set Environment Variable**:
   ```bash
   export OPENAI_API_KEY=sk-...
   ```

3. **Configure Agent**:
   ```yaml
   primary:
     model: gpt-4-turbo
     provider: openai
     provider_url: https://api.openai.com
   ```

### Ollama Setup

1. **Install Ollama**:
   - Download from https://ollama.ai
   - Or use Docker: `docker run -it -p 11434:11434 ollama/ollama`

2. **Pull Model**:
   ```bash
   ollama pull deepseek-r1:1.5b
   ollama pull gemma3:1b
   ```

3. **Start Ollama** (if not running):
   ```bash
   ollama serve
   ```

4. **Verify**:
   ```bash
   curl http://localhost:11434/api/tags
   ```

5. **Configure Agent**:
   ```yaml
   primary:
     model: deepseek-r1:1.5b
     provider: ollama
     provider_url: http://localhost:11434
   ```

---

## 7. Complete Agent Configuration Examples

### Simple OpenAI Agent

```yaml
id: simple-openai
name: Simple OpenAI Agent
role: General Assistant
description: Simple agent using OpenAI

backstory: |
  I am a helpful assistant powered by GPT-4.
  I provide accurate and thoughtful responses.

primary:
  model: gpt-4-turbo
  provider: openai

temperature: 0.7
is_terminal: true
tools: []
```

### Resilient Production Agent

```yaml
id: resilient-prod
name: Resilient Production Agent
role: Production Specialist
description: High-availability agent for production

backstory: |
  I am a production-grade agent with automatic failover.
  I ensure service reliability and cost efficiency.

# Primary: Fast and reliable
primary:
  model: gpt-4-turbo
  provider: openai

# Backup: Cost-effective fallback
backup:
  model: gpt-4o-mini
  provider: openai

temperature: 0.5
is_terminal: false

tools:
  - GetSystemStatus
  - MonitorHealth

handoff_targets:
  - executor
```

### Local Development Agent

```yaml
id: local-dev
name: Local Development Agent
role: Development Helper
description: Local agent for development with zero cost

backstory: |
  I run locally on your machine.
  I provide instant feedback without API costs.

primary:
  model: deepseek-r1:1.5b
  provider: ollama
  provider_url: http://localhost:11434

backup:
  model: gemma3:1b
  provider: ollama
  provider_url: http://localhost:11434

temperature: 0.6
is_terminal: true
tools: []
```

### Mixed Provider Agent

```yaml
id: mixed-provider
name: Mixed Provider Agent
role: Flexible Specialist
description: Uses local first, cloud fallback

backstory: |
  I prefer privacy and cost savings with local execution.
  But I can fallback to cloud when needed.

# Primary: Local (private, free)
primary:
  model: deepseek-r1:1.5b
  provider: ollama
  provider_url: http://localhost:11434

# Backup: Cloud (reliable, paid)
backup:
  model: gpt-4o-mini
  provider: openai

temperature: 0.7
is_terminal: true
tools:
  - AnalyzeData
  - GenerateReport
```

---

## 8. Performance & Cost Optimization

### Speed Optimization

```yaml
# Fast responses prioritized
id: fast-agent
primary:
  model: gpt-4o-mini          # Fast
  provider: openai

backup:
  model: gemma3:1b            # Very fast
  provider: ollama
  provider_url: http://localhost:11434

temperature: 0.3              # Lower = faster
```

### Cost Optimization

```yaml
# Cost minimization prioritized
id: cheap-agent
primary:
  model: gemma3:1b            # Free
  provider: ollama
  provider_url: http://localhost:11434

backup:
  model: gpt-4o-mini          # Cheap
  provider: openai

temperature: 0.5
```

### Quality Optimization

```yaml
# Best quality, cost secondary
id: quality-agent
primary:
  model: gpt-4-turbo          # Best quality
  provider: openai

backup:
  model: gpt-4o               # Good quality
  provider: openai

temperature: 0.6              # Balanced
```

---

## 9. Troubleshooting

### Problem: "Provider not found: openai"

**Solution**: Ensure you have `OPENAI_API_KEY` environment variable set:
```bash
export OPENAI_API_KEY=sk-...
```

### Problem: "Cannot connect to Ollama at http://localhost:11434"

**Solution**: Make sure Ollama is running:
```bash
ollama serve
```

Or verify connection:
```bash
curl http://localhost:11434/api/tags
```

### Problem: "Model not found: gpt-4-turbo"

**Solution**: Check your OpenAI API key has access to that model

### Problem: Agent always uses backup

**Solution**: Check primary model configuration is correct. Monitor logs for errors.

### Problem: Too slow / Too expensive

**Solution**: Adjust temperature (lower = faster) or switch to cheaper model

---

## 10. Best Practices

✅ **DO**:
- Use new format with primary/backup
- Set appropriate temperature for task
- Use Ollama for development/testing
- Use OpenAI for production
- Monitor model performance
- Have a backup model for production

❌ **DON'T**:
- Use hardcoded model names
- Commit API keys to repository
- Use gpt-4-turbo for simple tasks
- Run Ollama on production without load balancing
- Ignore model errors

---

## References

- [CORE_LIBRARY_UPDATES.md](CORE_LIBRARY_UPDATES.md) - New features overview
- [CONFIG_SPECIFICATION.md](CONFIG_SPECIFICATION.md) - Full config spec
- [TEAM_SETUP_EXAMPLES.md](TEAM_SETUP_EXAMPLES.md) - Real examples

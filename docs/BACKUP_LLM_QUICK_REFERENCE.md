# Backup LLM Model - Quick Reference

## ⚡ 30-Second Overview

Agents can now have a **primary** and **backup** LLM model. If primary fails → automatically tries backup.

## 🎯 Basic Usage

### Configuration (YAML)

```yaml
primary:
  model: gpt-4o
  provider: openai
  provider_url: https://api.openai.com

backup:
  model: llama2:70b
  provider: ollama
  provider_url: http://localhost:11434
```

### Execution Flow

```
Try Primary (gpt-4o on OpenAI)
    ├─ Success? ✅ Return
    └─ Failed? → Try Backup
        ├─ Success? ✅ Return
        └─ Failed? ❌ Error
```

---

## 📋 Configuration Examples

### Example 1: Dev + Prod Fallback
```yaml
primary:
  model: llama2:70b
  provider: ollama
  provider_url: http://localhost:11434

backup:
  model: gpt-4o
  provider: openai
  provider_url: https://api.openai.com
```
**Use case:** Free local model, API as backup

### Example 2: Single Model (No Backup)
```yaml
primary:
  model: gpt-4o
  provider: openai
  provider_url: https://api.openai.com

# backup: (omitted - no fallback)
```
**Use case:** Simple setup, no redundancy

### Example 3: Multi-Region Resilience
```yaml
primary:
  model: gpt-4o
  provider: openai
  provider_url: https://api.openai.com

backup:
  model: gpt-4o
  provider: openai
  provider_url: https://api-eu.openai.com  # Different region
```
**Use case:** Geographic redundancy

---

## 🔄 Supported Fallback Paths

✅ **Any Provider → Any Provider:**
- OpenAI → Ollama
- Ollama → OpenAI
- OpenAI → Different OpenAI region
- Ollama → Different Ollama instance

---

## ✅ What Works

| Feature | Status |
|---------|--------|
| Primary + Backup config | ✅ |
| Automatic fallback on error | ✅ |
| Streaming with fallback | ✅ |
| Cross-provider fallback | ✅ |
| Multiple agents with different configs | ✅ |
| Backward compatibility (old format) | ✅ |
| Config validation | ✅ |

---

## ⚠️ Known Limitations

| Limitation | Details |
|-----------|---------|
| **Mid-stream fallback** | If primary fails mid-stream, response incomplete |
| **Latency** | Fallback adds ~1-2s if primary fails |
| **Cost** | Backup charges when primary fails |
| **No retry** | Only one fallback attempt (primary → backup) |

---

## 🧪 Testing

### Run Tests
```bash
cd core
go test -v -run "Primary|Backup"
```

### Expected Output
```
✅ TestAgentWithPrimaryModelConfig
✅ TestAgentWithPrimaryAndBackupConfig
✅ TestValidateAgentConfigWithPrimaryModel
✅ TestValidateAgentConfigWithPrimaryAndBackup
✅ TestValidateAgentConfigEmptyPrimaryModel
✅ TestValidateAgentConfigEmptyPrimaryProvider
✅ TestValidateAgentConfigEmptyBackupModel
✅ TestValidateAgentConfigEmptyBackupProvider
```

---

## 🔧 Migration

### From Old Format

**Old:**
```yaml
model: gpt-4o
provider: openai
provider_url: https://api.openai.com
```

**New:**
```yaml
primary:
  model: gpt-4o
  provider: openai
  provider_url: https://api.openai.com

backup:  # optional
  model: llama2
  provider: ollama
  provider_url: http://localhost:11434
```

**Important:** Old format still works! Auto-converted to `primary` internally.

---

## 📊 When Fallback Triggers

### ✅ Triggers Fallback
- `429 Too Many Requests` (rate limit)
- `500/502/503` (server errors)
- `timeout` (request timeout)
- `connection refused` (provider down)
- Network errors

### ❌ Does NOT Trigger Fallback
- `401 Unauthorized` (auth failed)
- `404 Not Found` (model doesn't exist)
- `400 Bad Request` (malformed request)

---

## 📈 Monitoring

### Console Output
```
[FALLBACK] Primary 'gpt-4o' (openai) failed: 429
Trying backup 'llama2' (ollama)...
[FALLBACK SUCCESS] Backup succeeded
```

### What's Logged
- Primary model and error
- Backup model being tried
- Success/failure result

---

## 🚀 Common Patterns

### Pattern 1: Cost Optimization
```yaml
primary:
  model: mistral:7b  # cheap/free
  provider: ollama

backup:
  model: gpt-4o      # expensive
  provider: openai
```

### Pattern 2: Development
```yaml
primary:
  model: llama2      # local, no API key needed
  provider: ollama

backup:
  model: gpt-4o      # production fallback
  provider: openai
```

### Pattern 3: Reliability
```yaml
primary:
  model: gpt-4o
  provider: openai
  provider_url: https://api.openai.com

backup:
  model: gpt-4o
  provider: openai
  provider_url: https://api-eu.openai.com  # different region
```

---

## 🎓 Best Practices

1. **Always set Primary** - Required field
2. **Optional Backup** - Add when needed
3. **Test Both Models** - Ensure backup works with your use case
4. **Monitor Fallbacks** - Track how often primary fails
5. **Different Models** - Backup should differ from primary
6. **Cost Aware** - Understand backup cost implications

---

## ❓ FAQ

**Q: Do I need a backup model?**
A: No. If no backup, failure returns error as before.

**Q: Can I have 3+ models?**
A: Not yet. Max 1 primary + 1 backup. Feature request: tertiary model.

**Q: Will old YAML files break?**
A: No! Auto-converted to `primary`. Fully backward compatible.

**Q: How do I test the fallback?**
A: Stop the primary provider (e.g., kill Ollama) → fallback triggers.

**Q: Does streaming support fallback?**
A: Yes, but only if primary fails before first token.

---

## 📚 Files Modified

```
core/types.go         (+15 lines)  - ModelConfig struct, Agent.Primary/Backup
core/config.go        (+90 lines)  - ModelConfigYAML, config parsing, validation
core/agent.go         (+120 lines) - Fallback execution logic
core/agent_test.go    (+90 lines)  - Structure tests
core/config_test.go   (+150 lines) - Validation tests
examples/.../agent.yaml (updated) - Example with primary/backup
docs/BACKUP_LLM_MODEL_FEATURE.md  - Full documentation
```

---

## 🔗 Related

- **Issue #23:** Agent config validation
- **Issue #6:** YAML validation at load time
- **Hardcoded Values Audit:** Default provider selection

---

**Version:** 1.0
**Status:** ✅ Production Ready
**Last Updated:** 2025-12-22

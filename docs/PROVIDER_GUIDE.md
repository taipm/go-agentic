# Provider Guide: Multi-Provider LLM Support

This guide explains how to use go-crewai with different LLM providers (Ollama and OpenAI).

## Overview

go-crewai now supports multiple LLM backends through a provider abstraction layer:

- **Ollama** - Local development with open-source models (recommended)
- **OpenAI** - Cloud-based API for production deployments

Switch between providers by modifying your agent configuration in YAML.

## Ollama (Recommended for Local Development)

### What is Ollama?

[Ollama](https://ollama.com) is a tool that lets you run open-source LLMs locally on your machine. No API keys needed, completely free, and fast.

### Installation

1. **Download Ollama**: Visit https://ollama.com and download for your platform
2. **Install**: Follow the platform-specific instructions
3. **Verify installation**:
   ```bash
   ollama --version
   ```

### Running Ollama

Start the Ollama server in a terminal:

```bash
ollama serve
```

The server will be available at `http://localhost:11434` by default.

### Pulling Models

Before using a model, pull it first:

```bash
# Ultra lightweight (recommended for testing)
ollama pull gemma3:1b

# Small reasoning model (recommended for IT support)
ollama pull deepseek-r1:1.5b

# Larger models (requires more RAM)
ollama pull llama3.1:8b
ollama pull mistral:7b
ollama pull qwen2.5:7b
```

### Checking Available Models

```bash
ollama list
```

### Configuring go-crewai for Ollama

Add these fields to your agent YAML configuration:

```yaml
# Use Ollama with deepseek-r1:1.5b (default model)
model: deepseek-r1:1.5b
provider: ollama
provider_url: http://localhost:11434  # Default URL, can be omitted
```

Or use different models:

```yaml
# Ultra lightweight option
model: gemma3:1b
provider: ollama
provider_url: http://localhost:11434
```

```yaml
# Larger model with more capabilities
model: llama3.1:8b
provider: ollama
provider_url: http://localhost:11434
```

### Complete Example

```yaml
name: my-agent
role: System Diagnostician
description: An AI that helps diagnose system issues

model: deepseek-r1:1.5b
temperature: 0.7
provider: ollama
provider_url: http://localhost:11434

backstory: |
  You are an expert in system diagnostics.
  Your role is to help users understand and fix their system issues.
```

### Model Recommendations

| Model | Size | Best For | RAM |
|-------|------|----------|-----|
| **gemma3:1b** | 1B | Quick tests, minimal resources | ~1GB |
| **deepseek-r1:1.5b** | 1.5B | IT support, reasoning (DEFAULT) | ~2GB |
| **llama3.1:8b** | 8B | Better reasoning, multi-step tasks | ~8GB |
| **mistral:7b** | 7B | Fast, good quality responses | ~7GB |
| **qwen2.5:7b** | 7B | Excellent for multilingual | ~7GB |

### Tool Calling with Ollama

Ollama models don't have native tool calling support like OpenAI. Instead, go-crewai uses **text parsing** to extract tool calls from responses.

Model responses should include tool calls in this format:

```
Let me analyze the system.
GetCPUUsage()
GetMemoryUsage()
CheckDiskStatus("/")
```

The framework will automatically extract `GetCPUUsage()`, `GetMemoryUsage()`, and `CheckDiskStatus()` as tool calls.

**Recommended**: Use models with good reasoning capabilities like `deepseek-r1:1.5b` to ensure reliable tool call generation.

### Streaming Support

Ollama supports real-time streaming responses. Enable streaming in your code:

```go
// Streaming will automatically work with ExecuteAgentStream()
streamChan := make(chan providers.StreamChunk)
go agent.ExecuteAgentStream(ctx, req, streamChan)

for chunk := range streamChan {
    if chunk.Error != nil {
        fmt.Println("Error:", chunk.Error)
        break
    }

    if chunk.Content != "" {
        fmt.Print(chunk.Content) // Real-time streaming output
    }

    if chunk.Done {
        break
    }
}
```

### Troubleshooting Ollama

**Issue**: `Connection refused` error

- Make sure Ollama server is running: `ollama serve`
- Check if server is listening on port 11434: `curl http://localhost:11434/api/tags`

**Issue**: Model not found

```bash
# Pull the model first
ollama pull deepseek-r1:1.5b
```

**Issue**: Slow responses

- Use a smaller model (gemma3:1b instead of llama3.1:8b)
- Reduce temperature for faster deterministic responses
- Use GPU acceleration if available

**Issue**: Tool calls not being extracted

- Check if model output follows the format: `ToolName(args)`
- Use models with better reasoning: deepseek-r1:1.5b
- Ensure tool names start with uppercase letter (e.g., `GetCPUUsage`, not `getCPUUsage`)

---

## OpenAI (Cloud-Based)

### Prerequisites

- [OpenAI API key](https://platform.openai.com/api-keys)
- Active OpenAI account with API access

### Configuring go-crewai for OpenAI

Add these fields to your agent YAML:

```yaml
model: gpt-4o-mini
provider: openai
```

The API key is passed at runtime, not in config files for security.

### Model Options

| Model | Cost | Speed | Quality |
|-------|------|-------|---------|
| **gpt-4o-mini** | $ | Fast | Good |
| **gpt-4-turbo** | $$ | Medium | Very Good |
| **gpt-4o** | $$$ | Slow | Best |

### Complete Example

```yaml
name: my-agent
role: System Analyzer
description: Analyzes system behavior with OpenAI

model: gpt-4o-mini
temperature: 0.7
provider: openai

backstory: |
  You are an expert system analyst powered by OpenAI's advanced models.
```

### Using OpenAI in Code

```go
import "github.com/taipm/go-agentic/core/providers"

// Get OpenAI provider with API key
factory := providers.GetGlobalFactory()
provider, err := factory.GetProvider("openai", "", "sk-xxx")

// Use provider for completions
request := &providers.CompletionRequest{
    Model: "gpt-4o-mini",
    Messages: messages,
    Temperature: 0.7,
}

response, err := provider.Complete(ctx, request)
```

### Tool Calling with OpenAI

OpenAI models have native tool calling support via the `tool_calls` field in responses. go-crewai automatically extracts these native tool calls.

### Streaming with OpenAI

OpenAI's streaming API is fully supported:

```go
streamChan := make(chan providers.StreamChunk)
go provider.CompleteStream(ctx, request, streamChan)

for chunk := range streamChan {
    if chunk.Error != nil {
        log.Fatal(chunk.Error)
    }
    fmt.Print(chunk.Content)
    if chunk.Done {
        break
    }
}
```

### Cost Control

- Use `gpt-4o-mini` for cost-effective solutions
- Set `temperature: 0` for deterministic responses (slightly cheaper)
- Monitor usage in OpenAI dashboard
- Set rate limits if needed

### Troubleshooting OpenAI

**Issue**: Authentication error

- Verify API key is correct: `echo $OPENAI_API_KEY`
- Check API key has necessary permissions
- Visit https://platform.openai.com/api-keys to verify key status

**Issue**: Model not found

- Use valid model names: `gpt-4o-mini`, `gpt-4-turbo`, `gpt-4o`
- Check your account has access to the model

**Issue**: Rate limit exceeded

- Implement exponential backoff retry logic
- Reduce request frequency
- Upgrade to higher tier if needed

---

## Switching Between Providers

### Quick Switch in YAML

**From Ollama to OpenAI**:
```yaml
# Change this:
model: deepseek-r1:1.5b
provider: ollama
provider_url: http://localhost:11434

# To this:
model: gpt-4o-mini
provider: openai
```

**From OpenAI to Ollama**:
```yaml
# Change this:
model: gpt-4o-mini
provider: openai

# To this:
model: deepseek-r1:1.5b
provider: ollama
provider_url: http://localhost:11434
```

### Provider-Agnostic Code

The abstraction layer ensures your code works with both providers:

```go
// This code works with any provider (OpenAI, Ollama, etc.)
func executeAgent(ctx context.Context, agent *Agent, apiKey string) (*AgentResponse, error) {
    factory := providers.GetGlobalFactory()
    provider, err := factory.GetProvider(agent.Provider, agent.ProviderURL, apiKey)
    if err != nil {
        return nil, err
    }

    request := &providers.CompletionRequest{
        Model: agent.Model,
        Messages: messages,
        Temperature: agent.Temperature,
    }

    return provider.Complete(ctx, request)
}
```

---

## Performance Comparison

### Latency

- **Ollama**: 0.5-5 seconds per request (local, depends on model)
- **OpenAI**: 1-10 seconds per request (API call + network)

### Cost

- **Ollama**: Free (one-time setup)
- **OpenAI**: $0.15-3.00 per 1M tokens (depending on model)

### Quality (Tool Calling)

- **Ollama**: Good with deepseek-r1:1.5b (text parsing)
- **OpenAI**: Excellent (native tool calling support)

### Offline Capability

- **Ollama**: Yes (fully local)
- **OpenAI**: No (requires internet connection)

---

## Best Practices

### For Development

1. Use **Ollama with deepseek-r1:1.5b** for rapid testing
2. No API keys needed - faster iteration
3. Test locally before deploying to OpenAI

### For Production

1. Use **OpenAI with gpt-4o-mini** for best quality
2. Better tool calling support
3. More reliable and feature-rich
4. Monitor costs and set rate limits

### For Cost-Sensitive Projects

1. Start with Ollama locally
2. Test core functionality
3. Switch to OpenAI only for final deployment
4. Use gpt-4o-mini (cheaper than gpt-4-turbo)

### For Tool-Heavy Applications

- **Ollama**: Ensure models can handle tool calling format
- **OpenAI**: Native support, more reliable

---

## Example: IT Support System

The example IT support system uses Ollama by default:

```yaml
# orchestrator.yaml
model: deepseek-r1:1.5b
provider: ollama
provider_url: http://localhost:11434

# clarifier.yaml
model: deepseek-r1:1.5b
provider: ollama
provider_url: http://localhost:11434

# executor.yaml
model: deepseek-r1:1.5b
provider: ollama
provider_url: http://localhost:11434
```

To use OpenAI instead:

```yaml
# Change all agents to:
model: gpt-4o-mini
provider: openai
```

---

## Advanced Configuration

### Custom Ollama URL

For remote Ollama servers:

```yaml
model: deepseek-r1:1.5b
provider: ollama
provider_url: http://192.168.1.100:11434  # Remote server
```

### Environment-Based Configuration

In code, support both local and cloud:

```go
ollama := "http://localhost:11434"
if os.Getenv("OLLAMA_HOST") != "" {
    ollama = os.Getenv("OLLAMA_HOST")
}

provider, err := factory.GetProvider("ollama", ollama, "")
```

### Model Aliasing

Create shortcuts for common configurations:

```yaml
# config.yaml
models:
  local-reasoning: deepseek-r1:1.5b
  local-lightweight: gemma3:1b
  cloud-cheap: gpt-4o-mini
  cloud-quality: gpt-4o

agents:
  analyzer:
    model: local-reasoning  # Resolves to deepseek-r1:1.5b
```

---

## Debugging Provider Issues

### Enable Logging

```go
import "log"

// Logs will show which provider is used and any errors
log.SetFlags(log.Ldate | log.Ltime | log.Lshortfile)
```

### Verify Provider Connection

```bash
# Test Ollama
curl http://localhost:11434/api/tags

# Test OpenAI (if you have API key)
curl https://api.openai.com/v1/models \
  -H "Authorization: Bearer sk-xxx"
```

### Check Configuration

```go
agent := loadAgent("config.yaml")
fmt.Printf("Provider: %s\n", agent.Provider)
fmt.Printf("URL: %s\n", agent.ProviderURL)
fmt.Printf("Model: %s\n", agent.Model)
```

---

## FAQ

**Q: Can I use both providers in the same application?**

A: Yes! Different agents can use different providers. The factory handles caching and routing.

**Q: Does streaming work with both providers?**

A: Yes, both Ollama and OpenAI support streaming through the `CompleteStream()` method.

**Q: How do I upgrade a model without changing code?**

A: Just update the `model:` field in your YAML config. No code changes needed.

**Q: What if Ollama is offline?**

A: The provider will return a connection error. Handle errors gracefully or switch to OpenAI as fallback.

**Q: Can I cache responses?**

A: The provider factory caches provider instances (connections), not responses. Implement response caching separately if needed.

**Q: How do I measure token usage?**

A: OpenAI provides token counts in responses. Ollama doesn't expose token counts directly, but you can estimate based on model documentation.

---

## References

- [Ollama Official Site](https://ollama.com)
- [OpenAI API Documentation](https://platform.openai.com/docs)
- [go-crewai Architecture](./ARCHITECTURE.md)
- [Provider Implementation](../core/providers/)

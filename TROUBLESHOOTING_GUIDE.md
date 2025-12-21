# 🔧 Troubleshooting Guide - go-agentic

**Status**: Production Ready
**Version**: 1.0
**Last Updated**: 2025-12-22

---

## 📋 Quick Reference

| Issue | Symptom | Solution |
|-------|---------|----------|
| Server won't start | Port already in use | Use different port: `--port 8082` |
| API key error | 401 Unauthorized | Set `OPENAI_API_KEY` environment variable |
| Requests timing out | Slow responses | Increase timeout in config |
| High memory usage | Memory growing | Check for goroutine leaks or large tool results |
| Agent stuck in loop | Infinite back-and-forth | Reduce `max_handoffs` in config |
| Tool execution fails | Tool errors | Check tool parameters and available tools |

---

## 🚨 Common Issues and Solutions

### Issue 1: Server Won't Start

**Symptom**:
```
Error: listen tcp :8081: bind: address already in use
```

**Causes**:
- Port 8081 already in use
- Previous instance still running
- Permission issue (port < 1024)

**Solutions**:

**Option A**: Use different port
```bash
go-agentic --port 8082
```

**Option B**: Kill existing process
```bash
# Find process on port 8081
lsof -i :8081

# Kill process
kill -9 <PID>

# Or directly (macOS/Linux)
killall go-agentic
```

**Option C**: Check what's using the port
```bash
# Linux
sudo netstat -tlnp | grep 8081

# macOS
lsof -i :8081

# Windows
netstat -ano | findstr :8081
```

---

### Issue 2: API Key Not Found

**Symptom**:
```
Error: OPENAI_API_KEY environment variable not set
```

**Causes**:
- Environment variable not set
- Set in wrong shell
- Using wrong variable name

**Solutions**:

**Option A**: Set environment variable
```bash
# Temporary (this shell session only)
export OPENAI_API_KEY="sk-..."
go-agentic --server

# Permanent (add to ~/.bashrc or ~/.zshrc)
echo 'export OPENAI_API_KEY="sk-..."' >> ~/.bashrc
source ~/.bashrc
```

**Option B**: Verify it's set
```bash
echo $OPENAI_API_KEY
# Should print: sk-...
```

**Option C**: Set at runtime
```bash
OPENAI_API_KEY="sk-..." go-agentic --server
```

**Option D**: Create .env file (if supported)
```bash
# .env
OPENAI_API_KEY=sk-...

# Load before running
source .env
go-agentic --server
```

---

### Issue 3: Requests Keep Timing Out

**Symptom**:
```
event: error
data: {"error_type":"TOOL_TIMEOUT","message":"Tool exceeded timeout (5s)"}
```

**Causes**:
- Tool is slow (network, system load)
- Timeout too short for your environment
- Tool is hanging

**Solutions**:

**Option A**: Increase timeout in config
```yaml
# config/crew.yaml
timeouts:
  default_tool_timeout: 10    # Increase from 5s to 10s
  sequence_timeout: 60        # Increase from 30s to 60s
```

**Option B**: Increase specific tool timeout
```yaml
timeouts:
  per_tool_timeout:
    "PingHost": 15            # Network might be slow
    "GetCPUUsage": 3          # Should be fast
```

**Option C**: Diagnose which tool is slow
```bash
# Enable debug logging
export GO_AGENTIC_LOG_LEVEL=debug
go-agentic --server

# Look for slow tool logs:
# [DEBUG] Tool GetCPUUsage took 4.5s (near timeout!)
```

**Option D**: Check system resources
```bash
# Linux
top             # Check CPU, memory, I/O
iostat -x 1     # Check disk I/O

# macOS
activity monitor  # GUI tool

# Windows
Task Manager      # GUI tool or:
Get-Process | Sort-Object -Property CPU -Descending | Select-Object -First 5
```

---

### Issue 4: High Memory Usage

**Symptom**:
```
Memory usage: 500MB and growing
```

**Causes**:
- Goroutine leak
- Large tool results buffered
- No garbage collection
- Too many concurrent requests

**Solutions**:

**Option A**: Check goroutine count
```bash
# In your monitoring:
curl http://localhost:8081/metrics | jq '.goroutines'

# Should be relatively stable, not growing
```

**Option B**: Limit concurrent requests
```yaml
# config/crew.yaml
performance:
  max_concurrent_requests: 50    # Reduce from 100
```

**Option C**: Monitor and restart strategy
```bash
#!/bin/bash
# Restart server if memory exceeds 500MB
while true; do
  MEMORY=$(ps aux | grep go-agentic | awk '{print $6}' | head -1)
  if [ $MEMORY -gt 500000 ]; then
    echo "Memory too high: $MEMORY"
    killall go-agentic
    sleep 2
    go-agentic --server &
  fi
  sleep 60
done
```

**Option D**: Enable profiling
```bash
# Run with profiling
go-agentic --server --profile :6060

# Then analyze:
go tool pprof http://localhost:6060/debug/pprof/heap
```

---

### Issue 5: Agent Stuck in Loop

**Symptom**:
```
Agent keeps calling same tool repeatedly
Or: Agent keeps routing back and forth
Request takes very long time (minutes)
```

**Causes**:
- Agent doesn't recognize completion signal
- Routing misconfiguration
- `max_rounds` or `max_handoffs` too high

**Solutions**:

**Option A**: Reduce limits
```yaml
# config/crew.yaml
max_handoffs: 3       # Reduce from 5
max_rounds: 5         # Reduce from 10
```

**Option B**: Fix agent signals
Check `config/agents/executor.yaml`:
```yaml
role: |
  ...
  When done, emit: [DONE]  # Critical!
```

**Option C**: Check routing configuration
```yaml
# config/crew.yaml
routing:
  signals:
    executor:
      - signal: "[DONE]"
        target: null      # Must be null to stop
```

**Option D**: Debug agent thinking
```bash
# Enable agent logging
export GO_AGENTIC_LOG_AGENT_THINKING=true
go-agentic --server

# Watch the logs to see what agent is thinking
```

---

### Issue 6: Tool Execution Fails

**Symptom**:
```
event: tool_error
data: {"tool":"GetCPUUsage","error":"tool not found"}
```

**Causes**:
- Tool not registered
- Tool name mismatch
- Tool not available on this OS
- Tool dependencies missing

**Solutions**:

**Option A**: Check tool is registered
```bash
# List available tools
curl http://localhost:8081/tools

# Or check logs for:
# [INFO] Registered tool: GetCPUUsage
```

**Option B**: Check tool name spelling
```yaml
# config/agents/executor.yaml
tools:
  - "GetCPUUsage"      # Correct: CamelCase
  - "GetMemoryUsage"

# Not:
# - "getcpuusage"      # Wrong: lowercase
# - "get_cpu_usage"    # Wrong: snake_case
```

**Option C**: Verify tool is available
```bash
# Check if system command exists
which ping       # For PingHost tool
which nslookup   # For ResolveDNS tool

# Install if missing
# Ubuntu:
sudo apt-get install iputils-ping dnsutils

# macOS:
brew install bind-tools
```

**Option D**: Check tool parameters
```yaml
# If tool needs parameters, pass them correctly
# Example: GetDiskSpace needs path
tool_call: GetDiskSpace("/home")
# Not: GetDiskSpace()  # Missing required param
```

---

### Issue 7: Slow Response Times

**Symptom**:
```
Simple queries take 2-3 seconds
Agent takes long time to respond
```

**Causes**:
- LLM API is slow
- Tools are slow
- Network latency
- Server overloaded

**Solutions**:

**Option A**: Use faster model
```yaml
# config/agents/orchestrator.yaml
model: "gpt-4-turbo"    # Faster than gpt-4o
temperature: 0.1        # More deterministic = faster
max_tokens: 500         # Shorter responses
```

**Option B**: Reduce response verbosity
```yaml
role: |
  Be concise. Respond in 2-3 sentences maximum.
```

**Option C**: Monitor API latency
```bash
# Check OpenAI API status
curl https://status.openai.com/api/v2/status.json | jq

# Or time your requests
time curl -X POST http://localhost:8081/api/crew/stream \
  -H "Content-Type: application/json" \
  -d '{"user_input":"Check CPU"}'
```

**Option D**: Use caching
```yaml
# config/crew.yaml
collect_metrics: true   # Enables cache tracking

# Monitor cache hit rate
curl http://localhost:8081/metrics | jq '.system_metrics.cache_hit_rate'
```

---

### Issue 8: SSL/TLS Certificate Errors

**Symptom**:
```
Error: x509: certificate signed by unknown authority
or
Error: unable to verify the first certificate
```

**Causes**:
- Self-signed certificate
- Invalid certificate chain
- Expired certificate
- Wrong CA bundle

**Solutions**:

**Option A**: Disable certificate verification (dev only!)
```bash
export INSECURE_SKIP_VERIFY=true
go-agentic --server
```

**Option B**: Specify CA certificate
```bash
export CURL_CA_BUNDLE=/path/to/ca-bundle.crt
curl https://localhost:8081/api/crew/stream ...
```

**Option C**: Use HTTP instead of HTTPS (dev only!)
```bash
# In development, use plain HTTP
go-agentic --server --port 8081
```

---

## 🔍 Debug Techniques

### Technique 1: Enable Debug Logging

```bash
# Set log level to debug
export GO_AGENTIC_LOG_LEVEL=debug

# Enable agent thinking logs
export GO_AGENTIC_LOG_AGENT_THINKING=true

# Enable tool execution logs
export GO_AGENTIC_LOG_TOOL_EXECUTION=true

# Start server
go-agentic --server
```

### Technique 2: Stream the Response

```bash
# Save SSE stream to file for analysis
curl -X POST http://localhost:8081/api/crew/stream \
  -H "Content-Type: application/json" \
  -d '{"user_input":"Check CPU"}' > stream.txt

# Analyze
cat stream.txt | grep "event:" | sort | uniq -c
```

### Technique 3: Check Metrics

```bash
# Get current metrics
curl http://localhost:8081/metrics?format=json | jq

# Monitor in real-time (macOS/Linux)
watch -n 1 'curl -s http://localhost:8081/metrics | jq .system_metrics'

# Export to Prometheus format
curl http://localhost:8081/metrics?format=prometheus
```

### Technique 4: Trace Requests

```bash
# Use verbose curl to see headers
curl -vvv -X POST http://localhost:8081/api/crew/stream \
  -H "Content-Type: application/json" \
  -d '{"user_input":"test"}'

# Use tcpdump to see network traffic
sudo tcpdump -i lo0 'tcp port 8081'  # macOS

# Use Wireshark for GUI analysis
wireshark
```

### Technique 5: Validate Configuration

```bash
# Validate YAML syntax
go-agentic --validate config/crew.yaml

# Check for common issues
go-agentic --check-config config/

# Show loaded configuration
go-agentic --show-config
```

---

## 📈 Performance Tuning

### Baseline Metrics

| Metric | Baseline | Good | Excellent |
|--------|----------|------|-----------|
| Avg Response | 1-2s | <1s | <500ms |
| P99 Latency | <5s | <3s | <2s |
| Memory/Request | 2-3MB | <2MB | <1MB |
| CPU/Request | 10-20% | <10% | <5% |
| Throughput | 10 req/s | 20 req/s | 50+ req/s |

### Tuning Parameters

```yaml
# For low-latency scenarios:
performance:
  max_concurrent_requests: 50    # Lower load
  queue_size: 100

timeouts:
  default_tool_timeout: 3        # Stricter
  sequence_timeout: 15

agent_behaviors:
  temperature: 0.1               # Deterministic
  max_tokens: 500                # Shorter responses

# For high-throughput scenarios:
performance:
  max_concurrent_requests: 200   # Handle more
  queue_size: 500

timeouts:
  default_tool_timeout: 10       # More generous
  sequence_timeout: 60

agent_behaviors:
  temperature: 0.7               # Balance
  max_tokens: 2000               # Full responses
```

---

## 🧪 Testing and Validation

### Unit Test a Tool

```bash
# Create test script
cat > test_tool.sh << 'EOF'
#!/bin/bash
query='Check CPU usage'
result=$(curl -s -X POST http://localhost:8081/api/crew/stream \
  -H "Content-Type: application/json" \
  -d "{\"user_input\":\"$query\"}" | grep -c "GetCPUUsage")

if [ $result -gt 0 ]; then
  echo "✅ Tool execution test passed"
else
  echo "❌ Tool execution test failed"
fi
EOF

chmod +x test_tool.sh
./test_tool.sh
```

### Load Testing

```bash
# Simple load test with ab (Apache Bench)
ab -n 100 -c 10 -p payload.json \
   -T application/json \
   http://localhost:8081/api/crew/stream

# Or use wrk (better for SSE)
wrk -t4 -c100 -d30s \
    -s load_test.lua \
    http://localhost:8081/api/crew/stream
```

### Health Check Script

```bash
#!/bin/bash
# Health check for monitoring

health_check() {
  status=$(curl -s http://localhost:8081/health | jq .status)
  if [ "$status" = '"ok"' ]; then
    echo "✅ Server is healthy"
    return 0
  else
    echo "❌ Server is unhealthy"
    return 1
  fi
}

if health_check; then
  exit 0
else
  exit 1
fi
```

---

## 📚 Getting Help

### Check Documentation

1. [Architecture Overview](ARCHITECTURE_OVERVIEW.md) - Understand the system
2. [Configuration Guide](CONFIGURATION_GUIDE.md) - Configuration issues
3. [API Reference](API_REFERENCE.md) - API usage
4. [Deployment Guide](DEPLOYMENT_GUIDE.md) - Deployment issues

### Enable Diagnostic Mode

```bash
# Comprehensive diagnostics
go-agentic --diagnose

# Output includes:
# - System info
# - Configuration validation
# - Dependency checks
# - Port availability
# - API key status
```

### Collect Logs for Support

```bash
# Collect all diagnostic information
go-agentic --diagnose > diagnostics.txt
go-agentic --show-config > config.txt
curl http://localhost:8081/metrics > metrics.json

# Bundle for sharing (exclude secrets)
tar -czf go-agentic-diagnostics.tar.gz \
  diagnostics.txt config.txt metrics.json
```

---

## ✅ Troubleshooting Checklist

Before asking for help:

- [ ] Server is running (`ps aux | grep go-agentic`)
- [ ] Port is accessible (`telnet localhost 8081`)
- [ ] API key is set (`echo $OPENAI_API_KEY`)
- [ ] Configuration is valid (`go-agentic --validate config/`)
- [ ] Logs show no errors (check `GO_AGENTIC_LOG_LEVEL=debug`)
- [ ] Metrics show reasonable values (check `/metrics`)
- [ ] DNS resolution works (`nslookup api.openai.com`)
- [ ] Network connectivity works (`curl https://api.openai.com`)
- [ ] No recent code changes break things
- [ ] Tried restarting the server

---

**Version**: 1.0
**Last Updated**: 2025-12-22
**Status**: Production Ready ✅

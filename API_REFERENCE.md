# 🔌 API Reference - go-agentic

**Status**: Production Ready
**Version**: 1.0
**Last Updated**: 2025-12-22

---

## 🎯 Overview

go-agentic provides a **REST API** for executing agent workflows via HTTP. The API uses **Server-Sent Events (SSE)** to stream real-time execution updates to clients.

### Base URL

```
http://localhost:8081
```

### Authentication

Currently, no authentication is required. For production deployment with authentication:
- Use API gateway with OAuth2
- Or implement middleware authentication layer
- Or use reverse proxy (nginx, Caddy) with auth module

### Content Types

- **Request**: `application/json`
- **Response**: `text/event-stream` (SSE)

---

## 📋 Endpoints

### 1. Stream Crew Execution

Execute agents and stream results in real-time via SSE.

**Endpoint**:
```http
POST /api/crew/stream
```

**Purpose**: Execute a user query through the agent crew, streaming results in real-time

**Request Headers**:
```http
Content-Type: application/json
```

**Request Body**:
```json
{
  "user_input": "Check my system health",
  "model": "gpt-4o",
  "conversation_history": [
    {
      "role": "user",
      "content": "What tools do you have?"
    },
    {
      "role": "assistant",
      "content": "I can check CPU, memory, disk space, etc."
    }
  ]
}
```

**Request Fields**:

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `user_input` | string | Yes | The user's query/command |
| `model` | string | No | LLM model to use (default: `gpt-4o`) |
| `conversation_history` | array | No | Previous messages for context |

**Response**:

Server-Sent Events stream (text/event-stream):

```
event: start
data: {"timestamp": "2025-12-22T10:30:00Z", "request_id": "abc123"}

event: agent_thinking
data: {"agent": "orchestrator", "thinking": "User wants system health check..."}

event: tool_call
data: {"agent": "executor", "tool": "GetCPUUsage", "arguments": {}}

event: tool_result
data: {"tool": "GetCPUUsage", "result": "85%", "duration": 0.123}

event: agent_response
data: {"agent": "executor", "response": "CPU usage is 85%, which is high..."}

event: complete
data: {"status": "success", "total_duration": 1.245, "total_rounds": 3}
```

**Event Types**:

| Event | Payload | Description |
|-------|---------|-------------|
| `start` | `{timestamp, request_id}` | Execution started |
| `agent_thinking` | `{agent, thinking}` | Agent processing |
| `tool_call` | `{agent, tool, arguments}` | About to call tool |
| `tool_result` | `{tool, result, duration}` | Tool result available |
| `tool_error` | `{tool, error}` | Tool execution failed |
| `agent_response` | `{agent, response}` | Agent's final response |
| `agent_error` | `{agent, error}` | Agent execution error |
| `complete` | `{status, total_duration, total_rounds}` | Execution finished |
| `error` | `{error_type, message}` | Critical error |

**Example cURL Request**:

```bash
curl -X POST http://localhost:8081/api/crew/stream \
  -H "Content-Type: application/json" \
  -d '{
    "user_input": "Check my system health",
    "model": "gpt-4o"
  }'
```

**Example JavaScript Client**:

```javascript
const eventSource = new EventSource('/api/crew/stream', {
  method: 'POST',
  headers: {
    'Content-Type': 'application/json'
  },
  body: JSON.stringify({
    user_input: 'Check my system health'
  })
});

eventSource.addEventListener('start', (event) => {
  const data = JSON.parse(event.data);
  console.log('Request started:', data.request_id);
});

eventSource.addEventListener('agent_thinking', (event) => {
  const data = JSON.parse(event.data);
  console.log(`${data.agent} thinking: ${data.thinking}`);
});

eventSource.addEventListener('tool_result', (event) => {
  const data = JSON.parse(event.data);
  console.log(`Tool ${data.tool} returned: ${data.result}`);
});

eventSource.addEventListener('complete', (event) => {
  const data = JSON.parse(event.data);
  console.log('Execution complete:', data.status);
  eventSource.close();
});

eventSource.addEventListener('error', (event) => {
  const data = JSON.parse(event.data);
  console.error('Error:', data.message);
  eventSource.close();
});
```

**Example Python Client**:

```python
import requests
import json

url = 'http://localhost:8081/api/crew/stream'
payload = {
    'user_input': 'Check my system health',
    'model': 'gpt-4o'
}

response = requests.post(url, json=payload, stream=True)

for line in response.iter_lines():
    if line:
        if line.startswith(b'event: '):
            event_type = line.decode().split(': ')[1]
        elif line.startswith(b'data: '):
            event_data = json.loads(line.decode().split(': ')[1])
            print(f'{event_type}: {event_data}')
```

**Example Go Client**:

```go
package main

import (
    "bufio"
    "bytes"
    "encoding/json"
    "fmt"
    "net/http"
)

func main() {
    payload := map[string]interface{}{
        "user_input": "Check my system health",
        "model": "gpt-4o",
    }

    jsonPayload, _ := json.Marshal(payload)
    resp, _ := http.Post(
        "http://localhost:8081/api/crew/stream",
        "application/json",
        bytes.NewBuffer(jsonPayload),
    )
    defer resp.Body.Close()

    scanner := bufio.NewScanner(resp.Body)
    for scanner.Scan() {
        line := scanner.Bytes()
        if bytes.HasPrefix(line, []byte("data: ")) {
            data := line[6:]
            var event map[string]interface{}
            json.Unmarshal(data, &event)
            fmt.Printf("Event data: %v\n", event)
        }
    }
}
```

---

### 2. Health Check

Check if the server is running and healthy.

**Endpoint**:
```http
GET /health
```

**Response**:
```json
{
  "status": "ok",
  "timestamp": "2025-12-22T10:30:00Z",
  "uptime": "2h 30m 45s",
  "version": "1.0"
}
```

**Example**:

```bash
curl http://localhost:8081/health

# Response:
# {
#   "status": "ok",
#   "timestamp": "2025-12-22T10:30:00Z",
#   "uptime": "2h30m45s",
#   "version": "1.0"
# }
```

**Use Case**: Load balancer health checks, readiness probes

---

### 3. Get Metrics

Retrieve system metrics (execution stats, performance data).

**Endpoint**:
```http
GET /metrics?format=json
```

**Query Parameters**:

| Parameter | Values | Default | Description |
|-----------|--------|---------|-------------|
| `format` | `json`, `prometheus` | `json` | Export format |

**Response (JSON Format)**:

```json
{
  "system_metrics": {
    "start_time": "2025-12-22T09:00:00Z",
    "last_updated": "2025-12-22T10:30:00Z",
    "total_requests": 150,
    "successful_requests": 145,
    "failed_requests": 5,
    "total_execution_time": "150s",
    "average_request_time": "1s",
    "memory_usage": 52428800,
    "max_memory_usage": 62914560,
    "cache_hits": 1200,
    "cache_misses": 300,
    "cache_hit_rate": 0.8
  },
  "agent_metrics": {
    "orchestrator": {
      "agent_id": "orchestrator",
      "agent_name": "Intelligent Router",
      "execution_count": 150,
      "success_count": 145,
      "error_count": 5,
      "total_duration": "45s",
      "average_duration": "300ms",
      "min_duration": "100ms",
      "max_duration": "2s",
      "tool_metrics": {}
    },
    "executor": {
      "execution_count": 130,
      "success_count": 125,
      "error_count": 5,
      "tool_metrics": {
        "GetCPUUsage": {
          "execution_count": 130,
          "success_count": 130,
          "error_count": 0,
          "average_duration": "150ms"
        }
      }
    }
  }
}
```

**Response (Prometheus Format)**:

```prometheus
# HELP crew_requests_total Total requests processed
# TYPE crew_requests_total counter
crew_requests_total{status="success"} 145
crew_requests_total{status="error"} 5

# HELP crew_request_duration_seconds Average request duration
# TYPE crew_request_duration_seconds gauge
crew_average_request_duration_seconds 1.000000

# HELP crew_cache_hits_total Total cache hits
# TYPE crew_cache_hits_total counter
crew_cache_hits_total 1200

# HELP crew_cache_misses_total Total cache misses
# TYPE crew_cache_misses_total counter
crew_cache_misses_total 300

# HELP crew_cache_hit_rate Cache hit rate
# TYPE crew_cache_hit_rate gauge
crew_cache_hit_rate 0.800000

# HELP crew_memory_usage_bytes Current memory usage
# TYPE crew_memory_usage_bytes gauge
crew_memory_usage_bytes 52428800

# HELP crew_max_memory_usage_bytes Maximum memory usage
# TYPE crew_max_memory_usage_bytes gauge
crew_max_memory_usage_bytes 62914560
```

**Example**:

```bash
# Get JSON metrics
curl http://localhost:8081/metrics?format=json | jq

# Get Prometheus metrics (for Grafana, Prometheus, etc.)
curl http://localhost:8081/metrics?format=prometheus
```

---

## 🔄 Streaming Examples

### Example 1: Simple Health Check Query

**Request**:
```bash
curl -X POST http://localhost:8081/api/crew/stream \
  -H "Content-Type: application/json" \
  -d '{"user_input": "Check my system health"}'
```

**Expected Stream**:
```
event: start
data: {"timestamp": "2025-12-22T10:30:00Z", "request_id": "abc123"}

event: agent_thinking
data: {"agent": "orchestrator", "thinking": "User wants system diagnostics..."}

event: agent_response
data: {"agent": "orchestrator", "response": "I'll check your system health..."}

event: tool_call
data: {"agent": "executor", "tool": "GetCPUUsage", "arguments": {}}

event: tool_result
data: {"tool": "GetCPUUsage", "result": "45%", "duration": 0.05}

event: tool_call
data: {"agent": "executor", "tool": "GetMemoryUsage", "arguments": {}}

event: tool_result
data: {"tool": "GetMemoryUsage", "result": "62%", "duration": 0.03}

event: agent_response
data: {"agent": "executor", "response": "Your system health is good..."}

event: complete
data: {"status": "success", "total_duration": 0.5, "total_rounds": 2}
```

### Example 2: Query with Conversation History

**Request**:
```bash
curl -X POST http://localhost:8081/api/crew/stream \
  -H "Content-Type: application/json" \
  -d '{
    "user_input": "What about disk space?",
    "model": "gpt-4o",
    "conversation_history": [
      {"role": "user", "content": "Check my system health"},
      {"role": "assistant", "content": "System is healthy..."}
    ]
  }'
```

---

## ⚠️ Error Handling

### Common HTTP Status Codes

| Status | Meaning | Example |
|--------|---------|---------|
| 200 | OK - Stream started | SSE connection established |
| 400 | Bad Request | Invalid JSON, missing fields |
| 404 | Not Found | Unknown endpoint |
| 500 | Internal Server Error | Uncaught exception |
| 503 | Service Unavailable | Server shutting down |

### Error Response Format

When an error occurs in the SSE stream:

```
event: error
data: {
  "error_type": "TOOL_TIMEOUT",
  "message": "Tool 'GetCPUUsage' exceeded timeout (5s)",
  "tool": "GetCPUUsage",
  "timestamp": "2025-12-22T10:30:00Z"
}
```

### Error Types

| Type | Description | Action |
|------|-------------|--------|
| `TOOL_TIMEOUT` | Tool took too long | Retry or skip |
| `TOOL_ERROR` | Tool execution failed | Check logs |
| `AGENT_ERROR` | Agent failed | Check agent config |
| `INVALID_INPUT` | Bad request format | Fix input |
| `MAX_HANDOFFS` | Too many agent transfers | Check routing |
| `NO_AGENT_FOUND` | Agent not found | Check config |
| `INTERNAL_ERROR` | Unexpected error | Check server logs |

---

## 🔐 Security Considerations

### Input Validation

All user inputs are validated:
- Maximum input length: 10,000 characters
- Invalid characters: None (all UTF-8 allowed)
- Injection prevention: Parameterized tool calls

### Rate Limiting (Recommended)

For production, implement rate limiting:

**With nginx**:
```nginx
limit_req_zone $binary_remote_addr zone=crew:10m rate=10r/s;

location /api/crew/stream {
    limit_req zone=crew burst=20;
    proxy_pass http://backend;
}
```

**With reverse proxy gateway**:
```
RateLimit:
  - 10 requests per second per IP
  - Burst: 20 requests
  - Error: 429 Too Many Requests
```

### CORS (Cross-Origin Resource Sharing)

For browser-based clients:

**Using nginx**:
```nginx
add_header 'Access-Control-Allow-Origin' '*';
add_header 'Access-Control-Allow-Methods' 'GET, POST, OPTIONS';
add_header 'Access-Control-Allow-Headers' 'Content-Type';
```

---

## 📊 API Usage Examples

### Example 1: CLI with curl

```bash
#!/bin/bash

# Stream crew execution
curl -X POST http://localhost:8081/api/crew/stream \
  -H "Content-Type: application/json" \
  -d '{
    "user_input": "Is the system healthy?",
    "model": "gpt-4o"
  }' | while IFS= read -r line; do
    if [[ $line == "data: "* ]]; then
      echo "${line#data: }"
    fi
  done
```

### Example 2: Web Browser

```html
<!DOCTYPE html>
<html>
<head>
    <title>go-agentic Client</title>
</head>
<body>
    <h1>System Health Check</h1>
    <div id="output"></div>

    <script>
        function streamQuery(query) {
            const url = 'http://localhost:8081/api/crew/stream';
            const payload = {
                user_input: query,
                model: 'gpt-4o'
            };

            fetch(url, {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json'
                },
                body: JSON.stringify(payload)
            })
            .then(response => response.body.getReader())
            .then(reader => {
                const output = document.getElementById('output');
                const decoder = new TextDecoder();

                function read() {
                    reader.read().then(({done, value}) => {
                        if (done) return;

                        const chunk = decoder.decode(value);
                        if (chunk.includes('data: ')) {
                            const lines = chunk.split('\n');
                            lines.forEach(line => {
                                if (line.startsWith('data: ')) {
                                    const data = JSON.parse(line.slice(6));
                                    output.innerHTML += JSON.stringify(data) + '<br>';
                                }
                            });
                        }
                        read();
                    });
                }
                read();
            });
        }

        // Query on page load
        window.onload = () => {
            streamQuery('Check my system health');
        };
    </script>
</body>
</html>
```

### Example 3: Node.js

```javascript
const http = require('http');

function streamQuery(userInput) {
    const payload = JSON.stringify({
        user_input: userInput,
        model: 'gpt-4o'
    });

    const options = {
        hostname: 'localhost',
        port: 8081,
        path: '/api/crew/stream',
        method: 'POST',
        headers: {
            'Content-Type': 'application/json',
            'Content-Length': payload.length
        }
    };

    const req = http.request(options, (res) => {
        let buffer = '';

        res.on('data', (chunk) => {
            buffer += chunk;
            const lines = buffer.split('\n');

            for (let i = 0; i < lines.length - 1; i++) {
                const line = lines[i];
                if (line.startsWith('data: ')) {
                    const data = JSON.parse(line.slice(6));
                    console.log('Event data:', data);
                }
            }

            buffer = lines[lines.length - 1];
        });

        res.on('end', () => {
            console.log('Stream complete');
        });
    });

    req.on('error', (error) => {
        console.error('Request error:', error);
    });

    req.write(payload);
    req.end();
}

// Execute query
streamQuery('Is my system healthy?');
```

---

## 🚀 Performance Tips

### Optimize Requests

1. **Use specific queries** (not vague)
   - ❌ "Help" (requires clarification)
   - ✅ "Check CPU usage" (direct action)

2. **Include context** in conversation history
   - Reduces back-and-forth
   - Faster execution

3. **Choose right model**
   - gpt-4o: Better quality, slower
   - gpt-4-turbo: Faster, good quality

### Monitor Performance

```bash
# Check metrics regularly
watch -n 5 'curl -s http://localhost:8081/metrics?format=json | jq .system_metrics'
```

### Handle Slow Clients

For slow clients, increase timeout:

```go
// In config:
timeouts:
  sequence_timeout: 60  // Increase from default 30
  default_tool_timeout: 10
```

---

## 📚 Related Documentation

- [Architecture Overview](ARCHITECTURE_OVERVIEW.md) - System design
- [Configuration Guide](CONFIGURATION_GUIDE.md) - API configuration
- [Deployment Guide](DEPLOYMENT_GUIDE.md) - Production setup
- [Troubleshooting Guide](TROUBLESHOOTING_GUIDE.md) - Common issues

---

**Version**: 1.0
**Last Updated**: 2025-12-22
**Status**: Production Ready ✅

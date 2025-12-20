# go-crewai Streaming Guide

## 📡 Server-Sent Events (SSE) Implementation

This guide explains how to use the new **SSE streaming** feature in go-crewai for real-time crew execution feedback.

---

## 🚀 Quick Start

### Start the HTTP Server

```bash
cd go-crewai
go run ./cmd/main.go --server --port 8080
```

**Output:**
```
🚀 HTTP Server starting on http://localhost:8080
📡 SSE Endpoint: http://localhost:8080/api/crew/stream
🌐 Web Client: http://localhost:8080
```

### Access Web Client

Open your browser: **http://localhost:8080**

### Test with curl

```bash
curl -X POST http://localhost:8080/api/crew/stream \
  -H "Content-Type: application/json" \
  -d '{"query": "Tôi không vào được Internet"}'
```

---

## 📊 Streaming Events

### Event Types

| Event Type | Description | Example |
|------------|-------------|---------|
| `start` | Stream initialization | `"🚀 Starting crew execution..."` |
| `agent_start` | Before agent execution | `"🔄 Starting Orchestrator..."` |
| `agent_response` | Agent's response | `"Tôi sẽ chuyển sang..."` |
| `tool_start` | Before tool execution | `"🔧 [Tool] PingHost → Executing..."` |
| `tool_result` | Tool execution result | `"✅ [Tool] PingHost → Success"` |
| `pause` | Waiting for user input | `"[PAUSE] Waiting for user input"` |
| `done` | Execution completed | `"✅ Execution completed"` |
| `error` | Error occurred | `"❌ Error: Agent failed"` |
| `ping` | Keep-alive (every 30s) | `""` |

### Event Structure

Each event is sent as SSE format:

```
data: {"type":"agent_start","agent":"Orchestrator","content":"🔄 Starting Orchestrator...","timestamp":"2025-12-19T10:30:00Z","metadata":null}

```

**JSON Fields:**
- `type` (string): Event type
- `agent` (string): Agent name/ID
- `content` (string): Main message
- `timestamp` (ISO 8601): When event occurred
- `metadata` (object): Additional data (optional)

---

## 💻 JavaScript Client Example

### Using EventSource API

```javascript
// Send request
const payload = {
    query: "Tôi không vào được Internet",
    history: []  // For continuation after pause
};

// Open SSE connection
const eventSource = new EventSource(
    '/api/crew/stream?q=' + encodeURIComponent(JSON.stringify(payload))
);

// Handle events
eventSource.onmessage = (event) => {
    const streamEvent = JSON.parse(event.data);

    switch(streamEvent.type) {
        case 'agent_start':
            console.log('🔄', streamEvent.content);
            break;
        case 'agent_response':
            console.log('💬', streamEvent.agent + ':', streamEvent.content);
            break;
        case 'tool_result':
            console.log('✅', streamEvent.content);
            break;
        case 'pause':
            console.log('⏸️ Waiting for user input');
            // User should provide next input here
            break;
        case 'done':
            console.log('✅ Done');
            eventSource.close();
            break;
    }
};

// Handle errors
eventSource.onerror = (error) => {
    console.error('Connection error:', error);
    eventSource.close();
};
```

---

## 🔄 Conversation Flow with Pause/Resume

### Scenario: Vague Problem requiring Clarification

**Request 1:** User asks vague question
```bash
curl -X POST http://localhost:8080/api/crew/stream \
  -H "Content-Type: application/json" \
  -d '{
    "query": "Tôi không vào được Internet",
    "history": []
  }'
```

**Stream Response:**
```
data: {"type":"agent_start","agent":"Orchestrator",...}
data: {"type":"agent_response","agent":"Orchestrator","content":"Tôi sẽ chuyển sang Ngân...",...}
data: {"type":"agent_start","agent":"Clarifier",...}
data: {"type":"agent_response","agent":"Clarifier","content":"Bạn đang cố kết nối từ máy nào?",...}
data: {"type":"pause","agent":"Clarifier","content":"[PAUSE] Waiting for user input",...}
```

**Stream closes** ↔ User sees question, provides answer

**Request 2:** User provides clarification with history
```bash
curl -X POST http://localhost:8080/api/crew/stream \
  -H "Content-Type: application/json" \
  -d '{
    "query": "Máy 192.168.1.101, không ping được 8.8.8.8",
    "history": [
      {"role": "user", "content": "Tôi không vào được Internet"},
      {"role": "assistant", "content": "Tôi sẽ chuyển sang Ngân..."},
      {"role": "assistant", "content": "Bạn đang cố kết nối từ máy nào?"}
    ]
  }'
```

**Stream Response:**
```
data: {"type":"agent_start","agent":"Executor",...}
data: {"type":"tool_start",...}
data: {"type":"tool_result",...}
data: {"type":"agent_response","agent":"Executor","content":"DIAGNOSIS: Network interface DOWN",...}
data: {"type":"done",...}
```

**Stream closes** ↔ Conversation complete

---

## 🛠️ API Endpoints

### POST /api/crew/stream

**Description:** Execute crew with SSE streaming

**Request:**
```json
{
  "query": "Your IT support question",
  "history": [
    {"role": "user", "content": "Previous message"},
    {"role": "assistant", "content": "Agent response"}
  ]
}
```

**Response:**
- Content-Type: `text/event-stream`
- Cache-Control: `no-cache`
- Connection: `keep-alive`
- Streaming SSE events until completion or pause

**Example with curl:**
```bash
curl -X POST http://localhost:8080/api/crew/stream \
  -H "Content-Type: application/json" \
  -d '{"query":"Your question here"}'
```

### GET /health

**Description:** Health check endpoint

**Response:**
```json
{
  "status": "ok",
  "service": "go-crewai-streaming"
}
```

### GET /

**Description:** Web client HTML interface

**Response:** Interactive streaming client UI

---

## 🎯 Features & Limitations

### ✅ Supported Features

- Real-time SSE streaming of agent execution
- Tool execution progress tracking
- Conversation history preservation
- Pause/resume flow (Clarifier waits for input)
- Error handling and reporting
- Keep-alive pings every 30 seconds
- Cross-origin requests (CORS enabled)
- Both JSON and plain text formats
- Built-in HTML web client

### ⚠️ Limitations

- **One-way communication** (server → client only)
  - Use separate POST requests for client → server
  - No real-time bidirectional streaming

- **No automatic reconnect**
  - Client responsible for reconnection logic
  - Implement exponential backoff for retries

- **30-second timeout**
  - Long-running operations may need keep-alive handling
  - Configure via timeout constants in http.go

- **Browser compatibility**
  - Requires EventSource API (IE11+, all modern browsers)
  - Not suitable for very old browsers

---

## 🧪 Testing

### Manual Testing

1. **Start server:**
   ```bash
   go run ./cmd/main.go --server --port 8080
   ```

2. **Open web client:**
   - Go to http://localhost:8080
   - Type: "Máy chậm lắm"
   - Watch real-time progress

3. **Test with curl:**
   ```bash
   curl -X POST http://localhost:8080/api/crew/stream \
     -H "Content-Type: application/json" \
     -d '{"query":"Test question"}'
   ```

4. **Check health:**
   ```bash
   curl http://localhost:8080/health
   ```

### Test Scenarios

**Scenario 1: Simple Question**
```
Query: "Server 192.168.1.50 không ping được, check cho tôi"
Expected: Orchestrator → Executor (direct routing)
Stream: Multiple tool executions with results
```

**Scenario 2: Vague Question**
```
Query: "Máy chậm lắm"
Expected: Orchestrator → Clarifier → pause
Stream: Questions, then pause event
```

**Scenario 3: Error Handling**
```
Query: (with invalid OPENAI_API_KEY)
Expected: Error event in stream
Stream: "❌ Error: ..." then stream closes
```

---

## 🔧 Configuration

### Command Line Flags

```bash
go run ./cmd/main.go --server --port 8080
```

- `--server` (bool): Enable HTTP server mode (default: false - CLI mode)
- `--port` (int): HTTP server port (default: 8080)

### Environment Variables

```bash
export OPENAI_API_KEY="sk-..."        # Required
export CREWAI_CONFIG_DIR="./config"   # Optional
```

### Code Configuration

Edit constants in `http.go`:
- `30 * time.Second` - Keep-alive ping interval
- `10` - Stream channel buffer size
- `8080` - Default port

---

## 📝 Example: Building a Custom Client

### Minimal HTML5 Client

```html
<!DOCTYPE html>
<html>
<body>
    <input id="query" type="text" placeholder="Enter query...">
    <button onclick="sendQuery()">Send</button>
    <pre id="output"></pre>

    <script>
        function sendQuery() {
            const query = document.getElementById('query').value;
            const eventSource = new EventSource(
                '/api/crew/stream?q=' + encodeURIComponent(JSON.stringify({query}))
            );

            eventSource.onmessage = (e) => {
                const event = JSON.parse(e.data);
                document.getElementById('output').textContent +=
                    `[${event.type}] ${event.content}\n`;
            };

            eventSource.onerror = () => eventSource.close();
        }
    </script>
</body>
</html>
```

---

## 🔗 Integration with Existing Code

### CLI Mode (Unchanged)

Still works as before:
```bash
go run ./cmd/main.go
```

### New HTTP Mode

```bash
go run ./cmd/main.go --server
```

### Both Modes

- Share the same crew executor
- Share the same configuration
- Reuse existing agent/tool implementations
- No breaking changes to existing code

---

## 📚 File Structure

```
go-crewai/
├── streaming.go          (NEW) Streaming utilities
├── http.go              (NEW) HTTP server & handlers
├── html_client.go       (NEW) Frontend HTML template
├── crew.go              (MODIFIED) Added ExecuteStream()
├── types.go             (MODIFIED) Added StreamEvent struct
├── cmd/main.go          (MODIFIED) Added --server flag
└── STREAMING_GUIDE.md   (THIS FILE)
```

---

## 🐛 Troubleshooting

### Port Already in Use

```bash
# Use different port
go run ./cmd/main.go --server --port 9000
```

### OPENAI_API_KEY Not Set

```bash
# Set environment variable
export OPENAI_API_KEY="sk-..."
go run ./cmd/main.go --server
```

### EventSource Connection Failed

1. Check server is running: `curl http://localhost:8080/health`
2. Check CORS headers in response
3. Try with different query format
4. Check browser console for errors

### Stream Closes Unexpectedly

1. Check agent WaitForSignal configuration
2. Verify conversation history format
3. Check for errors in stream (look for error events)
4. Verify OpenAI API key validity

---

## 📞 Support

For issues or questions:
1. Check this guide first
2. Review technical spec: `tech-spec-sse-streaming.md`
3. Check implementation code in `streaming.go`, `http.go`
4. Test with curl before debugging JavaScript

---

**Version:** 1.0
**Last Updated:** 2025-12-19
**Status:** Production Ready ✅

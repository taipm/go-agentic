# 🚀 Quick Start - SSE Streaming Demos

**Status:** ✅ Ready to Use
**Port:** 8081
**Server:** http://localhost:8081

---

## Start the Server

```bash
cd go-crewai
./crewai-server --server --port 8081

# Or with go run
go run ./cmd/main.go --server --port 8081
```

You should see:
```
🚀 HTTP Server starting on http://localhost:8081
📡 SSE Endpoint: http://localhost:8081/api/crew/stream
🌐 Web Client: http://localhost:8081
```

---

## Try Demo (Pick One)

### Option 1️⃣: Web Browser (Easiest) ⭐

```bash
# Open in browser
open http://localhost:8081

# Or navigate manually to: http://localhost:8081
```

**What to do:**
1. Type query: `Máy chậm lắm`
2. Click: `Send Query`
3. Watch real-time events stream in!

---

### Option 2️⃣: Interactive Script (Recommended) ⭐⭐

```bash
cd go-crewai
export TERM=xterm
./demo.sh

# Menu will appear with 6 different demo scenarios
# Choose one and watch it run!
```

---

### Option 3️⃣: curl Command

**Simple query:**
```bash
curl -X POST http://localhost:8081/api/crew/stream \
  -H "Content-Type: application/json" \
  -d '{"query":"Máy chậm lắm"}'
```

**With conversation history:**
```bash
curl -X POST http://localhost:8081/api/crew/stream \
  -H "Content-Type: application/json" \
  -d '{
    "query":"Ubuntu 192.168.1.101 không ping được 8.8.8.8",
    "history":[
      {"role":"user","content":"Tôi không vào được Internet"},
      {"role":"assistant","content":"Tôi sẽ chuyển sang Clarifier..."},
      {"role":"assistant","content":"Bạn đang cố kết nối từ máy nào?"}
    ]
  }'
```

---

## Demo Scenarios

| Scenario | Query | Expected Result |
|----------|-------|-----------------|
| **Machine Slow** | `Máy chậm lắm` | Orchestrator → Executor |
| **Network Issue** | `Server 192.168.1.50 không ping được` | Direct problem routing |
| **Vague Question** | `Tôi không vào được Internet` | Orchestrator → Clarifier → PAUSE |
| **Resume Flow** | Previous query + history | Clarifier → Executor |

---

## Health Check

```bash
curl http://localhost:8081/health
```

Expected response:
```json
{"service":"go-crewai-streaming","status":"ok"}
```

---

## Event Types

When streaming, you'll see these event types:

| Type | Icon | Meaning |
|------|------|---------|
| `start` | 🚀 | Execution started |
| `agent_start` | 🔄 | Agent starting |
| `agent_response` | 💬 | Agent response |
| `tool_start` | 🔧 | Tool execution started |
| `tool_result` | ✅ | Tool result |
| `pause` | ⏸️ | Waiting for input |
| `done` | ✅ | Completed |
| `error` | ❌ | Error occurred |

---

## Real-Time Viewing

### View raw SSE stream
```bash
curl -s http://localhost:8081/api/crew/stream \
  -d '{"query":"Test"}' | \
  while IFS= read -r line; do
    echo "$(date '+%H:%M:%S') $line"
  done
```

### Save to file
```bash
curl -s http://localhost:8081/api/crew/stream \
  -d '{"query":"Test"}' > events.log

cat events.log
```

### Pretty print with jq
```bash
curl -s http://localhost:8081/api/crew/stream \
  -d '{"query":"Test"}' | \
  grep 'data:' | \
  sed 's/data: //' | \
  jq '.'
```

---

## Files Included

```
go-crewai/
├── QUICKSTART.md              ← You are here
├── DEMO_QUICK_START.md        - Quick start guide
├── DEMO_README.md             - Complete guide
├── DEMO_EXAMPLES.md           - Detailed examples
├── FIX_VERIFICATION.md        - Technical fix details
├── demo.sh                    - Interactive demo script
├── test_sse_client.html       - Web client (auto-served)
├── STREAMING_GUIDE.md         - Full API reference
├── DEPLOYMENT_CHECKLIST.md    - Deployment steps
└── tech-spec-sse-streaming.md - Technical specification
```

---

## Troubleshooting

### Server won't start
```bash
# Check if port 8081 is in use
lsof -i :8081

# Kill any existing process
pkill -f crewai-server

# Try different port
./crewai-server --server --port 9000
```

### OPENAI_API_KEY not set
```bash
export OPENAI_API_KEY="sk-..."
./crewai-server --server --port 8081
```

### EventSource connection failed
```bash
# Verify server health
curl http://localhost:8081/health

# Check server logs
tail -f /tmp/server.log
```

### jq not installed
```bash
# macOS
brew install jq

# Ubuntu/Debian
sudo apt-get install jq

# Or just use web client instead
open http://localhost:8081
```

---

## API Endpoints

| Endpoint | Method | Purpose |
|----------|--------|---------|
| `/api/crew/stream` | POST, GET | Stream crew execution |
| `/health` | GET | Server health check |
| `/` | GET | Web client UI |

---

## Next Steps

1. ✅ Run server: `./crewai-server --server --port 8081`
2. ✅ Try web client: `open http://localhost:8081`
3. ✅ Run demo script: `./demo.sh`
4. ✅ Read STREAMING_GUIDE.md for API details
5. ✅ Integrate into your application

---

## Need Help?

- **Quick ref:** DEMO_QUICK_START.md
- **Full guide:** DEMO_README.md
- **Examples:** DEMO_EXAMPLES.md
- **API docs:** STREAMING_GUIDE.md
- **Deployment:** DEPLOYMENT_CHECKLIST.md
- **Technical:** tech-spec-sse-streaming.md

---

**Ready to demo?** Pick Option 1, 2, or 3 above and get started! 🎉

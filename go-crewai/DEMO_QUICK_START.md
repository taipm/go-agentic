# 🎬 Demo Quick Start

**Hướng dẫn nhanh để test SSE Streaming**

---

## 🚀 Step 1: Khởi động Server

### Terminal 1:

```bash
cd go-crewai
go run ./cmd/main.go --server --port 8081
```

**Kết quả mong đợi:**

```
🚀 HTTP Server starting on http://localhost:8081
📡 SSE Endpoint: http://localhost:8081/api/crew/stream
🌐 Web Client: http://localhost:8081
```

✅ Server sẵn sàng!

---

## 🎯 Step 2: Chọn cách test

### Cách 1️⃣: Web Browser (Dễ nhất) ⭐

```bash
# Mở trình duyệt
open http://localhost:8081

# Hoặc
firefox http://localhost:8081
```

**Cách dùng:**
1. Nhập: `Máy chậm lắm`
2. Click: `Send`
3. Xem: Real-time events

**✅ Demo hoàn tất!**

---

### Cách 2️⃣: Interactive Demo Script (Dễ nhất + Chi tiết)

```bash
# Terminal 2:
cd go-crewai
chmod +x demo.sh
./demo.sh
```

**Menu interactive:**
```
1️⃣  Simple Query - Machine Slow
2️⃣  Network Issue (Direct Problem)
3️⃣  Vague Question (Pause/Resume)
4️⃣  Resume with Clarification
5️⃣  Load Test
6️⃣  Health Check
7️⃣  Run All Demos
8️⃣  Open Web Client
9️⃣  Exit
```

**✅ Chọn demo và xem kết quả!**

---

### Cách 3️⃣: curl Commands (Untuk Technical)

**Simple Demo:**

```bash
curl -X POST http://localhost:8081/api/crew/stream \
  -H "Content-Type: application/json" \
  -d '{"query":"Máy chậm lắm"}'
```

**Pause/Resume Demo:**

```bash
# Step 1: Trigger pause
curl -X POST http://localhost:8081/api/crew/stream \
  -H "Content-Type: application/json" \
  -d '{"query":"Tôi không vào được Internet"}'
```

**Step 2: Resume dengan history**

```bash
curl -X POST http://localhost:8081/api/crew/stream \
  -H "Content-Type: application/json" \
  -d '{
    "query":"Máy 192.168.1.101, Ubuntu, không ping được 8.8.8.8",
    "history":[
      {"role":"user","content":"Tôi không vào được Internet"},
      {"role":"assistant","content":"Tôi sẽ chuyển sang Clarifier..."},
      {"role":"assistant","content":"Bạn đang cố kết nối từ máy nào?"}
    ]
  }'
```

**✅ Xem streaming events!**

---

## 📋 Demo Scenarios Khác

### Scenario A: Direct Problem → Executor

```bash
curl -X POST http://localhost:8081/api/crew/stream \
  -H "Content-Type: application/json" \
  -d '{"query":"Server 192.168.1.50 không ping được"}'
```

**Kỳ vọng:** Orchestrator → Executor (direct routing)

---

### Scenario B: Vague Problem → Clarifier → Pause

```bash
curl -X POST http://localhost:8081/api/crew/stream \
  -H "Content-Type: application/json" \
  -d '{"query":"Network bị vấn đề"}'
```

**Kỳ vọng:** Orchestrator → Clarifier → [PAUSE]

---

### Scenario C: Tool Execution

```bash
curl -X POST http://localhost:8081/api/crew/stream \
  -H "Content-Type: application/json" \
  -d '{"query":"Check disk space máy 192.168.1.100"}'
```

**Kỳ vọng:** Tool execution events (tool_start, tool_result)

---

## 🔍 Xem Chi Tiết Events

### Pretty Print với jq

```bash
curl -X POST http://localhost:8081/api/crew/stream \
  -H "Content-Type: application/json" \
  -d '{"query":"Máy chậm"}' | grep 'data:' | sed 's/data: //' | jq '.'
```

---

## ✅ Verification

### Check Server Health

```bash
curl http://localhost:8081/health
```

**Kết quả:**
```json
{"status":"ok","service":"go-crewai-streaming"}
```

---

## 📚 Tài Liệu Chi Tiết

- **DEMO_EXAMPLES.md** - Đầy đủ các ví dụ
- **STREAMING_GUIDE.md** - Hướng dẫn toàn bộ
- **DEPLOYMENT_CHECKLIST.md** - Deployment procedures
- **INDEX.md** - Navigation guide

---

## 🎓 Hiểu Events

### Event Types

| Type | Icon | Ý Nghĩa |
|------|------|---------|
| `start` | 🚀 | Bắt đầu execution |
| `agent_start` | 🔄 | Bắt đầu agent gọi API |
| `agent_response` | 💬 | Phản hồi từ agent |
| `tool_start` | 🔧 | Bắt đầu tool execution |
| `tool_result` | ✅ | Kết quả tool execution |
| `pause` | ⏸️ | Chờ input user |
| `done` | ✅ | Hoàn tất |
| `error` | ❌ | Lỗi |

---

## 🚨 Troubleshooting

### Server không chạy

```bash
# Kiểm tra port
lsof -i :8081

# Dùng port khác
go run ./cmd/main.go --server --port 9000
```

### OPENAI_API_KEY không set

```bash
export OPENAI_API_KEY="sk-..."
go run ./cmd/main.go --server --port 8081
```

### EventSource connection failed

```bash
# Verify server health
curl http://localhost:8081/health

# Check headers
curl -v http://localhost:8081/health
```

---

## 📊 Kết Quả Mong Đợi

### Scenario 1: Máy chậm
```
🚀 Starting crew execution...
🔄 Starting Orchestrator...
💬 Orchestrator: Tôi sẽ chuyển sang Executor...
🔄 Starting Executor...
🔧 [Tool] DiskSpaceCheck → Executing...
✅ [Tool] DiskSpaceCheck → Success
💬 Executor: Ổ cứng 95% đầy...
✅ Execution completed
```

### Scenario 2: Không vào Internet
```
🚀 Starting crew execution...
🔄 Starting Orchestrator...
💬 Orchestrator: Tôi sẽ chuyển sang Clarifier...
🔄 Starting Clarifier...
💬 Clarifier: Bạn đang cố kết nối từ máy nào?
⏸️ [PAUSE] Waiting for user input
```

---

## 💡 Tips & Tricks

### Xem real-time logs

```bash
# Terminal 1
go run ./cmd/main.go --server 2>&1 | tee server.log

# Terminal 2
tail -f server.log | grep -E "Starting|agent|Tool"
```

### Test concurrent requests

```bash
for i in {1..5}; do
  curl -X POST http://localhost:8081/api/crew/stream \
    -H "Content-Type: application/json" \
    -d "{\"query\":\"Test $i\"}" &
done
wait
```

### Save events to file

```bash
curl -X POST http://localhost:8081/api/crew/stream \
  -H "Content-Type: application/json" \
  -d '{"query":"Máy chậm"}' > events.log

cat events.log | grep 'data:' | sed 's/data: //' | jq '.'
```

---

## 🎯 Next Steps

1. ✅ Chạy demo với web browser
2. ✅ Thử script `./demo.sh`
3. ✅ Test curl commands
4. ✅ Đọc STREAMING_GUIDE.md để hiểu sâu
5. ✅ Tích hợp vào ứng dụng của bạn

---

## 📞 Support

- **Web Client:** http://localhost:8081
- **API Endpoint:** http://localhost:8081/api/crew/stream
- **Health Check:** http://localhost:8081/health
- **Documentation:** STREAMING_GUIDE.md

---

**Ready to demo? Start with:**

```bash
# Terminal 1: Start server
cd go-crewai
go run ./cmd/main.go --server --port 8081

# Terminal 2: Run interactive demo
cd go-crewai
./demo.sh
```

**✅ Enjoy! 🎉**

---

**Version:** 1.0
**Date:** 2025-12-19
**Status:** Ready to Use ✅

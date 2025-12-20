# 🎬 SSE Streaming Demo Examples

**Bản hướng dẫn:** Các ví dụ demo thực tế để test SSE streaming
**Ngôn ngữ:** Tiếng Việt (Vietnamese)
**Ngày:** 2025-12-19

---

## 📋 Mục Lục

1. [Khởi động Server](#khởi-động-server)
2. [Demo 1: Web Client đơn giản](#demo-1-web-client-đơn-giản)
3. [Demo 2: curl test Scenario 1 - Máy chậm](#demo-2-curl-test-scenario-1)
4. [Demo 3: curl test Scenario 2 - Không vào Internet](#demo-3-curl-test-scenario-2)
5. [Demo 4: curl test Scenario 3 - Vague Question (Pause/Resume)](#demo-4-curl-test-scenario-3)
6. [Demo 5: JavaScript Client](#demo-5-javascript-client)
7. [Demo 6: PowerShell Demo (Windows)](#demo-6-powershell-demo-windows)
8. [Demo 7: Monitoring & Logging](#demo-7-monitoring--logging)

---

## 🚀 Khởi động Server

### Bước 1: Đặt OPENAI_API_KEY

```bash
# macOS/Linux
export OPENAI_API_KEY="sk-..."

# hoặc từ file .env
source .env
```

### Bước 2: Khởi động HTTP Server

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

### Bước 3: Verify Server đang chạy

```bash
curl http://localhost:8081/health
```

**Kết quả:**
```json
{"status":"ok","service":"go-crewai-streaming"}
```

✅ Server sẵn sàng!

---

## Demo 1: Web Client đơn giản

### Cách 1: Mở trình duyệt

```
1. Mở: http://localhost:8081
2. Nhập: "Máy chậm lắm"
3. Click: Send
4. Xem: Real-time streaming events
```

### Kết quả mong đợi

```
🚀 Starting crew execution...
🔄 Starting Orchestrator... [orchestrator]
💬 Agent (Orchestrator): Tôi sẽ chuyển sang Executor...
🔄 Starting Executor... [executor]
🔧 [Tool] DiskSpaceCheck → Executing...
✅ [Tool] DiskSpaceCheck → Success
💬 Agent (Executor): Tìm thấy ổ cứng 95% đầy...
✅ Execution completed
```

---

## Demo 2: curl test Scenario 1

### Scenario: Machine chậm - đơn giản

**Mô tả:** User hỏi máy chậm. Orchestrator trực tiếp gửi tới Executor để kiểm tra.

**Command:**

```bash
curl -X POST http://localhost:8081/api/crew/stream \
  -H "Content-Type: application/json" \
  -d '{"query":"Máy chậm lắm"}'
```

**Kết quả mong đợi:**

```
data: {"type":"start","agent":"system","content":"🚀 Starting crew execution...","timestamp":"2025-12-19T10:00:00Z","metadata":null}

data: {"type":"agent_start","agent":"Orchestrator","content":"🔄 Starting Orchestrator...","timestamp":"2025-12-19T10:00:01Z","metadata":null}

data: {"type":"agent_response","agent":"Orchestrator","content":"Tôi sẽ chuyển sang Executor để kiểm tra tài nguyên hệ thống...","timestamp":"2025-12-19T10:00:02Z","metadata":null}

data: {"type":"agent_start","agent":"Executor","content":"🔄 Starting Executor...","timestamp":"2025-12-19T10:00:03Z","metadata":null}

data: {"type":"tool_start","agent":"Executor","content":"🔧 [Tool] DiskSpaceCheck → Executing...","timestamp":"2025-12-19T10:00:04Z","metadata":null}

data: {"type":"tool_result","agent":"Executor","content":"✅ [Tool] DiskSpaceCheck → Success","timestamp":"2025-12-19T10:00:05Z","metadata":null}

data: {"type":"tool_start","agent":"Executor","content":"🔧 [Tool] MemoryCheck → Executing...","timestamp":"2025-12-19T10:00:06Z","metadata":null}

data: {"type":"tool_result","agent":"Executor","content":"✅ [Tool] MemoryCheck → Success","timestamp":"2025-12-19T10:00:07Z","metadata":null}

data: {"type":"agent_response","agent":"Executor","content":"DIAGNOSIS: Ổ cứng 95% đầy, bộ nhớ 80% sử dụng. Khuyến cáo xóa file cũ hoặc nâng cấp SSD.","timestamp":"2025-12-19T10:00:08Z","metadata":null}

data: {"type":"done","agent":"system","content":"✅ Execution completed","timestamp":"2025-12-19T10:00:09Z","metadata":null}
```

---

## Demo 3: curl test Scenario 2

### Scenario: Không vào Internet - Clarifier hỏi

**Mô tả:** User hỏi không vào được Internet (vague). Orchestrator gửi tới Clarifier để hỏi thêm.

**Command:**

```bash
curl -X POST http://localhost:8081/api/crew/stream \
  -H "Content-Type: application/json" \
  -d '{"query":"Tôi không vào được Internet"}'
```

**Kết quả mong đợi:**

```
data: {"type":"start","agent":"system","content":"🚀 Starting crew execution..."}

data: {"type":"agent_start","agent":"Orchestrator","content":"🔄 Starting Orchestrator..."}

data: {"type":"agent_response","agent":"Orchestrator","content":"Tôi sẽ chuyển sang Clarifier để làm rõ vấn đề..."}

data: {"type":"agent_start","agent":"Clarifier","content":"🔄 Starting Clarifier..."}

data: {"type":"agent_response","agent":"Clarifier","content":"Bạn đang cố kết nối từ máy nào? (Windows/Mac/Linux)"}

data: {"type":"pause","agent":"Clarifier","content":"[PAUSE] Waiting for user input"}
```

### Bước tiếp theo: User trả lời

User nhìn thấy câu hỏi từ Clarifier và trả lời. Gửi request mới với history:

```bash
curl -X POST http://localhost:8081/api/crew/stream \
  -H "Content-Type: application/json" \
  -d '{
    "query":"Máy 192.168.1.101, Ubuntu Linux, không ping được 8.8.8.8",
    "history":[
      {"role":"user","content":"Tôi không vào được Internet"},
      {"role":"assistant","content":"Tôi sẽ chuyển sang Clarifier..."},
      {"role":"assistant","content":"Bạn đang cố kết nối từ máy nào?"}
    ]
  }'
```

**Kết quả mong đợi:**

```
data: {"type":"agent_start","agent":"Executor","content":"🔄 Starting Executor..."}

data: {"type":"tool_start","agent":"Executor","content":"🔧 [Tool] PingHost → Executing..."}

data: {"type":"tool_result","agent":"Executor","content":"✅ [Tool] PingHost → Failed"}

data: {"type":"tool_start","agent":"Executor","content":"🔧 [Tool] NetworkDiagnostics → Executing..."}

data: {"type":"tool_result","agent":"Executor","content":"✅ [Tool] NetworkDiagnostics → IP config OK, Gateway không response"}

data: {"type":"agent_response","agent":"Executor","content":"DIAGNOSIS: Gateway không phản hồi. Kiểm tra: 1) Kết nối Ethernet, 2) Restart router"}

data: {"type":"done","agent":"system","content":"✅ Execution completed"}
```

---

## Demo 4: curl test Scenario 3

### Scenario: Vague Question - Pause/Resume Flow

**Mô tả:** Test đầy đủ pause/resume flow

**Step 1: User hỏi vague question**

```bash
curl -X POST http://localhost:8081/api/crew/stream \
  -H "Content-Type: application/json" \
  -d '{"query":"Network bị vấn đề"}'
```

**Kết quả:** Stream tạm dừng ở câu hỏi của Clarifier

**Step 2: User trả lời**

```bash
curl -X POST http://localhost:8081/api/crew/stream \
  -H "Content-Type: application/json" \
  -d '{
    "query":"Windows 10, có Internet nhưng chậm lắm",
    "history":[
      {"role":"user","content":"Network bị vấn đề"},
      {"role":"assistant","content":"Tôi sẽ chuyển sang Clarifier..."},
      {"role":"assistant","content":"Hệ điều hành nào? Internet có hoạt động không?"}
    ]
  }'
```

**Kết quả:** Executor kiểm tra và diagnose

**Step 3 (Optional): User hỏi thêm**

Nếu cần thêm info, lặp lại step 2 với history mới

---

## Demo 5: JavaScript Client

### Tạo HTML file để test JavaScript

Tạo file `test_streaming.html`:

```html
<!DOCTYPE html>
<html>
<head>
    <title>SSE Streaming Test</title>
    <style>
        body {
            font-family: 'Courier New', monospace;
            max-width: 800px;
            margin: 0 auto;
            padding: 20px;
            background: #f5f5f5;
        }
        #input-section {
            background: white;
            padding: 20px;
            border-radius: 8px;
            margin-bottom: 20px;
        }
        input {
            width: 100%;
            padding: 10px;
            font-size: 14px;
            border: 1px solid #ddd;
            border-radius: 4px;
            margin-bottom: 10px;
        }
        button {
            background: #4CAF50;
            color: white;
            padding: 10px 20px;
            border: none;
            border-radius: 4px;
            cursor: pointer;
            font-size: 14px;
        }
        button:hover {
            background: #45a049;
        }
        #output {
            background: white;
            border: 1px solid #ddd;
            border-radius: 8px;
            padding: 20px;
            height: 400px;
            overflow-y: auto;
            margin-bottom: 20px;
        }
        .event {
            padding: 8px;
            margin: 5px 0;
            border-left: 3px solid #ddd;
            padding-left: 10px;
        }
        .agent_start {
            border-left-color: #FFA500;
            color: #FF6347;
        }
        .agent_response {
            border-left-color: #4CAF50;
            color: #2E7D32;
        }
        .tool_start {
            border-left-color: #2196F3;
            color: #1565C0;
        }
        .tool_result {
            border-left-color: #4CAF50;
            color: #2E7D32;
        }
        .pause {
            border-left-color: #FFC107;
            background: #FFF8DC;
            color: #F57F17;
        }
        .error {
            border-left-color: #F44336;
            color: #C62828;
        }
        .done {
            border-left-color: #4CAF50;
            color: #1B5E20;
            font-weight: bold;
        }
        .ping {
            color: #999;
            font-size: 12px;
        }
        .start {
            border-left-color: #4CAF50;
            color: #1B5E20;
            font-weight: bold;
        }
    </style>
</head>
<body>
    <h1>🚀 SSE Streaming Demo</h1>

    <div id="input-section">
        <h2>Submit Query</h2>
        <input id="query" type="text" placeholder="Nhập câu hỏi IT support..."
               value="Máy chậm lắm">
        <button onclick="sendQuery()">Send Query</button>
        <button onclick="clearOutput()" style="background: #f44336; margin-left: 10px;">Clear</button>
    </div>

    <h2>📡 Streaming Events</h2>
    <div id="output"></div>

    <script>
        let eventSource = null;
        let history = [];

        function sendQuery() {
            const query = document.getElementById('query').value;
            if (!query.trim()) {
                alert('Vui lòng nhập câu hỏi');
                return;
            }

            if (eventSource) {
                eventSource.close();
            }

            clearOutput();

            const payload = {
                query: query,
                history: history
            };

            // Gửi request
            const url = '/api/crew/stream?q=' + encodeURIComponent(JSON.stringify(payload));

            console.log('Sending query:', payload);
            eventSource = new EventSource(url);

            eventSource.onmessage = function(event) {
                const data = event.data;
                if (!data) return;

                try {
                    const streamEvent = JSON.parse(data);
                    handleStreamEvent(streamEvent);
                } catch (e) {
                    console.error('Failed to parse event:', e, data);
                }
            };

            eventSource.onerror = function(error) {
                console.error('Connection error:', error);
                if (eventSource.readyState === EventSource.CLOSED) {
                    addEvent('Stream closed', 'done');
                } else {
                    addEvent('Connection error: ' + error, 'error');
                }
                eventSource.close();
            };
        }

        function handleStreamEvent(event) {
            console.log('Event:', event);

            let displayText = '';
            switch(event.type) {
                case 'start':
                    displayText = event.content;
                    break;
                case 'agent_start':
                    displayText = '🔄 ' + event.content + ' [' + event.agent + ']';
                    break;
                case 'agent_response':
                    displayText = '💬 ' + event.agent + ': ' + event.content;
                    history.push({role: 'assistant', content: event.content});
                    break;
                case 'tool_start':
                    displayText = '🔧 ' + event.content;
                    break;
                case 'tool_result':
                    displayText = '✅ ' + event.content;
                    break;
                case 'pause':
                    displayText = '⏸️  WAITING FOR INPUT';
                    document.getElementById('query').focus();
                    eventSource.close();
                    break;
                case 'done':
                    displayText = event.content;
                    eventSource.close();
                    break;
                case 'error':
                    displayText = '❌ ' + event.content;
                    eventSource.close();
                    break;
                case 'ping':
                    displayText = '(keep-alive ping)';
                    break;
                default:
                    displayText = '[' + event.type + '] ' + event.content;
            }

            addEvent(displayText, event.type);

            // Add user message to history on first user query
            if (event.type === 'start') {
                history = [{role: 'user', content: document.getElementById('query').value}];
            }
        }

        function addEvent(text, type = 'info') {
            const output = document.getElementById('output');
            const eventDiv = document.createElement('div');
            eventDiv.className = 'event ' + type;
            eventDiv.textContent = text;
            output.appendChild(eventDiv);
            output.scrollTop = output.scrollHeight;
        }

        function clearOutput() {
            document.getElementById('output').innerHTML = '';
            history = [];
        }

        // Test on load
        window.addEventListener('load', function() {
            console.log('Page loaded. Ready to test SSE streaming.');
        });
    </script>
</body>
</html>
```

### Cách sử dụng:

```bash
# Copy file vào thư mục server hoặc mở trực tiếp
cp test_streaming.html /Users/taipm/GitHub/go-bit-server-alpha/go-crewai/

# Mở trong trình duyệt
open http://localhost:8081/test_streaming.html

# Hoặc nếu dùng Linux
firefox http://localhost:8081/test_streaming.html
```

---

## Demo 6: PowerShell Demo (Windows)

### Script PowerShell để test

Tạo file `test_streaming.ps1`:

```powershell
# SSE Streaming Test Script (Windows PowerShell)

$ServerUrl = "http://localhost:8081"
$HealthUrl = "$ServerUrl/health"
$StreamUrl = "$ServerUrl/api/crew/stream"

Write-Host "🎬 SSE Streaming Demo - PowerShell" -ForegroundColor Cyan
Write-Host "==================================" -ForegroundColor Cyan

# 1. Check server health
Write-Host "`n[1] Checking server health..." -ForegroundColor Yellow
try {
    $health = Invoke-RestMethod -Uri $HealthUrl -Method Get
    Write-Host "✅ Server is healthy: $($health.status)" -ForegroundColor Green
} catch {
    Write-Host "❌ Server is not responding" -ForegroundColor Red
    exit
}

# 2. Test Scenario 1: Machine chậm
Write-Host "`n[2] Testing Scenario 1: Machine chậm" -ForegroundColor Yellow

$query1 = @{
    query = "Máy chậm lắm"
    history = @()
} | ConvertTo-Json

$params = @{
    Uri = $StreamUrl
    Method = "POST"
    Body = $query1
    ContentType = "application/json"
}

Write-Host "Sending query: 'Máy chậm lắm'" -ForegroundColor Cyan

try {
    $response = Invoke-RestMethod @params
    Write-Host "Response: $response" -ForegroundColor Green
} catch {
    Write-Host "Error: $($_.Exception.Message)" -ForegroundColor Red
}

# 3. Test with curl (if available)
Write-Host "`n[3] Testing with curl..." -ForegroundColor Yellow

$curlCommand = @"
curl -X POST $StreamUrl `
  -H "Content-Type: application/json" `
  -d '{"query":"Không vào được Internet"}'
"@

Write-Host "Command: $curlCommand" -ForegroundColor Cyan
Write-Host "Running... (watch the streaming events)" -ForegroundColor Yellow

Invoke-Expression $curlCommand

Write-Host "`n✅ Demo completed!" -ForegroundColor Green
```

### Chạy script:

```powershell
# Windows PowerShell
Set-ExecutionPolicy -ExecutionPolicy RemoteSigned -Scope CurrentUser
.\test_streaming.ps1
```

---

## Demo 7: Monitoring & Logging

### Demo: Real-time Monitoring

**Terminal 1: Start Server với logging**

```bash
cd go-crewai
go run ./cmd/main.go --server --port 8081 2>&1 | tee server.log
```

**Terminal 2: Watch logs in real-time**

```bash
tail -f server.log | grep -E "(Starting|Execution|Event|Error)"
```

**Terminal 3: Send test requests**

```bash
# Request 1
curl -X POST http://localhost:8081/api/crew/stream \
  -H "Content-Type: application/json" \
  -d '{"query":"Server 192.168.1.50 không ping được"}'

# Request 2 (sau khi request 1 hoàn tất)
curl -X POST http://localhost:8081/api/crew/stream \
  -H "Content-Type: application/json" \
  -d '{"query":"Máy chậm"}'
```

### Demo: Monitor Performance

```bash
# Theo dõi kết nối mạng trong real-time
watch -n 1 'curl -s http://localhost:8081/health | jq'

# Hoặc dùng lsof để xem kết nối
lsof -i :8081

# Theo dõi CPU & Memory (macOS)
while true; do
  echo "=== $(date) ==="
  top -l 1 | head -20
  sleep 5
done
```

---

## 📝 Tóm Tắt Demo Commands

### Quick Reference

```bash
# 1. Khởi động server
cd go-crewai
go run ./cmd/main.go --server --port 8081

# 2. Check health
curl http://localhost:8081/health

# 3. Test web client
open http://localhost:8081

# 4. Test curl - Scenario 1
curl -X POST http://localhost:8081/api/crew/stream \
  -H "Content-Type: application/json" \
  -d '{"query":"Máy chậm lắm"}'

# 5. Test curl - Scenario 2
curl -X POST http://localhost:8081/api/crew/stream \
  -H "Content-Type: application/json" \
  -d '{"query":"Tôi không vào được Internet"}'

# 6. Test curl - Scenario 3 (Pause/Resume)
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

---

## 🎯 Các Trường Hợp Test Đáng Chú Ý

### Test Case 1: Routing Flow
**Mục đích:** Test routing từ Orchestrator → Executor

```bash
curl -X POST http://localhost:8081/api/crew/stream \
  -H "Content-Type: application/json" \
  -d '{"query":"Server 192.168.1.50 có thể ping được không?"}'
```

**Kỳ vọng:** Orchestrator → Executor (direct routing)

### Test Case 2: Clarification Flow
**Mục đích:** Test routing từ Orchestrator → Clarifier → Executor

```bash
curl -X POST http://localhost:8081/api/crew/stream \
  -H "Content-Type: application/json" \
  -d '{"query":"Máy bị vấn đề"}'
```

**Kỳ vọng:** Orchestrator → Clarifier → PAUSE

### Test Case 3: Tool Execution
**Mục đích:** Test tool execution tracking

```bash
curl -X POST http://localhost:8081/api/crew/stream \
  -H "Content-Type: application/json" \
  -d '{"query":"Check disk space máy 192.168.1.100"}'
```

**Kỳ vọng:** Stream events cho từng tool execution

### Test Case 4: Error Handling
**Mục đích:** Test error event handling

```bash
# Gửi API key sai hoặc điều kiện error
curl -X POST http://localhost:8081/api/crew/stream \
  -H "Content-Type: application/json" \
  -d '{"query":""}'  # Empty query
```

**Kỳ vọng:** Error event với mô tả rõ

---

## 🔍 Debugging Tips

### Xem chi tiết events

```bash
# Redirect output vào file để xem chi tiết
curl -X POST http://localhost:8081/api/crew/stream \
  -H "Content-Type: application/json" \
  -d '{"query":"Máy chậm"}' > events.log 2>&1

# Xem file
cat events.log | jq '.'
```

### Kiểm tra headers

```bash
curl -v -X POST http://localhost:8081/api/crew/stream \
  -H "Content-Type: application/json" \
  -d '{"query":"Test"}'
```

**Kỳ vọng headers:**
```
< Content-Type: text/event-stream
< Cache-Control: no-cache
< Connection: keep-alive
< Access-Control-Allow-Origin: *
```

### Test concurrent requests

```bash
# Gửi 3 requests đồng thời
for i in {1..3}; do
  curl -X POST http://localhost:8081/api/crew/stream \
    -H "Content-Type: application/json" \
    -d "{\"query\":\"Test $i\"}" &
done
wait
```

---

## 📊 Performance Testing

### Load test đơn giản

```bash
#!/bin/bash
# save as load_test.sh

echo "🚀 Starting load test..."
TOTAL_REQUESTS=10

for i in $(seq 1 $TOTAL_REQUESTS); do
  echo "Request $i/TOTAL_REQUESTS"
  curl -X POST http://localhost:8081/api/crew/stream \
    -H "Content-Type: application/json" \
    -d "{\"query\":\"Load test request $i\"}" > /dev/null 2>&1 &
done

wait
echo "✅ Load test completed!"
```

```bash
chmod +x load_test.sh
./load_test.sh
```

---

## 🎓 Kết Luận

Các demo này giúp bạn:

1. ✅ Kiểm tra server hoạt động đúng
2. ✅ Test các scenario khác nhau
3. ✅ Xem events streaming real-time
4. ✅ Debug vấn đề nếu có
5. ✅ Thử nghiệm performance

**Chọn demo phù hợp với nhu cầu của bạn và bắt đầu test!** 🚀

---

**Version:** 1.0
**Last Updated:** 2025-12-19
**Status:** Ready to Demo ✅

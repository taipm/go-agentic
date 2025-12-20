# IT Support System - Quick Start Guide

## ⚡ 30 Second Setup

### 1. Set your OpenAI API key

```bash
export OPENAI_API_KEY=sk-proj-your-key-here
```

### 2. Run CLI mode (Interactive)

```bash
cd examples/it-support
go run ./cmd/main.go
```

Enter an IT issue and get instant diagnosis!

---

## 🌐 Run Web Server with Live Streaming

```bash
export OPENAI_API_KEY=sk-proj-your-key-here
go run ./cmd/main.go --server --port 8081
```

Open browser: [http://localhost:8081](http://localhost:8081)

### Features

- ✅ Real-time SSE streaming of agent interactions
- ✅ 5 preset scenarios (Máy chậm, No Internet, Network lỗi, Disk đầy, RAM cao)
- ✅ Conversation history tracking
- ✅ Color-coded event visualization

---

## 📝 Example Issues to Test

### CLI Mode

```bash
# Check my machine
$ go run ./cmd/main.go
Describe your IT issue:
> Bạn tự lấy thông tin máy hiện tại

# Check specific server
$ go run ./cmd/main.go
Describe your IT issue:
> Server 192.168.1.50 không ping được

# Vague issue (will ask clarifying questions)
$ go run ./cmd/main.go
Describe your IT issue:
> Máy tính của tôi chậm
```

### Web Mode

1. Run server: `go run ./cmd/main.go --server --port 8081`
2. Open [http://localhost:8081](http://localhost:8081)
3. Click preset scenario buttons or enter custom query
4. Watch real-time SSE streaming events

---

## 🔧 Customize

### Change Server Port

```bash
go run ./cmd/main.go --server --port 9000
```

### Available Tools (13)

System Info: `GetSystemInfo`, `GetCPUUsage`, `GetMemoryUsage`, `GetDiskSpace`

Network: `PingHost`, `ResolveDNS`, `CheckNetworkStatus`

Advanced: `CheckMemoryStatus`, `CheckDiskStatus`, `GetRunningProcesses`, `GetSystemDiagnostics`

Services: `CheckServiceStatus`, `ExecuteCommand`

---

## 📚 Architecture

```
User Input
    ↓
Orchestrator (Router)
    ├─ If has info → Executor (13 diagnostic tools)
    └─ If vague → Clarifier (asks 2-3 questions)
         ↓
      Clarifier gathers details
         ↓
      Routes to Executor
         ↓
      Returns diagnosis
```

---

## 🚀 Next Steps

- See [README.md](README.md) for full documentation
- Check [../README.md](../README.md) for other examples
- Review [../../README.md](../../README.md) for framework overview

---

## ✅ Verify Installation

```bash
# Test build
go build ./cmd/main.go

# Run with dry run
OPENAI_API_KEY=test go run ./cmd/main.go <<< "test"
```

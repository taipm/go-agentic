# 🎯 START HERE - go-agentic Navigation Guide

**Welcome to go-agentic!**

This guide helps you find exactly what you need. Choose your path below:

---

## 🚀 I Want to Get Started NOW (3 minutes)

```bash
# 1. Start the server
go run ./cmd/main.go --server --port 8081

# 2. Open in browser
open http://localhost:8081

# 3. Try it!
Type: "Máy chậm lắm" (Machine is slow)
```

👉 **Read:** [QUICKSTART.md](QUICKSTART.md) (3 min)

---

## 💬 Try These Requests:

Type any of these and press Enter:

### 1. System Health Check
```
Check my system health
```
**What it does:** Runs CPU, memory, disk, and system info checks

### 2. Vague Problem (Tests Clarification)
```
My machine is slow
```
**What it does:** Asks you questions, then diagnoses

### 3. Network Test
```
Can you reach 8.8.8.8?
```
**What it does:** Tests network connectivity

### 4. Service Status
```
Is SSH running?
```
**What it does:** Checks if SSH service is active

### 5. Exit
```
quit
```
**What it does:** Closes the application

---

## 🎯 What Happens?

When you submit a request:

1. **Orchestrator Agent** reads your request (1-2 sec)
2. **Router** decides: Is this clear? Or vague?
   - **Clear** → Goes straight to Executor
   - **Vague** → Goes to Clarifier first
3. **Executor Agent** (if needed)
   - Runs diagnostic tools
   - Analyzes results
   - Provides recommendations
4. **Workflow ends** with "[Conversation ended - terminal agent reached]"
5. **Prompt returns** for next request

---

## ⏱️ Timing

- First request: 2-5 seconds (API call + tools)
- Following requests: 1-3 seconds
- Tools: <1 second each

---

## ❌ If Something Goes Wrong

### "Error: OPENAI_API_KEY environment variable not set"
**Fix:** Check `.env` file has your API key:
```bash
cat .env | grep OPENAI_API_KEY
```
Should show: `OPENAI_API_KEY=sk-proj-...`

### "Error loading agent configs"
**Fix:** Verify config files exist:
```bash
ls -la config/agents/
```

### Tools don't seem to execute
**Fix:** Tools usually run silently. Look for results in the agent's response.

---

## 📊 What You'll See

Example output for "Check my system health":

```
Agent: Executor
Response: System Analysis:
- OS: macOS Ventura 13.5
- Hostname: your-machine
- Uptime: 45 days
- CPU Usage: 35%
- Memory: 8GB total, 5GB used (62%)
- Disk: 500GB total, 360GB used (72%)

Recommendations:
✓ System is generally healthy
⚠ Monitor disk usage (72% full)
→ Consider clearing old files or upgrading storage

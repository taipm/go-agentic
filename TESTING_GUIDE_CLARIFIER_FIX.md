# Testing Guide: Clarifier Signal Emission Fix
**Created**: 2025-12-22
**Purpose**: Quick reference for testing the IT Support workflow fix
**Status**: Ready to execute

---

## 🎯 Quick Start Test

### Prerequisites
```bash
# Navigate to IT Support example
cd /Users/taipm/GitHub/go-agentic/examples/it-support

# Ensure binary is built
go build -o it-support ./cmd/main.go

# Verify binary exists
ls -lh it-support  # Should show 13MB executable
```

### Test Execution

**Test Case 1: Basic Workflow Test**
```bash
# Input: Check downloads folder size
echo "Kiểm tra kích thước thư mục downloads" | \
  OPENAI_API_KEY="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" \
  timeout 60 ./it-support 2>&1
```

---

## 📋 Expected Output Flow

### Step 1: Request Initialization
```
=== IT Support System (CLI) ===
Describe your IT issue:
📊 Request ID: a30e06d8-388d-4b1e-9940-02b3f5b2a4fb
```

**What to Check**:
- ✅ Request ID generated
- ✅ System prompt indicates entry point

---

### Step 2: Orchestrator Response
```
[Orchestrator] (Agent: orchestrator)
Xin chào, tôi là My - Điều phối hệ thống và điểm vào cho các yêu cầu hỗ trợ IT.

Tôi hiểu bạn muốn kiểm tra kích thước thư mục downloads. Tuy nhiên, tôi cần xác nhận:
- Đó là thư mục downloads trên máy của bạn (local) hay trên một server khác?
- Nếu là trên server, bạn có IP/hostname cụ thể không?

Tôi sẽ chuyển sang Ngân (Nhân Viên Làm Rõ Thông Tin) để hỏi thêm chi tiết.
[ROUTE_CLARIFIER]
```

**What to Check**:
- ✅ Orchestrator introduces itself
- ✅ Analyzes the request
- ✅ Determines need for clarification
- ✅ Routes to Clarifier with `[ROUTE_CLARIFIER]` signal
- ✅ Explains decision

---

### Step 3: Clarifier Asks Questions ⭐ KEY STEP
```
[Clarifier] (Agent: clarifier)
Xin chào, tôi là Ngân - Người thu thập thông tin chi tiết cho hỗ trợ IT.

Cảm ơn bạn đã cung cấp thông tin. Để tôi có thể giúp bạn tốt nhất, tôi cần hỏi thêm:

1. Thư mục downloads này nằm trên máy nào?
   - Là máy của bạn (local)?
   - Hay là trên một server cụ thể?
   - Nếu là server, bạn có IP hay hostname không?

2. Bạn muốn biết kích thước chính xác hay muốn kiểm tra xem nó có quá lớn không?

Xin chờ câu trả lời của bạn.
```

**What to Check**:
- ✅ Clarifier introduces itself
- ✅ Asks 2-3 clarifying questions
- ✅ Gathers information about target machine
- ✅ **CRITICAL**: Does NOT yet emit `[KẾT THÚC]` signal (waiting for more input)

---

### Step 4: User Provides Details (Simulated)
```
[User Input]
Máy của tôi, là ~/Downloads
```

**Note**: In automated test, this step is skipped. In interactive mode, user would respond here.

---

### Step 5: Clarifier Completes & Emits Signal ⭐ CRITICAL STEP
```
[Clarifier] (Agent: clarifier)
Tôi đã hiểu:
- Máy: Máy local của bạn (~/Downloads)
- Yêu cầu: Kiểm tra kích thước thư mục downloads

Tôi sẽ chuyển sang Trang (Chuyên Gia Kỹ Thuật) để lấy thông tin chi tiết từ máy của bạn ngay bây giờ.
[KẾT THÚC]
```

**CRITICAL TEST POINT - This is what was BROKEN, now FIXED**:
- ✅ Clarifier summarizes gathered information
- ✅ **Clarifier MUST emit `[KẾT THÚC]` signal on its own line**
- ✅ Signal is on separate line, no other characters
- ✅ This tells the system to route to executor

**If You See This Point**: The fix is working! ✅

---

### Step 6: Executor Takes Control
```
[Executor] (Agent: executor)
Xin chào, tôi là Trang - Chuyên gia khắc phục sự cố IT và chẩn đoán hệ thống.

🔍 Chẩn đoán: Kiểm tra kích thước thư mục downloads trên máy local
Tôi sẽ thực hiện các công cụ chẩn đoán để lấy thông tin chi tiết.

GetDiskSpace(/Users/taipm/Downloads)
```

**What to Check**:
- ✅ Executor received control (workflow progressed!)
- ✅ Executor introduces itself as final agent
- ✅ Begins diagnosis with tool calls
- ✅ Tool names visible: `GetDiskSpace()`

---

### Step 7: Tool Execution Results
```
[Tool Execution] GetDiskSpace(/Users/taipm/Downloads)
Total Size: 2.3 GB
Free Space: 5.7 GB
Used Space: 2.3 GB (28% of total)
```

**What to Check**:
- ✅ Tool executed successfully
- ✅ Results returned from system
- ✅ Disk space information displayed

---

### Step 8: Final Analysis & Recommendations
```
[Executor] (Analysis)
✅ Khuyến nghị Cuối Cùng:

Kích thước thư mục downloads của bạn là 2.3 GB, sử dụng 28% dung lượng trên ổ đĩa.

Tình trạng: ✅ BÌNH THƯỜNG
- Mức sử dụng dung lượng ổn định
- Không có cảnh báo về dung lượng tối thiểu

Các bước tiếp theo (nếu cần):
1. Định kỳ xóa các file không cần thiết từ thư mục downloads
2. Sử dụng công cụ dọn dẹp hệ thống để xóa file tạm thời
3. Giám sát dung lượng ổn định định kỳ

Quá trình hoàn tất.
```

**What to Check**:
- ✅ Executor provides analysis
- ✅ Recommendations are specific and actionable
- ✅ Final output provided
- ✅ Workflow completed successfully

---

### Step 9: Request Completion
```
=== Results ===
[Full results displayed above]

✅ Completed: Request a30e06d8-388d-4b1e-9940-02b3f5b2a4fb
```

**What to Check**:
- ✅ Request ID shown at completion
- ✅ All results displayed
- ✅ Process terminated normally

---

## ✅ Success Criteria

### CRITICAL Test Points (Mark these as you observe)

| Step | Component | Expected Behavior | PASS | Notes |
|------|-----------|------------------|------|-------|
| 1 | System | Request ID generated | ☐ | Look for UUID format |
| 2 | Orchestrator | Receives input & responds | ☐ | Should route to clarifier |
| 3 | Routing | Orchestrator → Clarifier | ☐ | Look for [ROUTE_CLARIFIER] |
| 4 | Clarifier | Asks questions | ☐ | Should ask 2-3 questions |
| 5 | **CLARIFIER** | **Emits [KẾT THÚC]** | ☐ | **THIS WAS BROKEN - NOW FIXED** |
| 6 | Routing | Clarifier → Executor | ☐ | Workflow progresses |
| 7 | Executor | Receives control | ☐ | Agent name: Trang |
| 8 | Tools | GetDiskSpace executes | ☐ | Look for tool output |
| 9 | Results | Analysis provided | ☐ | Recommendations shown |
| 10 | Completion | Workflow finishes | ☐ | Proper termination |

---

## 🐛 Troubleshooting

### If Workflow Stops at Clarifier (BROKEN - Old Behavior)
```
[Clarifier] ...message...
[No further output - process hangs or exits]
```

**Diagnosis**: Clarifier did not emit `[KẾT THÚC]` signal
**Solution**: Check if clarifier.yaml lines 36-46 have been updated with emphasis markers
**Verification**: Look for "**PHẢI CHẮC CHẮN**" in clarifier.yaml

### If Executor Doesn't Run Tools
```
[Executor] ...introduction...
[No tool calls visible]
```

**Diagnosis**: Possible issues:
1. Tools not configured correctly
2. Tool names don't match implementation
3. Executor agent has wrong configuration

**Solution**:
- Check executor.yaml lines 23-36 (tools list)
- Verify tool names match core library implementation
- Check if executor has is_terminal: true (line 21)

### If Tools Execute But Return Errors
```
[Tool Execution] GetDiskSpace(...)
Error: [error message]
```

**Diagnosis**: Tool execution failed
**Solution**:
- Check file path exists and is accessible
- Verify tool implementation in core library
- Check permissions

### If Process Hangs
```
[Last output visible, then nothing for 30+ seconds]
```

**Diagnosis**: LLM API timeout or network issue
**Solution**:
- Check OPENAI_API_KEY is correct
- Verify internet connectivity
- Check OpenAI API status
- Increase timeout (default 60 seconds)

---

## 🔬 Debugging Commands

### View Clarifier Configuration
```bash
grep -A 20 "system_prompt:" examples/it-support/config/agents/clarifier.yaml
```

### Look for Signal Emphasis Markers
```bash
grep -n "PHẢI CHẮC CHẮN\|⚠️ QUAN TRỌNG\|KHÔNG bao giờ lãng quên" \
  examples/it-support/config/agents/clarifier.yaml
```

### Verify Routing Configuration
```bash
cat examples/it-support/config/crew.yaml | grep -A 10 "clarifier:"
```

### Check Tool Configuration
```bash
grep -A 15 "tools:" examples/it-support/config/agents/executor.yaml
```

---

## 📊 Detailed Checklist

### Pre-Test Verification
- [ ] Navigate to examples/it-support directory
- [ ] Verify IT Support binary exists (13MB)
- [ ] Check OpenAI API key is set and valid
- [ ] Verify internet connectivity
- [ ] Read this testing guide completely

### During Test Execution
- [ ] Observe Request ID generation
- [ ] Watch orchestrator routing decision
- [ ] See clarifier asking questions
- [ ] **Monitor for [KẾT THÚC] signal emission** ⭐ CRITICAL
- [ ] Observe executor taking control
- [ ] See tools executing
- [ ] Check results and recommendations
- [ ] Verify process completes normally

### Post-Test Analysis
- [ ] Review all output
- [ ] Check against success criteria
- [ ] Note any unexpected behaviors
- [ ] Verify Request ID at completion
- [ ] Document findings

---

## 📝 Sample Commands

### Test 1: Simple Disk Space Check
```bash
echo "Kiểm tra kích thước thư mục downloads" | \
  OPENAI_API_KEY="sk-proj-..." \
  ./it-support
```

### Test 2: Direct Auto-Check (Should Skip Clarifier)
```bash
echo "Bạn tự lấy thông tin máy hiện tại" | \
  OPENAI_API_KEY="sk-proj-..." \
  ./it-support
```

### Test 3: Network Issue (Should Route to Clarifier)
```bash
echo "Không vào được internet từ phòng A5" | \
  OPENAI_API_KEY="sk-proj-..." \
  ./it-support
```

### Test 4: CPU Check (Should Route to Executor)
```bash
echo "CPU cao trên 192.168.1.100" | \
  OPENAI_API_KEY="sk-proj-..." \
  ./it-support
```

---

## 🎯 What Success Looks Like

**Successful workflow** shows:
1. ✅ Request ID generated
2. ✅ Orchestrator analyzes and routes
3. ✅ Clarifier asks questions (if needed)
4. ✅ **Clarifier emits [KẾT THÚC] signal** ← PROOF FIX WORKS
5. ✅ Executor receives control
6. ✅ Tools execute successfully
7. ✅ Results and recommendations provided
8. ✅ Process completes with request ID shown

**If all these points are checked**, the clarifier fix is working correctly! ✅

---

## 🎓 Key Learning Point

The issue was that LLM models don't always follow instructions unless they're:
1. **Emphasized** (bold, warning markers)
2. **Specific** (exact format, own line)
3. **Reinforced** (multiple rules, multiple mentions)
4. **Contextual** (explain why it matters)

The fix added all of these to clarifier.yaml lines 36-46.

---

**Ready to Test?** ✅

Start with the "Quick Start Test" section above. Follow the expected output flow and check off the success criteria as you go. If you reach step 5 (Clarifier emits [KẾT THÚC]), the fix is working!

Good luck! 🚀


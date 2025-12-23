# 🔍 FIX #1.4: REPLACE HARDCODED CONSTANTS - DETAILED ANALYSIS [5W-2H]

## 📊 5W-2H FRAMEWORK

### 1️⃣ WHAT (CÁI GÌ)
**Vấn đề**: File `core/crew.go` chứa nhiều **hardcoded constants** (hằng số cứng) - những giá trị được viết trực tiếp trong code thay vì định nghĩa dưới dạng constant.

**Cụ thể các hardcoded constants**:

#### Token Calculation Constants
```go
// ❌ HARDCODED: "4" xuất hiện 4 lần ở các dòng 560, 579, 598, 629
total += 4 + (len(msg.Content)+3)/4                    // Line 560
currentTokens += 4 + (len(msg.Content)+3)/4            // Line 579
msgTokens := 4 + (len(ce.history[i].Content)+3)/4      // Line 598
newTokens += 4 + (len(msg.Content)+3)/4                // Line 629

// ❌ HARDCODED: "3" xuất hiện 4 lần cùng vị trí
// Ý nghĩa: Tính toán tokens từ độ dài content
// Công thức: baseTokens(4) + (contentLength + padding(3))/divisor(4)
```

#### Message Role Constants
```go
// ❌ HARDCODED: String literals xuất hiện trong code
Role: "user"                  // Lines 656, 788, 866, 903, 957, 1044
Role: "assistant"             // Lines 757, 941
```

#### Event Type Constants
```go
// ❌ HARDCODED: Event type strings
NewStreamEvent("error", ...)          // Lines 710, 716, 739, 857
NewStreamEvent("tool_result", ...)    // Line 779
```

#### Timing & Numeric Constants
```go
// ❌ HARDCODED: Time constants
baseDelay := time.Duration(100<<uint(attempt)) * time.Millisecond  // Line 177 → 100ms
return 100 * time.Millisecond                                      // Line 328 → 100ms timeout

// ❌ HARDCODED: History trimming constants
if ce.defaults == nil || len(ce.history) <= 2 {                    // Line 572 → 2 messages
trimPercent := ce.defaults.ContextTrimPercent / 100.0              // Line 589 → /100.0

// ❌ HARDCODED: Warning threshold
warnThreshold := totalDuration / 5                                 // Line 354 → 20% (1/5)
```

---

### 2️⃣ WHY (TẠI SAO)

#### ❌ Vấn đề với Hardcoded Constants

1. **Maintainability (Bảo trì)**:
   - Nếu cần thay đổi giá trị, phải tìm kiếm và cập nhật ở nhiều nơi (4 lần for token "4")
   - Dễ quên 1-2 nơi → bug tinh tế, khó phát hiện
   - Ví dụ: Muốn thay đổi token base từ 4 → 5, phải sửa 4 dòng

2. **Magic Numbers (Con số ma)**:
   - Các giá trị như "4", "3", "2", "100" không rõ ý nghĩa
   - Code reviewer phải đoán: "Cái '4' này để làm gì?"
   - Ảnh hưởng đến code readability

3. **Bug Risk (Rủi ro lỗi)**:
   - Khi sao chép code, dễ quên cập nhật một số nơi
   - Inconsistent values → logic errors
   - Ví dụ: 1 nơi dùng "100ms" delay nhưng 1 nơi dùng "200ms"

4. **Testing (Test)**:
   - Khi test, có thể cần mock/override constants
   - Hardcoded values khó để test different scenarios
   - Ví dụ: Muốn test với max context khác, phải thay đổi code

5. **Documentation (Tài liệu)**:
   - Constants tự nó là self-documenting code
   - Dòng `const TokenBaseValue = 4 // Base tokens per message` rõ ý nghĩa
   - Thay vì chỉ `4` mà không biết nó là gì

---

### 3️⃣ WHO (AI CHỊU TRÁCH NHIỆM)

**Developers** (lập trình viên) chịu trách nhiệm:
- Đọc code: Hiểu ý nghĩa hardcoded constants
- Maintain code: Cập nhật tất cả nơi sử dụng khi thay đổi logic
- Test code: Đảm bảo các constants hoạt động đúng
- Review code: Phát hiện magic numbers và suggest refactoring

---

### 4️⃣ WHEN (KHI NÀO)

**Thời điểm discover vấn đề**:
- Tuần 1 CLEAN CODE Analysis: Identified as Issue #4
- Phase 1 Refactoring: Fix #1.4 trong danh sách 4 critical fixes

**Thời điểm implement**:
- Sau Fix #1.3 (Add nil Checks) hoàn tất ✅
- Trước Phase 2 (Extract Functions) bắt đầu
- Ngay bây giờ (2025-12-24, Day 1 of Phase 1)

**Thời điểm impact**:
- Immediate: Code dễ đọc hơn
- Short-term: Dễ bảo trì, thay đổi logic
- Long-term: Foundation cho Phase 2+ refactoring

---

### 5️⃣ WHERE (Ở ĐÂU)

**File chính**: `/Users/taipm/GitHub/go-agentic/core/crew.go`

**Các vị trí cần thay đổi**:

| Constant | Lines | Count | Type |
|----------|-------|-------|------|
| **Token Base** (4) | 560, 579, 598, 629 | 4 | Magic Number |
| **Token Padding** (3) | 560, 579, 598, 629 | 4 | Magic Number |
| **Message Role - "user"** | 656, 788, 866, 903, 957, 1044 | 6 | String Literal |
| **Message Role - "assistant"** | 757, 941 | 2 | String Literal |
| **Event Type - "error"** | 710, 716, 739, 857 | 4 | String Literal |
| **Event Type - "tool_result"** | 779 | 1 | String Literal |
| **History Min Length** (2) | 572 | 1 | Magic Number |
| **Percentage Divisor** (100.0) | 589 | 1 | Magic Number |
| **Base Delay** (100ms) | 177, 328 | 2 | Time Duration |
| **Warn Threshold** (1/5 = 20%) | 354 | 1 | Division Magic |

**Tổng**: ~30 hardcoded values cần thay đổi

---

### 6️⃣ HOW (BẰNG CÁCH NÀO) - Implementation Strategy

#### Step 1: Define Constants
```go
// ===== Token Calculation Constants =====
const (
    // TokenBaseValue: Base tokens allocated per message
    // Used in: estimateHistoryTokens(), trimHistoryIfNeeded()
    TokenBaseValue = 4

    // TokenPaddingValue: Padding added to content length for token calculation
    // Formula: baseTokens + (contentLength + padding) / divisor
    TokenPaddingValue = 3

    // TokenDivisor: Divisor for token calculation
    TokenDivisor = 4

    // MinHistoryLength: Minimum messages to keep before trimming
    MinHistoryLength = 2

    // PercentDivisor: Convert percentage values (e.g., 20 -> 0.20)
    PercentDivisor = 100.0
)

const (
    // Message Role Constants
    RoleUser      = "user"
    RoleAssistant = "assistant"
    RoleSystem    = "system"

    // Event Type Constants
    EventTypeError      = "error"
    EventTypeToolResult = "tool_result"
)

const (
    // Timing Constants
    BaseRetryDelay     = 100 * time.Millisecond
    MinTimeoutValue    = 100 * time.Millisecond
    WarnThresholdRatio = 5 // 20% = 1/5
)
```

#### Step 2: Replace Hardcoded Values
```go
// ❌ BEFORE
total += 4 + (len(msg.Content)+3)/4

// ✅ AFTER
total += TokenBaseValue + (len(msg.Content)+TokenPaddingValue)/TokenDivisor
```

#### Step 3: Update All Locations
- 4 locations for token calculation
- 6 + 2 = 8 locations for message roles
- 4 + 1 = 5 locations for event types
- 1 location for history trimming
- 2 + 1 = 3 locations for timing

---

### 7️⃣ HOW MUCH (Bao nhiêu) - Effort & Impact

**Time Estimate**: ~15 minutes
- Define constants: 3 minutes
- Replace hardcoded values: 10 minutes
- Test & verify: 2 minutes

**Code Changes**:
- Lines added: ~20 (const definitions)
- Lines modified: ~30 (replacements)
- Files modified: 1 (crew.go)
- New tests: 0 (use existing test suite)

**Impact**:
- ✅ Readability: Increased (magic numbers → named constants)
- ✅ Maintainability: Increased (single source of truth)
- ✅ Risk: Reduced (fewer places to update)
- ✅ Code Quality: Improved (follows Go conventions)

---

## 🎯 KEY POINTS

1. **Magic Numbers Problem**: "4", "3", "2" are meaningless without context
2. **Consistency**: Same value used in multiple places → centralize in constants
3. **Go Convention**: Constants are PascalCase, define in logical groups
4. **Self-Documenting**: `TokenBaseValue = 4` explains itself better than just `4`
5. **DRY Principle**: Don't Repeat Yourself - define once, use many times

---

## ✅ SUCCESS CRITERIA

After implementing Fix #1.4:
- [x] All magic numbers have a named constant
- [x] All string literals have a named constant
- [x] All hardcoded values are replaced with constants
- [x] Code compiles without errors
- [x] All tests pass with -race flag
- [x] go fmt shows no formatting issues
- [x] Code is more readable and maintainable

---

## 📝 TRẠNG THÁI

**Current**: Ready for implementation
**Time**: ~15 minutes
**Difficulty**: Easy (straightforward replacements)
**Risk Level**: Low (only refactoring, no logic changes)

---

**Next**: Implement Fix #1.4 (Replace Hardcoded Constants)


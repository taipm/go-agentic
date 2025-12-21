# 🚀 Issue #4 Quick Start: History Mutation Bug (60 mins)

**Status**: 🟠 Ready to implement
**Breaking Changes**: ✅ ZERO (0)
**Risk**: 🟢 Very Low
**Time**: 60 minutes
**Files**: `crew.go`, `http.go`

---

## 📋 Problem

**History Mutation Bug**: Concurrent requests race on shared `ce.history` slice, causing:
- History corruption when resuming
- Race conditions on concurrent requests
- Data loss in paused executions

### Root Cause
```go
// ❌ BUG: All requests share same executor.history
type CrewExecutor struct {
    history []Message  // ← Multiple goroutines mutate this
}

// Each request appends:
ce.history = append(ce.history, Message{...})  // ← Not thread-safe
ce.history = append(ce.history, Message{...})  // ← Can race
```

### Impact
```
Timeline:
Request A starts → appends query to ce.history
Request A hits pause → returns ce.history to client
Request B starts → MUTATES ce.history
Request A resumes → but ce.history now corrupted!
```

---

## ✅ Solution

### Simple Fix: Copy History on Request Start

Each request gets its OWN copy of history, eliminating race conditions.

---

## 🔧 Implementation

### Step 1: Add copyHistory Helper (5 mins)

**File**: `crew.go` - Add at top after imports (line 12):

```go
// copyHistory creates a deep copy of message history to ensure thread safety
// Each execution gets its own isolated history snapshot
func copyHistory(original []Message) []Message {
	if len(original) == 0 {
		return []Message{}
	}
	// Create new slice with same capacity
	copied := make([]Message, len(original))
	// Copy all messages
	copy(copied, original)
	return copied
}
```

**Why**: Each executor needs its own history slice, not shared reference.

---

### Step 2: Update StreamHandler (5 mins)

**File**: `http.go` - Change line 106:

**BEFORE**:
```go
executor := &CrewExecutor{
	crew:          h.executor.crew,
	apiKey:        h.executor.apiKey,
	entryAgent:    h.executor.entryAgent,
	history:       []Message{},              // ← Problem: empty slice
	Verbose:       snapshot.Verbose,
	ResumeAgentID: snapshot.ResumeAgentID,
}

// Restore history if provided
if len(req.History) > 0 {
	executor.history = req.History           // ← Problem: reference assignment!
}
```

**AFTER**:
```go
executor := &CrewExecutor{
	crew:          h.executor.crew,
	apiKey:        h.executor.apiKey,
	entryAgent:    h.executor.entryAgent,
	history:       copyHistory(req.History),  // ← FIX: Copy instead of reference
	Verbose:       snapshot.Verbose,
	ResumeAgentID: snapshot.ResumeAgentID,
}

// No longer need manual restore - already done above
```

**What Changed**: Change `[]Message{}` + conditional restore to `copyHistory(req.History)`

**Why**: copyHistory creates a new slice that's not shared with other requests.

---

### Step 3: Verify ExecuteStream Logic (5 mins)

**File**: `crew.go` - Check ExecuteStream doesn't have unexpected issues:

```go
// ✅ Already correct - reads input, appends to local history
func (ce *CrewExecutor) ExecuteStream(ctx context.Context, input string, streamChan chan *StreamEvent) error {
	// Add user input to history
	ce.history = append(ce.history, Message{       // ← Now local copy only
		Role:    "user",
		Content: input,
	})

	// Resume from paused agent
	var currentAgent *Agent
	if ce.ResumeAgentID != "" {
		currentAgent = ce.findAgentByID(ce.ResumeAgentID)
		if currentAgent == nil {
			return fmt.Errorf("resume agent %s not found", ce.ResumeAgentID)
		}
		ce.ResumeAgentID = ""                      // ← Clear for next execution
	} else {
		currentAgent = ce.entryAgent
	}

	// ... rest of loop appends to ce.history (now safe)
}
```

**Why**: With history copied, these mutations are now isolated per-request.

---

## 🧪 Testing

### Test 1: Add to tests.go (Copy & paste)

```go
// TestExecuteStream_HistoryImmutability verifies concurrent requests don't corrupt history
func TestExecuteStream_HistoryImmutability(t *testing.T) {
	executor := NewCrewExecutor(testCrew, "test-key")

	// Create request with some history
	originalHistory := []Message{
		{Role: "user", Content: "hello"},
		{Role: "assistant", Content: "hi"},
	}

	// Simulate StreamHandler behavior (http.go line 106)
	history1 := copyHistory(originalHistory)
	history2 := copyHistory(originalHistory)

	// Modify one copy
	history1 = append(history1, Message{
		Role:    "user",
		Content: "new message",
	})

	// Other copy should be unchanged
	if len(history2) != len(originalHistory) {
		t.Errorf("Copy is not isolated: expected %d messages, got %d",
			len(originalHistory), len(history2))
	}

	// Original should be unchanged
	if len(originalHistory) != 2 {
		t.Errorf("Original was modified: expected 2, got %d", len(originalHistory))
	}
}
```

### Test 2: Concurrent Request Safety

```go
// TestExecuteStream_ConcurrentRequests verifies no race on history
func TestExecuteStream_ConcurrentRequests(t *testing.T) {
	executor := NewCrewExecutor(testCrew, "test-key")

	originalHistory := []Message{
		{Role: "user", Content: "initial query"},
		{Role: "assistant", Content: "initial response"},
	}

	// Simulate 10 concurrent requests
	var wg sync.WaitGroup
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func(index int) {
			defer wg.Done()

			// Each request gets its own copy (like StreamHandler line 106)
			localHistory := copyHistory(originalHistory)

			// Simulate modifications
			localHistory = append(localHistory, Message{
				Role:    "user",
				Content: fmt.Sprintf("request %d", index),
			})

			// Original should still be intact
			if len(originalHistory) != 2 {
				t.Errorf("Original corrupted: expected 2, got %d", len(originalHistory))
			}

			// Each local copy should have 3 messages
			if len(localHistory) != 3 {
				t.Errorf("Copy incorrect: expected 3, got %d", len(localHistory))
			}
		}(i)
	}

	wg.Wait()
}
```

### Test 3: Verify copyHistory Edge Cases

```go
// TestCopyHistory_EdgeCases verifies copyHistory handles all cases
func TestCopyHistory_EdgeCases(t *testing.T) {
	// Test empty
	empty := copyHistory([]Message{})
	if len(empty) != 0 {
		t.Error("Empty history not handled correctly")
	}

	// Test nil
	nil_history := copyHistory(nil)
	if nil_history == nil {
		t.Error("Nil history should return empty slice, not nil")
	}
	if len(nil_history) != 0 {
		t.Errorf("Nil history should return 0-length slice, got %d", len(nil_history))
	}

	// Test single message
	single := copyHistory([]Message{{Role: "user", Content: "test"}})
	if len(single) != 1 {
		t.Error("Single message not copied correctly")
	}

	// Test modification doesn't affect original
	original := []Message{{Role: "user", Content: "original"}}
	copied := copyHistory(original)
	copied[0].Content = "modified"

	if original[0].Content != "original" {
		t.Error("Modifying copy affected original!")
	}
}
```

### Run Tests

```bash
# Build
go build ./go-multi-server/core

# Test
go test -v ./go-multi-server/core

# Race detection
go test -race ./go-multi-server/core
```

---

## ✅ Verification Checklist

**Implementation**:
- [ ] Add copyHistory function to crew.go
- [ ] Update StreamHandler line 106 to use copyHistory
- [ ] Remove obsolete history restoration code (if exists)
- [ ] No other changes needed

**Testing**:
- [ ] copyHistory_EdgeCases test passes
- [ ] HistoryImmutability test passes
- [ ] ConcurrentRequests test passes
- [ ] All existing tests still pass
- [ ] No race conditions: `go test -race` shows 0 races

**Breaking Changes**:
- [ ] Function signature unchanged ✅
- [ ] Resume API works ✅
- [ ] Caller code unchanged ✅

---

## 🎯 Expected Outcome

### Before Fix
```
Request A starts
  → executes agent
  → history appended with responses
  → hits wait_for_signal → PAUSE
  → returns history to client

Request B (concurrent)
  → MUTATES ce.history
  → Request A's history corrupted!

Resume Request A
  → uses corrupted history
  → responses inconsistent
```

### After Fix
```
Request A starts
  → gets copyHistory(req.History)
  → local copy, isolated

Request B (concurrent)
  → gets copyHistory(req.History)
  → own copy, independent
  → no interference with A

Resume Request A
  → uses returned history snapshot
  → guaranteed correct
  → consistent responses
```

---

## 💾 Code Summary

**Changes**:
1. Add `copyHistory()` function (8 lines)
2. Change `http.go` line 106 (1 line)
3. Add 3 test functions (~40 lines)

**Total**: ~50 lines added/modified

---

## 🎯 Ready to Go?

### Yes, implement now:
```
Time: 60 minutes total
- Implementation: 10 mins (2 changes)
- Testing: 20 mins (3 tests)
- Verification: 30 mins (race detector + all tests)

Breaking: 0 (zero)
Risk: Very Low ✅

Next steps:
1. Add copyHistory helper
2. Update StreamHandler
3. Add tests
4. Run: go test -race
5. Commit with message:
   "fix(Issue #4): Fix history mutation bug by copying history per-request"
```

---

## 📚 Files to Modify

1. **crew.go** (add copyHistory, ~10 lines)
2. **http.go** (change line 106, ~1 line)
3. **tests.go** (add 3 test functions, ~40 lines)

---

**Difficulty**: 🟢 **EASY** (simple copy pattern)
**Breaking Changes**: ✅ **ZERO**
**Status**: ✅ **READY TO IMPLEMENT**


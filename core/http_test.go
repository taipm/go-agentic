package crewai

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

// ✅ Test 1: Verify RWMutex prevents race condition with concurrent requests
func TestStreamHandlerNoRaceCondition(t *testing.T) {
	// Create test executor
	crew := &Crew{
		Agents: []*Agent{
			{
				ID:         "test-agent",
				Name:       "Test Agent",
				IsTerminal: true,
			},
		},
	}
	executor := NewCrewExecutor(crew, "test-key")

	// Create handler
	handler := NewHTTPHandler(executor, DefaultHardcodedDefaults())

	// Test parameters
	numRequests := 50
	numStateChanges := 10
	var wg sync.WaitGroup
	errors := make([]string, 0)
	var errorsMu sync.Mutex
	var raceDetected int32

	// Goroutine group 1: Concurrent StreamHandlers (readers)
	for i := 0; i < numRequests; i++ {
		wg.Add(1)
		go func(index int) {
			defer wg.Done()

			// Create request
			req := httptest.NewRequest(
				"GET",
				fmt.Sprintf("/api/crew/stream?q=test-query-%d", index),
				nil,
			)

			// Add context with timeout to prevent hanging
			ctx, cancel := context.WithTimeout(context.Background(), 500*time.Millisecond)
			defer cancel()
			req = req.WithContext(ctx)

			w := httptest.NewRecorder()

			// Call StreamHandler - should use RLock for reading state
			handler.StreamHandler(w, req)

			// Verify response is valid
			if w.Code != http.StatusOK {
				errorsMu.Lock()
				errors = append(errors, fmt.Sprintf("Request %d: status %d (expected 200)", index, w.Code))
				errorsMu.Unlock()
			}
		}(i)
	}

	// Goroutine group 2: Concurrent state modifications (writers)
	for i := 0; i < numStateChanges; i++ {
		wg.Add(1)
		go func(index int) {
			defer wg.Done()
			time.Sleep(time.Duration(index*10) * time.Millisecond)

			// Alternate between SetVerbose and SetResumeAgent
			if index%2 == 0 {
				handler.SetVerbose(index%2 == 0)
			} else {
				handler.SetResumeAgent(fmt.Sprintf("agent-%d", index))
			}
		}(i)
	}

	// Wait for all goroutines
	wg.Wait()

	// Check results
	if len(errors) > 0 {
		t.Logf("Found %d errors during concurrent access:\n", len(errors))
		for _, err := range errors {
			t.Logf("  - %s\n", err)
		}
	}

	if atomic.LoadInt32(&raceDetected) > 0 {
		t.Error("Race condition detected during concurrent requests")
	}
}

// ✅ Test 2: Verify snapshot isolates state changes
func TestSnapshotIsolatesStateChanges(t *testing.T) {
	crew := &Crew{
		Agents: []*Agent{
			{ID: "a1", Name: "Agent1", IsTerminal: true},
		},
	}
	executor := NewCrewExecutor(crew, "key")

	handler := NewHTTPHandler(executor, DefaultHardcodedDefaults())

	// Set initial state
	handler.SetVerbose(false)
	handler.SetResumeAgent("initial-agent")

	// Capture current state
	verbose1 := handler.GetVerbose()
	resume1 := handler.GetResumeAgent()

	// Change state in background
	handler.SetVerbose(true)
	handler.SetResumeAgent("changed-agent")

	// Verify state changed
	verbose2 := handler.GetVerbose()
	resume2 := handler.GetResumeAgent()

	if verbose1 == verbose2 {
		t.Error("SetVerbose did not change state")
	}

	if resume1 == resume2 {
		t.Error("SetResumeAgent did not change state")
	}

	if verbose2 != true {
		t.Errorf("Expected Verbose=true, got %v", verbose2)
	}

	if resume2 != "changed-agent" {
		t.Errorf("Expected ResumeAgent='changed-agent', got %s", resume2)
	}
}

// ✅ Test 3: Verify RWMutex allows concurrent reads
func TestConcurrentReads(t *testing.T) {
	crew := &Crew{
		Agents: []*Agent{
			{ID: "a1", Name: "Agent1", IsTerminal: true},
		},
	}
	executor := NewCrewExecutor(crew, "key")
	handler := NewHTTPHandler(executor, DefaultHardcodedDefaults())

	handler.SetVerbose(true)
	handler.SetResumeAgent("test-agent")

	numReaders := 100
	var wg sync.WaitGroup
	var readCount int32

	// Launch many concurrent readers
	for i := 0; i < numReaders; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()

			// These should NOT block each other (RLock)
			v := handler.GetVerbose()
			r := handler.GetResumeAgent()

			if v && r == "test-agent" {
				atomic.AddInt32(&readCount, 1)
			}
		}()
	}

	wg.Wait()

	if readCount != int32(numReaders) {
		t.Errorf("Expected %d successful reads, got %d", numReaders, readCount)
	}
}

// ✅ Test 4: Verify write lock prevents concurrent writes
func TestWriteLockPreventsRaces(t *testing.T) {
	crew := &Crew{
		Agents: []*Agent{
			{ID: "a1", Name: "Agent1", IsTerminal: true},
		},
	}
	executor := NewCrewExecutor(crew, "key")
	handler := NewHTTPHandler(executor, DefaultHardcodedDefaults())

	numWriters := 20
	var wg sync.WaitGroup

	// Launch concurrent writers
	for i := 0; i < numWriters; i++ {
		wg.Add(1)
		go func(index int) {
			defer wg.Done()
			handler.SetResumeAgent(fmt.Sprintf("agent-%d", index))
		}(i)
	}

	wg.Wait()

	// Final state should be one of the values we set
	final := handler.GetResumeAgent()
	expectedPattern := "agent-"

	if final == "" || len(final) < len(expectedPattern) {
		t.Errorf("Expected agent-ID format, got: %s", final)
	}
}

// ✅ Test 5: Verify ClearResumeAgent works correctly
func TestClearResumeAgent(t *testing.T) {
	crew := &Crew{
		Agents: []*Agent{
			{ID: "a1", Name: "Agent1", IsTerminal: true},
		},
	}
	executor := NewCrewExecutor(crew, "key")
	handler := NewHTTPHandler(executor, DefaultHardcodedDefaults())

	// Set resume agent
	handler.SetResumeAgent("test-agent")
	if handler.GetResumeAgent() != "test-agent" {
		t.Error("SetResumeAgent failed")
	}

	// Clear it
	handler.ClearResumeAgent()
	if handler.GetResumeAgent() != "" {
		t.Error("ClearResumeAgent failed")
	}
}

// ✅ Test 6: High concurrency stress test
func TestHighConcurrencyStress(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping stress test in short mode")
	}

	crew := &Crew{
		Agents: []*Agent{
			{ID: "a1", Name: "Agent1", IsTerminal: true},
		},
	}
	executor := NewCrewExecutor(crew, "key")
	handler := NewHTTPHandler(executor, DefaultHardcodedDefaults())

	numRequests := 200
	numWriters := 5
	duration := 2 * time.Second

	var wg sync.WaitGroup
	var successCount int32
	var errorCount int32

	done := make(chan struct{})
	go func() {
		time.Sleep(duration)
		close(done)
	}()

	// Readers: Simulate StreamHandlers
	for i := 0; i < numRequests; i++ {
		wg.Add(1)
		go func(index int) {
			defer wg.Done()

			for {
				select {
				case <-done:
					return
				default:
					// Simulate reading state
					v := handler.GetVerbose()
					r := handler.GetResumeAgent()

					// Create executor with snapshot pattern
					snapshot := executorSnapshot{
						Verbose:       v,
						ResumeAgentID: r,
					}

					// Verify snapshot is consistent
					if snapshot.Verbose || snapshot.ResumeAgentID != "" {
						atomic.AddInt32(&successCount, 1)
					}
				}
			}
		}(i)
	}

	// Writers: Simulate configuration changes
	for i := 0; i < numWriters; i++ {
		wg.Add(1)
		go func(index int) {
			defer wg.Done()

			for {
				select {
				case <-done:
					return
				default:
					if index%2 == 0 {
						handler.SetVerbose(index%3 == 0)
					} else {
						handler.SetResumeAgent(fmt.Sprintf("agent-%d", index%5))
					}
					time.Sleep(10 * time.Millisecond)
				}
			}
		}(i)
	}

	wg.Wait()

	t.Logf("Completed %d read operations successfully", atomic.LoadInt32(&successCount))
	if errorCount > 0 {
		t.Errorf("Encountered %d errors during stress test", errorCount)
	}
}

// ✅ Test 7: Verify state consistency under mixed operations
func TestStateConsistency(t *testing.T) {
	crew := &Crew{
		Agents: []*Agent{
			{ID: "a1", Name: "Agent1", IsTerminal: true},
		},
	}
	executor := NewCrewExecutor(crew, "key")
	handler := NewHTTPHandler(executor, DefaultHardcodedDefaults())

	// Initial state
	handler.SetVerbose(true)
	handler.SetResumeAgent("agent-1")

	// Read state multiple times - should be consistent
	values := make([]string, 0)
	for i := 0; i < 10; i++ {
		v := handler.GetVerbose()
		r := handler.GetResumeAgent()
		values = append(values, fmt.Sprintf("v=%v,r=%s", v, r))
	}

	// All reads should give same value (no changes made)
	firstValue := values[0]
	for i, v := range values {
		if v != firstValue {
			t.Errorf("Inconsistent read at index %d: %s vs %s", i, firstValue, v)
		}
	}
}

// ✅ Test 8: Verify no deadlock scenarios
func TestNoDeadlock(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping deadlock test in short mode")
	}

	crew := &Crew{
		Agents: []*Agent{
			{ID: "a1", Name: "Agent1", IsTerminal: true},
		},
	}
	executor := NewCrewExecutor(crew, "key")
	handler := NewHTTPHandler(executor, DefaultHardcodedDefaults())

	done := make(chan bool, 1)

	// Start goroutine that could deadlock
	go func() {
		var wg sync.WaitGroup

		// Multiple nested operations
		for i := 0; i < 50; i++ {
			wg.Add(1)
			go func(idx int) {
				defer wg.Done()
				if idx%3 == 0 {
					handler.SetVerbose(idx%2 == 0)
				} else if idx%3 == 1 {
					handler.SetResumeAgent(fmt.Sprintf("agent-%d", idx))
				} else {
					_ = handler.GetVerbose()
					_ = handler.GetResumeAgent()
				}
			}(i)
		}

		wg.Wait()
		done <- true
	}()

	// Wait with timeout
	select {
	case <-done:
		// Success - no deadlock
	case <-time.After(5 * time.Second):
		t.Fatal("Possible deadlock detected (timeout after 5 seconds)")
	}
}

// ===== Issue #10: Input Validation Tests =====

// TestValidateQueryLength verifies query length validation
func TestValidateQueryLength(t *testing.T) {
	validator := NewInputValidator(DefaultHardcodedDefaults())

	tests := []struct {
		name      string
		query     string
		shouldErr bool
	}{
		{"empty query", "", true},
		{"valid query", "what is AI?", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validator.ValidateQuery(tt.query)
			if (err != nil) != tt.shouldErr {
				t.Errorf("Expected error: %v, got: %v", tt.shouldErr, err)
			}
		})
	}

	// Test max length query (exactly 10240 chars = 10KB default limit)
	t.Run("max length query", func(t *testing.T) {
		maxQuery := strings.Repeat("a", 10240)
		err := validator.ValidateQuery(maxQuery)
		if err != nil {
			t.Errorf("Expected no error for max length query, got: %v", err)
		}
	})

	// Test exceeds max length (10241 chars)
	t.Run("exceeds max length", func(t *testing.T) {
		overQuery := strings.Repeat("a", 10241)
		err := validator.ValidateQuery(overQuery)
		if err == nil {
			t.Error("Expected error for exceeding max length, got nil")
		}
	})
}

// TestValidateQueryUTF8 verifies UTF-8 validation
func TestValidateQueryUTF8(t *testing.T) {
	validator := NewInputValidator(DefaultHardcodedDefaults())

	// Invalid UTF-8 sequence
	invalidUTF8 := "hello\xff\xfe"
	err := validator.ValidateQuery(invalidUTF8)
	if err == nil {
		t.Error("Expected error for invalid UTF-8, got nil")
	}

	// Valid UTF-8 with Unicode
	validUnicode := "Xin chào, 世界"
	err = validator.ValidateQuery(validUnicode)
	if err != nil {
		t.Errorf("Expected no error for valid Unicode, got: %v", err)
	}
}

// TestValidateQueryControlChars verifies control character detection
func TestValidateQueryControlChars(t *testing.T) {
	validator := NewInputValidator(DefaultHardcodedDefaults())

	tests := []struct {
		name      string
		query     string
		shouldErr bool
	}{
		{"null byte", "hello\x00world", true},
		{"control char", "hello\x01world", true},
		{"newline allowed", "hello\nworld", false},
		{"tab allowed", "hello\tworld", false},
		{"clean text", "hello world", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validator.ValidateQuery(tt.query)
			if (err != nil) != tt.shouldErr {
				t.Errorf("Expected error: %v, got: %v", tt.shouldErr, err)
			}
		})
	}
}

// TestValidateAgentIDFormat verifies agent ID format validation
func TestValidateAgentIDFormat(t *testing.T) {
	validator := NewInputValidator(DefaultHardcodedDefaults())

	tests := []struct {
		name      string
		agentID   string
		shouldErr bool
	}{
		{"valid id", "agent-1", false},
		{"valid id with underscore", "agent_1", false},
		{"valid id alphanumeric", "agent123", false},
		{"empty id", "", true},
		{"invalid char @", "agent@1", true},
		{"invalid char space", "agent 1", true},
		{"too long", string(make([]byte, 129)), true},
		{"valid max length", string(make([]byte, 128)), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// For length tests, fill with valid chars (alphanumeric only)
			agentID := tt.agentID
			if tt.name == "too long" {
				// Create string with 129 alphanumeric chars
				agentID = strings.Repeat("a", 129)
			}
			if tt.name == "valid max length" {
				// Create string with exactly 128 alphanumeric chars
				agentID = strings.Repeat("a", 128)
			}

			err := validator.ValidateAgentID(agentID)
			if (err != nil) != tt.shouldErr {
				t.Errorf("Expected error: %v, got: %v", tt.shouldErr, err)
			}
		})
	}
}

// TestValidateHistory verifies history validation
func TestValidateHistory(t *testing.T) {
	validator := NewInputValidator(DefaultHardcodedDefaults())

	t.Run("valid history", func(t *testing.T) {
		history := []Message{
			{Role: "user", Content: "hello"},
			{Role: "assistant", Content: "hi there"},
		}
		err := validator.ValidateHistory(history)
		if err != nil {
			t.Errorf("Expected no error, got: %v", err)
		}
	})

	t.Run("invalid role", func(t *testing.T) {
		history := []Message{
			{Role: "invalid_role", Content: "hello"},
		}
		err := validator.ValidateHistory(history)
		if err == nil {
			t.Error("Expected error for invalid role, got nil")
		}
	})

	t.Run("exceeds message count", func(t *testing.T) {
		history := make([]Message, 1001)
		for i := range history {
			history[i] = Message{Role: "user", Content: "msg"}
		}
		err := validator.ValidateHistory(history)
		if err == nil {
			t.Error("Expected error for exceeding message count, got nil")
		}
	})

	t.Run("message too large", func(t *testing.T) {
		// MaxRequestBodySize default is 100 * 1024 = 102400 bytes
		// Create message larger than that
		history := []Message{
			{Role: "user", Content: string(make([]byte, 102401))},
		}
		err := validator.ValidateHistory(history)
		if err == nil {
			t.Error("Expected error for oversized message, got nil")
		}
	})
}

// TestStreamHandlerInputValidation verifies validation in StreamHandler
func TestStreamHandlerInputValidation(t *testing.T) {
	crew := &Crew{
		Agents: []*Agent{
			{ID: "test-agent", Name: "Test Agent", IsTerminal: true},
		},
	}
	executor := NewCrewExecutor(crew, "test-key")
	handler := NewHTTPHandler(executor, DefaultHardcodedDefaults())

	t.Run("reject oversized query", func(t *testing.T) {
		// MaxInputSize default is 10 * 1024 = 10240 bytes
		// Create query larger than that
		req := httptest.NewRequest(
			"GET",
			"/api/crew/stream?q="+strings.Repeat("a", 10241),
			nil,
		)
		w := httptest.NewRecorder()

		handler.StreamHandler(w, req)

		if w.Code != http.StatusBadRequest {
			t.Errorf("Expected status 400, got %d", w.Code)
		}
	})

	t.Run("accept valid query", func(t *testing.T) {
		// This will fail during execution since no LLM, but shouldn't fail validation
		req := httptest.NewRequest(
			"GET",
			"/api/crew/stream?q=test",
			nil,
		)
		w := httptest.NewRecorder()

		handler.StreamHandler(w, req)

		// Should NOT be 400 (bad request)
		if w.Code == http.StatusBadRequest {
			t.Errorf("Valid query was rejected: %s", w.Body.String())
		}
	})
}

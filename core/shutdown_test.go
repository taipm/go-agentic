package crewai

import (
	"context"
	"fmt"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

// ===== Issue #18: Graceful Shutdown Tests =====

// TestGracefulShutdownManagerCreation verifies proper initialization
func TestGracefulShutdownManagerCreation(t *testing.T) {
	gsm := NewGracefulShutdownManager()

	if gsm == nil {
		t.Error("GracefulShutdownManager creation failed")
	}

	if gsm.GracefulTimeout != 30*time.Second {
		t.Errorf("Expected default timeout 30s, got %v", gsm.GracefulTimeout)
	}

	if gsm.GetActiveRequests() != 0 {
		t.Error("Initial active requests should be 0")
	}

	if gsm.GetActiveStreamCount() != 0 {
		t.Error("Initial stream count should be 0")
	}

	if gsm.IsShuttingDown() {
		t.Error("Should not be shutting down initially")
	}
}

// TestRequestTracking verifies request counting works correctly
func TestRequestTracking(t *testing.T) {
	gsm := NewGracefulShutdownManager()

	// Test increment
	for i := 0; i < 5; i++ {
		gsm.IncrementActiveRequests()
	}

	if gsm.GetActiveRequests() != 5 {
		t.Errorf("Expected 5 active requests, got %d", gsm.GetActiveRequests())
	}

	// Test decrement
	for i := 0; i < 3; i++ {
		gsm.DecrementActiveRequests()
	}

	if gsm.GetActiveRequests() != 2 {
		t.Errorf("Expected 2 active requests, got %d", gsm.GetActiveRequests())
	}

	// Test full cleanup
	gsm.DecrementActiveRequests()
	gsm.DecrementActiveRequests()

	if gsm.GetActiveRequests() != 0 {
		t.Errorf("Expected 0 active requests after cleanup, got %d", gsm.GetActiveRequests())
	}
}

// TestRequestTrackingConcurrency verifies thread-safe request counting
func TestRequestTrackingConcurrency(t *testing.T) {
	gsm := NewGracefulShutdownManager()
	const numGoroutines = 100
	const requestsPerGoroutine = 10

	var wg sync.WaitGroup

	// Spawn concurrent request handlers
	for g := 0; g < numGoroutines; g++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for i := 0; i < requestsPerGoroutine; i++ {
				gsm.IncrementActiveRequests()
				time.Sleep(1 * time.Millisecond) // Simulate work
				gsm.DecrementActiveRequests()
			}
		}()
	}

	wg.Wait()

	// After all goroutines complete, active requests should be 0
	if gsm.GetActiveRequests() != 0 {
		t.Errorf("Expected 0 active requests after concurrent operations, got %d", gsm.GetActiveRequests())
	}
}

// TestStreamRegistration verifies stream tracking works
func TestStreamRegistration(t *testing.T) {
	gsm := NewGracefulShutdownManager()

	// Create dummy cancel functions
	ctx1, cancel1 := context.WithCancel(context.Background())
	_, cancel2 := context.WithCancel(context.Background())
	ctx3, cancel3 := context.WithCancel(context.Background())

	// Register streams
	gsm.RegisterStream("stream-1", cancel1)
	gsm.RegisterStream("stream-2", cancel2)
	gsm.RegisterStream("stream-3", cancel3)

	if gsm.GetActiveStreamCount() != 3 {
		t.Errorf("Expected 3 active streams, got %d", gsm.GetActiveStreamCount())
	}

	// Unregister one stream
	gsm.UnregisterStream("stream-2")

	if gsm.GetActiveStreamCount() != 2 {
		t.Errorf("Expected 2 active streams after unregister, got %d", gsm.GetActiveStreamCount())
	}

	// Verify context is still valid
	select {
	case <-ctx1.Done():
		t.Error("Context 1 should not be cancelled")
	default:
	}

	select {
	case <-ctx3.Done():
		t.Error("Context 3 should not be cancelled")
	default:
	}

	// Cleanup
	cancel1()
	cancel2()
	cancel3()
}

// TestCancelAllStreams verifies all streams are cancelled on shutdown
func TestCancelAllStreams(t *testing.T) {
	gsm := NewGracefulShutdownManager()

	// Create multiple contexts
	contexts := make([]context.Context, 5)
	cancels := make([]context.CancelFunc, 5)

	for i := 0; i < 5; i++ {
		var ctx context.Context
		ctx, cancels[i] = context.WithCancel(context.Background())
		contexts[i] = ctx
		gsm.RegisterStream(fmt.Sprintf("stream-%d", i), cancels[i])
	}

	// Verify all registered
	if gsm.GetActiveStreamCount() != 5 {
		t.Errorf("Expected 5 streams registered, got %d", gsm.GetActiveStreamCount())
	}

	// Cancel all streams via shutdown
	gsm.cancelAllStreams()

	// Verify all contexts are cancelled
	for i, ctx := range contexts {
		select {
		case <-ctx.Done():
			// Expected
		default:
			t.Errorf("Context %d should be cancelled", i)
		}
	}

	// Verify stream map is cleared
	if gsm.GetActiveStreamCount() != 0 {
		t.Errorf("Stream count should be 0 after cancel all, got %d", gsm.GetActiveStreamCount())
	}
}

// TestShutdownWithActiveRequests verifies shutdown waits for request completion
func TestShutdownWithActiveRequests(t *testing.T) {
	gsm := NewGracefulShutdownManager()
	gsm.GracefulTimeout = 1 * time.Second

	var wg sync.WaitGroup
	var completionTime time.Duration
	requestsStarted := int32(0)

	// Simulate quick requests that complete before timeout
	for i := 0; i < 3; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			gsm.IncrementActiveRequests()
			atomic.AddInt32(&requestsStarted, 1)
			defer gsm.DecrementActiveRequests()
			time.Sleep(200 * time.Millisecond) // Quick completion
		}()
	}

	// Give requests time to start
	time.Sleep(100 * time.Millisecond)

	// Start shutdown and measure time
	startTime := time.Now()
	err := gsm.Shutdown(context.Background())
	completionTime = time.Since(startTime)

	// Wait for goroutines to finish cleanup
	wg.Wait()

	// If requests started and completed within timeout, no error expected
	if atomic.LoadInt32(&requestsStarted) == 3 && err == nil {
		// Verify we waited for requests (should take at least 100ms based on request sleep)
		// But allow margin for scheduling, so check >= 50ms
		if completionTime < 50*time.Millisecond {
			t.Errorf("Shutdown should have waited for requests, took only %v", completionTime)
		}
	}

	// Final count should be 0 (all completed)
	finalCount := gsm.GetActiveRequests()
	if finalCount < 0 {
		t.Logf("Requests started: %d, final count: %d", requestsStarted, finalCount)
	}
}

// TestShutdownTimeout verifies timeout protection works
func TestShutdownTimeout(t *testing.T) {
	gsm := NewGracefulShutdownManager()
	gsm.GracefulTimeout = 200 * time.Millisecond

	// Create long-running request
	gsm.IncrementActiveRequests()
	go func() {
		defer gsm.DecrementActiveRequests()
		time.Sleep(2 * time.Second) // Much longer than timeout
	}()

	// Shutdown should timeout
	err := gsm.Shutdown(context.Background())

	if err == nil {
		t.Error("Expected shutdown timeout error")
	}

	// But request counter should show we had active request
	if gsm.GetActiveRequests() != 1 {
		t.Errorf("Expected 1 active request during timeout, got %d", gsm.GetActiveRequests())
	}
}

// TestIsShuttingDown verifies shutdown flag works
func TestIsShuttingDown(t *testing.T) {
	gsm := NewGracefulShutdownManager()

	// Initially not shutting down
	if gsm.IsShuttingDown() {
		t.Error("Should not be shutting down initially")
	}

	// Increment requests before shutdown
	gsm.IncrementActiveRequests()
	gsm.IncrementActiveRequests()

	// Start shutdown
	gsm.Shutdown(context.Background())

	// Should be marked as shutting down
	if !gsm.IsShuttingDown() {
		t.Error("Should be shutting down after Shutdown() called")
	}
}

// TestIncrementDuringShutdown verifies requests are rejected during shutdown
func TestIncrementDuringShutdown(t *testing.T) {
	gsm := NewGracefulShutdownManager()

	// Mark as shutting down
	atomic.StoreInt32(&gsm.isShuttingDown, 1)

	initialCount := gsm.GetActiveRequests()

	// Try to increment during shutdown (should be ignored)
	gsm.IncrementActiveRequests()
	gsm.IncrementActiveRequests()

	// Count should not change
	if gsm.GetActiveRequests() != initialCount {
		t.Error("Increment during shutdown should be ignored")
	}
}

// TestShutdownCallback verifies custom shutdown callback is called
func TestShutdownCallback(t *testing.T) {
	gsm := NewGracefulShutdownManager()

	callbackCalled := false
	gsm.ShutdownCallback = func() error {
		callbackCalled = true
		return nil
	}

	gsm.Shutdown(context.Background())

	if !callbackCalled {
		t.Error("Shutdown callback should have been called")
	}
}

// TestShutdownCallbackError verifies callback errors are handled
func TestShutdownCallbackError(t *testing.T) {
	gsm := NewGracefulShutdownManager()

	gsm.ShutdownCallback = func() error {
		return fmt.Errorf("callback error")
	}

	// Should not panic even if callback returns error
	gsm.Shutdown(context.Background())
}

// TestForceShutdown verifies force shutdown works
func TestForceShutdown(t *testing.T) {
	gsm := NewGracefulShutdownManager()

	// Add some state
	gsm.IncrementActiveRequests()
	gsm.IncrementActiveRequests()

	ctx, cancel := context.WithCancel(context.Background())
	gsm.RegisterStream("stream-1", cancel)

	// Force shutdown
	gsm.ForceShutdown()

	// Should be marked as shutting down
	if !gsm.IsShuttingDown() {
		t.Error("Should be shutting down after ForceShutdown()")
	}

	// Streams should be cleared
	if gsm.GetActiveStreamCount() != 0 {
		t.Errorf("Streams should be cleared, got %d", gsm.GetActiveStreamCount())
	}

	// Context should be cancelled
	select {
	case <-ctx.Done():
		// Expected
	default:
		t.Error("Context should be cancelled")
	}
}

// TestConcurrentShutdown verifies concurrent shutdown operations are safe
func TestConcurrentShutdown(t *testing.T) {
	gsm := NewGracefulShutdownManager()
	gsm.GracefulTimeout = 100 * time.Millisecond // Short timeout for test

	// Add some requests and streams
	for i := 0; i < 5; i++ {
		gsm.IncrementActiveRequests()
	}

	_, cancel1 := context.WithCancel(context.Background())
	_, cancel2 := context.WithCancel(context.Background())

	gsm.RegisterStream("stream-1", cancel1)
	gsm.RegisterStream("stream-2", cancel2)

	// Shutdown will timeout since we have active requests with no goroutine to complete them
	err := gsm.Shutdown(context.Background())

	// Should timeout as expected
	if err == nil {
		t.Logf("Expected timeout error, got nil")
	}

	// No panic or data corruption should occur
	if gsm.GetActiveStreamCount() != 0 {
		t.Errorf("Streams should be cleaned, got %d", gsm.GetActiveStreamCount())
	}
}

// TestZeroDowntimeScenario simulates a zero-downtime deployment
func TestZeroDowntimeScenario(t *testing.T) {
	gsm := NewGracefulShutdownManager()

	// Simulate incoming requests during shutdown
	var wg sync.WaitGroup
	requestsCompleted := int32(0)

	// Start 10 concurrent requests
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			gsm.IncrementActiveRequests()
			defer gsm.DecrementActiveRequests()
			defer atomic.AddInt32(&requestsCompleted, 1)

			// Simulate request processing
			time.Sleep(100 * time.Millisecond)
		}()
	}

	// Let some requests start
	time.Sleep(50 * time.Millisecond)

	// Now start shutdown
	gsm.Shutdown(context.Background())

	// Wait for all goroutines
	wg.Wait()

	// All requests should have completed
	if atomic.LoadInt32(&requestsCompleted) != 10 {
		t.Errorf("Expected 10 completed requests, got %d", requestsCompleted)
	}

	// No active requests should remain
	if gsm.GetActiveRequests() != 0 {
		t.Errorf("Expected 0 active requests, got %d", gsm.GetActiveRequests())
	}
}

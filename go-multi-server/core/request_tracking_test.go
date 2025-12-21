package crewai

import (
	"context"
	"fmt"
	"strings"
	"testing"
	"time"
)

// TestGenerateRequestID verifies unique ID generation
func TestGenerateRequestID(t *testing.T) {
	id1 := GenerateRequestID()
	id2 := GenerateRequestID()

	if id1 == "" {
		t.Fatal("Generated ID should not be empty")
	}

	if id1 == id2 {
		t.Fatal("Generated IDs should be unique")
	}

	// Check UUID format (should be 36 chars with hyphens)
	if len(id1) != 36 {
		t.Fatalf("UUID should be 36 chars, got %d", len(id1))
	}

	// Check for UUID hyphens
	if strings.Count(id1, "-") != 4 {
		t.Fatalf("UUID should have 4 hyphens, got %d", strings.Count(id1, "-"))
	}
}

// TestGenerateShortRequestID verifies short ID generation
func TestGenerateShortRequestID(t *testing.T) {
	id := GenerateShortRequestID()

	if !strings.HasPrefix(id, "req-") {
		t.Fatalf("Short ID should start with 'req-', got %s", id)
	}

	// Should be "req-" + 12 chars = 16 total
	if len(id) != 16 {
		t.Fatalf("Short ID should be 16 chars, got %d", len(id))
	}
}

// TestGetRequestID tests context request ID retrieval
func TestGetRequestID(t *testing.T) {
	// Test missing ID
	ctx := context.Background()
	id := GetRequestID(ctx)
	if id != "unknown" {
		t.Fatalf("Missing ID should return 'unknown', got %s", id)
	}

	// Test present ID
	ctx = context.WithValue(ctx, RequestIDKey, "test-123")
	id = GetRequestID(ctx)
	if id != "test-123" {
		t.Fatalf("Should return stored ID, got %s", id)
	}

	// Test nil context
	id = GetRequestID(nil)
	if id != "unknown" {
		t.Fatalf("Nil context should return 'unknown', got %s", id)
	}
}

// TestGetOrCreateRequestID tests ID creation if missing
func TestGetOrCreateRequestID(t *testing.T) {
	// Test creation
	ctx := context.Background()
	id, newCtx := GetOrCreateRequestID(ctx)

	if id == "" || id == "unknown" {
		t.Fatal("Should create new ID if missing")
	}

	// Test retrieval from new context
	id2, _ := GetOrCreateRequestID(newCtx)
	if id != id2 {
		t.Fatal("Should retrieve same ID from modified context")
	}

	// Test with existing ID
	ctx = context.WithValue(context.Background(), RequestIDKey, "existing-id")
	id, _ = GetOrCreateRequestID(ctx)
	if id != "existing-id" {
		t.Fatalf("Should keep existing ID, got %s", id)
	}
}

// TestRequestMetadataAddEvent tests event tracking
func TestRequestMetadataAddEvent(t *testing.T) {
	meta := &RequestMetadata{
		ID:      "test-123",
		ShortID: "req-test",
	}

	meta.AddEvent("agent_thinking", "orchestrator", "", map[string]string{"msg": "thinking"})
	meta.AddEvent("tool_call", "executor", "GetCPUUsage", map[string]int{"timeout": 5})

	if len(meta.Events) != 2 {
		t.Fatalf("Should have 2 events, got %d", len(meta.Events))
	}

	if meta.Events[0].Type != "agent_thinking" {
		t.Fatal("First event should be agent_thinking")
	}

	if meta.Events[1].Tool != "GetCPUUsage" {
		t.Fatal("Second event should have tool name")
	}
}

// TestRequestMetadataCounters tests counter increments
func TestRequestMetadataCounters(t *testing.T) {
	meta := &RequestMetadata{ID: "test-123"}

	meta.IncrementAgentCalls()
	meta.IncrementAgentCalls()
	meta.IncrementToolCalls()

	if meta.AgentCalls != 2 {
		t.Fatalf("Should have 2 agent calls, got %d", meta.AgentCalls)
	}

	if meta.ToolCalls != 1 {
		t.Fatalf("Should have 1 tool call, got %d", meta.ToolCalls)
	}
}

// TestRequestMetadataStatus tests status tracking
func TestRequestMetadataStatus(t *testing.T) {
	meta := &RequestMetadata{ID: "test-123"}

	meta.SetStatus("success", "")
	if meta.Status != "success" {
		t.Fatal("Status should be set to success")
	}

	meta.SetStatus("error", "Something went wrong")
	if meta.Status != "error" || meta.ErrorMessage != "Something went wrong" {
		t.Fatal("Status and error message should be set")
	}
}

// TestRequestMetadataFinalize tests request completion
func TestRequestMetadataFinalize(t *testing.T) {
	meta := &RequestMetadata{
		ID:        "test-123",
		StartTime: time.Now(),
	}

	time.Sleep(10 * time.Millisecond)
	meta.Finalize()

	if meta.EndTime.IsZero() {
		t.Fatal("End time should be set")
	}

	if meta.Duration == 0 {
		t.Fatal("Duration should be calculated")
	}

	if meta.Status == "" {
		t.Fatal("Default status should be 'success'")
	}
}

// TestRequestMetadataGetSnapshot tests snapshot creation
func TestRequestMetadataGetSnapshot(t *testing.T) {
	meta := &RequestMetadata{
		ID:        "test-123",
		UserInput: "Check CPU",
		Status:    "success",
	}

	meta.AddEvent("test", "agent", "", nil)

	snapshot := meta.GetSnapshot()

	if snapshot.ID != "test-123" {
		t.Fatal("Snapshot should preserve ID")
	}

	if len(snapshot.Events) != 1 {
		t.Fatal("Snapshot should preserve events")
	}

	// Modify original
	meta.UserInput = "Modified"

	// Snapshot should not change
	if snapshot.UserInput != "Check CPU" {
		t.Fatal("Snapshot should be independent copy")
	}
}

// TestRequestStorageBasic tests basic store operations
func TestRequestStorageBasic(t *testing.T) {
	store := NewRequestStore(100)

	meta := &RequestMetadata{
		ID:      "req-1",
		ShortID: "req-1",
	}

	store.Add(meta)

	if store.Size() != 1 {
		t.Fatalf("Store should have 1 item, got %d", store.Size())
	}

	retrieved := store.Get("req-1")
	if retrieved == nil || retrieved.ID != "req-1" {
		t.Fatal("Should retrieve stored request")
	}

	// Test non-existent request
	retrieved = store.Get("non-existent")
	if retrieved != nil {
		t.Fatal("Non-existent request should return nil")
	}
}

// TestRequestStoreMaxSize tests max size limit
func TestRequestStoreMaxSize(t *testing.T) {
	store := NewRequestStore(5)

	// Add 7 requests (exceeds max of 5)
	for i := 1; i <= 7; i++ {
		id := fmt.Sprintf("req-%d", i)
		meta := &RequestMetadata{
			ID:      id,
			ShortID: id,
		}
		store.Add(meta)
	}

	if store.Size() > 5 {
		t.Fatalf("Store should respect max size of 5, got %d", store.Size())
	}

	// Oldest requests should be removed
	if store.Get("req-1") != nil {
		t.Fatal("Oldest request should have been removed")
	}

	// Newest requests should still be there
	if store.Get("req-7") == nil {
		t.Fatal("Newest request should still be in store")
	}
}

// TestRequestStoreGetAll tests GetAll method
func TestRequestStoreGetAll(t *testing.T) {
	store := NewRequestStore(100)

	for i := 1; i <= 3; i++ {
		id := fmt.Sprintf("req-%d", i)
		meta := &RequestMetadata{ID: id}
		store.Add(meta)
	}

	all := store.GetAll()
	if len(all) != 3 {
		t.Fatalf("GetAll should return all requests, got %d", len(all))
	}
}

// TestRequestStoreGetRecent tests GetRecent method
func TestRequestStoreGetRecent(t *testing.T) {
	store := NewRequestStore(100)

	for i := 1; i <= 5; i++ {
		id := fmt.Sprintf("req-%d", i)
		meta := &RequestMetadata{
			ID:        id,
			StartTime: time.Now().Add(time.Duration(i) * time.Second),
		}
		store.Add(meta)
	}

	recent := store.GetRecent(2)
	if len(recent) != 2 {
		t.Fatalf("GetRecent(2) should return 2 items, got %d", len(recent))
	}

	// Most recent should be first
	if recent[0].ID != "req-5" {
		t.Fatalf("Most recent request should be first, got %s", recent[0].ID)
	}
}

// TestRequestStoreGetByStatus tests filtering by status
func TestRequestStoreGetByStatus(t *testing.T) {
	store := NewRequestStore(100)

	store.Add(&RequestMetadata{ID: "req-1", Status: "success"})
	store.Add(&RequestMetadata{ID: "req-2", Status: "error"})
	store.Add(&RequestMetadata{ID: "req-3", Status: "success"})

	successes := store.GetByStatus("success")
	if len(successes) != 2 {
		t.Fatalf("Should find 2 success requests, got %d", len(successes))
	}

	errors := store.GetByStatus("error")
	if len(errors) != 1 {
		t.Fatalf("Should find 1 error request, got %d", len(errors))
	}
}

// TestRequestStoreGetStats tests statistics
func TestRequestStoreGetStats(t *testing.T) {
	store := NewRequestStore(100)

	store.Add(&RequestMetadata{
		ID:         "req-1",
		Status:     "success",
		AgentCalls: 2,
		ToolCalls:  3,
	})
	store.Add(&RequestMetadata{
		ID:         "req-2",
		Status:     "error",
		AgentCalls: 1,
		ToolCalls:  1,
	})

	stats := store.GetStats()

	if stats["total_requests"] != 2 {
		t.Fatalf("Should report 2 total requests, got %d", stats["total_requests"])
	}

	if stats["total_agent_calls"] != 3 {
		t.Fatalf("Should report 3 total agent calls, got %d", stats["total_agent_calls"])
	}

	if stats["total_tool_calls"] != 4 {
		t.Fatalf("Should report 4 total tool calls, got %d", stats["total_tool_calls"])
	}

	statusCounts := stats["requests"].(map[string]int)
	if statusCounts["success"] != 1 || statusCounts["error"] != 1 {
		t.Fatal("Status counts should be accurate")
	}
}

// TestRequestStoreClear tests store clearing
func TestRequestStoreClear(t *testing.T) {
	store := NewRequestStore(100)

	store.Add(&RequestMetadata{ID: "req-1"})
	store.Add(&RequestMetadata{ID: "req-2"})

	if store.Size() != 2 {
		t.Fatal("Store should have 2 items before clear")
	}

	store.Clear()

	if store.Size() != 0 {
		t.Fatal("Store should be empty after clear")
	}
}

// TestRequestStoreCleanup tests old request removal
func TestRequestStoreCleanup(t *testing.T) {
	store := NewRequestStore(100)

	// Add old request
	oldMeta := &RequestMetadata{
		ID:        "req-old",
		StartTime: time.Now().Add(-2 * time.Hour),
		EndTime:   time.Now().Add(-2 * time.Hour),
	}
	store.Add(oldMeta)

	// Add recent request
	newMeta := &RequestMetadata{
		ID:        "req-new",
		StartTime: time.Now(),
		EndTime:   time.Now(),
	}
	store.Add(newMeta)

	if store.Size() != 2 {
		t.Fatal("Store should have 2 items before cleanup")
	}

	// Cleanup requests older than 1 hour
	store.Cleanup(1 * time.Hour)

	if store.Size() != 1 {
		t.Fatal("Store should have 1 item after cleanup")
	}

	if store.Get("req-old") != nil {
		t.Fatal("Old request should be removed")
	}

	if store.Get("req-new") == nil {
		t.Fatal("Recent request should remain")
	}
}

// TestRequestStoreThreadSafety tests concurrent access
func TestRequestStoreThreadSafety(t *testing.T) {
	store := NewRequestStore(100)

	// Concurrent adds
	done := make(chan bool)

	for i := 1; i <= 5; i++ {
		go func(id int) {
			meta := &RequestMetadata{ID: "req-" + string(rune(id))}
			store.Add(meta)
			done <- true
		}(i)
	}

	// Wait for all goroutines
	for i := 0; i < 5; i++ {
		<-done
	}

	if store.Size() != 5 {
		t.Fatalf("Should have 5 items after concurrent adds, got %d", store.Size())
	}
}

// TestRequestMetadataSummary tests summary generation
func TestRequestMetadataSummary(t *testing.T) {
	meta := &RequestMetadata{
		ID:         "test-123",
		ShortID:    "req-test",
		UserInput:  "Check system status",
		Status:     "success",
		Duration:   1500 * time.Millisecond,
		AgentCalls: 2,
		ToolCalls:  3,
	}

	meta.Events = []Event{
		{Type: "agent_thinking"},
		{Type: "tool_call"},
	}

	summary := meta.Summary()

	if !strings.Contains(summary, "req-test") {
		t.Fatal("Summary should include request ID")
	}

	if !strings.Contains(summary, "success") {
		t.Fatal("Summary should include status")
	}

	if !strings.Contains(summary, "2") {
		t.Fatal("Summary should include agent call count")
	}
}

// TestRequestStoreExport tests export functionality
func TestRequestStoreExport(t *testing.T) {
	store := NewRequestStore(100)

	store.Add(&RequestMetadata{
		ID:        "req-1",
		ShortID:   "req-1",
		Status:    "success",
		StartTime: time.Now(),
		EndTime:   time.Now(),
	})

	exported := store.Export()

	if len(exported) != 1 {
		t.Fatalf("Export should return 1 item, got %d", len(exported))
	}

	item := exported[0]
	if item["id"] != "req-1" {
		t.Fatal("Exported item should have ID")
	}

	if item["status"] != "success" {
		t.Fatal("Exported item should have status")
	}
}

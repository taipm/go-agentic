package crewai

import (
	"context"
	"fmt"
	"log"
	"sort"
	"sync"
	"time"

	"github.com/google/uuid"
)

// RequestIDKey is the context key for storing request IDs
const RequestIDKey = "request-id"

// ✅ Issue #17: Request ID Tracking - Distributed request tracking system
// Enables correlation of logs and events across all components

// GenerateRequestID generates a unique request ID using UUID
func GenerateRequestID() string {
	return uuid.New().String()
}

// GenerateShortRequestID generates a shorter request ID for logs (16 chars)
// Format: "req-" + first 12 chars of UUID
func GenerateShortRequestID() string {
	id := uuid.New()
	return "req-" + id.String()[:12]
}

// GetRequestID retrieves request ID from context
// Returns "unknown" if no ID found
func GetRequestID(ctx context.Context) string {
	if ctx == nil {
		return "unknown"
	}
	id, ok := ctx.Value(RequestIDKey).(string)
	if !ok {
		return "unknown"
	}
	return id
}

// GetOrCreateRequestID gets request ID from context or creates new one
func GetOrCreateRequestID(ctx context.Context) (string, context.Context) {
	id := GetRequestID(ctx)
	if id == "unknown" {
		id = GenerateRequestID()
		ctx = context.WithValue(ctx, RequestIDKey, id)
	}
	return id, ctx
}

// Event represents a tracked event in request lifecycle
type Event struct {
	Type      string        `json:"type"`      // agent_thinking, tool_call, agent_response, etc.
	Agent     string        `json:"agent"`     // Agent ID that triggered event
	Tool      string        `json:"tool"`      // Tool name (if applicable)
	Timestamp time.Time     `json:"timestamp"` // When event occurred
	Data      interface{}   `json:"data"`      // Event-specific data
}

// RequestMetadata tracks request execution metadata
// ✅ Issue #17: Complete request lifecycle tracking
type RequestMetadata struct {
	// Identity
	ID        string `json:"id"`         // Unique request ID
	ShortID   string `json:"short_id"`   // Short version for logs
	UserInput string `json:"user_input"` // Original user input

	// Timing
	StartTime time.Time     `json:"start_time"` // When request started
	EndTime   time.Time     `json:"end_time"`   // When request ended
	Duration  time.Duration `json:"duration"`   // Total duration

	// Execution metrics
	AgentCalls int `json:"agent_calls"` // Number of agent calls
	ToolCalls  int `json:"tool_calls"`  // Number of tool calls
	RoundCount int `json:"round_count"` // Execution rounds

	// Status
	Status       string `json:"status"`        // success, error, timeout
	ErrorMessage string `json:"error_message"` // If failed

	// Events
	Events []Event `json:"events"` // All events in order

	// Additional metadata
	Model       string            `json:"model"`        // Primary LLM model
	Metadata    map[string]string `json:"metadata"`     // Custom metadata
	mu          sync.RWMutex      `json:"-"`            // Thread safety
}

// AddEvent adds an event to the request metadata
func (rm *RequestMetadata) AddEvent(eventType, agent, tool string, data interface{}) {
	if rm == nil {
		return
	}
	rm.mu.Lock()
	defer rm.mu.Unlock()

	rm.Events = append(rm.Events, Event{
		Type:      eventType,
		Agent:     agent,
		Tool:      tool,
		Timestamp: time.Now(),
		Data:      data,
	})
}

// IncrementAgentCalls increments agent call counter
func (rm *RequestMetadata) IncrementAgentCalls() {
	if rm == nil {
		return
	}
	rm.mu.Lock()
	defer rm.mu.Unlock()
	rm.AgentCalls++
}

// IncrementToolCalls increments tool call counter
func (rm *RequestMetadata) IncrementToolCalls() {
	if rm == nil {
		return
	}
	rm.mu.Lock()
	defer rm.mu.Unlock()
	rm.ToolCalls++
}

// SetStatus sets request status and optionally error message
func (rm *RequestMetadata) SetStatus(status, errorMsg string) {
	if rm == nil {
		return
	}
	rm.mu.Lock()
	defer rm.mu.Unlock()
	rm.Status = status
	if errorMsg != "" {
		rm.ErrorMessage = errorMsg
	}
}

// Finalize completes the request metadata
func (rm *RequestMetadata) Finalize() {
	if rm == nil {
		return
	}
	rm.mu.Lock()
	defer rm.mu.Unlock()
	rm.EndTime = time.Now()
	rm.Duration = rm.EndTime.Sub(rm.StartTime)
	if rm.Status == "" {
		rm.Status = "success"
	}
}

// GetSnapshot returns a thread-safe copy of metadata
func (rm *RequestMetadata) GetSnapshot() RequestMetadata {
	if rm == nil {
		return RequestMetadata{}
	}
	rm.mu.RLock()
	defer rm.mu.RUnlock()

	// Deep copy events
	eventsCopy := make([]Event, len(rm.Events))
	copy(eventsCopy, rm.Events)

	// Deep copy metadata
	metaCopy := make(map[string]string)
	for k, v := range rm.Metadata {
		metaCopy[k] = v
	}

	return RequestMetadata{
		ID:           rm.ID,
		ShortID:      rm.ShortID,
		UserInput:    rm.UserInput,
		StartTime:    rm.StartTime,
		EndTime:      rm.EndTime,
		Duration:     rm.Duration,
		AgentCalls:   rm.AgentCalls,
		ToolCalls:    rm.ToolCalls,
		RoundCount:   rm.RoundCount,
		Status:       rm.Status,
		ErrorMessage: rm.ErrorMessage,
		Events:       eventsCopy,
		Model:        rm.Model,
		Metadata:     metaCopy,
	}
}

// RequestStore manages request history in-memory
// ✅ Issue #17: Thread-safe request storage with automatic cleanup
type RequestStore struct {
	mu       sync.RWMutex
	requests map[string]*RequestMetadata
	maxSize  int // Keep last N requests
	order    []string // Track insertion order for FIFO cleanup
}

// NewRequestStore creates a new request store
// maxSize: Keep last N requests (default: 1000)
func NewRequestStore(maxSize int) *RequestStore {
	if maxSize <= 0 {
		maxSize = 1000
	}
	return &RequestStore{
		requests: make(map[string]*RequestMetadata),
		maxSize:  maxSize,
		order:    make([]string, 0, maxSize),
	}
}

// Add adds or updates a request in the store
func (rs *RequestStore) Add(meta *RequestMetadata) {
	if rs == nil || meta == nil {
		return
	}
	rs.mu.Lock()
	defer rs.mu.Unlock()

	// Check if already exists
	exists := rs.requests[meta.ID] != nil

	// Store request
	rs.requests[meta.ID] = meta

	// Track insertion order for FIFO cleanup
	if !exists {
		rs.order = append(rs.order, meta.ID)

		// Remove oldest if over capacity
		if len(rs.requests) > rs.maxSize {
			oldestID := rs.order[0]
			rs.order = rs.order[1:]
			delete(rs.requests, oldestID)
			log.Printf("[REQUEST_STORE] Removed oldest request %s (store size: %d/%d)", oldestID, len(rs.requests), rs.maxSize)
		}
	}
}

// Get retrieves a request by ID
func (rs *RequestStore) Get(id string) *RequestMetadata {
	if rs == nil {
		return nil
	}
	rs.mu.RLock()
	defer rs.mu.RUnlock()
	return rs.requests[id]
}

// GetAll returns all stored requests (snapshot)
func (rs *RequestStore) GetAll() map[string]*RequestMetadata {
	if rs == nil {
		return nil
	}
	rs.mu.RLock()
	defer rs.mu.RUnlock()

	// Return a copy to prevent external modification
	copy := make(map[string]*RequestMetadata)
	for k, v := range rs.requests {
		copy[k] = v
	}
	return copy
}

// GetRecent returns last N requests in reverse chronological order
func (rs *RequestStore) GetRecent(limit int) []*RequestMetadata {
	if rs == nil {
		return nil
	}
	rs.mu.RLock()
	defer rs.mu.RUnlock()

	if limit <= 0 {
		limit = 10
	}

	// Create list from order
	var results []*RequestMetadata
	start := len(rs.order) - limit
	if start < 0 {
		start = 0
	}

	// Return in reverse order (most recent first)
	for i := len(rs.order) - 1; i >= start; i-- {
		if i >= 0 && i < len(rs.order) {
			id := rs.order[i]
			if req, ok := rs.requests[id]; ok {
				results = append(results, req)
			}
		}
	}

	return results
}

// GetByStatus returns requests with matching status
func (rs *RequestStore) GetByStatus(status string) []*RequestMetadata {
	if rs == nil {
		return nil
	}
	rs.mu.RLock()
	defer rs.mu.RUnlock()

	var results []*RequestMetadata
	for _, req := range rs.requests {
		if req.Status == status {
			results = append(results, req)
		}
	}
	return results
}

// GetStats returns store statistics
func (rs *RequestStore) GetStats() map[string]interface{} {
	if rs == nil {
		return nil
	}
	rs.mu.RLock()
	defer rs.mu.RUnlock()

	stats := map[string]interface{}{
		"total_requests": len(rs.requests),
		"max_size":       rs.maxSize,
		"requests": map[string]int{
			"success": 0,
			"error":   0,
			"timeout": 0,
			"pending": 0,
		},
		"total_agent_calls": 0,
		"total_tool_calls":  0,
	}

	statusCounts := stats["requests"].(map[string]int)
	var totalAgentCalls int
	var totalToolCalls int

	for _, req := range rs.requests {
		switch req.Status {
		case "success":
			statusCounts["success"]++
		case "error":
			statusCounts["error"]++
		case "timeout":
			statusCounts["timeout"]++
		default:
			statusCounts["pending"]++
		}
		totalAgentCalls += req.AgentCalls
		totalToolCalls += req.ToolCalls
	}

	stats["total_agent_calls"] = totalAgentCalls
	stats["total_tool_calls"] = totalToolCalls

	return stats
}

// Clear removes all requests from the store
func (rs *RequestStore) Clear() {
	if rs == nil {
		return
	}
	rs.mu.Lock()
	defer rs.mu.Unlock()
	rs.requests = make(map[string]*RequestMetadata)
	rs.order = make([]string, 0, rs.maxSize)
	log.Printf("[REQUEST_STORE] Store cleared")
}

// Size returns number of stored requests
func (rs *RequestStore) Size() int {
	if rs == nil {
		return 0
	}
	rs.mu.RLock()
	defer rs.mu.RUnlock()
	return len(rs.requests)
}

// Cleanup removes old requests based on age
// olderThan: Remove requests older than this duration
func (rs *RequestStore) Cleanup(olderThan time.Duration) {
	if rs == nil {
		return
	}
	rs.mu.Lock()
	defer rs.mu.Unlock()

	cutoff := time.Now().Add(-olderThan)
	removed := 0

	for id, req := range rs.requests {
		if req.EndTime.Before(cutoff) {
			delete(rs.requests, id)
			// Also remove from order
			for i, oid := range rs.order {
				if oid == id {
					rs.order = append(rs.order[:i], rs.order[i+1:]...)
					break
				}
			}
			removed++
		}
	}

	if removed > 0 {
		log.Printf("[REQUEST_STORE] Cleanup removed %d old requests (older than %v)", removed, olderThan)
	}
}

// Export exports all requests as JSON-compatible format
func (rs *RequestStore) Export() []map[string]interface{} {
	if rs == nil {
		return nil
	}
	rs.mu.RLock()
	defer rs.mu.RUnlock()

	var result []map[string]interface{}
	for _, meta := range rs.requests {
		snapshot := meta.GetSnapshot()
		result = append(result, map[string]interface{}{
			"id":             snapshot.ID,
			"short_id":       snapshot.ShortID,
			"user_input":     snapshot.UserInput,
			"start_time":     snapshot.StartTime,
			"end_time":       snapshot.EndTime,
			"duration_ms":    snapshot.Duration.Milliseconds(),
			"agent_calls":    snapshot.AgentCalls,
			"tool_calls":     snapshot.ToolCalls,
			"round_count":    snapshot.RoundCount,
			"status":         snapshot.Status,
			"error_message":  snapshot.ErrorMessage,
			"event_count":    len(snapshot.Events),
			"model":          snapshot.Model,
		})
	}

	// Sort by start time (most recent first)
	sort.Slice(result, func(i, j int) bool {
		ti := result[i]["start_time"].(time.Time)
		tj := result[j]["start_time"].(time.Time)
		return ti.After(tj)
	})

	return result
}

// Summary returns a human-readable summary of request
func (rm *RequestMetadata) Summary() string {
	if rm == nil {
		return "empty request metadata"
	}
	rm.mu.RLock()
	defer rm.mu.RUnlock()

	return fmt.Sprintf(
		"[%s] Request: %s | Status: %s | Duration: %v | Agents: %d | Tools: %d | Events: %d",
		rm.ShortID,
		truncateString(rm.UserInput, 50),
		rm.Status,
		rm.Duration,
		rm.AgentCalls,
		rm.ToolCalls,
		len(rm.Events),
	)
}

// truncateString truncates a string to max length and adds ellipsis
func truncateString(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen-3] + "..."
}

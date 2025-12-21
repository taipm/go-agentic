package crewai

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sync"
	"time"
)

// StreamRequest represents a request to stream crew execution
type StreamRequest struct {
	Query   string    `json:"query"`
	History []Message `json:"history"`
}

// executorSnapshot safely copies executor state for concurrent access
// This struct is used to safely read mutable fields from CrewExecutor
// under RWMutex protection before creating a request-scoped executor
type executorSnapshot struct {
	Verbose       bool
	ResumeAgentID string
}

// HTTPHandler handles HTTP requests for crew execution
// Uses RWMutex for optimal read-heavy workload (many concurrent StreamHandlers, few SetVerbose/SetResumeAgent calls)
type HTTPHandler struct {
	executor *CrewExecutor
	mu       sync.RWMutex // Changed from sync.Mutex for better concurrency (read-heavy pattern)
}

// NewHTTPHandler creates a new HTTP handler
func NewHTTPHandler(executor *CrewExecutor) *HTTPHandler {
	return &HTTPHandler{
		executor: executor,
	}
}

// StreamHandler handles SSE stream requests
func (h *HTTPHandler) StreamHandler(w http.ResponseWriter, r *http.Request) {
	// Support both GET (EventSource API) and POST methods
	if r.Method != http.MethodGet && r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Parse request - support both JSON body and query parameter
	var req StreamRequest

	// Try to parse JSON body first (for POST requests)
	if r.Method == http.MethodPost {
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			// Fall back to query parameter
			req.Query = r.URL.Query().Get("q")
			if req.Query != "" {
				// Try to unmarshal as JSON (from URL encoded JSON)
				var temp StreamRequest
				if err := json.Unmarshal([]byte(req.Query), &temp); err == nil {
					req = temp
				}
			}
		}
	} else {
		// GET request - parse from query parameter
		req.Query = r.URL.Query().Get("q")
		if req.Query != "" {
			// Try to unmarshal as JSON (from URL encoded JSON)
			var temp StreamRequest
			if err := json.Unmarshal([]byte(req.Query), &temp); err == nil {
				req = temp
			}
		}
	}

	if req.Query == "" {
		http.Error(w, "Query is required", http.StatusBadRequest)
		return
	}

	// Set up SSE response headers
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	// Create a channel for streaming events
	// Buffer size of 100 to prevent deadlock with parallel agent execution
	streamChan := make(chan *StreamEvent, 100)

	// ✅ FIX for Issue #1 (Race Condition): Safely snapshot executor state
	// Using RLock (read lock) allows multiple concurrent StreamHandlers without blocking each other
	// This is optimal for read-heavy pattern: many StreamHandlers (readers) vs few SetVerbose/SetResumeAgent calls (writers)
	h.mu.RLock()
	snapshot := executorSnapshot{
		Verbose:       h.executor.Verbose,       // Protected read
		ResumeAgentID: h.executor.ResumeAgentID, // Protected read
	}
	h.mu.RUnlock()

	// Create a new executor context for this request
	// Use copyHistory to create isolated copy of history (no shared references)
	executor := &CrewExecutor{
		crew:          h.executor.crew,              // Immutable pointer
		apiKey:        h.executor.apiKey,            // Immutable string
		entryAgent:    h.executor.entryAgent,        // Immutable pointer
		history:       copyHistory(req.History),     // ✅ Deep copy for thread safety
		Verbose:       snapshot.Verbose,             // Safe copy from snapshot
		ResumeAgentID: snapshot.ResumeAgentID,       // Safe copy from snapshot
	}

	// ✅ FIX for Issue #8 (Streaming Buffer Race Condition)
	// Use channel closing as synchronization signal instead of separate done channel
	// This eliminates all race conditions:
	// 1. Channel closing guarantees ExecuteStream has finished
	// 2. No separate execErr synchronization needed (happens-before guarantee from Go memory model)
	// 3. Automatic buffer draining when channel closes
	// 4. Most idiomatic Go pattern for goroutine completion
	var execErr error

	go func() {
		defer close(streamChan) // Signal completion by closing channel on exit
		execErr = executor.ExecuteStream(r.Context(), req.Query, streamChan)
	}()

	// Send events to client
	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "Streaming not supported", http.StatusInternalServerError)
		return
	}

	// Send opening message
	SendStreamEvent(w, NewStreamEvent("start", "system", "🚀 Starting crew execution..."))
	flusher.Flush()

	// Event loop - receives from streamChan until it closes
	for {
		select {
		case event, ok := <-streamChan:
			if !ok {
				// Channel closed = ExecuteStream finished
				// streamChan is now guaranteed empty
				// execErr read is synchronized by channel closing (Go memory model)
				if execErr != nil {
					SendStreamEvent(w, NewStreamEvent("error", "system", fmt.Sprintf("Execution error: %v", execErr)))
				} else {
					SendStreamEvent(w, NewStreamEvent("done", "system", "✅ Execution completed"))
				}
				flusher.Flush()
				return
			}
			if event != nil {
				SendStreamEvent(w, event)
				flusher.Flush()
			}

		case <-time.After(30 * time.Second):
			// Keep-alive ping
			SendStreamEvent(w, NewStreamEvent("ping", "system", ""))
			flusher.Flush()

		case <-r.Context().Done():
			// Client disconnected
			log.Println("Client disconnected from stream")
			return
		}
	}
}

// HealthHandler returns health status
func (h *HTTPHandler) HealthHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"status":  "ok",
		"service": "go-crewai-streaming",
	})
}

// ✅ Wrapper methods for thread-safe state modifications
// These methods ensure all writes to CrewExecutor fields go through Write locks
// This prevents race conditions between StreamHandlers and configuration changes

// SetVerbose enables or disables verbose output with proper synchronization
func (h *HTTPHandler) SetVerbose(verbose bool) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.executor.Verbose = verbose
}

// SetResumeAgent sets the agent to resume from with proper synchronization
func (h *HTTPHandler) SetResumeAgent(agentID string) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.executor.ResumeAgentID = agentID
}

// ClearResumeAgent clears the resume agent with proper synchronization
func (h *HTTPHandler) ClearResumeAgent() {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.executor.ResumeAgentID = ""
}

// GetVerbose safely reads the verbose flag with read lock
func (h *HTTPHandler) GetVerbose() bool {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return h.executor.Verbose
}

// GetResumeAgent safely reads the resume agent ID with read lock
func (h *HTTPHandler) GetResumeAgent() string {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return h.executor.ResumeAgentID
}

// StartHTTPServer starts the HTTP server with SSE streaming
func StartHTTPServer(executor *CrewExecutor, port int) error {
	handler := NewHTTPHandler(executor)

	http.HandleFunc("/api/crew/stream", handler.StreamHandler)
	http.HandleFunc("/health", handler.HealthHandler)

	// Serve example client
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/" {
			w.Header().Set("Content-Type", "text/html; charset=utf-8")
			w.Write([]byte(exampleHTMLClient))
			return
		}
		http.NotFound(w, r)
	})

	addr := fmt.Sprintf(":%d", port)
	log.Printf("🚀 HTTP Server starting on http://localhost:%d", port)
	log.Printf("📡 SSE Endpoint: http://localhost:%d/api/crew/stream", port)
	log.Printf("🌐 Web Client: http://localhost:%d", port)

	return http.ListenAndServe(addr, nil)
}

// StartHTTPServerWithCustomUI starts the HTTP server with custom HTML UI
func StartHTTPServerWithCustomUI(executor *CrewExecutor, port int, htmlContent string) error {
	handler := NewHTTPHandler(executor)

	http.HandleFunc("/api/crew/stream", handler.StreamHandler)
	http.HandleFunc("/health", handler.HealthHandler)

	// Serve custom client UI
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/" {
			w.Header().Set("Content-Type", "text/html; charset=utf-8")
			w.Write([]byte(htmlContent))
			return
		}
		http.NotFound(w, r)
	})

	addr := fmt.Sprintf(":%d", port)
	log.Printf("🚀 HTTP Server starting on http://localhost:%d", port)
	log.Printf("📡 SSE Endpoint: http://localhost:%d/api/crew/stream", port)
	log.Printf("🌐 Web Client: http://localhost:%d", port)

	return http.ListenAndServe(addr, nil)
}

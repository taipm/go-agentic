package agentic

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sync"
	"time"
)

// StreamRequest represents a request to stream team execution
type StreamRequest struct {
	Query   string    `json:"query"`
	History []Message `json:"history"`
}

// HTTPHandler handles HTTP requests for team execution
type HTTPHandler struct {
	executor *TeamExecutor
	mu       sync.Mutex
}

// NewHTTPHandler creates a new HTTP handler
func NewHTTPHandler(executor *TeamExecutor) *HTTPHandler {
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
	streamChan := make(chan *StreamEvent, 10)
	defer close(streamChan)

	// Create a new executor context for this request
	h.mu.Lock()
	executor := h.createRequestExecutor()
	h.mu.Unlock()

	// Restore history if provided
	if len(req.History) > 0 {
		executor.history = req.History
	}

	// Run crew execution in a goroutine
	done := make(chan bool)
	var execErr error

	go func() {
		execErr = executor.ExecuteStream(r.Context(), req.Query, streamChan)
		done <- true
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

	// Event loop
	for {
		select {
		case <-done:
			// Execution completed
			if execErr != nil {
				SendStreamEvent(w, NewStreamEvent("error", "system", fmt.Sprintf("Execution error: %v", execErr)))
			} else {
				SendStreamEvent(w, NewStreamEvent("done", "system", "✅ Execution completed"))
			}
			flusher.Flush()
			return

		case event := <-streamChan:
			if event == nil {
				continue
			}
			SendStreamEvent(w, event)
			flusher.Flush()

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
		"service": "go-agentic-streaming",
	})
}

// createRequestExecutor creates a new executor for this request
func (h *HTTPHandler) createRequestExecutor() *TeamExecutor {
	return NewTeamExecutor(h.executor.team, h.executor.apiKey)
}

// StartHTTPServer starts the HTTP server with SSE streaming
func StartHTTPServer(executor *TeamExecutor, port int) error {
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

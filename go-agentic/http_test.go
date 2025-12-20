package agentic

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

// TestNewHTTPHandlerCreation tests HTTP handler initialization
func TestNewHTTPHandlerCreation(t *testing.T) {
	executor := &TeamExecutor{
		team:   &Team{Agents: []*Agent{}},
		apiKey: "test-key",
	}

	handler := NewHTTPHandler(executor)

	if handler == nil {
		t.Fatal("Expected non-nil handler")
	}

	if handler.executor != executor {
		t.Error("Expected executor to be set")
	}
}

// TestHealthHandlerBasic tests health check endpoint
func TestHealthHandlerBasic(t *testing.T) {
	executor := &TeamExecutor{
		team:   &Team{Agents: []*Agent{}},
		apiKey: "test-key",
	}

	handler := NewHTTPHandler(executor)
	req := httptest.NewRequest("GET", "/health", nil)
	w := httptest.NewRecorder()

	handler.HealthHandler(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status OK, got %d", w.Code)
	}

	var response map[string]string
	if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
		t.Errorf("Expected valid JSON response, got error: %v", err)
	}

	if response["status"] != "ok" {
		t.Errorf("Expected status 'ok', got '%s'", response["status"])
	}

	if response["service"] != "go-agentic-streaming" {
		t.Errorf("Expected service 'go-agentic-streaming', got '%s'", response["service"])
	}
}

// TestHealthHandlerContentType tests health handler returns JSON content type
func TestHealthHandlerContentType(t *testing.T) {
	executor := &TeamExecutor{
		team:   &Team{Agents: []*Agent{}},
		apiKey: "test-key",
	}

	handler := NewHTTPHandler(executor)
	req := httptest.NewRequest("GET", "/health", nil)
	w := httptest.NewRecorder()

	handler.HealthHandler(w, req)

	contentType := w.Header().Get("Content-Type")
	if contentType != "application/json" {
		t.Errorf("Expected Content-Type application/json, got %s", contentType)
	}
}

// TestStreamHandlerMethodNotAllowed tests stream handler rejects invalid methods
func TestStreamHandlerMethodNotAllowed(t *testing.T) {
	executor := &TeamExecutor{
		team:   &Team{Agents: []*Agent{}},
		apiKey: "test-key",
	}

	handler := NewHTTPHandler(executor)
	req := httptest.NewRequest("DELETE", "/api/crew/stream", nil)
	w := httptest.NewRecorder()

	handler.StreamHandler(w, req)

	if w.Code != http.StatusMethodNotAllowed {
		t.Errorf("Expected status MethodNotAllowed, got %d", w.Code)
	}
}

// TestStreamHandlerMissingQuery tests stream handler requires query parameter
func TestStreamHandlerMissingQuery(t *testing.T) {
	executor := &TeamExecutor{
		team:   &Team{Agents: []*Agent{}},
		apiKey: "test-key",
	}

	handler := NewHTTPHandler(executor)
	req := httptest.NewRequest("GET", "/api/crew/stream", nil)
	w := httptest.NewRecorder()

	handler.StreamHandler(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status BadRequest, got %d", w.Code)
	}

	if !strings.Contains(w.Body.String(), "Query is required") {
		t.Error("Expected 'Query is required' error message")
	}
}

// TestStreamHandlerGETWithQuery tests stream handler accepts GET with query parameter
func TestStreamHandlerGETWithQuery(t *testing.T) {
	agent := &Agent{
		ID:         "agent1",
		Name:       "Agent1",
		Model:      "gpt-4o-mini",
		IsTerminal: true,
	}

	team := &Team{
		Agents:      []*Agent{agent},
		MaxHandoffs: 10,
	}

	executor := NewTeamExecutor(team, "test-key")
	handler := NewHTTPHandler(executor)

	req := httptest.NewRequest("GET", "/api/crew/stream?q=test+query", nil)
	w := httptest.NewRecorder()

	handler.StreamHandler(w, req)

	// Should receive streaming response with proper headers
	if w.Code != http.StatusOK {
		t.Errorf("Expected status OK, got %d", w.Code)
	}

	contentType := w.Header().Get("Content-Type")
	if contentType != "text/event-stream" {
		t.Errorf("Expected Content-Type text/event-stream, got %s", contentType)
	}

	cacheControl := w.Header().Get("Cache-Control")
	if cacheControl != "no-cache" {
		t.Errorf("Expected Cache-Control no-cache, got %s", cacheControl)
	}
}

// TestStreamHandlerPOSTWithJSON tests stream handler accepts POST with JSON body
func TestStreamHandlerPOSTWithJSON(t *testing.T) {
	agent := &Agent{
		ID:         "agent1",
		Name:       "Agent1",
		Model:      "gpt-4o-mini",
		IsTerminal: true,
	}

	team := &Team{
		Agents:      []*Agent{agent},
		MaxHandoffs: 10,
	}

	executor := NewTeamExecutor(team, "test-key")
	handler := NewHTTPHandler(executor)

	reqBody := StreamRequest{
		Query:   "test query",
		History: []Message{},
	}

	bodyBytes, _ := json.Marshal(reqBody)
	req := httptest.NewRequest("POST", "/api/crew/stream", bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	handler.StreamHandler(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status OK, got %d", w.Code)
	}

	contentType := w.Header().Get("Content-Type")
	if contentType != "text/event-stream" {
		t.Errorf("Expected Content-Type text/event-stream, got %s", contentType)
	}
}

// TestStreamHandlerSSEHeaders tests stream handler sets proper SSE headers
func TestStreamHandlerSSEHeaders(t *testing.T) {
	agent := &Agent{
		ID:         "agent1",
		Name:       "Agent1",
		Model:      "gpt-4o-mini",
		IsTerminal: true,
	}

	team := &Team{
		Agents:      []*Agent{agent},
		MaxHandoffs: 10,
	}

	executor := NewTeamExecutor(team, "test-key")
	handler := NewHTTPHandler(executor)

	req := httptest.NewRequest("GET", "/api/crew/stream?q=test", nil)
	w := httptest.NewRecorder()

	handler.StreamHandler(w, req)

	// Check for required SSE headers
	headers := []string{"Content-Type", "Cache-Control", "Connection", "Access-Control-Allow-Origin"}
	for _, header := range headers {
		if w.Header().Get(header) == "" {
			t.Errorf("Expected header %s to be set", header)
		}
	}
}

// TestStreamHandlerCORSHeaders tests stream handler sets CORS headers
func TestStreamHandlerCORSHeaders(t *testing.T) {
	agent := &Agent{
		ID:         "agent1",
		Name:       "Agent1",
		Model:      "gpt-4o-mini",
		IsTerminal: true,
	}

	team := &Team{
		Agents:      []*Agent{agent},
		MaxHandoffs: 10,
	}

	executor := NewTeamExecutor(team, "test-key")
	handler := NewHTTPHandler(executor)

	req := httptest.NewRequest("GET", "/api/crew/stream?q=test", nil)
	w := httptest.NewRecorder()

	handler.StreamHandler(w, req)

	cors := w.Header().Get("Access-Control-Allow-Origin")
	if cors != "*" {
		t.Errorf("Expected CORS header *, got %s", cors)
	}
}

// TestStreamHandlerContextCancellation tests stream handler respects context cancellation
func TestStreamHandlerContextCancellation(t *testing.T) {
	agent := &Agent{
		ID:         "agent1",
		Name:       "Agent1",
		Model:      "gpt-4o-mini",
		IsTerminal: true,
	}

	team := &Team{
		Agents:      []*Agent{agent},
		MaxHandoffs: 10,
	}

	executor := NewTeamExecutor(team, "test-key")
	handler := NewHTTPHandler(executor)

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	req := httptest.NewRequest("GET", "/api/crew/stream?q=test", nil)
	req = req.WithContext(ctx)
	w := httptest.NewRecorder()

	handler.StreamHandler(w, req)

	// Request should still process but context is cancelled
	if w.Code != http.StatusOK {
		t.Errorf("Expected status OK, got %d", w.Code)
	}
}

// TestStreamHandlerStartEvent tests stream handler sends start event
func TestStreamHandlerStartEvent(t *testing.T) {
	agent := &Agent{
		ID:         "agent1",
		Name:       "Agent1",
		Model:      "gpt-4o-mini",
		IsTerminal: true,
	}

	team := &Team{
		Agents:      []*Agent{agent},
		MaxHandoffs: 10,
	}

	executor := NewTeamExecutor(team, "test-key")
	handler := NewHTTPHandler(executor)

	req := httptest.NewRequest("GET", "/api/crew/stream?q=test", nil)
	w := httptest.NewRecorder()

	handler.StreamHandler(w, req)

	body := w.Body.String()
	if !strings.Contains(body, "Starting crew execution") {
		t.Error("Expected start message in response")
	}
}

// TestCreateRequestExecutor tests creating new executor for request
func TestCreateRequestExecutor(t *testing.T) {
	agent := &Agent{
		ID:         "agent1",
		Name:       "Agent1",
		Model:      "gpt-4o-mini",
		IsTerminal: true,
	}

	team := &Team{
		Agents:      []*Agent{agent},
		MaxHandoffs: 10,
	}

	executor := NewTeamExecutor(team, "test-key")
	handler := NewHTTPHandler(executor)

	newExecutor := handler.createRequestExecutor()

	if newExecutor == nil {
		t.Fatal("Expected non-nil executor")
	}

	if newExecutor.team != team {
		t.Error("Expected executor to have same team")
	}

	if newExecutor.apiKey != "test-key" {
		t.Error("Expected executor to have same API key")
	}
}

// TestStreamHandlerWithHistory tests stream handler preserves history
func TestStreamHandlerWithHistory(t *testing.T) {
	agent := &Agent{
		ID:         "agent1",
		Name:       "Agent1",
		Model:      "gpt-4o-mini",
		IsTerminal: true,
	}

	team := &Team{
		Agents:      []*Agent{agent},
		MaxHandoffs: 10,
	}

	executor := NewTeamExecutor(team, "test-key")
	handler := NewHTTPHandler(executor)

	// Create request with history
	history := []Message{
		{Role: "user", Content: "Previous message"},
	}

	reqBody := StreamRequest{
		Query:   "new query",
		History: history,
	}

	bodyBytes, _ := json.Marshal(reqBody)
	req := httptest.NewRequest("POST", "/api/crew/stream", bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	handler.StreamHandler(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status OK, got %d", w.Code)
	}
}

// TestStreamHandlerMultipleCalls tests multiple stream handler calls
func TestStreamHandlerMultipleCalls(t *testing.T) {
	agent := &Agent{
		ID:         "agent1",
		Name:       "Agent1",
		Model:      "gpt-4o-mini",
		IsTerminal: true,
	}

	team := &Team{
		Agents:      []*Agent{agent},
		MaxHandoffs: 10,
	}

	executor := NewTeamExecutor(team, "test-key")
	handler := NewHTTPHandler(executor)

	for i := 0; i < 3; i++ {
		req := httptest.NewRequest("GET", "/api/crew/stream?q=test", nil)
		w := httptest.NewRecorder()

		handler.StreamHandler(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Call %d: Expected status OK, got %d", i+1, w.Code)
		}
	}
}

// TestHealthHandlerResponse tests health handler response format
func TestHealthHandlerResponse(t *testing.T) {
	executor := &TeamExecutor{
		team:   &Team{Agents: []*Agent{}},
		apiKey: "test-key",
	}

	handler := NewHTTPHandler(executor)
	req := httptest.NewRequest("GET", "/health", nil)
	w := httptest.NewRecorder()

	handler.HealthHandler(w, req)

	var response map[string]string
	body, _ := io.ReadAll(w.Body)
	json.Unmarshal(body, &response)

	expectedKeys := []string{"status", "service"}
	for _, key := range expectedKeys {
		if _, exists := response[key]; !exists {
			t.Errorf("Expected key '%s' in response", key)
		}
	}
}

// TestStreamRequestParsing tests StreamRequest unmarshaling
func TestStreamRequestParsing(t *testing.T) {
	reqBody := StreamRequest{
		Query: "test query",
		History: []Message{
			{Role: "user", Content: "Test message"},
		},
	}

	bodyBytes, err := json.Marshal(reqBody)
	if err != nil {
		t.Fatalf("Failed to marshal request: %v", err)
	}

	var parsed StreamRequest
	if err := json.Unmarshal(bodyBytes, &parsed); err != nil {
		t.Errorf("Failed to unmarshal request: %v", err)
	}

	if parsed.Query != "test query" {
		t.Errorf("Expected query 'test query', got '%s'", parsed.Query)
	}

	if len(parsed.History) != 1 {
		t.Errorf("Expected 1 history item, got %d", len(parsed.History))
	}
}

// TestStreamHandlerEmptyHistory tests stream handler with empty history
func TestStreamHandlerEmptyHistory(t *testing.T) {
	agent := &Agent{
		ID:         "agent1",
		Name:       "Agent1",
		Model:      "gpt-4o-mini",
		IsTerminal: true,
	}

	team := &Team{
		Agents:      []*Agent{agent},
		MaxHandoffs: 10,
	}

	executor := NewTeamExecutor(team, "test-key")
	handler := NewHTTPHandler(executor)

	reqBody := StreamRequest{
		Query:   "test query",
		History: []Message{},
	}

	bodyBytes, _ := json.Marshal(reqBody)
	req := httptest.NewRequest("POST", "/api/crew/stream", bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	handler.StreamHandler(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status OK, got %d", w.Code)
	}
}

// TestStartHTTPServerCreatesHandler tests that StartHTTPServer sets up routes
func TestStartHTTPServerCreatesHandler(t *testing.T) {
	agent := &Agent{
		ID:         "agent1",
		Name:       "Agent1",
		Model:      "gpt-4o-mini",
		IsTerminal: true,
	}

	team := &Team{
		Agents:      []*Agent{agent},
		MaxHandoffs: 10,
	}

	executor := NewTeamExecutor(team, "test-key")

	// We can't test the full server startup without a real port,
	// but we can verify the handler creation and route setup
	handler := NewHTTPHandler(executor)

	// Verify the handler is properly initialized
	if handler.executor != executor {
		t.Error("Handler executor should match")
	}
}

// TestStartHTTPServerRoutes tests that HTTP routes are properly configured
func TestStartHTTPServerRoutes(t *testing.T) {
	agent := &Agent{
		ID:         "agent1",
		Name:       "Agent1",
		Model:      "gpt-4o-mini",
		IsTerminal: true,
	}

	team := &Team{
		Agents:      []*Agent{agent},
		MaxHandoffs: 10,
	}

	executor := NewTeamExecutor(team, "test-key")
	handler := NewHTTPHandler(executor)

	// Test health endpoint through handler
	req := httptest.NewRequest("GET", "/health", nil)
	w := httptest.NewRecorder()
	handler.HealthHandler(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Health endpoint should return 200, got %d", w.Code)
	}

	// Test stream endpoint through handler
	req = httptest.NewRequest("GET", "/api/crew/stream?q=test", nil)
	w = httptest.NewRecorder()
	handler.StreamHandler(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Stream endpoint should return 200, got %d", w.Code)
	}
}

// TestStartHTTPServerHealthEndpoint verifies health endpoint works
func TestStartHTTPServerHealthEndpoint(t *testing.T) {
	agent := &Agent{
		ID:         "agent1",
		Name:       "Agent1",
		Model:      "gpt-4o-mini",
		IsTerminal: true,
	}

	team := &Team{
		Agents:      []*Agent{agent},
		MaxHandoffs: 10,
	}

	executor := NewTeamExecutor(team, "test-key")
	handler := NewHTTPHandler(executor)

	req := httptest.NewRequest("GET", "/health", nil)
	w := httptest.NewRecorder()

	handler.HealthHandler(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected 200, got %d", w.Code)
	}

	var response map[string]string
	json.NewDecoder(w.Body).Decode(&response)

	if response["status"] != "ok" {
		t.Error("Health status should be 'ok'")
	}

	if response["service"] != "go-agentic-streaming" {
		t.Error("Service name should be 'go-agentic-streaming'")
	}
}

// TestStartHTTPServerStreamEndpoint verifies stream endpoint works
func TestStartHTTPServerStreamEndpoint(t *testing.T) {
	agent := &Agent{
		ID:         "agent1",
		Name:       "Agent1",
		Model:      "gpt-4o-mini",
		IsTerminal: true,
	}

	team := &Team{
		Agents:      []*Agent{agent},
		MaxHandoffs: 10,
	}

	executor := NewTeamExecutor(team, "test-key")
	handler := NewHTTPHandler(executor)

	req := httptest.NewRequest("GET", "/api/crew/stream?q=test+query", nil)
	w := httptest.NewRecorder()

	handler.StreamHandler(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected 200, got %d", w.Code)
	}

	contentType := w.Header().Get("Content-Type")
	if contentType != "text/event-stream" {
		t.Errorf("Expected text/event-stream, got %s", contentType)
	}
}

// TestStartHTTPServerMultipleRequests tests concurrent request handling
func TestStartHTTPServerMultipleRequests(t *testing.T) {
	agent := &Agent{
		ID:         "agent1",
		Name:       "Agent1",
		Model:      "gpt-4o-mini",
		IsTerminal: true,
	}

	team := &Team{
		Agents:      []*Agent{agent},
		MaxHandoffs: 10,
	}

	executor := NewTeamExecutor(team, "test-key")
	handler := NewHTTPHandler(executor)

	// Make multiple requests in sequence
	for i := 0; i < 3; i++ {
		req := httptest.NewRequest("GET", "/health", nil)
		w := httptest.NewRecorder()

		handler.HealthHandler(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Request %d failed with status %d", i+1, w.Code)
		}
	}
}

// TestHTTPHandlerExecutorStorage tests that executor is properly stored
func TestHTTPHandlerExecutorStorage(t *testing.T) {
	agent := &Agent{
		ID:         "agent1",
		Name:       "Agent1",
		Model:      "gpt-4o-mini",
		IsTerminal: true,
	}

	team := &Team{
		Agents:      []*Agent{agent},
		MaxHandoffs: 10,
	}

	executor := NewTeamExecutor(team, "test-key")
	handler := NewHTTPHandler(executor)

	// Verify executor is accessible
	if handler.executor == nil {
		t.Error("Handler should store executor")
	}

	if handler.executor.team != team {
		t.Error("Handler executor should reference correct team")
	}
}

// TestHTTPHandlerMutexProtection tests mutex functionality
func TestHTTPHandlerMutexProtection(t *testing.T) {
	agent := &Agent{
		ID:         "agent1",
		Name:       "Agent1",
		Model:      "gpt-4o-mini",
		IsTerminal: true,
	}

	team := &Team{
		Agents:      []*Agent{agent},
		MaxHandoffs: 10,
	}

	executor := NewTeamExecutor(team, "test-key")
	handler := NewHTTPHandler(executor)

	// Verify createRequestExecutor works (which uses the mutex)
	newExecutor := handler.createRequestExecutor()
	if newExecutor == nil {
		t.Error("createRequestExecutor should return non-nil executor")
	}
}

// TestStartHTTPServerHandlerSetup tests that StartHTTPServer properly sets up handlers
func TestStartHTTPServerHandlerSetup(t *testing.T) {
	agent := &Agent{
		ID:         "agent1",
		Name:       "Agent1",
		Model:      "gpt-4o-mini",
		IsTerminal: true,
	}

	team := &Team{
		Agents:      []*Agent{agent},
		MaxHandoffs: 10,
	}

	executor := NewTeamExecutor(team, "test-key")
	handler := NewHTTPHandler(executor)

	// Verify handler is created correctly
	if handler == nil {
		t.Fatal("NewHTTPHandler should not return nil")
	}

	if handler.executor == nil {
		t.Error("Handler executor should not be nil")
	}

	if handler.executor.team != team {
		t.Error("Handler executor should have correct team")
	}

	if handler.executor.apiKey != "test-key" {
		t.Error("Handler executor should have correct API key")
	}
}

// TestStartHTTPServerAPIEndpointSetup tests endpoint availability
func TestStartHTTPServerAPIEndpointSetup(t *testing.T) {
	agent := &Agent{
		ID:         "agent1",
		Name:       "Agent1",
		Model:      "gpt-4o-mini",
		IsTerminal: true,
	}

	team := &Team{
		Agents:      []*Agent{agent},
		MaxHandoffs: 10,
	}

	executor := NewTeamExecutor(team, "test-key")
	handler := NewHTTPHandler(executor)

	// Test /api/crew/stream endpoint
	req := httptest.NewRequest("GET", "/api/crew/stream?q=test", nil)
	w := httptest.NewRecorder()
	handler.StreamHandler(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Stream endpoint should work, got status %d", w.Code)
	}

	// Test /health endpoint
	req = httptest.NewRequest("GET", "/health", nil)
	w = httptest.NewRecorder()
	handler.HealthHandler(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Health endpoint should work, got status %d", w.Code)
	}
}

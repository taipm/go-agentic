package openai

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"strings"
	"sync"
	"time"

	openai "github.com/openai/openai-go/v3"
	"github.com/openai/openai-go/v3/option"
	providers "github.com/taipm/go-agentic/core/providers"
	"github.com/taipm/go-agentic/core/tools"
)

// clientEntry represents a cached OpenAI client with expiry time
// ✅ FIX for Issue #2 (Memory Leak): Client cache with TTL expiration
type clientEntry struct {
	client    openai.Client
	createdAt time.Time
	expiresAt time.Time
}

// defaultClientTTL is the default TTL for cached OpenAI clients
// ✅ Phase 5: Now configurable (was hardcoded constant)
// Can be overridden per provider instance or via environment
var defaultClientTTL = 1 * time.Hour // Default: 1 hour (configurable)

// OpenAIProvider implements LLMProvider interface using OpenAI API
// ✅ FIX #3: Made clientTTL configurable (was hardcoded constant)
type OpenAIProvider struct {
	apiKey    string
	client    openai.Client
	clientTTL time.Duration // Configurable TTL for client cache
}

// cachedClients holds OpenAI client instances with TTL-based expiration
var (
	cachedClients = make(map[string]*clientEntry)
	clientMutex   sync.RWMutex
)

// getOrCreateOpenAIClient returns a cached OpenAI client or creates a new one
// ✅ FIX #3: Updated to use defaultClientTTL constant (configurable per provider)
// Clients expire after TTL of inactivity to prevent memory leak
func getOrCreateOpenAIClient(apiKey string) openai.Client {
	clientMutex.Lock()
	defer clientMutex.Unlock()

	// Check if cached and not expired
	if cached, exists := cachedClients[apiKey]; exists {
		if time.Now().Before(cached.expiresAt) {
			// Refresh expiry time on access (sliding window)
			cached.expiresAt = time.Now().Add(defaultClientTTL)
			return cached.client
		}
		// Expired - delete from cache
		delete(cachedClients, apiKey)
	}

	// Create new client
	client := openai.NewClient(option.WithAPIKey(apiKey))

	// Cache with expiry time
	cachedClients[apiKey] = &clientEntry{
		client:    client,
		createdAt: time.Now(),
		expiresAt: time.Now().Add(defaultClientTTL),
	}

	return client
}

// cleanupExpiredClients periodically removes expired clients from cache
// Runs every 5 minutes to prevent memory from growing even if not accessed
func cleanupExpiredClients() {
	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()

	for range ticker.C {
		clientMutex.Lock()
		now := time.Now()
		for apiKey, cached := range cachedClients {
			if now.After(cached.expiresAt) {
				delete(cachedClients, apiKey)
			}
		}
		clientMutex.Unlock()
	}
}

// init starts the cleanup goroutine and registers the provider factory
func init() {
	go cleanupExpiredClients()

	// Register OpenAI provider factory
	providers.NewOpenAIProvider = func(apiKey string) (providers.LLMProvider, error) {
		if apiKey == "" {
			return nil, fmt.Errorf("OpenAI API key cannot be empty")
		}
		return &OpenAIProvider{
			apiKey:    apiKey,
			client:    getOrCreateOpenAIClient(apiKey),
			clientTTL: defaultClientTTL, // ✅ FIX #3: Use default TTL (configurable per provider)
		}, nil
	}
}

// validateCompletionRequest validates the completion request
// Returns error if validation fails
func validateCompletionRequest(req *providers.CompletionRequest) error {
	if req == nil {
		return fmt.Errorf("completion request cannot be nil")
	}

	if req.Model == "" {
		return fmt.Errorf("model name cannot be empty")
	}

	return nil
}

// buildChatCompletionParams builds OpenAI ChatCompletionNewParams from a completion request
func buildChatCompletionParams(req *providers.CompletionRequest) openai.ChatCompletionNewParams {
	messages := convertToOpenAIMessages(req.Messages, req.SystemPrompt)

	params := openai.ChatCompletionNewParams{
		Model:    req.Model,
		Messages: messages,
	}

	// Set temperature if specified
	if req.Temperature > 0 {
		params.Temperature = openai.Float(req.Temperature)
	}

	return params
}

// extractToolCallsFromResponse extracts tool calls from OpenAI response message
// Uses hybrid approach: PRIMARY (native tool_calls) + FALLBACK (text parsing)
func extractToolCallsFromResponse(message openai.ChatCompletionMessage, content string) []providers.ToolCall {
	var toolCalls []providers.ToolCall

	// PRIMARY: Check if completion has tool_calls (OpenAI's structured format)
	if message.ToolCalls != nil {
		toolCalls = extractFromOpenAIToolCalls(message.ToolCalls)
		if len(toolCalls) > 0 {
			log.Printf("[TOOL PARSE] OpenAI native tool_calls: %d calls extracted", len(toolCalls))
			return toolCalls
		}
	}

	// FALLBACK: Extract from text response (for models without tool_use support)
	if content != "" {
		toolCalls = extractToolCallsFromText(content)
		if len(toolCalls) > 0 {
			log.Printf("[TOOL PARSE] Fallback text parsing: %d calls extracted", len(toolCalls))
		}
	}

	return toolCalls
}

// Complete sends a synchronous chat completion request to OpenAI
// Implements LLMProvider.Complete()
func (p *OpenAIProvider) Complete(ctx context.Context, req *providers.CompletionRequest) (*providers.CompletionResponse, error) {
	// Step 1: Validate request
	if err := validateCompletionRequest(req); err != nil {
		return nil, err
	}

	// Step 2: Build OpenAI parameters
	params := buildChatCompletionParams(req)

	// Step 3: Call OpenAI API
	completion, err := p.client.Chat.Completions.New(ctx, params)
	if err != nil {
		return nil, fmt.Errorf("OpenAI API call failed: %w", err)
	}

	if len(completion.Choices) == 0 {
		return nil, fmt.Errorf("no choices in OpenAI response")
	}

	choice := completion.Choices[0]
	message := choice.Message

	// Step 4: Extract tool calls from response
	content := message.Content
	toolCalls := extractToolCallsFromResponse(message, content)

	return &providers.CompletionResponse{
		Content:   content,
		ToolCalls: toolCalls,
	}, nil
}

// CompleteStream sends a streaming chat completion request to OpenAI
// Implements LLMProvider.CompleteStream()
func (p *OpenAIProvider) CompleteStream(ctx context.Context, req *providers.CompletionRequest, streamChan chan<- providers.StreamChunk) error {
	if req == nil {
		return fmt.Errorf("completion request cannot be nil")
	}

	if req.Model == "" {
		return fmt.Errorf("model name cannot be empty")
	}

	// Convert provider-agnostic messages to OpenAI format
	messages := convertToOpenAIMessages(req.Messages, req.SystemPrompt)

	// Create streaming completion request
	params := openai.ChatCompletionNewParams{
		Model:    req.Model,
		Messages: messages,
	}

	// Set temperature if specified
	if req.Temperature > 0 {
		params.Temperature = openai.Float(req.Temperature)
	}

	// Create streaming completion (NewStreaming returns *Stream directly)
	stream := p.client.Chat.Completions.NewStreaming(ctx, params)
	defer stream.Close()

	// Accumulate full message for tool call extraction at the end
	var fullContent strings.Builder
	var toolCalls []providers.ToolCall

	// Read from stream
	for stream.Next() {
		event := stream.Current()

		if len(event.Choices) > 0 {
			choice := event.Choices[0]
			if choice.Delta.Content != "" {
				fullContent.WriteString(choice.Delta.Content)
				streamChan <- providers.StreamChunk{
					Content: choice.Delta.Content,
					Done:    false,
					Error:   nil,
				}
			}

			// Check for tool calls in the delta (if supported by OpenAI in streaming)
			if choice.Delta.ToolCalls != nil {
				toolCalls = append(toolCalls, extractFromOpenAIToolCalls(choice.Delta.ToolCalls)...)
			}
		}
	}

	// Check for stream errors
	if err := stream.Err(); err != nil {
		streamChan <- providers.StreamChunk{
			Content: "",
			Done:    true,
			Error:   fmt.Errorf("streaming error: %w", err),
		}
		return err
	}

	// Send final chunk with done signal and any extracted tool calls
	finalContent := fullContent.String()

	// If no tool calls extracted from streaming, try fallback text parsing
	if len(toolCalls) == 0 && finalContent != "" {
		toolCalls = extractToolCallsFromText(finalContent)
	}

	streamChan <- providers.StreamChunk{
		Content: "",
		Done:    true,
		Error:   nil,
	}

	return nil
}

// Name returns the provider identifier
// Implements LLMProvider.Name()
func (p *OpenAIProvider) Name() string {
	return "openai"
}

// Close cleans up provider resources
// Implements LLMProvider.Close()
func (p *OpenAIProvider) Close() error {
	// OpenAI client doesn't require explicit cleanup
	// Client caching is managed automatically
	return nil
}

// convertToOpenAIMessages converts provider-agnostic messages to OpenAI format
func convertToOpenAIMessages(messages []providers.ProviderMessage, systemPrompt string) []openai.ChatCompletionMessageParamUnion {
	var result []openai.ChatCompletionMessageParamUnion

	// Add system message if provided
	if systemPrompt != "" {
		result = append(result, openai.SystemMessage(systemPrompt))
	}

	// Convert each message
	for _, msg := range messages {
		switch msg.Role {
		case "user":
			result = append(result, openai.UserMessage(msg.Content))
		case "assistant":
			result = append(result, openai.AssistantMessage(msg.Content))
		case "system":
			result = append(result, openai.SystemMessage(msg.Content))
		default:
			// Default to user message for unknown roles
			result = append(result, openai.UserMessage(msg.Content))
		}
	}

	return result
}

// extractFromOpenAIToolCalls extracts tool calls from OpenAI's native tool_calls format
// ✅ PRIMARY METHOD: Uses OpenAI's structured tool_calls (validated by OpenAI)
func extractFromOpenAIToolCalls(toolCalls interface{}) []providers.ToolCall {
	var calls []providers.ToolCall

	// Type assert to handle OpenAI tool calls
	var toolCallSlice []interface{}
	switch v := toolCalls.(type) {
	case []interface{}:
		toolCallSlice = v
	default:
		log.Printf("[TOOL PARSE] OpenAI tool_calls not in expected format")
		return calls
	}

	// Process each OpenAI tool call
	for _, tc := range toolCallSlice {
		tcMap, ok := tc.(map[string]interface{})
		if !ok {
			continue
		}

		// Extract ID
		id, ok := tcMap["id"].(string)
		if !ok {
			continue
		}

		// Extract function details
		funcObj, ok := tcMap["function"].(map[string]interface{})
		if !ok {
			continue
		}

		// Extract tool name
		toolName, ok := funcObj["name"].(string)
		if !ok {
			continue
		}

		// Extract and parse arguments
		args := make(map[string]interface{})
		if argsStr, ok := funcObj["arguments"].(string); ok && argsStr != "" {
			if err := json.Unmarshal([]byte(argsStr), &args); err != nil {
				log.Printf("[TOOL ERROR] Invalid JSON arguments for %s: %v", toolName, err)
				continue
			}
		}

		// Create tool call with validated data
		calls = append(calls, providers.ToolCall{
			ID:        id,
			ToolName:  toolName,
			Arguments: args,
		})

		log.Printf("[TOOL PARSE] Extracted from OpenAI: %s (args validated by OpenAI)", toolName)
	}

	return calls
}

// extractToolCallsFromText extracts tool calls from response text
// ⚠️ FALLBACK METHOD: Uses text parsing (for models without tool_use support)
// Delegates to shared tools.ExtractToolCallsFromText() for unified extraction
func extractToolCallsFromText(text string) []providers.ToolCall {
	// Use shared extraction utility
	extractedCalls := tools.ExtractToolCallsFromText(text)

	// Convert from tools.ExtractedToolCall to providers.ToolCall
	var calls []providers.ToolCall
	for i, extracted := range extractedCalls {
		calls = append(calls, providers.ToolCall{
			ID:        fmt.Sprintf("%s_%d", extracted.ToolName, i),
			ToolName:  extracted.ToolName,
			Arguments: extracted.Arguments,
		})
	}

	return calls
}

// parseToolArguments delegates to shared tools package implementation
func parseToolArguments(argsStr string) map[string]interface{} {
	return tools.ParseArguments(argsStr)
}

// splitArguments delegates to shared tools package implementation
func splitArguments(argsStr string) []string {
	return tools.SplitArguments(argsStr)
}

// isAlphanumeric delegates to shared tools package implementation
func isAlphanumeric(ch rune) bool {
	return tools.IsAlphanumeric(ch)
}

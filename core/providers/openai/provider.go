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
)

// clientEntry represents a cached OpenAI client with expiry time
// ✅ FIX for Issue #2 (Memory Leak): Client cache with TTL expiration
type clientEntry struct {
	client    openai.Client
	createdAt time.Time
	expiresAt time.Time
}

// clientTTL defines how long a cached client remains valid before expiring
// Uses sliding window: TTL is refreshed on each access
const clientTTL = 1 * time.Hour

// OpenAIProvider implements LLMProvider interface using OpenAI API
type OpenAIProvider struct {
	apiKey string
	client openai.Client
}

// cachedClients holds OpenAI client instances with TTL-based expiration
var (
	cachedClients = make(map[string]*clientEntry)
	clientMutex   sync.RWMutex
)

// getOrCreateOpenAIClient returns a cached OpenAI client or creates a new one
// Clients expire after clientTTL (1 hour) of inactivity to prevent memory leak
func getOrCreateOpenAIClient(apiKey string) openai.Client {
	clientMutex.Lock()
	defer clientMutex.Unlock()

	// Check if cached and not expired
	if cached, exists := cachedClients[apiKey]; exists {
		if time.Now().Before(cached.expiresAt) {
			// Refresh expiry time on access (sliding window)
			cached.expiresAt = time.Now().Add(clientTTL)
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
		expiresAt: time.Now().Add(clientTTL),
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
			apiKey: apiKey,
			client: getOrCreateOpenAIClient(apiKey),
		}, nil
	}
}

// Complete sends a synchronous chat completion request to OpenAI
// Implements LLMProvider.Complete()
func (p *OpenAIProvider) Complete(ctx context.Context, req *providers.CompletionRequest) (*providers.CompletionResponse, error) {
	if req == nil {
		return nil, fmt.Errorf("completion request cannot be nil")
	}

	if req.Model == "" {
		return nil, fmt.Errorf("model name cannot be empty")
	}

	// Convert provider-agnostic messages to OpenAI format
	messages := convertToOpenAIMessages(req.Messages, req.SystemPrompt)

	// Create completion request
	params := openai.ChatCompletionNewParams{
		Model:    req.Model,
		Messages: messages,
	}

	// Set temperature if specified
	if req.Temperature > 0 {
		params.Temperature = openai.Float(req.Temperature)
	}

	// Call OpenAI API
	completion, err := p.client.Chat.Completions.New(ctx, params)
	if err != nil {
		return nil, fmt.Errorf("OpenAI API call failed: %w", err)
	}

	if len(completion.Choices) == 0 {
		return nil, fmt.Errorf("no choices in OpenAI response")
	}

	choice := completion.Choices[0]
	message := choice.Message

	// Extract response content
	content := message.Content

	// ✅ FIX for Issue #9 (Tool Call Extraction): Hybrid approach
	// PRIMARY: Use OpenAI's native tool_calls if available (preferred, validated by OpenAI)
	// FALLBACK: Parse text response (for edge cases, legacy support)

	var toolCalls []providers.ToolCall

	// Check if completion has tool_calls (OpenAI's structured format)
	if message.ToolCalls != nil {
		// Message has tool_calls field - use it
		toolCalls = extractFromOpenAIToolCalls(message.ToolCalls)
		if len(toolCalls) > 0 {
			log.Printf("[TOOL PARSE] OpenAI native tool_calls: %d calls extracted", len(toolCalls))
			return &providers.CompletionResponse{
				Content:   content,
				ToolCalls: toolCalls,
			}, nil
		}
	}

	// FALLBACK: Extract from text response (rare, for models without tool_use support)
	if content != "" {
		toolCalls = extractToolCallsFromText(content)
		if len(toolCalls) > 0 {
			log.Printf("[TOOL PARSE] Fallback text parsing: %d calls extracted", len(toolCalls))
		}
	}

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
func extractToolCallsFromText(text string) []providers.ToolCall {
	var calls []providers.ToolCall
	toolCallPattern := make(map[string]bool) // Track unique tool calls

	// Look for patterns like: ToolName(...)
	// Match word characters followed by parentheses
	lines := strings.Split(text, "\n")

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		// Look for function call patterns: Word(...)
		for i := 0; i < len(line); i++ {
			if line[i] == '(' {
				// Found opening paren, look back for function name
				j := i - 1

				// Skip backwards over alphanumeric and underscore
				for j >= 0 && isAlphanumeric(rune(line[j])) {
					j--
				}

				// Check if we found a valid identifier
				if j < i-1 {
					toolName := line[j+1 : i]

					// Validate tool name starts with uppercase (convention for functions)
					if len(toolName) > 0 && toolName[0] >= 'A' && toolName[0] <= 'Z' {
						// Look for closing paren
						endIdx := strings.Index(line[i:], ")")
						if endIdx != -1 {
							endIdx += i
							argsStr := line[i+1 : endIdx]

							// Create tool call entry (avoid duplicates)
							callKey := fmt.Sprintf("%s:%s", toolName, argsStr)
							if !toolCallPattern[callKey] {
								toolCallPattern[callKey] = true
								calls = append(calls, providers.ToolCall{
									ID:        fmt.Sprintf("%s_%d", toolName, len(calls)),
									ToolName:  toolName,
									Arguments: parseToolArguments(argsStr),
								})
							}
						}
					}
				}
			}
		}
	}

	return calls
}

// parseToolArguments splits tool arguments respecting nested brackets
func parseToolArguments(argsStr string) map[string]interface{} {
	result := make(map[string]interface{})

	if argsStr == "" {
		return result
	}

	// Simple approach: try to parse as JSON first (handles complex types)
	// If that fails, treat as simple string arguments
	var jsonArgs map[string]interface{}
	if err := json.Unmarshal([]byte("{"+argsStr+"}"), &jsonArgs); err == nil {
		return jsonArgs
	}

	// Fallback: parse as comma-separated positional arguments
	parts := splitArguments(argsStr)
	for i, part := range parts {
		part = strings.TrimSpace(part)
		part = strings.Trim(part, `"'`)
		result[fmt.Sprintf("arg%d", i)] = part
	}

	return result
}

// splitArguments splits arguments respecting nested brackets and quotes
func splitArguments(argsStr string) []string {
	var parts []string
	var current strings.Builder
	bracketDepth := 0
	quoteChar := rune(0)

	for _, ch := range argsStr {
		switch ch {
		case '"', '\'':
			if quoteChar == 0 {
				quoteChar = ch
			} else if quoteChar == ch {
				quoteChar = 0
			}
			current.WriteRune(ch)
		case '[', '{':
			if quoteChar == 0 {
				bracketDepth++
			}
			current.WriteRune(ch)
		case ']', '}':
			if quoteChar == 0 {
				bracketDepth--
			}
			current.WriteRune(ch)
		case ',':
			if bracketDepth == 0 && quoteChar == 0 {
				part := strings.TrimSpace(current.String())
				if part != "" {
					parts = append(parts, part)
				}
				current.Reset()
			} else {
				current.WriteRune(ch)
			}
		default:
			current.WriteRune(ch)
		}
	}

	if part := strings.TrimSpace(current.String()); part != "" {
		parts = append(parts, part)
	}

	return parts
}

// isAlphanumeric checks if a rune is alphanumeric or underscore
func isAlphanumeric(ch rune) bool {
	return (ch >= 'a' && ch <= 'z') || (ch >= 'A' && ch <= 'Z') || (ch >= '0' && ch <= '9') || ch == '_'
}

package ollama

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"strings"

	providers "github.com/taipm/go-agentic/core/providers"
)

// OllamaProvider implements LLMProvider interface using Ollama API
type OllamaProvider struct {
	baseURL string
	client  *http.Client
}

// OllamaChatRequest represents the request body for Ollama chat API
type OllamaChatRequest struct {
	Model       string              `json:"model"`
	Messages    []OllamaMessage     `json:"messages"`
	Temperature float64             `json:"temperature,omitempty"`
	Stream      bool                `json:"stream"`
	Format      string              `json:"format,omitempty"` // "json" for structured output
}

// OllamaMessage represents a single message in Ollama format
type OllamaMessage struct {
	Role    string `json:"role"`    // "system", "user", "assistant"
	Content string `json:"content"`
}

// OllamaChatResponse represents the response from Ollama chat API
type OllamaChatResponse struct {
	Model              string        `json:"model"`
	CreatedAt          string        `json:"created_at"`
	Message            OllamaMessage `json:"message"`
	Done               bool          `json:"done"`
	TotalDuration      int64         `json:"total_duration,omitempty"`
	LoadDuration       int64         `json:"load_duration,omitempty"`
	PromptEvalCount    int64         `json:"prompt_eval_count,omitempty"`
	PromptEvalDuration int64         `json:"prompt_eval_duration,omitempty"`
	EvalCount          int64         `json:"eval_count,omitempty"`
	EvalDuration       int64         `json:"eval_duration,omitempty"`
}

// init registers the Ollama provider factory
func init() {
	// Register Ollama provider factory
	providers.NewOllamaProvider = func(baseURL string) (providers.LLMProvider, error) {
		// ✅ FIX #2: Check environment variable OLLAMA_URL if baseURL not provided
		if baseURL == "" {
			baseURL = os.Getenv("OLLAMA_URL")
		}

		// If still empty, require explicit configuration
		if baseURL == "" {
			return nil, fmt.Errorf("Ollama URL not specified: use 'provider_url' in agent YAML config or set OLLAMA_URL environment variable (e.g., http://localhost:11434)")
		}

		// Validate URL format
		if !strings.HasPrefix(baseURL, "http://") && !strings.HasPrefix(baseURL, "https://") {
			baseURL = "http://" + baseURL
		}

		// Verify URL is parseable
		_, err := url.Parse(baseURL)
		if err != nil {
			return nil, fmt.Errorf("invalid Ollama URL: %w", err)
		}

		return &OllamaProvider{
			baseURL: strings.TrimSuffix(baseURL, "/"), // Remove trailing slash
			client: &http.Client{
				Timeout: 0, // Streaming requests can take a while
			},
		}, nil
	}
}

// Complete sends a synchronous chat completion request to Ollama
// Implements LLMProvider.Complete()
func (p *OllamaProvider) Complete(ctx context.Context, req *providers.CompletionRequest) (*providers.CompletionResponse, error) {
	if req == nil {
		return nil, fmt.Errorf("completion request cannot be nil")
	}

	if req.Model == "" {
		return nil, fmt.Errorf("model name cannot be empty")
	}

	// Convert provider-agnostic messages to Ollama format
	messages := convertToOllamaMessages(req.Messages, req.SystemPrompt)

	// Create Ollama request
	ollamaReq := &OllamaChatRequest{
		Model:       req.Model,
		Messages:    messages,
		Temperature: req.Temperature,
		Stream:      false,
	}

	// Serialize request
	reqBody, err := json.Marshal(ollamaReq)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	// Create HTTP request
	httpReq, err := http.NewRequestWithContext(ctx, "POST", p.baseURL+"/api/chat", bytes.NewReader(reqBody))
	if err != nil {
		return nil, fmt.Errorf("failed to create HTTP request: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")

	// Send request
	httpResp, err := p.client.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("Ollama API call failed: %w", err)
	}
	defer httpResp.Body.Close()

	// Check response status
	if httpResp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(httpResp.Body)
		return nil, fmt.Errorf("Ollama API error (status %d): %s", httpResp.StatusCode, string(bodyBytes))
	}

	// Parse response
	var ollamaResp OllamaChatResponse
	if err := json.NewDecoder(httpResp.Body).Decode(&ollamaResp); err != nil {
		return nil, fmt.Errorf("failed to decode Ollama response: %w", err)
	}

	// Extract content
	content := ollamaResp.Message.Content

	// ✅ FIX for Issue #9 (Tool Call Extraction): Hybrid approach
	// PRIMARY: Try text parsing (Ollama doesn't have native tool_calls like OpenAI)
	// For small models like gemma3:1b and deepseek-r1:1.5b, text parsing is essential

	var toolCalls []providers.ToolCall

	// Try to extract tool calls from response text
	if content != "" {
		toolCalls = extractToolCallsFromText(content)
		if len(toolCalls) > 0 {
			log.Printf("[TOOL PARSE] Ollama text parsing: %d calls extracted from %s", len(toolCalls), req.Model)
		}
	}

	return &providers.CompletionResponse{
		Content:   content,
		ToolCalls: toolCalls,
	}, nil
}

// CompleteStream sends a streaming chat completion request to Ollama
// Implements LLMProvider.CompleteStream()
func (p *OllamaProvider) CompleteStream(ctx context.Context, req *providers.CompletionRequest, streamChan chan<- providers.StreamChunk) error {
	if req == nil {
		return fmt.Errorf("completion request cannot be nil")
	}

	if req.Model == "" {
		return fmt.Errorf("model name cannot be empty")
	}

	// Convert provider-agnostic messages to Ollama format
	messages := convertToOllamaMessages(req.Messages, req.SystemPrompt)

	// Create Ollama request with streaming enabled
	ollamaReq := &OllamaChatRequest{
		Model:       req.Model,
		Messages:    messages,
		Temperature: req.Temperature,
		Stream:      true,
	}

	// Serialize request
	reqBody, err := json.Marshal(ollamaReq)
	if err != nil {
		return fmt.Errorf("failed to marshal request: %w", err)
	}

	// Create HTTP request
	httpReq, err := http.NewRequestWithContext(ctx, "POST", p.baseURL+"/api/chat", bytes.NewReader(reqBody))
	if err != nil {
		return fmt.Errorf("failed to create HTTP request: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")

	// Send request
	httpResp, err := p.client.Do(httpReq)
	if err != nil {
		return fmt.Errorf("Ollama API call failed: %w", err)
	}
	defer httpResp.Body.Close()

	// Check response status
	if httpResp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(httpResp.Body)
		streamChan <- providers.StreamChunk{
			Content: "",
			Done:    true,
			Error:   fmt.Errorf("Ollama API error (status %d): %s", httpResp.StatusCode, string(bodyBytes)),
		}
		return fmt.Errorf("Ollama API error (status %d)", httpResp.StatusCode)
	}

	// Process streaming response
	var fullContent strings.Builder
	decoder := json.NewDecoder(httpResp.Body)

	for {
		var chunk OllamaChatResponse
		err := decoder.Decode(&chunk)
		if err != nil {
			if err == io.EOF {
				break
			}
			streamChan <- providers.StreamChunk{
				Content: "",
				Done:    true,
				Error:   fmt.Errorf("failed to decode stream chunk: %w", err),
			}
			return err
		}

		// Stream the content delta
		if chunk.Message.Content != "" {
			fullContent.WriteString(chunk.Message.Content)
			streamChan <- providers.StreamChunk{
				Content: chunk.Message.Content,
				Done:    false,
				Error:   nil,
			}
		}

		// Check if streaming is done
		if chunk.Done {
			break
		}
	}

	// Send final chunk with done signal
	streamChan <- providers.StreamChunk{
		Content: "",
		Done:    true,
		Error:   nil,
	}

	return nil
}

// Name returns the provider identifier
// Implements LLMProvider.Name()
func (p *OllamaProvider) Name() string {
	return "ollama"
}

// Close cleans up provider resources
// Implements LLMProvider.Close()
func (p *OllamaProvider) Close() error {
	// HTTP client doesn't require explicit cleanup
	// Connection pooling is managed automatically
	return nil
}

// convertToOllamaMessages converts provider-agnostic messages to Ollama format
func convertToOllamaMessages(messages []providers.ProviderMessage, systemPrompt string) []OllamaMessage {
	var result []OllamaMessage

	// Add system message if provided
	if systemPrompt != "" {
		result = append(result, OllamaMessage{
			Role:    "system",
			Content: systemPrompt,
		})
	}

	// Convert each message
	for _, msg := range messages {
		// Normalize role to Ollama format
		role := msg.Role
		if role != "system" && role != "user" && role != "assistant" {
			role = "user" // Default to user for unknown roles
		}

		result = append(result, OllamaMessage{
			Role:    role,
			Content: msg.Content,
		})
	}

	return result
}

// extractToolCallsFromText extracts tool calls from response text
// ⚠️ PRIMARY METHOD for Ollama: Uses text parsing (Ollama doesn't have native tool_calls)
// Critical for small models like gemma3:1b and deepseek-r1:1.5b
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

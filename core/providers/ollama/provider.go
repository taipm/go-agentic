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
	"github.com/taipm/go-agentic/core/tools"
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

		// If still empty, use default local URL
		if baseURL == "" {
			baseURL = "http://localhost:11434" // Default Ollama local server
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

	// Extract usage information from response
	var usage *providers.UsageInfo
	if ollamaResp.PromptEvalCount > 0 || ollamaResp.EvalCount > 0 {
		usage = &providers.UsageInfo{
			InputTokens:  int(ollamaResp.PromptEvalCount),
			OutputTokens: int(ollamaResp.EvalCount),
			TotalTokens:  int(ollamaResp.PromptEvalCount + ollamaResp.EvalCount),
		}
	}

	return &providers.CompletionResponse{
		Content:   content,
		ToolCalls: toolCalls,
		Usage:     usage,
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
// Supports JSON, key=value, and positional argument formats with type conversion
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

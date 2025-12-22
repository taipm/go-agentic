package llms

import (
	"context"
	"fmt"
	"sync"
)

// LLMProvider defines the interface for LLM backends
type LLMProvider interface {
	// Complete sends a chat completion request (synchronous)
	Complete(ctx context.Context, req *CompletionRequest) (*CompletionResponse, error)

	// CompleteStream sends a chat completion request with streaming responses
	CompleteStream(ctx context.Context, req *CompletionRequest, streamChan chan<- StreamChunk) error

	// Name returns the provider identifier ("openai", "ollama", etc.)
	Name() string

	// Close cleans up resources (connection pooling, caches, etc.)
	Close() error
}

// StreamChunk represents a chunk of streamed response
type StreamChunk struct {
	Content string // Response text content
	Done    bool   // True if this is the final chunk
	Error   error  // Error if streaming failed
}

// CompletionRequest is provider-agnostic request structure
type CompletionRequest struct {
	Model        string              // Model name ("gpt-4o-mini" or "llama3.1:8b")
	SystemPrompt string              // System message for context
	Messages     []ProviderMessage   // Conversation history
	Temperature  float64             // Sampling temperature (0.0-2.0)
	Tools        []ProviderTool      // Available tools for function calling
}

// ProviderMessage is provider-agnostic message format
type ProviderMessage struct {
	Role    string // "user", "assistant", "system"
	Content string // Message content
}

// ProviderTool is provider-agnostic tool/function definition
type ProviderTool struct {
	Name        string                 // Tool name
	Description string                 // What the tool does
	Parameters  map[string]interface{} // JSON schema for parameters
}

// CompletionResponse is provider-agnostic response
type CompletionResponse struct {
	Content   string          // Response text from model
	ToolCalls []ToolCall      // Extracted tool calls
}

// ToolCall represents a tool invocation (matches crewai.ToolCall)
type ToolCall struct {
	ID        string                 // Unique ID for this call
	ToolName  string                 // Name of tool to invoke
	Arguments map[string]interface{} // Parsed arguments
}

// ProviderFactory creates and caches LLM provider instances
type ProviderFactory struct {
	providers map[string]LLMProvider // Cache: key -> provider instance
	mu        sync.RWMutex           // Thread-safe access
}

// NewProviderFactory creates a new provider factory
func NewProviderFactory() *ProviderFactory {
	return &ProviderFactory{
		providers: make(map[string]LLMProvider),
	}
}

// GetProvider returns a cached or creates a new provider instance
// providerType: "openai" or "ollama" (defaults to "ollama")
// providerURL: provider-specific URL (for Ollama, defaults to "http://localhost:11434")
// apiKey: API key (for OpenAI)
func (f *ProviderFactory) GetProvider(providerType, providerURL, apiKey string) (LLMProvider, error) {
	// Default to Ollama (local development)
	if providerType == "" {
		providerType = "ollama"
	}

	// Build cache key
	cacheKey := f.buildCacheKey(providerType, providerURL, apiKey)

	// Check cache with read lock (fast path)
	f.mu.RLock()
	if provider, exists := f.providers[cacheKey]; exists {
		f.mu.RUnlock()
		return provider, nil
	}
	f.mu.RUnlock()

	// Not in cache, acquire write lock to create new provider
	f.mu.Lock()
	defer f.mu.Unlock()

	// Double-check after acquiring lock (prevent duplicate creation)
	if provider, exists := f.providers[cacheKey]; exists {
		return provider, nil
	}

	// Create new provider based on type
	var provider LLMProvider
	var err error

	switch providerType {
	case "openai":
		// OpenAI provider is created and registered via init()
		provider, err = NewOpenAIProvider(apiKey)
	case "ollama":
		// Ollama provider is created and registered via init()
		if providerURL == "" {
			providerURL = "http://localhost:11434" // Default Ollama URL
		}
		provider, err = NewOllamaProvider(providerURL)
	default:
		return nil, fmt.Errorf("unsupported provider: %s", providerType)
	}

	if err != nil {
		return nil, err
	}

	// Cache the provider instance
	f.providers[cacheKey] = provider
	return provider, nil
}

// buildCacheKey generates a unique cache key for provider + config
func (f *ProviderFactory) buildCacheKey(providerType, providerURL, apiKey string) string {
	switch providerType {
	case "openai":
		// OpenAI: key includes API key (different keys = different instances)
		return fmt.Sprintf("openai:%s", apiKey)
	case "ollama":
		// Ollama: key includes URL (different URLs = different instances)
		if providerURL == "" {
			providerURL = "http://localhost:11434"
		}
		return fmt.Sprintf("ollama:%s", providerURL)
	default:
		return providerType
	}
}

// Close closes all cached provider instances
func (f *ProviderFactory) Close() error {
	f.mu.Lock()
	defer f.mu.Unlock()

	var errs []error
	for _, provider := range f.providers {
		if err := provider.Close(); err != nil {
			errs = append(errs, err)
		}
	}

	if len(errs) > 0 {
		return fmt.Errorf("errors closing providers: %v", errs)
	}
	return nil
}

// Global factory instance for singleton access
var globalFactory *ProviderFactory

func init() {
	globalFactory = NewProviderFactory()
}

// GetGlobalFactory returns the global provider factory instance
func GetGlobalFactory() *ProviderFactory {
	return globalFactory
}

// Helper constructors - implemented by each provider package
// These are set during provider package init()

var (
	// NewOpenAIProvider is set by llms/openai package
	NewOpenAIProvider func(apiKey string) (LLMProvider, error)

	// NewOllamaProvider is set by llms/ollama package
	NewOllamaProvider func(baseURL string) (LLMProvider, error)
)

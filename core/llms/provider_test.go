package llms

import (
	"context"
	"errors"
	"testing"
)

// MockProvider is a test implementation of LLMProvider
type MockProvider struct {
	name string
	err  error
}

func (m *MockProvider) Complete(ctx context.Context, req *CompletionRequest) (*CompletionResponse, error) {
	if m.err != nil {
		return nil, m.err
	}
	return &CompletionResponse{
		Content:   "Mock response from " + m.name,
		ToolCalls: []ToolCall{},
	}, nil
}

func (m *MockProvider) CompleteStream(ctx context.Context, req *CompletionRequest, streamChan chan<- StreamChunk) error {
	if m.err != nil {
		return m.err
	}
	streamChan <- StreamChunk{
		Content: "Mock streamed response",
		Done:    true,
	}
	return nil
}

func (m *MockProvider) Name() string {
	return m.name
}

func (m *MockProvider) Close() error {
	return nil
}

// TestProviderFactory_CreateOpenAIProvider tests OpenAI provider creation
func TestProviderFactory_CreateOpenAIProvider(t *testing.T) {
	factory := NewProviderFactory()

	// Register mock OpenAI provider for testing
	NewOpenAIProvider = func(apiKey string) (LLMProvider, error) {
		return &MockProvider{name: "openai"}, nil
	}

	provider, err := factory.GetProvider("openai", "", "test-key")
	if err != nil {
		t.Fatalf("failed to create OpenAI provider: %v", err)
	}

	if provider.Name() != "openai" {
		t.Errorf("expected provider name 'openai', got %s", provider.Name())
	}
}

// TestProviderFactory_CreateOllamaProvider tests Ollama provider creation
func TestProviderFactory_CreateOllamaProvider(t *testing.T) {
	factory := NewProviderFactory()

	// Register mock Ollama provider for testing
	NewOllamaProvider = func(baseURL string) (LLMProvider, error) {
		return &MockProvider{name: "ollama"}, nil
	}

	provider, err := factory.GetProvider("ollama", "http://localhost:11434", "")
	if err != nil {
		t.Fatalf("failed to create Ollama provider: %v", err)
	}

	if provider.Name() != "ollama" {
		t.Errorf("expected provider name 'ollama', got %s", provider.Name())
	}
}

// TestProviderFactory_DefaultsToOllama tests that empty provider type defaults to Ollama
func TestProviderFactory_DefaultsToOllama(t *testing.T) {
	factory := NewProviderFactory()

	// Register mock Ollama provider
	NewOllamaProvider = func(baseURL string) (LLMProvider, error) {
		return &MockProvider{name: "ollama"}, nil
	}

	// Empty provider type should default to Ollama (local development)
	provider, err := factory.GetProvider("", "", "")
	if err != nil {
		t.Fatalf("failed to create default provider: %v", err)
	}

	if provider.Name() != "ollama" {
		t.Errorf("expected default provider 'ollama', got %s", provider.Name())
	}
}

// TestProviderFactory_CacheReuse tests that providers are cached
func TestProviderFactory_CacheReuse(t *testing.T) {
	factory := NewProviderFactory()

	// Register mock OpenAI provider
	NewOpenAIProvider = func(apiKey string) (LLMProvider, error) {
		return &MockProvider{name: "openai"}, nil
	}

	// Get provider twice with same config
	provider1, err := factory.GetProvider("openai", "", "test-key")
	if err != nil {
		t.Fatalf("failed to create provider: %v", err)
	}

	provider2, err := factory.GetProvider("openai", "", "test-key")
	if err != nil {
		t.Fatalf("failed to get cached provider: %v", err)
	}

	// Should be same instance (cached)
	if provider1 != provider2 {
		t.Error("expected cached provider to be the same instance")
	}
}

// TestProviderFactory_DifferentKeysCreateDifferentInstances tests cache key isolation
func TestProviderFactory_DifferentKeysCreateDifferentInstances(t *testing.T) {
	factory := NewProviderFactory()

	// Register mock OpenAI provider
	NewOpenAIProvider = func(apiKey string) (LLMProvider, error) {
		return &MockProvider{name: "openai:" + apiKey}, nil
	}

	// Get providers with different API keys
	provider1, err := factory.GetProvider("openai", "", "key1")
	if err != nil {
		t.Fatalf("failed to create provider 1: %v", err)
	}

	provider2, err := factory.GetProvider("openai", "", "key2")
	if err != nil {
		t.Fatalf("failed to create provider 2: %v", err)
	}

	// Should be different instances
	if provider1 == provider2 {
		t.Error("expected different instances for different API keys")
	}

	// Names should reflect different keys
	if provider1.Name() != "openai:key1" {
		t.Errorf("expected name 'openai:key1', got %s", provider1.Name())
	}
	if provider2.Name() != "openai:key2" {
		t.Errorf("expected name 'openai:key2', got %s", provider2.Name())
	}
}

// TestProviderFactory_UnsupportedProvider tests error handling
func TestProviderFactory_UnsupportedProvider(t *testing.T) {
	factory := NewProviderFactory()

	_, err := factory.GetProvider("unsupported", "", "")
	if err == nil {
		t.Error("expected error for unsupported provider")
	}

	if !errors.Is(err, errors.New("unsupported provider: unsupported")) && !errors.Is(err, errors.New("")) {
		// Just check error message contains "unsupported"
		if err.Error() != "unsupported provider: unsupported" {
			t.Errorf("unexpected error message: %v", err)
		}
	}
}

// TestProviderFactory_DefaultOllamaURL tests default Ollama URL
func TestProviderFactory_DefaultOllamaURL(t *testing.T) {
	factory := NewProviderFactory()

	// Register mock that captures the URL
	var capturedURL string
	NewOllamaProvider = func(baseURL string) (LLMProvider, error) {
		capturedURL = baseURL
		return &MockProvider{name: "ollama"}, nil
	}

	// Create without specifying URL
	_, err := factory.GetProvider("ollama", "", "")
	if err != nil {
		t.Fatalf("failed to create Ollama provider: %v", err)
	}

	// Should use default URL
	if capturedURL != "http://localhost:11434" {
		t.Errorf("expected default URL 'http://localhost:11434', got %s", capturedURL)
	}
}

// TestBuildCacheKey tests cache key generation
func TestBuildCacheKey(t *testing.T) {
	factory := NewProviderFactory()

	tests := []struct {
		providerType string
		providerURL  string
		apiKey       string
		expectedKey  string
	}{
		{"openai", "", "key123", "openai:key123"},
		{"ollama", "http://localhost:11434", "", "ollama:http://localhost:11434"},
		{"ollama", "", "", "ollama:http://localhost:11434"}, // Default URL
	}

	for _, tt := range tests {
		key := factory.buildCacheKey(tt.providerType, tt.providerURL, tt.apiKey)
		if key != tt.expectedKey {
			t.Errorf("expected key %s, got %s", tt.expectedKey, key)
		}
	}
}

// TestGlobalFactory tests singleton access
func TestGlobalFactory(t *testing.T) {
	factory1 := GetGlobalFactory()
	factory2 := GetGlobalFactory()

	if factory1 != factory2 {
		t.Error("expected global factory to be a singleton")
	}
}

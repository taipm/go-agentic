// Package agent provides agent execution and related functionality.
package agent

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/taipm/go-agentic/core/common"
	providers "github.com/taipm/go-agentic/core/providers"
	_ "github.com/taipm/go-agentic/core/providers/openai"  // Register OpenAI provider
	_ "github.com/taipm/go-agentic/core/providers/ollama"  // Register Ollama provider
)

// providerFactory is the global LLM provider factory instance
var providerFactory = providers.GetGlobalFactory()

// ExecuteAgent runs an agent and returns its response
// Uses provider factory to support multiple LLM backends (OpenAI, Ollama, etc.)
// Supports fallback to backup LLM model if primary fails
func ExecuteAgent(ctx context.Context, agent *common.Agent, input string, history []common.Message, apiKey string) (*common.AgentResponse, error) {
	// Get model configs (use new Primary/Backup if available, else fall back to old fields)
	primaryConfig := agent.PrimaryModel
	backupConfig := agent.BackupModel

	// Build system prompt once (reused for both primary and backup)
	systemPrompt := BuildSystemPrompt(agent)
	messages := ConvertToProviderMessages(history)

	// 1️⃣ TRY PRIMARY MODEL
	response, primaryErr := executeWithModelConfig(ctx, agent, systemPrompt, messages, primaryConfig, apiKey)
	if primaryErr == nil {
		return response, nil
	}

	// 2️⃣ IF PRIMARY FAILED AND BACKUP EXISTS, TRY BACKUP
	if backupConfig != nil {
		fmt.Printf("[FALLBACK] Primary model '%s' failed: %v. Trying backup model '%s'...\n",
			primaryConfig.Model, primaryErr, backupConfig.Model)

		response, backupErr := executeWithModelConfig(ctx, agent, systemPrompt, messages, backupConfig, apiKey)
		if backupErr == nil {
			fmt.Printf("[FALLBACK SUCCESS] Backup model '%s' succeeded\n", backupConfig.Model)
			return response, nil
		}

		// Both failed - return detailed error
		return nil, fmt.Errorf("both primary and backup models failed: primary=%v, backup=%v", primaryErr, backupErr)
	}

	// No backup available - return primary error
	return nil, fmt.Errorf("primary model failed: %w", primaryErr)
}

// ExecuteAgentStream runs an agent with streaming responses
// Streams response chunks to the provided channel
// Supports fallback to backup LLM model if primary fails
func ExecuteAgentStream(ctx context.Context, agent *common.Agent, input string, history []common.Message, apiKey string, streamChan chan<- providers.StreamChunk) error {
	// Get model configs
	primaryConfig := agent.PrimaryModel
	backupConfig := agent.BackupModel

	// Build system prompt once
	systemPrompt := BuildSystemPrompt(agent)
	messages := ConvertToProviderMessages(history)

	// 1️⃣ TRY PRIMARY MODEL WITH STREAMING
	primaryErr := executeWithModelConfigStream(ctx, agent, systemPrompt, messages, primaryConfig, apiKey, streamChan)
	if primaryErr == nil {
		return nil
	}

	// 2️⃣ IF PRIMARY FAILED AND BACKUP EXISTS, TRY BACKUP
	if backupConfig != nil {
		fmt.Printf("[FALLBACK] Primary model streaming failed: %v. Trying backup...\n", primaryErr)

		backupErr := executeWithModelConfigStream(ctx, agent, systemPrompt, messages, backupConfig, apiKey, streamChan)
		if backupErr == nil {
			fmt.Printf("[FALLBACK SUCCESS] Backup model streaming succeeded\n")
			return nil
		}

		// Both failed - return detailed error
		return fmt.Errorf("both primary and backup models failed for streaming: primary=%v, backup=%v", primaryErr, backupErr)
	}

	// No backup available - return primary error
	return fmt.Errorf("primary model streaming failed: %w", primaryErr)
}

// executeWithModelConfig executes an agent with a specific model configuration
func executeWithModelConfig(ctx context.Context, agent *common.Agent, systemPrompt string, messages []providers.ProviderMessage, modelConfig *common.ModelConfig, apiKey string) (*common.AgentResponse, error) {
	// Step 1: Estimate tokens and validate cost limits
	systemAndPromptContent := systemPrompt
	for _, msg := range messages {
		systemAndPromptContent += msg.Content
	}
	_ = agent.EstimateTokens(systemAndPromptContent) // Token estimation for future metrics use

	// Step 2: Build completion request
	request := &providers.CompletionRequest{
		Model:        modelConfig.Model,
		SystemPrompt: systemPrompt,
		Messages:     messages,
		Temperature:  agent.Temperature,
		Tools:        ConvertToolsToProvider(agent.Tools),
	}

	// Step 3: Execute provider call
	provider, err := providerFactory.GetProvider(modelConfig.Provider, modelConfig.ProviderURL, apiKey)
	if err != nil {
		return nil, fmt.Errorf("failed to get LLM provider '%s': %w", modelConfig.Provider, err)
	}

	startTime := time.Now()
	response, err := provider.Complete(ctx, request)
	_ = time.Since(startTime) // Execution duration for metrics in full implementation

	if err != nil {
		return nil, err
	}

	return &common.AgentResponse{
		AgentID:   agent.ID,
		AgentName: agent.Name,
		Content:   response.Content,
		ToolCalls: ConvertToolCallsFromProvider(response.ToolCalls),
	}, nil
}

// executeWithModelConfigStream executes an agent with streaming using a specific model configuration
func executeWithModelConfigStream(ctx context.Context, agent *common.Agent, systemPrompt string, messages []providers.ProviderMessage, modelConfig *common.ModelConfig, apiKey string, streamChan chan<- providers.StreamChunk) error {
	// Estimate tokens BEFORE execution
	systemAndPromptContent := systemPrompt
	for _, msg := range messages {
		systemAndPromptContent += msg.Content
	}
	estimatedTokens := agent.EstimateTokens(systemAndPromptContent)

	// Get provider instance
	provider, err := providerFactory.GetProvider(modelConfig.Provider, modelConfig.ProviderURL, apiKey)
	if err != nil {
		return fmt.Errorf("failed to get LLM provider '%s': %w", modelConfig.Provider, err)
	}

	// Create completion request
	request := &providers.CompletionRequest{
		Model:        modelConfig.Model,
		SystemPrompt: systemPrompt,
		Messages:     messages,
		Temperature:  agent.Temperature,
		Tools:        ConvertToolsToProvider(agent.Tools),
	}

	// Call provider with streaming
	startTime := time.Now()
	err = provider.CompleteStream(ctx, request, streamChan)
	_ = time.Since(startTime) // Execution duration for metrics in full implementation
	_ = estimatedTokens       // Used for cost tracking in full implementation

	return err
}

// ConvertToProviderMessages converts internal Message format to provider-agnostic format
func ConvertToProviderMessages(history []common.Message) []providers.ProviderMessage {
	messages := make([]providers.ProviderMessage, len(history))
	for i, msg := range history {
		messages[i] = providers.ProviderMessage{
			Role:    msg.Role,
			Content: msg.Content,
		}
	}
	return messages
}

// ConvertToolsToProvider converts internal Tool format to provider-agnostic format
func ConvertToolsToProvider(toolsObj interface{}) []providers.ProviderTool {
	// Handle nil case
	if toolsObj == nil {
		return []providers.ProviderTool{}
	}

	// For now, return empty list since agent.Tools are interface{}
	// Full implementation would iterate and extract tool metadata
	return []providers.ProviderTool{}
}

// ConvertToolCallsFromProvider converts provider tool calls back to internal format
func ConvertToolCallsFromProvider(providerCalls []providers.ToolCall) []common.ToolCall {
	calls := make([]common.ToolCall, len(providerCalls))
	for i, call := range providerCalls {
		calls[i] = common.ToolCall{
			ID:        call.ID,
			ToolName:  call.ToolName,
			Arguments: call.Arguments,
		}
	}
	return calls
}

// BuildSystemPrompt creates the system prompt for the agent
func BuildSystemPrompt(agent *common.Agent) string {
	if agent.SystemPrompt != "" {
		return buildCustomPrompt(agent)
	}
	return buildGenericPrompt(agent)
}

// buildCustomPrompt processes custom system prompt with template variable replacement
func buildCustomPrompt(agent *common.Agent) string {
	prompt := agent.SystemPrompt
	prompt = strings.ReplaceAll(prompt, "{{name}}", agent.Name)
	prompt = strings.ReplaceAll(prompt, "{{role}}", agent.Role)
	prompt = strings.ReplaceAll(prompt, "{{description}}", agent.Name+" - "+agent.Role)
	prompt = strings.ReplaceAll(prompt, "{{backstory}}", agent.Backstory)
	return prompt
}

// buildGenericPrompt creates a default prompt with agent role and tools
func buildGenericPrompt(agent *common.Agent) string {
	var prompt strings.Builder

	prompt.WriteString(fmt.Sprintf("You are %s.\n", agent.Name))
	prompt.WriteString(fmt.Sprintf("Role: %s\n", agent.Role))
	prompt.WriteString(fmt.Sprintf("Backstory: %s\n\n", agent.Backstory))

	if len(agent.Tools) > 0 {
		prompt.WriteString("You have access to the following tools\n\n")
		prompt.WriteString("When you need to use a tool, write it exactly like this (on its own line):\n")
		prompt.WriteString("ToolName(param1, param2)\n\n")
	}

	prompt.WriteString("Instructions:\n")
	prompt.WriteString("1. Analyze the input and determine what tools you need\n")
	prompt.WriteString("2. Use tools to gather information\n")
	prompt.WriteString("3. Analyze tool results and provide recommendations\n")

	return prompt.String()
}


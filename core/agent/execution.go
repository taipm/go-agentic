// Package agent provides agent execution and related functionality.
package agent

import (
	"context"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"github.com/taipm/go-agentic/core/common"
	"github.com/taipm/go-agentic/core/logging"
	providers "github.com/taipm/go-agentic/core/providers"
	_ "github.com/taipm/go-agentic/core/providers/openai"  // Register OpenAI provider
	_ "github.com/taipm/go-agentic/core/providers/ollama"  // Register Ollama provider
)

// providerFactory is the global LLM provider factory instance
var providerFactory = providers.GetGlobalFactory()

// ExtractSignalsFromContent extracts signal markers from agent response text
// Signals are markers like [QUESTION], [ANSWER], [END_EXAM] etc.
func ExtractSignalsFromContent(content string) []string {
	var signals []string

	// Define known signal patterns to extract
	signalPatterns := []string{
		"[QUESTION]",
		"[ANSWER]",
		"[END_EXAM]",
		"[OK]",
		"[DONE]",
	}

	// Check each signal pattern in content
	for _, pattern := range signalPatterns {
		if strings.Contains(content, pattern) {
			signals = append(signals, pattern)
		}
	}

	return signals
}

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
		Tools:        ConvertAgentToolsToProviderTools(agent.Tools),
	}

	// Step 3: Execute provider call
	provider, err := providerFactory.GetProvider(modelConfig.Provider, modelConfig.ProviderURL, apiKey)
	if err != nil {
		return nil, fmt.Errorf("failed to get LLM provider '%s': %w", modelConfig.Provider, err)
	}

	// Log: LLM call start
	traceID := logging.GetTraceID(ctx)
	logging.GetLogger().InfoContext(ctx, "llm_call",
		slog.String("event", "llm_call"),
		slog.String("trace_id", traceID),
		slog.String("agent_id", agent.ID),
		slog.String("agent_name", agent.Name),
		slog.String("model", modelConfig.Model),
	)

	startTime := time.Now()
	response, err := provider.Complete(ctx, request)
	duration := time.Since(startTime)

	if err != nil {
		return nil, err
	}

	// Log: LLM response with actual cost
	if response.Usage != nil {
		actualCost := agent.CalculateCost(response.Usage.InputTokens, response.Usage.OutputTokens)

		logging.GetLogger().InfoContext(ctx, "llm_response",
			slog.String("event", "llm_response"),
			slog.String("trace_id", traceID),
			slog.String("agent_id", agent.ID),
			slog.String("agent_name", agent.Name),
			slog.String("model", modelConfig.Model),
			slog.Int("input_tokens", response.Usage.InputTokens),
			slog.Int("output_tokens", response.Usage.OutputTokens),
			slog.Int("total_tokens", response.Usage.TotalTokens),
			slog.Float64("cost_usd", actualCost),
			slog.Int64("duration_ms", duration.Milliseconds()),
		)
	}

	// Extract signals from response content
	signals := ExtractSignalsFromContent(response.Content)

	// Build cost summary
	var cost *common.CostSummary
	if response.Usage != nil {
		actualCost := agent.CalculateCost(response.Usage.InputTokens, response.Usage.OutputTokens)
		cost = &common.CostSummary{
			InputTokens:  response.Usage.InputTokens,
			OutputTokens: response.Usage.OutputTokens,
			TotalTokens:  response.Usage.TotalTokens,
			CostUSD:      actualCost,
		}
	}

	return &common.AgentResponse{
		AgentID:   agent.ID,
		AgentName: agent.Name,
		Content:   response.Content,
		ToolCalls: ConvertToolCallsFromProvider(response.ToolCalls),
		Signals:   signals,
		Cost:      cost,
	}, nil
}

// ConvertAgentToolsToProviderTools converts agent tools to provider format
// Handles various tool formats: Tool struct, map[string]interface{}, etc.
func ConvertAgentToolsToProviderTools(agentTools []interface{}) []providers.ProviderTool {
	var providerTools []providers.ProviderTool

	for _, tool := range agentTools {
		if tool == nil {
			continue
		}

		providerTool := convertSingleTool(tool)
		if providerTool != nil {
			providerTools = append(providerTools, *providerTool)
		}
	}

	return providerTools
}

// convertSingleTool converts a single tool to provider format
func convertSingleTool(tool interface{}) *providers.ProviderTool {
	// Handle *common.Tool
	if toolPtr, ok := tool.(*common.Tool); ok {
		return &providers.ProviderTool{
			Name:        toolPtr.Name,
			Description: toolPtr.Description,
			Parameters:  extractToolParameters(toolPtr.Parameters),
		}
	}

	// Handle common.Tool (by value)
	if toolVal, ok := tool.(common.Tool); ok {
		return &providers.ProviderTool{
			Name:        toolVal.Name,
			Description: toolVal.Description,
			Parameters:  extractToolParameters(toolVal.Parameters),
		}
	}

	// Handle providers.ProviderTool
	if providerTool, ok := tool.(providers.ProviderTool); ok {
		return &providerTool
	}

	// Unknown type - log warning
	fmt.Printf("[WARN] Unknown tool type in ConvertAgentToolsToProviderTools: %T\n", tool)
	return nil
}

// extractToolParameters safely extracts parameters from tool definition
func extractToolParameters(params interface{}) map[string]interface{} {
	if params == nil {
		return make(map[string]interface{})
	}

	if paramMap, ok := params.(map[string]interface{}); ok {
		return paramMap
	}

	return make(map[string]interface{})
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
		Tools:        ConvertAgentToolsToProviderTools(agent.Tools),
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


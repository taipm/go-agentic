package crewai

import (
	"context"
	"fmt"
	"strings"
	"time"

	providers "github.com/taipm/go-agentic/core/providers"
	_ "github.com/taipm/go-agentic/core/providers/openai"  // Register OpenAI provider
	_ "github.com/taipm/go-agentic/core/providers/ollama"  // Register Ollama provider
)

// providerFactory is the global LLM provider factory instance
var providerFactory = providers.GetGlobalFactory()

// ExecuteAgent runs an agent and returns its response
// Uses provider factory to support multiple LLM backends (OpenAI, Ollama, etc.)
// Supports fallback to backup LLM model if primary fails
// ✅ FIX for hardcoded model bug: Now uses agent.Primary.Model from config
// ✅ NEW: Backup LLM model support with automatic fallback
func ExecuteAgent(ctx context.Context, agent *Agent, input string, history []Message, apiKey string) (*AgentResponse, error) {
	// Get model configs (use new Primary/Backup if available, else fall back to old fields)
	primaryConfig := agent.Primary
	backupConfig := agent.Backup

	// Handle backward compatibility: if Primary is nil, use old format
	if primaryConfig == nil {
		// ✅ FIX #1: Validation instead of hardcoding "openai" default
		if agent.Provider == "" {
			return nil, fmt.Errorf("agent '%s': provider not specified in config - must be 'openai' or 'ollama'", agent.ID)
		}
		primaryConfig = &ModelConfig{
			Model:       agent.Model,
			Provider:    agent.Provider,
			ProviderURL: agent.ProviderURL,
		}
	}

	// Build system prompt once (reused for both primary and backup)
	systemPrompt := buildSystemPrompt(agent)
	messages := convertToProviderMessages(history)

	// 1️⃣ TRY PRIMARY MODEL
	response, primaryErr := executeWithModelConfig(ctx, agent, systemPrompt, messages, primaryConfig, apiKey)
	if primaryErr == nil {
		return response, nil
	}

	// 2️⃣ IF PRIMARY FAILED AND BACKUP EXISTS, TRY BACKUP
	if backupConfig != nil {
		fmt.Printf("[FALLBACK] Primary model '%s' (%s) failed: %v. Trying backup model '%s' (%s)...\n",
			primaryConfig.Model, primaryConfig.Provider, primaryErr, backupConfig.Model, backupConfig.Provider)

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

// validateAndCheckCostLimits estimates tokens and validates cost limits before execution
// Returns estimated token count for cost calculation
func validateAndCheckCostLimits(agent *Agent, systemPrompt string, messages []providers.ProviderMessage) (int, error) {
	// Estimate tokens from system prompt and messages
	systemAndPromptContent := systemPrompt
	for _, msg := range messages {
		systemAndPromptContent += msg.Content
	}
	estimatedTokens := agent.EstimateTokens(systemAndPromptContent)

	// Validate cost limits (BEFORE execution)
	if err := agent.CheckCostLimits(estimatedTokens); err != nil {
		return 0, err // BLOCK if limit exceeded
	}

	return estimatedTokens, nil
}

// buildCompletionRequest creates a provider-agnostic completion request
func buildCompletionRequest(modelConfig *ModelConfig, systemPrompt string, messages []providers.ProviderMessage, agent *Agent) *providers.CompletionRequest {
	return &providers.CompletionRequest{
		Model:        modelConfig.Model,
		SystemPrompt: systemPrompt,
		Messages:     messages,
		Temperature:  agent.Temperature,
		Tools:        convertToolsToProvider(agent.Tools),
	}
}

// executeProviderCall invokes the LLM provider and returns the response
// Tracks execution time for performance metrics
func executeProviderCall(ctx context.Context, agent *Agent, modelConfig *ModelConfig, request *providers.CompletionRequest, apiKey string) (*providers.CompletionResponse, time.Duration, error) {
	provider, err := providerFactory.GetProvider(modelConfig.Provider, modelConfig.ProviderURL, apiKey)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to get LLM provider '%s': %w", modelConfig.Provider, err)
	}

	fmt.Printf("[MODEL] Agent '%s' using model: %s (provider: %s)\n", agent.ID, modelConfig.Model, modelConfig.Provider)

	startTime := time.Now()
	response, err := provider.Complete(ctx, request)
	executionDuration := time.Since(startTime)

	return response, executionDuration, err
}

// validatePostExecutionQuotas checks memory and error quotas after execution
// Returns an error if quotas are exceeded
func validatePostExecutionQuotas(agent *Agent) error {
	if quotaErr := agent.CheckErrorQuota(); quotaErr != nil {
		return quotaErr
	}

	if memoryErr := agent.CheckMemoryQuota(); memoryErr != nil {
		fmt.Printf("[QUOTA] Agent %s exceeded memory quota: %v\n", agent.ID, memoryErr)
		return memoryErr
	}

	return nil
}

// updateMetricsOnSuccess updates cost, memory, and performance metrics after successful execution
func updateMetricsOnSuccess(agent *Agent, estimatedTokens int, response *providers.CompletionResponse, executionDuration time.Duration) {
	executionDurationMs := executionDuration.Milliseconds()

	// Update cost metrics
	actualCost := agent.CalculateCost(estimatedTokens)
	agent.UpdateCostMetrics(estimatedTokens, actualCost)

	// Update memory metrics (estimated from response size)
	responseSize := len(response.Content)
	for _, call := range response.ToolCalls {
		responseSize += len(call.ToolName) + len(call.ID)
		// Add size of arguments (rough JSON encoding estimate)
		for k, v := range call.Arguments {
			responseSize += len(k) + len(fmt.Sprintf("%v", v))
		}
	}
	estimatedMemoryMB := (responseSize / 1024)
	if estimatedMemoryMB < 1 {
		estimatedMemoryMB = 1
	}
	agent.UpdateMemoryMetrics(estimatedMemoryMB, executionDurationMs)

	// Update performance metrics
	agent.UpdatePerformanceMetrics(true, "")

	// Check slow call threshold
	agent.CheckSlowCall(executionDuration)

	// Log cost information
	agent.CostMetrics.Mutex.RLock()
	dailyCost := agent.CostMetrics.DailyCost
	callCount := agent.CostMetrics.CallCount
	totalTokens := agent.CostMetrics.TotalTokens
	agent.CostMetrics.Mutex.RUnlock()

	fmt.Printf("[COST] Agent '%s': +%d tokens ($%.6f) | Daily: %d tokens, $%.4f spent | Calls: %d\n",
		agent.ID, estimatedTokens, actualCost, totalTokens, dailyCost, callCount)

	// Log metadata
	LogMetadataMetrics(agent)
	LogMetadataQuotaStatus(agent)
	LogMemoryMetrics(agent)
	LogPerformanceMetrics(agent)
}

// executeWithModelConfig executes an agent with a specific model configuration
// Helper function used by ExecuteAgent to reduce code duplication
// ✅ WEEK 1: Now includes cost checking before execution and metric updates after
func executeWithModelConfig(ctx context.Context, agent *Agent, systemPrompt string, messages []providers.ProviderMessage, modelConfig *ModelConfig, apiKey string) (*AgentResponse, error) {
	// Step 1: Validate cost limits and estimate tokens
	estimatedTokens, err := validateAndCheckCostLimits(agent, systemPrompt, messages)
	if err != nil {
		return nil, err
	}

	// Step 2: Build completion request
	request := buildCompletionRequest(modelConfig, systemPrompt, messages, agent)

	// Step 3: Execute provider call
	response, executionDuration, err := executeProviderCall(ctx, agent, modelConfig, request, apiKey)
	if err != nil {
		agent.UpdatePerformanceMetrics(false, err.Error())
		if quotaErr := agent.CheckErrorQuota(); quotaErr != nil {
			// Quota enforcement is handled
		}
		return nil, err
	}

	// Step 4: Validate post-execution quotas
	if quotaErr := validatePostExecutionQuotas(agent); quotaErr != nil {
		return nil, quotaErr
	}

	// Step 5: Update all metrics on success
	updateMetricsOnSuccess(agent, estimatedTokens, response, executionDuration)

	return &AgentResponse{
		AgentID:   agent.ID,
		AgentName: agent.Name,
		Content:   response.Content,
		ToolCalls: convertToolCallsFromProvider(response.ToolCalls),
	}, nil
}

// ExecuteAgentStream runs an agent with streaming responses
// Streams response chunks to the provided channel
// Supports fallback to backup LLM model if primary fails
// ✅ NEW: Backup LLM model support with streaming fallback
func ExecuteAgentStream(ctx context.Context, agent *Agent, input string, history []Message, apiKey string, streamChan chan<- providers.StreamChunk) error {
	// Get model configs (use new Primary/Backup if available, else fall back to old fields)
	primaryConfig := agent.Primary
	backupConfig := agent.Backup

	// Handle backward compatibility: if Primary is nil, use old format
	if primaryConfig == nil {
		// ✅ FIX #1: Validation instead of hardcoding "openai" default
		if agent.Provider == "" {
			return fmt.Errorf("agent '%s': provider not specified in config - must be 'openai' or 'ollama'", agent.ID)
		}
		primaryConfig = &ModelConfig{
			Model:       agent.Model,
			Provider:    agent.Provider,
			ProviderURL: agent.ProviderURL,
		}
	}

	// Build system prompt once (reused for both primary and backup)
	systemPrompt := buildSystemPrompt(agent)
	messages := convertToProviderMessages(history)

	// 1️⃣ TRY PRIMARY MODEL WITH STREAMING
	primaryErr := executeWithModelConfigStream(ctx, agent, systemPrompt, messages, primaryConfig, apiKey, streamChan)
	if primaryErr == nil {
		return nil
	}

	// 2️⃣ IF PRIMARY FAILED AND BACKUP EXISTS, TRY BACKUP
	if backupConfig != nil {
		fmt.Printf("[FALLBACK] Primary model '%s' (%s) streaming failed: %v. Trying backup model '%s' (%s)...\n",
			primaryConfig.Model, primaryConfig.Provider, primaryErr, backupConfig.Model, backupConfig.Provider)

		backupErr := executeWithModelConfigStream(ctx, agent, systemPrompt, messages, backupConfig, apiKey, streamChan)
		if backupErr == nil {
			fmt.Printf("[FALLBACK SUCCESS] Backup model '%s' streaming succeeded\n", backupConfig.Model)
			return nil
		}

		// Both failed - return detailed error
		return fmt.Errorf("both primary and backup models failed for streaming: primary=%v, backup=%v", primaryErr, backupErr)
	}

	// No backup available - return primary error
	return fmt.Errorf("primary model streaming failed: %w", primaryErr)
}

// executeWithModelConfigStream executes an agent with streaming using a specific model configuration
// Helper function used by ExecuteAgentStream to reduce code duplication
// ✅ WEEK 1: Now includes cost checking before execution and metric updates after
func executeWithModelConfigStream(ctx context.Context, agent *Agent, systemPrompt string, messages []providers.ProviderMessage, modelConfig *ModelConfig, apiKey string, streamChan chan<- providers.StreamChunk) error {
	// ✅ Step 1: Estimate tokens BEFORE execution
	systemAndPromptContent := systemPrompt
	for _, msg := range messages {
		systemAndPromptContent += msg.Content
	}
	estimatedTokens := agent.EstimateTokens(systemAndPromptContent)

	// ✅ Step 2: Check cost limits (BEFORE execution)
	if err := agent.CheckCostLimits(estimatedTokens); err != nil {
		return err // BLOCK if limit exceeded (when EnforceCostLimits=true)
	}

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
		Tools:        convertToolsToProvider(agent.Tools),
	}

	// ✅ WEEK 3: Track execution time for performance metrics
	startTime := time.Now()

	// Call provider with streaming
	err = provider.CompleteStream(ctx, request, streamChan)

	// ✅ WEEK 3: Calculate execution duration
	executionDuration := time.Since(startTime)
	executionDurationMs := executionDuration.Milliseconds()

	// ✅ Step 3: Update metrics AFTER execution (success or failure)
	if err == nil {
		actualCost := agent.CalculateCost(estimatedTokens)
		agent.UpdateCostMetrics(estimatedTokens, actualCost)

		// ✅ WEEK 3: Update memory metrics (estimated for streaming - use token estimate as proxy)
		// For streaming, estimate memory based on tokens (rough: 1 token ≈ 4 bytes on average)
		estimatedMemoryMB := (estimatedTokens * 4) / (1024 * 1024)
		if estimatedMemoryMB < 1 {
			estimatedMemoryMB = 1 // Minimum 1 MB
		}
		agent.UpdateMemoryMetrics(estimatedMemoryMB, executionDurationMs)

		// ✅ WEEK 3: Update performance metrics on success
		agent.UpdatePerformanceMetrics(true, "")

		// ✅ WEEK 3: Check if this was a slow call
		agent.CheckSlowCall(executionDuration)

		// ✅ Log cost information for visibility
		agent.CostMetrics.Mutex.RLock()
		dailyCost := agent.CostMetrics.DailyCost
		callCount := agent.CostMetrics.CallCount
		totalTokens := agent.CostMetrics.TotalTokens
		agent.CostMetrics.Mutex.RUnlock()

		fmt.Printf("[COST] Agent '%s': +%d tokens ($%.6f) | Daily: %d tokens, $%.4f spent | Calls: %d\n",
			agent.ID, estimatedTokens, actualCost, totalTokens, dailyCost, callCount)

		// ✅ NEW: Automatic metadata logging - display quota, memory, and performance info
		LogMetadataMetrics(agent)
		LogMetadataQuotaStatus(agent)
		LogMemoryMetrics(agent)
		LogPerformanceMetrics(agent)
	} else {
		// ✅ WEEK 3: Update performance metrics on failure
		agent.UpdatePerformanceMetrics(false, err.Error())
		// Check and enforce error quotas
		if quotaErr := agent.CheckErrorQuota(); quotaErr != nil {
			// Quota enforcement is handled (warning or error based on settings)
		}
	}

	return err
}

// convertToProviderMessages converts internal Message format to provider-agnostic format
func convertToProviderMessages(history []Message) []providers.ProviderMessage {
	messages := make([]providers.ProviderMessage, len(history))
	for i, msg := range history {
		messages[i] = providers.ProviderMessage{
			Role:    msg.Role,
			Content: msg.Content,
		}
	}
	return messages
}

// convertToolsToProvider converts internal Tool format to provider-agnostic format
func convertToolsToProvider(tools []*Tool) []providers.ProviderTool {
	providerTools := make([]providers.ProviderTool, len(tools))
	for i, tool := range tools {
		providerTools[i] = providers.ProviderTool{
			Name:        tool.Name,
			Description: tool.Description,
			Parameters:  tool.Parameters,
		}
	}
	return providerTools
}

// convertToolCallsFromProvider converts provider tool calls back to internal format
func convertToolCallsFromProvider(providerCalls []providers.ToolCall) []ToolCall {
	calls := make([]ToolCall, len(providerCalls))
	for i, call := range providerCalls {
		calls[i] = ToolCall{
			ID:        call.ID,
			ToolName:  call.ToolName,
			Arguments: call.Arguments,
		}
	}
	return calls
}

// buildSystemPrompt creates the system prompt for the agent
// ✅ FIX for HIGH Issue #2: Now uses caching to prevent repeated token cost
// The system prompt is built once and cached for subsequent calls
func buildSystemPrompt(agent *Agent) string {
	// Check cache first (thread-safe read)
	agent.systemPromptMutex.RLock()
	if agent.cachedSystemPrompt != "" {
		cached := agent.cachedSystemPrompt
		agent.systemPromptMutex.RUnlock()
		return cached
	}
	agent.systemPromptMutex.RUnlock()

	// Build the prompt (cache miss)
	var builtPrompt string

	// If agent has a custom system prompt, use it (with template variable replacement)
	if agent.SystemPrompt != "" {
		prompt := agent.SystemPrompt
		// Replace template variables
		prompt = strings.ReplaceAll(prompt, "{{name}}", agent.Name)
		prompt = strings.ReplaceAll(prompt, "{{role}}", agent.Role)
		prompt = strings.ReplaceAll(prompt, "{{description}}", agent.Name+" - "+agent.Role)
		prompt = strings.ReplaceAll(prompt, "{{backstory}}", agent.Backstory)
		builtPrompt = prompt
	} else {
		// Otherwise, build a generic prompt
		var prompt strings.Builder

		prompt.WriteString(fmt.Sprintf("You are %s.\n", agent.Name))
		prompt.WriteString(fmt.Sprintf("Role: %s\n", agent.Role))
		prompt.WriteString(fmt.Sprintf("Backstory: %s\n\n", agent.Backstory))

		if len(agent.Tools) > 0 {
			prompt.WriteString("You have access to the following tools:\n\n")
			for i, tool := range agent.Tools {
				prompt.WriteString(fmt.Sprintf("%d. %s: %s\n", i+1, tool.Name, tool.Description))
			}

			prompt.WriteString("\nWhen you need to use a tool, write it exactly like this (on its own line):\n")
			prompt.WriteString("ToolName(param1, param2)\n\n")
			prompt.WriteString("Examples of tool calls:\n")
			prompt.WriteString("  GetCPUUsage()\n")
			prompt.WriteString("  PingHost(192.168.1.100)\n")
			prompt.WriteString("  CheckServiceStatus(nginx)\n\n")
		}

		prompt.WriteString("Instructions:\n")
		prompt.WriteString("1. Analyze the input and determine what tools you need\n")
		prompt.WriteString("2. Use tools to gather information\n")
		prompt.WriteString("3. Analyze tool results and provide recommendations\n")
		prompt.WriteString("4. If you need more information, use additional tools\n")

		if agent.IsTerminal {
			prompt.WriteString("5. You are the FINAL agent in the workflow - after you respond, the conversation ends\n")
		}

		builtPrompt = prompt.String()
	}

	// Cache the built prompt (thread-safe write)
	agent.systemPromptMutex.Lock()
	agent.cachedSystemPrompt = builtPrompt
	agent.systemPromptMutex.Unlock()

	return builtPrompt
}

// InvalidateSystemPromptCache clears the cached system prompt
// Call this if agent configuration changes (tools, role, etc.)
// ✅ FIX for HIGH Issue #2: Allow cache invalidation when needed
func (a *Agent) InvalidateSystemPromptCache() {
	a.systemPromptMutex.Lock()
	a.cachedSystemPrompt = ""
	a.systemPromptMutex.Unlock()
}

// parseToolArguments splits tool arguments respecting nested brackets
// Handles cases like: collection_name, [1.0, 2.0, 3.0], 5
func parseToolArguments(argsStr string) []string {
	var parts []string
	var current strings.Builder
	bracketDepth := 0
	parenDepth := 0

	for _, ch := range argsStr {
		switch ch {
		case '[':
			bracketDepth++
			current.WriteRune(ch)
		case ']':
			bracketDepth--
			current.WriteRune(ch)
		case '(':
			parenDepth++
			current.WriteRune(ch)
		case ')':
			parenDepth--
			current.WriteRune(ch)
		case ',':
			if bracketDepth == 0 && parenDepth == 0 {
				// This is a top-level comma, so split here
				part := strings.TrimSpace(current.String())
				if part != "" {
					parts = append(parts, part)
				}
				current.Reset()
			} else {
				// Comma is inside brackets, keep it
				current.WriteRune(ch)
			}
		default:
			current.WriteRune(ch)
		}
	}

	// Add last part
	if part := strings.TrimSpace(current.String()); part != "" {
		parts = append(parts, part)
	}

	return parts
}

// extractToolCallsFromText extracts tool calls from the response text
// ⚠️ FALLBACK METHOD: Uses text parsing (for models without tool_use support)
// This is kept for backward compatibility and edge cases only
// Preferred method is extractFromOpenAIToolCalls for robustness
func extractToolCallsFromText(text string, agent *Agent) []ToolCall {
	var calls []ToolCall

	validToolNames := make(map[string]*Tool)
	for _, tool := range agent.Tools {
		validToolNames[tool.Name] = tool
	}

	// Look for patterns like: ToolName(...)
	lines := strings.Split(text, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		// Try to find tool calls in this line
		for toolName := range validToolNames {
			if strings.Contains(line, toolName+"(") {
				// Extract the arguments
				startIdx := strings.Index(line, toolName+"(")
				if startIdx != -1 {
					endIdx := strings.Index(line[startIdx:], ")")
					if endIdx != -1 {
						endIdx += startIdx
						argsStr := line[startIdx+len(toolName)+1 : endIdx]

						// Parse arguments - handle nested arrays/objects
						args := make(map[string]interface{})
						if argsStr != "" {
							// Split arguments respecting nested brackets
							argParts := parseToolArguments(argsStr)

							// Map positional arguments to named parameters
							tool := validToolNames[toolName]
							paramNames := getToolParameterNames(tool)

							for i, part := range argParts {
								part = strings.TrimSpace(part)
								part = strings.Trim(part, `"'`)

								// Use parameter name if available, otherwise use arg0, arg1, etc.
								if i < len(paramNames) {
									args[paramNames[i]] = part
								} else {
									args[fmt.Sprintf("arg%d", i)] = part
								}
							}
						}

						calls = append(calls, ToolCall{
							ID:        fmt.Sprintf("%s_%d", toolName, len(calls)),
							ToolName:  toolName,
							Arguments: args,
						})
					}
				}
			}
		}
	}

	return calls
}

// getToolParameterNames extracts parameter names from tool definition in order
func getToolParameterNames(tool *Tool) []string {
	var paramNames []string

	if tool == nil || tool.Parameters == nil {
		return paramNames
	}

	// Extract properties from the tool definition
	if props, ok := tool.Parameters["properties"]; ok {
		if propsMap, ok := props.(map[string]interface{}); ok {
			// Get required parameters first (in order)
			if required, ok := tool.Parameters["required"]; ok {
				if requiredList, ok := required.([]string); ok {
					for _, paramName := range requiredList {
						if _, exists := propsMap[paramName]; exists {
							paramNames = append(paramNames, paramName)
						}
					}
				}
			}

			// Add optional parameters (those not in required list)
			requiredSet := make(map[string]bool)
			if required, ok := tool.Parameters["required"]; ok {
				if requiredList, ok := required.([]string); ok {
					for _, name := range requiredList {
						requiredSet[name] = true
					}
				}
			}

			// Go through properties in iteration order (maps are unordered, but this is best effort)
			for paramName := range propsMap {
				if !requiredSet[paramName] {
					paramNames = append(paramNames, paramName)
				}
			}
		}
	}

	return paramNames
}

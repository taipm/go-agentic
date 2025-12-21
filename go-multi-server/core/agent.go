package crewai

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	openai "github.com/openai/openai-go/v3"
	"github.com/openai/openai-go/v3/option"
)

// ✅ FIX for Issue #2 (Memory Leak): Client cache with TTL expiration
// clientEntry represents a cached OpenAI client with expiry time
type clientEntry struct {
	client    openai.Client
	createdAt time.Time
	expiresAt time.Time
}

// clientTTL defines how long a cached client remains valid before expiring
// Uses sliding window: TTL is refreshed on each access
const clientTTL = 1 * time.Hour

// OpenAI client caching for connection reuse
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

// init starts the cleanup goroutine
func init() {
	go cleanupExpiredClients()
}

// ExecuteAgent runs an agent and returns its response
func ExecuteAgent(ctx context.Context, agent *Agent, input string, history []Message, apiKey string) (*AgentResponse, error) {
	// Reuse cached client if available, otherwise create new one
	client := getOrCreateOpenAIClient(apiKey)

	// Build system prompt
	systemPrompt := buildSystemPrompt(agent)

	// Convert history to openai messages
	messages := buildOpenAIMessages(agent, input, history, systemPrompt)

	// Create completion request
	params := openai.ChatCompletionNewParams{
		Model:    "gpt-4o-mini",
		Messages: messages,
	}

	// Call OpenAI API
	completion, err := client.Chat.Completions.New(ctx, params)
	if err != nil {
		return nil, fmt.Errorf("failed to call OpenAI API: %w", err)
	}

	if len(completion.Choices) == 0 {
		return nil, fmt.Errorf("no choices in completion")
	}

	choice := completion.Choices[0]
	message := choice.Message

	// Extract response content
	content := message.Content

	// For now, extract tool calls by parsing the response text
	// This is a workaround until we get the proper tool calling API working
	toolCalls := extractToolCallsFromText(content, agent)

	return &AgentResponse{
		AgentID:   agent.ID,
		AgentName: agent.Name,
		Content:   content,
		ToolCalls: toolCalls,
	}, nil
}

// buildSystemPrompt creates the system prompt for the agent
func buildSystemPrompt(agent *Agent) string {
	// If agent has a custom system prompt, use it (with template variable replacement)
	if agent.SystemPrompt != "" {
		prompt := agent.SystemPrompt
		// Replace template variables
		prompt = strings.ReplaceAll(prompt, "{{name}}", agent.Name)
		prompt = strings.ReplaceAll(prompt, "{{role}}", agent.Role)
		prompt = strings.ReplaceAll(prompt, "{{description}}", agent.Name+" - "+agent.Role)
		prompt = strings.ReplaceAll(prompt, "{{backstory}}", agent.Backstory)
		return prompt
	}

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

	return prompt.String()
}

// buildOpenAIMessages converts history and input to OpenAI message format
func buildOpenAIMessages(agent *Agent, input string, history []Message, systemPrompt string) []openai.ChatCompletionMessageParamUnion {
	var messages []openai.ChatCompletionMessageParamUnion

	// Add system message
	messages = append(messages, openai.SystemMessage(systemPrompt))

	// Add conversation history
	for _, msg := range history {
		switch msg.Role {
		case "user":
			messages = append(messages, openai.UserMessage(msg.Content))
		case "assistant":
			messages = append(messages, openai.AssistantMessage(msg.Content))
		}
	}

	// Add current user input
	messages = append(messages, openai.UserMessage(input))

	return messages
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
// This uses a simple regex approach: ToolName(args)
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

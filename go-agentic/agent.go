package agentic

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	openai "github.com/openai/openai-go/v3"
	"github.com/openai/openai-go/v3/option"
)

// ExecuteAgent runs an agent and returns its response
func ExecuteAgent(ctx context.Context, agent *Agent, input string, history []Message, apiKey string) (*AgentResponse, error) {
	client := openai.NewClient(option.WithAPIKey(apiKey))

	// Log agent initialization
	fmt.Printf("[INFO] Agent '%s' (ID: %s) using model '%s' with temperature %.1f\n",
		agent.Name, agent.ID, agent.Model, agent.Temperature)

	// Build system prompt
	systemPrompt := buildSystemPrompt(agent)

	// Convert history to openai messages
	messages := buildOpenAIMessages(agent, input, history, systemPrompt)

	// Create completion request
	params := openai.ChatCompletionNewParams{
		Model:    agent.Model,
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

	// Try native tool calls first (official OpenAI API)
	var toolCalls []ToolCall
	if len(message.ToolCalls) > 0 {
		toolCalls = parseNativeToolCalls(message.ToolCalls, agent)
	}

	// Fallback to text parsing if no native tool calls
	if len(toolCalls) == 0 {
		toolCalls = extractToolCallsFromText(content, agent)
	}

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

		prompt.WriteString("\nWhen you need to use a tool, use the function calling mechanism.\n")
		prompt.WriteString("The system will handle executing the tool calls you emit.\n")
		prompt.WriteString("Provide tool calls in the proper format and I will execute them for you.\n\n")
		prompt.WriteString("Examples of tools you can call:\n")
		prompt.WriteString("  GetCPUUsage()\n")
		prompt.WriteString("  PingHost(host=\"192.168.1.100\")\n")
		prompt.WriteString("  CheckServiceStatus(service=\"nginx\")\n\n")
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

						// Parse arguments
						args := make(map[string]interface{})
						if argsStr != "" {
							// Split by comma and trim
							argParts := strings.Split(argsStr, ",")

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

// parseNativeToolCalls extracts tool calls from OpenAI's native tool_calls field
// This is the official API mechanism for tool calling
func parseNativeToolCalls(nativeToolCalls []openai.ChatCompletionMessageToolCallUnion, agent *Agent) []ToolCall {
	var calls []ToolCall

	validToolNames := make(map[string]*Tool)
	for _, tool := range agent.Tools {
		validToolNames[tool.Name] = tool
	}

	for _, nativeCall := range nativeToolCalls {
		// Only process function tool calls (not custom tools)
		if nativeCall.Type != "function" {
			continue
		}

		// Get function name and arguments
		toolName := nativeCall.Function.Name
		argumentsJSON := nativeCall.Function.Arguments

		// Only process if this is a valid tool for this agent
		if _, isValid := validToolNames[toolName]; !isValid {
			continue
		}

		// Parse arguments from JSON
		args := make(map[string]interface{})
		if argumentsJSON != "" {
			// Parse JSON arguments
			err := parseJSONArguments(argumentsJSON, args)
			if err != nil {
				// Log warning but continue - use empty args
				fmt.Printf("[WARN] Failed to parse JSON arguments for %s: %v\n", toolName, err)
			}
		}

		calls = append(calls, ToolCall{
			ID:        nativeCall.ID,
			ToolName:  toolName,
			Arguments: args,
		})
	}

	return calls
}

// parseJSONArguments parses JSON argument string into map
func parseJSONArguments(jsonStr string, args map[string]interface{}) error {
	// Use json.Unmarshal to parse the JSON string
	var parsedArgs map[string]interface{}

	// Try to parse as JSON object
	err := json.Unmarshal([]byte(jsonStr), &parsedArgs)
	if err != nil {
		return fmt.Errorf("failed to unmarshal JSON: %w", err)
	}

	// Copy parsed arguments to args map
	for key, value := range parsedArgs {
		args[key] = value
	}

	return nil
}

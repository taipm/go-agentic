package internal

import (
	"encoding/json"
	"fmt"
	"strings"

	agenticcore "github.com/taipm/go-agentic/core"
)

// MessageAnalyzerTools provides conversation analysis tools for agents
type MessageAnalyzerTools struct {
	executor *agenticcore.CrewExecutor
}

// NewMessageAnalyzerTools creates a new tool set for message analysis
func NewMessageAnalyzerTools(executor *agenticcore.CrewExecutor) *MessageAnalyzerTools {
	return &MessageAnalyzerTools{
		executor: executor,
	}
}

// GetMessageCount returns the count of messages in conversation
func (mat *MessageAnalyzerTools) GetMessageCount() map[string]interface{} {
	history := mat.executor.GetHistory()

	userCount := 0
	assistantCount := 0

	for _, msg := range history {
		if msg.Role == "user" {
			userCount++
		} else if msg.Role == "assistant" {
			assistantCount++
		}
	}

	result := map[string]interface{}{
		"count": len(history),
		"role_breakdown": map[string]int{
			"user":      userCount,
			"assistant": assistantCount,
		},
	}

	// Log tool call with results
	fmt.Printf("\n[TOOL CALL] get_message_count()\n")
	fmt.Printf("[TOOL RESULT] Total: %d messages (User: %d, Assistant: %d)\n\n", len(history), userCount, assistantCount)

	return result
}

// ConversationMessage represents a single message in the conversation
type ConversationMessage struct {
	Index   int    `json:"index"`
	Role    string `json:"role"`
	Content string `json:"content"`
}

// ExtractedFacts contains facts extracted from the conversation
type ExtractedFacts struct {
	UserName   string   `json:"user_name,omitempty"`
	KeyTopics  []string `json:"key_topics,omitempty"`
	Mentioned  []string `json:"mentioned,omitempty"`
}

// GetConversationSummary returns a summary of the conversation with extracted facts
func (mat *MessageAnalyzerTools) GetConversationSummary() map[string]interface{} {
	history := mat.executor.GetHistory()

	messages := make([]ConversationMessage, 0)
	for i, msg := range history {
		messages = append(messages, ConversationMessage{
			Index:   i,
			Role:    msg.Role,
			Content: msg.Content,
		})
	}

	// Extract facts from conversation
	facts := mat.extractFacts(history)

	result := map[string]interface{}{
		"total_messages": len(history),
		"messages":       messages,
		"extracted_facts": facts,
	}

	// Log tool call with results
	fmt.Printf("\n[TOOL CALL] get_conversation_summary()\n")
	fmt.Printf("[TOOL RESULT] %d messages found\n", len(history))
	if facts.UserName != "" {
		fmt.Printf("[EXTRACTED FACT] User Name: %s\n", facts.UserName)
	}
	if len(facts.KeyTopics) > 0 {
		fmt.Printf("[EXTRACTED FACTS] Topics: %v\n", facts.KeyTopics)
	}
	fmt.Printf("\n")

	return result
}

// SearchMessagesParams are the parameters for search_messages tool
type SearchMessagesParams struct {
	Query string `json:"query"`
	Limit int    `json:"limit"`
}

// SearchResult represents a single search result
type SearchResult struct {
	Index     int     `json:"index"`
	Role      string  `json:"role"`
	Content   string  `json:"content"`
	Relevance float64 `json:"relevance"`
}

// SearchMessages searches for keywords in the conversation history
func (mat *MessageAnalyzerTools) SearchMessages(query string, limit int) map[string]interface{} {
	if limit <= 0 {
		limit = 10
	}

	history := mat.executor.GetHistory()
	results := make([]SearchResult, 0)

	lowerQuery := strings.ToLower(query)

	for i, msg := range history {
		if strings.Contains(strings.ToLower(msg.Content), lowerQuery) {
			// Calculate simple relevance score
			relevance := 0.5
			if strings.EqualFold(msg.Content, query) {
				relevance = 1.0
			} else if strings.HasPrefix(strings.ToLower(msg.Content), lowerQuery) {
				relevance = 0.9
			}

			results = append(results, SearchResult{
				Index:     i,
				Role:      msg.Role,
				Content:   msg.Content,
				Relevance: relevance,
			})

			if len(results) >= limit {
				break
			}
		}
	}

	result := map[string]interface{}{
		"query":   query,
		"count":   len(results),
		"results": results,
	}

	// Log tool call with results
	fmt.Printf("\n[TOOL CALL] search_messages(query=\"%s\", limit=%d)\n", query, limit)
	fmt.Printf("[TOOL RESULT] Found %d matching messages\n", len(results))
	for _, r := range results {
		fmt.Printf("  [%d] %s: \"%s\" (relevance: %.1f)\n", r.Index, r.Role, r.Content, r.Relevance)
	}
	fmt.Printf("\n")

	return result
}

// CountMessagesByParams are the parameters for count_messages_by tool
type CountMessagesByParams struct {
	FilterBy    string `json:"filter_by"`
	FilterValue string `json:"filter_value,omitempty"`
}

// CountMessagesBy counts messages filtered by role or keyword
func (mat *MessageAnalyzerTools) CountMessagesBy(filterBy string, filterValue string) map[string]interface{} {
	history := mat.executor.GetHistory()
	count := 0

	switch strings.ToLower(filterBy) {
	case "role":
		for _, msg := range history {
			if strings.EqualFold(msg.Role, filterValue) {
				count++
			}
		}
		result := map[string]interface{}{
			"filter": map[string]string{
				"by":    "role",
				"value": filterValue,
			},
			"count": count,
		}
		fmt.Printf("\n[TOOL CALL] count_messages_by(filter_by=\"role\", filter_value=\"%s\")\n", filterValue)
		fmt.Printf("[TOOL RESULT] Found %d messages with role '%s'\n\n", count, filterValue)
		return result

	case "keyword":
		lowerValue := strings.ToLower(filterValue)
		for _, msg := range history {
			if strings.Contains(strings.ToLower(msg.Content), lowerValue) {
				count++
			}
		}
		result := map[string]interface{}{
			"filter": map[string]string{
				"by":    "keyword",
				"value": filterValue,
			},
			"count": count,
		}
		fmt.Printf("\n[TOOL CALL] count_messages_by(filter_by=\"keyword\", filter_value=\"%s\")\n", filterValue)
		fmt.Printf("[TOOL RESULT] Found %d messages containing keyword '%s'\n\n", count, filterValue)
		return result

	case "all":
		count = len(history)
		result := map[string]interface{}{
			"filter": map[string]string{
				"by": "all",
			},
			"count": count,
		}
		fmt.Printf("\n[TOOL CALL] count_messages_by(filter_by=\"all\")\n")
		fmt.Printf("[TOOL RESULT] Total message count: %d\n\n", count)
		return result

	default:
		fmt.Printf("\n[TOOL CALL] count_messages_by(filter_by=\"%s\", filter_value=\"%s\")\n", filterBy, filterValue)
		fmt.Printf("[TOOL ERROR] Invalid filter_by value. Use 'role', 'keyword', or 'all'\n\n")
		return map[string]interface{}{
			"error": "Invalid filter_by value. Use 'role', 'keyword', or 'all'",
		}
	}
}

// extractFacts extracts interesting facts from conversation history
func (mat *MessageAnalyzerTools) extractFacts(history []agenticcore.Message) ExtractedFacts {
	facts := ExtractedFacts{
		KeyTopics: make([]string, 0),
		Mentioned: make([]string, 0),
	}

	// Simple fact extraction: look for common patterns
	for _, msg := range history {
		content := msg.Content

		// Look for name patterns (very simple)
		if msg.Role == "user" {
			if strings.Contains(content, "Tôi là ") {
				// Extract name after "Tôi là "
				parts := strings.Split(content, "Tôi là ")
				if len(parts) > 1 {
					name := strings.TrimSpace(parts[1])
					// Remove common suffixes
					name = strings.TrimSuffix(name, ".")
					name = strings.TrimSuffix(name, "!")
					if name != "" && !strings.Contains(name, " ") {
						facts.UserName = name
					}
				}
			}

			// Extract topics
			lowerContent := strings.ToLower(content)
			if strings.Contains(lowerContent, "tên") {
				if !contains(facts.KeyTopics, "personal_info") {
					facts.KeyTopics = append(facts.KeyTopics, "personal_info")
				}
			}
		}
	}

	return facts
}

// contains checks if a string is in a slice
func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

// ToolExecutor executes tool calls and returns results
type ToolExecutor struct {
	tools *MessageAnalyzerTools
}

// NewToolExecutor creates a new tool executor
func NewToolExecutor(tools *MessageAnalyzerTools) *ToolExecutor {
	return &ToolExecutor{
		tools: tools,
	}
}

// ExecuteToolCall executes a tool call and returns the result as JSON string
func (te *ToolExecutor) ExecuteToolCall(toolName string, toolInput json.RawMessage) (string, error) {
	switch toolName {
	case "get_message_count":
		result := te.tools.GetMessageCount()
		jsonBytes, err := json.MarshalIndent(result, "", "  ")
		if err != nil {
			return "", fmt.Errorf("failed to marshal result: %w", err)
		}
		return string(jsonBytes), nil

	case "get_conversation_summary":
		result := te.tools.GetConversationSummary()
		jsonBytes, err := json.MarshalIndent(result, "", "  ")
		if err != nil {
			return "", fmt.Errorf("failed to marshal result: %w", err)
		}
		return string(jsonBytes), nil

	case "search_messages":
		var params SearchMessagesParams
		if err := json.Unmarshal(toolInput, &params); err != nil {
			return "", fmt.Errorf("failed to unmarshal params: %w", err)
		}
		result := te.tools.SearchMessages(params.Query, params.Limit)
		jsonBytes, err := json.MarshalIndent(result, "", "  ")
		if err != nil {
			return "", fmt.Errorf("failed to marshal result: %w", err)
		}
		return string(jsonBytes), nil

	case "count_messages_by":
		var params CountMessagesByParams
		if err := json.Unmarshal(toolInput, &params); err != nil {
			return "", fmt.Errorf("failed to unmarshal params: %w", err)
		}
		result := te.tools.CountMessagesBy(params.FilterBy, params.FilterValue)
		jsonBytes, err := json.MarshalIndent(result, "", "  ")
		if err != nil {
			return "", fmt.Errorf("failed to marshal result: %w", err)
		}
		return string(jsonBytes), nil

	default:
		return "", fmt.Errorf("unknown tool: %s", toolName)
	}
}

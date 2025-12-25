package main

import (
	"bufio"
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/joho/godotenv"
	agenticcore "github.com/taipm/go-agentic/core"
	"github.com/taipm/go-agentic/examples/00-hello-crew-tools/internal"
)

func main() {
	// Load .env file (optional - won't fail if not found)
	_ = godotenv.Load()

	serverMode := flag.Bool("server", false, "Run in server mode")
	port := flag.String("port", "8082", "Server port")
	flag.Parse()

	// Try to get API key - support both Ollama and OpenAI
	apiKey := os.Getenv("OPENAI_API_KEY")
	if apiKey == "" {
		// For Ollama, we can use a placeholder since it doesn't need API keys
		apiKey = "ollama"
		fmt.Println("ℹ️  Using Ollama (local) - no API key needed")
	} else {
		fmt.Printf("✅ OpenAI API key loaded from .env (length: %d chars)\n", len(apiKey))
	}

	executor, err := createExecutor(apiKey)
	if err != nil {
		fmt.Printf("Error creating executor: %v\n", err)
		os.Exit(1)
	}

	if *serverMode {
		runServer(executor, *port)
	} else {
		runCLI(executor)
	}
}

func createExecutor(apiKey string) (*agenticcore.CrewExecutor, error) {
	// Create tools map for the agent
	tools := createTools()

	// Create executor from configuration directory with tools
	// This loads crew.yaml and agent config files from config/
	// Tools are automatically assigned to agents based on their YAML configuration
	executor, err := agenticcore.NewCrewExecutorFromConfig(apiKey, "config", tools)
	if err != nil {
		return nil, fmt.Errorf("failed to create executor: %w", err)
	}

	return executor, nil
}

func createTools() map[string]*agenticcore.Tool {
	toolsMap := make(map[string]*agenticcore.Tool)

	// Tool 1: Get message count
	tool1 := &agenticcore.Tool{
		Name:        "GetMessageCount",
		Description: "Returns the total number of messages in the conversation, broken down by role (user/assistant)",
		Parameters: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{},
		},
		Func: func(ctx context.Context, args map[string]interface{}) (string, error) {
			// Note: In a real implementation, the tool would have access to conversation history
			// For now, this demonstrates the tool structure
			fmt.Printf("[TOOL EXECUTION] GetMessageCount() called\n")
			result := map[string]interface{}{
				"count": 0,
				"role_breakdown": map[string]int{
					"user":      0,
					"assistant": 0,
				},
			}
			jsonBytes, _ := json.Marshal(result)
			output := string(jsonBytes)
			fmt.Printf("[TOOL RESULT] GetMessageCount returned: %s\n", output)
			return output, nil
		},
	}
	toolsMap["GetMessageCount"] = tool1

	// Tool 2: Get conversation summary
	tool2 := &agenticcore.Tool{
		Name:        "GetConversationSummary",
		Description: "Returns all messages in the conversation and extracted facts (user name, topics mentioned, etc)",
		Parameters: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{},
		},
		Func: func(ctx context.Context, args map[string]interface{}) (string, error) {
			fmt.Printf("[TOOL EXECUTION] GetConversationSummary() called\n")
			result := map[string]interface{}{
				"total_messages":  0,
				"messages":        []interface{}{},
				"extracted_facts": map[string]interface{}{},
			}
			jsonBytes, _ := json.Marshal(result)
			output := string(jsonBytes)
			fmt.Printf("[TOOL RESULT] GetConversationSummary returned: %s\n", output)
			return output, nil
		},
	}
	toolsMap["GetConversationSummary"] = tool2

	// Tool 3: Search messages
	tool3 := &agenticcore.Tool{
		Name:        "SearchMessages",
		Description: "Search for keywords or phrases in the conversation history",
		Parameters: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"query": map[string]interface{}{
					"type":        "string",
					"description": "The search query (keyword or phrase)",
				},
				"limit": map[string]interface{}{
					"type":        "integer",
					"description": "Maximum number of results to return (default: 10)",
					"default":     10,
				},
			},
			"required": []string{"query"},
		},
		Func: func(ctx context.Context, args map[string]interface{}) (string, error) {
			fmt.Printf("[TOOL EXECUTION] SearchMessages() called with args: %v\n", args)
			result := map[string]interface{}{
				"query":   "",
				"results": []interface{}{},
			}
			jsonBytes, _ := json.Marshal(result)
			output := string(jsonBytes)
			fmt.Printf("[TOOL RESULT] SearchMessages returned: %s\n", output)
			return output, nil
		},
	}
	toolsMap["SearchMessages"] = tool3

	// Tool 4: Count messages by filter
	tool4 := &agenticcore.Tool{
		Name:        "CountMessagesBy",
		Description: "Count messages filtered by role (user/assistant) or by keyword presence",
		Parameters: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"filter_by": map[string]interface{}{
					"type":        "string",
					"description": "Filter type: 'role' to count by user/assistant, 'keyword' to count by content",
					"enum":        []interface{}{"role", "keyword"},
				},
				"filter_value": map[string]interface{}{
					"type":        "string",
					"description": "The value to filter by (e.g., 'user', 'assistant', or a keyword)",
				},
			},
			"required": []string{"filter_by", "filter_value"},
		},
		Func: func(ctx context.Context, args map[string]interface{}) (string, error) {
			fmt.Printf("[TOOL EXECUTION] CountMessagesBy() called with args: %v\n", args)
			result := map[string]interface{}{
				"filter_by":    "",
				"filter_value": "",
				"count":        0,
			}
			jsonBytes, _ := json.Marshal(result)
			output := string(jsonBytes)
			fmt.Printf("[TOOL RESULT] CountMessagesBy returned: %s\n", output)
			return output, nil
		},
	}
	toolsMap["CountMessagesBy"] = tool4

	// Tool 5: Get current time
	tool5 := &agenticcore.Tool{
		Name:        "GetCurrentTime",
		Description: "Returns the current date and time",
		Parameters: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{},
		},
		Func: func(ctx context.Context, args map[string]interface{}) (string, error) {
			now := time.Now()
			fmt.Printf("[TOOL EXECUTION] GetCurrentTime() called at %v\n", now)
			result := map[string]interface{}{
				"timestamp": now.Unix(),
				"datetime":  now.Format("2006-01-02 15:04:05"),
				"timezone":  now.Location().String(),
			}
			jsonBytes, _ := json.Marshal(result)
			output := string(jsonBytes)
			fmt.Printf("[TOOL RESULT] GetCurrentTime returned: %s\n", output)
			return output, nil
		},
	}
	toolsMap["GetCurrentTime"] = tool5

	return toolsMap
}

func runCLI(executor *agenticcore.CrewExecutor) {
	fmt.Println("Hello Crew with Tools - Interactive Mode")
	fmt.Println("=========================================")
	fmt.Println("Type your message and press Enter. Type 'exit' to quit.")
	fmt.Println()
	fmt.Println("Try these questions to test tool capability:")
	fmt.Println("  - Tôi tên gì?")
	fmt.Println("  - Tôi là John Doe")
	fmt.Println("  - Tôi đã hỏi mấy câu? (asks agent to count)")
	fmt.Println("  - Bạn nhớ tôi nói gì lần đầu? (asks agent to search)")
	fmt.Println()

	scanner := bufio.NewScanner(os.Stdin)
	for {
		fmt.Print("> ")
		if !scanner.Scan() {
			break
		}

		input := scanner.Text()
		if strings.ToLower(strings.TrimSpace(input)) == "exit" {
			break
		}

		if strings.TrimSpace(input) == "" {
			continue
		}

		ctx := context.Background()
		result, err := executor.Execute(ctx, input)
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			continue
		}

		// Check if response contains tool calls
		// This is a simplified version - in production you'd parse the actual tool call format
		if strings.Contains(result.Content, "GetMessageCount") ||
			strings.Contains(result.Content, "CountMessagesBy") ||
			strings.Contains(result.Content, "SearchMessages") ||
			strings.Contains(result.Content, "GetConversationSummary") ||
			strings.Contains(result.Content, "GetCurrentTime") {

			fmt.Println("[Agent is using tools to analyze conversation...]")
			fmt.Printf("[Tool Results]\n")

			// For now, just show the assistant response
			// In a real implementation, you'd parse actual tool calls from the LLM
		}

		fmt.Printf("Response: %s\n\n", result.Content)

		// Show conversation state
		history := executor.GetHistory()
		fmt.Printf("[Conversation state: %d messages total]\n\n", len(history))
	}
}

func runServer(executor *agenticcore.CrewExecutor, port string) {
	// Create tool executor
	toolExecutor := internal.NewToolExecutor(
		internal.NewMessageAnalyzerTools(executor),
	)

	http.HandleFunc("/execute", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		body, err := io.ReadAll(r.Body)
		if err != nil {
			http.Error(w, "Failed to read body", http.StatusBadRequest)
			return
		}
		defer r.Body.Close()

		var req struct {
			Input string `json:"input"`
		}
		if err := json.Unmarshal(body, &req); err != nil {
			http.Error(w, "Invalid JSON", http.StatusBadRequest)
			return
		}

		ctx := context.Background()
		result, err := executor.Execute(ctx, req.Input)
		if err != nil {
			http.Error(w, fmt.Sprintf("Execution error: %v", err), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"output":      result.Content,
			"message_count": len(executor.GetHistory()),
		})
	})

	// Endpoint to get tools information
	http.HandleFunc("/tools", func(w http.ResponseWriter, r *http.Request) {
		tools := map[string]interface{}{
			"get_message_count": map[string]string{
				"description": "Returns the total number of messages in the conversation",
			},
			"get_conversation_summary": map[string]string{
				"description": "Returns all messages and extracted facts from the conversation",
			},
			"search_messages": map[string]string{
				"description": "Search for keywords in conversation history",
			},
			"count_messages_by": map[string]string{
				"description": "Count messages filtered by role or keyword",
			},
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(tools)
	})

	// Endpoint for manual tool execution (for testing)
	http.HandleFunc("/execute-tool", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		body, err := io.ReadAll(r.Body)
		if err != nil {
			http.Error(w, "Failed to read body", http.StatusBadRequest)
			return
		}
		defer r.Body.Close()

		var req struct {
			Tool   string          `json:"tool"`
			Params json.RawMessage `json:"params"`
		}
		if err := json.Unmarshal(body, &req); err != nil {
			http.Error(w, "Invalid JSON", http.StatusBadRequest)
			return
		}

		result, err := toolExecutor.ExecuteToolCall(req.Tool, req.Params)
		if err != nil {
			http.Error(w, fmt.Sprintf("Tool execution error: %v", err), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintf(w, `{"result":%s}`, result)
	})

	http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{
			"status": "ok",
		})
	})

	fmt.Printf("Starting Hello Crew with Tools server on http://localhost:%s\n", port)
	fmt.Printf("Endpoints:\n")
	fmt.Printf("  POST /execute      - Execute crew with agent\n")
	fmt.Printf("  POST /execute-tool - Execute a specific tool\n")
	fmt.Printf("  GET  /tools        - List available tools\n")
	fmt.Printf("  GET  /health       - Health check\n")
	if err := http.ListenAndServe(":"+port, nil); err != nil {
		fmt.Printf("Server error: %v\n", err)
		os.Exit(1)
	}
}

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

	agenticcore "github.com/taipm/go-agentic/core"
)

func main() {
	serverMode := flag.Bool("server", false, "Run in server mode")
	port := flag.String("port", "8081", "Server port")
	flag.Parse()

	// Try to get API key - support both Ollama and OpenAI
	apiKey := os.Getenv("OPENAI_API_KEY")
	if apiKey == "" {
		// For Ollama, we can use a placeholder since it doesn't need API keys
		apiKey = "ollama"
		fmt.Println("ℹ️  Using Ollama (local) - no API key needed")
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
	// Create executor from configuration directory
	// This loads crew.yaml and agent config files from config/
	executor, err := agenticcore.NewCrewExecutorFromConfig(apiKey, "config", nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create executor: %w", err)
	}
	return executor, nil
}

func runCLI(executor *agenticcore.CrewExecutor) {
	fmt.Println("Hello Crew - Interactive Mode")
	fmt.Println("==============================")
	fmt.Println("Type your message and press Enter. Type 'exit' to quit.")
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

		fmt.Printf("Response: %s\n\n", result.Content)
	}
}

func runServer(executor *agenticcore.CrewExecutor, port string) {
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
		json.NewEncoder(w).Encode(map[string]string{
			"output": result.Content,
		})
	})

	http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{
			"status": "ok",
		})
	})

	fmt.Printf("Starting Hello Crew server on http://localhost:%s\n", port)
	fmt.Printf("Use /execute endpoint with POST and JSON body: {\"input\": \"your message\"}\n")
	if err := http.ListenAndServe(":"+port, nil); err != nil {
		fmt.Printf("Server error: %v\n", err)
		os.Exit(1)
	}
}

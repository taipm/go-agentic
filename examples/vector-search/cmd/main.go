package main

import (
	"bufio"
	"context"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/joho/godotenv"
	crewai "github.com/taipm/go-agentic/core"
	"github.com/taipm/go-agentic/examples/vector-search/internal"
)

func main() {
	// Load environment variables from .env file
	if err := godotenv.Load(".env"); err != nil {
		fmt.Println("⚠️  No .env file found, using system environment variables")
	}

	// Parse command-line flags
	serverMode := flag.Bool("server", false, "Run in server mode with HTTP API")
	port := flag.Int("port", 0, "HTTP server port (default: from .env SERVER_PORT or 8082)")
	flag.Parse()

	// Get API keys from environment
	openaiKey := os.Getenv("OPENAI_API_KEY")
	if openaiKey == "" {
		fmt.Println("❌ Error: OPENAI_API_KEY not set in .env or environment variables")
		os.Exit(1)
	}

	qdrantURL := os.Getenv("QDRANT_URL")
	if qdrantURL == "" {
		fmt.Println("❌ Error: QDRANT_URL not set in .env or environment variables")
		os.Exit(1)
	}

	qdrantKey := os.Getenv("QDRANT_API_KEY")
	if qdrantKey == "" {
		fmt.Println("❌ Error: QDRANT_API_KEY not set in .env or environment variables")
		os.Exit(1)
	}

	// Determine server port: command-line flag > .env SERVER_PORT > default 8082
	finalPort := *port
	if finalPort == 0 {
		portFromEnv := os.Getenv("SERVER_PORT")
		if portFromEnv != "" {
			if p, err := strconv.Atoi(portFromEnv); err == nil {
				finalPort = p
			} else {
				finalPort = 8082
			}
		} else {
			finalPort = 8082
		}
	}

	// Create Qdrant client
	fmt.Println("🔌 Connecting to Qdrant...")
	qc, err := internal.NewQdrantClient(qdrantURL, qdrantKey)
	if err != nil {
		fmt.Printf("❌ Failed to connect to Qdrant: %v\n", err)
		os.Exit(1)
	}
	defer qc.Close()

	fmt.Println("✅ Connected to Qdrant")

	// Get all tools
	allTools := internal.GetAllQdrantTools(qc, openaiKey)

	// Load crew configuration
	configDir := "config"
	executor, err := crewai.NewCrewExecutorFromConfig(openaiKey, configDir, allTools)
	if err != nil {
		fmt.Printf("❌ Failed to load crew configuration: %v\n", err)
		os.Exit(1)
	}

	// Run in appropriate mode
	if *serverMode {
		runServerMode(executor, finalPort)
	} else {
		runCLIMode(executor)
	}
}

// runCLIMode runs the interactive CLI mode
func runCLIMode(executor *crewai.CrewExecutor) {
	fmt.Println("\n🔍 Qdrant Vector Search - Interactive CLI Mode")
	fmt.Println("=" + strings.Repeat("=", 50))
	fmt.Println("Type your question in Vietnamese (Tiếng Việt)")
	fmt.Println("Type 'exit' or 'quit' to exit")
	fmt.Println("=" + strings.Repeat("=", 50) + "\n")

	reader := bufio.NewReader(os.Stdin)

	for {
		fmt.Print("> ")
		input, err := reader.ReadString('\n')
		if err != nil {
			fmt.Printf("Error reading input: %v\n", err)
			continue
		}

		input = strings.TrimSpace(input)
		if input == "" {
			continue
		}

		if input == "exit" || input == "quit" {
			fmt.Println("\n👋 Goodbye!")
			break
		}

		// Execute query with streaming
		ctx, cancel := context.WithTimeout(context.Background(), 5*60*1000000000) // 5 minutes
		streamChan := make(chan *crewai.StreamEvent, 100)

		fmt.Println("\n" + strings.Repeat("-", 52))

		go func() {
			err := executor.ExecuteStream(ctx, input, streamChan)
			if err != nil {
				fmt.Printf("Error: %v\n", err)
			}
			close(streamChan)
		}()

		// Process stream events
		for event := range streamChan {
			displayStreamEvent(event)
		}

		cancel()
		fmt.Println(strings.Repeat("-", 52) + "\n")
	}
}

// displayStreamEvent prints a stream event in a readable format
func displayStreamEvent(event *crewai.StreamEvent) {
	// Add timestamp for performance tracking
	timestamp := event.Timestamp.Format("15:04:05.000")

	// Hide long embedding vectors in tool_result
	content := event.Content
	if event.Type == "tool_result" && strings.Contains(content, "Embedding generated") {
		// Extract just the first line (summary) and hide the JSON vector
		lines := strings.Split(content, "\n")
		if len(lines) > 1 {
			content = lines[0] + " (vector hidden for readability)"
		}
	}

	switch event.Type {
	case "start":
		fmt.Printf("[%s] 🚀 %s\n", timestamp, content)
	case "agent_start":
		fmt.Printf("[%s] 🔄 [%s] %s\n", timestamp, event.Agent, content)
	case "agent_response":
		fmt.Printf("[%s] 💬 [%s] %s\n", timestamp, event.Agent, content)
	case "tool_start":
		fmt.Printf("[%s] 🔧 [%s] %s\n", timestamp, event.Agent, content)
	case "tool_result":
		fmt.Printf("[%s] ✅ [%s] %s\n", timestamp, event.Agent, content)
	case "pause":
		fmt.Printf("[%s] ⏸️  [%s] %s\n", timestamp, event.Agent, content)
	case "warning":
		fmt.Printf("[%s] ⚠️  [%s] %s\n", timestamp, event.Agent, content)
	case "error":
		fmt.Printf("[%s] ❌ [%s] %s\n", timestamp, event.Agent, content)
	case "done":
		fmt.Printf("[%s] ✨ %s\n", timestamp, content)
	default:
		fmt.Printf("[%s] 📝 [%s - %s] %s\n", timestamp, event.Type, event.Agent, content)
	}
}

// runServerMode runs the HTTP server mode
func runServerMode(executor *crewai.CrewExecutor, port int) {
	fmt.Printf("\n🚀 Starting HTTP server on port %d\n", port)
	fmt.Println("=" + strings.Repeat("=", 50))
	fmt.Printf("Access the web UI at: http://localhost:%d\n", port)
	fmt.Printf("API Endpoint: POST http://localhost:%d/api/crew/stream\n", port)
	fmt.Println("=" + strings.Repeat("=", 50))

	// Try to load custom Vietnamese web UI
	webUIPath := filepath.Join("web", "client.html")
	htmlContent := ""

	if data, err := os.ReadFile(webUIPath); err == nil {
		htmlContent = string(data)
		fmt.Printf("✅ Loaded Vietnamese web UI from %s\n\n", webUIPath)
	} else {
		fmt.Printf("⚠️  Could not load Vietnamese web UI from %s, using default UI\n\n", webUIPath)
	}

	var err error
	if htmlContent != "" {
		err = crewai.StartHTTPServerWithCustomUI(executor, port, htmlContent)
	} else {
		err = crewai.StartHTTPServer(executor, port)
	}

	if err != nil {
		fmt.Printf("❌ Server error: %v\n", err)
		os.Exit(1)
	}
}

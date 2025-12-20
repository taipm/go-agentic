package main

import (
	"bufio"
	"context"
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/taipm/go-agentic/examples/it-support/internal"
	agenticcore "github.com/taipm/go-agentic/core"
)

func main() {
	// Parse command line flags
	serverMode := flag.Bool("server", false, "Run in server mode with HTTP API and web UI")
	port := flag.Int("port", 8081, "HTTP server port (only used with --server)")
	flag.Parse()

	// Create the IT Support crew
	crew := internal.CreateITSupportCrew()

	// Get API key from environment
	apiKey := os.Getenv("OPENAI_API_KEY")
	if apiKey == "" {
		fmt.Fprintf(os.Stderr, "Error: OPENAI_API_KEY environment variable not set\n")
		os.Exit(1)
	}

	// Create crew executor
	executor := agenticcore.NewCrewExecutor(crew, apiKey)

	// Check if server mode is requested
	if *serverMode {
		// Start HTTP server
		fmt.Printf("🚀 Starting IT Support HTTP Server on port %d\n", *port)
		fmt.Printf("📡 SSE Endpoint: http://localhost:%d/api/crew/stream\n", *port)
		fmt.Printf("🌐 Web Client: http://localhost:%d\n", *port)
		fmt.Printf("📝 Test Client: http://localhost:%d/web/test_sse_client.html\n", *port)
		if err := agenticcore.StartHTTPServer(executor, *port); err != nil {
			fmt.Fprintf(os.Stderr, "Error starting server: %v\n", err)
			os.Exit(1)
		}
		return
	}

	// CLI mode (interactive)
	reader := bufio.NewReader(os.Stdin)
	fmt.Println("=== IT Support System (CLI) ===")
	fmt.Println("Describe your IT issue:")

	input, err := reader.ReadString('\n')
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error reading input: %v\n", err)
		os.Exit(1)
	}

	task := strings.TrimSpace(input)
	if task == "" {
		fmt.Println("No input provided. Exiting.")
		os.Exit(0)
	}

	// Execute the crew with the user's task
	ctx := context.Background()
	result, err := executor.Execute(ctx, task)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error executing crew: %v\n", err)
		os.Exit(1)
	}

	// Print results
	fmt.Println("\n=== Results ===")
	fmt.Println(result.Content)
}

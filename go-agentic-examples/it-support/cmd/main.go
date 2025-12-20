package main

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/taipm/go-agentic-examples/it-support/internal"
	"github.com/taipm/go-crewai"
)

func main() {
	// Create the IT Support crew
	crew := internal.CreateITSupportCrew()

	// Get API key from environment
	apiKey := os.Getenv("OPENAI_API_KEY")
	if apiKey == "" {
		fmt.Fprintf(os.Stderr, "Error: OPENAI_API_KEY environment variable not set\n")
		os.Exit(1)
	}

	// Create crew executor
	executor := crewai.NewCrewExecutor(crew, apiKey)

	// Interactive input from user
	reader := bufio.NewReader(os.Stdin)
	fmt.Println("=== IT Support System ===")
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

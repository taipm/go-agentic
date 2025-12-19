package main

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/taipm/go-agentic"
)

func main() {
	// Load environment
	if err := loadEnvFile(); err != nil {
		fmt.Printf("Note: %v\n", err)
	}

	apiKey := os.Getenv("OPENAI_API_KEY")
	if apiKey == "" {
		fmt.Println("Error: OPENAI_API_KEY environment variable not set")
		os.Exit(1)
	}

	// Create customer service crew
	crew := createCustomerServiceCrew()
	executor := agentic.NewTeamExecutor(crew, apiKey)

	// Run interactive loop
	runCustomerServiceLoop(executor)
}

func loadEnvFile() error {
	data, err := os.ReadFile(".env")
	if err != nil {
		return fmt.Errorf("no .env file found")
	}

	lines := strings.Split(string(data), "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		parts := strings.SplitN(line, "=", 2)
		if len(parts) == 2 {
			key := strings.TrimSpace(parts[0])
			value := strings.TrimSpace(parts[1])
			os.Setenv(key, value)
		}
	}
	return nil
}

func runCustomerServiceLoop(executor *agentic.TeamExecutor) {
	fmt.Println("\n🎯 Customer Service Support System v1.0")
	fmt.Println("======================================")
	fmt.Println("Chat with our customer service team (type 'quit' to exit)")
	fmt.Println()

	scanner := bufio.NewScanner(os.Stdin)

	for {
		fmt.Print("You: ")
		if !scanner.Scan() {
			break
		}

		input := strings.TrimSpace(scanner.Text())
		if input == "" {
			continue
		}

		if input == "quit" {
			fmt.Println("Thank you for contacting us. Goodbye!")
			break
		}

		response, err := executor.Execute(context.Background(), input)
		if err != nil {
			fmt.Printf("Error: %v\n\n", err)
			continue
		}

		fmt.Printf("\n✅ Response from: %s\n", response.AgentName)
		fmt.Printf("Message: %s\n\n", response.Content)

		if response.IsTerminal {
			fmt.Println("[Session concluded]\n")
		}
	}
}

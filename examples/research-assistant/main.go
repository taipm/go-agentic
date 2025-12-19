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

	// Create research assistant crew
	crew := createResearchAssistantCrew()
	executor := agentic.NewTeamExecutor(crew, apiKey)

	// Run interactive loop
	runResearchLoop(executor)
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

func runResearchLoop(executor *agentic.TeamExecutor) {
	fmt.Println("\n🔬 Research Assistant Team v1.0")
	fmt.Println("================================")
	fmt.Println("Ask our research team to investigate topics (type 'quit' to exit)")
	fmt.Println()
	fmt.Println("Example queries:")
	fmt.Println("- Research the latest developments in machine learning")
	fmt.Println("- Find key authors in AI safety research")
	fmt.Println("- Identify research gaps in deep learning")
	fmt.Println("- Analyze citation patterns in quantum computing papers")
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
			fmt.Println("Thank you for using Research Assistant Team. Goodbye!")
			break
		}

		response, err := executor.Execute(context.Background(), input)
		if err != nil {
			fmt.Printf("Error: %v\n\n", err)
			continue
		}

		fmt.Printf("\n🎓 %s says:\n", response.AgentName)
		fmt.Printf("%s\n\n", response.Content)

		if response.IsTerminal {
			fmt.Println("[Research complete]\n")
		}
	}
}

package main

import (
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
		fmt.Println("❌ Error: OPENAI_API_KEY environment variable not set")
		fmt.Println("📝 Please add your API key to .env file or set it as environment variable")
		os.Exit(1)
	}

	// Create simple chat crew with 2 agents
	crew := createSimpleChatCrew()
	executor := agentic.NewTeamExecutor(crew, apiKey)

	// Test topics for the agents to discuss
	topics := []string{
		"What are the best practices for writing Go code?",
		"How can AI agents improve software development?",
		"Tell me about the latest trends in machine learning",
	}

	fmt.Println("🤖 Simple Multi-Agent Chat System")
	fmt.Println("==================================================")
	fmt.Println()

	for _, topic := range topics {
		fmt.Printf("📌 Topic: %s\n", topic)
		fmt.Println("--------------------------------------------------")

		ctx := context.Background()
		response, err := executor.Execute(ctx, topic)
		if err != nil {
			fmt.Printf("❌ Error: %v\n", err)
			continue
		}

		fmt.Printf("✅ Final Response:\n%s\n\n", response)
	}
}

// createSimpleChatCrew creates a crew with 2 simple agents that discuss topics
func createSimpleChatCrew() *agentic.Crew {
	// Agent 1: Enthusiast - Asks questions and explores ideas
	enthusiast := &agentic.Agent{
		ID:          "enthusiast",
		Name:        "Enthusiast",
		Role:        "Curious learner who asks insightful questions",
		Backstory:   "You are an enthusiastic learner who loves exploring new ideas. Your role is to ask the Expert thoughtful questions and engage in meaningful discussion.",
		Model:       "gpt-4o-mini",
		Tools:       []*agentic.Tool{},
		Temperature: 0.8,
		IsTerminal:  false,
	}

	// Agent 2: Expert - Provides answers and expertise
	expert := &agentic.Agent{
		ID:          "expert",
		Name:        "Expert",
		Role:        "Subject matter expert with deep knowledge",
		Backstory:   "You are a knowledgeable expert in various fields. Your role is to provide comprehensive answers to questions and share your expertise with the Enthusiast. Provide clear, concise, and helpful responses.",
		Model:       "gpt-4o-mini",
		Tools:       []*agentic.Tool{},
		Temperature: 0.7,
		IsTerminal:  true,
	}

	// Create crew with 2 agents
	crew := &agentic.Crew{
		Agents:      []*agentic.Agent{enthusiast, expert},
		MaxRounds:   3,      // Allow 3 rounds of conversation
		MaxHandoffs: 2,      // Allow 2 handoffs between agents
	}

	return crew
}

// loadEnvFile loads environment variables from .env file
func loadEnvFile() error {
	data, err := os.ReadFile(".env")
	if err != nil {
		return fmt.Errorf("no .env file found - using system environment variables")
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

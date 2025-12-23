package main

import (
	"context"
	"fmt"
	"os"

	agenticcore "github.com/taipm/go-agentic/core"
)

func runDebug() {
	apiKey := os.Getenv("OPENAI_API_KEY")
	if apiKey == "" {
		apiKey = "ollama"
		fmt.Println("ℹ️  Using Ollama (local) - no API key needed")
	}

	executor, err := agenticcore.NewCrewExecutorFromConfig(apiKey, "config", nil)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	ctx := context.Background()

	fmt.Println("\n=== TEST 1: User introduces ===")
	fmt.Println("USER: Tôi tên Tài đó nha")
	result1, _ := executor.Execute(ctx, "Tôi tên Tài đó nha")
	fmt.Printf("AGENT: %s\n", result1.Content)
	
	history := executor.GetHistory()
	fmt.Printf("\nHISTORY NOW (%d messages):\n", len(history))
	for i, msg := range history {
		content := msg.Content
		if len(content) > 100 {
			content = content[:100] + "..."
		}
		fmt.Printf("  [%d] %s: %s\n", i, msg.Role, content)
	}

	fmt.Println("\n=== TEST 2: User asks name ===")
	fmt.Println("USER: Tôi tên gì vậy ?")
	result2, _ := executor.Execute(ctx, "Tôi tên gì vậy ?")
	fmt.Printf("AGENT: %s\n", result2.Content)
	
	history = executor.GetHistory()
	fmt.Printf("\nHISTORY NOW (%d messages):\n", len(history))
	for i, msg := range history {
		content := msg.Content
		if len(content) > 100 {
			content = content[:100] + "..."
		}
		fmt.Printf("  [%d] %s: %s\n", i, msg.Role, content)
	}
}

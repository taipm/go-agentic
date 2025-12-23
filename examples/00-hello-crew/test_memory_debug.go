package main

import (
	"context"
	"fmt"
	"os"

	agenticcore "github.com/taipm/go-agentic/core"
)

// Test để debug tại sao crew không nhớ tên
func testMemoryDebug() {
	apiKey := os.Getenv("OPENAI_API_KEY")
	if apiKey == "" {
		apiKey = "ollama"
	}

	executor, err := agenticcore.NewCrewExecutorFromConfig(apiKey, "config", nil)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	ctx := context.Background()

	// MESSAGE 1: User says their name
	fmt.Println("\n=== MESSAGE 1: User introduces themselves ===")
	fmt.Println("USER INPUT: Tôi tên Tài đó nha")

	result1, err := executor.Execute(ctx, "Tôi tên Tài đó nha")
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	fmt.Printf("AGENT RESPONSE: %s\n", result1.Content)

	// DEBUG: Print history after first message
	printHistoryDebug(executor, "After message 1")

	// MESSAGE 2: User asks their name
	fmt.Println("\n=== MESSAGE 2: User asks their name ===")
	fmt.Println("USER INPUT: Tôi tên gì vậy ?")

	result2, err := executor.Execute(ctx, "Tôi tên gì vậy ?")
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	fmt.Printf("AGENT RESPONSE: %s\n", result2.Content)

	// DEBUG: Print history after second message
	printHistoryDebug(executor, "After message 2")

	// MESSAGE 3: Another question
	fmt.Println("\n=== MESSAGE 3: Ask for full name ===")
	fmt.Println("USER INPUT: Tên đầy đủ của tôi là gì?")

	result3, err := executor.Execute(ctx, "Tên đầy đủ của tôi là gì?")
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	fmt.Printf("AGENT RESPONSE: %s\n", result3.Content)

	// DEBUG: Print final history
	printHistoryDebug(executor, "After message 3")
}

func printHistoryDebug(executor *agenticcore.CrewExecutor, label string) {
	history := executor.GetHistory()
	fmt.Printf("\n📝 HISTORY (%s): %d messages\n", label, len(history))
	for i, msg := range history {
		fmt.Printf("  [%d] Role=%q Content=%q\n", i, msg.Role, truncate(msg.Content, 80))
	}
}

func truncate(s string, maxLen int) string {
	if len(s) > maxLen {
		return s[:maxLen] + "..."
	}
	return s
}

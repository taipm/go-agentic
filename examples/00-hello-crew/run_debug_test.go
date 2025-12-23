package main

import (
	"os"
	"context"
	"fmt"
	agenticcore "github.com/taipm/go-agentic/core"
)

func main() {
	testMemoryDebug()
}

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
	printHistoryDebug(executor, "After message 2")
}

func printHistoryDebug(executor *agenticcore.CrewExecutor, label string) {
	history := executor.GetHistory()
	fmt.Printf("\n📝 HISTORY (%s): %d messages\n", label, len(history))
	for i, msg := range history {
		fmt.Printf("  [%d] Role=%q Content=%q\n", i, msg.Role, truncate(msg.Content, 100))
	}
}

func truncate(s string, maxLen int) string {
	if len(s) > maxLen {
		return s[:maxLen] + "..."
	}
	return s
}

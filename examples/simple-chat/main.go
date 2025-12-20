package main

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/taipm/go-agentic"
)

func main() {
	// Load API key
	apiKey := getEnvVar("OPENAI_API_KEY")
	if apiKey == "" {
		fmt.Println("❌ Lỗi: OPENAI_API_KEY không được thiết lập")
		os.Exit(1)
	}

	// Load team from YAML (no tools needed for this example)
	team, err := agentic.LoadTeamFromYAML("team.yaml", agentic.ToolHandlerRegistry{})
	if err != nil {
		fmt.Printf("❌ Lỗi tải cấu hình: %v\n", err)
		os.Exit(1)
	}

	// Create executor and run
	executor := agentic.NewTeamExecutor(team, apiKey)
	fmt.Println("\n🤖 Hệ Thống Thảo Luận Multi-Agent\n" + strings.Repeat("=", 50))

	// Sample topics for demonstration
	topics := []string{
		"Những thực hành tốt nhất khi viết code Go là gì?",
		"Làm thế nào mà các AI agent có thể cải thiện phát triển phần mềm?",
	}

	for i, topic := range topics {
		fmt.Printf("\n📌 Chủ đề %d: %s\n%s\n", i+1, topic, strings.Repeat("-", 50))
		resp, err := executor.Execute(context.Background(), topic)
		if err == nil {
			fmt.Printf("✅ %s\n", resp.Content)
		} else {
			fmt.Printf("❌ Lỗi: %v\n", err)
		}
	}
	fmt.Println("\n" + strings.Repeat("=", 50) + "\n🎉 Hoàn thành!\n")
}

func getEnvVar(key string) string {
	data, _ := os.ReadFile(".env")
	for _, line := range strings.Split(string(data), "\n") {
		if strings.HasPrefix(strings.TrimSpace(line), key+"=") {
			return strings.TrimSpace(strings.Split(line, "=")[1])
		}
	}
	return ""
}

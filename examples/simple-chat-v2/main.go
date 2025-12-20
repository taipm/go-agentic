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
	loadEnv()
	key := os.Getenv("OPENAI_API_KEY")
	if key == "" {
		fmt.Println("❌ Lỗi: OPENAI_API_KEY không được thiết lập")
		os.Exit(1)
	}

	// Create agents using fluent builder API
	enthusiast := agentic.NewAgent("enthusiast", "Người Tò Mò").
		WithRole("Người học hỏi đầy tò mò").
		WithBackstory(`Bạn là một người yêu thích khám phá những ý tưởng mới.
Vai trò của bạn là đặt những câu hỏi sâu sắc cho Chuyên Gia
và tham gia vào cuộc thảo luận có ý nghĩa về các chủ đề khác nhau.
Hãy nói tiếng Việt một cách tự nhiên và thân thiện.`).
		WithModel("gpt-4o-mini").
		WithTemperature(0.8).
		SetTerminal(false).
		WithHandoff("expert").
		Build()

	expert := agentic.NewAgent("expert", "Chuyên Gia").
		WithRole("Chuyên gia có kiến thức sâu").
		WithBackstory(`Bạn là một chuyên gia thông thái với kiến thức sâu rộng
về nhiều lĩnh vực khác nhau. Vai trò của bạn là cung cấp
những câu trả lời toàn diện cho Người Tò Mò.
Hãy chia sẻ kiến thức của bạn bằng tiếng Việt
một cách rõ ràng, súc tích và hữu ích.
Luôn thân thiện và dễ hiểu khi giải thích.`).
		WithModel("gpt-4o-mini").
		WithTemperature(0.7).
		SetTerminal(true).
		Build()

	// Create team using fluent builder API
	team := agentic.NewTeam().
		AddAgents(enthusiast, expert).
		WithMaxRounds(4).
		WithMaxHandoffs(3).
		Build()

	// Create executor
	executor := agentic.NewTeamExecutor(team, key)

	// Discussion topics
	topics := []string{
		"Những thực hành tốt nhất khi viết code Go là gì?",
		"Làm thế nào mà các AI agent có thể cải thiện phát triển phần mềm?",
		"Hãy cho tôi biết về những xu hướng mới nhất trong máy học?",
		"Ứng dụng của Go trong các hệ thống distributed có những đặc điểm gì?",
	}

	// Run discussions
	fmt.Println("\n🤖 Hệ Thống Thảo Luận Multi-Agent\n" + strings.Repeat("=", 50))
	for i, topic := range topics {
		fmt.Printf("\n📌 Chủ đề %d: %s\n%s\n", i+1, topic, strings.Repeat("-", 50))
		if response, err := executor.Execute(context.Background(), topic); err == nil {
			fmt.Printf("✅ Kết Quả:\n%s\n", response)
		} else {
			fmt.Printf("❌ Lỗi: %v\n", err)
		}
	}
	fmt.Println("\n" + strings.Repeat("=", 50) + "\n🎉 Hoàn thành!\n")
}

func loadEnv() {
	data, _ := os.ReadFile(".env")
	for _, line := range strings.Split(string(data), "\n") {
		line = strings.TrimSpace(line)
		if line != "" && !strings.HasPrefix(line, "#") {
			if parts := strings.SplitN(line, "=", 2); len(parts) == 2 {
				os.Setenv(strings.TrimSpace(parts[0]), strings.TrimSpace(parts[1]))
			}
		}
	}
}

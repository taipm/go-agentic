package main

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/taipm/go-agentic"
	"gopkg.in/yaml.v3"
)

func main() {
	// Load API key
	loadEnv()
	key := os.Getenv("OPENAI_API_KEY")
	if key == "" {
		fmt.Println("❌ Lỗi: OPENAI_API_KEY không được thiết lập")
		os.Exit(1)
	}

	// Load config from crew.yaml
	var cfg struct {
		Crew struct {
			MaxRounds   int `yaml:"maxRounds"`
			MaxHandoffs int `yaml:"maxHandoffs"`
		} `yaml:"crew"`
		Agents []struct {
			ID          string  `yaml:"id"`
			Name        string  `yaml:"name"`
			Role        string  `yaml:"role"`
			Backstory   string  `yaml:"backstory"`
			Model       string  `yaml:"model"`
			Temperature float64 `yaml:"temperature"`
			IsTerminal  bool    `yaml:"isTerminal"`
		} `yaml:"agents"`
		Topics []string `yaml:"topics"`
	}

	data, _ := os.ReadFile("crew.yaml")
	yaml.Unmarshal(data, &cfg)

	// Create agents from config
	agents := make([]*agentic.Agent, len(cfg.Agents))
	for i, a := range cfg.Agents {
		agents[i] = &agentic.Agent{
			ID: a.ID, Name: a.Name, Role: a.Role, Backstory: a.Backstory,
			Model: a.Model, Temperature: a.Temperature, IsTerminal: a.IsTerminal,
			Tools: []*agentic.Tool{},
		}
	}

	// Create crew and executor
	crew := &agentic.Crew{Agents: agents, MaxRounds: cfg.Crew.MaxRounds, MaxHandoffs: cfg.Crew.MaxHandoffs}
	executor := agentic.NewTeamExecutor(crew, key)

	// Run discussions
	fmt.Println("\n🤖 Hệ Thống Thảo Luận Multi-Agent\n" + strings.Repeat("=", 50))
	for i, topic := range cfg.Topics {
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

package main

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/taipm/go-agentic"
	"gopkg.in/yaml.v3"
)

// Config structure for team.yaml
type Config struct {
	Team struct {
		MaxRounds   int `yaml:"maxRounds"`
		MaxHandoffs int `yaml:"maxHandoffs"`
	} `yaml:"team"`
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

func main() {
	// Load environment and config
	os.Setenv("OPENAI_API_KEY", getEnvVar("OPENAI_API_KEY"))
	if os.Getenv("OPENAI_API_KEY") == "" {
		fmt.Println("❌ Lỗi: OPENAI_API_KEY không được thiết lập")
		os.Exit(1)
	}

	// Parse YAML config
	var cfg Config
	data, _ := os.ReadFile("team.yaml")
	yaml.Unmarshal(data, &cfg)

	// Build agents and routing in one go
	agents := make([]*agentic.Agent, len(cfg.Agents))
	for i, a := range cfg.Agents {
		agents[i] = &agentic.Agent{
			ID: a.ID, Name: a.Name, Role: a.Role, Backstory: a.Backstory,
			Model: a.Model, Temperature: a.Temperature, IsTerminal: a.IsTerminal,
		}
	}

	// Phase 3: Declarative routing with trigger detection
	routing, _ := agentic.NewRouter().
		RegisterAgents("enthusiast", "expert").
		FromAgent("enthusiast").
		To("expert", agentic.NewKeywordDetector([]string{"?", "hỏi", "gì"}, false)).
		Done().
		Build()

	// Create team with routing and run
	team := &agentic.Team{Agents: agents, MaxRounds: cfg.Team.MaxRounds, MaxHandoffs: cfg.Team.MaxHandoffs, Routing: routing}
	executor := agentic.NewTeamExecutor(team, os.Getenv("OPENAI_API_KEY"))

	fmt.Println("\n🤖 Hệ Thống Thảo Luận Multi-Agent\n" + strings.Repeat("=", 50))
	for i, topic := range cfg.Topics {
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

// getEnvVar reads from .env file
func getEnvVar(key string) string {
	data, _ := os.ReadFile(".env")
	for _, line := range strings.Split(string(data), "\n") {
		if strings.HasPrefix(strings.TrimSpace(line), key+"=") {
			return strings.TrimSpace(strings.Split(line, "=")[1])
		}
	}
	return ""
}

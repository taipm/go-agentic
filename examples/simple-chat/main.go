package main

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/taipm/go-agentic"
	"gopkg.in/yaml.v3"
)

// Config represents the crew configuration from YAML
type Config struct {
	Crew struct {
		MaxRounds   int `yaml:"maxRounds"`
		MaxHandoffs int `yaml:"maxHandoffs"`
	} `yaml:"crew"`
	Agents []AgentConfig `yaml:"agents"`
	Topics []string      `yaml:"topics"`
}

// AgentConfig represents a single agent configuration
type AgentConfig struct {
	ID          string  `yaml:"id"`
	Name        string  `yaml:"name"`
	Role        string  `yaml:"role"`
	Backstory   string  `yaml:"backstory"`
	Model       string  `yaml:"model"`
	Temperature float64 `yaml:"temperature"`
	IsTerminal  bool    `yaml:"isTerminal"`
}

func main() {
	// Load environment
	if err := loadEnvFile(); err != nil {
		fmt.Printf("📝 Ghi chú: %v\n", err)
	}

	apiKey := os.Getenv("OPENAI_API_KEY")
	if apiKey == "" {
		fmt.Println("❌ Lỗi: OPENAI_API_KEY chưa được thiết lập")
		fmt.Println("📝 Vui lòng thêm API key của bạn vào file .env hoặc đặt biến môi trường")
		os.Exit(1)
	}

	// Load configuration from YAML
	config, err := loadConfig("crew.yaml")
	if err != nil {
		fmt.Printf("❌ Lỗi tải config: %v\n", err)
		os.Exit(1)
	}

	// Create crew from configuration
	crew := createCrewFromConfig(config)
	executor := agentic.NewTeamExecutor(crew, apiKey)

	// Display header
	fmt.Println()
	fmt.Println("🤖 Hệ Thống Thảo Luận Multi-Agent Đơn Giản")
	fmt.Println("=" + strings.Repeat("=", 48))
	fmt.Println()

	// Run discussions on each topic
	for i, topic := range config.Topics {
		fmt.Printf("📌 Chủ đề %d: %s\n", i+1, topic)
		fmt.Println("-" + strings.Repeat("-", 48))

		ctx := context.Background()
		response, err := executor.Execute(ctx, topic)
		if err != nil {
			fmt.Printf("❌ Lỗi: %v\n", err)
			continue
		}

		fmt.Printf("✅ Kết Quả Cuối Cùng:\n%s\n\n", response)
	}

	fmt.Println("=" + strings.Repeat("=", 48))
	fmt.Println("🎉 Hoàn thành tất cả các cuộc thảo luận!")
	fmt.Println()
}

// loadConfig loads the crew configuration from YAML file
func loadConfig(filename string) (*Config, error) {
	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("không thể đọc file config %s: %w", filename, err)
	}

	var config Config
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("lỗi parse YAML: %w", err)
	}

	return &config, nil
}

// createCrewFromConfig creates a crew from the YAML configuration
func createCrewFromConfig(config *Config) *agentic.Crew {
	agents := make([]*agentic.Agent, len(config.Agents))

	for i, agentCfg := range config.Agents {
		agents[i] = &agentic.Agent{
			ID:          agentCfg.ID,
			Name:        agentCfg.Name,
			Role:        agentCfg.Role,
			Backstory:   agentCfg.Backstory,
			Model:       agentCfg.Model,
			Tools:       []*agentic.Tool{},
			Temperature: agentCfg.Temperature,
			IsTerminal:  agentCfg.IsTerminal,
		}
	}

	crew := &agentic.Crew{
		Agents:      agents,
		MaxRounds:   config.Crew.MaxRounds,
		MaxHandoffs: config.Crew.MaxHandoffs,
	}

	return crew
}

// loadEnvFile loads environment variables from .env file
func loadEnvFile() error {
	data, err := os.ReadFile(".env")
	if err != nil {
		return fmt.Errorf("không tìm thấy file .env - sử dụng biến môi trường của hệ thống")
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

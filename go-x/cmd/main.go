package main

import (
	"bufio"
	"context"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/taipm/go-crewai"
)

// loadEnvFile loads environment variables from .env file
func loadEnvFile() error {
	// Try to find .env file in current directory or parent directories
	paths := []string{
		".env",
		filepath.Join("..", ".env"),
		filepath.Join(filepath.Dir(os.Args[0]), ".env"),
		filepath.Join(filepath.Dir(os.Args[0]), "..", ".env"),
	}

	for _, path := range paths {
		if data, err := os.ReadFile(path); err == nil {
			lines := strings.Split(string(data), "\n")
			for _, line := range lines {
				line = strings.TrimSpace(line)
				// Skip comments and empty lines
				if line == "" || strings.HasPrefix(line, "#") {
					continue
				}

				// Parse key=value
				parts := strings.SplitN(line, "=", 2)
				if len(parts) == 2 {
					key := strings.TrimSpace(parts[0])
					value := strings.TrimSpace(parts[1])
					os.Setenv(key, value)
				}
			}
			return nil
		}
	}
	return fmt.Errorf("no .env file found")
}

func getConfigDir() string {
	if configDir := os.Getenv("CREWAI_CONFIG_DIR"); configDir != "" {
		return configDir
	}

	// Try current working directory first
	if _, err := os.Stat("config"); err == nil {
		return "config"
	}

	// Try relative to executable
	if ex, err := os.Executable(); err == nil {
		relPath := filepath.Join(filepath.Dir(ex), "..", "config")
		if _, err := os.Stat(relPath); err == nil {
			return relPath
		}
	}

	return "config"
}

func getDefaultCrewConfig() *crewai.CrewConfig {
	return &crewai.CrewConfig{
		EntryPoint: "orchestrator",
		Settings: struct {
			MaxHandoffs    int    `yaml:"max_handoffs"`
			MaxRounds      int    `yaml:"max_rounds"`
			TimeoutSeconds int    `yaml:"timeout_seconds"`
			Language       string `yaml:"language"`
			Organization   string `yaml:"organization"`
		}{
			MaxHandoffs:    5,
			MaxRounds:      10,
			TimeoutSeconds: 300,
			Language:       "en",
			Organization:   "IT-Support",
		},
	}
}

func loadCrewConfig(configDir string) *crewai.CrewConfig {
	crewConfigPath := filepath.Join(configDir, "crew.yaml")
	crewConfig, err := crewai.LoadCrewConfig(crewConfigPath)
	if err != nil {
		fmt.Printf("Error loading crew config: %v\n", err)
		fmt.Println("Creating default crew...")
		return getDefaultCrewConfig()
	}
	return crewConfig
}

func createAgentsFromConfig(crewConfig *crewai.CrewConfig, agentConfigs map[string]*crewai.AgentConfig, allTools map[string]*crewai.Tool) []*crewai.Agent {
	var agents []*crewai.Agent
	for _, agentID := range crewConfig.Agents {
		if config, exists := agentConfigs[agentID]; exists {
			agent := crewai.CreateAgentFromConfig(config, allTools)
			agents = append(agents, agent)
		}
	}
	return agents
}

func runInteractiveLoop(executor *crewai.CrewExecutor) {
	fmt.Println("\n🚀 go-crewai IT Support Crew v1.0")
	fmt.Println("======================================")
	fmt.Println("Enter your IT support request (type 'quit' to exit):")
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
			fmt.Println("Goodbye!")
			break
		}

		response, err := executor.Execute(context.Background(), input)
		if err != nil {
			fmt.Printf("Error: %v\n\n", err)
			continue
		}

		fmt.Printf("\n✅ Final Response:\n")
		fmt.Printf("Agent: %s\n", response.AgentName)
		fmt.Printf("Response: %s\n\n", response.Content)

		if response.IsTerminal {
			fmt.Println("[Conversation ended - terminal agent reached]\n")
		}
	}
}

func main() {
	// Parse command line flags
	serverMode := flag.Bool("server", false, "Run in HTTP server mode (SSE streaming)")
	port := flag.Int("port", 8080, "HTTP server port (only in server mode)")
	flag.Parse()

	// Load .env file if it exists
	if err := loadEnvFile(); err != nil {
		fmt.Printf("Note: %v (using environment variables if set)\n", err)
	}

	// Get API key from environment
	apiKey := os.Getenv("OPENAI_API_KEY")
	if apiKey == "" {
		fmt.Println("Error: OPENAI_API_KEY environment variable not set")
		fmt.Println("Please create a .env file with: OPENAI_API_KEY=sk-...")
		os.Exit(1)
	}

	// Load configuration
	configDir := getConfigDir()
	crewConfig := loadCrewConfig(configDir)

	agentsDir := filepath.Join(configDir, "agents")
	agentConfigs, err := crewai.LoadAgentConfigs(agentsDir)
	if err != nil {
		fmt.Printf("Error loading agent configs: %v\n", err)
		os.Exit(1)
	}

	// Create agents
	allTools := crewai.GetAllITSupportTools()
	agents := createAgentsFromConfig(crewConfig, agentConfigs, allTools)

	if len(agents) == 0 {
		fmt.Println("Error: No agents loaded from configuration")
		os.Exit(1)
	}

	// Check if test mode is requested
	args := flag.Args()
	if len(args) > 0 && args[0] == "test" {
		fmt.Println("Test mode not yet implemented")
		return
	}

	// Create crew and executor
	crew := &crewai.Crew{
		Agents:      agents,
		MaxRounds:   crewConfig.Settings.MaxRounds,
		MaxHandoffs: crewConfig.Settings.MaxHandoffs,
		Routing:     crewConfig.Routing,
	}
	executor := crewai.NewCrewExecutor(crew, apiKey)

	// Run in requested mode
	if *serverMode {
		// HTTP server mode (SSE streaming)
		if err := crewai.StartHTTPServer(executor, *port); err != nil {
			fmt.Printf("HTTP server error: %v\n", err)
			os.Exit(1)
		}
	} else {
		// CLI interactive mode (default)
		runInteractiveLoop(executor)
	}
}

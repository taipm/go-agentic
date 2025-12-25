package main

import (
	"fmt"
	"os"

	"github.com/taipm/go-agentic/core/config"
)

func main() {
	// Get API key
	apiKey := os.Getenv("OPENAI_API_KEY")
	if apiKey == "" {
		apiKey = "ollama"
		fmt.Println("ℹ️  Using Ollama (local)")
	}

	// Load agent config directly to access it
	agentConfig, err := config.LoadAgentConfig("config/agents/hello-agent.yaml")
	if err != nil {
		fmt.Printf("Error loading agent config: %v\n", err)
		os.Exit(1)
	}

	// Create agent from config
	agent := config.CreateAgentFromConfig(agentConfig, map[string]interface{}{})

	// Print agent metadata
	fmt.Printf("\n╔════════════════════════════════════════════════════════════╗\n")
	fmt.Printf("║              AGENT METADATA INSPECTION                    ║\n")
	fmt.Printf("╚════════════════════════════════════════════════════════════╝\n\n")

	fmt.Printf("📋 Agent Information:\n")
	fmt.Printf("  ID: %s\n", agent.ID)
	fmt.Printf("  Name: %s\n", agent.Name)
	fmt.Printf("  Role: %s\n", agent.Role)
	fmt.Printf("  Backstory: %s\n", agent.Backstory)

	fmt.Printf("\n💰 Cost Configuration:\n")
	if agent.Quota != nil {
		fmt.Printf("  Quotas:\n")
		fmt.Printf("    - MaxTokensPerCall: %d\n", agent.Quota.MaxTokensPerCall)
		fmt.Printf("    - MaxTokensPerDay: %d\n", agent.Quota.MaxTokensPerDay)
		fmt.Printf("    - MaxCostPerDay: $%.2f\n", agent.Quota.MaxCostPerDay)
		fmt.Printf("    - BlockOnQuotaExceed: %v\n", agent.Quota.BlockOnQuotaExceed)
	}

	if agent.CostMetrics != nil {
		fmt.Printf("\n  Current Metrics:\n")
		agent.CostMetrics.Mutex.RLock()
		fmt.Printf("    - CallCount: %d\n", agent.CostMetrics.CallCount)
		fmt.Printf("    - TotalTokens: %d\n", agent.CostMetrics.TotalTokens)
		fmt.Printf("    - DailyCost: $%.6f\n", agent.CostMetrics.DailyCost)
		agent.CostMetrics.Mutex.RUnlock()
	}

	fmt.Printf("\n🧠 Memory Configuration:\n")
	if agent.MemoryMetrics != nil {
		agent.MemoryMetrics.Mutex.RLock()
		fmt.Printf("  Current Metrics:\n")
		fmt.Printf("    - CurrentMemoryMB: %d\n", agent.MemoryMetrics.CurrentMemoryMB)
		fmt.Printf("    - PeakMemoryMB: %d\n", agent.MemoryMetrics.PeakMemoryMB)
		fmt.Printf("    - AverageMemoryMB: %d\n", agent.MemoryMetrics.AverageMemoryMB)
		fmt.Printf("    - MaxMemoryMB: %d\n", agent.MemoryMetrics.MaxMemoryMB)
		fmt.Printf("    - MaxContextWindow: %d tokens\n", agent.MemoryMetrics.MaxContextWindow)
		agent.MemoryMetrics.Mutex.RUnlock()
	}

	fmt.Printf("\n📊 Performance Metrics:\n")
	if agent.PerformanceMetrics != nil {
		agent.PerformanceMetrics.Mutex.RLock()
		fmt.Printf("  Quality:\n")
		fmt.Printf("    - SuccessfulCalls: %d\n", agent.PerformanceMetrics.SuccessfulCalls)
		fmt.Printf("    - FailedCalls: %d\n", agent.PerformanceMetrics.FailedCalls)
		fmt.Printf("    - SuccessRate: %.1f%%\n", agent.PerformanceMetrics.SuccessRate)
		fmt.Printf("    - MaxErrorsPerDay: %d\n", agent.PerformanceMetrics.MaxErrorsPerDay)
		agent.PerformanceMetrics.Mutex.RUnlock()
	}

	fmt.Printf("\n✅ Agent loaded successfully from config!\n")
	fmt.Printf("   (Metrics will be updated when agent is executed)\n\n")
}

package main

import (
	"fmt"
	"os"

	agenticcore "github.com/taipm/go-agentic/core"
)

func main() {
	// Get API key
	apiKey := os.Getenv("OPENAI_API_KEY")
	if apiKey == "" {
		apiKey = "ollama"
		fmt.Println("ℹ️  Using Ollama (local)")
	}

	// Load agent config directly to access it
	// ✅ FIX for Issue #5: Pass configMode (default to PERMISSIVE for backward compatibility)
	agentConfig, err := agenticcore.LoadAgentConfig("config/agents/hello-agent.yaml", agenticcore.PermissiveMode)
	if err != nil {
		fmt.Printf("Error loading agent config: %v\n", err)
		os.Exit(1)
	}

	// Create agent from config
	agent := agenticcore.CreateAgentFromConfig(agentConfig, map[string]*agenticcore.Tool{})

	// Print agent metadata
	fmt.Printf("\n╔════════════════════════════════════════════════════════════╗\n")
	fmt.Printf("║              AGENT METADATA INSPECTION                    ║\n")
	fmt.Printf("╚════════════════════════════════════════════════════════════╝\n\n")

	fmt.Printf("📋 Agent Information:\n")
	fmt.Printf("  ID: %s\n", agent.ID)
	fmt.Printf("  Name: %s\n", agent.Name)
	fmt.Printf("  Role: %s\n", agent.Role)

	// Access metadata safely with mutex
	agent.Metadata.Mutex.RLock()
	defer agent.Metadata.Mutex.RUnlock()

	fmt.Printf("\n💰 Cost Configuration & Metrics:\n")
	fmt.Printf("  Quotas:\n")
	fmt.Printf("    - MaxTokensPerCall: %d\n", agent.Metadata.Quotas.MaxTokensPerCall)
	fmt.Printf("    - MaxTokensPerDay: %d\n", agent.Metadata.Quotas.MaxTokensPerDay)
	fmt.Printf("    - MaxCostPerDay: $%.2f\n", agent.Metadata.Quotas.MaxCostPerDay)
	fmt.Printf("    - CostAlertPercent: %.0f%%\n", agent.Metadata.Quotas.CostAlertPercent*100)
	fmt.Printf("    - BlockOnQuotaExceed: %v\n", agent.Metadata.Quotas.BlockOnQuotaExceed)

	fmt.Printf("\n  Current Metrics:\n")
	fmt.Printf("    - CallCount: %d\n", agent.Metadata.Cost.CallCount)
	fmt.Printf("    - TotalTokens: %d\n", agent.Metadata.Cost.TotalTokens)
	fmt.Printf("    - DailyCost: $%.6f\n", agent.Metadata.Cost.DailyCost)

	fmt.Printf("\n🧠 Memory Configuration & Metrics:\n")
	fmt.Printf("  Quotas:\n")
	fmt.Printf("    - MaxMemoryPerCall: %d MB\n", agent.Metadata.Quotas.MaxMemoryPerCall)
	fmt.Printf("    - MaxMemoryPerDay: %d MB\n", agent.Metadata.Quotas.MaxMemoryPerDay)
	fmt.Printf("    - MaxContextWindow: %d tokens\n", agent.Metadata.Quotas.MaxContextWindow)

	fmt.Printf("\n  Current Metrics:\n")
	fmt.Printf("    - CurrentMemoryMB: %d\n", agent.Metadata.Memory.CurrentMemoryMB)
	fmt.Printf("    - PeakMemoryMB: %d\n", agent.Metadata.Memory.PeakMemoryMB)
	fmt.Printf("    - AverageMemoryMB: %d\n", agent.Metadata.Memory.AverageMemoryMB)
	fmt.Printf("    - CurrentContextSize: %d tokens\n", agent.Metadata.Memory.CurrentContextSize)
	fmt.Printf("    - MaxContextWindow: %d tokens\n", agent.Metadata.Memory.MaxContextWindow)

	fmt.Printf("\n⚙️  Execution Quotas:\n")
	fmt.Printf("    - MaxCallsPerMinute: %d\n", agent.Metadata.Quotas.MaxCallsPerMinute)
	fmt.Printf("    - MaxCallsPerHour: %d\n", agent.Metadata.Quotas.MaxCallsPerHour)
	fmt.Printf("    - MaxCallsPerDay: %d\n", agent.Metadata.Quotas.MaxCallsPerDay)
	fmt.Printf("    - MaxErrorsPerHour: %d\n", agent.Metadata.Quotas.MaxErrorsPerHour)
	fmt.Printf("    - MaxErrorsPerDay: %d\n", agent.Metadata.Quotas.MaxErrorsPerDay)

	fmt.Printf("\n📊 Performance Metrics:\n")
	fmt.Printf("  Quality:\n")
	fmt.Printf("    - SuccessfulCalls: %d\n", agent.Metadata.Performance.SuccessfulCalls)
	fmt.Printf("    - FailedCalls: %d\n", agent.Metadata.Performance.FailedCalls)
	fmt.Printf("    - SuccessRate: %.1f%%\n", agent.Metadata.Performance.SuccessRate)

	fmt.Printf("\n  Error Tracking:\n")
	fmt.Printf("    - ConsecutiveErrors: %d\n", agent.Metadata.Performance.ConsecutiveErrors)
	fmt.Printf("    - ErrorCountToday: %d\n", agent.Metadata.Performance.ErrorCountToday)
	fmt.Printf("    - MaxErrorsPerDay: %d\n", agent.Metadata.Performance.MaxErrorsPerDay)

	fmt.Printf("\n⏱️  Timestamps:\n")
	fmt.Printf("    - Created: %s\n", agent.Metadata.CreatedTime.Format("2006-01-02 15:04:05"))
	fmt.Printf("    - LastAccess: %s\n", agent.Metadata.LastAccessTime.Format("2006-01-02 15:04:05"))

	fmt.Printf("\n✅ Metadata inspection complete!\n")
	fmt.Printf("   (Metrics will be updated when agent is executed)\n\n")
}

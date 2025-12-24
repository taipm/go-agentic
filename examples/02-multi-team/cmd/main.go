package main

import (
	"bufio"
	"context"
	"flag"
	"fmt"
	"os"
	"strings"
	"time"

	agenticcore "github.com/taipm/go-agentic/core"
)

// TeamExecutor wraps a CrewExecutor for a specific team
type TeamExecutor struct {
	Name     string
	Executor *agenticcore.CrewExecutor
}

// MultiTeamOrchestrator coordinates multiple teams
type MultiTeamOrchestrator struct {
	MasterTeam *TeamExecutor
	TeamAlpha  *TeamExecutor
	TeamBeta   *TeamExecutor
	Verbose    bool
}

func main() {
	verbose := flag.Bool("v", false, "Verbose mode")
	flag.Parse()

	printBanner()

	// Get API key
	apiKey := os.Getenv("OPENAI_API_KEY")
	if apiKey == "" {
		apiKey = "ollama"
		fmt.Println("ℹ️  Using Ollama (local)")
	}

	// Create orchestrator
	orchestrator, err := NewMultiTeamOrchestrator(apiKey, *verbose)
	if err != nil {
		fmt.Printf("❌ Error creating orchestrator: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("\nNhập topic để bắt đầu (hoặc 'exit' để thoát):")
	fmt.Println("Ví dụ: 'Viết về lợi ích của AI trong giáo dục'")
	fmt.Println()

	scanner := bufio.NewScanner(os.Stdin)
	for {
		fmt.Print("📝 Topic: ")
		if !scanner.Scan() {
			break
		}

		input := strings.TrimSpace(scanner.Text())
		if strings.ToLower(input) == "exit" {
			fmt.Println("👋 Goodbye!")
			break
		}

		if input == "" {
			continue
		}

		// Run the multi-team workflow
		result, err := orchestrator.Execute(context.Background(), input)
		if err != nil {
			fmt.Printf("❌ Error: %v\n", err)
			continue
		}

		fmt.Println("\n" + strings.Repeat("═", 60))
		fmt.Println("📋 KẾT QUẢ CUỐI CÙNG:")
		fmt.Println(strings.Repeat("═", 60))
		fmt.Println(result)
		fmt.Println()
	}
}

func printBanner() {
	fmt.Println("╔════════════════════════════════════════════════════════════╗")
	fmt.Println("║           MULTI-TEAM COORDINATION DEMO                    ║")
	fmt.Println("╠════════════════════════════════════════════════════════════╣")
	fmt.Println("║  Cấu trúc:                                                 ║")
	fmt.Println("║  ┌─────────────────────────────────────────────────────┐  ║")
	fmt.Println("║  │  🎯 Master Team: Coordinator                        │  ║")
	fmt.Println("║  │       ├── 🔬 Team Alpha: Researcher → Analyst       │  ║")
	fmt.Println("║  │       └── ✍️  Team Beta:  Writer → Reviewer         │  ║")
	fmt.Println("║  └─────────────────────────────────────────────────────┘  ║")
	fmt.Println("╚════════════════════════════════════════════════════════════╝")
}

// NewMultiTeamOrchestrator creates orchestrator with all teams
func NewMultiTeamOrchestrator(apiKey string, verbose bool) (*MultiTeamOrchestrator, error) {
	// Load Master Team
	masterExec, err := agenticcore.NewCrewExecutorFromConfig(apiKey, "master-team/config", nil)
	if err != nil {
		return nil, fmt.Errorf("failed to load master-team: %w", err)
	}
	masterExec.SetVerbose(verbose)

	// Load Team Alpha
	alphaExec, err := agenticcore.NewCrewExecutorFromConfig(apiKey, "team-alpha/config", nil)
	if err != nil {
		return nil, fmt.Errorf("failed to load team-alpha: %w", err)
	}
	alphaExec.SetVerbose(verbose)

	// Load Team Beta
	betaExec, err := agenticcore.NewCrewExecutorFromConfig(apiKey, "team-beta/config", nil)
	if err != nil {
		return nil, fmt.Errorf("failed to load team-beta: %w", err)
	}
	betaExec.SetVerbose(verbose)

	return &MultiTeamOrchestrator{
		MasterTeam: &TeamExecutor{Name: "Master Team", Executor: masterExec},
		TeamAlpha:  &TeamExecutor{Name: "Team Alpha", Executor: alphaExec},
		TeamBeta:   &TeamExecutor{Name: "Team Beta", Executor: betaExec},
		Verbose:    verbose,
	}, nil
}

// Execute runs the multi-team workflow
func (o *MultiTeamOrchestrator) Execute(ctx context.Context, input string) (string, error) {
	fmt.Println("\n🚀 Bắt đầu workflow...")
	fmt.Println(strings.Repeat("─", 60))

	// Phase 1: Master Team receives input and delegates
	fmt.Println("\n🎯 [MASTER TEAM] Đang phân tích yêu cầu...")
	masterResult, err := o.executeTeam(ctx, o.MasterTeam, input)
	if err != nil {
		return "", fmt.Errorf("master team error: %w", err)
	}
	fmt.Printf("   Response: %s\n", truncate(masterResult, 200))

	// Check if Master wants to delegate to Alpha
	if strings.Contains(masterResult, "[DELEGATE_ALPHA]") {
		// Phase 2: Team Alpha researches
		fmt.Println("\n🔬 [TEAM ALPHA] Đang nghiên cứu...")
		alphaResult, err := o.executeTeam(ctx, o.TeamAlpha, input)
		if err != nil {
			return "", fmt.Errorf("team alpha error: %w", err)
		}
		fmt.Printf("   Response: %s\n", truncate(alphaResult, 200))

		// Phase 3: Send Alpha result back to Master
		fmt.Println("\n🎯 [MASTER TEAM] Đã nhận kết quả từ Team Alpha...")
		combinedInput := fmt.Sprintf("Kết quả từ Team Alpha:\n%s\n\nTopic gốc: %s", alphaResult, input)
		masterResult2, err := o.executeTeam(ctx, o.MasterTeam, combinedInput)
		if err != nil {
			return "", fmt.Errorf("master team phase 2 error: %w", err)
		}
		fmt.Printf("   Response: %s\n", truncate(masterResult2, 200))

		// Check if Master wants to delegate to Beta
		if strings.Contains(masterResult2, "[DELEGATE_BETA]") {
			// Phase 4: Team Beta writes content
			fmt.Println("\n✍️  [TEAM BETA] Đang viết content...")
			betaBrief := fmt.Sprintf("Brief từ Master:\n%s\n\nInsights từ Team Alpha:\n%s", masterResult2, alphaResult)
			betaResult, err := o.executeTeam(ctx, o.TeamBeta, betaBrief)
			if err != nil {
				return "", fmt.Errorf("team beta error: %w", err)
			}
			fmt.Printf("   Response: %s\n", truncate(betaResult, 200))

			// Phase 5: Final report from Master
			fmt.Println("\n🎯 [MASTER TEAM] Đang tổng hợp kết quả cuối...")
			finalInput := fmt.Sprintf("Kết quả từ Team Beta:\n%s\n\nTopic gốc: %s", betaResult, input)
			finalResult, err := o.executeTeam(ctx, o.MasterTeam, finalInput)
			if err != nil {
				return "", fmt.Errorf("master team final error: %w", err)
			}

			return finalResult, nil
		}

		return masterResult2, nil
	}

	return masterResult, nil
}

// executeTeam runs a single team and returns result
func (o *MultiTeamOrchestrator) executeTeam(ctx context.Context, team *TeamExecutor, input string) (string, error) {
	// Create timeout context
	ctx, cancel := context.WithTimeout(ctx, 60*time.Second)
	defer cancel()

	// Use non-streaming Execute for simplicity
	result, err := team.Executor.Execute(ctx, input)
	if err != nil {
		return "", err
	}

	// Clear history for next call
	team.Executor.ClearHistory()

	return result.Content, nil
}

// truncate shortens string for display
func truncate(s string, maxLen int) string {
	s = strings.ReplaceAll(s, "\n", " ")
	if len(s) > maxLen {
		return s[:maxLen] + "..."
	}
	return s
}

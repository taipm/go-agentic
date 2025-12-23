package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/joho/godotenv"
	agenticcore "github.com/taipm/go-agentic/core"
	"github.com/taipm/go-agentic/examples/01-quiz-exam/internal"
)

func main() {
	// Load .env file (optional)
	_ = godotenv.Load()

	verbose := flag.Bool("verbose", false, "Enable verbose output")
	outputDir := flag.String("output", "reports", "Output directory for exam reports")
	flag.Parse()

	// Get API key
	apiKey := os.Getenv("OPENAI_API_KEY")
	if apiKey == "" {
		apiKey = "ollama"
		fmt.Println("Using Ollama (local) - no API key needed")
	}

	// Create output directory if not exists
	if err := os.MkdirAll(*outputDir, 0755); err != nil {
		fmt.Printf("Error creating output directory: %v\n", err)
		os.Exit(1)
	}

	// Create quiz state (shared memory for the exam)
	quizState := internal.NewQuizState(*outputDir)

	// Create tools
	tools := internal.CreateQuizTools(quizState)

	// Create executor from config
	executor, err := agenticcore.NewCrewExecutorFromConfig(apiKey, "config", tools)
	if err != nil {
		fmt.Printf("Error creating executor: %v\n", err)
		os.Exit(1)
	}

	executor.SetVerbose(*verbose)

	// Print banner
	printBanner(quizState.ReportPath)

	// Run the quiz exam
	runQuizExam(executor, quizState)
}

func printBanner(reportPath string) {
	fmt.Println()
	fmt.Println("╔══════════════════════════════════════════════════════════╗")
	fmt.Println("║           THI VẤN ĐÁP - QUIZ EXAM DEMO                   ║")
	fmt.Println("╠══════════════════════════════════════════════════════════╣")
	fmt.Println("║  - 2 Agents: Thầy Giáo (Teacher) + Học Sinh (Student)    ║")
	fmt.Println("║  - 10 câu hỏi, mỗi câu đúng = 1 điểm                     ║")
	fmt.Println("║  - Điểm > 5 = ĐẠT, <= 5 = CHƯA ĐẠT                       ║")
	fmt.Println("║  - Demo: Signal-based routing & Tool memory              ║")
	fmt.Println("╠══════════════════════════════════════════════════════════╣")
	fmt.Printf("║  Biên bản: %-47s ║\n", reportPath)
	fmt.Println("╚══════════════════════════════════════════════════════════╝")
	fmt.Println()
}

func runQuizExam(executor *agenticcore.CrewExecutor, quizState *internal.QuizState) {
	ctx := context.Background()

	// Create stream channel for events
	streamChan := make(chan *agenticcore.StreamEvent, 100)

	// Start the exam with an initial prompt
	initialPrompt := "Bắt đầu kỳ thi vấn đáp. Thầy giáo hãy đặt câu hỏi đầu tiên cho học sinh."

	fmt.Println("Bắt đầu kỳ thi...")
	fmt.Println(strings.Repeat("─", 60))
	fmt.Println()

	// Run in a goroutine to collect events
	go func() {
		err := executor.ExecuteStream(ctx, initialPrompt, streamChan)
		if err != nil {
			fmt.Printf("\n[ERROR] %v\n", err)
		}
		close(streamChan)
	}()

	// Process events
	currentAgent := ""
	for event := range streamChan {
		switch event.Type {
		case "agent_start":
			if event.Agent != currentAgent {
				currentAgent = event.Agent
				fmt.Printf("\n[%s]\n", event.Agent)
				fmt.Println(strings.Repeat("─", 40))
			}

		case "agent_response":
			// Print agent's response (cleaned up)
			content := cleanResponse(event.Content)
			if content != "" {
				fmt.Printf("%s\n", content)
			}

		case "tool_start":
			// Tool execution happening
			fmt.Printf("  %s\n", event.Content)

		case "tool_result":
			// Tool result
			// Already printed by the tool itself with formatting

		case "error":
			fmt.Printf("\n[LỖI] %s: %s\n", event.Agent, event.Content)

		case "pause":
			fmt.Printf("\n[TẠM DỪNG] %s\n", event.Content)
		}
	}

	// Print final summary
	fmt.Println()
	fmt.Println(strings.Repeat("═", 60))
	fmt.Println("KẾT THÚC KỲ THI")
	fmt.Println(strings.Repeat("═", 60))

	// Get final result from quiz state
	result := quizState.GetFinalResult()
	fmt.Printf("\nĐiểm số: %d/%d (%s)\n", result["final_score"], result["max_score"], result["percentage"])
	fmt.Printf("Kết quả: %s\n", result["grade"])
	fmt.Printf("\nBiên bản thi: %s\n", quizState.ReportPath)
	fmt.Println()
}

// cleanResponse removes signal markers from response for cleaner output
func cleanResponse(content string) string {
	// Remove common signal markers
	signals := []string{
		"[CÂU HỎI]",
		"[TRẢ LỜI]",
		"[KẾT THÚC THI]",
	}

	result := content
	for _, signal := range signals {
		result = strings.ReplaceAll(result, signal, "")
	}

	return strings.TrimSpace(result)
}

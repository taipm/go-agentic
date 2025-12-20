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
		fmt.Println("❌ Error: OPENAI_API_KEY not set")
		os.Exit(1)
	}

	// Load team from YAML with tool handlers
	team, err := agentic.LoadTeamFromYAML("team.yaml", getToolHandlers())
	if err != nil {
		fmt.Printf("❌ Failed to load team config: %v\n", err)
		os.Exit(1)
	}

	// Create executor
	executor := agentic.NewTeamExecutor(team, apiKey)

	// Support tickets to process
	tickets := []struct {
		id    string
		issue string
	}{
		{
			id:    "TKT-001",
			issue: "Computer won't turn on. No lights, no fans, completely dead.",
		},
		{
			id:    "TKT-002",
			issue: "Microsoft Office installation failed with error 0x80070005. I don't have admin rights.",
		},
		{
			id:    "TKT-003",
			issue: "Cannot connect to company VPN. Getting timeout errors.",
		},
	}

	// Process tickets
	fmt.Println("\n🎫 IT Support System\n" + strings.Repeat("=", 50))
	for _, ticket := range tickets {
		fmt.Printf("\n📋 Ticket: %s\n%s\n", ticket.id, strings.Repeat("-", 50))

		prompt := fmt.Sprintf("[Ticket %s] %s", ticket.id, ticket.issue)
		response, err := executor.Execute(context.Background(), prompt)

		if err == nil {
			fmt.Printf("✅ %s\n", response.Content)
		} else {
			fmt.Printf("❌ Error: %v\n", err)
		}
	}

	fmt.Println("\n" + strings.Repeat("=", 50) + "\n🎉 Done!\n")
}

func getToolHandlers() agentic.ToolHandlerRegistry {
	return agentic.ToolHandlerRegistry{
		"get_system_info": func(ctx context.Context, args map[string]interface{}) (string, error) {
			return "System: Intel i7-13700K | Memory: 32GB | Disk: 512GB SSD | OS: Ubuntu 22.04", nil
		},
		"ping_host": func(ctx context.Context, args map[string]interface{}) (string, error) {
			target, _ := args["target"].(string)
			return fmt.Sprintf("✓ %s is reachable (15ms latency)", target), nil
		},
		"check_disk_space": func(ctx context.Context, args map[string]interface{}) (string, error) {
			return "Disk: 450GB / 512GB available (88% used)", nil
		},
		"list_installed_apps": func(ctx context.Context, args map[string]interface{}) (string, error) {
			return "Firefox 121 | Chrome 121 | VSCode 1.84 | Git 2.42", nil
		},
		"check_app_version": func(ctx context.Context, args map[string]interface{}) (string, error) {
			return "Version: 1.2.3 (latest)", nil
		},
		"verify_licenses": func(ctx context.Context, args map[string]interface{}) (string, error) {
			return "Licensed (Valid until 2025-12-31)", nil
		},
		"trace_route": func(ctx context.Context, args map[string]interface{}) (string, error) {
			return "Route: 1→router 2→ISP 3→CDN (3 hops, healthy)", nil
		},
		"check_dns": func(ctx context.Context, args map[string]interface{}) (string, error) {
			return "DNS: 8.8.8.8 | TTL: 300s | Status: ✓ Healthy", nil
		},
		"log_ticket": func(ctx context.Context, args map[string]interface{}) (string, error) {
			return "✓ Ticket logged and archived", nil
		},
		"send_summary": func(ctx context.Context, args map[string]interface{}) (string, error) {
			return "✓ Summary sent via email", nil
		},
	}
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

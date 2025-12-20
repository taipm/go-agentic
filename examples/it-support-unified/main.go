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
		fmt.Println("❌ Error: OPENAI_API_KEY not set")
		os.Exit(1)
	}

	// Define tool handlers - actual implementations
	toolHandlers := agentic.ToolHandlerRegistry{
		"get_system_info": func(ctx context.Context, args map[string]interface{}) (string, error) {
			return "System: Intel i7-13700K | Memory: 32GB | Disk: 512GB SSD | OS: Ubuntu 22.04", nil
		},
		"ping_host": func(ctx context.Context, args map[string]interface{}) (string, error) {
			target, _ := args["target"].(string)
			return fmt.Sprintf("✓ %s is reachable (latency: 15ms)", target), nil
		},
		"check_disk_space": func(ctx context.Context, args map[string]interface{}) (string, error) {
			return "Disk: 450GB available / 512GB total (88% used)", nil
		},
		"list_installed_apps": func(ctx context.Context, args map[string]interface{}) (string, error) {
			return "Firefox 121.0 | Chrome 121.0 | VSCode 1.84.0 | Git 2.42.0", nil
		},
		"check_app_version": func(ctx context.Context, args map[string]interface{}) (string, error) {
			app, _ := args["app_name"].(string)
			return fmt.Sprintf("%s version: 1.2.3 (latest)", app), nil
		},
		"verify_licenses": func(ctx context.Context, args map[string]interface{}) (string, error) {
			app, _ := args["app_name"].(string)
			return fmt.Sprintf("%s: Licensed (Valid until 2025-12-31)", app), nil
		},
		"trace_route": func(ctx context.Context, args map[string]interface{}) (string, error) {
			dest, _ := args["destination"].(string)
			return fmt.Sprintf("Trace to %s: 1->router 2->ISP 3->CDN (3 hops, healthy)", dest), nil
		},
		"check_dns": func(ctx context.Context, args map[string]interface{}) (string, error) {
			domain, _ := args["domain"].(string)
			return fmt.Sprintf("DNS %s: 8.8.8.8 | TTL: 300s | Status: ✓ Healthy", domain), nil
		},
		"log_ticket": func(ctx context.Context, args map[string]interface{}) (string, error) {
			return "✓ Ticket logged and archived", nil
		},
		"send_summary": func(ctx context.Context, args map[string]interface{}) (string, error) {
			return "✓ Summary sent to user via email", nil
		},
	}

	// Load complete team from unified YAML
	yamlPath := "team.yaml"
	team, err := agentic.LoadTeamFromYAML(yamlPath, toolHandlers)
	if err != nil {
		fmt.Printf("❌ Failed to load team config: %v\n", err)
		os.Exit(1)
	}

	// Create executor
	executor := agentic.NewTeamExecutor(team, key)

	// Define support tickets
	tickets := []struct {
		id    string
		issue string
	}{
		{
			id:    "TKT-001",
			issue: "Computer won't turn on. I pressed the power button but nothing happens. No lights, no fans, completely dead.",
		},
		{
			id:    "TKT-002",
			issue: "Microsoft Office installation failed with error code 0x80070005. I don't have admin rights.",
		},
		{
			id:    "TKT-003",
			issue: "Cannot connect to company VPN. Getting timeout errors. Other sites load fine.",
		},
	}

	// Process tickets
	fmt.Println("\n🎫 IT Support System - Unified Configuration\n" + strings.Repeat("=", 60))
	for _, ticket := range tickets {
		fmt.Printf("\n📋 Ticket: %s\n%s\n", ticket.id, strings.Repeat("-", 60))

		prompt := fmt.Sprintf("[Support Ticket %s] %s", ticket.id, ticket.issue)
		response, err := executor.Execute(context.Background(), prompt)

		if err != nil {
			fmt.Printf("❌ Error processing ticket: %v\n", err)
		} else {
			fmt.Printf("✅ Resolution:\n%s\n", response)
		}
	}

	fmt.Println("\n" + strings.Repeat("=", 60) + "\n🎉 All tickets processed!\n")
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

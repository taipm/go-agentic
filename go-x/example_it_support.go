package crewai

import (
	"context"
	"fmt"
	"os/exec"
	"runtime"
	"strconv"
	"strings"
)

// CreateITSupportCrew creates a complete IT Support crew
func CreateITSupportCrew() *Crew {
	// Define tools
	tools := createITSupportTools()

	// Create agents
	orchestrator := &Agent{
		ID:          "orchestrator",
		Name:        "Orchestrator",
		Role:        "System coordinator and entry point",
		Backstory:   "You are the entry point for IT support requests. Analyze the problem and decide if you need more information before proceeding to execution.",
		Model:       "gpt-4o",
		Tools:       []*Tool{},
		Temperature: 0.7,
		IsTerminal:  false,
	}

	clarifier := &Agent{
		ID:          "clarifier",
		Name:        "Clarifier",
		Role:        "Information gatherer",
		Backstory:   "You specialize in gathering detailed information about IT issues. Ask clarifying questions to understand the problem better.",
		Model:       "gpt-4o",
		Tools:       []*Tool{},
		Temperature: 0.7,
		IsTerminal:  false,
	}

	executor := &Agent{
		ID:          "executor",
		Name:        "Executor",
		Role:        "IT troubleshooter and diagnostician",
		Backstory:   "You are an expert IT troubleshooter. Use available tools to diagnose issues and provide solutions.",
		Model:       "gpt-4o",
		Tools:       tools,
		Temperature: 0.7,
		IsTerminal:  true,
	}

	// Create crew
	crew := &Crew{
		Agents:      []*Agent{orchestrator, clarifier, executor},
		MaxRounds:   10,
		MaxHandoffs: 5,
	}

	return crew
}

// GetAllITSupportTools returns all IT support tools as a map
func GetAllITSupportTools() map[string]*Tool {
	tools := createITSupportTools()
	toolMap := make(map[string]*Tool)
	for _, tool := range tools {
		toolMap[tool.Name] = tool
	}
	return toolMap
}

// createITSupportTools creates all IT support tools
func createITSupportTools() []*Tool {
	return []*Tool{
		{
			Name:        "GetCPUUsage",
			Description: "Get current CPU usage percentage",
			Parameters: map[string]interface{}{
				"type":       "object",
				"properties": map[string]interface{}{},
			},
			Handler: func(ctx context.Context, args map[string]interface{}) (string, error) {
				return getCPUUsage(ctx)
			},
		},
		{
			Name:        "GetMemoryUsage",
			Description: "Get current memory usage",
			Parameters: map[string]interface{}{
				"type":       "object",
				"properties": map[string]interface{}{},
			},
			Handler: func(ctx context.Context, args map[string]interface{}) (string, error) {
				return getMemoryUsage(ctx)
			},
		},
		{
			Name:        "GetDiskSpace",
			Description: "Get disk space usage for a path",
			Parameters: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"path": map[string]interface{}{
						"type":        "string",
						"description": "Path to check (default: /)",
					},
				},
			},
			Handler: getDiskSpaceHandler,
		},
		{
			Name:        "GetSystemInfo",
			Description: "Get system information (OS, hostname, uptime)",
			Parameters: map[string]interface{}{
				"type":       "object",
				"properties": map[string]interface{}{},
			},
			Handler: getSystemInfoHandler,
		},
		{
			Name:        "GetRunningProcesses",
			Description: "Get top running processes by CPU/memory usage",
			Parameters: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"count": map[string]interface{}{
						"type":        "string",
						"description": "Number of processes to show (default: 5)",
					},
				},
			},
			Handler: getRunningProcessesHandler,
		},
		{
			Name:        "PingHost",
			Description: "Ping a remote host to test connectivity",
			Parameters: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"host": map[string]interface{}{
						"type":        "string",
						"description": "Hostname or IP address to ping",
					},
					"count": map[string]interface{}{
						"type":        "string",
						"description": "Number of ping attempts (default: 4)",
					},
				},
				"required": []string{"host"},
			},
			Handler: pingHostHandler,
		},
		{
			Name:        "CheckServiceStatus",
			Description: "Check if a service is running",
			Parameters: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"service": map[string]interface{}{
						"type":        "string",
						"description": "Service name to check",
					},
				},
				"required": []string{"service"},
			},
			Handler: checkServiceStatusHandler,
		},
		{
			Name:        "ResolveDNS",
			Description: "Resolve a hostname to IP address",
			Parameters: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"hostname": map[string]interface{}{
						"type":        "string",
						"description": "Hostname to resolve",
					},
				},
				"required": []string{"hostname"},
			},
			Handler: resolveDNSHandler,
		},
		{
			Name:        "CheckMemoryStatus",
			Description: "Get detailed memory usage information for the local machine",
			Parameters: map[string]interface{}{
				"type":       "object",
				"properties": map[string]interface{}{},
			},
			Handler: checkMemoryStatusHandler,
		},
		{
			Name:        "CheckDiskStatus",
			Description: "Get detailed disk space usage for specified path or root directory",
			Parameters: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"path": map[string]interface{}{
						"type":        "string",
						"description": "Path to check (e.g., '/', '/home', '/var'). Default: /",
					},
				},
			},
			Handler: checkDiskStatusHandler,
		},
		{
			Name:        "CheckNetworkStatus",
			Description: "Check network connectivity by pinging a host",
			Parameters: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"host": map[string]interface{}{
						"type":        "string",
						"description": "Hostname or IP to test (e.g., 8.8.8.8, google.com)",
					},
					"count": map[string]interface{}{
						"type":        "string",
						"description": "Number of pings (default: 3)",
					},
				},
				"required": []string{"host"},
			},
			Handler: checkNetworkStatusHandler,
		},
		{
			Name:        "ExecuteCommand",
			Description: "Execute a shell command and return its output. Use with caution!",
			Parameters: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"command": map[string]interface{}{
						"type":        "string",
						"description": "Shell command to execute (e.g., 'ps aux', 'ls -la', 'uname -a')",
					},
				},
				"required": []string{"command"},
			},
			Handler: executeCommandHandler,
		},
		{
			Name:        "GetSystemDiagnostics",
			Description: "Get comprehensive system diagnostics (CPU, Memory, Disk, Network)",
			Parameters: map[string]interface{}{
				"type":       "object",
				"properties": map[string]interface{}{},
			},
			Handler: getSystemDiagnosticsHandler,
		},
	}
}

// System information tool implementations

func getCPUUsage(ctx context.Context) (string, error) {
	if runtime.GOOS == "darwin" {
		cmd := exec.CommandContext(ctx, "sh", "-c", "ps aux | awk 'BEGIN{sum=0}{sum+=$3}END{printf \"%.1f\", sum}'")
		output, err := cmd.Output()
		if err != nil {
			return "", fmt.Errorf("failed to get CPU usage: %w", err)
		}
		return string(output) + "%", nil
	}

	cmd := exec.CommandContext(ctx, "sh", "-c", "grep -o \"cpu [0-9.]*\" /proc/stat | awk '{print $2}'")
	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("failed to get CPU usage: %w", err)
	}

	return strings.TrimSpace(string(output)) + "%", nil
}

func getMemoryUsage(ctx context.Context) (string, error) {
	if runtime.GOOS == "darwin" {
		cmd := exec.CommandContext(ctx, "vm_stat")
		output, err := cmd.Output()
		if err != nil {
			return "", fmt.Errorf("failed to get memory usage: %w", err)
		}
		return strings.TrimSpace(string(output)), nil
	}

	cmd := exec.CommandContext(ctx, "free", "-h")
	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("failed to get memory usage: %w", err)
	}

	return strings.TrimSpace(string(output)), nil
}

func getDiskSpaceHandler(ctx context.Context, args map[string]interface{}) (string, error) {
	path := "/"
	if p, ok := args["path"]; ok {
		if ps, ok := p.(string); ok {
			path = ps
		}
	}

	cmd := exec.CommandContext(ctx, "df", "-h", path)
	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("failed to get disk space: %w", err)
	}

	return strings.TrimSpace(string(output)), nil
}

func getSystemInfoHandler(ctx context.Context, args map[string]interface{}) (string, error) {
	var info strings.Builder

	info.WriteString(fmt.Sprintf("OS: %s\n", runtime.GOOS))

	cmd := exec.CommandContext(ctx, "hostname")
	if output, err := cmd.Output(); err == nil {
		info.WriteString(fmt.Sprintf("Hostname: %s\n", strings.TrimSpace(string(output))))
	}

	return info.String(), nil
}

func getRunningProcessesHandler(ctx context.Context, args map[string]interface{}) (string, error) {
	count := "5"
	if c, ok := args["count"]; ok {
		if cs, ok := c.(string); ok {
			count = cs
		}
	}

	// Validate count is a number
	if _, err := strconv.Atoi(count); err != nil {
		count = "5"
	}

	cmd := exec.CommandContext(ctx, "sh", "-c", fmt.Sprintf("ps aux | head -n %s", count))
	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("failed to get processes: %w", err)
	}

	return strings.TrimSpace(string(output)), nil
}

func pingHostHandler(ctx context.Context, args map[string]interface{}) (string, error) {
	host, ok := args["host"].(string)
	if !ok {
		return "", fmt.Errorf("host parameter required")
	}

	count := "4"
	if c, ok := args["count"]; ok {
		if cs, ok := c.(string); ok {
			count = cs
		}
	}

	// Validate count
	if _, err := strconv.Atoi(count); err != nil {
		count = "4"
	}

	cmd := exec.CommandContext(ctx, "ping", "-c", count, host)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("ping failed: %w", err)
	}

	return strings.TrimSpace(string(output)), nil
}

func checkServiceStatusHandler(ctx context.Context, args map[string]interface{}) (string, error) {
	service, ok := args["service"].(string)
	if !ok {
		return "", fmt.Errorf("service parameter required")
	}

	if runtime.GOOS == "darwin" {
		cmd := exec.CommandContext(ctx, "sh", "-c", fmt.Sprintf("launchctl list | grep %s", service))
		output, _ := cmd.Output()
		if len(output) > 0 {
			return fmt.Sprintf("Service %s is running", service), nil
		}
		return fmt.Sprintf("Service %s is not running", service), nil
	}

	cmd := exec.CommandContext(ctx, "systemctl", "is-active", service)
	output, _ := cmd.Output()

	if strings.Contains(strings.TrimSpace(string(output)), "active") {
		return fmt.Sprintf("Service %s is running", service), nil
	}

	return fmt.Sprintf("Service %s is not running", service), nil
}

func resolveDNSHandler(ctx context.Context, args map[string]interface{}) (string, error) {
	hostname, ok := args["hostname"].(string)
	if !ok {
		return "", fmt.Errorf("hostname parameter required")
	}

	cmd := exec.CommandContext(ctx, "nslookup", hostname)
	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("DNS resolution failed: %w", err)
	}

	return strings.TrimSpace(string(output)), nil
}

// checkMemoryStatusHandler gets detailed memory information
func checkMemoryStatusHandler(ctx context.Context, args map[string]interface{}) (string, error) {
	if runtime.GOOS == "darwin" {
		// macOS: use vm_stat
		cmd := exec.CommandContext(ctx, "vm_stat")
		output, err := cmd.Output()
		if err != nil {
			return "", fmt.Errorf("failed to get memory status: %w", err)
		}
		return strings.TrimSpace(string(output)), nil
	}

	// Linux: use free command
	cmd := exec.CommandContext(ctx, "free", "-h")
	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("failed to get memory status: %w", err)
	}
	return strings.TrimSpace(string(output)), nil
}

// checkDiskStatusHandler gets detailed disk usage information
func checkDiskStatusHandler(ctx context.Context, args map[string]interface{}) (string, error) {
	path := "/"
	if p, ok := args["path"]; ok {
		if ps, ok := p.(string); ok && ps != "" {
			path = ps
		}
	}

	cmd := exec.CommandContext(ctx, "df", "-h", path)
	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("failed to get disk status for %s: %w", path, err)
	}
	return strings.TrimSpace(string(output)), nil
}

// checkNetworkStatusHandler tests network connectivity
func checkNetworkStatusHandler(ctx context.Context, args map[string]interface{}) (string, error) {
	host, ok := args["host"].(string)
	if !ok {
		return "", fmt.Errorf("host parameter required")
	}

	count := "3"
	if c, ok := args["count"]; ok {
		if cs, ok := c.(string); ok && cs != "" {
			count = cs
		}
	}

	// Validate count
	if _, err := strconv.Atoi(count); err != nil {
		count = "3"
	}

	cmd := exec.CommandContext(ctx, "ping", "-c", count, host)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Sprintf("Ping to %s failed:\n%s", host, string(output)), nil
	}
	return strings.TrimSpace(string(output)), nil
}

// executeCommandHandler executes an arbitrary shell command
func executeCommandHandler(ctx context.Context, args map[string]interface{}) (string, error) {
	command, ok := args["command"].(string)
	if !ok || command == "" {
		return "", fmt.Errorf("command parameter required")
	}

	// Safety check: prevent dangerous commands
	dangerousPatterns := []string{"rm -rf", "mkfs", "dd if=", ":(){:|:", "fork"}
	for _, pattern := range dangerousPatterns {
		if strings.Contains(strings.ToLower(command), strings.ToLower(pattern)) {
			return "", fmt.Errorf("dangerous command blocked: %s", pattern)
		}
	}

	cmd := exec.CommandContext(ctx, "sh", "-c", command)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Sprintf("Command failed:\n%s\nError: %v", string(output), err), nil
	}
	return strings.TrimSpace(string(output)), nil
}

// getSystemDiagnosticsHandler provides comprehensive system diagnostics
func getSystemDiagnosticsHandler(ctx context.Context, args map[string]interface{}) (string, error) {
	var diagnostics strings.Builder

	diagnostics.WriteString("=== SYSTEM DIAGNOSTICS ===\n\n")

	// System Info
	diagnostics.WriteString("--- System Information ---\n")
	if sysInfo, err := getSystemInfoHandler(ctx, nil); err == nil {
		diagnostics.WriteString(sysInfo)
	}
	diagnostics.WriteString("\n")

	// CPU Usage
	diagnostics.WriteString("--- CPU Usage ---\n")
	if cpu, err := getCPUUsage(ctx); err == nil {
		diagnostics.WriteString("CPU Usage: " + cpu + "\n")
	}
	diagnostics.WriteString("\n")

	// Memory Usage
	diagnostics.WriteString("--- Memory Status ---\n")
	if mem, err := checkMemoryStatusHandler(ctx, nil); err == nil {
		diagnostics.WriteString(mem + "\n")
	}
	diagnostics.WriteString("\n")

	// Disk Usage
	diagnostics.WriteString("--- Disk Status (Root) ---\n")
	if disk, err := checkDiskStatusHandler(ctx, map[string]interface{}{"path": "/"}); err == nil {
		diagnostics.WriteString(disk + "\n")
	}
	diagnostics.WriteString("\n")

	// Running Processes
	diagnostics.WriteString("--- Top 5 Running Processes ---\n")
	if procs, err := getRunningProcessesHandler(ctx, map[string]interface{}{"count": "5"}); err == nil {
		diagnostics.WriteString(procs + "\n")
	}

	return diagnostics.String(), nil
}

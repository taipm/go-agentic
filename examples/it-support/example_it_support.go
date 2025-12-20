package main

import (
	"context"
	"fmt"
	"os/exec"
	"runtime"
	"strconv"
	"strings"

	"github.com/taipm/go-agentic"
)

// createITSupportCrew creates a complete IT Support crew
func createITSupportCrew() *agentic.Crew {
	// Define tools
	tools := createITSupportTools()

	// Create agents
	orchestrator := &agentic.Agent{
		ID:          "orchestrator",
		Name:        "Orchestrator",
		Role:        "System coordinator and entry point",
		Backstory:   "You are the entry point for IT support requests. Analyze the problem and decide if you need more information before proceeding to execution.",
		Model:       "gpt-4o-mini",
		Tools:       []*agentic.Tool{},
		Temperature: 0.7,
		IsTerminal:  false,
	}

	clarifier := &agentic.Agent{
		ID:          "clarifier",
		Name:        "Clarifier",
		Role:        "Information gatherer",
		Backstory:   "You specialize in gathering detailed information about IT issues. Ask clarifying questions to understand the problem better.",
		Model:       "gpt-4o-mini",
		Tools:       []*agentic.Tool{},
		Temperature: 0.7,
		IsTerminal:  false,
	}

	executor := &agentic.Agent{
		ID:          "executor",
		Name:        "Executor",
		Role:        "IT troubleshooter and diagnostician",
		Backstory:   "You are an expert IT troubleshooter. Use available tools to diagnose issues and provide solutions.",
		Model:       "gpt-4o-mini",
		Tools:       tools,
		Temperature: 0.7,
		IsTerminal:  true,
	}

	// Create crew
	crew := &agentic.Crew{
		Agents:      []*agentic.Agent{orchestrator, clarifier, executor},
		MaxRounds:   10,
		MaxHandoffs: 5,
	}

	return crew
}

// createITSupportTools creates all IT support tools
func createITSupportTools() []*agentic.Tool {
	return []*agentic.Tool{
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
	// Get platform-specific CPU usage command
	cmd, err := getCPUUsageCommand(ctx)
	if err != nil {
		return "", fmt.Errorf("[COMMAND_FAILED] Failed to create CPU usage command: %w. "+
			"Suggestion: Check if required system commands are available", err)
	}

	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("[COMMAND_FAILED] Failed to get CPU usage: %w. "+
			"Suggestion: Check system permissions or try running with elevated privileges", err)
	}

	return strings.TrimSpace(string(output)) + "%", nil
}

// getCPUUsageCommand returns platform-specific CPU usage command
func getCPUUsageCommand(ctx context.Context) (*exec.Cmd, error) {
	switch runtime.GOOS {
	case "windows":
		// Windows: use wmic (Windows Management Instrumentation Command-line)
		return exec.CommandContext(ctx, "wmic", "cpu", "get", "loadpercentage"), nil
	case "darwin":
		// macOS: use ps and awk to sum CPU usage
		return exec.CommandContext(ctx, "sh", "-c", "ps aux | awk 'BEGIN{sum=0}{sum+=$3}END{printf \"%.1f\", sum}'"), nil
	case "linux":
		// Linux: use /proc/stat
		return exec.CommandContext(ctx, "sh", "-c", "grep -o \"cpu [0-9.]*\" /proc/stat | awk '{print $2}'"), nil
	default:
		return nil, fmt.Errorf("unsupported OS: %s", runtime.GOOS)
	}
}

func getMemoryUsage(ctx context.Context) (string, error) {
	// Get platform-specific memory usage command
	cmd, err := getMemoryUsageCommand(ctx)
	if err != nil {
		return "", fmt.Errorf("[COMMAND_FAILED] Failed to create memory usage command: %w. "+
			"Suggestion: Check if required system commands are available", err)
	}

	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("[COMMAND_FAILED] Failed to get memory usage: %w. "+
			"Suggestion: Check system permissions or try running with elevated privileges", err)
	}

	return strings.TrimSpace(string(output)), nil
}

// getMemoryUsageCommand returns platform-specific memory usage command
func getMemoryUsageCommand(ctx context.Context) (*exec.Cmd, error) {
	switch runtime.GOOS {
	case "windows":
		// Windows: use wmic to get memory info
		return exec.CommandContext(ctx, "wmic", "OS", "get", "TotalVisibleMemorySize,FreePhysicalMemory"), nil
	case "darwin":
		// macOS: use vm_stat
		return exec.CommandContext(ctx, "vm_stat"), nil
	case "linux":
		// Linux: use free command
		return exec.CommandContext(ctx, "free", "-h"), nil
	default:
		return nil, fmt.Errorf("unsupported OS: %s", runtime.GOOS)
	}
}

func getDiskSpaceHandler(ctx context.Context, args map[string]interface{}) (string, error) {
	path := "/"
	if p, ok := args["path"]; ok {
		if ps, ok := p.(string); ok {
			path = ps
		}
	}

	// Get platform-specific disk space command
	cmd, err := getDiskSpaceCommand(ctx, path)
	if err != nil {
		return "", fmt.Errorf("[COMMAND_FAILED] Failed to create disk space command: %w. "+
			"Suggestion: Check if required system commands are available", err)
	}

	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("[COMMAND_FAILED] Failed to get disk space for %s: %w. "+
			"Suggestion: Check path exists and you have permission to access it", path, err)
	}

	return strings.TrimSpace(string(output)), nil
}

// getDiskSpaceCommand returns platform-specific disk space command
func getDiskSpaceCommand(ctx context.Context, path string) (*exec.Cmd, error) {
	switch runtime.GOOS {
	case "windows":
		// Windows: use dir or wmic for disk info
		// Use dir to get volume info
		volume := "C:"
		if len(path) > 0 {
			volume = strings.Split(path, ":")[0] + ":"
		}
		return exec.CommandContext(ctx, "dir", volume), nil
	case "darwin", "linux":
		// Unix-like: use df
		return exec.CommandContext(ctx, "df", "-h", path), nil
	default:
		return nil, fmt.Errorf("unsupported OS: %s", runtime.GOOS)
	}
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

	// Get platform-specific process listing command
	cmd, err := getProcessListCommand(ctx, count)
	if err != nil {
		return "", fmt.Errorf("[COMMAND_FAILED] Failed to create process listing command: %w. "+
			"Suggestion: Check if ps/tasklist is available on your system", err)
	}

	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("[COMMAND_FAILED] Failed to get processes: %w. "+
			"Suggestion: Run with elevated privileges or check command syntax", err)
	}

	return strings.TrimSpace(string(output)), nil
}

// getProcessListCommand returns platform-specific process listing command
func getProcessListCommand(ctx context.Context, count string) (*exec.Cmd, error) {
	countInt, _ := strconv.Atoi(count)
	switch runtime.GOOS {
	case "windows":
		// Windows: use tasklist
		return exec.CommandContext(ctx, "tasklist"), nil
	case "darwin":
		// macOS: use ps with limit
		return exec.CommandContext(ctx, "sh", "-c", fmt.Sprintf("ps aux | head -n %d", countInt+1)), nil
	case "linux":
		// Linux: use ps with limit
		return exec.CommandContext(ctx, "sh", "-c", fmt.Sprintf("ps aux | head -n %d", countInt+1)), nil
	default:
		return nil, fmt.Errorf("unsupported OS: %s", runtime.GOOS)
	}
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

	// Get platform-specific ping command
	cmd, err := getPingCommand(ctx, host, count)
	if err != nil {
		return "", fmt.Errorf("[COMMAND_FAILED] Failed to create ping command: %w. "+
			"Suggestion: Check if ping is available on your system", err)
	}

	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("[COMMAND_FAILED] Ping to %s failed: %w. "+
			"Suggestion: Check network connectivity and target host availability", host, err)
	}

	return strings.TrimSpace(string(output)), nil
}

// getPingCommand returns platform-specific ping command
func getPingCommand(ctx context.Context, host string, count string) (*exec.Cmd, error) {
	switch runtime.GOOS {
	case "windows":
		return exec.CommandContext(ctx, "ping", "-n", count, host), nil
	case "darwin", "linux":
		return exec.CommandContext(ctx, "ping", "-c", count, host), nil
	default:
		return nil, fmt.Errorf("unsupported OS: %s", runtime.GOOS)
	}
}

func checkServiceStatusHandler(ctx context.Context, args map[string]interface{}) (string, error) {
	service, ok := args["service"].(string)
	if !ok {
		return "", fmt.Errorf("service parameter required")
	}

	// Get platform-specific service check command
	cmd, err := getServiceStatusCommand(ctx, service)
	if err != nil {
		return "", fmt.Errorf("[COMMAND_FAILED] Failed to create service status command: %w. "+
			"Suggestion: Check if required system commands are available", err)
	}

	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("[COMMAND_FAILED] Failed to check service status for %s: %w. "+
			"Suggestion: Run with elevated privileges (sudo) or check if service exists", service, err)
	}

	outputStr := strings.TrimSpace(string(output))
	isRunning := isServiceRunning(outputStr, runtime.GOOS)
	if isRunning {
		return fmt.Sprintf("Service %s is running", service), nil
	}
	return fmt.Sprintf("Service %s is not running", service), nil
}

// getServiceStatusCommand returns platform-specific service status command
func getServiceStatusCommand(ctx context.Context, service string) (*exec.Cmd, error) {
	switch runtime.GOOS {
	case "windows":
		// Windows: use Get-Service PowerShell command
		return exec.CommandContext(ctx, "powershell", "-Command", fmt.Sprintf("Get-Service -Name %s | Select-Object Status", service)), nil
	case "darwin":
		// macOS: use launchctl
		return exec.CommandContext(ctx, "sh", "-c", fmt.Sprintf("launchctl list | grep %s", service)), nil
	case "linux":
		// Linux: use systemctl
		return exec.CommandContext(ctx, "systemctl", "is-active", service), nil
	default:
		return nil, fmt.Errorf("unsupported OS: %s", runtime.GOOS)
	}
}

// isServiceRunning determines if service is running based on output and OS
func isServiceRunning(output string, goos string) bool {
	output = strings.TrimSpace(output)
	switch goos {
	case "windows":
		return strings.Contains(output, "Running")
	case "darwin":
		return len(output) > 0
	case "linux":
		// Check for exact "active" status (not "inactive")
		return output == "active"
	default:
		return false
	}
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
	// Get platform-specific memory status command
	cmd, err := getMemoryStatusCommand(ctx)
	if err != nil {
		return "", fmt.Errorf("[COMMAND_FAILED] Failed to create memory status command: %w. "+
			"Suggestion: Check if required system commands are available", err)
	}

	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("[COMMAND_FAILED] Failed to get memory status: %w. "+
			"Suggestion: Check system permissions or try running with elevated privileges", err)
	}
	return strings.TrimSpace(string(output)), nil
}

// getMemoryStatusCommand returns platform-specific memory status command
func getMemoryStatusCommand(ctx context.Context) (*exec.Cmd, error) {
	switch runtime.GOOS {
	case "windows":
		// Windows: use wmic for memory info
		return exec.CommandContext(ctx, "wmic", "OS", "get", "TotalVisibleMemorySize,FreePhysicalMemory"), nil
	case "darwin":
		// macOS: use vm_stat
		return exec.CommandContext(ctx, "vm_stat"), nil
	case "linux":
		// Linux: use free command
		return exec.CommandContext(ctx, "free", "-h"), nil
	default:
		return nil, fmt.Errorf("unsupported OS: %s", runtime.GOOS)
	}
}

// checkDiskStatusHandler gets detailed disk usage information
func checkDiskStatusHandler(ctx context.Context, args map[string]interface{}) (string, error) {
	path := "/"
	if p, ok := args["path"]; ok {
		if ps, ok := p.(string); ok && ps != "" {
			path = ps
		}
	}

	// Get platform-specific disk status command
	cmd, err := getDiskStatusCommand(ctx, path)
	if err != nil {
		return "", fmt.Errorf("[COMMAND_FAILED] Failed to create disk status command: %w. "+
			"Suggestion: Check if required system commands are available", err)
	}

	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("[COMMAND_FAILED] Failed to get disk status for %s: %w. "+
			"Suggestion: Check path exists and you have permission to access it", path, err)
	}
	return strings.TrimSpace(string(output)), nil
}

// getDiskStatusCommand returns platform-specific disk status command
func getDiskStatusCommand(ctx context.Context, path string) (*exec.Cmd, error) {
	switch runtime.GOOS {
	case "windows":
		// Windows: use dir or wmic for disk info
		volume := "C:"
		if len(path) > 0 {
			parts := strings.Split(path, ":")
			if len(parts) > 0 {
				volume = parts[0] + ":"
			}
		}
		return exec.CommandContext(ctx, "dir", volume), nil
	case "darwin", "linux":
		// Unix-like: use df
		return exec.CommandContext(ctx, "df", "-h", path), nil
	default:
		return nil, fmt.Errorf("unsupported OS: %s", runtime.GOOS)
	}
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

	// Get platform-specific ping command
	cmd, err := getPingCommand(ctx, host, count)
	if err != nil {
		return fmt.Sprintf("Ping command error on %s: %v\nSuggestion: Check if ping is available", runtime.GOOS, err), nil
	}

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

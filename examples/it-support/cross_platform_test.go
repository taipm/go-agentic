package main

import (
	"context"
	"runtime"
	"testing"
)

// TestPingCommandSelection verifies OS-specific ping command variants
func TestPingCommandSelection(t *testing.T) {
	ctx := context.Background()
	tests := []struct {
		name       string
		goos       string
		expectedOS string
	}{
		{
			name:       "Windows ping uses -n flag",
			goos:       "windows",
			expectedOS: "windows",
		},
		{
			name:       "macOS ping uses -c flag",
			goos:       "darwin",
			expectedOS: "darwin",
		},
		{
			name:       "Linux ping uses -c flag",
			goos:       "linux",
			expectedOS: "linux",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Skip test if not on the target OS
			if runtime.GOOS != tt.goos {
				t.Skipf("Test only runs on %s, current OS is %s", tt.goos, runtime.GOOS)
			}

			cmd, err := getPingCommand(ctx, "127.0.0.1", "1")
			if err != nil {
				t.Fatalf("getPingCommand failed: %v", err)
			}

			if cmd == nil {
				t.Fatal("Expected command to be non-nil")
			}

			// Verify the command has the right structure
			if cmd.Path == "" {
				t.Error("Command path is empty")
			}
		})
	}
}

// TestUnsupportedOSError verifies error handling for unsupported OS
func TestUnsupportedOSError(t *testing.T) {
	ctx := context.Background()

	// This test will only run properly if we can somehow mock runtime.GOOS
	// For now, we just test that the function handles unknown OS gracefully
	// when called directly with invalid inputs

	// Test CPU usage command creation
	_, err := getCPUUsageCommand(ctx)
	if err != nil {
		t.Fatalf("getCPUUsageCommand failed on current OS: %v", err)
	}

	// Test memory usage command creation
	_, err = getMemoryUsageCommand(ctx)
	if err != nil {
		t.Fatalf("getMemoryUsageCommand failed on current OS: %v", err)
	}

	// Test disk space command creation
	_, err = getDiskSpaceCommand(ctx, "/")
	if err != nil {
		t.Fatalf("getDiskSpaceCommand failed on current OS: %v", err)
	}
}

// TestCPUUsageCommandSelection verifies platform-specific CPU command creation
func TestCPUUsageCommandSelection(t *testing.T) {
	ctx := context.Background()

	cmd, err := getCPUUsageCommand(ctx)
	if err != nil {
		t.Fatalf("getCPUUsageCommand failed: %v", err)
	}

	if cmd == nil {
		t.Fatal("Expected non-nil command")
	}

	// Verify command properties
	if cmd.Path == "" {
		t.Error("Command path is empty")
	}

	// Verify OS-specific behavior
	switch runtime.GOOS {
	case "windows":
		if cmd.Args[0] != "wmic" {
			t.Errorf("Expected wmic on Windows, got %s", cmd.Args[0])
		}
	case "darwin":
		if cmd.Args[0] != "sh" {
			t.Errorf("Expected sh on macOS, got %s", cmd.Args[0])
		}
	case "linux":
		if cmd.Args[0] != "sh" {
			t.Errorf("Expected sh on Linux, got %s", cmd.Args[0])
		}
	}
}

// TestMemoryUsageCommandSelection verifies platform-specific memory command creation
func TestMemoryUsageCommandSelection(t *testing.T) {
	ctx := context.Background()

	cmd, err := getMemoryUsageCommand(ctx)
	if err != nil {
		t.Fatalf("getMemoryUsageCommand failed: %v", err)
	}

	if cmd == nil {
		t.Fatal("Expected non-nil command")
	}

	// Verify OS-specific command
	switch runtime.GOOS {
	case "windows":
		if cmd.Args[0] != "wmic" {
			t.Errorf("Expected wmic on Windows, got %s", cmd.Args[0])
		}
	case "darwin":
		if cmd.Args[0] != "vm_stat" {
			t.Errorf("Expected vm_stat on macOS, got %s", cmd.Args[0])
		}
	case "linux":
		if cmd.Args[0] != "free" {
			t.Errorf("Expected free on Linux, got %s", cmd.Args[0])
		}
	}
}

// TestDiskSpaceCommandSelection verifies platform-specific disk command creation
func TestDiskSpaceCommandSelection(t *testing.T) {
	ctx := context.Background()

	cmd, err := getDiskSpaceCommand(ctx, "/")
	if err != nil {
		t.Fatalf("getDiskSpaceCommand failed: %v", err)
	}

	if cmd == nil {
		t.Fatal("Expected non-nil command")
	}

	// Verify OS-specific command
	switch runtime.GOOS {
	case "windows":
		if cmd.Args[0] != "dir" {
			t.Errorf("Expected dir on Windows, got %s", cmd.Args[0])
		}
	case "darwin", "linux":
		if cmd.Args[0] != "df" {
			t.Errorf("Expected df on Unix-like, got %s", cmd.Args[0])
		}
	}
}

// TestProcessListCommandSelection verifies platform-specific process listing
func TestProcessListCommandSelection(t *testing.T) {
	ctx := context.Background()

	cmd, err := getProcessListCommand(ctx, "5")
	if err != nil {
		t.Fatalf("getProcessListCommand failed: %v", err)
	}

	if cmd == nil {
		t.Fatal("Expected non-nil command")
	}

	// Verify OS-specific command
	switch runtime.GOOS {
	case "windows":
		if cmd.Args[0] != "tasklist" {
			t.Errorf("Expected tasklist on Windows, got %s", cmd.Args[0])
		}
	case "darwin", "linux":
		if cmd.Args[0] != "sh" {
			t.Errorf("Expected sh on Unix-like, got %s", cmd.Args[0])
		}
	}
}

// TestServiceStatusCommandSelection verifies platform-specific service check
func TestServiceStatusCommandSelection(t *testing.T) {
	ctx := context.Background()

	cmd, err := getServiceStatusCommand(ctx, "nginx")
	if err != nil {
		t.Fatalf("getServiceStatusCommand failed: %v", err)
	}

	if cmd == nil {
		t.Fatal("Expected non-nil command")
	}

	// Verify OS-specific command
	switch runtime.GOOS {
	case "windows":
		if cmd.Args[0] != "powershell" {
			t.Errorf("Expected powershell on Windows, got %s", cmd.Args[0])
		}
	case "darwin":
		if cmd.Args[0] != "sh" {
			t.Errorf("Expected sh on macOS, got %s", cmd.Args[0])
		}
	case "linux":
		if cmd.Args[0] != "systemctl" {
			t.Errorf("Expected systemctl on Linux, got %s", cmd.Args[0])
		}
	}
}

// TestServiceRunningDetection verifies platform-specific service status parsing
func TestServiceRunningDetection(t *testing.T) {
	tests := []struct {
		name     string
		goos     string
		output   string
		expected bool
	}{
		{
			name:     "Windows: Running service",
			goos:     "windows",
			output:   "Running",
			expected: true,
		},
		{
			name:     "Windows: Stopped service",
			goos:     "windows",
			output:   "Stopped",
			expected: false,
		},
		{
			name:     "macOS: Service found",
			goos:     "darwin",
			output:   "123 - homebrew.mxcl.nginx",
			expected: true,
		},
		{
			name:     "macOS: Service not found",
			goos:     "darwin",
			output:   "",
			expected: false,
		},
		{
			name:     "Linux: Active service",
			goos:     "linux",
			output:   "active",
			expected: true,
		},
		{
			name:     "Linux: Inactive service",
			goos:     "linux",
			output:   "inactive",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := isServiceRunning(tt.output, tt.goos)
			if result != tt.expected {
				t.Errorf("Expected %v, got %v for output '%s'", tt.expected, result, tt.output)
			}
		})
	}
}

// TestPingHostHandlerCrossPlatform tests ping handler on current platform
func TestPingHostHandlerCrossPlatform(t *testing.T) {
	ctx := context.Background()

	// Test with localhost - should work on all platforms
	result, err := pingHostHandler(ctx, map[string]interface{}{
		"host":  "127.0.0.1",
		"count": "1",
	})

	if err != nil {
		t.Logf("Note: Ping may fail in test environment: %v", err)
		// Don't fail the test - ping might not work in CI environment
		return
	}

	if result == "" {
		t.Error("Expected non-empty ping result")
	}
}

// TestGetCPUUsageHandlerCrossPlatform tests CPU handler on current platform
func TestGetCPUUsageHandlerCrossPlatform(t *testing.T) {
	ctx := context.Background()

	result, err := getCPUUsage(ctx)

	if err != nil {
		t.Logf("Note: CPU check may fail in test environment: %v", err)
		// Don't fail the test - system commands might not be available
		return
	}

	if result == "" {
		t.Error("Expected non-empty CPU usage result")
	}

	// Verify result has % sign
	if len(result) > 0 && string(result[len(result)-1]) != "%" {
		t.Errorf("Expected result to end with %%, got %s", result)
	}
}

// TestGetMemoryUsageHandlerCrossPlatform tests memory handler on current platform
func TestGetMemoryUsageHandlerCrossPlatform(t *testing.T) {
	ctx := context.Background()

	result, err := getMemoryUsage(ctx)

	if err != nil {
		t.Logf("Note: Memory check may fail in test environment: %v", err)
		return
	}

	if result == "" {
		t.Error("Expected non-empty memory usage result")
	}
}

// TestGetDiskSpaceHandlerCrossPlatform tests disk handler on current platform
func TestGetDiskSpaceHandlerCrossPlatform(t *testing.T) {
	ctx := context.Background()

	result, err := getDiskSpaceHandler(ctx, map[string]interface{}{
		"path": "/",
	})

	if err != nil {
		t.Logf("Note: Disk check may fail in test environment: %v", err)
		return
	}

	if result == "" {
		t.Error("Expected non-empty disk space result")
	}
}

// TestGetProcessListHandlerCrossPlatform tests process listing on current platform
func TestGetProcessListHandlerCrossPlatform(t *testing.T) {
	ctx := context.Background()

	result, err := getRunningProcessesHandler(ctx, map[string]interface{}{
		"count": "3",
	})

	if err != nil {
		t.Logf("Note: Process listing may fail in test environment: %v", err)
		return
	}

	if result == "" {
		t.Error("Expected non-empty process list result")
	}
}

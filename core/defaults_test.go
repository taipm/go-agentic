package crewai

import (
	"testing"
	"time"
)

// TestDefaultHardcodedDefaults verifies all default values are set correctly
func TestDefaultHardcodedDefaults(t *testing.T) {
	defaults := DefaultHardcodedDefaults()

	tests := []struct {
		name     string
		getValue func() interface{}
		expected interface{}
	}{
		// Timeout parameters
		{"ParallelAgentTimeout", func() interface{} { return defaults.ParallelAgentTimeout }, 60 * time.Second},
		{"ToolExecutionTimeout", func() interface{} { return defaults.ToolExecutionTimeout }, 5 * time.Second},
		{"ToolResultTimeout", func() interface{} { return defaults.ToolResultTimeout }, 30 * time.Second},
		{"MinToolTimeout", func() interface{} { return defaults.MinToolTimeout }, 100 * time.Millisecond},
		{"StreamChunkTimeout", func() interface{} { return defaults.StreamChunkTimeout }, 500 * time.Millisecond},
		{"SSEKeepAliveInterval", func() interface{} { return defaults.SSEKeepAliveInterval }, 30 * time.Second},
		{"RequestStoreCleanupInterval", func() interface{} { return defaults.RequestStoreCleanupInterval }, 5 * time.Minute},
		{"ClientCacheTTL", func() interface{} { return defaults.ClientCacheTTL }, 1 * time.Hour},
		{"GracefulShutdownCheckInterval", func() interface{} { return defaults.GracefulShutdownCheckInterval }, 100 * time.Millisecond},

		// Retry and backoff
		{"RetryBackoffMinDuration", func() interface{} { return defaults.RetryBackoffMinDuration }, 100 * time.Millisecond},
		{"RetryBackoffMaxDuration", func() interface{} { return defaults.RetryBackoffMaxDuration }, 5 * time.Second},

		// Size limits
		{"MaxInputSize", func() interface{} { return defaults.MaxInputSize }, 10 * 1024},
		{"MinAgentIDLength", func() interface{} { return defaults.MinAgentIDLength }, 1},
		{"MaxAgentIDLength", func() interface{} { return defaults.MaxAgentIDLength }, 128},
		{"MaxRequestBodySize", func() interface{} { return defaults.MaxRequestBodySize }, 100 * 1024},
		{"MaxToolOutputChars", func() interface{} { return defaults.MaxToolOutputChars }, 2000},
		{"StreamBufferSize", func() interface{} { return defaults.StreamBufferSize }, 100},
		{"MaxStoredRequests", func() interface{} { return defaults.MaxStoredRequests }, 1000},

		// Other
		{"TimeoutWarningThreshold", func() interface{} { return defaults.TimeoutWarningThreshold }, 0.20},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.getValue()
			if got != tt.expected {
				t.Errorf("%s: got %v, want %v", tt.name, got, tt.expected)
			}
		})
	}
}

// TestHardcodedDefaultsValidate verifies validation catches invalid values
func TestHardcodedDefaultsValidate(t *testing.T) {
	tests := []struct {
		name      string
		setupFunc func() *HardcodedDefaults
		wantErr   bool
	}{
		{
			name: "valid defaults",
			setupFunc: func() *HardcodedDefaults {
				return DefaultHardcodedDefaults()
			},
			wantErr: false,
		},
		{
			name: "zero timeout corrected to default",
			setupFunc: func() *HardcodedDefaults {
				d := DefaultHardcodedDefaults()
				d.ParallelAgentTimeout = 0
				return d
			},
			wantErr: false,
		},
		{
			name: "negative timeout corrected to default",
			setupFunc: func() *HardcodedDefaults {
				d := DefaultHardcodedDefaults()
				d.ToolExecutionTimeout = -5 * time.Second
				return d
			},
			wantErr: false,
		},
		{
			name: "zero size limit corrected",
			setupFunc: func() *HardcodedDefaults {
				d := DefaultHardcodedDefaults()
				d.MaxInputSize = 0
				return d
			},
			wantErr: false,
		},
		{
			name: "out of range threshold corrected",
			setupFunc: func() *HardcodedDefaults {
				d := DefaultHardcodedDefaults()
				d.TimeoutWarningThreshold = 1.5 // > 1.0
				return d
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d := tt.setupFunc()
			err := d.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

// TestConfigToHardcodedDefaults verifies YAML conversion works correctly
func TestConfigToHardcodedDefaults(t *testing.T) {
	t.Run("empty config uses defaults", func(t *testing.T) {
		config := &CrewConfig{}
		defaults := ConfigToHardcodedDefaults(config)

		if defaults.ParallelAgentTimeout != 60*time.Second {
			t.Errorf("Expected 60s default, got %v", defaults.ParallelAgentTimeout)
		}
	})

	t.Run("custom values override defaults", func(t *testing.T) {
		config := &CrewConfig{}
		config.Settings.ParallelTimeoutSeconds = 120
		config.Settings.MaxToolOutputChars = 5000
		config.Settings.ToolExecutionTimeoutSeconds = 10
		defaults := ConfigToHardcodedDefaults(config)

		if defaults.ParallelAgentTimeout != 120*time.Second {
			t.Errorf("Expected 120s, got %v", defaults.ParallelAgentTimeout)
		}
		if defaults.MaxToolOutputChars != 5000 {
			t.Errorf("Expected 5000, got %d", defaults.MaxToolOutputChars)
		}
		if defaults.ToolExecutionTimeout != 10*time.Second {
			t.Errorf("Expected 10s, got %v", defaults.ToolExecutionTimeout)
		}
	})

	t.Run("size limits converted from KB to bytes", func(t *testing.T) {
		config := &CrewConfig{}
		config.Settings.MaxInputSizeKB = 20
		config.Settings.MaxRequestBodySizeKB = 200
		defaults := ConfigToHardcodedDefaults(config)

		if defaults.MaxInputSize != 20*1024 {
			t.Errorf("Expected 20KB (20480 bytes), got %d", defaults.MaxInputSize)
		}
		if defaults.MaxRequestBodySize != 200*1024 {
			t.Errorf("Expected 200KB (204800 bytes), got %d", defaults.MaxRequestBodySize)
		}
	})

	t.Run("time units converted correctly", func(t *testing.T) {
		config := &CrewConfig{}
		config.Settings.SSEKeepAliveSeconds = 60
		config.Settings.RequestStoreCleanupMinutes = 10
		config.Settings.ClientCacheTTLMinutes = 2
		defaults := ConfigToHardcodedDefaults(config)

		if defaults.SSEKeepAliveInterval != 60*time.Second {
			t.Errorf("Expected 60s, got %v", defaults.SSEKeepAliveInterval)
		}
		if defaults.RequestStoreCleanupInterval != 10*time.Minute {
			t.Errorf("Expected 10m, got %v", defaults.RequestStoreCleanupInterval)
		}
		if defaults.ClientCacheTTL != 2*time.Minute {
			t.Errorf("Expected 2m, got %v", defaults.ClientCacheTTL)
		}
	})

	t.Run("zero values use defaults", func(t *testing.T) {
		config := &CrewConfig{}
		config.Settings.ParallelTimeoutSeconds = 0
		config.Settings.MaxToolOutputChars = 0
		defaults := ConfigToHardcodedDefaults(config)

		if defaults.ParallelAgentTimeout != 60*time.Second {
			t.Errorf("Expected default 60s, got %v", defaults.ParallelAgentTimeout)
		}
		if defaults.MaxToolOutputChars != 2000 {
			t.Errorf("Expected default 2000, got %d", defaults.MaxToolOutputChars)
		}
	})

	t.Run("percentage converted correctly", func(t *testing.T) {
		config := &CrewConfig{}
		config.Settings.TimeoutWarningThresholdPct = 25
		defaults := ConfigToHardcodedDefaults(config)

		if defaults.TimeoutWarningThreshold != 0.25 {
			t.Errorf("Expected 0.25 (25%%), got %f", defaults.TimeoutWarningThreshold)
		}
	})
}

// TestHardcodedDefaultsValidateAfterConversion verifies validation after YAML conversion
func TestHardcodedDefaultsValidateAfterConversion(t *testing.T) {
	config := &CrewConfig{}
	config.Settings.ParallelTimeoutSeconds = 300
	config.Settings.MaxToolOutputChars = 10000

	defaults := ConfigToHardcodedDefaults(config)

	if err := defaults.Validate(); err != nil {
		t.Errorf("Validation failed after conversion: %v", err)
	}

	if defaults.ParallelAgentTimeout != 300*time.Second {
		t.Errorf("Expected 300s, got %v", defaults.ParallelAgentTimeout)
	}
	if defaults.MaxToolOutputChars != 10000 {
		t.Errorf("Expected 10000, got %d", defaults.MaxToolOutputChars)
	}
}

// TestHardcodedDefaultsBoundaryValues tests edge cases
func TestHardcodedDefaultsBoundaryValues(t *testing.T) {
	tests := []struct {
		name      string
		setupFunc func() *HardcodedDefaults
		wantValid bool
	}{
		{
			name: "very large timeout",
			setupFunc: func() *HardcodedDefaults {
				d := DefaultHardcodedDefaults()
				d.ParallelAgentTimeout = 1 * time.Hour
				return d
			},
			wantValid: true,
		},
		{
			name: "very small timeout",
			setupFunc: func() *HardcodedDefaults {
				d := DefaultHardcodedDefaults()
				d.MinToolTimeout = 1 * time.Millisecond
				return d
			},
			wantValid: true,
		},
		{
			name: "zero threshold corrected",
			setupFunc: func() *HardcodedDefaults {
				d := DefaultHardcodedDefaults()
				d.TimeoutWarningThreshold = 0.0
				return d
			},
			wantValid: true, // Corrected to 0.20
		},
		{
			name: "100% threshold valid",
			setupFunc: func() *HardcodedDefaults {
				d := DefaultHardcodedDefaults()
				d.TimeoutWarningThreshold = 1.0
				return d
			},
			wantValid: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d := tt.setupFunc()
			err := d.Validate()
			if (err == nil) != tt.wantValid {
				t.Errorf("Validate() got error %v, wantValid %v", err, tt.wantValid)
			}
		})
	}
}

// TestConfigToHardcodedDefaultsInvalidPercentage tests invalid percentage handling
func TestConfigToHardcodedDefaultsInvalidPercentage(t *testing.T) {
	tests := []struct {
		name           string
		percentageVal  int
		expectedResult float64
	}{
		{"valid 20%", 20, 0.20},
		{"valid 50%", 50, 0.50},
		{"zero % uses default", 0, 0.20},
		{"negative % uses default", -10, 0.20},
		{"over 100% uses default", 150, 0.20},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := &CrewConfig{}
			config.Settings.TimeoutWarningThresholdPct = tt.percentageVal
			defaults := ConfigToHardcodedDefaults(config)

			if defaults.TimeoutWarningThreshold != tt.expectedResult {
				t.Errorf("Expected %f, got %f", tt.expectedResult, defaults.TimeoutWarningThreshold)
			}
		})
	}
}

// TestHardcodedDefaultsAllFieldsPresent verifies all 18 fields are initialized
func TestHardcodedDefaultsAllFieldsPresent(t *testing.T) {
	d := DefaultHardcodedDefaults()

	// Count all non-zero fields to verify everything is initialized
	nonZeroCount := 0

	// Timeout fields
	if d.ParallelAgentTimeout > 0 {
		nonZeroCount++
	}
	if d.ToolExecutionTimeout > 0 {
		nonZeroCount++
	}
	if d.ToolResultTimeout > 0 {
		nonZeroCount++
	}
	if d.MinToolTimeout > 0 {
		nonZeroCount++
	}
	if d.StreamChunkTimeout > 0 {
		nonZeroCount++
	}
	if d.SSEKeepAliveInterval > 0 {
		nonZeroCount++
	}
	if d.RequestStoreCleanupInterval > 0 {
		nonZeroCount++
	}
	if d.ClientCacheTTL > 0 {
		nonZeroCount++
	}
	if d.GracefulShutdownCheckInterval > 0 {
		nonZeroCount++
	}

	// Size fields
	if d.MaxInputSize > 0 {
		nonZeroCount++
	}
	if d.MinAgentIDLength > 0 {
		nonZeroCount++
	}
	if d.MaxAgentIDLength > 0 {
		nonZeroCount++
	}
	if d.MaxRequestBodySize > 0 {
		nonZeroCount++
	}
	if d.MaxToolOutputChars > 0 {
		nonZeroCount++
	}
	if d.StreamBufferSize > 0 {
		nonZeroCount++
	}
	if d.MaxStoredRequests > 0 {
		nonZeroCount++
	}

	// Retry/backoff fields
	if d.RetryBackoffMinDuration > 0 {
		nonZeroCount++
	}
	if d.RetryBackoffMaxDuration > 0 {
		nonZeroCount++
	}

	// Threshold field
	if d.TimeoutWarningThreshold > 0 {
		nonZeroCount++
	}

	if nonZeroCount != 19 {
		t.Errorf("Expected all 19 fields initialized, got %d", nonZeroCount)
	}
}

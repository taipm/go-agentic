package crewai

import (
	"testing"
)

// ===== Phase 2: Signal Validation Tests =====

// TestValidateSignalsValidConfig tests that valid signal configuration passes
func TestValidateSignalsValidConfig(t *testing.T) {
	crew := &Crew{
		Agents: []*Agent{
			{ID: "teacher"},
			{ID: "reporter"},
		},
		Routing: &RoutingConfig{
			Signals: map[string][]RoutingSignal{
				"teacher": {
					{Signal: "[QUESTION_READY]", Target: "reporter"},
					{Signal: "[END_EXAM]", Target: ""}, // Termination signal
				},
				"reporter": {
					{Signal: "[REPORT_DONE]", Target: ""}, // Termination signal
				},
			},
		},
	}

	executor := &CrewExecutor{crew: crew}
	err := executor.ValidateSignals()
	if err != nil {
		t.Errorf("Valid signal config should pass validation, got error: %v", err)
	}
}

// TestValidateSignalsEmptySignalName tests that empty signal name is rejected
func TestValidateSignalsEmptySignalName(t *testing.T) {
	crew := &Crew{
		Agents: []*Agent{
			{ID: "teacher"},
		},
		Routing: &RoutingConfig{
			Signals: map[string][]RoutingSignal{
				"teacher": {
					{Signal: "", Target: "reporter"}, // Empty signal!
				},
			},
		},
	}

	executor := &CrewExecutor{crew: crew}
	err := executor.ValidateSignals()
	if err == nil {
		t.Error("Should reject empty signal name")
	}
	if err != nil && err.Error() != "agent 'teacher' has signal with empty name - signal must be in [NAME] format" {
		t.Errorf("Wrong error message: %v", err)
	}
}

// TestValidateSignalsInvalidFormat tests that signals must match [NAME] format
func TestValidateSignalsInvalidFormat(t *testing.T) {
	tests := []struct {
		name      string
		signal    string
		wantError bool
	}{
		{"Valid bracket", "[END_EXAM]", false},
		{"Valid with space", "[END EXAM]", false},
		{"Missing opening bracket", "END_EXAM]", true},
		{"Missing closing bracket", "[END_EXAM", true},
		{"No brackets", "END_EXAM", true},
		{"Empty brackets", "[]", true}, // Invalid: must have content inside brackets
		{"Too short", "[", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			crew := &Crew{
				Agents: []*Agent{
					{ID: "teacher"},
					{ID: "reporter"},
				},
				Routing: &RoutingConfig{
					Signals: map[string][]RoutingSignal{
						"teacher": {
							{Signal: tt.signal, Target: "reporter"},
						},
					},
				},
			}

			executor := &CrewExecutor{crew: crew}
			err := executor.ValidateSignals()

			if tt.wantError && err == nil {
				t.Errorf("Expected error for signal %q, got nil", tt.signal)
			}
			if !tt.wantError && err != nil {
				t.Errorf("Expected no error for signal %q, got: %v", tt.signal, err)
			}
		})
	}
}

// TestValidateSignalsUnknownTarget tests that signal targets must exist
func TestValidateSignalsUnknownTarget(t *testing.T) {
	crew := &Crew{
		Agents: []*Agent{
			{ID: "teacher"},
			{ID: "reporter"},
		},
		Routing: &RoutingConfig{
			Signals: map[string][]RoutingSignal{
				"teacher": {
					{Signal: "[NEXT]", Target: "unknown_agent"}, // Target doesn't exist!
				},
			},
		},
	}

	executor := &CrewExecutor{crew: crew}
	err := executor.ValidateSignals()
	if err == nil {
		t.Error("Should reject signal targeting unknown agent")
	}
	if err != nil {
		expected := "agent 'teacher' emits signal '[NEXT]' targeting unknown agent 'unknown_agent' - target must be empty (terminate) or valid agent ID"
		if err.Error() != expected {
			t.Errorf("Wrong error message:\nGot:      %v\nExpected: %v", err.Error(), expected)
		}
	}
}

// TestValidateSignalsEmptyTargetTermination tests that empty target (termination) is valid
func TestValidateSignalsEmptyTargetTermination(t *testing.T) {
	crew := &Crew{
		Agents: []*Agent{
			{ID: "teacher"},
		},
		Routing: &RoutingConfig{
			Signals: map[string][]RoutingSignal{
				"teacher": {
					{Signal: "[END]", Target: ""}, // Empty target = termination, should be valid
				},
			},
		},
	}

	executor := &CrewExecutor{crew: crew}
	err := executor.ValidateSignals()
	if err != nil {
		t.Errorf("Empty target should be valid for termination signal, got error: %v", err)
	}
}

// TestValidateSignalsNoRouting tests that missing routing doesn't cause error
func TestValidateSignalsNoRouting(t *testing.T) {
	crew := &Crew{
		Agents: []*Agent{
			{ID: "teacher"},
		},
		Routing: nil, // No routing
	}

	executor := &CrewExecutor{crew: crew}
	err := executor.ValidateSignals()
	if err != nil {
		t.Errorf("Missing routing should not cause error, got: %v", err)
	}
}

// TestValidateSignalsEmptySignalMap tests that empty signal map doesn't cause error
func TestValidateSignalsEmptySignalMap(t *testing.T) {
	crew := &Crew{
		Agents: []*Agent{
			{ID: "teacher"},
		},
		Routing: &RoutingConfig{
			Signals: map[string][]RoutingSignal{}, // Empty signal map
		},
	}

	executor := &CrewExecutor{crew: crew}
	err := executor.ValidateSignals()
	if err != nil {
		t.Errorf("Empty signal map should not cause error, got: %v", err)
	}
}

// TestValidateSignalsMultipleSignalsPerAgent tests validation with multiple signals per agent
func TestValidateSignalsMultipleSignalsPerAgent(t *testing.T) {
	crew := &Crew{
		Agents: []*Agent{
			{ID: "teacher"},
			{ID: "reporter"},
			{ID: "auditor"},
		},
		Routing: &RoutingConfig{
			Signals: map[string][]RoutingSignal{
				"teacher": {
					{Signal: "[QUESTIONS_READY]", Target: "reporter"},
					{Signal: "[SKIP_REPORT]", Target: "auditor"},
					{Signal: "[END_EXAM]", Target: ""}, // Termination
				},
				"reporter": {
					{Signal: "[REPORT_DONE]", Target: "auditor"},
					{Signal: "[REPORT_SKIP]", Target: ""}, // Termination
				},
			},
		},
	}

	executor := &CrewExecutor{crew: crew}
	err := executor.ValidateSignals()
	if err != nil {
		t.Errorf("Valid multi-signal config should pass, got error: %v", err)
	}
}

// TestValidateSignalsVietnamseSignals tests Vietnamese signal names
func TestValidateSignalsVietnamseSignals(t *testing.T) {
	crew := &Crew{
		Agents: []*Agent{
			{ID: "giao_vien"},
			{ID: "bao_cao"},
		},
		Routing: &RoutingConfig{
			Signals: map[string][]RoutingSignal{
				"giao_vien": {
					{Signal: "[KẾT_THÚC_THI]", Target: ""}, // Vietnamese signal
				},
				"bao_cao": {
					{Signal: "[BÁO_CÁO_XONG]", Target: "giao_vien"}, // Vietnamese signal
				},
			},
		},
	}

	executor := &CrewExecutor{crew: crew}
	err := executor.ValidateSignals()
	if err != nil {
		t.Errorf("Vietnamese signals should be valid, got error: %v", err)
	}
}

// TestValidateSignalsCaseSensitivity tests that signal names are case-sensitive in validation
func TestValidateSignalsCaseSensitivity(t *testing.T) {
	crew := &Crew{
		Agents: []*Agent{
			{ID: "agent1"},
		},
		Routing: &RoutingConfig{
			Signals: map[string][]RoutingSignal{
				"agent1": {
					{Signal: "[END_EXAM]", Target: ""},
					{Signal: "[end_exam]", Target: ""}, // Different case
				},
			},
		},
	}

	executor := &CrewExecutor{crew: crew}
	err := executor.ValidateSignals()
	// Both should be valid as separate signals
	if err != nil {
		t.Errorf("Signals with different cases should both be valid, got error: %v", err)
	}
}

// TestIsSignalFormatValid tests the signal format validation function directly
func TestIsSignalFormatValid(t *testing.T) {
	tests := []struct {
		signal    string
		wantValid bool
	}{
		{signal: "[END]", wantValid: true},
		{signal: "[A]", wantValid: true},
		{signal: "[END_EXAM]", wantValid: true},
		{signal: "[READY TO GO]", wantValid: true},
		{signal: "[KẾT THÚC]", wantValid: true},
		{signal: "END]", wantValid: false},
		{signal: "[END", wantValid: false},
		{signal: "END", wantValid: false},
		{signal: "[]", wantValid: false}, // Invalid: must have content inside brackets
		{signal: "[", wantValid: false},
		{signal: "]", wantValid: false},
		{signal: "", wantValid: false},
	}

	for _, tt := range tests {
		t.Run("signal="+tt.signal, func(t *testing.T) {
			got := isSignalFormatValid(tt.signal)
			if got != tt.wantValid {
				t.Errorf("isSignalFormatValid(%q) = %v, want %v", tt.signal, got, tt.wantValid)
			}
		})
	}
}

// TestCountTotalSignals tests counting total signals across all agents
func TestCountTotalSignals(t *testing.T) {
	signals := map[string][]RoutingSignal{
		"agent1": {
			{Signal: "[A]", Target: "agent2"},
			{Signal: "[B]", Target: ""},
		},
		"agent2": {
			{Signal: "[C]", Target: "agent1"},
		},
		"agent3": {}, // No signals
	}

	count := countTotalSignals(signals)
	expected := 3
	if count != expected {
		t.Errorf("countTotalSignals() = %d, want %d", count, expected)
	}
}

// TestValidateSignalsComplexWorkflow tests a realistic complex workflow
func TestValidateSignalsComplexWorkflow(t *testing.T) {
	crew := &Crew{
		Agents: []*Agent{
			{ID: "coordinator"},
			{ID: "researcher"},
			{ID: "analyst"},
			{ID: "writer"},
		},
		Routing: &RoutingConfig{
			Signals: map[string][]RoutingSignal{
				"coordinator": {
					{Signal: "[START_RESEARCH]", Target: "researcher"},
					{Signal: "[CANCEL]", Target: ""}, // User cancellation = terminate
				},
				"researcher": {
					{Signal: "[RESEARCH_COMPLETE]", Target: "analyst"},
					{Signal: "[RESEARCH_FAILED]", Target: "coordinator"}, // Back to coordinator
				},
				"analyst": {
					{Signal: "[ANALYSIS_DONE]", Target: "writer"},
					{Signal: "[NEEDS_MORE_DATA]", Target: "researcher"}, // Loop back
				},
				"writer": {
					{Signal: "[DOCUMENT_READY]", Target: ""}, // Final termination
					{Signal: "[REVISE_NEEDED]", Target: "analyst"}, // Loop back
				},
			},
		},
	}

	executor := &CrewExecutor{crew: crew}
	err := executor.ValidateSignals()
	if err != nil {
		t.Errorf("Complex valid workflow should pass validation, got error: %v", err)
	}
}

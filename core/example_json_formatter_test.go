package crewai

import (
	"encoding/json"
	"fmt"
)

// ExampleToJSON demonstrates JSON output from ValidationErrorFormatter
func ExampleConfigValidator_ToJSON() {
	// Create a config with validation errors
	config := &CrewConfig{
		EntryPoint: "bad_entry", // Invalid - not in agents list
		Agents:     []string{"orchestrator", "executor"},
	}
	config.Settings.MaxHandoffs = 5
	config.Settings.MaxRounds = 10
	config.Settings.TimeoutSeconds = 300

	agents := map[string]*AgentConfig{
		"orchestrator": {
			ID:          "orchestrator",
			Name:        "Orchestrator",
			Model:       "gpt-4o",
			Temperature: 0.7,
			IsTerminal:  false,
		},
		"executor": {
			ID:          "executor",
			Name:        "Executor",
			Model:       "gpt-4o",
			Temperature: 0.3,
			IsTerminal:  true,
		},
	}

	// Validate configuration
	validator := NewConfigValidator(config, agents)
	validator.ValidateAll()

	// Get JSON output
	jsonData, _ := validator.ToJSON()

	// Pretty print for demonstration
	var resp ErrorResponse
	json.Unmarshal(jsonData, &resp)

	fmt.Printf("Validation Success: %v\n", resp.Success)
	fmt.Printf("Total Errors: %d\n", resp.Summary.TotalErrors)
	fmt.Printf("Total Warnings: %d\n", resp.Summary.TotalWarnings)
	fmt.Printf("\nJSON Output:\n%s\n", string(jsonData))

	// Output:
	// Validation Success: false
	// Total Errors: 1
	// Total Warnings: 0
}

// ExampleValidConfigToJSON demonstrates JSON output for valid configuration
func ExampleConfigValidator_ToJSON_valid() {
	// Create a valid configuration
	config := &CrewConfig{
		EntryPoint: "orchestrator",
		Agents:     []string{"orchestrator", "executor"},
	}
	config.Settings.MaxHandoffs = 5
	config.Settings.MaxRounds = 10
	config.Settings.TimeoutSeconds = 300

	agents := map[string]*AgentConfig{
		"orchestrator": {
			ID:          "orchestrator",
			Name:        "Orchestrator",
			Model:       "gpt-4o",
			Temperature: 0.7,
			IsTerminal:  false,
		},
		"executor": {
			ID:          "executor",
			Name:        "Executor",
			Model:       "gpt-4o",
			Temperature: 0.3,
			IsTerminal:  true,
		},
	}

	validator := NewConfigValidator(config, agents)
	validator.ValidateAll()

	// Get JSON output
	jsonData, _ := validator.ToJSON()

	var resp ErrorResponse
	json.Unmarshal(jsonData, &resp)

	fmt.Printf("Validation Success: %v\n", resp.Success)
	fmt.Printf("Is Valid: %v\n", resp.Summary.IsValid)
	fmt.Printf("Errors: %d, Warnings: %d\n", resp.Summary.TotalErrors, resp.Summary.TotalWarnings)

	// Output:
	// Validation Success: true
	// Is Valid: true
	// Errors: 0, Warnings: 0
}

package internal

import (
	"github.com/taipm/go-crewai"
)

// CreateITSupportCrew creates a complete IT Support crew
func CreateITSupportCrew() *crewai.Crew {
	// Define tools
	tools := createITSupportTools()

	// Create agents
	orchestrator := &crewai.Agent{
		ID:          "orchestrator",
		Name:        "Orchestrator",
		Role:        "System coordinator and entry point",
		Backstory:   "You are the entry point for IT support requests. Analyze the problem and decide if you need more information before proceeding to execution.",
		Model:       "gpt-4o",
		Tools:       []*crewai.Tool{},
		Temperature: 0.7,
		IsTerminal:  false,
	}

	clarifier := &crewai.Agent{
		ID:          "clarifier",
		Name:        "Clarifier",
		Role:        "Information gatherer",
		Backstory:   "You specialize in gathering detailed information about IT issues. Ask clarifying questions to understand the problem better.",
		Model:       "gpt-4o",
		Tools:       []*crewai.Tool{},
		Temperature: 0.7,
		IsTerminal:  false,
	}

	executor := &crewai.Agent{
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
	crew := &crewai.Crew{
		Agents:      []*crewai.Agent{orchestrator, clarifier, executor},
		MaxRounds:   10,
		MaxHandoffs: 5,
	}

	return crew
}

// GetAllITSupportTools returns all IT support tools as a map
func GetAllITSupportTools() map[string]*crewai.Tool {
	tools := createITSupportTools()
	toolMap := make(map[string]*crewai.Tool)
	for _, tool := range tools {
		toolMap[tool.Name] = tool
	}
	return toolMap
}

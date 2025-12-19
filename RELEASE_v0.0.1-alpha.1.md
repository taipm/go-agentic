# Release: v0.0.1-alpha.1

## Release Date
December 19, 2025

## Library Information
**Module**: `github.com/taipm/go-agentic`
**Go Version**: 1.25.5
**Status**: Alpha Release

## Included Components

### Core Library Files
- ✅ `types.go` - Type definitions for agents, teams, tools, messages, and configurations
- ✅ `team.go` - Team orchestration and execution engine
- ✅ `agent.go` - Agent implementation and lifecycle management
- ✅ `config.go` - YAML configuration loading and parsing
- ✅ `http.go` - HTTP server with SSE streaming support
- ✅ `streaming.go` - Stream event handling
- ✅ `html_client.go` - HTML client for testing and monitoring
- ✅ `tests.go` - Testing framework and test scenarios
- ✅ `report.go` - Execution reporting and analytics

### Key Features
- ✅ Multi-agent orchestration system
- ✅ Real-time SSE streaming for agent events
- ✅ Intelligent agent routing and handoff with configuration-driven signals
- ✅ YAML-based configuration for agents and teams
- ✅ Comprehensive testing framework with predefined scenarios
- ✅ Production-ready error handling
- ✅ OpenAI GPT integration (v3.15.0)
- ✅ Support for custom tools and agents
- ✅ Backward compatibility aliases (Crew, CrewConfig, CrewResponse)

## Dependencies

### Direct Dependencies
- `github.com/openai/openai-go/v3` v3.15.0 - OpenAI API SDK
- `gopkg.in/yaml.v3` v3.0.1 - YAML configuration support

### Indirect Dependencies
- `github.com/tidwall/gjson` v1.18.0
- `github.com/tidwall/match` v1.2.0
- `github.com/tidwall/pretty` v1.2.1
- `github.com/tidwall/sjson` v1.2.5

## Security Checklist

### Completed
- ✅ All environment files (.env) verified and cleared of API keys
- ✅ Created comprehensive .gitignore
- ✅ Security scan for exposed credentials
- ✅ Source code review for sensitive data
- ✅ Git history cleaned of sensitive information

### Files Verified
- ✅ `_old_files/.env` - Cleared of sensitive data
- ✅ `go-crewai/.env` - Cleared of sensitive data
- ✅ All `.go` files - No hardcoded credentials
- ✅ Examples - Properly documented with placeholder credentials

## Build Status

### Library Build
✅ Successful with Go 1.25.5

### Examples Build
✅ All 4 examples build successfully:
- customer-service
- data-analysis
- it-support
- research-assistant

### Quality Checks
✅ Code compiles without warnings
✅ All modules tidy and locked

## Backward Compatibility

The library maintains full backward compatibility with previous "Crew" naming through type aliases:
```go
type Crew = Team                    // Deprecated: Use Team instead
type CrewConfig = TeamConfig        // Deprecated: Use TeamConfig instead
type CrewResponse = TeamResponse    // Deprecated: Use TeamResponse instead
```

Deprecated function wrappers are also provided:
```go
func NewCrewExecutor(...) *TeamExecutor // Deprecated: Use NewTeamExecutor instead
func LoadCrewConfig(...) *TeamConfig    // Deprecated: Use LoadTeamConfig instead
```

## Breaking Changes
None - this is the initial alpha release.

## Known Limitations
- This is an alpha release - API may change before stable release
- Some advanced routing features still in development
- Performance optimization pending

## Installation

```bash
go get github.com/taipm/go-agentic@v0.0.1-alpha.1
```

## Usage Example

```go
package main

import (
    "context"
    "github.com/taipm/go-agentic"
)

func main() {
    // Create agents
    agent := &agentic.Agent{
        ID:    "agent-1",
        Name:  "Agent One",
        Role:  "Primary agent",
        Model: "gpt-4o",
        Tools: tools,
    }

    // Create team
    team := &agentic.Team{
        Agents:      []*agentic.Agent{agent},
        MaxRounds:   10,
        MaxHandoffs: 5,
    }

    // Execute
    executor := agentic.NewTeamExecutor(team, apiKey)
    response, err := executor.Execute(context.Background(), "Your task here")
}
```

## Documentation
See `go-agentic/docs/` directory for comprehensive documentation:
- `LIBRARY_README.md` - Library overview and features
- `LIBRARY_STRUCTURE.md` - Architecture and design patterns
- `ARCHITECTURE.txt` - Visual diagrams and flows
- `START_HERE.md` - Quick start guide

## Support and Issues
For issues, questions, or feedback, refer to the project repository.

## License
See LICENSE file in repository root.

---
**Release Created**: December 19, 2025
**Release Manager**: Claude Code Agent
**Status**: Ready for Alpha Testing

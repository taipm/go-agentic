# go-agentic Examples

Complete, production-ready examples demonstrating how to use go-agentic for different use cases.

## Quick Start

Each example is a self-contained project. To run any example:

```bash
cd examples/<example-name>
cp .env.example .env  # or copy from parent
export OPENAI_API_KEY="your-key-here"
go run main.go
```

## Examples Overview

### 1. IT Support 🆘

A system for IT diagnostics and support ticket management.

**Use Case**: Automate infrastructure monitoring and support requests

**Agents**:
- Orchestrator: Routes requests
- Clarifier: Gathers system details
- Executor: Runs diagnostics with 12+ tools

**Tools**: CPU, Memory, Disk, Network, Services, DNS, Processes, etc.

```bash
cd it-support
go run main.go
```

[Learn more](./it-support/README.md)

### 2. Customer Service 👥

A multi-agent system for handling customer support interactions.

**Use Case**: Automated customer service with knowledge base, refunds, account management

**Agents**:
- Classifier: Prioritizes requests
- Resolver: Solves with tools
- Responder: Professional responses

**Tools**: Knowledge base search, account status, order history, refunds, support tickets

```bash
cd customer-service
go run main.go
```

[Learn more](./customer-service/README.md)

### 3. Data Analysis 📊

A team of data analysts for insights and reporting.

**Use Case**: Automated data analysis, reporting, and forecasting

**Agents**:
- Researcher: Gathers data
- Analyst: Statistical analysis
- Synthesizer: Generates insights

**Tools**: Dataset analysis, trend identification, correlation, anomaly detection, forecasting

```bash
cd data-analysis
go run main.go
```

[Learn more](./data-analysis/README.md)

### 4. Research Assistant 🔬

An AI research team for literature review and analysis.

**Use Case**: Academic research, literature review, methodology analysis

**Agents**:
- Investigator: Searches databases
- Analyst: Analyzes findings
- Documentor: Creates reports

**Tools**: Database search, citation analysis, theory comparison, author identification, gap analysis

```bash
cd research-assistant
go run main.go
```

[Learn more](./research-assistant/README.md)

## Architecture Pattern

All examples follow the same multi-agent orchestration pattern:

```
User Input
    ↓
Orchestrator Agent (entry point, analyzes request)
    ↓
Optional: Clarifier Agent (gathers more info if needed)
    ↓
Executor Agent (performs work with tools, is terminal)
    ↓
User Response
```

## Common Features

### Tool Integration

Each example includes domain-specific tools:
- IT Support: System commands, diagnostics
- Customer Service: Database queries, business logic
- Data Analysis: Statistical functions
- Research: API integrations

### Interactive CLI

All examples provide an interactive command-line interface:

```
You: [ask a question]
Agent: [response based on multi-agent analysis]
```

### Configuration

Examples can be customized via:
- `config/crew.yaml`: Agent routing and settings
- Agent definitions in code
- Tool implementations

## Directory Structure

```
examples/
├── it-support/
│   ├── main.go
│   ├── config/
│   ├── example_it_support.go
│   └── README.md
├── customer-service/
│   ├── main.go
│   └── README.md
├── data-analysis/
│   ├── main.go
│   └── README.md
├── research-assistant/
│   ├── main.go
│   └── README.md
└── README.md
```

## Creating Your Own Example

1. **Create directory structure**:
```bash
mkdir examples/my-example
cd examples/my-example
```

2. **Create main.go** with your crew:
```go
package main

import "github.com/taipm/go-agentic"

func main() {
    // Define your crew
    crew := &agentic.Crew{
        Agents: []*agentic.Agent{
            // Your agents
        },
    }
    
    // Execute
    executor := agentic.NewCrewExecutor(crew, apiKey)
    executor.Execute(context.Background(), userInput)
}
```

3. **Define your agents** with custom tools

4. **Add README.md** documenting your example

## Best Practices

### Tool Design

- Keep tools focused on single responsibility
- Provide clear parameter documentation
- Return meaningful error messages
- Use context for cancellation

### Agent Configuration

- Use specific, detailed roles
- Provide rich backstory context
- Match temperature to use case (0.5 for analysis, 0.7 for creative)
- Set `IsTerminal: true` on final agent only

### Error Handling

- Gracefully handle tool failures
- Provide user-friendly error messages
- Log detailed errors for debugging
- Implement retry logic where appropriate

## Extending Examples

### Add Custom Tools

```go
tool := &agentic.Tool{
    Name:        "MyTool",
    Description: "What this does",
    Parameters: map[string]interface{}{
        "type": "object",
        "properties": map[string]interface{}{
            "param": map[string]interface{}{
                "type": "string",
            },
        },
    },
    Handler: func(ctx context.Context, args map[string]interface{}) (string, error) {
        // Implementation
        return "result", nil
    },
}
```

### Modify Agent Behavior

Edit agent definitions:
```go
agent := &agentic.Agent{
    Role:        "Your custom role",
    Backstory:   "Detailed background...",
    Temperature: 0.7, // Adjust creativity
    // ... other fields
}
```

### Integration with External Systems

Connect to APIs, databases, services:
```go
Handler: func(ctx context.Context, args map[string]interface{}) (string, error) {
    // Call external API or database
    data := fetchFromExternal(args)
    return formatResult(data), nil
}
```

## Performance Tips

- **Batch operations**: Group related calls
- **Cache results**: Store frequently used data
- **Async tools**: Implement non-blocking operations
- **Monitor costs**: Track token usage

## Troubleshooting

### Agent not responding
- Check `IsTerminal` is set correctly
- Verify agent tools are properly configured
- Check OpenAI API key and quota

### Tools not being called
- Ensure tool description is clear
- Verify tool parameters are documented
- Check tool handlers for errors

### Slow responses
- Review agent temperature settings
- Check for blocking operations
- Monitor API latency

## Support

For issues or questions:
1. Check example-specific README
2. Review parent directory documentation
3. Check tool implementation examples
4. Test with simpler input first

## Contributing

To contribute new examples:
1. Create well-documented code
2. Include comprehensive README
3. Add test scenarios
4. Follow existing patterns

## License

All examples are part of go-agentic and follow the same license.

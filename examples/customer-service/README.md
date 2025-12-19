# Customer Service Example

A complete example of using go-agentic for customer support and service interactions.

## Overview

This example demonstrates a multi-agent system that handles customer service:
- **Request Classifier**: Classifies and prioritizes customer issues
- **Issue Resolver**: Solves issues using available tools and knowledge
- **Support Specialist**: Provides empathetic and professional responses

## Features

- Knowledge base search
- Account and order management
- Refund processing
- Support ticket creation
- Customer data management
- 7+ customer service tools

## Quick Start

### 1. Setup

```bash
cd examples/customer-service
cp ../.env.example .env
# Edit .env and add your OPENAI_API_KEY
```

### 2. Run

```bash
go run main.go
```

### 3. Try It

Example requests:
- "I want to reset my password"
- "How do I track my order?"
- "I'd like a refund for order ABC123"
- "What's your billing policy?"
- "I have a technical issue with the API"

## Available Tools

- **SearchKnowledgeBase** - Find help articles
- **CheckAccountStatus** - View account info
- **ViewOrderHistory** - See past orders
- **IssueRefund** - Process refunds
- **ResetPassword** - Password reset links
- **CreateTicket** - Escalate to support team

## Configuration

Customize the crew behavior by modifying the agent definitions in the code or create a `config.yaml`.

## Architecture

```
Customer Service Crew
├── Classifier (request analysis)
├── Resolver (issue solving with tools)
└── Responder (customer-facing response)
    ├── Knowledge Management
    ├── Account Management
    ├── Order Processing
    └── Support Escalation
```

## Adding Custom Tools

1. Define your tool function:

```go
func myCustomTool(ctx context.Context, args map[string]interface{}) (string, error) {
    // Implementation
    return "result", nil
}
```

2. Add to the tools list in `CreateCustomerServiceCrew()`

3. Update the tool definitions with proper parameters

## Testing

To test the customer service crew:

```bash
# Add test scenario to main.go
```

## Extending

To add new capabilities:

1. Define new tools in the handler functions
2. Update crew configuration
3. Add agent-specific prompts for better context
4. Test with real customer scenarios

## Learn More

- See parent directory README for general go-agentic information
- Review tool handlers for implementation examples
- Check agent configuration for customization options

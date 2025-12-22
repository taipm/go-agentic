# Hello Crew - Your First Agent Example

Welcome to **Hello Crew**, the simplest introduction to the go-agentic framework. This example demonstrates a minimal crew with just one agent.

## What You'll Learn

- How to set up a crew with a single agent
- How to create agent and crew configurations
- How to run your agent in CLI mode
- How to run your agent in server mode
- How to extend the agent with tools

## Project Structure

```
00-hello-crew/
├── cmd/
│   └── main.go              # Entry point for CLI and server modes
├── internal/
│   └── hello.go             # Hello agent definition
├── config/
│   ├── crew.yaml            # Crew configuration
│   └── agents/
│       └── hello-agent.yaml # Agent configuration
├── go.mod                   # Go module definition
├── .env.example             # Environment variables template
├── Makefile                 # Build and run commands
└── README.md               # This file
```

## Quick Start (3 Minutes)

### 1. Set Up Your Environment

```bash
# Copy environment template
cp .env.example .env

# Edit .env and add your OpenAI API key
# OPENAI_API_KEY=sk-...
```

### 2. Run in CLI Mode

```bash
make run
```

Then type your message and press Enter:

```
> Hello, how are you?
Response: I'm doing great, thank you for asking! I'm here to help you with any questions or tasks you might have. What can I assist you with today?

> exit
```

### 3. Run in Server Mode

```bash
make run-server
```

In another terminal:

```bash
curl -X POST http://localhost:8081/execute \
  -H "Content-Type: application/json" \
  -d '{"input": "Hello, how are you?"}'
```

## Understanding the Code (10 Minutes)

### Key Files

#### 1. `cmd/main.go` - Entry Point

The main.go file handles two modes:

**CLI Mode** (default):
- Reads user input from stdin
- Displays agent responses
- Simple interactive loop

**Server Mode** (`-server` flag):
- Starts HTTP server on port 8081
- Accepts POST requests to `/execute`
- Returns JSON responses

```go
// Usage:
// CLI mode: go run cmd/main.go
// Server mode: go run cmd/main.go -server -port 8081
```

#### 2. `config/crew.yaml` - Crew Configuration

This defines your crew (team of agents):

```yaml
name: hello-crew                    # Name of your crew
description: A minimal crew...      # What it does
agents:
  - hello-agent                     # List of agents (just one here)
tasks:
  - name: respond-to-user          # Task definition
    description: Respond...
    agent: hello-agent
```

#### 3. `config/agents/hello-agent.yaml` - Agent Configuration

This defines the Hello agent:

```yaml
name: hello-agent
role: Friendly Assistant
description: A simple and friendly assistant...
backstory: |
  You are a warm and welcoming assistant...
max_iterations: 5                   # How many times to think/act
temperature: 0.7                    # 0=deterministic, 1=creative
model: gpt-4-turbo                  # LLM model to use
tools: []                           # No tools for now
```

#### 4. `internal/hello.go` - Agent Definition

Defines the agent structure:

```go
type HelloAgent struct {
    Name        string
    Role        string
    Description string
    Backstory   string
}
```

### How It Works

1. **User Input**: You send a message
2. **Executor**: Receives the message and routes it to the crew
3. **Hello Agent**: Receives the message, thinks about it using GPT-4
4. **Response**: Returns the agent's response to you

## Customize Your Crew (15 Minutes)

### Change the Agent's Personality

Edit `config/agents/hello-agent.yaml`:

```yaml
backstory: |
  You are a pirate captain with a sense of humor.
  You respond to questions with pirate dialect.
```

Then run again and see how it changes!

### Add More Information to the Backstory

```yaml
backstory: |
  You are a helpful assistant specializing in software development.
  You provide clear, concise answers with code examples when relevant.
  You ask clarifying questions when needed.
```

### Change the Model

```yaml
model: gpt-4o              # Faster and cheaper
# or
model: gpt-4-turbo        # More capable
```

### Adjust Temperature (Creativity)

```yaml
temperature: 0.3           # More focused and consistent
# or
temperature: 0.9           # More creative and varied
```

## Build Your Own Crew

Ready to extend this? Here's what you can do next:

### 1. Add More Agents

Create `config/agents/researcher.yaml`:

```yaml
name: researcher
role: Research Analyst
description: Researches topics and provides detailed information
backstory: You are an expert researcher...
max_iterations: 5
temperature: 0.5
model: gpt-4-turbo
tools: []
```

Update `config/crew.yaml`:

```yaml
agents:
  - hello-agent
  - researcher
```

### 2. Add Tools

Tools allow agents to take actions. Example: web search, file operations, etc.

```yaml
tools:
  - name: search
    description: Search the web for information
    schema: {...}
```

### 3. Add Tasks

Define specific tasks for each agent:

```yaml
tasks:
  - name: greet-user
    description: Greet the user warmly
    agent: hello-agent
  - name: research-topic
    description: Research the topic if needed
    agent: researcher
```

## Make Commands

```bash
make run              # Run in CLI mode
make run-server       # Run in server mode
make build            # Build the binary
make clean            # Clean up compiled files
```

## Troubleshooting

### Issue: "OPENAI_API_KEY not set"

**Solution**: Create a `.env` file in the project root:

```bash
cp .env.example .env
# Edit .env and add your OpenAI API key
export $(cat .env | xargs)
go run cmd/main.go
```

### Issue: "failed to load crew config"

**Solution**: Make sure you're running from the project root:

```bash
cd examples/00-hello-crew
go run cmd/main.go
```

### Issue: Port 8081 already in use

**Solution**: Use a different port:

```bash
go run cmd/main.go -server -port 8082
```

## Next Steps

1. **Run Hello Crew**: `make run`
2. **Customize**: Edit `config/agents/hello-agent.yaml`
3. **Understand**: Read the code in `cmd/main.go` and understand the flow
4. **Extend**: Add more agents or tools (see "Build Your Own Crew" section)
5. **Learn More**: Check out the [IT Support Example](../01-it-support/) for a more complex crew

## Key Concepts

- **Agent**: An AI entity with a specific role and backstory
- **Crew**: A team of agents working together
- **Task**: Work to be done by an agent
- **Tool**: An action an agent can take (web search, run code, etc.)
- **Executor**: The engine that runs your crew
- **Message History**: Conversation history between user and agents

## Cost Optimization

This example uses OpenAI's GPT-4 Turbo. To reduce costs:

1. Change `model: gpt-4-turbo` to `model: gpt-4o` in agent config (50% cheaper)
2. Lower `temperature: 0.7` to `temperature: 0.3` (faster responses)
3. Use `max_iterations: 3` instead of 5 (fewer API calls)

With these changes, you can run this example for ~$0.01 per request instead of $0.05.

## Questions?

- Check the [Getting Started Guide](../../docs/GUIDE_GETTING_STARTED.md)
- Review the [API Documentation](../../docs/API_REFERENCE.md)
- Ask on our GitHub discussions

Happy coding! 🚀

# Simple Chat Example - Two Agent Conversation

This is the **simplest example** of using go-agentic with just **2 agents** that automatically have a conversation with each other.

## Features

- 🤖 **2 Simple Agents**: Enthusiast and Expert
- 💬 **Automatic Conversation**: Agents take turns discussing topics
- 🎯 **Minimal Setup**: No complex tools or configurations
- 📚 **Easy to Understand**: Perfect for learning the basics

## Architecture

```
┌─────────────────────────────────────────┐
│         Initial Topic                    │
└──────────────┬──────────────────────────┘
               │
        ┌──────▼──────┐
        │  Enthusiast  │ (Asks questions)
        └──────┬──────┘
               │
        ┌──────▼──────┐
        │   Expert     │ (Provides answers)
        └──────┬──────┘
               │
        ┌──────▼──────┐
        │  Enthusiast  │ (Follows up)
        └──────┬──────┘
               │
        ┌──────▼──────┐
        │   Expert     │ (Final response)
        └──────┬──────┘
               │
        ┌──────▼──────────┐
        │  Response Ready  │
        └─────────────────┘
```

## How It Works

1. **Enthusiast Agent** - Curious learner who asks insightful questions
   - Role: Asks the Expert thoughtful questions
   - Behavior: Explores ideas and engages in discussion
   - IsTerminal: `false` (can pass to next agent)

2. **Expert Agent** - Subject matter expert
   - Role: Provides comprehensive answers
   - Behavior: Shares expertise and knowledge
   - IsTerminal: `true` (final response)

## Setup

1. **Copy environment template:**
   ```bash
   cp .env.example .env
   ```

2. **Add your OpenAI API key:**
   ```bash
   # Edit .env and replace with your actual key
   OPENAI_API_KEY=sk-proj-your-actual-api-key-here
   ```

3. **Verify .env is NOT committed:**
   ```bash
   git status
   # .env should NOT appear in the list
   ```

## Running the Example

```bash
# Using go run
go run main.go

# Or build and run
go build
./simple-chat
```

## Expected Output

```
🤖 Simple Multi-Agent Chat System
==================================================

📌 Topic: What are the best practices for writing Go code?
--------------------------------------------------
✅ Final Response:
[Expert's comprehensive answer about Go best practices]

📌 Topic: How can AI agents improve software development?
--------------------------------------------------
✅ Final Response:
[Expert's thoughts on AI agents in development]

...
```

## Conversation Flow

```
User: "What are the best practices for writing Go code?"
  ↓
Enthusiast: "That's a great topic! What are the most important principles?"
  ↓
Expert: "Here are the key practices: [detailed answer]"
  ↓
Enthusiast: "Can you elaborate on error handling?"
  ↓
Expert: "Absolutely! Error handling in Go... [detailed response]"
```

## Customization

### Change Topics
Edit the `topics` slice in `main.go`:

```go
topics := []string{
    "Your custom topic here",
    "Another topic",
    "And more...",
}
```

### Change Agent Behavior
Modify the agent properties:

```go
enthusiast := &agentic.Agent{
    ID:          "enthusiast",
    Name:        "Enthusiast",
    Role:        "Your custom role",
    Backstory:   "Your custom backstory",
    Temperature: 0.8,  // Higher = more creative
    IsTerminal:  false,
}
```

### Adjust Conversation Length
```go
crew := &agentic.Crew{
    Agents:      []*agentic.Agent{enthusiast, expert},
    MaxRounds:   5,      // More rounds = longer conversation
    MaxHandoffs: 3,      // More handoffs = more back-and-forth
}
```

## Troubleshooting

### "OPENAI_API_KEY environment variable not set"
- Make sure you've created `.env` file from `.env.example`
- Verify the API key is correct
- Make sure the file is in the same directory as the executable

### "setup failed" error
- Check that your go-agentic library is properly installed
- Verify the go.mod replace directive points to the correct path

### API Errors
- Verify your OpenAI API key is valid
- Check that you have sufficient credits
- Ensure the API key has the necessary permissions

## Learning Path

1. **Start here** - Understand basic agent setup
2. Read the agent properties (Role, Backstory, Temperature)
3. Modify the agents and observe behavior changes
4. Add more complex examples from the other example directories

## Files

- `main.go` - Main application with 2-agent crew
- `.env.example` - Template for environment variables
- `go.mod` - Module definition
- `README.md` - This file

## Security Notes

⚠️ **Never commit your actual API keys!**
- `.env` file is ignored by git (see `.gitignore`)
- Always use `.env.example` as template
- For more security guidelines, see `/SECURITY.md` in the root directory

## Next Steps

After understanding this simple example, explore:
- **customer-service** - More complex crew with 3+ agents and tools
- **it-support** - Real-world IT support workflow
- **research-assistant** - Multi-step research process
- **data-analysis** - Data analysis with specialized agents

## Questions?

See the main go-agentic documentation in the root README for more information about:
- Building tools and handlers
- Custom prompts
- Streaming responses
- Error handling

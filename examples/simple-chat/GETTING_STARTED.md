# Getting Started with Simple Chat Example

This guide will help you run the simplest go-agentic example in just 5 minutes.

## Prerequisites

- Go 1.25 or later
- OpenAI API key (get one from https://platform.openai.com/account/api-keys)

## Quick Start (5 minutes)

### Step 1: Get Your API Key

1. Go to https://platform.openai.com/account/api-keys
2. Create a new API key (or use an existing one)
3. Copy it to clipboard

### Step 2: Create .env File

```bash
# From the simple-chat directory
cp .env.example .env
```

Then open `.env` and replace `sk-proj-your-api-key-here` with your actual API key:

```env
OPENAI_API_KEY=sk-proj-xxxxxxxxxxxxxxxxxxxxxxxxxx
```

**⚠️ Security Warning:**
- `.env` is ignored by git (see `.gitignore`)
- Never commit your actual API key!
- Never share it in pull requests or documentation

### Step 3: Install Dependencies

```bash
go mod download
```

### Step 4: Run the Example

```bash
go run main.go
```

## Expected Output

```
🤖 Simple Multi-Agent Chat System
==================================================

📌 Topic: What are the best practices for writing Go code?
--------------------------------------------------
✅ Final Response:
[Expert's comprehensive answer...]

📌 Topic: How can AI agents improve software development?
--------------------------------------------------
✅ Final Response:
[Expert's thoughts on AI agents...]

📌 Topic: Tell me about the latest trends in machine learning
--------------------------------------------------
✅ Final Response:
[Expert's insights on ML trends...]
```

## What's Happening?

The example creates a crew with 2 agents that automatically have a conversation:

1. **Enthusiast** (first agent)
   - Asks insightful questions about the topic
   - Keeps the conversation going
   - Can pass to the Expert agent

2. **Expert** (second agent)
   - Provides knowledgeable answers
   - Gives the final response
   - Terminal agent (stops the conversation)

The conversation flows like this:

```
Enthusiast: "Great topic! I'm curious about..."
     ↓
Expert: "Good question! Here's what I know..."
     ↓
Enthusiast: "Can you tell me more about..."
     ↓
Expert: "Absolutely! In detail, this means..."
     ↓ (End - Expert is terminal agent)
```

## Troubleshooting

### Problem: "OPENAI_API_KEY environment variable not set"

**Solution:** Make sure you've created the `.env` file with your actual API key:

```bash
# Check if .env exists
ls -la .env

# If not, create it
cp .env.example .env

# Then edit it with your API key
nano .env
```

### Problem: "module not found: github.com/taipm/go-agentic"

**Solution:** Make sure `go.mod` has the correct path. Check that:

```bash
# From simple-chat directory
cat go.mod | grep replace

# Should show:
# replace github.com/taipm/go-agentic => ../../go-agentic

# Verify the path exists
ls ../../go-agentic/go.mod
```

### Problem: "No such file or directory: .env"

**Solution:** The program tries to load `.env` if it exists, but will also work with environment variables:

```bash
# Option 1: Create .env file (recommended for development)
cp .env.example .env
nano .env

# Option 2: Use environment variable directly
export OPENAI_API_KEY="sk-proj-your-key-here"
go run main.go

# Option 3: Inline environment variable
OPENAI_API_KEY="sk-proj-your-key-here" go run main.go
```

### Problem: "OpenAI API error" or rate limiting

**Solution:** 
- Check your API key is correct
- Check you have remaining credits at https://platform.openai.com/account/usage
- Wait a moment and try again (rate limiting)
- Use `gpt-4o-mini` model (cheaper than gpt-4)

### Problem: Agent responses seem cut off or incomplete

**Solution:** This is normal! The crew has:
- `MaxRounds: 3` - Maximum 3 conversation rounds
- `MaxHandoffs: 2` - Maximum 2 handoffs between agents

To see longer conversations, modify `main.go`:

```go
crew := &agentic.Crew{
    Agents:      []*agentic.Agent{enthusiast, expert},
    MaxRounds:   5,      // Increase for longer conversations
    MaxHandoffs: 3,      // Increase for more back-and-forth
}
```

## Customization Examples

### Add More Topics

Edit the `topics` slice in `main.go`:

```go
topics := []string{
    "Your custom topic here",
    "Another interesting topic",
    "And one more...",
}
```

### Change Agent Personalities

Modify the agent's `Backstory` and `Role`:

```go
enthusiast := &agentic.Agent{
    ID:          "enthusiast",
    Name:        "Curious Student",
    Role:        "A student eager to learn",
    Backstory:   "You are a diligent student with lots of curiosity...",
    Temperature: 0.8,  // Higher = more creative, varied responses
    IsTerminal:  false,
}
```

### Adjust Creativity

The `Temperature` parameter controls how creative responses are:

```go
Temperature: 0.3,  // More focused, consistent
Temperature: 0.7,  // Balanced (default)
Temperature: 0.9,  // More creative, varied
```

### Change the Model

Replace `"gpt-4o-mini"` with other OpenAI models:

```go
// Faster, cheaper
Model: "gpt-4o-mini",

// More capable
Model: "gpt-4o",

// Standard GPT-4
Model: "gpt-4-turbo",

// Older, cheaper
Model: "gpt-3.5-turbo",
```

## Next Steps

1. **Explore the code** - Read through `main.go` to understand the structure
2. **Try customizations** - Modify topics, agents, or behaviors
3. **Read the README** - See `README.md` for more details
4. **Build something** - Use this as a template for your own multi-agent system
5. **Check other examples** - Look at `customer-service` for a more complex example

## File Structure

```
simple-chat/
├── main.go           # The application (readable, ~130 lines)
├── .env.example      # Template (copy to .env)
├── go.mod            # Dependencies
├── go.sum            # Checksums
├── README.md         # Full documentation
└── GETTING_STARTED.md # This file
```

## Key Files Explained

### main.go

The entire application in one file:

- **Line 1-10**: Imports
- **Line 12-28**: main() function
  - Loads .env file
  - Gets API key
  - Creates crew
  - Runs agent conversation
- **Line 30-56**: createSimpleChatCrew()
  - Defines Enthusiast agent
  - Defines Expert agent
  - Creates crew with MaxRounds and MaxHandoffs
- **Line 58-77**: loadEnvFile()
  - Reads .env file
  - Sets environment variables

### .env.example

Template showing what environment variables are needed:

```env
OPENAI_API_KEY=sk-proj-your-api-key-here
```

Copy to `.env` and fill in your actual API key.

## Security Best Practices

✅ **Do:**
- Use `.env.example` as template
- Never commit `.env` file
- Rotate API keys regularly
- Check `.gitignore` excludes `.env`
- Use specific API keys for development

❌ **Don't:**
- Hardcode API keys in code
- Commit `.env` files
- Share API keys via email/chat
- Reuse API keys across services

## Getting Help

If you have issues:

1. Check this guide's Troubleshooting section
2. Read the main `README.md` in this directory
3. See `/SECURITY.md` for security issues
4. Check the go-agentic main documentation
5. Review error messages carefully

## What to Do Next

After successfully running this example:

1. **Modify the example:**
   - Change topics
   - Add more agents
   - Add tools to agents
   - Increase conversation rounds

2. **Explore other examples:**
   - `customer-service` - Real-world customer support workflow
   - `it-support` - IT help desk automation
   - `research-assistant` - Multi-step research
   - `data-analysis` - Data analysis workflow

3. **Build your own:**
   - Design your own agents
   - Define their roles and responsibilities
   - Create custom tools
   - Implement your business logic

## Tips for Success

1. **Start small** - This example is intentionally minimal
2. **Understand the agents** - Read their Role and Backstory
3. **Try customizations** - Modify Temperature, MaxRounds, etc.
4. **Check the output** - See how agents interact
5. **Read the code** - It's well-commented and readable

---

**Ready to run?**

```bash
cp .env.example .env
# Edit .env with your API key
go run main.go
```

Good luck! 🚀

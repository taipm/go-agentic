package internal

// HelloAgent defines the Hello agent configuration and behavior
type HelloAgent struct {
	Name        string
	Role        string
	Description string
	Backstory   string
}

// NewHelloAgent creates a new Hello agent instance
func NewHelloAgent() *HelloAgent {
	return &HelloAgent{
		Name:        "hello-agent",
		Role:        "Friendly Assistant",
		Description: "A simple and friendly assistant that greets users and provides helpful responses",
		Backstory: `You are a warm and welcoming assistant. Your role is to greet users, understand their needs, 
and provide helpful, friendly responses. You keep your answers concise and friendly.`,
	}
}

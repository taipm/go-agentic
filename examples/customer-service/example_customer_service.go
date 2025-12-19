package main

import (
	"context"
	"fmt"
	"math/rand"

	"github.com/taipm/go-agentic"
)

// CreateCustomerServiceCrew creates a complete customer service crew
func createCustomerServiceCrew() *agentic.Crew {
	tools := createCustomerServiceTools()

	classifier := &agentic.Agent{
		ID:          "classifier",
		Name:        "Issue Classifier",
		Role:        "Customer issue categorizer",
		Backstory:   "You are expert at understanding customer issues and categorizing them. Determine the type of issue and decide which team should handle it.",
		Model:       "gpt-4o",
		Tools:       []*agentic.Tool{},
		Temperature: 0.7,
		IsTerminal:  false,
	}

	resolver := &agentic.Agent{
		ID:          "resolver",
		Name:        "Issue Resolver",
		Role:        "Problem solver",
		Backstory:   "You are expert at resolving customer issues. Use available tools to investigate and resolve customer problems.",
		Model:       "gpt-4o",
		Tools:       tools,
		Temperature: 0.7,
		IsTerminal:  false,
	}

	responder := &agentic.Agent{
		ID:          "responder",
		Name:        "Response Specialist",
		Role:        "Customer response coordinator",
		Backstory:   "You specialize in communicating solutions to customers in a friendly and professional manner.",
		Model:       "gpt-4o",
		Tools:       []*agentic.Tool{},
		Temperature: 0.7,
		IsTerminal:  true,
	}

	crew := &agentic.Crew{
		Agents:      []*agentic.Agent{classifier, resolver, responder},
		MaxRounds:   10,
		MaxHandoffs: 5,
	}

	return crew
}

// createCustomerServiceTools creates all customer service tools
func createCustomerServiceTools() []*agentic.Tool {
	return []*agentic.Tool{
		{
			Name:        "SearchKnowledgeBase",
			Description: "Search the knowledge base for solutions to common customer issues",
			Parameters: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"query": map[string]interface{}{
						"type":        "string",
						"description": "The search query",
					},
				},
				"required": []string{"query"},
			},
			Handler: func(ctx context.Context, args map[string]interface{}) (string, error) {
				query, ok := args["query"].(string)
				if !ok {
					return "", fmt.Errorf("invalid query parameter")
				}
				return fmt.Sprintf("Found knowledge base articles for: %s\n- Article 1: How to reset your password\n- Article 2: Troubleshooting common issues\n- Article 3: Billing FAQ", query), nil
			},
		},
		{
			Name:        "CheckAccountStatus",
			Description: "Check the status of a customer account",
			Parameters: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"account_id": map[string]interface{}{
						"type":        "string",
						"description": "The customer account ID",
					},
				},
				"required": []string{"account_id"},
			},
			Handler: func(ctx context.Context, args map[string]interface{}) (string, error) {
				accountID, ok := args["account_id"].(string)
				if !ok {
					return "", fmt.Errorf("invalid account_id parameter")
				}
				return fmt.Sprintf("Account %s Status:\n- Status: Active\n- Plan: Premium\n- Last Login: 2 hours ago\n- Account Balance: $150.00", accountID), nil
			},
		},
		{
			Name:        "ViewOrderHistory",
			Description: "View order history for a customer",
			Parameters: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"account_id": map[string]interface{}{
						"type":        "string",
						"description": "The customer account ID",
					},
				},
				"required": []string{"account_id"},
			},
			Handler: func(ctx context.Context, args map[string]interface{}) (string, error) {
				accountID, ok := args["account_id"].(string)
				if !ok {
					return "", fmt.Errorf("invalid account_id parameter")
				}
				return fmt.Sprintf("Order History for %s:\n- Order #001: $99.99 (Delivered)\n- Order #002: $49.99 (In Transit)\n- Order #003: $199.99 (Processing)", accountID), nil
			},
		},
		{
			Name:        "IssueRefund",
			Description: "Process a refund for a customer",
			Parameters: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"account_id": map[string]interface{}{
						"type":        "string",
						"description": "The customer account ID",
					},
					"amount": map[string]interface{}{
						"type":        "string",
						"description": "The refund amount",
					},
				},
				"required": []string{"account_id", "amount"},
			},
			Handler: func(ctx context.Context, args map[string]interface{}) (string, error) {
				accountID, ok := args["account_id"].(string)
				if !ok {
					return "", fmt.Errorf("invalid account_id parameter")
				}
				amount, _ := args["amount"].(string)
				_ = amount
				refundID := fmt.Sprintf("REF-%06d", rand.Intn(999999))
				return fmt.Sprintf("Refund processed!\n- Refund ID: %s\n- Account: %s\n- Amount: %s\n- Status: Pending (3-5 business days)", refundID, accountID, amount), nil
			},
		},
		{
			Name:        "ResetPassword",
			Description: "Initiate a password reset for a customer",
			Parameters: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"account_id": map[string]interface{}{
						"type":        "string",
						"description": "The customer account ID",
					},
				},
				"required": []string{"account_id"},
			},
			Handler: func(ctx context.Context, args map[string]interface{}) (string, error) {
				accountID, ok := args["account_id"].(string)
				if !ok {
					return "", fmt.Errorf("invalid account_id parameter")
				}
				return fmt.Sprintf("Password reset initiated for account %s\n- Reset link sent to registered email\n- Valid for 24 hours\n- Check spam folder if not received", accountID), nil
			},
		},
		{
			Name:        "CreateTicket",
			Description: "Create a support ticket for further investigation",
			Parameters: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"account_id": map[string]interface{}{
						"type":        "string",
						"description": "The customer account ID",
					},
					"issue_description": map[string]interface{}{
						"type":        "string",
						"description": "Description of the issue",
					},
				},
				"required": []string{"account_id", "issue_description"},
			},
			Handler: func(ctx context.Context, args map[string]interface{}) (string, error) {
				accountID, ok := args["account_id"].(string)
				if !ok {
					return "", fmt.Errorf("invalid account_id parameter")
				}
				desc, _ := args["issue_description"].(string)
				ticketID := fmt.Sprintf("TKT-%06d", rand.Intn(999999))
				return fmt.Sprintf("Support ticket created!\n- Ticket ID: %s\n- Account: %s\n- Issue: %s\n- Priority: Normal\n- Expected Response: 24 hours", ticketID, accountID, desc), nil
			},
		},
	}
}

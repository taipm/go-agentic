package cost

import (
	"context"
	"testing"
)

func TestNewCrewCostAggregator(t *testing.T) {
	agg := NewCrewCostAggregator("crew1")

	if agg.GetTotalCost() != 0.0 {
		t.Errorf("Initial total cost should be 0, got %f", agg.GetTotalCost())
	}

	if agg.GetExecutionCount() != 0 {
		t.Errorf("Initial execution count should be 0, got %d", agg.GetExecutionCount())
	}
}

func TestRecordAgentCost(t *testing.T) {
	agg := NewCrewCostAggregator("crew1")
	ctx := context.Background()

	agg.RecordAgentCost(ctx, "agent1", "Agent 1", 10.0, 1)
	agg.RecordAgentCost(ctx, "agent2", "Agent 2", 20.0, 1)

	if agg.GetTotalCost() != 30.0 {
		t.Errorf("Total cost should be 30.0, got %f", agg.GetTotalCost())
	}

	if agg.GetExecutionCount() != 2 {
		t.Errorf("Execution count should be 2, got %d", agg.GetExecutionCount())
	}
}

func TestGetAgentCostBreakdown(t *testing.T) {
	agg := NewCrewCostAggregator("crew1")
	ctx := context.Background()

	agg.RecordAgentCost(ctx, "agent1", "Agent 1", 10.0, 1)
	agg.RecordAgentCost(ctx, "agent1", "Agent 1", 15.0, 2)

	breakdown := agg.GetAgentCostBreakdown("agent1")

	if breakdown == nil {
		t.Fatal("Expected agent breakdown, got nil")
	}

	if breakdown.TotalCost != 25.0 {
		t.Errorf("Agent total cost should be 25.0, got %f", breakdown.TotalCost)
	}

	if breakdown.ExecutionCount != 2 {
		t.Errorf("Agent execution count should be 2, got %d", breakdown.ExecutionCount)
	}

	if breakdown.AverageCost != 12.5 {
		t.Errorf("Agent average cost should be 12.5, got %f", breakdown.AverageCost)
	}

	if breakdown.MinCost != 10.0 {
		t.Errorf("Agent min cost should be 10.0, got %f", breakdown.MinCost)
	}

	if breakdown.MaxCost != 15.0 {
		t.Errorf("Agent max cost should be 15.0, got %f", breakdown.MaxCost)
	}
}

func TestGetAllAgentCosts(t *testing.T) {
	agg := NewCrewCostAggregator("crew1")
	ctx := context.Background()

	agg.RecordAgentCost(ctx, "agent1", "Agent 1", 10.0, 1)
	agg.RecordAgentCost(ctx, "agent2", "Agent 2", 20.0, 1)
	agg.RecordAgentCost(ctx, "agent3", "Agent 3", 5.0, 1)

	costs := agg.GetAllAgentCosts()

	if len(costs) != 3 {
		t.Errorf("Expected 3 agents, got %d", len(costs))
	}

	// Should be sorted by total cost descending
	if costs[0].TotalCost != 20.0 {
		t.Errorf("First agent should have cost 20.0, got %f", costs[0].TotalCost)
	}

	if costs[1].TotalCost != 10.0 {
		t.Errorf("Second agent should have cost 10.0, got %f", costs[1].TotalCost)
	}

	if costs[2].TotalCost != 5.0 {
		t.Errorf("Third agent should have cost 5.0, got %f", costs[2].TotalCost)
	}
}

func TestCostPercentage(t *testing.T) {
	agg := NewCrewCostAggregator("crew1")
	ctx := context.Background()

	agg.RecordAgentCost(ctx, "agent1", "Agent 1", 30.0, 1)
	agg.RecordAgentCost(ctx, "agent2", "Agent 2", 20.0, 1)
	agg.RecordAgentCost(ctx, "agent3", "Agent 3", 50.0, 1)

	breakdown := agg.GetAgentCostBreakdown("agent1")

	expectedPercent := 30.0
	if breakdown.CostPercentage != expectedPercent {
		t.Errorf("Agent percentage should be %f, got %f", expectedPercent, breakdown.CostPercentage)
	}
}

func TestMostExpensiveAgent(t *testing.T) {
	agg := NewCrewCostAggregator("crew1")
	ctx := context.Background()

	agg.RecordAgentCost(ctx, "agent1", "Agent 1", 10.0, 1)
	agg.RecordAgentCost(ctx, "agent2", "Agent 2", 30.0, 1)
	agg.RecordAgentCost(ctx, "agent3", "Agent 3", 5.0, 1)

	most := agg.GetMostExpensiveAgent()

	if most == nil {
		t.Fatal("Expected most expensive agent, got nil")
	}

	if most.AgentID != "agent2" {
		t.Errorf("Most expensive agent should be agent2, got %s", most.AgentID)
	}

	if most.TotalCost != 30.0 {
		t.Errorf("Most expensive agent cost should be 30.0, got %f", most.TotalCost)
	}
}

func TestAverageCostPerExecution(t *testing.T) {
	agg := NewCrewCostAggregator("crew1")
	ctx := context.Background()

	agg.RecordAgentCost(ctx, "agent1", "Agent 1", 10.0, 1)
	agg.RecordAgentCost(ctx, "agent2", "Agent 2", 20.0, 1)
	agg.RecordAgentCost(ctx, "agent3", "Agent 3", 30.0, 1)

	avg := agg.GetAverageCostPerExecution()

	expected := 20.0
	if avg != expected {
		t.Errorf("Average cost should be %f, got %f", expected, avg)
	}
}

func TestCrewCostReport(t *testing.T) {
	agg := NewCrewCostAggregator("crew1")
	ctx := context.Background()

	agg.RecordAgentCost(ctx, "agent1", "Agent 1", 10.0, 1)
	agg.RecordAgentCost(ctx, "agent2", "Agent 2", 20.0, 1)

	report := agg.GetCrewCostReport(ctx)

	if report == nil {
		t.Fatal("Expected crew cost report, got nil")
	}

	if report.TotalCost != 30.0 {
		t.Errorf("Report total cost should be 30.0, got %f", report.TotalCost)
	}

	if report.ExecutionCount != 2 {
		t.Errorf("Report execution count should be 2, got %d", report.ExecutionCount)
	}

	if report.AgentCount != 2 {
		t.Errorf("Report agent count should be 2, got %d", report.AgentCount)
	}
}

func TestTopExpensiveAgents(t *testing.T) {
	agg := NewCrewCostAggregator("crew1")
	ctx := context.Background()

	agg.RecordAgentCost(ctx, "agent1", "Agent 1", 10.0, 1)
	agg.RecordAgentCost(ctx, "agent2", "Agent 2", 30.0, 1)
	agg.RecordAgentCost(ctx, "agent3", "Agent 3", 20.0, 1)
	agg.RecordAgentCost(ctx, "agent4", "Agent 4", 5.0, 1)

	top2 := agg.GetTopExpensiveAgents(2)

	if len(top2) != 2 {
		t.Errorf("Expected 2 agents, got %d", len(top2))
	}

	if top2[0].AgentID != "agent2" || top2[0].TotalCost != 30.0 {
		t.Error("First agent should be agent2 with cost 30.0")
	}

	if top2[1].AgentID != "agent3" || top2[1].TotalCost != 20.0 {
		t.Error("Second agent should be agent3 with cost 20.0")
	}
}

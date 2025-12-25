package cost

import (
	"context"
	"sync"
)

// CostTracker tracks cumulative costs across execution
type CostTracker struct {
	mu              sync.RWMutex
	budget          *BudgetTracker
	totalCost       float64
	costByAgent     map[string]float64
	costByRound     map[int]float64
	executionCount  int
}

// NewCostTracker creates a new cost tracker
func NewCostTracker(budget *BudgetTracker) *CostTracker {
	if budget == nil {
		budget = NewBudgetTracker(nil)
	}

	return &CostTracker{
		budget:         budget,
		totalCost:      0.0,
		costByAgent:    make(map[string]float64),
		costByRound:    make(map[int]float64),
		executionCount: 0,
	}
}

// RecordAgentCost records cost for an agent execution
func (ct *CostTracker) RecordAgentCost(ctx context.Context, agentID string, cost float64, round int) error {
	ct.mu.Lock()
	defer ct.mu.Unlock()

	// Record to budget tracker
	if err := ct.budget.RecordCost(ctx, cost, agentID); err != nil {
		return err
	}

	// Update cumulative cost
	ct.totalCost += cost

	// Update agent cost
	ct.costByAgent[agentID] += cost

	// Update round cost
	ct.costByRound[round] += cost

	// Increment execution count
	ct.executionCount++

	return nil
}

// GetTotalCost returns total accumulated cost
func (ct *CostTracker) GetTotalCost() float64 {
	ct.mu.RLock()
	defer ct.mu.RUnlock()
	return ct.totalCost
}

// GetAgentCost returns cost for a specific agent
func (ct *CostTracker) GetAgentCost(agentID string) float64 {
	ct.mu.RLock()
	defer ct.mu.RUnlock()
	return ct.costByAgent[agentID]
}

// GetRoundCost returns cost for a specific round
func (ct *CostTracker) GetRoundCost(round int) float64 {
	ct.mu.RLock()
	defer ct.mu.RUnlock()
	return ct.costByRound[round]
}

// GetExecutionCount returns number of agent executions
func (ct *CostTracker) GetExecutionCount() int {
	ct.mu.RLock()
	defer ct.mu.RUnlock()
	return ct.executionCount
}

// GetAverageCostPerExecution returns average cost per execution
func (ct *CostTracker) GetAverageCostPerExecution() float64 {
	ct.mu.RLock()
	defer ct.mu.RUnlock()

	if ct.executionCount == 0 {
		return 0.0
	}
	return ct.totalCost / float64(ct.executionCount)
}

// GetCostByAgent returns cost breakdown by agent
func (ct *CostTracker) GetCostByAgent() map[string]float64 {
	ct.mu.RLock()
	defer ct.mu.RUnlock()

	// Return a copy
	result := make(map[string]float64, len(ct.costByAgent))
	for agent, cost := range ct.costByAgent {
		result[agent] = cost
	}
	return result
}

// GetCostByRound returns cost breakdown by round
func (ct *CostTracker) GetCostByRound() map[int]float64 {
	ct.mu.RLock()
	defer ct.mu.RUnlock()

	// Return a copy
	result := make(map[int]float64, len(ct.costByRound))
	for round, cost := range ct.costByRound {
		result[round] = cost
	}
	return result
}

// GetCostReport generates a comprehensive cost report
func (ct *CostTracker) GetCostReport() *CostReport {
	ct.mu.RLock()
	defer ct.mu.RUnlock()

	summary := ct.budget.GetCostSummary()

	return &CostReport{
		TotalCost:             ct.totalCost,
		ExecutionCount:        ct.executionCount,
		AverageCostPerExecution: ct.GetAverageCostPerExecution(),
		CostByAgent:           ct.getCostByAgentCopy(),
		CostByRound:           ct.getCostByRoundCopy(),
		SessionCost:           ct.budget.GetCurrentSessionCost(),
		SessionBudget:         ct.budget.config.SessionLimit,
		DailyCost:             ct.budget.GetCurrentDailyCost(),
		DailyBudget:           ct.budget.config.DailyLimit,
		BudgetSummary:         summary,
		Alerts:                ct.budget.GetAlerts(),
	}
}

// getCostByAgentCopy returns a copy of cost by agent
func (ct *CostTracker) getCostByAgentCopy() map[string]float64 {
	result := make(map[string]float64, len(ct.costByAgent))
	for agent, cost := range ct.costByAgent {
		result[agent] = cost
	}
	return result
}

// getCostByRoundCopy returns a copy of cost by round
func (ct *CostTracker) getCostByRoundCopy() map[int]float64 {
	result := make(map[int]float64, len(ct.costByRound))
	for round, cost := range ct.costByRound {
		result[round] = cost
	}
	return result
}

// CostReport represents a comprehensive cost report
type CostReport struct {
	TotalCost                float64                  `json:"total_cost"`
	ExecutionCount           int                      `json:"execution_count"`
	AverageCostPerExecution  float64                  `json:"average_cost_per_execution"`
	CostByAgent              map[string]float64       `json:"cost_by_agent"`
	CostByRound              map[int]float64          `json:"cost_by_round"`
	SessionCost              float64                  `json:"session_cost"`
	SessionBudget            float64                  `json:"session_budget"`
	DailyCost                float64                  `json:"daily_cost"`
	DailyBudget              float64                  `json:"daily_budget"`
	BudgetSummary            map[string]interface{}   `json:"budget_summary"`
	Alerts                   []CostAlert              `json:"alerts"`
}

package cost

import (
	"context"
	"log/slog"
	"sort"
	"sync"
	"time"

	"github.com/taipm/go-agentic/core/logging"
)

// AgentCostBreakdown represents cost breakdown for a single agent
type AgentCostBreakdown struct {
	AgentID         string                 `json:"agent_id"`
	AgentName       string                 `json:"agent_name"`
	TotalCost       float64                `json:"total_cost"`
	ExecutionCount  int                    `json:"execution_count"`
	AverageCost     float64                `json:"average_cost"`
	MinCost         float64                `json:"min_cost"`
	MaxCost         float64                `json:"max_cost"`
	LastExecuted    time.Time              `json:"last_executed"`
	CostPercentage  float64                `json:"cost_percentage"`
}

// CrewCostAggregator aggregates costs across all agents in a crew
type CrewCostAggregator struct {
	mu                  sync.RWMutex
	crewID              string
	agentCosts          map[string]*AgentCostBreakdown
	totalCost           float64
	executionCount      int
	startTime           time.Time
	costHistory         []CostEntry
	maxHistorySize      int
}

// CostEntry represents a single cost entry in history
type CostEntry struct {
	Timestamp   time.Time
	AgentID     string
	AgentName   string
	Cost        float64
	Round       int
}

// NewCrewCostAggregator creates a new crew cost aggregator
func NewCrewCostAggregator(crewID string) *CrewCostAggregator {
	return &CrewCostAggregator{
		crewID:         crewID,
		agentCosts:     make(map[string]*AgentCostBreakdown),
		totalCost:      0.0,
		executionCount: 0,
		startTime:      time.Now(),
		costHistory:    make([]CostEntry, 0),
		maxHistorySize: 1000,
	}
}

// RecordAgentCost records cost for an agent and aggregates it
func (cca *CrewCostAggregator) RecordAgentCost(ctx context.Context, agentID, agentName string, cost float64, round int) {
	cca.mu.Lock()
	defer cca.mu.Unlock()

	// Update total
	cca.totalCost += cost
	cca.executionCount++

	// Get or create agent breakdown
	breakdown, exists := cca.agentCosts[agentID]
	if !exists {
		breakdown = &AgentCostBreakdown{
			AgentID:      agentID,
			AgentName:    agentName,
			MinCost:      cost,
			MaxCost:      cost,
		}
		cca.agentCosts[agentID] = breakdown
	}

	// Update agent breakdown
	breakdown.TotalCost += cost
	breakdown.ExecutionCount++
	breakdown.AverageCost = breakdown.TotalCost / float64(breakdown.ExecutionCount)
	breakdown.LastExecuted = time.Now()

	if cost < breakdown.MinCost {
		breakdown.MinCost = cost
	}
	if cost > breakdown.MaxCost {
		breakdown.MaxCost = cost
	}

	// Add to history
	cca.costHistory = append(cca.costHistory, CostEntry{
		Timestamp: time.Now(),
		AgentID:   agentID,
		AgentName: agentName,
		Cost:      cost,
		Round:     round,
	})

	// Limit history
	if len(cca.costHistory) > cca.maxHistorySize {
		cca.costHistory = cca.costHistory[len(cca.costHistory)-cca.maxHistorySize:]
	}

	// Log agent cost recording
	logging.GetLogger().InfoContext(ctx, "crew.agent_cost",
		slog.String("event", "crew.agent_cost"),
		slog.String("trace_id", logging.GetTraceID(ctx)),
		slog.String("crew_id", cca.crewID),
		slog.String("agent_id", agentID),
		slog.String("agent_name", agentName),
		slog.Float64("cost_usd", cost),
		slog.Int("agent_execution_count", breakdown.ExecutionCount),
		slog.Float64("agent_total_cost", breakdown.TotalCost),
	)
}

// GetAgentCostBreakdown returns cost breakdown for a specific agent
func (cca *CrewCostAggregator) GetAgentCostBreakdown(agentID string) *AgentCostBreakdown {
	cca.mu.RLock()
	defer cca.mu.RUnlock()

	breakdown, exists := cca.agentCosts[agentID]
	if !exists {
		return nil
	}

	// Return a copy with percentage calculated
	copy := *breakdown
	copy.CostPercentage = (copy.TotalCost / cca.totalCost) * 100
	return &copy
}

// GetAllAgentCosts returns cost breakdown for all agents
func (cca *CrewCostAggregator) GetAllAgentCosts() []*AgentCostBreakdown {
	cca.mu.RLock()
	defer cca.mu.RUnlock()

	breakdowns := make([]*AgentCostBreakdown, 0, len(cca.agentCosts))
	for _, breakdown := range cca.agentCosts {
		copy := *breakdown
		copy.CostPercentage = (copy.TotalCost / cca.totalCost) * 100
		breakdowns = append(breakdowns, &copy)
	}

	// Sort by total cost descending
	sort.Slice(breakdowns, func(i, j int) bool {
		return breakdowns[i].TotalCost > breakdowns[j].TotalCost
	})

	return breakdowns
}

// GetTotalCost returns total cost across all agents
func (cca *CrewCostAggregator) GetTotalCost() float64 {
	cca.mu.RLock()
	defer cca.mu.RUnlock()
	return cca.totalCost
}

// GetMostExpensiveAgent returns the agent with the highest total cost
func (cca *CrewCostAggregator) GetMostExpensiveAgent() *AgentCostBreakdown {
	breakdowns := cca.GetAllAgentCosts()
	if len(breakdowns) == 0 {
		return nil
	}
	return breakdowns[0]
}

// GetCheapestAgent returns the agent with the lowest total cost
func (cca *CrewCostAggregator) GetCheapestAgent() *AgentCostBreakdown {
	breakdowns := cca.GetAllAgentCosts()
	if len(breakdowns) == 0 {
		return nil
	}
	return breakdowns[len(breakdowns)-1]
}

// GetExecutionCount returns total execution count across all agents
func (cca *CrewCostAggregator) GetExecutionCount() int {
	cca.mu.RLock()
	defer cca.mu.RUnlock()
	return cca.executionCount
}

// GetAverageCostPerExecution returns average cost per execution across all agents
func (cca *CrewCostAggregator) GetAverageCostPerExecution() float64 {
	cca.mu.RLock()
	defer cca.mu.RUnlock()

	if cca.executionCount == 0 {
		return 0.0
	}
	return cca.totalCost / float64(cca.executionCount)
}

// GetCostHistory returns the cost history
func (cca *CrewCostAggregator) GetCostHistory() []CostEntry {
	cca.mu.RLock()
	defer cca.mu.RUnlock()

	// Return a copy
	history := make([]CostEntry, len(cca.costHistory))
	copy(history, cca.costHistory)
	return history
}

// GetCrewCostReport generates a comprehensive crew cost report
func (cca *CrewCostAggregator) GetCrewCostReport(ctx context.Context) *CrewCostReport {
	cca.mu.RLock()
	defer cca.mu.RUnlock()

	report := &CrewCostReport{
		CrewID:                cca.crewID,
		TotalCost:             cca.totalCost,
		ExecutionCount:        cca.executionCount,
		AverageCostPerExecution: cca.GetAverageCostPerExecution(),
		AgentCount:            len(cca.agentCosts),
		ElapsedTime:           time.Since(cca.startTime),
		StartTime:             cca.startTime,
		ReportTime:            time.Now(),
		AgentBreakdowns:       cca.getAgentBreakdownsCopy(),
	}

	// Log the report generation
	logging.GetLogger().InfoContext(ctx, "crew.cost_report",
		slog.String("event", "crew.cost_report"),
		slog.String("trace_id", logging.GetTraceID(ctx)),
		slog.String("crew_id", cca.crewID),
		slog.Float64("total_cost", report.TotalCost),
		slog.Int("execution_count", report.ExecutionCount),
		slog.Int("agent_count", report.AgentCount),
		slog.String("elapsed_time", report.ElapsedTime.String()),
	)

	return report
}

// getAgentBreakdownsCopy returns a copy of all agent breakdowns with percentages
func (cca *CrewCostAggregator) getAgentBreakdownsCopy() []*AgentCostBreakdown {
	breakdowns := make([]*AgentCostBreakdown, 0, len(cca.agentCosts))
	for _, breakdown := range cca.agentCosts {
		copy := *breakdown
		if cca.totalCost > 0 {
			copy.CostPercentage = (copy.TotalCost / cca.totalCost) * 100
		}
		breakdowns = append(breakdowns, &copy)
	}

	// Sort by total cost descending
	sort.Slice(breakdowns, func(i, j int) bool {
		return breakdowns[i].TotalCost > breakdowns[j].TotalCost
	})

	return breakdowns
}

// CrewCostReport represents a comprehensive crew cost report
type CrewCostReport struct {
	CrewID                   string                 `json:"crew_id"`
	TotalCost                float64                `json:"total_cost"`
	ExecutionCount           int                    `json:"execution_count"`
	AverageCostPerExecution  float64                `json:"average_cost_per_execution"`
	AgentCount               int                    `json:"agent_count"`
	ElapsedTime              time.Duration          `json:"elapsed_time"`
	StartTime                time.Time              `json:"start_time"`
	ReportTime               time.Time              `json:"report_time"`
	AgentBreakdowns          []*AgentCostBreakdown  `json:"agent_breakdowns"`
}

// GetTopExpensiveAgents returns the top N most expensive agents
func (cca *CrewCostAggregator) GetTopExpensiveAgents(n int) []*AgentCostBreakdown {
	cca.mu.RLock()
	defer cca.mu.RUnlock()

	breakdowns := cca.getAgentBreakdownsCopy()
	if n > len(breakdowns) {
		n = len(breakdowns)
	}
	return breakdowns[:n]
}

// Reset resets the aggregator
func (cca *CrewCostAggregator) Reset(ctx context.Context) {
	cca.mu.Lock()
	defer cca.mu.Unlock()

	logging.GetLogger().InfoContext(ctx, "crew.cost_aggregator_reset",
		slog.String("event", "crew.cost_aggregator_reset"),
		slog.String("trace_id", logging.GetTraceID(ctx)),
		slog.String("crew_id", cca.crewID),
		slog.Float64("total_cost_reset", cca.totalCost),
		slog.Int("execution_count_reset", cca.executionCount),
	)

	cca.agentCosts = make(map[string]*AgentCostBreakdown)
	cca.totalCost = 0.0
	cca.executionCount = 0
	cca.startTime = time.Now()
	cca.costHistory = make([]CostEntry, 0)
}

# Quick Reference: Structured Logging + Cost Tracking

## Package Imports

```go
import (
    "github.com/taipm/go-agentic/core/logging"
    "github.com/taipm/go-agentic/core/cost"
    "log/slog"
)
```

## Logging

### Get Logger
```go
logger := logging.GetLogger()
```

### Log an Event
```go
logging.GetLogger().InfoContext(ctx, "event_name",
    slog.String("field1", "value1"),
    slog.Float64("field2", 123.45),
    slog.Int("field3", 42),
)
```

### Trace ID Management
```go
// Generate and add trace ID to context
traceID := uuid.New().String()
ctx = logging.WithTraceID(ctx, traceID)

// Retrieve trace ID from context
traceID := logging.GetTraceID(ctx)
```

### Testing
```go
var buf strings.Builder
logging.SetOutput(&buf)
// ... run test code ...
logOutput := buf.String()
```

## Cost Tracking

### Budget Tracking
```go
// Create tracker with default config (50 session, 100 daily)
bt := cost.NewBudgetTracker(nil)

// Or custom config
bt := cost.NewBudgetTracker(&cost.BudgetConfig{
    SessionLimit:      50.0,
    DailyLimit:        100.0,
    WarningThreshold:  0.75,   // 75%
    CriticalThreshold: 0.90,   // 90%
})

// Record a cost
err := bt.RecordCost(ctx, 10.5, "agent1")

// Check budget status
if bt.IsSessionBudgetExceeded() {
    // Handle exceeded budget
}

// Get remaining budget
remaining := bt.GetSessionBudgetRemaining()

// Get alerts
alerts := bt.GetAlerts()
for _, alert := range alerts {
    if alert.Level == cost.AlertCritical {
        log.Printf("CRITICAL: %s", alert.Message)
    }
}

// Reset session budget (keeps daily)
bt.Reset(ctx)

// Get summary
summary := bt.GetCostSummary()
// {
//   "session_cost": 15.5,
//   "session_budget": 50.0,
//   "session_remaining": 34.5,
//   "daily_cost": 25.5,
//   ...
// }
```

### Multi-Agent Cost Aggregation
```go
// Create aggregator
agg := cost.NewCrewCostAggregator("crew1")

// Record agent costs
agg.RecordAgentCost(ctx, "agent1", "Teacher", 10.5, 1)
agg.RecordAgentCost(ctx, "agent2", "Student", 5.2, 1)

// Get breakdown for specific agent
breakdown := agg.GetAgentCostBreakdown("agent1")
// {
//   AgentID: "agent1",
//   AgentName: "Teacher",
//   TotalCost: 10.5,
//   ExecutionCount: 1,
//   AverageCost: 10.5,
//   MinCost: 10.5,
//   MaxCost: 10.5,
//   CostPercentage: 66.88,
// }

// Get all costs (sorted by total descending)
all := agg.GetAllAgentCosts()

// Get rankings
mostExpensive := agg.GetMostExpensiveAgent()
cheapest := agg.GetCheapestAgent()
top3 := agg.GetTopExpensiveAgents(3)

// Get metrics
total := agg.GetTotalCost()
count := agg.GetExecutionCount()
avg := agg.GetAverageCostPerExecution()

// Get comprehensive report
report := agg.GetCrewCostReport(ctx)
// {
//   CrewID: "crew1",
//   TotalCost: 15.7,
//   ExecutionCount: 2,
//   AverageCostPerExecution: 7.85,
//   AgentCount: 2,
//   ElapsedTime: 5s234ms,
//   AgentBreakdowns: [...],
// }

// Get cost history
history := agg.GetCostHistory()
// []CostEntry with Timestamp, AgentID, AgentName, Cost, Round

// Reset aggregator
agg.Reset(ctx)
```

### Unified Cost Tracker
```go
// Create tracker (with optional budget)
ct := cost.NewCostTracker(bt)

// Record cost for agent
err := ct.RecordAgentCost(ctx, "agent1", 10.5, 1)

// Get breakdown
total := ct.GetTotalCost()
byAgent := ct.GetCostByAgent()  // map[string]float64
byRound := ct.GetCostByRound()  // map[int]float64

// Get comprehensive report
report := ct.GetCostReport()
// {
//   TotalCost: 15.7,
//   CostByAgent: {"agent1": 10.5, "agent2": 5.2},
//   CostByRound: {1: 15.7},
//   SessionCost: 15.7,
//   SessionBudget: 50.0,
//   DailyCost: 25.5,
//   DailyBudget: 100.0,
//   Alerts: [...],
// }
```

## Logging Events Reference

### Workflow Events
```
Event: workflow.start
Fields: trace_id, entry_agent, input_preview

Event: workflow.end
Fields: trace_id, final_agent, status, duration_ms
```

### Agent Events
```
Event: round.start
Fields: trace_id, agent_id, agent_name, round

Event: agent.start
Fields: trace_id, agent_id, agent_name, round

Event: agent.end
Fields: trace_id, agent_id, duration_ms
```

### Signal Events
```
Event: signals.extracted
Fields: trace_id, signals

Event: registry.signal_emitted
Fields: signal_name, agent_id

Event: handler.match
Fields: signal_name, handler_id, handler_name, target_agent
```

### Routing Events
```
Event: route.handoff
Fields: trace_id, from_agent, to_agent, reason, round

Event: route.terminal
Fields: trace_id, agent_id, terminal_reason
```

### Tool Events
```
Event: tool.start
Fields: trace_id, tool_name, agent_id

Event: tool.end
Fields: trace_id, tool_name, result_preview

Event: tool.error
Fields: trace_id, tool_name, error
```

### Parallel Events
```
Event: parallel.start
Fields: trace_id, agent_count

Event: parallel.agent_start
Fields: trace_id, agent_id, group_id

Event: parallel.agent_end
Fields: trace_id, agent_id, duration_ms

Event: parallel.end
Fields: trace_id, completed_count, duration_ms
```

### Cost Events
```
Event: cost.recorded
Fields: trace_id, agent_id, cost_usd, session_total_usd, daily_total_usd

Event: cost.alert
Fields: trace_id, level, message, current_cost, budget_limit, threshold_percent

Event: cost.budget_reset
Fields: trace_id, session_cost_reset

Event: crew.agent_cost
Fields: trace_id, crew_id, agent_id, agent_name, cost_usd, agent_execution_count, agent_total_cost

Event: crew.cost_report
Fields: trace_id, crew_id, total_cost, execution_count, agent_count, elapsed_time
```

## Constants

### Alert Levels
```go
cost.AlertWarning   // "warning"
cost.AlertCritical  // "critical"
```

### Default Budget Configuration
```go
DailyLimit:        100.0
SessionLimit:      50.0
WarningThreshold:  0.75   // 75%
CriticalThreshold: 0.90   // 90%
```

## Common Patterns

### Complete Workflow with Logging and Cost Tracking

```go
package main

import (
    "context"
    "log/slog"
    
    "github.com/taipm/go-agentic/core/logging"
    "github.com/taipm/go-agentic/core/cost"
)

func main() {
    ctx := context.Background()
    
    // Initialize cost tracking
    budgetConfig := &cost.BudgetConfig{
        SessionLimit: 100.0,
        DailyLimit: 500.0,
    }
    bt := cost.NewBudgetTracker(budgetConfig)
    agg := cost.NewCrewCostAggregator("crew1")
    
    // Execute workflow with logging
    traceID := generateTraceID()
    ctx = logging.WithTraceID(ctx, traceID)
    
    logging.GetLogger().InfoContext(ctx, "workflow.start",
        slog.String("trace_id", traceID),
        slog.String("entry_agent", "agent1"),
    )
    
    // Simulate agent execution
    agentCost := 10.5
    if err := bt.RecordCost(ctx, agentCost, "agent1"); err != nil {
        logging.GetLogger().ErrorContext(ctx, "budget_exceeded", slog.String("error", err.Error()))
        return
    }
    
    agg.RecordAgentCost(ctx, "agent1", "Agent 1", agentCost, 1)
    
    // Get final report
    report := agg.GetCrewCostReport(ctx)
    logging.GetLogger().InfoContext(ctx, "workflow.end",
        slog.String("trace_id", traceID),
        slog.Float64("total_cost", report.TotalCost),
        slog.Int("agent_count", report.AgentCount),
    )
}

func generateTraceID() string {
    return uuid.New().String()
}
```

### JSON Log Output Example

```json
{
  "time": "2025-12-25T10:00:00Z",
  "level": "INFO",
  "msg": "workflow.start",
  "event": "workflow.start",
  "trace_id": "550e8400-e29b-41d4-a716-446655440000",
  "entry_agent": "teacher",
  "input_preview": "What is 2+2?"
}

{
  "time": "2025-12-25T10:00:00Z",
  "level": "INFO",
  "msg": "cost.recorded",
  "event": "cost.recorded",
  "trace_id": "550e8400-e29b-41d4-a716-446655440000",
  "agent_id": "teacher",
  "cost_usd": 0.00312,
  "session_total_usd": 0.00312,
  "daily_total_usd": 0.00312
}

{
  "time": "2025-12-25T10:00:00Z",
  "level": "INFO",
  "msg": "workflow.end",
  "event": "workflow.end",
  "trace_id": "550e8400-e29b-41d4-a716-446655440000",
  "final_agent": "teacher",
  "status": "terminal",
  "duration_ms": 234
}
```

## Documentation

- **Full Documentation**: See `IMPLEMENTATION_COMPLETE.md`
- **Test Examples**: See `core/cost/*_test.go`
- **Logger Implementation**: See `core/logging/logger.go`
- **Cost Tracker**: See `core/cost/budget.go`, `core/cost/aggregator.go`

## Testing Your Code

```bash
# Test cost package
go test ./core/cost -v

# Test all core packages
go test ./core/... -v

# Build check
go build ./...

# Run with trace output
LOG_LEVEL=debug go run ./examples/01-quiz-exam
```

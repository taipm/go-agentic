// Package cost provides cost tracking, budgeting, and alert management for agent execution.
package cost

import (
	"context"
	"fmt"
	"log/slog"
	"sync"
	"time"

	"github.com/taipm/go-agentic/core/logging"
)

// AlertLevel represents the severity of a cost alert
type AlertLevel string

const (
	AlertWarning AlertLevel = "warning"
	AlertCritical AlertLevel = "critical"
)

// CostAlert represents a cost-related alert
type CostAlert struct {
	Level     AlertLevel
	Message   string
	Cost      float64
	Budget    float64
	Threshold float64
	Timestamp time.Time
}

// BudgetConfig represents budget configuration
type BudgetConfig struct {
	DailyLimit       float64 // Daily cost limit in USD
	SessionLimit     float64 // Session cost limit in USD
	WarningThreshold float64 // Warning threshold as % of limit (e.g., 0.75 = 75%)
	CriticalThreshold float64 // Critical threshold as % of limit (e.g., 0.90 = 90%)
}

// BudgetTracker tracks costs against budgets and generates alerts
type BudgetTracker struct {
	mu               sync.RWMutex
	config           *BudgetConfig
	currentDailyCost float64
	currentSessionCost float64
	dailyResetTime   time.Time
	alerts           []CostAlert
	maxAlerts        int
}

// NewBudgetTracker creates a new budget tracker with the given configuration
func NewBudgetTracker(config *BudgetConfig) *BudgetTracker {
	if config == nil {
		config = &BudgetConfig{
			DailyLimit:        100.0,
			SessionLimit:      50.0,
			WarningThreshold:  0.75,
			CriticalThreshold: 0.90,
		}
	}

	return &BudgetTracker{
		config:          config,
		currentDailyCost: 0.0,
		currentSessionCost: 0.0,
		dailyResetTime:  time.Now(),
		alerts:          make([]CostAlert, 0),
		maxAlerts:       100,
	}
}

// RecordCost records a cost and checks budget limits
func (bt *BudgetTracker) RecordCost(ctx context.Context, cost float64, agentID string) error {
	bt.mu.Lock()
	defer bt.mu.Unlock()

	// Check if daily limit should be reset (new day)
	if time.Since(bt.dailyResetTime) > 24*time.Hour {
		bt.currentDailyCost = 0.0
		bt.dailyResetTime = time.Now()
	}

	// Add cost
	bt.currentDailyCost += cost
	bt.currentSessionCost += cost

	// Check budgets and generate alerts
	bt.checkBudgets(ctx, cost, agentID)

	// Log cost recording
	logging.GetLogger().InfoContext(ctx, "cost.recorded",
		slog.String("event", "cost.recorded"),
		slog.String("trace_id", logging.GetTraceID(ctx)),
		slog.String("agent_id", agentID),
		slog.Float64("cost_usd", cost),
		slog.Float64("session_total_usd", bt.currentSessionCost),
		slog.Float64("daily_total_usd", bt.currentDailyCost),
	)

	return nil
}

// checkBudgets checks if budgets are exceeded and generates alerts
func (bt *BudgetTracker) checkBudgets(ctx context.Context, cost float64, agentID string) {
	// Check session budget
	sessionPercent := bt.currentSessionCost / bt.config.SessionLimit
	if sessionPercent >= bt.config.CriticalThreshold {
		bt.addAlert(CostAlert{
			Level:     AlertCritical,
			Message:   fmt.Sprintf("Session cost critical: %.2f%% of limit", sessionPercent*100),
			Cost:      bt.currentSessionCost,
			Budget:    bt.config.SessionLimit,
			Threshold: bt.config.CriticalThreshold,
			Timestamp: time.Now(),
		}, ctx)
	} else if sessionPercent >= bt.config.WarningThreshold {
		bt.addAlert(CostAlert{
			Level:     AlertWarning,
			Message:   fmt.Sprintf("Session cost warning: %.2f%% of limit", sessionPercent*100),
			Cost:      bt.currentSessionCost,
			Budget:    bt.config.SessionLimit,
			Threshold: bt.config.WarningThreshold,
			Timestamp: time.Now(),
		}, ctx)
	}

	// Check daily budget
	dailyPercent := bt.currentDailyCost / bt.config.DailyLimit
	if dailyPercent >= bt.config.CriticalThreshold {
		bt.addAlert(CostAlert{
			Level:     AlertCritical,
			Message:   fmt.Sprintf("Daily cost critical: %.2f%% of limit", dailyPercent*100),
			Cost:      bt.currentDailyCost,
			Budget:    bt.config.DailyLimit,
			Threshold: bt.config.CriticalThreshold,
			Timestamp: time.Now(),
		}, ctx)
	} else if dailyPercent >= bt.config.WarningThreshold {
		bt.addAlert(CostAlert{
			Level:     AlertWarning,
			Message:   fmt.Sprintf("Daily cost warning: %.2f%% of limit", dailyPercent*100),
			Cost:      bt.currentDailyCost,
			Budget:    bt.config.DailyLimit,
			Threshold: bt.config.WarningThreshold,
			Timestamp: time.Now(),
		}, ctx)
	}
}

// addAlert adds an alert and logs it
func (bt *BudgetTracker) addAlert(alert CostAlert, ctx context.Context) {
	bt.alerts = append(bt.alerts, alert)

	// Maintain max alerts limit
	if len(bt.alerts) > bt.maxAlerts {
		bt.alerts = bt.alerts[len(bt.alerts)-bt.maxAlerts:]
	}

	// Log alert
	logging.GetLogger().InfoContext(ctx, "cost.alert",
		slog.String("event", "cost.alert"),
		slog.String("trace_id", logging.GetTraceID(ctx)),
		slog.String("level", string(alert.Level)),
		slog.String("message", alert.Message),
		slog.Float64("current_cost", alert.Cost),
		slog.Float64("budget_limit", alert.Budget),
		slog.Float64("threshold_percent", alert.Threshold*100),
	)
}

// GetCurrentSessionCost returns the current session cost
func (bt *BudgetTracker) GetCurrentSessionCost() float64 {
	bt.mu.RLock()
	defer bt.mu.RUnlock()
	return bt.currentSessionCost
}

// GetCurrentDailyCost returns the current daily cost
func (bt *BudgetTracker) GetCurrentDailyCost() float64 {
	bt.mu.RLock()
	defer bt.mu.RUnlock()
	return bt.currentDailyCost
}

// IsSessionBudgetExceeded checks if session budget is exceeded
func (bt *BudgetTracker) IsSessionBudgetExceeded() bool {
	bt.mu.RLock()
	defer bt.mu.RUnlock()
	return bt.currentSessionCost >= bt.config.SessionLimit
}

// IsDailyBudgetExceeded checks if daily budget is exceeded
func (bt *BudgetTracker) IsDailyBudgetExceeded() bool {
	bt.mu.RLock()
	defer bt.mu.RUnlock()
	return bt.currentDailyCost >= bt.config.DailyLimit
}

// GetSessionBudgetRemaining returns remaining session budget
func (bt *BudgetTracker) GetSessionBudgetRemaining() float64 {
	bt.mu.RLock()
	defer bt.mu.RUnlock()
	remaining := bt.config.SessionLimit - bt.currentSessionCost
	if remaining < 0 {
		return 0
	}
	return remaining
}

// GetDailyBudgetRemaining returns remaining daily budget
func (bt *BudgetTracker) GetDailyBudgetRemaining() float64 {
	bt.mu.RLock()
	defer bt.mu.RUnlock()
	remaining := bt.config.DailyLimit - bt.currentDailyCost
	if remaining < 0 {
		return 0
	}
	return remaining
}

// GetAlerts returns all recorded alerts
func (bt *BudgetTracker) GetAlerts() []CostAlert {
	bt.mu.RLock()
	defer bt.mu.RUnlock()

	// Return a copy
	alerts := make([]CostAlert, len(bt.alerts))
	copy(alerts, bt.alerts)
	return alerts
}

// GetCriticalAlerts returns only critical alerts
func (bt *BudgetTracker) GetCriticalAlerts() []CostAlert {
	bt.mu.RLock()
	defer bt.mu.RUnlock()

	var critical []CostAlert
	for _, alert := range bt.alerts {
		if alert.Level == AlertCritical {
			critical = append(critical, alert)
		}
	}
	return critical
}

// ClearAlerts clears all alerts
func (bt *BudgetTracker) ClearAlerts() {
	bt.mu.Lock()
	defer bt.mu.Unlock()
	bt.alerts = make([]CostAlert, 0)
}

// Reset resets the session budget (keeps daily budget)
func (bt *BudgetTracker) Reset(ctx context.Context) {
	bt.mu.Lock()
	defer bt.mu.Unlock()

	logging.GetLogger().InfoContext(ctx, "cost.budget_reset",
		slog.String("event", "cost.budget_reset"),
		slog.String("trace_id", logging.GetTraceID(ctx)),
		slog.Float64("session_cost_reset", bt.currentSessionCost),
	)

	bt.currentSessionCost = 0.0
}

// ResetDaily resets the daily budget
func (bt *BudgetTracker) ResetDaily(ctx context.Context) {
	bt.mu.Lock()
	defer bt.mu.Unlock()

	logging.GetLogger().InfoContext(ctx, "cost.daily_budget_reset",
		slog.String("event", "cost.daily_budget_reset"),
		slog.String("trace_id", logging.GetTraceID(ctx)),
		slog.Float64("daily_cost_reset", bt.currentDailyCost),
	)

	bt.currentDailyCost = 0.0
	bt.dailyResetTime = time.Now()
}

// GetCostSummary returns a summary of the current cost state
func (bt *BudgetTracker) GetCostSummary() map[string]interface{} {
	bt.mu.RLock()
	defer bt.mu.RUnlock()

	return map[string]interface{}{
		"session_cost":              bt.currentSessionCost,
		"session_budget":            bt.config.SessionLimit,
		"session_remaining":         bt.config.SessionLimit - bt.currentSessionCost,
		"session_percent":           (bt.currentSessionCost / bt.config.SessionLimit) * 100,
		"daily_cost":                bt.currentDailyCost,
		"daily_budget":              bt.config.DailyLimit,
		"daily_remaining":           bt.config.DailyLimit - bt.currentDailyCost,
		"daily_percent":             (bt.currentDailyCost / bt.config.DailyLimit) * 100,
		"alert_count":               len(bt.alerts),
		"critical_alert_count":      countCriticalAlerts(bt.alerts),
		"session_budget_exceeded":   bt.currentSessionCost >= bt.config.SessionLimit,
		"daily_budget_exceeded":     bt.currentDailyCost >= bt.config.DailyLimit,
	}
}

// countCriticalAlerts counts critical alerts
func countCriticalAlerts(alerts []CostAlert) int {
	count := 0
	for _, alert := range alerts {
		if alert.Level == AlertCritical {
			count++
		}
	}
	return count
}

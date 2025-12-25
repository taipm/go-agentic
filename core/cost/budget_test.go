package cost

import (
	"context"
	"testing"
)

func TestNewBudgetTracker(t *testing.T) {
	config := &BudgetConfig{
		DailyLimit:        100.0,
		SessionLimit:      50.0,
		WarningThreshold:  0.75,
		CriticalThreshold: 0.90,
	}

	tracker := NewBudgetTracker(config)

	if tracker.GetCurrentSessionCost() != 0.0 {
		t.Errorf("Initial session cost should be 0, got %f", tracker.GetCurrentSessionCost())
	}

	if tracker.GetCurrentDailyCost() != 0.0 {
		t.Errorf("Initial daily cost should be 0, got %f", tracker.GetCurrentDailyCost())
	}
}

func TestRecordCost(t *testing.T) {
	config := &BudgetConfig{
		DailyLimit:        100.0,
		SessionLimit:      50.0,
		WarningThreshold:  0.75,
		CriticalThreshold: 0.90,
	}

	tracker := NewBudgetTracker(config)
	ctx := context.Background()

	tracker.RecordCost(ctx, 10.0, "agent1")

	if tracker.GetCurrentSessionCost() != 10.0 {
		t.Errorf("Session cost should be 10.0, got %f", tracker.GetCurrentSessionCost())
	}

	if tracker.GetCurrentDailyCost() != 10.0 {
		t.Errorf("Daily cost should be 10.0, got %f", tracker.GetCurrentDailyCost())
	}
}

func TestBudgetExceeded(t *testing.T) {
	config := &BudgetConfig{
		DailyLimit:        100.0,
		SessionLimit:      50.0,
		WarningThreshold:  0.75,
		CriticalThreshold: 0.90,
	}

	tracker := NewBudgetTracker(config)
	ctx := context.Background()

	tracker.RecordCost(ctx, 55.0, "agent1")

	if !tracker.IsSessionBudgetExceeded() {
		t.Error("Session budget should be exceeded")
	}
}

func TestWarningAlert(t *testing.T) {
	config := &BudgetConfig{
		DailyLimit:        100.0,
		SessionLimit:      50.0,
		WarningThreshold:  0.75,
		CriticalThreshold: 0.90,
	}

	tracker := NewBudgetTracker(config)
	ctx := context.Background()

	// Record 38 USD (75% of 50) to trigger warning
	tracker.RecordCost(ctx, 38.0, "agent1")

	alerts := tracker.GetAlerts()
	if len(alerts) == 0 {
		t.Error("Expected at least one warning alert")
	}

	// Check that warning was triggered
	found := false
	for _, alert := range alerts {
		if alert.Level == AlertWarning {
			found = true
			break
		}
	}

	if !found {
		t.Error("Expected warning alert not found")
	}
}

func TestCriticalAlert(t *testing.T) {
	config := &BudgetConfig{
		DailyLimit:        100.0,
		SessionLimit:      50.0,
		WarningThreshold:  0.75,
		CriticalThreshold: 0.90,
	}

	tracker := NewBudgetTracker(config)
	ctx := context.Background()

	// Record 46 USD (90% of 50) to trigger critical
	tracker.RecordCost(ctx, 46.0, "agent1")

	alerts := tracker.GetAlerts()
	criticalCount := 0
	for _, alert := range alerts {
		if alert.Level == AlertCritical {
			criticalCount++
		}
	}

	if criticalCount == 0 {
		t.Error("Expected critical alert not found")
	}
}

func TestBudgetRemaining(t *testing.T) {
	config := &BudgetConfig{
		DailyLimit:        100.0,
		SessionLimit:      50.0,
		WarningThreshold:  0.75,
		CriticalThreshold: 0.90,
	}

	tracker := NewBudgetTracker(config)
	ctx := context.Background()

	tracker.RecordCost(ctx, 20.0, "agent1")

	remaining := tracker.GetSessionBudgetRemaining()
	expected := 30.0

	if remaining != expected {
		t.Errorf("Remaining budget should be %f, got %f", expected, remaining)
	}
}

func TestReset(t *testing.T) {
	config := &BudgetConfig{
		DailyLimit:        100.0,
		SessionLimit:      50.0,
		WarningThreshold:  0.75,
		CriticalThreshold: 0.90,
	}

	tracker := NewBudgetTracker(config)
	ctx := context.Background()

	tracker.RecordCost(ctx, 20.0, "agent1")
	tracker.Reset(ctx)

	if tracker.GetCurrentSessionCost() != 0.0 {
		t.Errorf("Session cost after reset should be 0, got %f", tracker.GetCurrentSessionCost())
	}

	if tracker.GetCurrentDailyCost() != 20.0 {
		t.Errorf("Daily cost after session reset should remain 20.0, got %f", tracker.GetCurrentDailyCost())
	}
}

func TestGetCostSummary(t *testing.T) {
	config := &BudgetConfig{
		DailyLimit:        100.0,
		SessionLimit:      50.0,
		WarningThreshold:  0.75,
		CriticalThreshold: 0.90,
	}

	tracker := NewBudgetTracker(config)
	ctx := context.Background()

	tracker.RecordCost(ctx, 25.0, "agent1")

	summary := tracker.GetCostSummary()

	if summary["session_cost"].(float64) != 25.0 {
		t.Errorf("Session cost in summary should be 25.0")
	}

	if summary["daily_cost"].(float64) != 25.0 {
		t.Errorf("Daily cost in summary should be 25.0")
	}
}

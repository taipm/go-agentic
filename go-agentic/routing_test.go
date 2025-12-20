package agentic

import (
	"testing"
)

// ============================================
// Trigger Detector Tests
// ============================================

func TestKeywordDetectorSingleKeyword(t *testing.T) {
	detector := NewKeywordDetector([]string{"resolved"}, false)

	if !detector.Detect("Issue is resolved") {
		t.Error("expected detector to match 'resolved'")
	}

	if detector.Detect("Issue still pending") {
		t.Error("expected detector to not match")
	}
}

func TestKeywordDetectorMultipleKeywords(t *testing.T) {
	detector := NewKeywordDetector([]string{"resolved", "completed", "done"}, false)

	if !detector.Detect("Task completed successfully") {
		t.Error("expected detector to match 'completed'")
	}

	if !detector.Detect("We are done here") {
		t.Error("expected detector to match 'done'")
	}

	if detector.Detect("Still working on it") {
		t.Error("expected detector to not match")
	}
}

func TestKeywordDetectorCaseSensitive(t *testing.T) {
	detector := NewKeywordDetector([]string{"ERROR"}, true)

	if !detector.Detect("ERROR: Critical issue") {
		t.Error("expected case-sensitive match")
	}

	if detector.Detect("error: not matching") {
		t.Error("expected case-sensitive to not match lowercase")
	}
}

func TestKeywordDetectorCaseInsensitive(t *testing.T) {
	detector := NewKeywordDetector([]string{"ERROR"}, false)

	if !detector.Detect("error: critical issue") {
		t.Error("expected case-insensitive to match")
	}

	if !detector.Detect("ERROR: critical issue") {
		t.Error("expected case-insensitive to match uppercase")
	}
}

func TestPatternDetector(t *testing.T) {
	detector, err := NewPatternDetector(`\[ERROR:\s*\d+\]`)
	if err != nil {
		t.Fatalf("failed to create pattern detector: %v", err)
	}

	if !detector.Detect("Error code [ERROR: 404]") {
		t.Error("expected pattern detector to match")
	}

	if detector.Detect("Error code [ERROR: abc]") {
		t.Error("expected pattern detector to not match non-digits")
	}
}

func TestPatternDetectorInvalidRegex(t *testing.T) {
	_, err := NewPatternDetector(`[invalid(regex`)
	if err == nil {
		t.Error("expected error for invalid regex")
	}
}

func TestSignalDetector(t *testing.T) {
	detector := NewSignalDetector("resolved")

	if !detector.Detect("Issue resolved. [SIGNAL: resolved]") {
		t.Error("expected signal detector to match [SIGNAL: resolved]")
	}

	if !detector.Detect("[SIGNAL:   resolved   ]") {
		t.Error("expected signal detector to handle spaces")
	}

	if detector.Detect("[SIGNAL: pending]") {
		t.Error("expected signal detector to not match different signal")
	}

	if detector.Detect("No signal here") {
		t.Error("expected signal detector to not match missing signal")
	}
}

func TestPrefixDetector(t *testing.T) {
	detector := NewPrefixDetector([]string{"HARDWARE:", "NETWORK:"}, false)

	if !detector.Detect("HARDWARE: CPU usage is high") {
		t.Error("expected prefix detector to match HARDWARE:")
	}

	if !detector.Detect("NETWORK: Connection timeout") {
		t.Error("expected prefix detector to match NETWORK:")
	}

	if detector.Detect("Issue with something") {
		t.Error("expected prefix detector to not match")
	}
}

func TestPrefixDetectorMultiLine(t *testing.T) {
	detector := NewPrefixDetector([]string{"ACTION:"}, false)

	response := `Analyzing issue...
ACTION: Escalate to hardware team`

	if !detector.Detect(response) {
		t.Error("expected prefix detector to match on second line")
	}
}

func TestAnyDetectorMatches(t *testing.T) {
	detector := NewAnyDetector(
		NewKeywordDetector([]string{"resolved"}, false),
		NewKeywordDetector([]string{"done"}, false),
	)

	if !detector.Detect("Issue is resolved") {
		t.Error("expected any detector to match first option")
	}

	if !detector.Detect("Task is done") {
		t.Error("expected any detector to match second option")
	}

	if detector.Detect("Still pending") {
		t.Error("expected any detector to not match neither option")
	}
}

func TestAnyDetectorEmpty(t *testing.T) {
	detector := NewAnyDetector()
	if detector.Detect("anything") {
		t.Error("expected empty any detector to not match")
	}
}

func TestAllDetectorMatches(t *testing.T) {
	detector := NewAllDetector(
		NewKeywordDetector([]string{"completed"}, false),
		NewKeywordDetector([]string{"confirmed"}, false),
	)

	if !detector.Detect("Task is completed and confirmed") {
		t.Error("expected all detector to match when all conditions met")
	}

	if detector.Detect("Task is completed") {
		t.Error("expected all detector to not match when missing second condition")
	}
}

func TestAlwaysDetector(t *testing.T) {
	detector := NewAlwaysDetector()

	if !detector.Detect("anything") {
		t.Error("expected always detector to always match")
	}

	if !detector.Detect("") {
		t.Error("expected always detector to match empty string")
	}
}

func TestNeverDetector(t *testing.T) {
	detector := NewNeverDetector()

	if detector.Detect("anything") {
		t.Error("expected never detector to never match")
	}

	if detector.Detect("") {
		t.Error("expected never detector to never match empty string")
	}
}

// ============================================
// Router Builder Tests
// ============================================

func TestNewRouter(t *testing.T) {
	router := NewRouter()

	if router == nil {
		t.Error("expected router to be created")
	}

	if len(router.rules) != 0 {
		t.Error("expected empty rules initially")
	}
}

func TestRouterRegisterAgents(t *testing.T) {
	router := NewRouter().
		RegisterAgents("agent1", "agent2", "agent3")

	if len(router.agents) != 3 {
		t.Errorf("expected 3 agents registered, got %d", len(router.agents))
	}

	if !router.agents["agent1"] {
		t.Error("expected agent1 to be registered")
	}
}

func TestRouterFromAgentBuildsRoute(t *testing.T) {
	router := NewRouter().
		RegisterAgents("router", "handler")

	routeBuilder := router.FromAgent("router")

	if routeBuilder == nil {
		t.Error("expected route builder to be created")
	}

	if routeBuilder.fromAgent != "router" {
		t.Errorf("expected fromAgent to be 'router', got '%s'", routeBuilder.fromAgent)
	}
}

func TestRouterAddRule(t *testing.T) {
	router := NewRouter().
		RegisterAgents("a1", "a2")

	rule := &RoutingRule{
		FromAgent:   "a1",
		Trigger:     NewAlwaysDetector(),
		TargetAgent: "a2",
		Description: "Test rule",
	}

	router.AddRule(rule)

	if len(router.rules) != 1 {
		t.Errorf("expected 1 rule, got %d", len(router.rules))
	}
}

func TestRouterFluentAPI(t *testing.T) {
	router := NewRouter().
		RegisterAgents("orchestrator", "handler", "resolver")

	config, err := router.
		FromAgent("orchestrator").
		To("handler", NewKeywordDetector([]string{"needs_handling"}, false)).
		Done().
		FromAgent("handler").
		To("resolver", NewSignalDetector("done")).
		Done().
		Build()

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if len(config.CompiledRules) != 2 {
		t.Errorf("expected 2 rules, got %d", len(config.CompiledRules))
	}
}

func TestRouterToWithDescription(t *testing.T) {
	router := NewRouter().
		RegisterAgents("a1", "a2")

	config, err := router.
		FromAgent("a1").
		ToWithDescription("a2", NewAlwaysDetector(), "Route to a2").
		Done().
		Build()

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	rule := config.CompiledRules[0]
	if rule.Description != "Route to a2" {
		t.Errorf("expected description 'Route to a2', got '%s'", rule.Description)
	}
}

func TestRouterValidateMissingFromAgent(t *testing.T) {
	router := NewRouter().
		RegisterAgents("a1", "a2")

	rule := &RoutingRule{
		FromAgent:   "",
		Trigger:     NewAlwaysDetector(),
		TargetAgent: "a2",
	}

	router.AddRule(rule)
	_, err := router.Build()

	if err == nil {
		t.Error("expected error for missing FromAgent")
	}
}

func TestRouterValidateInvalidFromAgent(t *testing.T) {
	router := NewRouter().
		RegisterAgents("a1", "a2")

	rule := &RoutingRule{
		FromAgent:   "invalid",
		Trigger:     NewAlwaysDetector(),
		TargetAgent: "a2",
	}

	router.AddRule(rule)
	_, err := router.Build()

	if err == nil {
		t.Error("expected error for unregistered FromAgent")
	}
}

func TestRouterValidateMissingTargetAgent(t *testing.T) {
	router := NewRouter().
		RegisterAgents("a1", "a2")

	rule := &RoutingRule{
		FromAgent:   "a1",
		Trigger:     NewAlwaysDetector(),
		TargetAgent: "",
	}

	router.AddRule(rule)
	_, err := router.Build()

	if err == nil {
		t.Error("expected error for missing TargetAgent")
	}
}

func TestRouterValidateMissingTrigger(t *testing.T) {
	router := NewRouter().
		RegisterAgents("a1", "a2")

	rule := &RoutingRule{
		FromAgent:   "a1",
		Trigger:     nil,
		TargetAgent: "a2",
	}

	router.AddRule(rule)
	_, err := router.Build()

	if err == nil {
		t.Error("expected error for missing Trigger")
	}
}

// ============================================
// RoutingConfig Tests
// ============================================

func TestRoutingConfigFindRoute(t *testing.T) {
	router := NewRouter().
		RegisterAgents("a1", "a2", "a3")

	config, err := router.
		FromAgent("a1").
		To("a2", NewKeywordDetector([]string{"issue"}, false)).
		Done().
		FromAgent("a1").
		To("a3", NewKeywordDetector([]string{"urgent"}, false)).
		Done().
		Build()

	if err != nil {
		t.Fatalf("failed to build config: %v", err)
	}

	// Should match first route
	route := config.FindRoute("a1", "There is an issue")
	if route == nil || route.TargetAgent != "a2" {
		t.Error("expected to find route to a2")
	}

	// Should match second route
	route = config.FindRoute("a1", "This is urgent")
	if route == nil || route.TargetAgent != "a3" {
		t.Error("expected to find route to a3")
	}

	// Should not match any route
	route = config.FindRoute("a1", "Nothing to do")
	if route != nil {
		t.Error("expected no matching route")
	}
}

func TestRoutingConfigGetRoutesForAgent(t *testing.T) {
	router := NewRouter().
		RegisterAgents("orchestrator", "handler1", "handler2")

	config, err := router.
		FromAgent("orchestrator").
		To("handler1", NewKeywordDetector([]string{"type1"}, false)).
		Done().
		FromAgent("orchestrator").
		To("handler2", NewKeywordDetector([]string{"type2"}, false)).
		Done().
		Build()

	if err != nil {
		t.Fatalf("failed to build config: %v", err)
	}

	routes := config.GetRoutesForAgent("orchestrator")
	if len(routes) != 2 {
		t.Errorf("expected 2 routes for orchestrator, got %d", len(routes))
	}

	routes = config.GetRoutesForAgent("handler1")
	if len(routes) != 0 {
		t.Errorf("expected 0 routes for handler1, got %d", len(routes))
	}
}

func TestRoutingConfigGetTargetAgent(t *testing.T) {
	router := NewRouter().
		RegisterAgents("a1", "a2")

	config, err := router.
		FromAgent("a1").
		To("a2", NewKeywordDetector([]string{"match"}, false)).
		Done().
		Build()

	if err != nil {
		t.Fatalf("failed to build config: %v", err)
	}

	target := config.GetTargetAgent("a1", "This should match")
	if target == nil || *target != "a2" {
		t.Error("expected target agent to be a2")
	}

	target = config.GetTargetAgent("a1", "No response here")
	if target != nil {
		t.Error("expected no target agent")
	}
}

// ============================================
// Integration Tests
// ============================================

func TestComplexRoutingScenario(t *testing.T) {
	// Simulate customer support routing:
	// orchestrator -> based on issue type -> specialized handler -> resolver

	router := NewRouter().
		RegisterAgents("orchestrator", "billing", "technical", "resolver")

	config, err := router.
		FromAgent("orchestrator").
		ToWithDescription("billing", NewKeywordDetector([]string{"payment", "invoice", "billing"}, false), "Route billing issues").
		Done().
		FromAgent("orchestrator").
		ToWithDescription("technical", NewKeywordDetector([]string{"error", "crash", "bug"}, false), "Route technical issues").
		Done().
		FromAgent("billing").
		ToWithDescription("resolver", NewSignalDetector("resolved"), "Billing resolved").
		Done().
		FromAgent("technical").
		ToWithDescription("resolver", NewSignalDetector("resolved"), "Technical resolved").
		Done().
		Build()

	if err != nil {
		t.Fatalf("failed to build routing: %v", err)
	}

	// Test: billing issue routing
	route := config.FindRoute("orchestrator", "I have a payment issue")
	if route == nil || route.TargetAgent != "billing" {
		t.Error("expected billing issue to route to billing specialist")
	}

	// Test: technical issue routing
	route = config.FindRoute("orchestrator", "The app keeps crashing")
	if route == nil || route.TargetAgent != "technical" {
		t.Error("expected technical issue to route to technical specialist")
	}

	// Test: resolution signal
	route = config.FindRoute("billing", "Issue resolved. [SIGNAL: resolved]")
	if route == nil || route.TargetAgent != "resolver" {
		t.Error("expected resolved signal to route to resolver")
	}
}

func TestTriggerDetectorDescription(t *testing.T) {
	tests := []struct {
		detector TriggerDetector
		name     string
	}{
		{NewKeywordDetector([]string{"test"}, false), "KeywordDetector"},
		{NewSignalDetector("signal"), "SignalDetector"},
		{NewAlwaysDetector(), "AlwaysDetector"},
		{NewNeverDetector(), "NeverDetector"},
	}

	for _, test := range tests {
		desc := test.detector.Description()
		if desc == "" {
			t.Errorf("%s: expected non-empty description", test.name)
		}
	}
}

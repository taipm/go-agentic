package agentic

import (
	"fmt"
)

// ============================================
// Router Builder Pattern
// ============================================

// RouterBuilder provides a fluent API for building routing configuration
type RouterBuilder struct {
	rules  []*RoutingRule
	agents map[string]bool // Track valid agent IDs
}

// NewRouter creates a new router builder
func NewRouter() *RouterBuilder {
	return &RouterBuilder{
		rules:  make([]*RoutingRule, 0),
		agents: make(map[string]bool),
	}
}

// RegisterAgents registers valid agent IDs for validation
func (rb *RouterBuilder) RegisterAgents(agentIDs ...string) *RouterBuilder {
	for _, id := range agentIDs {
		rb.agents[id] = true
	}
	return rb
}

// FromAgent starts defining routes from a specific agent
func (rb *RouterBuilder) FromAgent(agentID string) *RouteBuilder {
	// Always create a new route builder (don't cache)
	// This allows multiple route definitions from the same agent
	route := &RouteBuilder{
		router:    rb,
		fromAgent: agentID,
		rules:     make([]*RoutingRule, 0),
	}
	return route
}

// AddRule directly adds a routing rule
func (rb *RouterBuilder) AddRule(rule *RoutingRule) *RouterBuilder {
	if rule != nil {
		rb.rules = append(rb.rules, rule)
	}
	return rb
}

// Build validates and returns the routing configuration
func (rb *RouterBuilder) Build() (*RoutingConfig, error) {
	// Collect all rules (rules are added via Done() or AddRule())
	allRules := make([]*RoutingRule, len(rb.rules))
	copy(allRules, rb.rules)

	// Validate all rules
	for i, rule := range allRules {
		if err := rb.validateRule(rule); err != nil {
			return nil, fmt.Errorf("rule %d validation failed: %w", i, err)
		}
	}

	// Convert rules to UnifiedRoutingRule format for compatibility
	unifiedRules := make([]UnifiedRoutingRule, len(allRules))
	for i, rule := range allRules {
		targetAgent := rule.TargetAgent
		triggerDesc := ""
		if detector, ok := rule.Trigger.(TriggerDetector); ok {
			triggerDesc = detector.Description()
		}
		unifiedRules[i] = UnifiedRoutingRule{
			FromAgent:   rule.FromAgent,
			Trigger:     triggerDesc,
			TargetAgent: &targetAgent,
			Description: rule.Description,
		}
	}

	return &RoutingConfig{
		CompiledRules: allRules,
	}, nil
}

// validateRule checks if a routing rule is valid
func (rb *RouterBuilder) validateRule(rule *RoutingRule) error {
	if rule.FromAgent == "" {
		return fmt.Errorf("FromAgent is required")
	}

	if len(rb.agents) > 0 && !rb.agents[rule.FromAgent] {
		return fmt.Errorf("FromAgent '%s' is not registered", rule.FromAgent)
	}

	if rule.Trigger == nil {
		return fmt.Errorf("Trigger is required")
	}

	if rule.TargetAgent == "" {
		return fmt.Errorf("TargetAgent is required")
	}

	if len(rb.agents) > 0 && !rb.agents[rule.TargetAgent] {
		return fmt.Errorf("TargetAgent '%s' is not registered", rule.TargetAgent)
	}

	return nil
}

// ============================================
// Route Builder - Fluent API for single agent
// ============================================

// RouteBuilder defines routes from a specific agent
type RouteBuilder struct {
	router    *RouterBuilder
	fromAgent string
	rules     []*RoutingRule
}

// To defines a route to a target agent with a trigger detector
func (rb *RouteBuilder) To(targetAgent string, trigger TriggerDetector) *RouteBuilder {
	return rb.ToWithDescription(targetAgent, trigger, "")
}

// ToWithDescription defines a route with a description
func (rb *RouteBuilder) ToWithDescription(targetAgent string, trigger TriggerDetector, description string) *RouteBuilder {
	rule := &RoutingRule{
		FromAgent:   rb.fromAgent,
		Trigger:     trigger,
		TargetAgent: targetAgent,
		Description: description,
	}
	rb.rules = append(rb.rules, rule)
	return rb
}

// When is an alias for To (more readable)
func (rb *RouteBuilder) When(trigger TriggerDetector) *RouteWhenBuilder {
	return &RouteWhenBuilder{
		routeBuilder: rb,
		trigger:      trigger,
	}
}

// Done returns to the router builder
func (rb *RouteBuilder) Done() *RouterBuilder {
	// Add rules to main router
	rb.router.rules = append(rb.router.rules, rb.rules...)
	return rb.router
}

// ============================================
// Route When Builder - For readability
// ============================================

// RouteWhenBuilder provides readable syntax: When(detector).GoTo(agent)
type RouteWhenBuilder struct {
	routeBuilder *RouteBuilder
	trigger      TriggerDetector
}

// GoTo sets the target agent
func (rwb *RouteWhenBuilder) GoTo(targetAgent string) *RouteBuilder {
	return rwb.GoToWithDescription(targetAgent, "")
}

// GoToWithDescription sets the target agent with description
func (rwb *RouteWhenBuilder) GoToWithDescription(targetAgent string, description string) *RouteBuilder {
	return rwb.routeBuilder.ToWithDescription(targetAgent, rwb.trigger, description)
}

// ThenRoute chains multiple routes
func (rwb *RouteWhenBuilder) ThenRoute(trigger TriggerDetector) *RouteWhenBuilder {
	return &RouteWhenBuilder{
		routeBuilder: rwb.routeBuilder,
		trigger:      trigger,
	}
}

// ============================================
// RoutingConfig Methods - Routing Engine
// ============================================

// FindRoute finds the matching route for an agent response
func (rc *RoutingConfig) FindRoute(fromAgent string, response string) *RoutingRule {
	if rc == nil || len(rc.CompiledRules) == 0 {
		return nil
	}

	for _, rule := range rc.CompiledRules {
		detector, ok := rule.Trigger.(TriggerDetector)
		if !ok {
			continue
		}
		if rule.FromAgent == fromAgent && detector.Detect(response) {
			return rule
		}
	}
	return nil
}

// GetRoutesForAgent returns all routes from a specific agent
func (rc *RoutingConfig) GetRoutesForAgent(agentID string) []*RoutingRule {
	if rc == nil || len(rc.CompiledRules) == 0 {
		return make([]*RoutingRule, 0)
	}

	routes := make([]*RoutingRule, 0)
	for _, rule := range rc.CompiledRules {
		if rule.FromAgent == agentID {
			routes = append(routes, rule)
		}
	}
	return routes
}

// GetTargetAgent returns the next agent based on response, or nil if no match
func (rc *RoutingConfig) GetTargetAgent(fromAgent string, response string) *string {
	route := rc.FindRoute(fromAgent, response)
	if route != nil {
		return &route.TargetAgent
	}
	return nil
}

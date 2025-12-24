#!/bin/bash

# Analyze which crewai types each .go file uses
for file in *.go; do
  echo "=== $file ==="
  # Look for patterns like: Crew, Agent, Tool, Message, etc.
  grep -E "(Crew|Agent|Tool|Message|CrewExecutor|MetricsCollector|SignalRegistry|AgentResponse|AgentBehavior)" "$file" | grep -v "^//" | head -3
  echo ""
done

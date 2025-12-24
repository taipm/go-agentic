#!/bin/bash

echo "=== ANALYZING DEPENDENCIES ==="
echo ""

# Function to count internal references
count_refs() {
    local file=$1
    # Count references to crewai package types
    grep -o '\b\(Crew\|Agent\|Tool\|Message\|CrewExecutor\|MetricsCollector\|SignalRegistry\|TerminationSignal\|HardcodedDefaults\|TimeoutTracker\|RoutingSignal\|AgentBehavior\|ParallelGroupConfig\|ConfigMode\|StrictMode\|RoutingConfig\|AgentConfig\|CrewConfig\|AgentResponse\)\b' "$file" 2>/dev/null | wc -l
}

# Check each .go file in root
for file in *.go; do
    if [[ "$file" != "go.mod" && "$file" != "go.sum" ]]; then
        count=$(count_refs "$file")
        echo "$file: $count references to core types"
    fi
done

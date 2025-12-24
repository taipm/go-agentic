#!/usr/bin/env python3

import re
import os

# Key types defined in crewai package
CORE_TYPES = {
    'Crew', 'Agent', 'Tool', 'Message', 'CrewExecutor', 
    'MetricsCollector', 'SignalRegistry', 'TerminationSignal',
    'HardcodedDefaults', 'TimeoutTracker', 'RoutingSignal',
    'AgentBehavior', 'ParallelGroupConfig', 'ConfigMode', 'StrictMode',
    'RoutingConfig', 'AgentConfig', 'CrewConfig', 'AgentResponse',
    'StreamEvent', 'AgentCostMetrics', 'AgentMetadata', 'ToolResult',
    'ModelConfig', 'GracefulShutdownManager', 'MetricsSnapshot'
}

# Analyze each .go file
files_analysis = {}
for file in sorted(os.listdir('.')):
    if not file.endswith('.go') or file in ['go.mod', 'go.sum']:
        continue
    
    with open(file, 'r') as f:
        content = f.read()
    
    # Find all referenced types
    refs = []
    for type_name in CORE_TYPES:
        if re.search(r'\b' + re.escape(type_name) + r'\b', content):
            refs.append(type_name)
    
    files_analysis[file] = {
        'refs': refs,
        'count': len(refs),
        'has_func': 'func ' in content,
        'has_struct': 'type ' in content,
    }

# Sort by reference count
sorted_files = sorted(files_analysis.items(), key=lambda x: x[1]['count'])

print("=== FILES SORTED BY COMPLEXITY (independent first) ===\n")
for file, info in sorted_files[:15]:  # Show first 15
    print(f"{file:40} | Refs: {info['count']:2d} | Types: {', '.join(info['refs'][:3])}")
    if info['count'] <= 3:
        print(f"  ✅ CANDIDATE FOR MOVING")
    print()

print("\n=== INDEPENDENT FILES (0-2 references) ===\n")
for file, info in sorted_files:
    if info['count'] <= 2:
        print(f"  - {file}: {info['refs']}")

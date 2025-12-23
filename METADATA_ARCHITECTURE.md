# 🏗️ Agent Metadata Architecture

**Status:** ✅ IMPLEMENTED
**Date:** Dec 23, 2025
**Phase:** WEEK 2

---

## 📐 Complete Architecture Diagram

```
┌─────────────────────────────────────────────────────────────────┐
│                    AGENT STRUCT (Agent)                         │
│                                                                 │
│  ┌──────────────────────────────────────────────────────────┐  │
│  │ Core Fields (unchanged)                                  │  │
│  │  - ID, Name, Role, Backstory                             │  │
│  │  - Primary, Backup (ModelConfig)                         │  │
│  │  - SystemPrompt, Tools, Temperature                      │  │
│  │  - IsTerminal, HandoffTargets                            │  │
│  └──────────────────────────────────────────────────────────┘  │
│                                                                 │
│  ┌──────────────────────────────────────────────────────────┐  │
│  │ ✅ NEW: Unified Metadata (WEEK 2)                         │  │
│  │                                                          │  │
│  │  Metadata *AgentMetadata ──┐                             │  │
│  │                            │                             │  │
│  │  [Details below]           │                             │  │
│  └────────────────────────────┼──────────────────────────────┘  │
│                               │                                  │
│  ┌────────────────────────────▼──────────────────────────────┐  │
│  │ ✅ LEGACY: Backward Compatibility (WEEK 1)               │  │
│  │  - MaxTokensPerCall, MaxTokensPerDay, MaxCostPerDay      │  │
│  │  - CostAlertThreshold, EnforceCostLimits                 │  │
│  │  - CostMetrics (AgentCostMetrics)                        │  │
│  │                                                          │  │
│  │  ⚠️  NOTE: These fields are now also in Metadata         │  │
│  └────────────────────────────────────────────────────────────┘  │
└─────────────────────────────────────────────────────────────────┘
```

---

## 🔍 AgentMetadata Deep Dive

```
┌────────────────────────────────────────────────────────────────┐
│                   AGENT METADATA (AgentMetadata)               │
│                                                                │
│  ✅ Unified Metadata Hub for Comprehensive Agent Monitoring   │
└────────────────────────────────────────────────────────────────┘

┌─────────────────────────────────────────────────────────────────┐
│ 1. CORE IDENTIFIERS                                             │
├─────────────────────────────────────────────────────────────────┤
│                                                                 │
│  AgentID        string         Agent unique identifier          │
│  AgentName      string         Agent display name              │
│  CreatedTime    time.Time      When agent was created          │
│  LastAccessTime time.Time      Last execution time             │
│                                                                 │
└─────────────────────────────────────────────────────────────────┘

┌─────────────────────────────────────────────────────────────────┐
│ 2. CONFIGURATION & QUOTAS                                       │
├─────────────────────────────────────────────────────────────────┤
│                                                                 │
│  Quotas AgentQuotaLimits ─────────┐                             │
│                                   │                             │
│                                   ▼                             │
│                    ┌──────────────────────────┐                 │
│                    │  Cost Quotas             │                 │
│                    ├──────────────────────────┤                 │
│                    │ MaxTokensPerCall: 1000   │                 │
│                    │ MaxTokensPerDay: 50000   │                 │
│                    │ MaxCostPerDay: $10.00    │                 │
│                    │ CostAlertPercent: 80%    │                 │
│                    └──────────────────────────┘                 │
│                                                                 │
│                    ┌──────────────────────────┐                 │
│                    │  Memory Quotas           │                 │
│                    ├──────────────────────────┤                 │
│                    │ MaxMemoryPerCall: 512 MB │                 │
│                    │ MaxMemoryPerDay: 10 GB   │                 │
│                    │ MaxContextWindow: 32K    │                 │
│                    └──────────────────────────┘                 │
│                                                                 │
│                    ┌──────────────────────────┐                 │
│                    │  Execution Quotas        │                 │
│                    ├──────────────────────────┤                 │
│                    │ MaxCallsPerMinute: 60    │                 │
│                    │ MaxCallsPerHour: 1000    │                 │
│                    │ MaxCallsPerDay: 10000    │                 │
│                    │ MaxErrorsPerHour: 10     │                 │
│                    │ MaxErrorsPerDay: 50      │                 │
│                    └──────────────────────────┘                 │
│                                                                 │
│  EnforceCostLimits bool   Legacy flag for enforcement mode     │
│                                                                 │
└─────────────────────────────────────────────────────────────────┘

┌─────────────────────────────────────────────────────────────────┐
│ 3. RUNTIME METRICS (Updated during execution)                  │
├─────────────────────────────────────────────────────────────────┤
│                                                                 │
│  Cost AgentCostMetrics ────────┐                               │
│                                │                               │
│                                ▼                               │
│                    ┌──────────────────────────┐                │
│                    │  Cost Tracking           │                │
│                    ├──────────────────────────┤                │
│                    │ CallCount: 5             │ ← # of calls   │
│                    │ TotalTokens: 3500        │ ← tokens used  │
│                    │ DailyCost: $0.000525     │ ← $ spent      │
│                    │ LastResetTime: 2025-...  │ ← reset time   │
│                    │ Mutex: RWMutex           │ ← thread-safe  │
│                    └──────────────────────────┘                │
│                                                                 │
│  Memory AgentMemoryMetrics ────┐                               │
│                                │                               │
│                                ▼                               │
│                    ┌──────────────────────────┐                │
│                    │  Memory Usage            │                │
│                    ├──────────────────────────┤                │
│                    │ CurrentMemoryMB: 256     │                │
│                    │ PeakMemoryMB: 512        │                │
│                    │ AverageMemoryMB: 300     │                │
│                    │ MemoryTrendPercent: 5.0  │                │
│                    └──────────────────────────┘                │
│                    ┌──────────────────────────┐                │
│                    │  Context Window          │                │
│                    ├──────────────────────────┤                │
│                    │ CurrentContextSize: 8000 │                │
│                    │ MaxContextWindow: 32000  │                │
│                    │ ContextTrimPercent: 20%  │                │
│                    └──────────────────────────┘                │
│                    ┌──────────────────────────┐                │
│                    │  Call Metrics            │                │
│                    ├──────────────────────────┤                │
│                    │ AverageCallDuration: 2s  │                │
│                    │ SlowCallThreshold: 30s   │                │
│                    │ Mutex: RWMutex           │                │
│                    └──────────────────────────┘                │
│                                                                 │
│  Performance AgentPerformanceMetrics ──┐                       │
│                                        │                       │
│                                        ▼                       │
│                    ┌──────────────────────────┐                │
│                    │  Quality Metrics         │                │
│                    ├──────────────────────────┤                │
│                    │ SuccessfulCalls: 48      │                │
│                    │ FailedCalls: 2           │                │
│                    │ SuccessRate: 96.0%       │                │
│                    │ AverageResponseTime: 2s  │                │
│                    └──────────────────────────┘                │
│                    ┌──────────────────────────┐                │
│                    │  Error Tracking          │                │
│                    ├──────────────────────────┤                │
│                    │ LastError: "timeout"     │                │
│                    │ LastErrorTime: 2025-...  │                │
│                    │ ConsecutiveErrors: 0     │                │
│                    │ ErrorCountToday: 2       │                │
│                    │ MaxErrorsPerDay: 50      │                │
│                    └──────────────────────────┘                │
│                    ┌──────────────────────────┐                │
│                    │  Thresholds              │                │
│                    ├──────────────────────────┤                │
│                    │ MaxErrorsPerHour: 10     │                │
│                    │ MaxErrorsPerDay: 50      │                │
│                    │ MaxConsecutiveErrors: 5  │                │
│                    │ Mutex: RWMutex           │                │
│                    └──────────────────────────┘                │
│                                                                 │
└─────────────────────────────────────────────────────────────────┘

┌─────────────────────────────────────────────────────────────────┐
│ 4. SYNCHRONIZATION                                              │
├─────────────────────────────────────────────────────────────────┤
│                                                                 │
│  Mutex sync.RWMutex  Global mutex protecting ALL metrics       │
│                                                                 │
│  Usage:                                                         │
│   - RLock()  for reading (multiple readers)                    │
│   - Lock()   for writing (exclusive)                           │
│   - RUnlock() / Unlock() to release                            │
│                                                                 │
└─────────────────────────────────────────────────────────────────┘
```

---

## 🔄 Data Flow: Creation and Initialization

```
YAML Configuration
    │
    ▼
LoadAgentConfig()
    │
    ├─ Parse YAML file
    ├─ Set defaults
    └─ Return AgentConfig
    │
    ▼
CreateAgentFromConfig()
    │
    ├─ Create Primary ModelConfig
    ├─ Create Backup ModelConfig (if specified)
    │
    ├─ Create AgentMetadata
    │  ├─ Set identifiers (AgentID, AgentName, timestamps)
    │  │
    │  ├─ Initialize Quotas
    │  │  ├─ From YAML: Cost quotas
    │  │  └─ Defaults: Memory, execution, error quotas
    │  │
    │  ├─ Initialize Cost metrics
    │  │  ├─ CallCount: 0
    │  │  ├─ TotalTokens: 0
    │  │  ├─ DailyCost: 0
    │  │  └─ LastResetTime: time.Time{} (zero)
    │  │
    │  ├─ Initialize Memory metrics
    │  │  ├─ Current usage: 0
    │  │  ├─ Peak: 0
    │  │  ├─ Average: 0
    │  │  ├─ Trend: 0
    │  │  ├─ Context: 0 / 32000
    │  │  └─ Defaults: Max thresholds
    │  │
    │  ├─ Initialize Performance metrics
    │  │  ├─ Successful: 0
    │  │  ├─ Failed: 0
    │  │  ├─ SuccessRate: 100%
    │  │  ├─ No errors yet
    │  │  └─ Error thresholds
    │  │
    │  └─ Create Mutex
    │
    ├─ Create Agent struct
    │  ├─ Set from config (ID, Name, Role, Backstory, etc.)
    │  │
    │  ├─ Set Metadata pointer ← NEW
    │  │
    │  └─ Set legacy fields (for backward compatibility)
    │      ├─ MaxTokensPerCall, MaxTokensPerDay, MaxCostPerDay
    │      ├─ CostAlertThreshold, EnforceCostLimits
    │      └─ CostMetrics
    │
    └─ Return Agent
    │
    ▼
Agent Ready for Use
```

---

## 📊 Memory Layout

```
Agent struct (in memory)
├─ [ID, Name, Role, Backstory] ─────────────────┐
├─ [Primary, Backup] ─────────────────┐         │
├─ [SystemPrompt, Tools, Temperature] │         │
├─ [IsTerminal, HandoffTargets]       │         │
│                                      │         │
├─ Metadata *AgentMetadata ───┐       │         │
│                             │       │         │
│  (points to heap memory)    │       │         │
│                             │       │         │
│  AgentMetadata {            │       │         │
│    AgentID: "agent-1"       │       │         │
│    AgentName: "Clarifier"   │       │         │
│    CreatedTime: time.Time   │       │         │
│    LastAccessTime: time.Time│       │         │
│                             │       │         │
│    Quotas: {                │       │         │
│      MaxTokensPerCall: 1000 │       │         │
│      MaxTokensPerDay: 50000 │       │         │
│      ... (13 fields)        │       │         │
│    }                        │       │         │
│                             │       │         │
│    Cost: {                  │       │         │
│      CallCount: 5           │       │         │
│      TotalTokens: 3500      │       │         │
│      DailyCost: 0.000525    │       │         │
│      LastResetTime: ...     │       │         │
│      Mutex: RWMutex         │       │         │
│    }                        │       │         │
│                             │       │         │
│    Memory: {                │       │         │
│      CurrentMemoryMB: 256   │       │         │
│      ... (12 fields)        │       │         │
│      Mutex: RWMutex         │       │         │
│    }                        │       │         │
│                             │       │         │
│    Performance: {           │       │         │
│      SuccessfulCalls: 48    │       │         │
│      ... (11 fields)        │       │         │
│      Mutex: RWMutex         │       │         │
│    }                        │       │         │
│                             │       │         │
│    Mutex: RWMutex           │       │         │
│  }                          │       │         │
│                             ▼       │         │
│  (Heap memory - ~5KB per agent)     │         │
│                                     │         │
├─ (Legacy fields - for backward compat)       │
│  ├─ MaxTokensPerCall: 1000          │         │
│  ├─ MaxTokensPerDay: 50000          │         │
│  ├─ MaxCostPerDay: 10.0             │         │
│  ├─ CostAlertThreshold: 0.8         │         │
│  ├─ EnforceCostLimits: false        │         │
│  └─ CostMetrics { ... }  ← Same as Metadata.Cost
│
└─ (Stack memory - ~200 bytes per agent)
```

---

## 🔐 Thread Safety Model

```
Multiple Goroutines ─┬─ Read Quota
                    ├─ Read Cost
                    ├─ Update Memory
                    ├─ Update Performance
                    └─ ...
                    │
                    ▼
              AgentMetadata.Mutex
                    │
        ┌───────────┼───────────┐
        │           │           │
        ▼           ▼           ▼
    RLock()     RLock()     Lock() ← exclusive for writes
  (shared)    (shared)     (exclusive)
        │           │           │
        ├─ Read  ─┤ Read  ─┤ Read/Write
        │ Quotas   │ Cost    │ All fields
        │ Cost     │ Memory  │
        │ Memory   │ ...     │
        │ Perf     │         │
        │           │           │
        └────┬──────┴─────┬─────┘
             │           │
             ▼           ▼
       RUnlock()    Unlock()
    (still mutex)  (release)
        │           │
        └─────┬─────┘
              │
              ▼
        Next operation
```

**Locking Strategy:**
- Multiple readers can access simultaneously (RLock)
- Only one writer at a time (Lock)
- Writers block readers and vice versa
- Always defer Unlock() to prevent deadlocks

---

## 🔄 Integration Points

### 1. Agent Creation
```
CreateAgentFromConfig()
    └─ Initializes AgentMetadata with quotas
    └─ Sets all metric values to defaults
    └─ Creates RWMutex for synchronization
```

### 2. Agent Execution (Current - WEEK 1)
```
agent.Execute()
    ├─ CheckCostLimits(agent, tokens)
    │  └─ Uses agent.MaxTokensPerCall, etc. (LEGACY)
    │
    ├─ CallLLM()
    │  └─ Returns response
    │
    └─ UpdateCostMetrics(agent, tokens, cost)
       └─ Updates agent.CostMetrics (LEGACY)
```

### 3. Agent Execution (Future - WEEK 2+)
```
agent.Execute()
    ├─ CheckCostLimits(agent)
    │  └─ Uses agent.Metadata.Quotas.MaxTokensPerCall
    │
    ├─ CallLLM()
    │  └─ Returns response
    │
    ├─ UpdateCostMetrics(agent, cost)
    │  └─ Updates agent.Metadata.Cost
    │
    ├─ UpdateMemoryMetrics(agent, memory)
    │  └─ Updates agent.Metadata.Memory (NEW)
    │
    └─ UpdatePerformanceMetrics(agent, duration, success)
       └─ Updates agent.Metadata.Performance (NEW)
```

---

## 📈 Quota Enforcement Hierarchy

```
Agent.Metadata.Quotas
├── COST QUOTAS
│   ├─ Per-Call: MaxTokensPerCall (1000)
│   │   └─ Checked BEFORE execution
│   │   └─ Returns error if exceeded (if EnforceQuotas=true)
│   │
│   ├─ Per-Day: MaxTokensPerDay (50000)
│   │   └─ Checked BEFORE execution
│   │   └─ Resets daily
│   │
│   └─ Per-Day: MaxCostPerDay ($10)
│       └─ Checked BEFORE execution
│       └─ Returns error if exceeded
│
├── MEMORY QUOTAS
│   ├─ Per-Call: MaxMemoryPerCall (512 MB)
│   │   └─ Tracked during execution
│   │   └─ Alert if exceeded
│   │
│   ├─ Per-Day: MaxMemoryPerDay (10 GB)
│   │   └─ Tracked during execution
│   │   └─ Alert if exceeded
│   │
│   └─ Context: MaxContextWindow (32K tokens)
│       └─ Tracked during execution
│       └─ Auto-trim if exceeded
│
├── EXECUTION QUOTAS
│   ├─ Per-Minute: MaxCallsPerMinute (60)
│   │   └─ Rate limiting
│   │
│   ├─ Per-Hour: MaxCallsPerHour (1000)
│   │   └─ Rate limiting
│   │
│   └─ Per-Day: MaxCallsPerDay (10000)
│       └─ Rate limiting
│
└── ERROR QUOTAS
    ├─ Per-Hour: MaxErrorsPerHour (10)
    │   └─ Alert if exceeded
    │
    ├─ Per-Day: MaxErrorsPerDay (50)
    │   └─ Block if exceeded
    │
    └─ Consecutive: MaxConsecutiveErrors (5)
        └─ Block if exceeded
```

---

## 🎯 Access Patterns

### Pattern 1: Read Multiple Metrics (Common)
```go
agent.Metadata.Mutex.RLock()
defer agent.Metadata.Mutex.RUnlock()

// Safe reads
callCount := agent.Metadata.Cost.CallCount
dailyCost := agent.Metadata.Cost.DailyCost
memoryUsed := agent.Metadata.Memory.CurrentMemoryMB
successRate := agent.Metadata.Performance.SuccessRate
```

### Pattern 2: Quota Check (Before Execution)
```go
agent.Metadata.Mutex.RLock()
defer agent.Metadata.Mutex.RUnlock()

// Check against quota
if estimatedTokens > agent.Metadata.Quotas.MaxTokensPerCall {
    return fmt.Errorf("token limit exceeded")
}
```

### Pattern 3: Update Metrics (After Execution)
```go
agent.Metadata.Mutex.Lock()
defer agent.Metadata.Mutex.Unlock()

// Safe updates
agent.Metadata.Cost.CallCount++
agent.Metadata.Cost.TotalTokens += estimatedTokens
agent.Metadata.Cost.DailyCost += actualCost
```

---

## 🚀 Scalability Considerations

### Memory Footprint
- **Per Agent**: ~5KB for Metadata + ~200 bytes for Legacy fields
- **100 Agents**: ~500 KB total overhead
- **1000 Agents**: ~5 MB total overhead (negligible)

### Mutex Performance
- **RWMutex**: Optimized for read-heavy workloads
- **Read Operations**: Multiple goroutines can read simultaneously
- **Write Operations**: Exclusive access, brief critical section
- **Expected Contention**: Low (metrics updated infrequently)

### Scalability Path
```
Current (WEEK 2):
  ├─ Per-agent metrics
  └─ RWMutex per agent

Future (WEEK 3+):
  ├─ Crew-level metrics aggregation
  ├─ Optional metrics persistence
  └─ Optional metrics export (Prometheus, etc.)
```

---

## ✅ Status Summary

- ✅ AgentMetadata structure implemented
- ✅ Four metric types defined
- ✅ CreateAgentFromConfig enhanced
- ✅ Thread-safe with RWMutex
- ✅ Sensible defaults for all quotas
- ✅ Backward compatible
- ✅ Build verified (zero errors)
- ✅ Tests verified (100% pass)

**Ready for memory and performance tracking implementation in next phase.**

---

**Document:** Agent Metadata Architecture
**Version:** WEEK 2
**Status:** ✅ COMPLETE


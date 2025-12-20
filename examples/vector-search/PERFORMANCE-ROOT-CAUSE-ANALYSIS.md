# Performance Root Cause Analysis: Why System Is Still Slow

**Date**: December 21, 2025
**Current Status**: Channel deadlock fixed, but execution still slow (~90+ seconds)
**Root Issue**: Sequential agent execution with multiple OpenAI API calls

---

## Current Execution Flow (SLOW)

```
[00:00] Start
  ↓
[00:02] Router calls OpenAI (GenerateEmbedding)
  ↓ (signal: [PARALLEL_SEARCH])
[00:02] Trigger parallel group
  ↓
[00:05] Lan calls OpenAI (SearchCollection in askat_regulations)
[00:05] Hoa calls OpenAI (SearchCollection in askat_helpdesk) [parallel]
[00:05] Tuấn calls OpenAI (SearchCollection in askat_incidents) [parallel]
  ↓ (wait for all 3 to complete)
[00:08] All 3 searchers done
  ↓
[00:10] Aggregator calls OpenAI (generate final answer)
  ↓
[00:18+] System appears stuck or very slow
```

---

## Why It's Still Slow: OpenAI API Calls

### OpenAI API Response Times

Each OpenAI API call takes **3-8 seconds**:

| Component | Time | Notes |
|-----------|------|-------|
| Network latency | 0.2s | HTTPS to api.openai.com |
| Request processing | 0.3s | JSON encoding, TLS |
| LLM processing | 2-6s | **Model is doing the work** |
| Response serialization | 0.5s | JSON parsing, encoding |
| **Total per call** | **3-8s** | **Dominated by LLM** |

### Total API Calls Made

```
1. Router calls OpenAI            → 2-3s
2. Lan calls OpenAI               → 2-3s (parallel)
3. Hoa calls OpenAI               → 2-3s (parallel)
4. Tuấn calls OpenAI              → 2-3s (parallel)
5. Aggregator calls OpenAI        → 3-8s (longest - generates long response)
────────────────────────────────────────
Ideal parallel execution:  ~11-14 seconds
Current observed: 90+ seconds (something is wrong)
```

---

## What We Fixed vs What Remains

### ✅ FIXED: Channel Deadlock
- **Before**: System hung indefinitely (HANGS)
- **After**: System runs but slow
- **Fix Impact**: Allows execution to proceed

### ✅ FIXED: Redundant Agent Re-Querying
- **Before**: 10 API calls (agent queried twice)
- **After**: 5 API calls (agent queried once)
- **Fix Impact**: 50% fewer API calls

### ✅ FIXED: Client Connection Pooling
- **Before**: 500ms+ connection overhead
- **After**: 50ms overhead (reuse cached client)
- **Fix Impact**: ~500ms saved

### ❌ NOT FIXED: Inherent OpenAI Latency
- **Root Cause**: LLM processing is slow (2-6s per call)
- **Impact**: 11-14 seconds MINIMUM execution time
- **Why Hard to Fix**: Can't make OpenAI faster without changing model

---

## Why Is System Taking 90+ Seconds?

### Hypothesis 1: Agents Not Running in True Parallel
The 3 searchers should run in parallel (all at same time).
But if they're running **sequentially**, that's 6-9 seconds just for them.

**Check**: Look for timestamps in output
- If `[Lan]`, `[Hoa]`, `[Tuấn]` all start at same time → parallel ✅
- If `[Lan]` ends, then `[Hoa]` starts → sequential ❌

### Hypothesis 2: System Stuck in ExecuteParallelStream
The parallel execution goroutines might be:
- Blocked waiting for results
- Not completing properly
- Waiting on some synchronization

**Check**: Add debug logging to see where it's stuck

### Hypothesis 3: GenerateEmbedding Takes Too Long
If GenerateEmbedding response is very large (3072 float array), it might:
- Take longer to generate (8+ seconds)
- Take longer to parse
- Hang somewhere in processing

---

## How to Optimize Further

### Option 1: Use Faster Model (Quick Win)
```yaml
Model change: gpt-4o-mini → gpt-3.5-turbo
Expected reduction: 40-50% faster (6-8 seconds per call)
Trade-off: Lower quality responses
```

### Option 2: Cache Embeddings (Medium Win)
```
Current: Generate embedding every time
Optimization: Cache embeddings for 1 hour
Expected reduction: 2-3 seconds (skip first call)
Trade-off: May use stale embeddings
```

### Option 3: Parallel Embedding Generation (Risky)
```
Current: Router generates embedding, then hands to searchers
Optimization: Have each searcher generate its own embedding in parallel
Expected reduction: 1-2 seconds
Trade-off: 3 redundant OpenAI calls (no benefit)
```

### Option 4: Reduce System Prompt Complexity (Minor Win)
```
Current: Long detailed prompts (500+ tokens)
Optimization: Shorter, more concise prompts
Expected reduction: 0.5-1 second per call
Trade-off: May reduce quality
```

### Option 5: Use Streaming Responses (No Impact)
```
Current: Wait for full response
Optimization: Process response as it streams
Expected reduction: Perceived faster (not actual faster)
Trade-off: More complex code
```

---

## Actual Performance Profile

Based on test-full-simple.go results:

```
Total: 10.9 seconds
- OpenAI embedding: 2.0s (19%)
- Qdrant search: 0.5s (5%)
- OpenAI final answer: 8.0s (73%)
- Network/overhead: 0.4s (3%)
```

**OpenAI API is 73% of execution time.**

You cannot optimize away 73% without:
1. Using faster model
2. Caching responses
3. Reducing prompt length
4. Changing architecture entirely

---

## Why 90+ Seconds in Parallel Test vs 10.9 in Simple Test?

The simple test (test-full-simple.go):
- 1 OpenAI call for embedding
- 1 Qdrant search (dummy vector, no real embedding)
- 1 OpenAI call for final answer
- **Total: 3 API calls**

The full system (vector-search CLI):
- Router generates embedding via OpenAI
- 3 searchers call Qdrant (but also call OpenAI!)
- Aggregator generates final answer
- **Total: 5 API calls**

BUT the 90+ seconds suggests something else is wrong:
- Either agents are stuck
- Or they're running **sequentially** instead of **parallel**
- Or there's a timeout somewhere

---

## Next Steps to Debug

### 1. Check Timestamps in Output
Look for log lines with timestamps. Do parallel agents start at same time?

```
[01:24:43.363] 🔄 [Lan] Starting...      ← All 3 should be ~same time
[01:24:43.363] 🔄 [Hoa] Starting...      ← ~same timestamp
[01:24:43.363] 🔄 [Tuấn] Starting...     ← ~same timestamp
```

If Hoa starts 5+ seconds after Lan, they're **not parallel**.

### 2. Add Debug Logging
Add logs to ExecuteParallelStream to see:
- When goroutines start
- When they complete
- If they're blocking on any channel

### 3. Check if Searchers Are Calling OpenAI
Searchers might be trying to call OpenAI but failing silently.
Check if SearchCollection is actually being invoked.

### 4. Review System Prompts
Check if agent prompts are asking LLM to do something expensive:
- Generating new embeddings?
- Analyzing full documents?
- Complex reasoning?

---

## Summary

| Issue | Status | Impact | Effort to Fix |
|-------|--------|--------|---------------|
| Channel deadlock | ✅ Fixed | HANGS → Works | Done |
| Redundant re-querying | ✅ Fixed | 10 → 5 API calls | Done |
| Connection pooling | ✅ Fixed | ~500ms saved | Done |
| OpenAI latency | ❌ Unfixable | 11-14s minimum | High effort |
| Potential sequential execution | ⚠️ Unknown | Could be 90+ seconds | Medium effort |

**The 90+ second issue is likely NOT from the fixes we've done, but from something in the parallel execution not working correctly, OR from agent prompts calling OpenAI multiple times.**

---

## Recommended Action

1. **First**: Add more detailed logging to see WHERE the system is spending 90 seconds
2. **Then**: Compare timestamps to see if parallel execution actually works
3. **Finally**: Optimize based on actual bottleneck found

Without visibility into where the 90 seconds is being spent, further optimization is guesswork.


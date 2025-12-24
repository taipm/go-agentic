# Signal-Based Routing Documentation Index

Complete documentation suite for signal-based routing in Go-CrewAI framework.

---

## 📚 Documentation Suite

### 1. **[SIGNAL_BASED_ROUTING_ANALYSIS.md](./SIGNAL_BASED_ROUTING_ANALYSIS.md)** - Comprehensive Analysis
**Level:** Advanced | **Length:** 15 sections | **Best for:** Deep understanding

Complete technical analysis covering:
- Signal definition and architecture
- Signal matching algorithm (3-level matching)
- Routing flow execution (complete state machine)
- Context preservation through handoffs
- Handoff limits and enforcement
- Parallel group execution
- Timeout and behavior signals
- Error handling and edge cases
- Best practices and design patterns

**Key sections:**
- [Section 1: Signal Definition](./SIGNAL_BASED_ROUTING_ANALYSIS.md#1-signal-definition--architecture)
- [Section 2: Signal Matching](./SIGNAL_BASED_ROUTING_ANALYSIS.md#2-signal-matching--detection)
- [Section 3: Routing Flow](./SIGNAL_BASED_ROUTING_ANALYSIS.md#3-routing-flow-execution)
- [Section 4: Context Preservation](./SIGNAL_BASED_ROUTING_ANALYSIS.md#4-context-preservation-through-handoffs)

---

### 2. **[SIGNAL_ROUTING_DIAGRAM.md](./SIGNAL_ROUTING_DIAGRAM.md)** - Visual Reference
**Level:** Intermediate | **Length:** 9 diagrams | **Best for:** Visual learners

ASCII diagrams and flowcharts showing:
1. Signal matching algorithm flowchart
2. ExecuteStream main loop flowchart
3. Signal matching examples
4. History preservation timeline
5. Handoff limit enforcement
6. Signal definition flow
7. Performance bottleneck analysis
8. Error recovery flow
9. Parallel group execution model

**Best for:** Visualizing flow and understanding state transitions

---

### 3. **[SIGNAL_ROUTING_GUIDE.md](./SIGNAL_ROUTING_GUIDE.md)** - Implementation Guide
**Level:** Beginner to Intermediate | **Length:** 8 sections | **Best for:** Implementation

Practical guide covering:
- Quick start configuration
- Signal design guide (naming conventions, scope, patterns)
- Implementation checklist
- Use case implementations (sequential, decision tree, parallel, etc)
- Debugging guide with common issues
- Performance optimization tips
- Production checklist
- Working examples

**Key sections:**
- [Section 1: Signal Design Guide](./SIGNAL_ROUTING_GUIDE.md#1-signal-design-guide)
- [Section 2: Implementation Checklist](./SIGNAL_ROUTING_GUIDE.md#2-implementation-checklist)
- [Section 3: Use Case Implementations](./SIGNAL_ROUTING_GUIDE.md#3-signal-design-for-specific-use-cases)

---

### 4. **[SIGNAL_ROUTING_FAQ.md](./SIGNAL_ROUTING_FAQ.md)** - Q&A Reference
**Level:** Beginner | **Length:** 30+ Q&A pairs | **Best for:** Quick answers

Frequently asked questions covering:
- Core concepts
- Configuration
- Execution and flow
- Handoffs and limits
- Parallel execution
- Behavior and signals
- Debugging
- Performance
- Common mistakes

**Quick reference table:** Common questions with instant answers

---

### 5. **[SIGNAL_ROUTING_QUICK_REF.md](./SIGNAL_ROUTING_QUICK_REF.md)** - Quick Reference
**Level:** Beginner | **Length:** One page | **Best for:** Quick lookup

One-page reference including:
- Signal matching (3-level)
- YAML config template
- Signal types table
- Execution order flowchart
- Common patterns
- Debugging checklist
- Performance notes
- Common mistakes
- Key functions

**Best for:** Quick lookup while coding

---

## 🎯 How to Use This Suite

### By Experience Level

**Beginner (Just starting with signal routing):**
1. Read [SIGNAL_ROUTING_QUICK_REF.md](./SIGNAL_ROUTING_QUICK_REF.md) - 5 minutes
2. Skim [SIGNAL_ROUTING_GUIDE.md](./SIGNAL_ROUTING_GUIDE.md) Section 1 - 10 minutes
3. Look at [examples/01-quiz-exam/config/crew.yaml](./examples/01-quiz-exam/config/crew.yaml) - 5 minutes
4. Implement your config following checklist

**Intermediate (Understanding existing implementation):**
1. Read [SIGNAL_ROUTING_ANALYSIS.md](./SIGNAL_BASED_ROUTING_ANALYSIS.md) Sections 1-4 - 20 minutes
2. Review [SIGNAL_ROUTING_DIAGRAM.md](./SIGNAL_ROUTING_DIAGRAM.md) - 15 minutes
3. Browse [SIGNAL_ROUTING_GUIDE.md](./SIGNAL_ROUTING_GUIDE.md) Section 3 (use cases) - 20 minutes
4. Read [core/crew_routing.go](./core/crew_routing.go) - 15 minutes

**Advanced (Debugging or extending):**
1. Deep read [SIGNAL_BASED_ROUTING_ANALYSIS.md](./SIGNAL_BASED_ROUTING_ANALYSIS.md) all sections - 45 minutes
2. Study all diagrams in [SIGNAL_ROUTING_DIAGRAM.md](./SIGNAL_ROUTING_DIAGRAM.md) - 20 minutes
3. Read [core/crew.go](./core/crew.go) lines 775-873 (execution loop) - 15 minutes
4. Review [core/crew_routing.go](./core/crew_routing.go) entire file - 20 minutes

### By Task

| Task | Start With |
|------|-----------|
| **I need to implement signal routing** | [SIGNAL_ROUTING_QUICK_REF.md](./SIGNAL_ROUTING_QUICK_REF.md) + [SIGNAL_ROUTING_GUIDE.md](./SIGNAL_ROUTING_GUIDE.md) Section 1-2 |
| **Signal not being detected (debug)** | [SIGNAL_ROUTING_GUIDE.md](./SIGNAL_ROUTING_GUIDE.md) Section 4 (Debugging) |
| **Want to understand the algorithm** | [SIGNAL_BASED_ROUTING_ANALYSIS.md](./SIGNAL_BASED_ROUTING_ANALYSIS.md) Sections 2-3 |
| **Need visual explanation** | [SIGNAL_ROUTING_DIAGRAM.md](./SIGNAL_ROUTING_DIAGRAM.md) |
| **Quick lookup while coding** | [SIGNAL_ROUTING_QUICK_REF.md](./SIGNAL_ROUTING_QUICK_REF.md) or [SIGNAL_ROUTING_FAQ.md](./SIGNAL_ROUTING_FAQ.md) |
| **I have a specific use case** | [SIGNAL_ROUTING_GUIDE.md](./SIGNAL_ROUTING_GUIDE.md) Section 3 (Use Cases) |
| **Production deployment** | [SIGNAL_ROUTING_GUIDE.md](./SIGNAL_ROUTING_GUIDE.md) Section 6 (Production Checklist) |

---

## 📖 Reading Paths

### Path 1: "Just Give Me the Essentials" (20 minutes)
1. [SIGNAL_ROUTING_QUICK_REF.md](./SIGNAL_ROUTING_QUICK_REF.md)
2. [SIGNAL_ROUTING_GUIDE.md](./SIGNAL_ROUTING_GUIDE.md) - Sections 1-2
3. Pick an example from [examples/01-quiz-exam](./examples/01-quiz-exam/config/crew.yaml)

### Path 2: "I Want to Understand It Properly" (60 minutes)
1. [SIGNAL_ROUTING_QUICK_REF.md](./SIGNAL_ROUTING_QUICK_REF.md) - 5 min
2. [SIGNAL_ROUTING_DIAGRAM.md](./SIGNAL_ROUTING_DIAGRAM.md) Sections 1-4 - 15 min
3. [SIGNAL_BASED_ROUTING_ANALYSIS.md](./SIGNAL_BASED_ROUTING_ANALYSIS.md) Sections 1-5 - 30 min
4. [SIGNAL_ROUTING_GUIDE.md](./SIGNAL_ROUTING_GUIDE.md) Section 3 - 10 min

### Path 3: "I Need to Debug This" (45 minutes)
1. [SIGNAL_ROUTING_GUIDE.md](./SIGNAL_ROUTING_GUIDE.md) Section 4 (Debugging) - 15 min
2. [SIGNAL_ROUTING_FAQ.md](./SIGNAL_ROUTING_FAQ.md) (Debugging section) - 10 min
3. [core/crew_routing.go](./core/crew_routing.go) - trace the code - 20 min

### Path 4: "Complete Mastery" (2-3 hours)
Read in this order:
1. [SIGNAL_ROUTING_QUICK_REF.md](./SIGNAL_ROUTING_QUICK_REF.md)
2. [SIGNAL_ROUTING_DIAGRAM.md](./SIGNAL_ROUTING_DIAGRAM.md) - all sections
3. [SIGNAL_BASED_ROUTING_ANALYSIS.md](./SIGNAL_BASED_ROUTING_ANALYSIS.md) - all sections
4. [SIGNAL_ROUTING_GUIDE.md](./SIGNAL_ROUTING_GUIDE.md) - all sections
5. [SIGNAL_ROUTING_FAQ.md](./SIGNAL_ROUTING_FAQ.md) - all sections
6. Source code: [core/crew_routing.go](./core/crew_routing.go) + [core/crew.go](./core/crew.go) lines 775-873

---

## 🔍 Quick Navigation by Topic

### Core Concepts
- **What is a signal?** → [SIGNAL_BASED_ROUTING_ANALYSIS.md §1.1](./SIGNAL_BASED_ROUTING_ANALYSIS.md#11-定义signal)
- **How does matching work?** → [SIGNAL_BASED_ROUTING_ANALYSIS.md §2.1](./SIGNAL_BASED_ROUTING_ANALYSIS.md#21-signal-matching-logic) or [SIGNAL_ROUTING_DIAGRAM.md §1](./SIGNAL_ROUTING_DIAGRAM.md#1-signal-matching-algorithm)
- **What are signal types?** → [SIGNAL_ROUTING_QUICK_REF.md](./SIGNAL_ROUTING_QUICK_REF.md) (Signal Types table)

### Configuration
- **Minimum config?** → [SIGNAL_ROUTING_FAQ.md](./SIGNAL_ROUTING_FAQ.md#q-whats-the-minimum-crewayaml-config-for-signal-routing)
- **YAML template** → [SIGNAL_ROUTING_GUIDE.md §2.2](./SIGNAL_ROUTING_GUIDE.md#22-agent-instruction-checklist) or [SIGNAL_ROUTING_QUICK_REF.md](./SIGNAL_ROUTING_QUICK_REF.md)
- **Signal naming** → [SIGNAL_ROUTING_GUIDE.md §1.1](./SIGNAL_ROUTING_GUIDE.md#11-naming-conventions)

### Execution
- **Execution order** → [SIGNAL_ROUTING_DIAGRAM.md §2](./SIGNAL_ROUTING_DIAGRAM.md#2-executestream-main-loop) or [SIGNAL_ROUTING_QUICK_REF.md](./SIGNAL_ROUTING_QUICK_REF.md)
- **How handoffs work** → [SIGNAL_BASED_ROUTING_ANALYSIS.md §3.1](./SIGNAL_BASED_ROUTING_ANALYSIS.md#31-signal-based-routing-flow)
- **Context preservation** → [SIGNAL_BASED_ROUTING_ANALYSIS.md §4](./SIGNAL_BASED_ROUTING_ANALYSIS.md#4-context-preservation-through-handoffs)

### Debugging
- **Signal not detected** → [SIGNAL_ROUTING_GUIDE.md §4.2](./SIGNAL_ROUTING_GUIDE.md#42-common-issues--solutions)
- **Debug checklist** → [SIGNAL_ROUTING_QUICK_REF.md](./SIGNAL_ROUTING_QUICK_REF.md) (Debugging Checklist)
- **Common mistakes** → [SIGNAL_ROUTING_FAQ.md](./SIGNAL_ROUTING_FAQ.md#common-mistakes)

### Use Cases
- **Sequential pipeline** → [SIGNAL_ROUTING_GUIDE.md §3.1](./SIGNAL_ROUTING_GUIDE.md#31-sequential-pipeline)
- **Decision tree** → [SIGNAL_ROUTING_GUIDE.md §3.2](./SIGNAL_ROUTING_GUIDE.md#32-decision-tree)
- **Parallel processing** → [SIGNAL_ROUTING_GUIDE.md §3.3](./SIGNAL_ROUTING_GUIDE.md#33-parallel-processing-with-aggregation)
- **Error handling** → [SIGNAL_ROUTING_GUIDE.md §3.4](./SIGNAL_ROUTING_GUIDE.md#34-error-handling--recovery)

### Parallel Execution
- **How parallel groups work** → [SIGNAL_BASED_ROUTING_ANALYSIS.md §6](./SIGNAL_BASED_ROUTING_ANALYSIS.md#6-parallel-group-execution)
- **Parallel diagram** → [SIGNAL_ROUTING_DIAGRAM.md §9](./SIGNAL_ROUTING_DIAGRAM.md#9-parallel-group-execution-model)
- **FAQ on parallel** → [SIGNAL_ROUTING_FAQ.md](./SIGNAL_ROUTING_FAQ.md#parallel-execution)

### Performance
- **Performance impact** → [SIGNAL_BASED_ROUTING_ANALYSIS.md §9](./SIGNAL_BASED_ROUTING_ANALYSIS.md#9-performance-analysis)
- **Performance diagram** → [SIGNAL_ROUTING_DIAGRAM.md §7](./SIGNAL_ROUTING_DIAGRAM.md#7-performance-bottleneck-analysis)
- **Optimization tips** → [SIGNAL_ROUTING_GUIDE.md §5](./SIGNAL_ROUTING_GUIDE.md#5-performance-optimization)

### Production
- **Pre-deployment checklist** → [SIGNAL_ROUTING_GUIDE.md §6](./SIGNAL_ROUTING_GUIDE.md#6-production-checklist)
- **Monitoring** → [SIGNAL_ROUTING_GUIDE.md §6.2](./SIGNAL_ROUTING_GUIDE.md#62-monitoring)

---

## 📝 Document Cheat Sheet

### SIGNAL_BASED_ROUTING_ANALYSIS.md
- **Type:** Technical deep-dive
- **Sections:** 15
- **Reading time:** 45 minutes
- **Best for:** Understanding algorithm details
- **Covers:** Everything in depth

### SIGNAL_ROUTING_DIAGRAM.md
- **Type:** Visual reference
- **Sections:** 9 (all flowcharts/diagrams)
- **Reading time:** 20 minutes
- **Best for:** Visual learners
- **Covers:** Flowcharts, state machines, examples

### SIGNAL_ROUTING_GUIDE.md
- **Type:** Implementation guide
- **Sections:** 8
- **Reading time:** 60 minutes
- **Best for:** Hands-on implementation
- **Covers:** How-to, checklists, examples

### SIGNAL_ROUTING_FAQ.md
- **Type:** Q&A reference
- **Sections:** Core concepts + 8 topics
- **Reading time:** 40 minutes (or use as lookup)
- **Best for:** Quick answers to specific questions
- **Covers:** 30+ common questions

### SIGNAL_ROUTING_QUICK_REF.md
- **Type:** One-page reference
- **Sections:** 12 mini-sections
- **Reading time:** 5-10 minutes
- **Best for:** Quick lookup while coding
- **Covers:** Most essential information

---

## 🔗 Source Code References

| Concept | File | Lines |
|---------|------|-------|
| Signal matching | `core/crew_routing.go` | 49-90 |
| Termination check | `core/crew_routing.go` | 98-124 |
| Routing check | `core/crew_routing.go` | 126-157 |
| Parallel detection | `core/crew_routing.go` | 202-223 |
| Execution loop | `core/crew.go` | 656-873 |
| History management | `core/crew.go` | 525-615 |
| Handoff counting | `core/crew.go` | 654, 791, 861 |
| Type definitions | `core/config.go` | 14-44 |

---

## 💡 Tips & Tricks

### For Quick Answers
- Use [SIGNAL_ROUTING_FAQ.md](./SIGNAL_ROUTING_FAQ.md) for Q&A
- Use [SIGNAL_ROUTING_QUICK_REF.md](./SIGNAL_ROUTING_QUICK_REF.md) for one-page reference

### For Visual Understanding
- Start with [SIGNAL_ROUTING_DIAGRAM.md](./SIGNAL_ROUTING_DIAGRAM.md) diagrams
- Match them against [SIGNAL_BASED_ROUTING_ANALYSIS.md](./SIGNAL_BASED_ROUTING_ANALYSIS.md) descriptions

### For Implementation
- Use [SIGNAL_ROUTING_GUIDE.md](./SIGNAL_ROUTING_GUIDE.md) implementation checklist
- Reference [examples/01-quiz-exam/config/crew.yaml](./examples/01-quiz-exam/config/crew.yaml)
- Follow debugging section if issues arise

### For Production
- Follow [SIGNAL_ROUTING_GUIDE.md §6](./SIGNAL_ROUTING_GUIDE.md#6-production-checklist)
- Keep [SIGNAL_ROUTING_QUICK_REF.md](./SIGNAL_ROUTING_QUICK_REF.md) handy for monitoring

---

## 🎓 Learning Summary

This documentation suite provides:

✅ **Complete Coverage** - From basics to advanced topics
✅ **Multiple Formats** - Text, diagrams, Q&A, checklists
✅ **Multiple Levels** - Beginner to advanced
✅ **Practical Examples** - Real-world use cases
✅ **Source Code** - Links to actual implementation
✅ **Quick Reference** - For lookup while coding

---

## 📞 Support Resources

| Need | Resource |
|------|----------|
| **Quick answer** | [SIGNAL_ROUTING_FAQ.md](./SIGNAL_ROUTING_FAQ.md) |
| **Visual explanation** | [SIGNAL_ROUTING_DIAGRAM.md](./SIGNAL_ROUTING_DIAGRAM.md) |
| **How to implement** | [SIGNAL_ROUTING_GUIDE.md](./SIGNAL_ROUTING_GUIDE.md) |
| **Technical details** | [SIGNAL_BASED_ROUTING_ANALYSIS.md](./SIGNAL_BASED_ROUTING_ANALYSIS.md) |
| **Quick lookup** | [SIGNAL_ROUTING_QUICK_REF.md](./SIGNAL_ROUTING_QUICK_REF.md) |
| **Source code** | [core/crew_routing.go](./core/crew_routing.go) |

---

## ✨ Document Quality

- ✅ First Principles: Each document explains WHY, not just WHAT
- ✅ Code References: Every concept linked to source code
- ✅ Examples: Multiple real-world examples for each pattern
- ✅ Completeness: No critical information missing
- ✅ Consistency: Terminology and examples consistent across documents
- ✅ Navigability: Easy cross-references between documents

---

**Last Updated:** 2025-12-23
**Coverage:** Signal-based routing in core/crew.go
**Status:** Complete ✅


# Removal Plan - Code Cleanup

## Status: STRATEGY UPDATED ✅

Created: 2025-12-25
Verified: YES
Risk Level: VERY LOW

---

## STRATEGY CHANGE: Remove Lowercase Duplicates Instead

After review: core/agent/execution.go (package `agent`) and core/agent_execution.go (package `crewai`) are in **different packages**.

**New Strategy:** Keep uppercase versions in `core/agent/` (official public API), remove lowercase duplicate versions from `core/agent_execution.go`.

---

## Functions to Remove

### Phase 1: Remove Lowercase Duplicates from core/agent_execution.go

These 4 functions are exact duplicates of the upstream versions in core/agent/execution.go:

1. **convertToProviderMessages** (lines 354-363)
   - Duplicate of: `ConvertToProviderMessages` in core/agent/execution.go:165
   - Usage in file: 1 call at line 239
   - Action: Remove and use `agent.ConvertToProviderMessages()` instead

2. **convertToolsToProvider** (lines 386-404)
   - Duplicate of: `ConvertToolsToProvider` in core/agent/execution.go:177
   - Usage in file: 2 calls (lines 94, 294)
   - Action: Remove and use `agent.ConvertToolsToProvider()` instead

3. **convertToolCallsFromProvider** (lines 407-417)
   - Duplicate of: `ConvertToolCallsFromProvider` in core/agent/execution.go:189
   - Usage in file: 1 call at line 211
   - Action: Remove and use `agent.ConvertToolCallsFromProvider()` instead

4. **buildCustomPrompt** (lines 419-427)
   - Duplicate of: buildCustomPrompt in core/agent/execution.go:210
   - Usage in file: 1 call at line 489
   - Action: Remove and use the one from core/agent/execution.go via import

### Phase 2: Remove Shadowing Function

5. **EstimateTokens** in core/agent/execution.go (lines 242-246)
   - Shadows: `agent.EstimateTokens()` method in common/types.go
   - Usage: 2 calls (lines 99, 139)
   - Action: Remove and replace with `agent.EstimateTokens()` method

### Phase 3: Inline Single-Caller Function

1. **buildCompletionRequest** in core/agent_execution.go (lines 87-96)
   - Single caller: `executeWithModelConfig()` at line 187
   - Action: Inline the 10-line function into its only caller

---

## Implementation Checklist

- [ ] 1. Update call sites in core/agent_execution.go to use agent.ConvertToProviderMessages()
- [ ] 2. Update call sites in core/agent_execution.go to use agent.ConvertToolsToProvider()
- [ ] 3. Update call sites in core/agent_execution.go to use agent.ConvertToolCallsFromProvider()
- [ ] 4. Update call site in core/agent_execution.go for buildCustomPrompt
- [ ] 5. Remove convertToProviderMessages from core/agent_execution.go (9 lines)
- [ ] 6. Remove convertToolsToProvider from core/agent_execution.go (19 lines)
- [ ] 7. Remove convertToolCallsFromProvider from core/agent_execution.go (11 lines)
- [ ] 8. Remove buildCustomPrompt from core/agent_execution.go (9 lines)
- [ ] 9. Remove EstimateTokens from core/agent/execution.go (5 lines)
- [ ] 10. Remove buildCompletionRequest and inline into caller
- [ ] 11. Run full test suite
- [ ] 12. Verify no breakage

---

## Files to Modify

1. **core/agent_execution.go** (main file)
   - Remove: convertToProviderMessages (9 lines) - lines 354-363
   - Remove: convertToolsToProvider (19 lines) - lines 386-404
   - Remove: convertToolCallsFromProvider (11 lines) - lines 407-417
   - Remove: buildCustomPrompt (9 lines) - lines 419-427
   - Remove: buildCompletionRequest (10 lines) - lines 87-96 + inline
   - Update: 5 call sites to use agent.* versions
   - Total lines removed: ~58 lines

2. **core/agent/execution.go** (minor)
   - Remove: EstimateTokens (5 lines) - lines 242-246
   - Update: 2 call sites to use agent.EstimateTokens() method

3. **core/** (package-level)
   - Add: import of core/agent in agent_execution.go if needed

---

## Verification Done

- ✅ Checked for duplicate functions across files
- ✅ Verified shadowing (EstimateTokens)
- ✅ Identified single-caller functions
- ✅ Checked package imports (no external dependencies)
- ✅ Identified false claims (Unwrap, InvalidateSystemPromptCache)
- ✅ Calculated actual impact (~63 lines, not 150-180)

---

## Risk Assessment

**Overall Risk: VERY LOW**

✅ No external API breakage
✅ All functions are internal to core/agent package
✅ Direct lowercase versions available as replacements
✅ Only used within local files (no cross-package deps)
✅ Comprehensive test coverage for verification

---

## Next Steps

When ready to execute:
1. Create feature branch: `refactor/remove-duplicate-functions`
2. Execute removals according to checklist
3. Run: `go test ./...`
4. Create PR with this as justification
5. Merge after approval

**Estimated Time:** 30 minutes
**Complexity:** LOW

# Removal Plan - Code Cleanup

## Status: READY FOR IMPLEMENTATION ✅

Created: 2025-12-25
Verified: YES
Risk Level: VERY LOW

---

## Functions to Remove

### Phase 1: Remove Public Duplicate Functions from core/agent/execution.go

These 4 functions are exact duplicates of lowercase versions in core/agent_execution.go:

1. **ConvertToProviderMessages** (lines 165-175)
   - Replace calls with: `convertToProviderMessages()` from core/agent_execution.go:354
   - Update locations: lines 29, 66

2. **ConvertToolsToProvider** (lines 177-187)
   - Replace calls with: `convertToolsToProvider()` from core/agent_execution.go:386
   - Update locations: lines 107, 153

3. **ConvertToolCallsFromProvider** (lines 189-200)
   - Replace calls with: `convertToolCallsFromProvider()` from core/agent_execution.go:407
   - Update locations: line 128

4. **buildCustomPrompt** (lines 210-218)
   - Remove from core/agent/execution.go
   - Keep in core/agent_execution.go (lines 419-427)
   - Called from: line 205

### Phase 2: Remove Shadowing Function

5. **EstimateTokens** (lines 242-246)
   - Replace calls with: `agent.EstimateTokens()` method
   - Update locations: lines 99, 139
   - Rationale: Shadows Agent method in common/types.go

### Phase 3: Inline Single-Caller Function

6. **buildCompletionRequest** in core/agent_execution.go (lines 87-96)
   - Inline into: `executeWithModelConfig()` at line 187
   - Rationale: Single caller, simple transformation

---

## Tests to Remove from core/agent_test.go

Remove these 6 test functions:
- TestConvertToProviderMessages
- TestConvertToProviderMessagesEmpty
- TestConvertToolsToProvider
- TestConvertToolsToProviderEmpty
- TestConvertToolCallsFromProvider
- TestConvertToolCallsFromProviderEmpty

---

## Implementation Checklist

- [ ] 1. Update ConvertToProviderMessages call sites (2 locations)
- [ ] 2. Update ConvertToolsToProvider call sites (2 locations)
- [ ] 3. Update ConvertToolCallsFromProvider call sites (1 location)
- [ ] 4. Remove buildCustomPrompt from core/agent/execution.go
- [ ] 5. Update EstimateTokens call sites (2 locations)
- [ ] 6. Remove buildCompletionRequest and inline into caller
- [ ] 7. Remove test functions from core/agent_test.go (6 functions)
- [ ] 8. Run full test suite
- [ ] 9. Verify no breakage
- [ ] 10. Commit changes

---

## Files to Modify

1. **core/agent/execution.go**
   - Remove: ConvertToProviderMessages (11 lines)
   - Remove: ConvertToolsToProvider (11 lines)
   - Remove: ConvertToolCallsFromProvider (12 lines)
   - Remove: EstimateTokens (5 lines)
   - Remove: buildCustomPrompt (9 lines)
   - Update: Call sites for above functions

2. **core/agent_execution.go**
   - Remove: buildCompletionRequest (10 lines) and inline
   - Update: Caller site

3. **core/agent_test.go**
   - Remove: 6 test functions

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

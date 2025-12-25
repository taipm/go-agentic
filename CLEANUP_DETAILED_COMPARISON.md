# Detailed Comparison: Old vs New Config Loader

## File Structure Comparison

### Old Version (DELETED)
```
core/config_loader.go (538 lines)
├── Package: crewai
├── Imports:
│   ├── fmt, log, os, path/filepath, sync, time
│   ├── gopkg.in/yaml.v3
│   ├── github.com/taipm/go-agentic/core/common
│
├── Functions (12):
│   ├── LoadCrewConfig() - Load crew config YAML
│   ├── LoadAndValidateCrewConfig() - Load + validate with circular detection
│   ├── LoadAgentConfig(path, configMode) - Load agent config
│   ├── LoadAgentConfigs(dir, configMode) - Load all agents
│   ├── CreateAgentFromConfig() - Create agent from config
│   ├── convertToModelConfig() - Helper
│   ├── buildAgentMetadata() - Helper
│   ├── buildAgentQuotas() - Helper
│   ├── addAgentTools() - Helper
│   ├── getInputTokenPrice() - Helper
│   ├── getOutputTokenPrice() - Helper
│   └── ConfigToHardcodedDefaults() - Config conversion
│
└── Status: ❌ DELETED - Not used anywhere
```

### New Version (ACTIVE)
```
core/config/loader.go (309 lines)
├── Package: config
├── Imports:
│   ├── fmt, log, os, path/filepath, sync, time
│   ├── gopkg.in/yaml.v3
│   ├── github.com/taipm/go-agentic/core/common
│   ├── github.com/taipm/go-agentic/core/validation
│
├── Functions (9):
│   ├── LoadCrewConfig() - Load crew config YAML
│   ├── LoadAgentConfig() - Load agent config
│   ├── LoadAgentConfigs() - Load all agents
│   ├── ExpandEnvVars() - Utility function
│   ├── CreateAgentFromConfig() - Create agent from config
│   ├── convertToModelConfig() - Helper
│   ├── buildAgentMetadata() - Helper
│   ├── buildAgentQuotas() - Helper
│   └── addAgentTools() - Helper
│
└── Status: ✅ ACTIVE - Used in core/crew.go
```

---

## Function-by-Function Comparison

### 1. LoadCrewConfig()

**OLD (config_loader.go, lines 16-52)**
```go
func LoadCrewConfig(path string) (*CrewConfig, error) {
    data, err := os.ReadFile(path)
    if err != nil {
        return nil, fmt.Errorf("failed to read crew config: %w", err)
    }

    var config CrewConfig  // ← OLD TYPE (crewai package)
    err = yaml.Unmarshal(data, &config)
    if err != nil {
        return nil, fmt.Errorf("failed to parse crew config YAML: %w", err)
    }

    // Set defaults
    if config.Settings.MaxHandoffs == 0 {
        config.Settings.MaxHandoffs = 5
    }
    // ... more defaults

    // ✅ FIX for Issue #6: Validate configuration at load time
    if err := ValidateCrewConfig(&config); err != nil {  // ← OLD validator
        log.Printf("[CONFIG ERROR] Failed to validate crew config: %v", err)
        return nil, fmt.Errorf("invalid crew configuration: %w", err)
    }

    log.Printf("[CONFIG SUCCESS] Crew config loaded: version=%s, agents=%d, entry=%s",
        config.Version, len(config.Agents), config.EntryPoint)
    return &config, nil
}
```

**NEW (config/loader.go, lines 17-52)**
```go
func LoadCrewConfig(path string) (*common.CrewConfig, error) {
    data, err := os.ReadFile(path)
    if err != nil {
        return nil, fmt.Errorf("failed to read crew config: %w", err)
    }

    var config common.CrewConfig  // ← NEW TYPE (common package)
    err = yaml.Unmarshal(data, &config)
    if err != nil {
        return nil, fmt.Errorf("failed to parse crew config YAML: %w", err)
    }

    // Set defaults
    if config.Settings.MaxHandoffs == 0 {
        config.Settings.MaxHandoffs = 5
    }
    // ... more defaults

    // Validate crew configuration
    if err := validation.ValidateCrewConfig(&config); err != nil {  // ← NEW validator
        return nil, fmt.Errorf("crew config validation failed: %w", err)
    }

    log.Printf("[CONFIG SUCCESS] Crew config loaded: version=%s, agents=%d, entry=%s",
        config.Version, len(config.Agents), config.EntryPoint)
    return &config, nil
}
```

**Differences:**
| Aspect | Old | New |
|--------|-----|-----|
| Return type | `*CrewConfig` (crewai) | `*common.CrewConfig` |
| Validator | `ValidateCrewConfig()` (local) | `validation.ValidateCrewConfig()` |
| Log on error | Yes | No (validation error directly) |
| Code length | 37 lines | 36 lines |
| **Used?** | ❌ No | ✅ Yes (core/crew.go) |

---

### 2. LoadAgentConfig()

**OLD (config_loader.go, lines 83-174)**
```go
func LoadAgentConfig(path string, configMode ConfigMode) (*AgentConfig, error) {
    // ... read and unmarshal

    // ✅ FIX for Issue #5: Add configMode parameter for STRICT/PERMISSIVE mode validation
    if err := ValidateAgentConfig(&config, configMode); err != nil {
        return nil, fmt.Errorf("invalid agent configuration: %w", err)
    }

    return &config, nil
}
```

**NEW (config/loader.go, lines 54-138)**
```go
func LoadAgentConfig(path string) (*common.AgentConfig, error) {
    // ... read and unmarshal

    // Validate agent configuration
    if err := validation.ValidateAgentConfig(&config); err != nil {
        return nil, fmt.Errorf("agent config validation failed: %w", err)
    }

    return &config, nil
}
```

**Differences:**
| Aspect | Old | New |
|--------|-----|-----|
| Signature | `LoadAgentConfig(path, configMode)` | `LoadAgentConfig(path)` |
| Validator | `ValidateAgentConfig(config, mode)` | `validation.ValidateAgentConfig(config)` |
| Type returned | `*AgentConfig` (old) | `*common.AgentConfig` |
| Config defaults | 92 lines | 84 lines |
| **Used?** | ❌ No | ✅ Yes |

---

### 3. LoadAgentConfigs()

**OLD (config_loader.go, lines 176-201)**
```go
func LoadAgentConfigs(dir string, configMode ConfigMode) (map[string]*AgentConfig, error) {
    // ...
    config, err := LoadAgentConfig(filePath, configMode)  // ← passes mode
    // ...
}
```

**NEW (config/loader.go, lines 140-164)**
```go
func LoadAgentConfigs(dir string) (map[string]*common.AgentConfig, error) {
    // ...
    config, err := LoadAgentConfig(filePath)  // ← no mode param
    // ...
}
```

**Differences:**
| Aspect | Old | New |
|--------|-----|-----|
| Signature | `LoadAgentConfigs(dir, mode)` | `LoadAgentConfigs(dir)` |
| Return type | `map[string]*AgentConfig` | `map[string]*common.AgentConfig` |
| Sub-call | Passes mode | No mode param |
| **Used?** | ❌ No | ✅ Yes |

---

### 4. CreateAgentFromConfig()

**OLD (lines 308-340) vs NEW (lines 174-206)**
- Functionality: Identical ✅
- Package change: Uses `common.Agent` vs old `Agent`
- Order of operations: Same
- **Code length**: Same (~32 lines)
- **Used?** Old: ❌ | New: ✅

---

### 5. Helper Functions

#### convertToModelConfig()
| Version | Implementation | Used | Status |
|---------|---|---|---|
| Old (206-223) | Identical code | ❌ | DELETE |
| New (208-226) | Identical code | ✅ | KEEP |

#### buildAgentMetadata()
| Version | Implementation | Used | Status |
|---------|---|---|---|
| Old (244-296) | Identical code | ❌ | DELETE |
| New (228-280) | Identical code | ✅ | KEEP |

#### buildAgentQuotas()
| Version | Implementation | Used | Status |
|---------|---|---|---|
| Old (225-242) | Identical code | ❌ | DELETE |
| New (282-299) | Identical code | ✅ | KEEP |

#### addAgentTools()
| Version | Implementation | Used | Status |
|---------|---|---|---|
| Old (298-305) | Identical code | ❌ | DELETE |
| New (301-308) | Identical code | ✅ | KEEP |

---

### 6. Functions Only in Old Version

#### LoadAndValidateCrewConfig() (lines 54-81)
```go
func LoadAndValidateCrewConfig(crewConfigPath string, agentConfigs map[string]*AgentConfig) (*CrewConfig, error) {
    config, err := LoadCrewConfig(crewConfigPath)

    validator := NewConfigValidator(config, agentConfigs)  // ← OLD API
    if err := validator.ValidateAll(); err != nil {
        // ...
    }

    warnings := validator.GetWarnings()
    return config, nil
}
```

**Status:** ❌ DELETED - Uses old validator API
- No replacement in new version
- NewConfigValidator() không tồn tại

#### getInputTokenPrice() / getOutputTokenPrice() (lines 342-356)
```go
func getInputTokenPrice(costLimits *CostLimitsConfig) float64 {
    if costLimits != nil && costLimits.InputTokenPricePerMillion > 0 {
        return costLimits.InputTokenPricePerMillion
    }
    return 0.15  // Default: gpt-4o-mini pricing
}
```

**Status:** ❌ DELETED - Unused utility
- No callers
- Token pricing handled in CostLimitsConfig directly

#### ConfigToHardcodedDefaults() (lines 358-537)
```go
func ConfigToHardcodedDefaults(config *CrewConfig) *HardcodedDefaults {
    // 180 lines of conversion logic
    // STRICT/PERMISSIVE mode handling
    // Timeout configurations
    // ...
}
```

**Status:** ❌ DELETED - Unused
- Seems like Phase 5 development
- No callers found
- Logic replaced elsewhere (if still needed)

---

## Package Organization

### OLD (Scattered)
```
Package: crewai
  ├── crew.go
  ├── config_loader.go  ← Config loading mixed with main logic
  ├── defaults.go
  ├── types.go
  └── other files...
```

**Problem:** Config loading in same package as crew execution

### NEW (Organized)
```
Package: crewai (core domain logic)
  ├── crew.go
  ├── defaults.go
  └── types.go

Package: config (configuration)
  ├── loader.go  ← Config loading isolated
  └── types.go

Package: common (shared types)
  ├── types.go  ← CrewConfig, AgentConfig definitions
  └── errors.go
```

**Benefit:** Clear separation of concerns

---

## Import Impact

### Using OLD config_loader.go
```
❌ Not recommended
├── Would import from "github.com/taipm/go-agentic/core"
├── Gets CrewConfig from crewai package
├── Gets AgentConfig from crewai package
└── Mixes concerns
```

### Using NEW core/config/loader.go (CURRENT)
```
✅ CORRECT (Already in use)
├── Import from "github.com/taipm/go-agentic/core/config"
├── Gets CrewConfig from common package
├── Gets AgentConfig from common package
└── Clear separation of concerns
```

---

## Conclusion

| Aspect | Old | New |
|--------|-----|-----|
| **Lines** | 538 | 309 |
| **Functions** | 12 | 9 |
| **Type source** | crewai | common |
| **Validator** | Local | From validation pkg |
| **Used** | ❌ NO | ✅ YES |
| **Status** | ❌ DELETED | ✅ ACTIVE |
| **Organization** | Mixed | Separated |
| **Maintainability** | Low | High |

**Recommendation**: ✅ Keep only NEW version
- Better organized
- Cleaner imports
- No duplication
- Actually used in codebase

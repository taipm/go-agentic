package agentic

import (
	"fmt"
	"regexp"
	"strings"
)

// ============================================
// Trigger Detection System
// ============================================

// TriggerDetector defines an interface for detecting routing signals in agent responses
type TriggerDetector interface {
	// Detect returns true if the trigger condition is met for the given response
	Detect(response string) bool
	// Description returns a human-readable description of what this detector matches
	Description() string
}

// ============================================
// Built-in Trigger Detectors
// ============================================

// KeywordDetector matches if response contains any of the specified keywords
type KeywordDetector struct {
	Keywords    []string
	CaseSensitive bool
}

// NewKeywordDetector creates a detector that matches any of the given keywords
func NewKeywordDetector(keywords []string, caseSensitive bool) *KeywordDetector {
	return &KeywordDetector{
		Keywords:    keywords,
		CaseSensitive: caseSensitive,
	}
}

func (kd *KeywordDetector) Detect(response string) bool {
	searchText := response
	if !kd.CaseSensitive {
		searchText = strings.ToLower(response)
	}

	for _, keyword := range kd.Keywords {
		if !kd.CaseSensitive {
			keyword = strings.ToLower(keyword)
		}
		if strings.Contains(searchText, keyword) {
			return true
		}
	}
	return false
}

func (kd *KeywordDetector) Description() string {
	return fmt.Sprintf("Contains keywords: %v", kd.Keywords)
}

// ============================================

// PatternDetector matches if response matches a regex pattern
type PatternDetector struct {
	Pattern *regexp.Regexp
	rawPattern string
}

// NewPatternDetector creates a detector that matches a regex pattern
func NewPatternDetector(pattern string) (*PatternDetector, error) {
	re, err := regexp.Compile(pattern)
	if err != nil {
		return nil, fmt.Errorf("invalid regex pattern: %w", err)
	}

	return &PatternDetector{
		Pattern: re,
		rawPattern: pattern,
	}, nil
}

func (pd *PatternDetector) Detect(response string) bool {
	return pd.Pattern.MatchString(response)
}

func (pd *PatternDetector) Description() string {
	return fmt.Sprintf("Matches pattern: %s", pd.rawPattern)
}

// ============================================

// SignalDetector matches explicit signal format: [SIGNAL: signal_name]
// Example: "Issue resolved. [SIGNAL: resolved]"
type SignalDetector struct {
	SignalName string
}

// NewSignalDetector creates a detector that matches explicit signals
func NewSignalDetector(signalName string) *SignalDetector {
	return &SignalDetector{
		SignalName: signalName,
	}
}

func (sd *SignalDetector) Detect(response string) bool {
	// Look for [SIGNAL: signal_name] format
	pattern := fmt.Sprintf(`\[SIGNAL:\s*%s\s*\]`, regexp.QuoteMeta(sd.SignalName))
	re, err := regexp.Compile(pattern)
	if err != nil {
		return false
	}
	return re.MatchString(response)
}

func (sd *SignalDetector) Description() string {
	return fmt.Sprintf("Explicit signal: [SIGNAL: %s]", sd.SignalName)
}

// ============================================

// PrefixDetector matches if response starts with specified prefixes
type PrefixDetector struct {
	Prefixes      []string
	CaseSensitive bool
}

// NewPrefixDetector creates a detector that matches line-based prefixes
func NewPrefixDetector(prefixes []string, caseSensitive bool) *PrefixDetector {
	return &PrefixDetector{
		Prefixes:      prefixes,
		CaseSensitive: caseSensitive,
	}
}

func (pd *PrefixDetector) Detect(response string) bool {
	lines := strings.Split(response, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		searchText := line
		if !pd.CaseSensitive {
			searchText = strings.ToLower(line)
		}

		for _, prefix := range pd.Prefixes {
			if !pd.CaseSensitive {
				prefix = strings.ToLower(prefix)
			}
			if strings.HasPrefix(searchText, prefix) {
				return true
			}
		}
	}
	return false
}

func (pd *PrefixDetector) Description() string {
	return fmt.Sprintf("Starts with prefixes: %v", pd.Prefixes)
}

// ============================================

// AnyDetector matches if ANY of the child detectors match (OR logic)
type AnyDetector struct {
	Detectors []TriggerDetector
}

// NewAnyDetector creates a detector that matches if any child detector matches
func NewAnyDetector(detectors ...TriggerDetector) *AnyDetector {
	return &AnyDetector{
		Detectors: detectors,
	}
}

func (ad *AnyDetector) Detect(response string) bool {
	for _, detector := range ad.Detectors {
		if detector.Detect(response) {
			return true
		}
	}
	return false
}

func (ad *AnyDetector) Description() string {
	descriptions := make([]string, len(ad.Detectors))
	for i, d := range ad.Detectors {
		descriptions[i] = d.Description()
	}
	return fmt.Sprintf("Any of: (%s)", strings.Join(descriptions, " OR "))
}

// ============================================

// AllDetector matches if ALL child detectors match (AND logic)
type AllDetector struct {
	Detectors []TriggerDetector
}

// NewAllDetector creates a detector that matches if all child detectors match
func NewAllDetector(detectors ...TriggerDetector) *AllDetector {
	return &AllDetector{
		Detectors: detectors,
	}
}

func (ad *AllDetector) Detect(response string) bool {
	for _, detector := range ad.Detectors {
		if !detector.Detect(response) {
			return false
		}
	}
	return true
}

func (ad *AllDetector) Description() string {
	descriptions := make([]string, len(ad.Detectors))
	for i, d := range ad.Detectors {
		descriptions[i] = d.Description()
	}
	return fmt.Sprintf("All of: (%s)", strings.Join(descriptions, " AND "))
}

// ============================================

// AlwaysDetector always matches (useful for default routes)
type AlwaysDetector struct{}

func NewAlwaysDetector() *AlwaysDetector {
	return &AlwaysDetector{}
}

func (ad *AlwaysDetector) Detect(response string) bool {
	return true
}

func (ad *AlwaysDetector) Description() string {
	return "Always matches (default route)"
}

// ============================================

// NeverDetector never matches (useful for disabled routes)
type NeverDetector struct{}

func NewNeverDetector() *NeverDetector {
	return &NeverDetector{}
}

func (nd *NeverDetector) Detect(response string) bool {
	return false
}

func (nd *NeverDetector) Description() string {
	return "Never matches (disabled)"
}

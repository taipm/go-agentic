package crewai

import (
	"fmt"
	"log"
	"sync"
)

// HistoryManager handles conversation history management with thread-safe operations
// Responsibilities:
// - Safely append/read messages with mutex protection
// - Estimate token counts for context window management
// - Trim history when exceeding context limits
// - Provide conversation inspection
type HistoryManager struct {
	history []Message
	mu      sync.RWMutex // Protect history from concurrent access
}

// NewHistoryManager creates a new history manager
func NewHistoryManager() *HistoryManager {
	return &HistoryManager{
		history: []Message{},
	}
}

// NewHistoryManagerWith creates a new history manager initialized with messages
func NewHistoryManagerWith(messages []Message) *HistoryManager {
	hm := NewHistoryManager()
	for _, msg := range messages {
		hm.Append(msg)
	}
	return hm
}

// Append safely appends a message to history with mutex protection
func (hm *HistoryManager) Append(msg Message) {
	hm.mu.Lock()
	defer hm.mu.Unlock()
	hm.history = append(hm.history, msg)
}

// AppendUser is a convenience method to append a user message
func (hm *HistoryManager) AppendUser(content string) {
	hm.Append(Message{
		Role:    RoleUser,
		Content: content,
	})
}

// AppendAssistant is a convenience method to append an assistant message
func (hm *HistoryManager) AppendAssistant(content string) {
	hm.Append(Message{
		Role:    RoleAssistant,
		Content: content,
	})
}

// AppendSystem is a convenience method to append a system message
func (hm *HistoryManager) AppendSystem(content string) {
	hm.Append(Message{
		Role:    RoleSystem,
		Content: content,
	})
}

// Copy returns a deep copy of history for safe reading
// Caller can safely read the returned copy without affecting concurrent writers
func (hm *HistoryManager) Copy() []Message {
	hm.mu.RLock()
	defer hm.mu.RUnlock()

	if len(hm.history) == 0 {
		return []Message{}
	}

	historyCopy := make([]Message, len(hm.history))
	copy(historyCopy, hm.history)
	return historyCopy
}

// GetAll returns a copy of the entire conversation history
// Safe for external access and debugging
func (hm *HistoryManager) GetAll() []Message {
	return hm.Copy()
}

// EstimateTokens estimates total tokens in conversation history
// Uses approximation: 1 token ≈ TokenDivisor characters (OpenAI convention)
// Formula: TokenBaseValue + (len(content) + TokenPaddingValue) / TokenDivisor
func (hm *HistoryManager) EstimateTokens() int {
	hm.mu.RLock()
	defer hm.mu.RUnlock()

	total := 0
	for _, msg := range hm.history {
		// Role overhead + content tokens
		total += TokenBaseValue + (len(msg.Content)+TokenPaddingValue)/TokenDivisor
	}
	return total
}

// Length returns the number of messages in history
func (hm *HistoryManager) Length() int {
	hm.mu.RLock()
	defer hm.mu.RUnlock()
	return len(hm.history)
}

// Clear clears all conversation history
func (hm *HistoryManager) Clear() {
	hm.mu.Lock()
	defer hm.mu.Unlock()
	hm.history = []Message{}
}

// TrimIfNeeded trims conversation history to fit within context window
// Strategy: Keep first + recent messages, remove oldest in middle when over limit
//
// Parameters:
// - maxContextWindow: Maximum tokens allowed in context
// - contextTrimPercent: Percentage to trim when over limit (e.g., 20 = remove 20%)
// - defaults: HardcodedDefaults containing max context window settings
//
// Returns true if trimming occurred, false otherwise
func (hm *HistoryManager) TrimIfNeeded(maxContextWindow int, contextTrimPercent float64) bool {
	hm.mu.Lock()
	defer hm.mu.Unlock()

	if len(hm.history) <= MinHistoryLength {
		return false
	}

	// Calculate tokens directly
	currentTokens := 0
	for _, msg := range hm.history {
		currentTokens += TokenBaseValue + (len(msg.Content)+TokenPaddingValue)/TokenDivisor
	}

	// Check if within limit
	if currentTokens <= maxContextWindow {
		return false
	}

	// Calculate target after trimming (remove contextTrimPercent of max)
	trimPercent := contextTrimPercent / PercentDivisor
	targetTokens := int(float64(maxContextWindow) * (1.0 - trimPercent))

	// Calculate how many messages to keep from end
	keepFromEnd := len(hm.history) - 1
	tokensFromEnd := 0

	for i := len(hm.history) - 1; i > 0 && tokensFromEnd < targetTokens; i-- {
		msgTokens := TokenBaseValue + (len(hm.history[i].Content)+TokenPaddingValue)/TokenDivisor
		tokensFromEnd += msgTokens
		keepFromEnd = len(hm.history) - i
	}

	// Ensure we keep at least MinHistoryLength messages from end
	if keepFromEnd < MinHistoryLength {
		keepFromEnd = MinHistoryLength
	}

	// Build trimmed history: first message + summary + last N messages
	if len(hm.history) > keepFromEnd+1 {
		trimmedCount := len(hm.history) - keepFromEnd - 1

		newHistory := make([]Message, 0, keepFromEnd+2)

		// Keep first message
		newHistory = append(newHistory, hm.history[0])

		// Add summary for trimmed content
		newHistory = append(newHistory, Message{
			Role:    RoleSystem,
			Content: fmt.Sprintf("[%d earlier messages trimmed to fit context window]", trimmedCount),
		})

		// Keep last N messages
		startIdx := len(hm.history) - keepFromEnd
		newHistory = append(newHistory, hm.history[startIdx:]...)

		newTokens := 0
		for _, msg := range newHistory {
			newTokens += TokenBaseValue + (len(msg.Content)+TokenPaddingValue)/TokenDivisor
		}

		log.Printf("[CONTEXT TRIM] %d→%d messages, ~%d→%d tokens (saved ~%d tokens)",
			len(hm.history), len(newHistory), currentTokens, newTokens, currentTokens-newTokens)

		hm.history = newHistory
		return true
	}

	return false
}

// TrimIfNeededWithDefaults is a convenience wrapper that uses HardcodedDefaults
func (hm *HistoryManager) TrimIfNeededWithDefaults(defaults *HardcodedDefaults) bool {
	if defaults == nil {
		return false
	}
	return hm.TrimIfNeeded(defaults.MaxContextWindow, defaults.ContextTrimPercent)
}

// Statistics returns a summary of history statistics
type HistoryStats struct {
	MessageCount  int
	TotalTokens   int
	AverageLength int
	FirstMessage  *Message
	LastMessage   *Message
}

// GetStatistics returns statistical information about the history
func (hm *HistoryManager) GetStatistics() HistoryStats {
	hm.mu.RLock()
	defer hm.mu.RUnlock()

	stats := HistoryStats{
		MessageCount: len(hm.history),
		TotalTokens:  0,
	}

	if len(hm.history) == 0 {
		return stats
	}

	totalLen := 0
	for _, msg := range hm.history {
		tokens := TokenBaseValue + (len(msg.Content)+TokenPaddingValue)/TokenDivisor
		stats.TotalTokens += tokens
		totalLen += len(msg.Content)
	}

	stats.AverageLength = totalLen / len(hm.history)
	stats.FirstMessage = &hm.history[0]
	stats.LastMessage = &hm.history[len(hm.history)-1]

	return stats
}

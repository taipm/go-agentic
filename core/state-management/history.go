// Package statemanagement provides state tracking for workflow execution.
package statemanagement

import (
	"sync"

	"github.com/taipm/go-agentic/core/common"
)

// HistoryManager manages conversation history with thread-safe operations.
// It handles message storage, trimming, and retrieval for agent execution context.
type HistoryManager struct {
	messages        []common.Message
	maxSize         int
	trimThreshold   int
	trimPercentage  float64
	mu              sync.RWMutex
}

// NewHistoryManager creates a new HistoryManager with default settings.
func NewHistoryManager() *HistoryManager {
	return &HistoryManager{
		messages:       make([]common.Message, 0),
		maxSize:        10000,            // Maximum history size in characters
		trimThreshold:  8000,             // Trigger trim at this threshold
		trimPercentage: 0.25,             // Remove oldest 25% when trimming
	}
}

// NewHistoryManagerWithConfig creates a HistoryManager with custom settings.
func NewHistoryManagerWithConfig(maxSize, trimThreshold int, trimPercentage float64) *HistoryManager {
	return &HistoryManager{
		messages:       make([]common.Message, 0),
		maxSize:        maxSize,
		trimThreshold:  trimThreshold,
		trimPercentage: trimPercentage,
	}
}

// Add adds a message to the history and trims if necessary.
func (hm *HistoryManager) Add(msg common.Message) {
	hm.mu.Lock()
	defer hm.mu.Unlock()

	hm.messages = append(hm.messages, msg)
	hm.trimIfNeededLocked()
}

// AddMessages adds multiple messages to history at once.
func (hm *HistoryManager) AddMessages(msgs []common.Message) {
	hm.mu.Lock()
	defer hm.mu.Unlock()

	hm.messages = append(hm.messages, msgs...)
	hm.trimIfNeededLocked()
}

// GetMessages returns a copy of all messages in the history.
func (hm *HistoryManager) GetMessages() []common.Message {
	hm.mu.RLock()
	defer hm.mu.RUnlock()

	// Return a copy to prevent external modification
	result := make([]common.Message, len(hm.messages))
	copy(result, hm.messages)
	return result
}

// GetRecentMessages returns the last N messages from history.
func (hm *HistoryManager) GetRecentMessages(count int) []common.Message {
	hm.mu.RLock()
	defer hm.mu.RUnlock()

	if count <= 0 || len(hm.messages) == 0 {
		return []common.Message{}
	}

	start := len(hm.messages) - count
	if start < 0 {
		start = 0
	}

	result := make([]common.Message, len(hm.messages)-start)
	copy(result, hm.messages[start:])
	return result
}

// Clear removes all messages from history.
func (hm *HistoryManager) Clear() {
	hm.mu.Lock()
	defer hm.mu.Unlock()

	hm.messages = make([]common.Message, 0)
}

// Length returns the number of messages in history.
func (hm *HistoryManager) Length() int {
	hm.mu.RLock()
	defer hm.mu.RUnlock()

	return len(hm.messages)
}

// TotalSize returns the total character count of all messages.
func (hm *HistoryManager) TotalSize() int {
	hm.mu.RLock()
	defer hm.mu.RUnlock()

	totalSize := 0
	for _, msg := range hm.messages {
		totalSize += len(msg.Content)
	}
	return totalSize
}

// trimIfNeededLocked checks if history needs trimming and trims if necessary.
// MUST be called with lock held.
func (hm *HistoryManager) trimIfNeededLocked() {
	size := 0
	for _, msg := range hm.messages {
		size += len(msg.Content)
	}

	if size > hm.trimThreshold {
		// Calculate how many messages to remove
		removeCount := int(float64(len(hm.messages)) * hm.trimPercentage)
		if removeCount < 1 {
			removeCount = 1
		}

		// Keep newer messages, remove older ones
		hm.messages = hm.messages[removeCount:]
	}
}

// Copy creates a deep copy of the history for independent execution contexts.
func (hm *HistoryManager) Copy() *HistoryManager {
	hm.mu.RLock()
	defer hm.mu.RUnlock()

	newHM := NewHistoryManagerWithConfig(hm.maxSize, hm.trimThreshold, hm.trimPercentage)
	newHM.messages = make([]common.Message, len(hm.messages))
	copy(newHM.messages, hm.messages)
	return newHM
}

// SetTrimConfig updates trim threshold and percentage.
func (hm *HistoryManager) SetTrimConfig(threshold int, percentage float64) {
	hm.mu.Lock()
	defer hm.mu.Unlock()

	if threshold > 0 {
		hm.trimThreshold = threshold
	}
	if percentage > 0 && percentage < 1 {
		hm.trimPercentage = percentage
	}
}

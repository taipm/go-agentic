package crewai

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"sync/atomic"
	"syscall"
	"time"
)

// ✅ FIX for Issue #18 (Graceful Shutdown)
// ✅ Phase 5: Integration - Uses HardcodedDefaults for configurable shutdown behavior
// GracefulShutdownManager manages server lifecycle and safe shutdown
// Ensures all active requests complete before stopping
// Prevents data loss and resource leaks during shutdown
type GracefulShutdownManager struct {
	// Server components
	server   *http.Server
	serverMu sync.RWMutex

	// Request tracking (atomic for thread-safety)
	activeRequests int32 // Atomic counter

	// Stream tracking
	activeStreams map[string]context.CancelFunc
	streamMu      sync.RWMutex

	// Shutdown signaling
	shutdownChan   chan os.Signal
	isShuttingDown int32 // Atomic flag (0 = running, 1 = shutting down)

	// Configuration
	GracefulTimeout time.Duration // Default 30s
	defaults        *HardcodedDefaults // ✅ Phase 5: Runtime configuration defaults
	logger          *log.Logger

	// Shutdown callback for custom cleanup
	ShutdownCallback func() error
}

// NewGracefulShutdownManager creates a new shutdown manager
// ✅ Phase 5: Initializes with HardcodedDefaults for shutdown configuration
func NewGracefulShutdownManager() *GracefulShutdownManager {
	return &GracefulShutdownManager{
		activeStreams:   make(map[string]context.CancelFunc),
		shutdownChan:    make(chan os.Signal, 1),
		GracefulTimeout: 30 * time.Second,
		defaults:        DefaultHardcodedDefaults(), // ✅ Phase 5: Initialize with default values
		logger:          log.New(os.Stdout, "[SHUTDOWN] ", log.LstdFlags),
	}
}

// Start begins monitoring for shutdown signals (SIGTERM, SIGINT)
// This should be called in a goroutine
func (gsm *GracefulShutdownManager) Start() {
	// Register signal handlers
	signal.Notify(gsm.shutdownChan, syscall.SIGTERM, syscall.SIGINT)

	// Wait for signal
	sig := <-gsm.shutdownChan
	gsm.logger.Printf("Received signal: %v", sig)

	// Start graceful shutdown
	if err := gsm.Shutdown(context.Background()); err != nil {
		gsm.logger.Printf("Shutdown error: %v", err)
		os.Exit(1)
	}

	os.Exit(0)
}

// Shutdown gracefully stops the server
// 1. Marks as shutting down (reject new requests)
// 2. Cancels all active streams
// 3. Waits for request completion (with timeout)
// 4. Cleans up resources
func (gsm *GracefulShutdownManager) Shutdown(ctx context.Context) error {
	// Mark as shutting down
	atomic.StoreInt32(&gsm.isShuttingDown, 1)
	gsm.logger.Println("Starting graceful shutdown...")

	// Cancel all active streams
	gsm.logger.Println("Cancelling active streams...")
	gsm.cancelAllStreams()

	// Create shutdown context with timeout
	shutdownCtx, cancel := context.WithTimeout(context.Background(), gsm.GracefulTimeout)
	defer cancel()

	// Wait for active requests to complete or timeout
	waitStart := time.Now()
	for {
		activeCount := atomic.LoadInt32(&gsm.activeRequests)
		if activeCount == 0 {
			gsm.logger.Printf("All requests completed in %v", time.Since(waitStart))
			break
		}

		select {
		case <-shutdownCtx.Done():
			gsm.logger.Printf("Shutdown timeout: %d requests still active", activeCount)
			// Force close server after timeout
			if gsm.server != nil {
				gsm.server.Close()
			}
			return fmt.Errorf("shutdown timeout: %d active requests", activeCount)

		case <-time.After(gsm.defaults.GracefulShutdownCheckInterval): // ✅ Phase 5: Configurable check interval
			// Check again after brief wait
			activeCount := atomic.LoadInt32(&gsm.activeRequests)
			if activeCount > 0 {
				gsm.logger.Printf("Waiting for %d active requests to complete...", activeCount)
			}
		}
	}

	// Call custom shutdown callback if provided
	if gsm.ShutdownCallback != nil {
		gsm.logger.Println("Running custom shutdown callback...")
		if err := gsm.ShutdownCallback(); err != nil {
			gsm.logger.Printf("Callback error: %v", err)
		}
	}

	// Shutdown HTTP server
	if gsm.server != nil {
		gsm.logger.Println("Shutting down HTTP server...")
		if err := gsm.server.Shutdown(shutdownCtx); err != nil {
			gsm.logger.Printf("HTTP server shutdown error: %v", err)
		}
	}

	gsm.logger.Println("Graceful shutdown complete")
	return nil
}

// IsShuttingDown returns true if shutdown is in progress
func (gsm *GracefulShutdownManager) IsShuttingDown() bool {
	return atomic.LoadInt32(&gsm.isShuttingDown) == 1
}

// IncrementActiveRequests increments the active request counter
// Call this at the start of request handling
func (gsm *GracefulShutdownManager) IncrementActiveRequests() {
	if !gsm.IsShuttingDown() {
		atomic.AddInt32(&gsm.activeRequests, 1)
	}
}

// DecrementActiveRequests decrements the active request counter
// Call this at the end of request handling (use defer)
func (gsm *GracefulShutdownManager) DecrementActiveRequests() {
	atomic.AddInt32(&gsm.activeRequests, -1)
}

// GetActiveRequests returns the current count of active requests
func (gsm *GracefulShutdownManager) GetActiveRequests() int {
	return int(atomic.LoadInt32(&gsm.activeRequests))
}

// RegisterStream registers an active stream for shutdown tracking
// This allows graceful cancellation during shutdown
func (gsm *GracefulShutdownManager) RegisterStream(streamID string, cancelFunc context.CancelFunc) {
	gsm.streamMu.Lock()
	defer gsm.streamMu.Unlock()

	gsm.activeStreams[streamID] = cancelFunc
}

// UnregisterStream removes a completed stream
func (gsm *GracefulShutdownManager) UnregisterStream(streamID string) {
	gsm.streamMu.Lock()
	defer gsm.streamMu.Unlock()

	delete(gsm.activeStreams, streamID)
}

// GetActiveStreamCount returns the number of active streams
func (gsm *GracefulShutdownManager) GetActiveStreamCount() int {
	gsm.streamMu.RLock()
	defer gsm.streamMu.RUnlock()

	return len(gsm.activeStreams)
}

// cancelAllStreams cancels all active streams
// Used during shutdown to stop long-running operations
func (gsm *GracefulShutdownManager) cancelAllStreams() {
	gsm.streamMu.Lock()
	defer gsm.streamMu.Unlock()

	count := len(gsm.activeStreams)
	if count > 0 {
		gsm.logger.Printf("Cancelling %d active streams...", count)
	}

	for streamID, cancelFunc := range gsm.activeStreams {
		if cancelFunc != nil {
			gsm.logger.Printf("Cancelling stream: %s", streamID)
			cancelFunc()
		}
	}

	// Clear the map after cancelling
	gsm.activeStreams = make(map[string]context.CancelFunc)
}

// WaitForShutdown blocks until shutdown signal is received
// Useful for testing
func (gsm *GracefulShutdownManager) WaitForShutdown() {
	sig := <-gsm.shutdownChan
	gsm.logger.Printf("Received shutdown signal: %v", sig)
}

// ForceShutdown immediately shuts down the server
// Should only be used in testing or emergency scenarios
func (gsm *GracefulShutdownManager) ForceShutdown() {
	atomic.StoreInt32(&gsm.isShuttingDown, 1)

	// Cancel all streams
	gsm.cancelAllStreams()

	// Force close server
	if gsm.server != nil {
		gsm.server.Close()
	}
}

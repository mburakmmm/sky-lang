package runtime

import (
	"runtime"
	"sync/atomic"
	"time"
)

// GCOptimizer manages GC optimization strategies
type GCOptimizer struct {
	enabled           bool
	targetPauseMs     int64
	actualPauseMs     int64
	collections       int64
	lastCollectionMs  int64
	adaptiveThreshold int64
}

// NewGCOptimizer creates a new GC optimizer
func NewGCOptimizer(targetPauseMs int64) *GCOptimizer {
	return &GCOptimizer{
		enabled:           true,
		targetPauseMs:     targetPauseMs,
		adaptiveThreshold: 1024 * 1024 * 10, // 10MB default
	}
}

// OptimizeGC applies GC optimizations
func (gco *GCOptimizer) OptimizeGC() {
	if !gco.enabled {
		return
	}

	// Measure GC pause time
	start := time.Now()
	runtime.GC()
	pauseMs := time.Since(start).Milliseconds()

	atomic.AddInt64(&gco.collections, 1)
	atomic.StoreInt64(&gco.lastCollectionMs, pauseMs)

	// Update running average
	atomic.StoreInt64(&gco.actualPauseMs,
		(gco.actualPauseMs*9+pauseMs)/10)

	// Adapt threshold based on pause time
	if gco.actualPauseMs > gco.targetPauseMs {
		// Increase threshold to reduce frequency
		newThreshold := gco.adaptiveThreshold * 110 / 100
		atomic.StoreInt64(&gco.adaptiveThreshold, newThreshold)
	} else {
		// Decrease threshold for more frequent collections
		newThreshold := gco.adaptiveThreshold * 95 / 100
		atomic.StoreInt64(&gco.adaptiveThreshold, newThreshold)
	}
}

// GetStats returns GC statistics
func (gco *GCOptimizer) GetStats() map[string]int64 {
	return map[string]int64{
		"collections":     atomic.LoadInt64(&gco.collections),
		"last_pause_ms":   atomic.LoadInt64(&gco.lastCollectionMs),
		"avg_pause_ms":    atomic.LoadInt64(&gco.actualPauseMs),
		"target_pause_ms": gco.targetPauseMs,
		"threshold_bytes": atomic.LoadInt64(&gco.adaptiveThreshold),
	}
}

// SetTargetPause sets the target pause time
func (gco *GCOptimizer) SetTargetPause(ms int64) {
	gco.targetPauseMs = ms
}

// Enable enables the optimizer
func (gco *GCOptimizer) Enable() {
	gco.enabled = true
}

// Disable disables the optimizer
func (gco *GCOptimizer) Disable() {
	gco.enabled = false
}

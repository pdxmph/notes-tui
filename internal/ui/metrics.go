package ui

import (
	"fmt"
	"sync"
	"time"
)

// Metrics tracks performance metrics for the UI
type Metrics struct {
	mu sync.RWMutex
	
	// Frame metrics
	LastFrameTime   time.Duration
	AvgFrameTime    time.Duration
	frameCount      int64
	totalFrameTime  time.Duration
	
	// Component metrics
	LastRenderTime  map[string]time.Duration
	AvgRenderTime   map[string]time.Duration
	componentCounts map[string]int64
	componentTotals map[string]time.Duration
	
	// Memory metrics
	AllocCount      int
	StringAllocSize int64
	
	// State sync metrics
	StateSyncCount  int64
	LastSyncTime    time.Duration
	AvgSyncTime     time.Duration
	totalSyncTime   time.Duration
}

// NewMetrics creates a new metrics tracker
func NewMetrics() *Metrics {
	return &Metrics{
		LastRenderTime:  make(map[string]time.Duration),
		AvgRenderTime:   make(map[string]time.Duration),
		componentCounts: make(map[string]int64),
		componentTotals: make(map[string]time.Duration),
	}
}

// RecordFrame records a frame render time
func (m *Metrics) RecordFrame(duration time.Duration) {
	m.mu.Lock()
	defer m.mu.Unlock()
	
	m.LastFrameTime = duration
	m.frameCount++
	m.totalFrameTime += duration
	m.AvgFrameTime = m.totalFrameTime / time.Duration(m.frameCount)
}

// RecordComponent records a component render time
func (m *Metrics) RecordComponent(name string, duration time.Duration) {
	m.mu.Lock()
	defer m.mu.Unlock()
	
	m.LastRenderTime[name] = duration
	m.componentCounts[name]++
	m.componentTotals[name] += duration
	m.AvgRenderTime[name] = m.componentTotals[name] / time.Duration(m.componentCounts[name])
}

// RecordStateSync records a state sync operation
func (m *Metrics) RecordStateSync(duration time.Duration) {
	m.mu.Lock()
	defer m.mu.Unlock()
	
	m.LastSyncTime = duration
	m.StateSyncCount++
	m.totalSyncTime += duration
	m.AvgSyncTime = m.totalSyncTime / time.Duration(m.StateSyncCount)
}

// RecordAllocation records memory allocation
func (m *Metrics) RecordAllocation(count int, size int64) {
	m.mu.Lock()
	defer m.mu.Unlock()
	
	m.AllocCount += count
	m.StringAllocSize += size
}

// Reset resets all metrics
func (m *Metrics) Reset() {
	m.mu.Lock()
	defer m.mu.Unlock()
	
	m.frameCount = 0
	m.totalFrameTime = 0
	m.AvgFrameTime = 0
	m.LastFrameTime = 0
	
	m.StateSyncCount = 0
	m.totalSyncTime = 0
	m.AvgSyncTime = 0
	m.LastSyncTime = 0
	
	m.AllocCount = 0
	m.StringAllocSize = 0
	
	// Clear maps
	m.LastRenderTime = make(map[string]time.Duration)
	m.AvgRenderTime = make(map[string]time.Duration)
	m.componentCounts = make(map[string]int64)
	m.componentTotals = make(map[string]time.Duration)
}

// Summary returns a formatted summary of metrics
func (m *Metrics) Summary() string {
	m.mu.RLock()
	defer m.mu.RUnlock()
	
	return fmt.Sprintf(
		"Frame: %v avg (%v last) | Sync: %v avg (%d calls) | Allocs: %d (%s)",
		m.AvgFrameTime.Round(time.Microsecond),
		m.LastFrameTime.Round(time.Microsecond),
		m.AvgSyncTime.Round(time.Microsecond),
		m.StateSyncCount,
		m.AllocCount,
		formatBytes(m.StringAllocSize),
	)
}

// formatBytes formats bytes in human readable format
func formatBytes(b int64) string {
	const unit = 1024
	if b < unit {
		return fmt.Sprintf("%d B", b)
	}
	div, exp := int64(unit), 0
	for n := b / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %ciB", float64(b)/float64(div), "KMGTPE"[exp])
}
// +build nogpu

package monitor

import (
	"log"
	"sync"

	"gpu-pro/analytics"
)

// GPUMonitor is a stub monitor for minimal builds (no GPU support)
type GPUMonitor struct {
	initialized     bool
	gpuData         map[string]interface{}
	mu              sync.RWMutex
	heartbeatClient *analytics.HeartbeatClient
}

// IsInitialized returns whether GPU monitoring is initialized
func (m *GPUMonitor) IsInitialized() bool {
	return m.initialized
}

// NewGPUMonitor creates a new GPU monitor (minimal build - no GPU support)
func NewGPUMonitor() *GPUMonitor {
	monitor := &GPUMonitor{
		initialized:     false, // GPU support is disabled in minimal build
		gpuData:         make(map[string]interface{}),
		heartbeatClient: analytics.NewHeartbeatClient("v2.0", "webui-minimal"), // GPU Pro version, minimal mode
	}

	log.Printf("📦 Running minimal build - GPU monitoring is disabled")
	log.Printf("✓  System metrics will be available")

	// Start analytics heartbeat
	monitor.heartbeatClient.Start()

	return monitor
}

// GetGPUData returns empty data (no GPU in minimal build)
func (m *GPUMonitor) GetGPUData() (map[string]interface{}, error) {
	// Return empty map - no GPU data in minimal build
	return make(map[string]interface{}), nil
}

// GetProcesses returns empty list (no GPU processes in minimal build)
func (m *GPUMonitor) GetProcesses() ([]map[string]interface{}, error) {
	// Return empty slice - no GPU processes in minimal build
	return []map[string]interface{}{}, nil
}

// Shutdown shuts down the monitor
func (m *GPUMonitor) Shutdown() {
	// Stop heartbeat client
	if m.heartbeatClient != nil {
		m.heartbeatClient.Stop()
	}

	log.Println("Minimal monitor shutdown")
}

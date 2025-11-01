// +build darwin

package monitor

import (
	"log"
	"sync"

	"gpu-pro/analytics"
)

// GPUMonitor is a stub monitor for macOS (no GPU support)
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

// NewGPUMonitor creates a new GPU monitor (macOS - no GPU support)
func NewGPUMonitor() *GPUMonitor {
	monitor := &GPUMonitor{
		initialized:     false, // GPU support is disabled on macOS
		gpuData:         make(map[string]interface{}),
		heartbeatClient: analytics.NewHeartbeatClient("v2.0", "webui"), // GPU Pro version, WebUI mode
	}

	log.Printf("🍎 Running on macOS - GPU monitoring is disabled")
	log.Printf("✓  System metrics will be available")

	// Start analytics heartbeat
	monitor.heartbeatClient.Start()

	return monitor
}

// GetGPUData returns empty data (no GPU on macOS)
func (m *GPUMonitor) GetGPUData() (map[string]interface{}, error) {
	// Return empty map - no GPU data on macOS
	return make(map[string]interface{}), nil
}

// GetProcesses returns empty list (no GPU processes on macOS)
func (m *GPUMonitor) GetProcesses() ([]map[string]interface{}, error) {
	// Return empty slice - no GPU processes on macOS
	return []map[string]interface{}{}, nil
}

// Shutdown shuts down the monitor
func (m *GPUMonitor) Shutdown() {
	// Stop heartbeat client
	if m.heartbeatClient != nil {
		m.heartbeatClient.Stop()
	}

	log.Println("macOS monitor shutdown")
}

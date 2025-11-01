// +build linux

package monitor

import (
	"fmt"
	"log"
	"sync"

	"gpu-pro/analytics"

	"github.com/NVIDIA/go-nvml/pkg/nvml"
	"github.com/shirou/gopsutil/v3/process"
)

// GPUMonitor monitors NVIDIA GPUs using NVML (Linux)
type GPUMonitor struct {
	initialized     bool
	collector       *MetricsCollector
	useSMI          map[string]bool // Track which GPUs use nvidia-smi
	gpuData         map[string]interface{}
	mu              sync.RWMutex
	heartbeatClient *analytics.HeartbeatClient
}

// IsInitialized returns whether GPU monitoring is initialized
func (m *GPUMonitor) IsInitialized() bool {
	return m.initialized
}

// NewGPUMonitor creates a new GPU monitor
func NewGPUMonitor() *GPUMonitor {
	monitor := &GPUMonitor{
		collector:       NewMetricsCollector(),
		useSMI:          make(map[string]bool),
		gpuData:         make(map[string]interface{}),
		heartbeatClient: analytics.NewHeartbeatClient("v2.0", "webui"), // GPU Pro version, WebUI mode
	}

	// Initialize NVML
	if ret := nvml.Init(); ret != nvml.SUCCESS {
		log.Printf("⚠️  No NVIDIA GPU detected or NVML not available (error code: %v)", ret)
		log.Printf("✓  System metrics will still be available")
		monitor.initialized = false
		// Still start heartbeat even without GPU
		monitor.heartbeatClient.Start()
		return monitor
	}

	monitor.initialized = true
	version, ret := nvml.SystemGetDriverVersion()
	if ret == nvml.SUCCESS {
		log.Printf("NVML initialized - Driver: %s", version)
	}

	// Detect which GPUs need nvidia-smi
	monitor.detectSMIGPUs()

	// Start analytics heartbeat
	monitor.heartbeatClient.Start()

	return monitor
}

func (m *GPUMonitor) detectSMIGPUs() {
	count, ret := nvml.DeviceGetCount()
	if ret != nvml.SUCCESS {
		log.Printf("Failed to get device count: %v", nvml.ErrorString(ret))
		return
	}

	log.Printf("Detected %d GPU(s)", count)

	for i := 0; i < count; i++ {
		gpuID := fmt.Sprintf("%d", i)
		device, ret := nvml.DeviceGetHandleByIndex(i)
		if ret != nvml.SUCCESS {
			m.useSMI[gpuID] = true
			log.Printf("GPU %d: Failed to get handle, using nvidia-smi fallback", i)
			continue
		}

		// Try to collect data
		data := m.collector.CollectAll(device, gpuID)
		gpuName := "Unknown"
		if name, ok := data["name"].(string); ok {
			gpuName = name
		}

		// Check if utilization is available
		if util, ok := data["utilization"].(float64); !ok || util < 0 {
			m.useSMI[gpuID] = true
			log.Printf("GPU %d (%s): Utilization metric not available via NVML", i, gpuName)
			log.Printf("GPU %d (%s): Switching to nvidia-smi mode", i, gpuName)
		} else {
			m.useSMI[gpuID] = false
			log.Printf("GPU %d (%s): Using NVML (utilization: %.1f%%)", i, gpuName, util)
		}
	}

	nvmlCount := 0
	smiCount := 0
	for _, useSMI := range m.useSMI {
		if useSMI {
			smiCount++
		} else {
			nvmlCount++
		}
	}

	if smiCount > 0 {
		log.Printf("Boot detection complete: %d GPU(s) using NVML, %d GPU(s) using nvidia-smi", nvmlCount, smiCount)
	} else {
		log.Printf("Boot detection complete: All %d GPU(s) using NVML", nvmlCount)
	}
}

// GetGPUData collects metrics from all detected GPUs
func (m *GPUMonitor) GetGPUData() (map[string]interface{}, error) {
	if !m.initialized {
		// Return empty map instead of error to allow graceful degradation
		return make(map[string]interface{}), nil
	}

	count, ret := nvml.DeviceGetCount()
	if ret != nvml.SUCCESS {
		return make(map[string]interface{}), nil
	}

	gpuData := make(map[string]interface{})

	for i := 0; i < count; i++ {
		gpuID := fmt.Sprintf("%d", i)
		device, ret := nvml.DeviceGetHandleByIndex(i)
		if ret != nvml.SUCCESS {
			log.Printf("GPU %d: Failed to get handle: %v", i, nvml.ErrorString(ret))
			continue
		}

		// Collect GPU data
		data := m.collector.CollectAll(device, gpuID)
		gpuData[gpuID] = data
	}

	m.mu.Lock()
	m.gpuData = gpuData
	m.mu.Unlock()

	// Update GPU info for heartbeat (first GPU only for simplicity)
	if len(gpuData) > 0 {
		if gpu0, ok := gpuData["0"].(map[string]interface{}); ok {
			if name, ok := gpu0["name"].(string); ok {
				m.heartbeatClient.SetGPUInfo(name)
			}
		}
	}

	return gpuData, nil
}

// GetProcesses gets GPU process information
func (m *GPUMonitor) GetProcesses() ([]map[string]interface{}, error) {
	if !m.initialized {
		// Return empty slice instead of error to allow graceful degradation
		return []map[string]interface{}{}, nil
	}

	count, ret := nvml.DeviceGetCount()
	if ret != nvml.SUCCESS {
		return []map[string]interface{}{}, nil
	}

	var allProcesses []map[string]interface{}
	gpuProcessCounts := make(map[string]map[string]int)

	for i := 0; i < count; i++ {
		gpuID := fmt.Sprintf("%d", i)
		device, ret := nvml.DeviceGetHandleByIndex(i)
		if ret != nvml.SUCCESS {
			continue
		}

		uuid, ret := device.GetUUID()
		if ret != nvml.SUCCESS {
			continue
		}

		gpuProcessCounts[gpuID] = map[string]int{"compute": 0, "graphics": 0}

		// Get compute processes
		procs, ret := device.GetComputeRunningProcesses()
		if ret == nvml.SUCCESS {
			gpuProcessCounts[gpuID]["compute"] = len(procs)

			for _, proc := range procs {
				procName := getProcessName(int(proc.Pid))
				procInfo := map[string]interface{}{
					"pid":      fmt.Sprintf("%d", proc.Pid),
					"name":     procName,
					"gpu_uuid": uuid,
					"gpu_id":   gpuID,
					"memory":   float64(proc.UsedGpuMemory) / (1024 * 1024), // MB
				"type":     "compute",
				}

				// Get additional process information
				if p, err := process.NewProcess(int32(proc.Pid)); err == nil {
					// Get command line
					if cmdline, err := p.Cmdline(); err == nil {
						procInfo["command"] = cmdline
					}

					// Get CPU utilization
					if cpuPercent, err := p.CPUPercent(); err == nil {
						procInfo["cpu_percent"] = cpuPercent
					}
				}

				// Try to get GPU utilization (NVML may not provide per-process util on all GPUs)
				// Note: GetProcessUtilization is not available in all NVML versions
				// We'll track this via SM utilization if available
				procInfo["gpu_percent"] = 0.0 // Default, will be updated if available

				allProcesses = append(allProcesses, procInfo)
			}
		}

		// Get graphics processes
		graphicsProcs, ret := device.GetGraphicsRunningProcesses()
		if ret == nvml.SUCCESS {
			gpuProcessCounts[gpuID]["graphics"] = len(graphicsProcs)

		for _, proc := range graphicsProcs {
			procName := getProcessName(int(proc.Pid))
			procInfo := map[string]interface{}{
				"pid":      fmt.Sprintf("%d", proc.Pid),
				"name":     procName,
				"gpu_uuid": uuid,
				"gpu_id":   gpuID,
				"memory":   float64(proc.UsedGpuMemory) / (1024 * 1024), // MB
				"type":     "graphics",
			}

			// Get additional process information
			if p, err := process.NewProcess(int32(proc.Pid)); err == nil {
				// Get command line
				if cmdline, err := p.Cmdline(); err == nil {
					procInfo["command"] = cmdline
				}

				// Get CPU utilization
				if cpuPercent, err := p.CPUPercent(); err == nil {
					procInfo["cpu_percent"] = cpuPercent
				}
			}

			// Try to get GPU utilization (NVML may not provide per-process util on all GPUs)
			// Note: GetProcessUtilization is not available in all NVML versions
			// We'll track this via SM utilization if available
			procInfo["gpu_percent"] = 0.0 // Default, will be updated if available

			allProcesses = append(allProcesses, procInfo)
		}
		}
	}

	// Update GPU data with process counts
	m.mu.Lock()
	for gpuID, counts := range gpuProcessCounts {
		if data, ok := m.gpuData[gpuID].(map[string]interface{}); ok {
			data["compute_processes_count"] = counts["compute"]
			data["graphics_processes_count"] = counts["graphics"]
		}
	}
	m.mu.Unlock()

	return allProcesses, nil
}

// Shutdown shuts down NVML and analytics
func (m *GPUMonitor) Shutdown() {
	// Stop heartbeat client
	if m.heartbeatClient != nil {
		m.heartbeatClient.Stop()
	}

	if m.initialized {
		if ret := nvml.Shutdown(); ret != nvml.SUCCESS {
			log.Printf("Failed to shutdown NVML: %v", nvml.ErrorString(ret))
		} else {
			log.Println("NVML shutdown")
		}
		m.initialized = false
	}
}

// getProcessName extracts readable process name from PID
func getProcessName(pid int) string {
	proc, err := process.NewProcess(int32(pid))
	if err != nil {
		return fmt.Sprintf("PID:%d", pid)
	}

	// Try to get process name
	name, err := proc.Name()
	if err == nil && name != "" && name != "python" && name != "python3" && name != "sh" && name != "bash" {
		return name
	}

	// Try to get cmdline for better name extraction
	cmdline, err := proc.Cmdline()
	if err == nil && cmdline != "" {
		// Simple parsing - get the last part of the first meaningful argument
		parts := splitCmdline(cmdline)
		for _, part := range parts {
			if part == "" || part[0] == '-' {
				continue
			}
			if part == "python" || part == "python3" || part == "node" || part == "java" {
				continue
			}
			// Extract filename from path
			if idx := lastIndex(part, '/'); idx >= 0 {
				part = part[idx+1:]
			}
			if idx := lastIndex(part, '\\'); idx >= 0 {
				part = part[idx+1:]
			}
			if part != "" {
				return part
			}
		}
	}

	return fmt.Sprintf("PID:%d", pid)
}

func splitCmdline(cmdline string) []string {
	// Simple split by space - good enough for most cases
	var parts []string
	current := ""
	for _, c := range cmdline {
		if c == ' ' {
			if current != "" {
				parts = append(parts, current)
				current = ""
			}
		} else {
			current += string(c)
		}
	}
	if current != "" {
		parts = append(parts, current)
	}
	return parts
}

func lastIndex(s string, c rune) int {
	for i := len(s) - 1; i >= 0; i-- {
		if rune(s[i]) == c {
			return i
		}
	}
	return -1
}

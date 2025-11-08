// +build windows

package monitor

import (
	"encoding/csv"
	"fmt"
	"log"
	"os/exec"
	"strconv"
	"strings"
	"sync"

	"gpu-pro/analytics"

	"github.com/shirou/gopsutil/v3/process"
)

// GPUMonitor monitors NVIDIA GPUs using nvidia-smi on Windows
type GPUMonitor struct {
	initialized     bool
	gpuData         map[string]interface{}
	gpuCount        int
	mu              sync.RWMutex
	heartbeatClient *analytics.HeartbeatClient
}

// IsInitialized returns whether GPU monitoring is initialized
func (m *GPUMonitor) IsInitialized() bool {
	return m.initialized
}

// NewGPUMonitor creates a new GPU monitor using nvidia-smi
func NewGPUMonitor() *GPUMonitor {
	monitor := &GPUMonitor{
		gpuData:         make(map[string]interface{}),
		heartbeatClient: analytics.NewHeartbeatClient("v2.0", "webui-windows"),
	}

	// Check if nvidia-smi is available
	cmd := exec.Command("nvidia-smi", "--list-gpus")
	output, err := cmd.Output()
	if err != nil {
		log.Printf("⚠️  nvidia-smi not found or no NVIDIA GPU detected")
		log.Printf("✓  System metrics will still be available")
		monitor.initialized = false
		monitor.heartbeatClient.Start()
		return monitor
	}

	// Count GPUs
	lines := strings.Split(strings.TrimSpace(string(output)), "\n")
	monitor.gpuCount = len(lines)
	monitor.initialized = true

	log.Printf("✓  GPU monitoring initialized using nvidia-smi")
	log.Printf("✓  Detected %d GPU(s)", monitor.gpuCount)

	// Log GPU names
	for i, line := range lines {
		if strings.Contains(line, "GPU") {
			log.Printf("   GPU %d: %s", i, strings.TrimSpace(line))
		}
	}

	monitor.heartbeatClient.Start()
	return monitor
}

// GetGPUData collects metrics from all detected GPUs using nvidia-smi
func (m *GPUMonitor) GetGPUData() (map[string]interface{}, error) {
	if !m.initialized {
		return make(map[string]interface{}), nil
	}

	// Query nvidia-smi with CSV format for easy parsing
	// Fields: index, name, temperature.gpu, utilization.gpu, utilization.memory,
	//         memory.total, memory.used, memory.free, power.draw, power.limit,
	//         clocks.current.graphics, clocks.current.memory, fan.speed, pcie.link.gen.current,
	//         pcie.link.width.current
	queryFields := []string{
		"index",
		"name",
		"temperature.gpu",
		"utilization.gpu",
		"utilization.memory",
		"memory.total",
		"memory.used",
		"memory.free",
		"power.draw",
		"power.limit",
		"clocks.current.graphics",
		"clocks.current.memory",
		"fan.speed",
		"pcie.link.gen.current",
		"pcie.link.width.current",
		"uuid",
	}

	query := strings.Join(queryFields, ",")
	cmd := exec.Command("nvidia-smi", "--query-gpu="+query, "--format=csv,noheader,nounits")

	output, err := cmd.Output()
	if err != nil {
		log.Printf("Failed to query nvidia-smi: %v", err)
		return make(map[string]interface{}), nil
	}

	// Parse CSV output
	reader := csv.NewReader(strings.NewReader(string(output)))
	records, err := reader.ReadAll()
	if err != nil {
		log.Printf("Failed to parse nvidia-smi output: %v", err)
		return make(map[string]interface{}), nil
	}

	gpuData := make(map[string]interface{})

	for _, record := range records {
		if len(record) < len(queryFields) {
			continue
		}

		// Parse GPU index
		gpuID := strings.TrimSpace(record[0])

		// Helper function to parse float
		parseFloat := func(s string) float64 {
			s = strings.TrimSpace(s)
			if s == "[N/A]" || s == "N/A" || s == "" {
				return 0.0
			}
			val, err := strconv.ParseFloat(s, 64)
			if err != nil {
				return 0.0
			}
			return val
		}

		// Helper function to parse int
		parseInt := func(s string) int {
			s = strings.TrimSpace(s)
			if s == "[N/A]" || s == "N/A" || s == "" {
				return 0
			}
			val, err := strconv.Atoi(s)
			if err != nil {
				return 0
			}
			return val
		}

		data := map[string]interface{}{
			"id":                     gpuID,
			"name":                   strings.TrimSpace(record[1]),
			"temperature":            parseFloat(record[2]),
			"utilization":            parseFloat(record[3]),
			"memory_utilization":     parseFloat(record[4]),
			"memory_total":           parseFloat(record[5]),
			"memory_used":            parseFloat(record[6]),
			"memory_free":            parseFloat(record[7]),
			"power_draw":             parseFloat(record[8]),
			"power_limit":            parseFloat(record[9]),
			"clock_graphics":         parseFloat(record[10]),
			"clock_memory":           parseFloat(record[11]),
			"fan_speed":              parseFloat(record[12]),
			"pcie_link_gen":          parseInt(record[13]),
			"pcie_link_width":        parseInt(record[14]),
			"uuid":                   strings.TrimSpace(record[15]),
			"compute_processes_count": 0,
			"graphics_processes_count": 0,
		}

		gpuData[gpuID] = data
	}

	m.mu.Lock()
	m.gpuData = gpuData
	m.mu.Unlock()

	// Update GPU info for heartbeat (first GPU only)
	if len(gpuData) > 0 {
		if gpu0, ok := gpuData["0"].(map[string]interface{}); ok {
			if name, ok := gpu0["name"].(string); ok {
				m.heartbeatClient.SetGPUInfo(name)
			}
		}
	}

	return gpuData, nil
}

// GetProcesses gets GPU process information using nvidia-smi
func (m *GPUMonitor) GetProcesses() ([]map[string]interface{}, error) {
	if !m.initialized {
		return []map[string]interface{}{}, nil
	}

	// Query nvidia-smi for compute processes
	// Fields: gpu_uuid, pid, used_memory, process_name
	cmd := exec.Command("nvidia-smi", "--query-compute-apps=gpu_uuid,pid,used_memory,name", "--format=csv,noheader,nounits")

	output, err := cmd.Output()
	if err != nil {
		// This is OK - might just mean no processes running
		return []map[string]interface{}{}, nil
	}

	var allProcesses []map[string]interface{}
	gpuProcessCounts := make(map[string]map[string]int)

	// Initialize process counts
	for i := 0; i < m.gpuCount; i++ {
		gpuID := fmt.Sprintf("%d", i)
		gpuProcessCounts[gpuID] = map[string]int{"compute": 0, "graphics": 0}
	}

	// Parse CSV output
	reader := csv.NewReader(strings.NewReader(string(output)))
	records, err := reader.ReadAll()
	if err == nil {
		for _, record := range records {
			if len(record) < 4 {
				continue
			}

			uuid := strings.TrimSpace(record[0])
			pidStr := strings.TrimSpace(record[1])
			memStr := strings.TrimSpace(record[2])
			procName := strings.TrimSpace(record[3])

			pid, err := strconv.Atoi(pidStr)
			if err != nil {
				continue
			}

			memory := 0.0
			if memStr != "[N/A]" && memStr != "N/A" && memStr != "" {
				if mem, err := strconv.ParseFloat(memStr, 64); err == nil {
					memory = mem
				}
			}

			// Find GPU ID from UUID
			gpuID := ""
			m.mu.RLock()
			for id, data := range m.gpuData {
				if gpuData, ok := data.(map[string]interface{}); ok {
					if gpuUUID, ok := gpuData["uuid"].(string); ok && gpuUUID == uuid {
						gpuID = id
						break
					}
				}
			}
			m.mu.RUnlock()

			if gpuID == "" {
				continue
			}

			procInfo := map[string]interface{}{
				"pid":      fmt.Sprintf("%d", pid),
				"name":     procName,
				"gpu_uuid": uuid,
				"gpu_id":   gpuID,
				"memory":   memory,
				"type":     "compute",
			}

			// Get additional process information
			if p, err := process.NewProcess(int32(pid)); err == nil {
				if cmdline, err := p.Cmdline(); err == nil {
					procInfo["command"] = cmdline
				}
				if cpuPercent, err := p.CPUPercent(); err == nil {
					procInfo["cpu_percent"] = cpuPercent
				}
			}

			procInfo["gpu_percent"] = 0.0 // nvidia-smi doesn't provide per-process GPU util

			allProcesses = append(allProcesses, procInfo)
			gpuProcessCounts[gpuID]["compute"]++
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

// Shutdown shuts down the monitor and analytics
func (m *GPUMonitor) Shutdown() {
	if m.heartbeatClient != nil {
		m.heartbeatClient.Stop()
	}
	log.Println("GPU Monitor (nvidia-smi) shutdown")
}

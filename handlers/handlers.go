package handlers

import (
	"encoding/json"
	"log"
	"os"
	"runtime"
	"sync"
	"time"

	"gpu-pro/config"
	"gpu-pro/monitor"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/websocket/v2"
	"github.com/shirou/gopsutil/v3/cpu"
	"github.com/shirou/gopsutil/v3/disk"
	"github.com/shirou/gopsutil/v3/mem"
)

// WebSocketClients holds all connected WebSocket clients
type WebSocketClients struct {
	clients map[*websocket.Conn]bool
	mu      sync.RWMutex
}

func NewWebSocketClients() *WebSocketClients {
	return &WebSocketClients{
		clients: make(map[*websocket.Conn]bool),
	}
}

func (wsc *WebSocketClients) Add(conn *websocket.Conn) {
	wsc.mu.Lock()
	defer wsc.mu.Unlock()
	wsc.clients[conn] = true
}

func (wsc *WebSocketClients) Remove(conn *websocket.Conn) {
	wsc.mu.Lock()
	defer wsc.mu.Unlock()
	delete(wsc.clients, conn)
}

func (wsc *WebSocketClients) Broadcast(data []byte) {
	wsc.mu.RLock()
	defer wsc.mu.RUnlock()

	var disconnected []*websocket.Conn
	for conn := range wsc.clients {
		if err := conn.WriteMessage(websocket.TextMessage, data); err != nil {
			disconnected = append(disconnected, conn)
		}
	}

	// Remove disconnected clients
	if len(disconnected) > 0 {
		wsc.mu.RUnlock()
		wsc.mu.Lock()
		for _, conn := range disconnected {
			delete(wsc.clients, conn)
		}
		wsc.mu.Unlock()
		wsc.mu.RLock()
	}
}

func (wsc *WebSocketClients) Count() int {
	wsc.mu.RLock()
	defer wsc.mu.RUnlock()
	return len(wsc.clients)
}

// RegisterHandlers registers FastAPI WebSocket handlers for monitor mode
func RegisterHandlers(app *fiber.App, mon *monitor.GPUMonitor, cfg *config.Config) {
	wsClients := NewWebSocketClients()
	monitorRunning := false
	var monitorMu sync.Mutex

	// API endpoint to get user's home directory
	app.Get("/api/home-directory", func(c *fiber.Ctx) error {
		homeDir, err := os.UserHomeDir()
		if err != nil {
			homeDir = "/" // Fallback to root if home directory cannot be determined
		}
		return c.JSON(fiber.Map{
			"home": homeDir,
		})
	})

	// API endpoint for fetching largest files from a specific directory
	app.Get("/api/largest-files", func(c *fiber.Ctx) error {
		directory := c.Query("directory", "/")
		files := GetTopLargestFiles(10, directory)
		return c.JSON(fiber.Map{
			"directory": directory,
			"files":     files,
		})
	})

	// API endpoint to get alert thresholds
	app.Get("/api/alert-thresholds", func(c *fiber.Ctx) error {
		thresholds, err := loadAlertThresholds()
		if err != nil {
			// Return default thresholds if file doesn't exist
			thresholds = getDefaultThresholds()
		}
		return c.JSON(thresholds)
	})

	// API endpoint to save alert thresholds
	app.Post("/api/alert-thresholds", func(c *fiber.Ctx) error {
		var thresholds map[string]interface{}
		if err := c.BodyParser(&thresholds); err != nil {
			return c.Status(400).JSON(fiber.Map{
				"error": "Invalid request body",
			})
		}

		if err := saveAlertThresholds(thresholds); err != nil {
			return c.Status(500).JSON(fiber.Map{
				"error": "Failed to save thresholds",
			})
		}

		return c.JSON(fiber.Map{
			"success": true,
			"message": "Alert thresholds saved successfully",
		})
	})

	// API endpoint to get alert history
	app.Get("/api/alert-history", func(c *fiber.Ctx) error {
		limit := c.QueryInt("limit", 50)
		alerts, err := loadAlertHistory(limit)
		if err != nil {
			return c.JSON(fiber.Map{
				"alerts": []interface{}{},
			})
		}
		return c.JSON(fiber.Map{
			"alerts": alerts,
		})
	})

	// WebSocket endpoint
	app.Get("/socket.io/", websocket.New(func(c *websocket.Conn) {
		wsClients.Add(c)
		log.Println("Dashboard client connected")

		// Start monitor loop if not already running
		monitorMu.Lock()
		if !monitorRunning {
			monitorRunning = true
			go monitorLoop(mon, wsClients, cfg)
		}
		monitorMu.Unlock()

		// Send immediate initial data to clear loading state
		go func() {
			sendInitialData(mon, c, cfg)
		}()

		// Keep connection alive
		for {
			if _, _, err := c.ReadMessage(); err != nil {
				log.Printf("Dashboard client disconnected: %v", err)
				break
			}
		}

		wsClients.Remove(c)
	}))
}

// sendInitialData sends immediate data to a newly connected client to clear loading state
func sendInitialData(mon *monitor.GPUMonitor, conn *websocket.Conn, cfg *config.Config) {
	// Collect initial data (will be empty if no GPU)
	gpuData, _ := mon.GetGPUData()
	processes, _ := mon.GetProcesses()

	// Ensure we have valid empty structures if nil
	if gpuData == nil {
		gpuData = make(map[string]interface{})
	}
	if processes == nil {
		processes = []map[string]interface{}{}
	}

	// Get system info
	// Use 500ms interval for CPU to get actual reading (shorter intervals return 0 on macOS)
	cpuPercent, _ := cpu.Percent(500*time.Millisecond, false)
	memInfo, _ := mem.VirtualMemory()

	systemInfo := map[string]interface{}{
		"cpu_percent":    0.0,
		"memory_percent": 0.0,
		"disk_percent":   0.0,
		"timestamp":      time.Now().Format(time.RFC3339),
	}

	if len(cpuPercent) > 0 {
		systemInfo["cpu_percent"] = cpuPercent[0]
	}
	if memInfo != nil {
		systemInfo["memory_percent"] = memInfo.UsedPercent
	}

	// Get disk usage for root partition
	diskPath := "/"
	if runtime.GOOS == "windows" {
		diskPath = "C:\\"
	}
	diskUsage, err := disk.Usage(diskPath)
	if err == nil {
		systemInfo["disk_percent"] = diskUsage.UsedPercent
		systemInfo["disk_used"] = float64(diskUsage.Used) / (1024 * 1024 * 1024)
		systemInfo["disk_total"] = float64(diskUsage.Total) / (1024 * 1024 * 1024)
	}

	// Build response
	response := map[string]interface{}{
		"mode":           cfg.Mode,
		"node_name":      cfg.NodeName,
		"gpus":           gpuData,
		"processes":      processes,
		"system":         systemInfo,
		"system_metrics": make(map[string]interface{}), // Empty for initial load
	}

	// Send to the client
	data, err := json.Marshal(response)
	if err != nil {
		log.Printf("Error marshaling initial response: %v", err)
		return
	}

	if err := conn.WriteMessage(websocket.TextMessage, data); err != nil {
		log.Printf("Error sending initial data: %v", err)
	}
}

// monitorLoop is the background loop that collects and emits GPU data
func monitorLoop(mon *monitor.GPUMonitor, wsClients *WebSocketClients, cfg *config.Config) {
	// Determine update interval
	updateInterval := cfg.UpdateInterval
	log.Printf("Using polling interval: %.2fs", updateInterval)

	ticker := time.NewTicker(time.Duration(updateInterval * float64(time.Second)))
	defer ticker.Stop()

	for range ticker.C {
		// Skip if no clients connected
		if wsClients.Count() == 0 {
			continue
		}

		// Collect GPU data and processes (will return empty if not initialized)
		var gpuData map[string]interface{}
		var processes []map[string]interface{}

		gpuData, _ = mon.GetGPUData()
		processes, _ = mon.GetProcesses()

		// Ensure we have valid empty structures if nil
		if gpuData == nil {
			gpuData = make(map[string]interface{})
		}
		if processes == nil {
			processes = []map[string]interface{}{}
		}

		// Get system info
		// Use 500ms interval for CPU to get actual reading (shorter intervals return 0 on macOS)
		cpuPercent, _ := cpu.Percent(500*time.Millisecond, false)
		memInfo, _ := mem.VirtualMemory()

		systemInfo := map[string]interface{}{
			"cpu_percent":    0.0,
			"memory_percent": 0.0,
			"disk_percent":   0.0,
			"disk_read_rate": 0.0,
			"disk_write_rate": 0.0,
			"timestamp":      time.Now().Format(time.RFC3339),
		}

		if len(cpuPercent) > 0 {
			systemInfo["cpu_percent"] = cpuPercent[0]
		}
		if memInfo != nil {
			systemInfo["memory_percent"] = memInfo.UsedPercent
		}

		// Get disk usage for root partition
		// Use platform-appropriate path (/ for Unix, C:\ for Windows)
		diskPath := "/"
		if runtime.GOOS == "windows" {
			diskPath = "C:\\"
		}
		diskUsage, err := disk.Usage(diskPath)
		if err == nil {
			systemInfo["disk_percent"] = diskUsage.UsedPercent
			systemInfo["disk_used"] = float64(diskUsage.Used) / (1024 * 1024 * 1024)  // GB
			systemInfo["disk_total"] = float64(diskUsage.Total) / (1024 * 1024 * 1024) // GB

		// Get system fan speeds (Linux only)
		fans := getSystemFanSpeeds()
		if len(fans) > 0 {
			systemInfo["system_fans"] = fans
			avgRPM := getAverageFanSpeed(fans)
			maxRPM := getMaxFanSpeed(fans)
			// Calculate percentage (assuming max RPM of 3000 or actual max seen)
			maxReference := maxRPM
			if maxReference < 3000 {
				maxReference = 3000
			}
			systemInfo["system_fan_speed"] = float64(avgRPM)
			systemInfo["system_fan_percent"] = (float64(avgRPM) / float64(maxReference)) * 100
		}
		}

		// Get extended system metrics (network I/O, disk I/O, connections, large files)
		systemMetrics := GetSystemMetrics()

		// Build response
		response := map[string]interface{}{
			"mode":           cfg.Mode,
			"node_name":      cfg.NodeName,
			"gpus":           gpuData,
			"processes":      processes,
			"system":         systemInfo,
			"system_metrics": systemMetrics,
		}

		// Send to all connected clients
		data, err := json.Marshal(response)
		if err != nil {
			log.Printf("Error marshaling response: %v", err)
			continue
		}

		wsClients.Broadcast(data)
	}
}

// Alert Management Functions

// getDefaultThresholds returns default alert threshold values
func getDefaultThresholds() map[string]interface{} {
	return map[string]interface{}{
		"temp_warning":     75,
		"temp_critical":    85,
		"memory_warning":   85,
		"memory_critical":  95,
		"power_warning":    90,
		"power_critical":   98,
	}
}

// loadAlertThresholds loads alert thresholds from gpu-thresholds.json
func loadAlertThresholds() (map[string]interface{}, error) {
	data, err := os.ReadFile("gpu-thresholds.json")
	if err != nil {
		return nil, err
	}

	var thresholds map[string]interface{}
	if err := json.Unmarshal(data, &thresholds); err != nil {
		return nil, err
	}

	return thresholds, nil
}

// saveAlertThresholds saves alert thresholds to gpu-thresholds.json
func saveAlertThresholds(thresholds map[string]interface{}) error {
	data, err := json.MarshalIndent(thresholds, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile("gpu-thresholds.json", data, 0644)
}

// Alert represents a single alert record
type Alert struct {
	Timestamp string      `json:"timestamp"`
	GPUIndex  int         `json:"gpu_index"`
	GPUName   string      `json:"gpu_name"`
	Level     string      `json:"level"`
	Metric    string      `json:"metric"`
	Value     float64     `json:"value"`
	Threshold float64     `json:"threshold"`
	Message   string      `json:"message"`
}

// loadAlertHistory loads recent alerts from gpu-alerts.log
func loadAlertHistory(limit int) ([]Alert, error) {
	data, err := os.ReadFile("gpu-alerts.log")
	if err != nil {
		return []Alert{}, nil // Return empty array if file doesn't exist
	}

	lines := []string{}
	for _, line := range splitLines(string(data)) {
		if line != "" {
			lines = append(lines, line)
		}
	}

	// Reverse to get newest first
	for i, j := 0, len(lines)-1; i < j; i, j = i+1, j-1 {
		lines[i], lines[j] = lines[j], lines[i]
	}

	// Limit the number of alerts
	if len(lines) > limit {
		lines = lines[:limit]
	}

	alerts := []Alert{}
	for _, line := range lines {
		alert := parseAlertLine(line)
		if alert != nil {
			alerts = append(alerts, *alert)
		}
	}

	return alerts, nil
}

// splitLines splits a string into lines
func splitLines(s string) []string {
	var lines []string
	start := 0
	for i := 0; i < len(s); i++ {
		if s[i] == '\n' {
			lines = append(lines, s[start:i])
			start = i + 1
		}
	}
	if start < len(s) {
		lines = append(lines, s[start:])
	}
	return lines
}

// parseAlertLine parses a single line from gpu-alerts.log
// Format: [2025-10-28 14:27:39] GPU 0 - warning Memory: 85.0 (threshold: 85.0)
func parseAlertLine(line string) *Alert {
	if len(line) < 20 {
		return nil
	}

	// Extract timestamp (first 21 characters: [YYYY-MM-DD HH:MM:SS])
	if line[0] != '[' {
		return nil
	}
	timestampEnd := 20
	if len(line) <= timestampEnd {
		return nil
	}
	timestamp := line[1:timestampEnd]

	// Parse the rest of the line
	// Format: GPU X - LEVEL Metric: Value (threshold: Threshold)
	rest := line[timestampEnd+2:] // Skip "] "

	alert := Alert{
		Timestamp: timestamp,
	}

	// This is a simple parser - you might want to make it more robust
	// For now, we'll just store the raw message
	alert.Message = rest

	return &alert
}

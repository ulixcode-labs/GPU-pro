package main

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"sort"
	"strconv"
	"strings"
	"time"

	"gpu-pro/analytics"
	"gpu-pro/config"
	"gpu-pro/monitor"

	"github.com/charmbracelet/bubbles/progress"
	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/shirou/gopsutil/v3/cpu"
	"github.com/shirou/gopsutil/v3/disk"
	"github.com/shirou/gopsutil/v3/mem"
)

const (
	historySize        = 60 // 60 data points for history (30 seconds at 0.5s refresh)
	sparklineLength    = 20 // Number of characters to display in sparkline
	alertHistoryLog    = "gpu-alerts.log"
)

// Styles
var (
	// Colors
	primaryColor   = lipgloss.Color("#4facfe")
	secondaryColor = lipgloss.Color("#00f2fe")
	successColor   = lipgloss.Color("#43e97b")
	warningColor   = lipgloss.Color("#fa709a")
	dangerColor    = lipgloss.Color("#f5576c")
	textColor      = lipgloss.Color("#ffffff")
	mutedColor     = lipgloss.Color("#888888")

	// Styles
	titleStyle = lipgloss.NewStyle().
			Foreground(primaryColor).
			Bold(true).
			Padding(0, 1)

	headerStyle = lipgloss.NewStyle().
			Foreground(secondaryColor).
			Bold(true).
			Underline(true)

	boxStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(primaryColor).
			Padding(0, 1).
			Margin(0, 1)

	metricStyle = lipgloss.NewStyle().
			Foreground(textColor).
			Padding(0, 1)

	labelStyle = lipgloss.NewStyle().
			Foreground(mutedColor).
			Width(20)

	valueStyle = lipgloss.NewStyle().
			Foreground(textColor).
			Bold(true)

	helpStyle = lipgloss.NewStyle().
			Foreground(mutedColor).
			Italic(true)

	alertStyle = lipgloss.NewStyle().
			Foreground(dangerColor).
			Bold(true)

	warningStyle = lipgloss.NewStyle().
			Foreground(warningColor).
			Bold(true)

	selectedStyle = lipgloss.NewStyle().
			Foreground(primaryColor).
			Background(lipgloss.Color("#1a1a1a")).
			Bold(true)
)

// MetricHistory stores historical data for sparklines
type MetricHistory struct {
	Utilization []float64
	Temperature []float64
	Memory      []float64
	Power       []float64
	MFU         []float64
}

// Alert represents a threshold alert
type Alert struct {
	Timestamp    time.Time
	GPUId        int
	Metric       string
	Value        float64
	Threshold    float64
	Level        string // "warning" or "critical"
	Acknowledged bool
	Snoozed      bool
	SnoozeUntil  time.Time
	Resolved     bool
	ResolvedAt   time.Time
}

// Thresholds configuration
type Thresholds struct {
	TempWarning     float64 `json:"temp_warning"`
	TempCritical    float64 `json:"temp_critical"`
	MemoryWarning   float64 `json:"memory_warning"`
	MemoryCritical  float64 `json:"memory_critical"`
	PowerWarning    float64 `json:"power_warning"`
	PowerCritical   float64 `json:"power_critical"`
}

// ProcessSort type
type ProcessSort string

const (
	SortByMemory ProcessSort = "memory"
	SortByGPU    ProcessSort = "gpu"
	SortByCPU    ProcessSort = "cpu"
	SortByPID    ProcessSort = "pid"
	SortByName   ProcessSort = "name"
)

// Model represents the TUI application state
type model struct {
	monitor         *monitor.GPUMonitor
	cfg             *config.Config
	spinner         spinner.Model
	progress        progress.Model
	gpuData         []map[string]interface{}
	processes       []map[string]interface{}
	systemInfo      map[string]interface{}
	width           int
	height          int
	err             error

	// Historical data
	gpuHistory      map[int]*MetricHistory

	// Alert system
	thresholds      Thresholds
	alerts          []Alert
	activeAlerts    map[string]bool

	// Process management
	processMode     bool
	selectedProcess int
	processFilter   string
	processSort     ProcessSort
	searchMode      bool
	searchInput     textinput.Model

	// Alert viewing
	alertViewMode   bool
	selectedAlert   int

	// Analytics
	heartbeatClient *analytics.HeartbeatClient
}

// Messages
type tickMsg time.Time
type dataMsg struct {
	gpus      []map[string]interface{}
	processes []map[string]interface{}
	system    map[string]interface{}
}

// Initialize the model
func initialModel() model {
	cfg := config.Load()

	// Initialize GPU monitor
	mon := monitor.NewGPUMonitor()

	// Initialize spinner
	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = lipgloss.NewStyle().Foreground(primaryColor)

	// Initialize progress bar
	p := progress.New(progress.WithGradient(string(primaryColor), string(secondaryColor)))

	// Load thresholds from config or use defaults
	thresholds := loadThresholds()

	// Initialize search input
	ti := textinput.New()
	ti.Placeholder = "Search processes..."
	ti.CharLimit = 50
	ti.Width = 30

	// Initialize analytics heartbeat client
	heartbeat := analytics.NewHeartbeatClient("v2.0", "tui")
	heartbeat.Start()

	return model{
		monitor:         mon,
		cfg:             cfg,
		spinner:         s,
		progress:        p,
		gpuHistory:      make(map[int]*MetricHistory),
		thresholds:      thresholds,
		alerts:          []Alert{},
		activeAlerts:    make(map[string]bool),
		processSort:     SortByMemory,
		searchInput:     ti,
		heartbeatClient: heartbeat,
	}
}

// Load thresholds from config file
func loadThresholds() Thresholds {
	defaults := Thresholds{
		TempWarning:    75.0,
		TempCritical:   85.0,
		MemoryWarning:  85.0,
		MemoryCritical: 95.0,
		PowerWarning:   90.0,
		PowerCritical:  98.0,
	}

	// Try to load from file
	data, err := os.ReadFile("gpu-thresholds.json")
	if err != nil {
		// Save defaults
		saveThresholds(defaults)
		return defaults
	}

	var t Thresholds
	if err := json.Unmarshal(data, &t); err != nil {
		return defaults
	}
	return t
}

// Save thresholds to config file
func saveThresholds(t Thresholds) {
	data, err := json.MarshalIndent(t, "", "  ")
	if err != nil {
		return
	}
	os.WriteFile("gpu-thresholds.json", data, 0644)
}

// Init initializes the application
func (m model) Init() tea.Cmd {
	return tea.Batch(
		m.spinner.Tick,
		tickCmd(),
	)
}

// tickCmd returns a command that waits for the next tick
func tickCmd() tea.Cmd {
	return tea.Tick(time.Duration(500*time.Millisecond), func(t time.Time) tea.Msg {
		return tickMsg(t)
	})
}

// fetchData fetches GPU, process, and system data
func (m model) fetchData() tea.Msg {
	// Initialize with empty data structures
	gpus := []map[string]interface{}{}
	processes := []map[string]interface{}{}

	if m.monitor != nil {
		// Get GPU data (returns map[gpuID]data, empty if not initialized)
		gpuDataMap, _ := m.monitor.GetGPUData()

		// Convert map to slice for easier iteration
		if gpuDataMap != nil {
			for _, gpuData := range gpuDataMap {
				if data, ok := gpuData.(map[string]interface{}); ok {
					gpus = append(gpus, data)
				}
			}
		}

		// Get processes (returns empty slice if not initialized)
		processes, _ = m.monitor.GetProcesses()
		if processes == nil {
			processes = []map[string]interface{}{}
		}
	}

	// Get system info
	// Use 500ms interval for CPU to get actual reading (shorter intervals return 0 on macOS)
	cpuPercent, _ := cpu.Percent(500*time.Millisecond, false)
	memInfo, _ := mem.VirtualMemory()
	diskUsage, _ := disk.Usage("/")

	systemInfo := map[string]interface{}{
		"cpu_percent":    0.0,
		"memory_percent": 0.0,
		"disk_percent":   0.0,
		"disk_used_gb":   0.0,
		"disk_total_gb":  0.0,
		"disk_free_gb":   0.0,
	}

	if len(cpuPercent) > 0 {
		systemInfo["cpu_percent"] = cpuPercent[0]
	}
	if memInfo != nil {
		systemInfo["memory_percent"] = memInfo.UsedPercent
	}
	if diskUsage != nil {
		systemInfo["disk_percent"] = diskUsage.UsedPercent
		systemInfo["disk_used_gb"] = float64(diskUsage.Used) / (1024 * 1024 * 1024)
		systemInfo["disk_total_gb"] = float64(diskUsage.Total) / (1024 * 1024 * 1024)
		systemInfo["disk_free_gb"] = float64(diskUsage.Free) / (1024 * 1024 * 1024)
	}

	return dataMsg{
		gpus:      gpus,
		processes: processes,
		system:    systemInfo,
	}
}

// Update handles messages and updates the model
func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		// Search mode handling
		if m.searchMode {
			switch msg.String() {
			case "enter", "esc":
				m.searchMode = false
				m.processFilter = m.searchInput.Value()
				return m, nil
			default:
				var cmd tea.Cmd
				m.searchInput, cmd = m.searchInput.Update(msg)
				return m, cmd
			}
		}

		// Normal key handling
		switch msg.String() {
		case "q", "ctrl+c":
			if m.monitor != nil {
				m.monitor.Shutdown()
			}
			if m.heartbeatClient != nil {
				m.heartbeatClient.Stop()
			}
			return m, tea.Quit
		case "r":
			// Refresh
			if !m.alertViewMode {
				return m, m.fetchData
			}
			return m, nil
		case "p":
			// Toggle process management mode (not in alert view)
			if !m.alertViewMode {
				m.processMode = !m.processMode
				if m.processMode {
					m.selectedProcess = 0
				}
			}
			return m, nil
		case "up", "k":
			if m.alertViewMode && len(m.alerts) > 0 {
				m.selectedAlert = max(0, m.selectedAlert-1)
			} else if m.processMode && len(m.processes) > 0 {
				m.selectedProcess = max(0, m.selectedProcess-1)
			}
			return m, nil
		case "down", "j":
			if m.alertViewMode && len(m.alerts) > 0 {
				displayLimit := min(len(m.alerts), 20)
				m.selectedAlert = min(displayLimit-1, m.selectedAlert+1)
			} else if m.processMode && len(m.processes) > 0 {
				m.selectedProcess = min(len(m.processes)-1, m.selectedProcess+1)
			}
			return m, nil
		case "K":
			// Kill selected process
			if m.processMode && len(m.processes) > 0 {
				m.killSelectedProcess()
			}
			return m, m.fetchData
		case "/":
			// Enter search mode (not in alert view)
			if !m.alertViewMode {
				m.searchMode = true
				m.searchInput.Focus()
				return m, textinput.Blink
			}
			return m, nil
		case "c":
			// Clear search filter (not in alert view)
			if !m.alertViewMode {
				m.processFilter = ""
				m.searchInput.SetValue("")
			}
			return m, nil
		case "s":
			if m.alertViewMode {
				// Snooze selected alert for 5 minutes
				m.snoozeAlert(m.selectedAlert, 5*time.Minute)
			} else {
				// Cycle sort order in normal mode
				m.cycleSortOrder()
			}
			return m, nil
		case "S":
			if m.alertViewMode {
				// Snooze selected alert for 30 minutes
				m.snoozeAlert(m.selectedAlert, 30*time.Minute)
			}
			return m, nil
		case "A":
			// Acknowledge selected alert (capital A in alert view)
			if m.alertViewMode {
				m.acknowledgeAlert(m.selectedAlert)
			}
			return m, nil
		case "a":
			// Toggle alert view mode
			m.alertViewMode = !m.alertViewMode
			if m.alertViewMode {
				m.selectedAlert = 0
			}
			return m, nil
		}

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, nil

	case tickMsg:
		return m, tea.Batch(
			tickCmd(),
			m.fetchData,
		)

	case dataMsg:
		m.gpuData = msg.gpus
		m.processes = msg.processes
		m.systemInfo = msg.system

		// Update GPU info for analytics
		if m.heartbeatClient != nil && len(msg.gpus) > 0 {
			// Collect GPU names for analytics
			gpuNames := []string{}
			for _, gpu := range msg.gpus {
				if name := getString(gpu, "name", ""); name != "" {
					gpuNames = append(gpuNames, name)
				}
			}
			if len(gpuNames) > 0 {
				m.heartbeatClient.SetGPUInfo(strings.Join(gpuNames, ", "))
			}
		}

		// Update history and check alerts
		m.updateHistory()
		m.checkAlerts()
		m.cleanupOldAlerts()

		return m, nil

	case spinner.TickMsg:
		var cmd tea.Cmd
		m.spinner, cmd = m.spinner.Update(msg)
		return m, cmd
	}

	return m, nil
}

// Snooze alert
func (m *model) snoozeAlert(alertIdx int, duration time.Duration) {
	// Get the actual alert index (accounting for reverse display)
	actualIdx := m.getActualAlertIndex(alertIdx)
	if actualIdx < 0 || actualIdx >= len(m.alerts) {
		return
	}

	alert := &m.alerts[actualIdx]
	alert.Snoozed = true
	alert.SnoozeUntil = time.Now().Add(duration)

	// Remove from active alerts temporarily
	key := fmt.Sprintf("gpu%d_%s_%s", alert.GPUId, alert.Metric, alert.Level)
	delete(m.activeAlerts, key)
}

// Acknowledge alert
func (m *model) acknowledgeAlert(alertIdx int) {
	// Get the actual alert index (accounting for reverse display)
	actualIdx := m.getActualAlertIndex(alertIdx)
	if actualIdx < 0 || actualIdx >= len(m.alerts) {
		return
	}

	alert := &m.alerts[actualIdx]
	alert.Acknowledged = true

	// Remove from active alerts permanently
	key := fmt.Sprintf("gpu%d_%s_%s", alert.GPUId, alert.Metric, alert.Level)
	delete(m.activeAlerts, key)
}

// Get actual alert index (for reverse display)
func (m *model) getActualAlertIndex(displayIdx int) int {
	startIdx := 0
	if len(m.alerts) > 20 {
		startIdx = len(m.alerts) - 20
	}
	recentAlerts := m.alerts[startIdx:]
	actualIdx := len(recentAlerts) - 1 - displayIdx + startIdx
	return actualIdx
}

// Clean up old alerts with smart TTL: 30s for resolved, 1min for acknowledged, 1hour for active
func (m *model) cleanupOldAlerts() {
	now := time.Now()
	oneHourAgo := now.Add(-1 * time.Hour)
	oneMinuteAgo := now.Add(-1 * time.Minute)
	thirtySecondsAgo := now.Add(-30 * time.Second)

	// Filter out old alerts
	newAlerts := []Alert{}
	for _, alert := range m.alerts {
		// Resolved alerts expire after 30 seconds (from resolution time)
		if alert.Resolved {
			if alert.ResolvedAt.After(thirtySecondsAgo) {
				newAlerts = append(newAlerts, alert)
			}
			continue
		}

		// Acknowledged alerts expire after 1 minute (from trigger time)
		if alert.Acknowledged {
			if alert.Timestamp.After(oneMinuteAgo) {
				newAlerts = append(newAlerts, alert)
			}
			continue
		}

		// Regular/Active alerts expire after 1 hour (from trigger time)
		if alert.Timestamp.After(oneHourAgo) {
			// Also check if snooze has expired
			if alert.Snoozed && now.After(alert.SnoozeUntil) {
				alert.Snoozed = false
			}
			newAlerts = append(newAlerts, alert)
		}
	}
	m.alerts = newAlerts
}

// Update historical data
func (m *model) updateHistory() {
	for i, gpu := range m.gpuData {
		if _, exists := m.gpuHistory[i]; !exists {
			m.gpuHistory[i] = &MetricHistory{
				Utilization: []float64{},
				Temperature: []float64{},
				Memory:      []float64{},
				Power:       []float64{},
				MFU:         []float64{},
			}
		}

		hist := m.gpuHistory[i]

		// Add new values
		hist.Utilization = append(hist.Utilization, getFloat(gpu, "utilization", 0))
		hist.Temperature = append(hist.Temperature, getFloat(gpu, "temperature", 0))

		memUsed := getFloat(gpu, "memory_used", 0)
		memTotal := getFloat(gpu, "memory_total", 1)
		hist.Memory = append(hist.Memory, (memUsed/memTotal)*100)

		power := getFloat(gpu, "power_draw", 0)
		powerLimit := getFloat(gpu, "power_limit", 1)
		hist.Power = append(hist.Power, (power/powerLimit)*100)

		mfu := getFloat(gpu, "mfu", 0)
		hist.MFU = append(hist.MFU, mfu)

		// Keep only last N values
		if len(hist.Utilization) > historySize {
			hist.Utilization = hist.Utilization[1:]
		}
		if len(hist.Temperature) > historySize {
			hist.Temperature = hist.Temperature[1:]
		}
		if len(hist.Memory) > historySize {
			hist.Memory = hist.Memory[1:]
		}
		if len(hist.Power) > historySize {
			hist.Power = hist.Power[1:]
		}
		if len(hist.MFU) > historySize {
			hist.MFU = hist.MFU[1:]
		}
	}
}

// Check for threshold violations
func (m *model) checkAlerts() {
	now := time.Now()

	for i, gpu := range m.gpuData {
		temp := getFloat(gpu, "temperature", 0)
		memUsed := getFloat(gpu, "memory_used", 0)
		memTotal := getFloat(gpu, "memory_total", 1)
		memPercent := (memUsed / memTotal) * 100
		power := getFloat(gpu, "power_draw", 0)
		powerLimit := getFloat(gpu, "power_limit", 1)
		powerPercent := (power / powerLimit) * 100

		// Check temperature
		if temp >= m.thresholds.TempCritical {
			m.addAlert(Alert{
				Timestamp: now,
				GPUId:     i,
				Metric:    "Temperature",
				Value:     temp,
				Threshold: m.thresholds.TempCritical,
				Level:     "critical",
			})
		} else if temp >= m.thresholds.TempWarning {
			m.addAlert(Alert{
				Timestamp: now,
				GPUId:     i,
				Metric:    "Temperature",
				Value:     temp,
				Threshold: m.thresholds.TempWarning,
				Level:     "warning",
			})
		} else {
			// Mark as resolved if it was active
			m.resolveAlert(i, "Temperature", "critical")
			m.resolveAlert(i, "Temperature", "warning")
		}

		// Check memory
		if memPercent >= m.thresholds.MemoryCritical {
			m.addAlert(Alert{
				Timestamp: now,
				GPUId:     i,
				Metric:    "Memory",
				Value:     memPercent,
				Threshold: m.thresholds.MemoryCritical,
				Level:     "critical",
			})
		} else if memPercent >= m.thresholds.MemoryWarning {
			m.addAlert(Alert{
				Timestamp: now,
				GPUId:     i,
				Metric:    "Memory",
				Value:     memPercent,
				Threshold: m.thresholds.MemoryWarning,
				Level:     "warning",
			})
		} else {
			// Mark as resolved if it was active
			m.resolveAlert(i, "Memory", "critical")
			m.resolveAlert(i, "Memory", "warning")
		}

		// Check power
		if powerPercent >= m.thresholds.PowerCritical {
			m.addAlert(Alert{
				Timestamp: now,
				GPUId:     i,
				Metric:    "Power",
				Value:     powerPercent,
				Threshold: m.thresholds.PowerCritical,
				Level:     "critical",
			})
		} else if powerPercent >= m.thresholds.PowerWarning {
			m.addAlert(Alert{
				Timestamp: now,
				GPUId:     i,
				Metric:    "Power",
				Value:     powerPercent,
				Threshold: m.thresholds.PowerWarning,
				Level:     "warning",
			})
		} else {
			// Mark as resolved if it was active
			m.resolveAlert(i, "Power", "critical")
			m.resolveAlert(i, "Power", "warning")
		}
	}
}

// Add alert and log it
func (m *model) addAlert(alert Alert) {
	key := fmt.Sprintf("gpu%d_%s_%s", alert.GPUId, alert.Metric, alert.Level)

	// Only add if not already active (debounce)
	if !m.activeAlerts[key] {
		m.alerts = append(m.alerts, alert)
		m.activeAlerts[key] = true

		// Log to file
		m.logAlert(alert)

		// Keep only last 100 alerts in memory
		if len(m.alerts) > 100 {
			m.alerts = m.alerts[1:]
		}
	}
}

// Mark alert as resolved when condition returns to normal
func (m *model) resolveAlert(gpuId int, metric string, level string) {
	key := fmt.Sprintf("gpu%d_%s_%s", gpuId, metric, level)

	// Only process if this alert was active
	if !m.activeAlerts[key] {
		return
	}

	now := time.Now()

	// Find the alert in the list and mark as resolved
	for i := range m.alerts {
		alert := &m.alerts[i]
		if alert.GPUId == gpuId &&
		   alert.Metric == metric &&
		   alert.Level == level &&
		   !alert.Resolved &&
		   !alert.Acknowledged {
			alert.Resolved = true
			alert.ResolvedAt = now
			m.activeAlerts[key] = false
			break
		}
	}
}

// Log alert to file
func (m *model) logAlert(alert Alert) {
	f, err := os.OpenFile(alertHistoryLog, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return
	}
	defer f.Close()

	logLine := fmt.Sprintf("[%s] GPU %d - %s %s: %.1f (threshold: %.1f)\n",
		alert.Timestamp.Format("2006-01-02 15:04:05"),
		alert.GPUId,
		alert.Level,
		alert.Metric,
		alert.Value,
		alert.Threshold,
	)
	f.WriteString(logLine)
}

// Kill selected process
func (m *model) killSelectedProcess() {
	if m.selectedProcess >= len(m.processes) {
		return
	}

	proc := m.processes[m.selectedProcess]
	pidStr := getString(proc, "pid", "")
	if pidStr == "" {
		return
	}

	pid, err := strconv.Atoi(pidStr)
	if err != nil {
		return
	}

	// Kill process (platform-specific)
	killProcess(pid)
}

// Cycle sort order
func (m *model) cycleSortOrder() {
	switch m.processSort {
	case SortByMemory:
		m.processSort = SortByGPU
	case SortByGPU:
		m.processSort = SortByCPU
	case SortByCPU:
		m.processSort = SortByPID
	case SortByPID:
		m.processSort = SortByName
	case SortByName:
		m.processSort = SortByMemory
	}
}

// Get filtered and sorted processes
func (m *model) getFilteredProcesses() []map[string]interface{} {
	procs := m.processes

	// Apply filter
	if m.processFilter != "" {
		filtered := []map[string]interface{}{}
		for _, proc := range procs {
			name := strings.ToLower(getString(proc, "name", ""))
			if strings.Contains(name, strings.ToLower(m.processFilter)) {
				filtered = append(filtered, proc)
			}
		}
		procs = filtered
	}

	// Apply sort
	sort.Slice(procs, func(i, j int) bool {
		switch m.processSort {
		case SortByMemory:
			return getFloat(procs[i], "memory", 0) > getFloat(procs[j], "memory", 0)
		case SortByGPU:
			return getFloat(procs[i], "gpu_percent", 0) > getFloat(procs[j], "gpu_percent", 0)
		case SortByCPU:
			return getFloat(procs[i], "cpu_percent", 0) > getFloat(procs[j], "cpu_percent", 0)
		case SortByPID:
			return getString(procs[i], "pid", "0") < getString(procs[j], "pid", "0")
		case SortByName:
			return getString(procs[i], "name", "") < getString(procs[j], "name", "")
		}
		return false
	})

	return procs
}

// View renders the TUI
func (m model) View() string {
	if m.width == 0 {
		return "Loading..."
	}

	// If in alert view mode, show alert history
	if m.alertViewMode {
		return m.renderAlertHistoryView()
	}

	var sections []string

	// Title
	title := titleStyle.Render("GPU Pro - Terminal Monitor")
	if m.processMode {
		title += " " + lipgloss.NewStyle().Foreground(primaryColor).Render("[Process Management Mode]")
	}
	sections = append(sections, title)

	// Active alerts banner
	if len(m.activeAlerts) > 0 {
		alertBanner := m.renderAlertBanner()
		sections = append(sections, alertBanner)
	}

	// System Info
	sections = append(sections, m.renderSystemInfo())

	// GPUs
	if len(m.gpuData) == 0 {
		// Show a friendly message when no GPUs are detected
		noGPUMsg := lipgloss.NewStyle().
			Foreground(mutedColor).
			Italic(true).
			Render("⚠️  No NVIDIA GPUs detected or NVML not available\n✓  System metrics are still available")
		sections = append(sections, boxStyle.Render(noGPUMsg))
	} else {
		for i, gpu := range m.gpuData {
			sections = append(sections, m.renderGPU(i, gpu))
		}
	}

	// Processes
	if len(m.processes) > 0 {
		sections = append(sections, m.renderProcesses())
	}

	// Search input
	if m.searchMode {
		searchBox := lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(primaryColor).
			Padding(0, 1).
			Render(m.searchInput.View())
		sections = append(sections, searchBox)
	}

	// Help
	help := m.renderHelp()
	sections = append(sections, "\n"+help)

	return lipgloss.JoinVertical(lipgloss.Left, sections...)
}

// Render alert history view
func (m model) renderAlertHistoryView() string {
	var sections []string

	// Title
	title := titleStyle.Render("GPU Pro - Alert History")
	sections = append(sections, title)

	// Header info
	totalAlerts := len(m.alerts)
	activeCount := len(m.activeAlerts)

	infoLine := fmt.Sprintf("Total Alerts: %d | Active: %d | Press 'a' to return", totalAlerts, activeCount)
	sections = append(sections, lipgloss.NewStyle().Foreground(mutedColor).Render(infoLine))
	sections = append(sections, "")

	// Display alerts (most recent first)
	if len(m.alerts) == 0 {
		noAlerts := lipgloss.NewStyle().
			Foreground(mutedColor).
			Italic(true).
			Render("No alerts recorded yet")
		sections = append(sections, boxStyle.Render(noAlerts))
	} else {
		// Show last 20 alerts
		startIdx := 0
		if len(m.alerts) > 20 {
			startIdx = len(m.alerts) - 20
		}

		recentAlerts := m.alerts[startIdx:]

		// Reverse to show most recent first
		displayIdx := 0
		for i := len(recentAlerts) - 1; i >= 0; i-- {
			alert := recentAlerts[i]
			isSelected := (displayIdx == m.selectedAlert)
			sections = append(sections, m.renderAlertItem(alert, isSelected))
			displayIdx++
		}

		if len(m.alerts) > 20 {
			moreInfo := lipgloss.NewStyle().
				Foreground(mutedColor).
				Italic(true).
				Render(fmt.Sprintf("... and %d more (see %s for full history)", len(m.alerts)-20, alertHistoryLog))
			sections = append(sections, moreInfo)
		}
	}

	// Help
	help := helpStyle.Render("↑/↓ or j/k: Navigate | s: Snooze 5m | S: Snooze 30m | A: Acknowledge | a: Return | q: Quit")
	ttlInfo := helpStyle.Render("TTL: Resolved 30s | Acknowledged 1min | Active 1hr")
	sections = append(sections, "\n"+help+"\n"+ttlInfo)

	return lipgloss.JoinVertical(lipgloss.Left, sections...)
}

// Render individual alert item
func (m model) renderAlertItem(alert Alert, isSelected bool) string {
	// Style based on level
	var levelStyle lipgloss.Style
	var levelIcon string

	if alert.Level == "critical" {
		levelStyle = alertStyle
		levelIcon = "🔴"
	} else {
		levelStyle = warningStyle
		levelIcon = "🟡"
	}

	timestamp := alert.Timestamp.Format("15:04:05")
	level := strings.ToUpper(alert.Level)

	// Add status indicators
	statusBadges := ""
	if alert.Resolved {
		// Show resolved badge with TTL countdown (30 seconds)
		remaining := time.Until(alert.ResolvedAt.Add(30 * time.Second))
		secs := int(remaining.Seconds())
		if secs < 0 {
			secs = 0
		}
		statusBadges += lipgloss.NewStyle().
			Foreground(successColor).
			Render(fmt.Sprintf(" [✓ RESOLVED | TTL: %ds]", secs))
	} else if alert.Acknowledged {
		// Show acknowledged badge with TTL countdown (1 minute)
		remaining := time.Until(alert.Timestamp.Add(1 * time.Minute))
		secs := int(remaining.Seconds())
		if secs < 0 {
			secs = 0
		}
		statusBadges += lipgloss.NewStyle().
			Foreground(successColor).
			Render(fmt.Sprintf(" [✓ ACK | TTL: %ds]", secs))
	} else if alert.Snoozed {
		remaining := time.Until(alert.SnoozeUntil)
		mins := int(remaining.Minutes())
		statusBadges += lipgloss.NewStyle().
			Foreground(primaryColor).
			Render(fmt.Sprintf(" [💤 %dm]", mins))
	}

	line := fmt.Sprintf(
		"%s %s [%s] GPU %d - %s: %.1f%s (threshold: %.1f%s)%s",
		levelIcon,
		lipgloss.NewStyle().Foreground(mutedColor).Render(timestamp),
		levelStyle.Render(level),
		alert.GPUId,
		alert.Metric,
		alert.Value,
		getMetricUnit(alert.Metric),
		alert.Threshold,
		getMetricUnit(alert.Metric),
		statusBadges,
	)

	// Highlight selected alert
	if isSelected {
		return selectedStyle.Render("→ " + line)
	}
	return "  " + line
}

// Get metric unit
func getMetricUnit(metric string) string {
	switch metric {
	case "Temperature":
		return "°C"
	case "Memory", "Power":
		return "%"
	default:
		return ""
	}
}

// Render alert banner
func (m model) renderAlertBanner() string {
	alertCount := len(m.activeAlerts)
	criticalCount := 0
	warningCount := 0

	for _, alert := range m.alerts[max(0, len(m.alerts)-20):] {
		if alert.Level == "critical" {
			criticalCount++
		} else {
			warningCount++
		}
	}

	banner := fmt.Sprintf("🚨 %d ACTIVE ALERTS", alertCount)
	if criticalCount > 0 {
		banner += fmt.Sprintf(" | %d CRITICAL", criticalCount)
	}
	if warningCount > 0 {
		banner += fmt.Sprintf(" | %d WARNING", warningCount)
	}
	banner += " | Press 'a' to view history"

	return alertStyle.Render(banner)
}

// Render help text
func (m model) renderHelp() string {
	if m.processMode {
		return helpStyle.Render("Process Mode: ↑/↓ or j/k: Navigate | K: Kill | /: Search | c: Clear filter | s: Sort | p: Exit | q: Quit")
	}
	return helpStyle.Render("q: Quit | r: Refresh | p: Process Mode | a: Toggle Alert View | Updates every 0.5s")
}

// renderSystemInfo renders system resource information
func (m model) renderSystemInfo() string {
	cpuPercent := getFloat(m.systemInfo, "cpu_percent", 0)
	memPercent := getFloat(m.systemInfo, "memory_percent", 0)
	diskPercent := getFloat(m.systemInfo, "disk_percent", 0)
	diskUsed := getFloat(m.systemInfo, "disk_used_gb", 0)
	diskTotal := getFloat(m.systemInfo, "disk_total_gb", 0)
	diskFree := getFloat(m.systemInfo, "disk_free_gb", 0)

	header := headerStyle.Render("System Resources")

	cpuBar := m.renderBar("CPU", cpuPercent, 100, "%")
	memBar := m.renderBar("Memory", memPercent, 100, "%")
	diskBar := m.renderBar("Disk", diskPercent, 100, "%")

	// Add detailed disk info
	diskInfo := fmt.Sprintf(
		"%s %s | %s %s | %s %s",
		labelStyle.Render("Used:"),
		valueStyle.Render(fmt.Sprintf("%.1f GB", diskUsed)),
		labelStyle.Render("Free:"),
		valueStyle.Render(fmt.Sprintf("%.1f GB", diskFree)),
		labelStyle.Render("Total:"),
		valueStyle.Render(fmt.Sprintf("%.1f GB", diskTotal)),
	)

	content := lipgloss.JoinVertical(lipgloss.Left,
		header,
		cpuBar,
		memBar,
		diskBar,
		diskInfo,
	)

	return boxStyle.Render(content)
}

// renderGPU renders a single GPU's information with sparklines
func (m model) renderGPU(id int, gpu map[string]interface{}) string {
	name := getString(gpu, "name", "Unknown GPU")
	util := getFloat(gpu, "utilization", 0)
	temp := getFloat(gpu, "temperature", 0)
	power := getFloat(gpu, "power_draw", 0)
	powerLimit := getFloat(gpu, "power_limit", 1)
	memUsed := getFloat(gpu, "memory_used", 0)
	memTotal := getFloat(gpu, "memory_total", 1)
	fanSpeed := getFloat(gpu, "fan_speed", 0)

	memPercent := (memUsed / memTotal) * 100
	powerPercent := (power / powerLimit) * 100

	header := headerStyle.Render(fmt.Sprintf("GPU %d: %s", id, name))

	// Get MFU and additional metrics
	mfu := getFloat(gpu, "mfu", 0)
	peakTFLOPs := getFloat(gpu, "peak_tflops", 0)
	achievedTFLOPs := getFloat(gpu, "achieved_tflops", 0)

	// Get sparklines
	hist := m.gpuHistory[id]
	var utilSparkline, tempSparkline, memSparkline, powerSparkline, mfuSparkline string
	if hist != nil {
		utilSparkline = renderSparkline(hist.Utilization)
		tempSparkline = renderSparkline(hist.Temperature)
		memSparkline = renderSparkline(hist.Memory)
		powerSparkline = renderSparkline(hist.Power)
		mfuSparkline = renderSparkline(hist.MFU)
	}

	// Metrics with sparklines and trend indicators
	utilBar := m.renderBarWithSparkline("Utilization", util, 100, "%", utilSparkline, getTrendIndicator(hist.Utilization))
	tempBar := m.renderBarWithSparkline("Temperature", temp, 100, "°C", tempSparkline, getTrendIndicator(hist.Temperature))
	memBar := m.renderBarWithSparkline("Memory", memPercent, 100, "%", memSparkline, getTrendIndicator(hist.Memory))
	powerBar := m.renderBarWithSparkline("Power", powerPercent, 100, "%", powerSparkline, getTrendIndicator(hist.Power))

	// MFU bar (only show if peak TFLOPs is known, otherwise show as info line)
	var mfuBar string
	if peakTFLOPs > 0 {
		mfuBar = m.renderBarWithSparkline("MFU", mfu, 100, "%", mfuSparkline, getTrendIndicator(hist.MFU))
	} else {
		mfuBar = labelStyle.Render("MFU:") + " " +
			lipgloss.NewStyle().Foreground(mutedColor).Render("N/A (GPU model not in database)")
	}

	// Additional info
	var info string
	if peakTFLOPs > 0 {
		info = fmt.Sprintf(
			"%s %s | %s %s | %s %.1f W / %.1f W | %s %.1f / %.1f TFLOPs",
			labelStyle.Render("Fan:"),
			valueStyle.Render(fmt.Sprintf("%.0f%%", fanSpeed)),
			labelStyle.Render("Memory:"),
			valueStyle.Render(fmt.Sprintf("%.0f / %.0f MiB", memUsed, memTotal)),
			labelStyle.Render("Power:"),
			power, powerLimit,
			labelStyle.Render("Compute:"),
			achievedTFLOPs, peakTFLOPs,
		)
	} else {
		info = fmt.Sprintf(
			"%s %s | %s %s | %s %.1f W / %.1f W",
			labelStyle.Render("Fan:"),
			valueStyle.Render(fmt.Sprintf("%.0f%%", fanSpeed)),
			labelStyle.Render("Memory:"),
			valueStyle.Render(fmt.Sprintf("%.0f / %.0f MiB", memUsed, memTotal)),
			labelStyle.Render("Power:"),
			power, powerLimit,
		)
	}

	content := lipgloss.JoinVertical(lipgloss.Left,
		header,
		utilBar,
		tempBar,
		memBar,
		powerBar,
		mfuBar,
		info,
	)

	return boxStyle.Render(content)
}

// Render sparkline with finer granularity
func renderSparkline(data []float64) string {
	if len(data) == 0 {
		return ""
	}

	// Use only the most recent sparklineLength data points
	startIdx := 0
	if len(data) > sparklineLength {
		startIdx = len(data) - sparklineLength
	}
	data = data[startIdx:]

	// Extended sparkline characters for finer granularity (9 levels)
	// Using block characters for smooth gradation
	chars := []rune{'▁', '▂', '▃', '▄', '▅', '▆', '▇', '█', '█'}

	// Find min and max
	min, max := data[0], data[0]
	for _, v := range data {
		if v < min {
			min = v
		}
		if v > max {
			max = v
		}
	}

	// Generate sparkline with finer steps
	var sparkline strings.Builder
	for _, v := range data {
		if max == min {
			// If all values are the same, show mid-level
			sparkline.WriteRune('▄')
		} else {
			normalized := (v - min) / (max - min)
			idx := int(normalized * float64(len(chars)-1))
			if idx >= len(chars) {
				idx = len(chars) - 1
			}
			if idx < 0 {
				idx = 0
			}
			sparkline.WriteRune(chars[idx])
		}
	}

	return sparkline.String()
}

// Get trend indicator
func getTrendIndicator(data []float64) string {
	if len(data) < 5 {
		return ""
	}

	// Compare recent average with older average
	recentAvg := average(data[len(data)-5:])
	olderAvg := average(data[max(0, len(data)-15):len(data)-5])

	diff := recentAvg - olderAvg

	if diff > 5 {
		return "🔥" // Rising fast
	} else if diff > 2 {
		return "↗" // Rising
	} else if diff < -5 {
		return "❄️" // Falling fast
	} else if diff < -2 {
		return "↘" // Falling
	}
	return "→" // Stable
}

// Calculate average
func average(data []float64) float64 {
	if len(data) == 0 {
		return 0
	}
	sum := 0.0
	for _, v := range data {
		sum += v
	}
	return sum / float64(len(data))
}

// renderBarWithSparkline renders a progress bar with sparkline and trend
func (m model) renderBarWithSparkline(label string, value, max float64, unit, sparkline, trend string) string {
	percent := value / max
	if percent > 1.0 {
		percent = 1.0
	}

	// Color based on value and thresholds
	color := successColor
	style := valueStyle

	// Check if this metric has an alert
	if label == "Temperature" {
		if value >= m.thresholds.TempCritical {
			color = dangerColor
			style = alertStyle
		} else if value >= m.thresholds.TempWarning {
			color = warningColor
			style = warningStyle
		}
	} else if label == "Memory" {
		if value >= m.thresholds.MemoryCritical {
			color = dangerColor
			style = alertStyle
		} else if value >= m.thresholds.MemoryWarning {
			color = warningColor
			style = warningStyle
		}
	} else {
		// Default color scheme
		if percent > 0.9 {
			color = dangerColor
		} else if percent > 0.7 {
			color = warningColor
		}
	}

	bar := m.progress.ViewAs(percent)

	// Add color
	coloredBar := lipgloss.NewStyle().
		Foreground(color).
		Render(bar)

	valueStr := fmt.Sprintf("%.1f%s", value, unit)

	// Add sparkline and trend if available
	if sparkline != "" {
		sparklineStr := lipgloss.NewStyle().Foreground(mutedColor).Render(sparkline)
		return fmt.Sprintf(
			"%s %s %s %s %s",
			labelStyle.Render(label+":"),
			coloredBar,
			style.Render(valueStr),
			sparklineStr,
			trend,
		)
	}

	return fmt.Sprintf(
		"%s %s %s",
		labelStyle.Render(label+":"),
		coloredBar,
		style.Render(valueStr),
	)
}

// renderProcesses renders GPU processes with interactive features
func (m model) renderProcesses() string {
	filteredProcs := m.getFilteredProcesses()

	headerText := fmt.Sprintf("Active Processes (%d)", len(filteredProcs))
	if m.processFilter != "" {
		headerText += fmt.Sprintf(" [Filter: %s]", m.processFilter)
	}
	headerText += fmt.Sprintf(" [Sort: %s]", m.processSort)

	header := headerStyle.Render(headerText)

	var procLines []string
	displayLimit := 10
	if m.processMode {
		displayLimit = 20
	}

	for i, proc := range filteredProcs {
		if i >= displayLimit {
			moreText := lipgloss.NewStyle().Foreground(mutedColor).Render(fmt.Sprintf("... and %d more", len(filteredProcs)-displayLimit))
			procLines = append(procLines, moreText)
			break
		}

		name := getString(proc, "name", "unknown")
		pid := getString(proc, "pid", "0")
		memory := getFloat(proc, "memory", 0)
		gpuPercent := getFloat(proc, "gpu_percent", 0)
		cpuPercent := getFloat(proc, "cpu_percent", 0)

		line := fmt.Sprintf(
			"%s %s | %s %s | %s %.1f MiB | %s %.1f%% | %s %.1f%%",
			labelStyle.Width(12).Render("Process:"),
			valueStyle.Render(truncate(name, 20)),
			labelStyle.Width(6).Render("PID:"),
			valueStyle.Render(pid),
			labelStyle.Width(8).Render("VRAM:"),
			memory,
			labelStyle.Width(6).Render("GPU:"),
			gpuPercent,
			labelStyle.Width(6).Render("CPU:"),
			cpuPercent,
		)

		// Highlight selected process in process mode
		if m.processMode && i == m.selectedProcess {
			line = selectedStyle.Render("→ " + line)
		} else {
			line = "  " + line
		}

		procLines = append(procLines, line)
	}

	content := lipgloss.JoinVertical(lipgloss.Left,
		header,
		strings.Join(procLines, "\n"),
	)

	return boxStyle.Render(content)
}

// renderBar renders a progress bar with label and value
func (m model) renderBar(label string, value, max float64, unit string) string {
	percent := value / max
	if percent > 1.0 {
		percent = 1.0
	}

	// Color based on value
	color := successColor
	if percent > 0.9 {
		color = dangerColor
	} else if percent > 0.7 {
		color = warningColor
	}

	bar := m.progress.ViewAs(percent)

	// Add color
	coloredBar := lipgloss.NewStyle().
		Foreground(color).
		Render(bar)

	valueStr := fmt.Sprintf("%.1f%s", value, unit)

	return fmt.Sprintf(
		"%s %s %s",
		labelStyle.Render(label+":"),
		coloredBar,
		valueStyle.Render(valueStr),
	)
}

// Helper functions
func getFloat(m map[string]interface{}, key string, def float64) float64 {
	if v, ok := m[key]; ok {
		switch val := v.(type) {
		case float64:
			return val
		case int:
			return float64(val)
		}
	}
	return def
}

func getString(m map[string]interface{}, key string, def string) string {
	if v, ok := m[key]; ok {
		if val, ok := v.(string); ok {
			return val
		}
	}
	return def
}

func truncate(s string, max int) string {
	if len(s) <= max {
		return s
	}
	return s[:max-3] + "..."
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func main() {
	// Check for command line flags
	if len(os.Args) > 1 {
		switch os.Args[1] {
		case "--view-alerts":
			viewAlertHistory()
			return
		case "--config-thresholds":
			configureThresholds()
			return
		case "--debug-mfu":
			debugMFU()
			return
		case "--help":
			printHelp()
			return
		}
	}

	// Run TUI
	p := tea.NewProgram(
		initialModel(),
		tea.WithAltScreen(),
	)

	if _, err := p.Run(); err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}
}

// View alert history
func viewAlertHistory() {
	data, err := os.ReadFile(alertHistoryLog)
	if err != nil {
		fmt.Println("No alert history found")
		return
	}

	// Use less or more to display
	cmd := exec.Command("less")
	cmd.Stdin = strings.NewReader(string(data))
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Run()
}

// Configure thresholds interactively
func configureThresholds() {
	t := loadThresholds()

	fmt.Println("Current Thresholds:")
	fmt.Printf("Temperature Warning: %.1f°C\n", t.TempWarning)
	fmt.Printf("Temperature Critical: %.1f°C\n", t.TempCritical)
	fmt.Printf("Memory Warning: %.1f%%\n", t.MemoryWarning)
	fmt.Printf("Memory Critical: %.1f%%\n", t.MemoryCritical)
	fmt.Printf("Power Warning: %.1f%%\n", t.PowerWarning)
	fmt.Printf("Power Critical: %.1f%%\n", t.PowerCritical)
	fmt.Println("\nEdit gpu-thresholds.json to modify")
}

// Debug MFU calculation
func debugMFU() {
	fmt.Println("MFU Debug Information")
	fmt.Println("====================\n")

	mon := monitor.NewGPUMonitor()
	if mon == nil {
		fmt.Println("Error: Could not initialize GPU monitor")
		return
	}
	defer mon.Shutdown()

	// Wait a moment for data collection
	time.Sleep(1 * time.Second)

	gpuData, err := mon.GetGPUData()
	if err != nil {
		fmt.Printf("Error getting GPU data: %v\n", err)
		return
	}

	if len(gpuData) == 0 {
		fmt.Println("No GPUs detected")
		return
	}

	for gpuID, gpuInfo := range gpuData {
		if data, ok := gpuInfo.(map[string]interface{}); ok {
			fmt.Printf("GPU %s:\n", gpuID)
			fmt.Printf("  Name: %s\n", getString(data, "name", "Unknown"))
			fmt.Printf("  Architecture: %s\n", getString(data, "architecture", "Unknown"))
			fmt.Println()

			fmt.Println("  MFU Metrics:")
			fmt.Printf("    MFU: %.2f%%\n", getFloat(data, "mfu", 0))
			fmt.Printf("    Peak TFLOPs: %.2f\n", getFloat(data, "peak_tflops", 0))
			fmt.Printf("    Achieved TFLOPs: %.2f\n", getFloat(data, "achieved_tflops", 0))
			fmt.Println()

			fmt.Println("  Debug Info:")
			fmt.Printf("    GPU Name (uppercase): %s\n", getString(data, "mfu_debug_gpu_name", "N/A"))
			fmt.Printf("    SM Clock: %.0f MHz\n", getFloat(data, "mfu_debug_sm_clock", 0))
			fmt.Printf("    Max SM Clock: %.0f MHz\n", getFloat(data, "mfu_debug_max_sm_clock", 0))
			fmt.Printf("    Utilization: %.2f%%\n", getFloat(data, "mfu_debug_utilization", 0))
			if status := getString(data, "mfu_debug_status", ""); status != "" {
				fmt.Printf("    Status: %s\n", status)
			}
			fmt.Println()

			if getFloat(data, "peak_tflops", 0) == 0 {
				fmt.Println("  ⚠️  This GPU model is not in the MFU database.")
				fmt.Println("      MFU calculation requires peak TFLOPs specification.")
				fmt.Println("      You can add support by editing monitor/metrics_linux.go")
				fmt.Printf("      and adding '%s' to the getPeakTFLOPs() function.\n", getString(data, "name", "Unknown"))
			}
			fmt.Println()
		}
	}
}

// Print help
func printHelp() {
	fmt.Println("GPU Pro CLI - Advanced GPU Monitoring")
	fmt.Println("\nUsage:")
	fmt.Println("  gpu-pro-cli                    Start interactive monitoring")
	fmt.Println("  gpu-pro-cli --view-alerts      View alert history")
	fmt.Println("  gpu-pro-cli --config-thresholds View current threshold configuration")
	fmt.Println("  gpu-pro-cli --debug-mfu        Show MFU debug information")
	fmt.Println("  gpu-pro-cli --help             Show this help")
	fmt.Println("\nInteractive Mode Controls:")
	fmt.Println("  q, Ctrl+C    Quit")
	fmt.Println("  r            Refresh data")
	fmt.Println("  p            Toggle process management mode")
	fmt.Println("  a            Toggle alert history view (interactive)")
	fmt.Println("\nAlert History View:")
	fmt.Println("  ↑/↓, j/k     Navigate alerts")
	fmt.Println("  s            Snooze selected alert for 5 minutes")
	fmt.Println("  S            Snooze selected alert for 30 minutes (capital S)")
	fmt.Println("  A            Acknowledge selected alert (capital A)")
	fmt.Println("  a            Return to monitoring")
	fmt.Println("\nProcess Management Mode:")
	fmt.Println("  ↑/↓, j/k     Navigate processes")
	fmt.Println("  K            Kill selected process (capital K)")
	fmt.Println("  /            Search processes")
	fmt.Println("  c            Clear search filter")
	fmt.Println("  s            Cycle sort order (memory, GPU, CPU, PID, name)")
	fmt.Println("\nFeatures:")
	fmt.Println("  • Historical sparklines showing 10-second trends")
	fmt.Println("  • MFU (Model FLOPs Utilization) calculation for supported GPUs")
	fmt.Println("  • Interactive alert management (snooze, acknowledge)")
	fmt.Println("  • Smart alert expiration (30s resolved, 1min acknowledged, 1h active)")
	fmt.Println("  • Automatic threshold alerts with persistent logging")
	fmt.Println("  • Interactive process management with filtering and sorting")
	fmt.Println("  • Configurable alert thresholds (edit gpu-thresholds.json)")
}

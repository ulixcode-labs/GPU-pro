package handlers

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"os/user"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"time"
)

// NetworkStats holds network I/O statistics
type NetworkStats struct {
	Interface     string  `json:"interface"`
	BytesReceived uint64  `json:"bytes_received"`
	BytesSent     uint64  `json:"bytes_sent"`
	RxRate        float64 `json:"rx_rate"` // bytes/sec
	TxRate        float64 `json:"tx_rate"` // bytes/sec
}

// DiskStats holds disk I/O statistics
type DiskStats struct {
	Device          string  `json:"device"`
	ReadsCompleted  uint64  `json:"reads_completed"`
	WritesCompleted uint64  `json:"writes_completed"`
	SectorsRead     uint64  `json:"-"` // Not exported to JSON
	SectorsWritten  uint64  `json:"-"` // Not exported to JSON
	ReadRate        float64 `json:"read_rate"`  // reads/sec
	WriteRate       float64 `json:"write_rate"` // writes/sec
	ReadKBps        float64 `json:"read_kbps"`  // KB/s
	WriteKBps       float64 `json:"write_kbps"` // KB/s
}

// NetworkConnection represents a network connection
type NetworkConnection struct {
	Protocol    string `json:"protocol"`
	LocalAddr   string `json:"local_addr"`
	ForeignAddr string `json:"foreign_addr"`
	State       string `json:"state"`
	PID         string `json:"pid"`
	Program     string `json:"program"`
	IsExternal  bool   `json:"is_external"`
	ForeignIP   string `json:"foreign_ip"`
	Duration    string `json:"duration"`     // Human-readable duration
	DurationSec int64  `json:"duration_sec"` // Duration in seconds for sorting
}

// LargeFile represents a large file
type LargeFile struct {
	Path           string `json:"path"`
	Size           int64  `json:"size"`              // Apparent size
	SizeHuman      string `json:"size_human"`        // Apparent size (human readable)
	ActualSize     int64  `json:"actual_size"`       // Actual disk usage
	ActualSizeHuman string `json:"actual_size_human"` // Actual disk usage (human readable)
	IsSparse       bool   `json:"is_sparse"`         // Whether file is sparse
	ModTime        string `json:"mod_time"`
}

var (
	lastNetStats          map[string]*NetworkStats
	lastDiskStats         map[string]*DiskStats
	lastNetStatsTime      time.Time
	lastDiskStatsTime     time.Time
	connectionFirstSeen   map[string]time.Time // Track when connections were first seen
	connectionFirstSeenMu sync.RWMutex
)

func init() {
	lastNetStats = make(map[string]*NetworkStats)
	lastDiskStats = make(map[string]*DiskStats)
	connectionFirstSeen = make(map[string]time.Time)
	lastNetStatsTime = time.Now()
	lastDiskStatsTime = time.Now()
}

// GetNetworkIO reads network I/O statistics from /proc/net/dev
func GetNetworkIO() []NetworkStats {
	stats := []NetworkStats{}
	
	file, err := os.Open("/proc/net/dev")
	if err != nil {
		return stats
	}
	defer file.Close()
	
	scanner := bufio.NewScanner(file)
	// Skip first two header lines
	scanner.Scan()
	scanner.Scan()
	
	now := time.Now()
	elapsed := now.Sub(lastNetStatsTime).Seconds()
	if elapsed == 0 {
		elapsed = 1
	}

	for scanner.Scan() {
		line := scanner.Text()
		fields := strings.Fields(line)
		if len(fields) < 10 {
			continue
		}

		iface := strings.TrimSuffix(fields[0], ":")
		// Skip loopback
		if iface == "lo" {
			continue
		}

		bytesReceived, _ := strconv.ParseUint(fields[1], 10, 64)
		bytesSent, _ := strconv.ParseUint(fields[9], 10, 64)

		stat := NetworkStats{
			Interface:     iface,
			BytesReceived: bytesReceived,
			BytesSent:     bytesSent,
		}

		// Calculate rates
		if last, ok := lastNetStats[iface]; ok {
			stat.RxRate = float64(bytesReceived-last.BytesReceived) / elapsed
			stat.TxRate = float64(bytesSent-last.BytesSent) / elapsed
		}

		stats = append(stats, stat)
		lastNetStats[iface] = &stat
	}

	lastNetStatsTime = now
	return stats
}

// GetDiskIO reads disk I/O statistics from /proc/diskstats
func GetDiskIO() []DiskStats {
	stats := []DiskStats{}
	
	file, err := os.Open("/proc/diskstats")
	if err != nil {
		return stats
	}
	defer file.Close()
	
	now := time.Now()
	elapsed := now.Sub(lastDiskStatsTime).Seconds()
	if elapsed == 0 {
		elapsed = 1
	}

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		fields := strings.Fields(scanner.Text())
		if len(fields) < 14 {
			continue
		}
		
		device := fields[2]
		// Only show main devices (sda, nvme0n1, etc.), skip partitions
		if strings.Contains(device, "loop") {
			continue
		}
		if len(device) > 0 && (device[len(device)-1] >= '0' && device[len(device)-1] <= '9') {
			// Skip if it's a partition (ends with number and parent exists)
			if !strings.HasPrefix(device, "nvme") && !strings.HasPrefix(device, "mmcblk") {
				continue
			}
		}
		
		readsCompleted, _ := strconv.ParseUint(fields[3], 10, 64)
		sectorsRead, _ := strconv.ParseUint(fields[5], 10, 64)
		writesCompleted, _ := strconv.ParseUint(fields[7], 10, 64)
		sectorsWritten, _ := strconv.ParseUint(fields[9], 10, 64)
		
		stat := DiskStats{
			Device:          device,
			ReadsCompleted:  readsCompleted,
			WritesCompleted: writesCompleted,
			SectorsRead:     sectorsRead,
			SectorsWritten:  sectorsWritten,
		}

		// Calculate rates (sectors are 512 bytes)
		if last, ok := lastDiskStats[device]; ok {
			stat.ReadRate = float64(readsCompleted-last.ReadsCompleted) / elapsed
			stat.WriteRate = float64(writesCompleted-last.WritesCompleted) / elapsed
			stat.ReadKBps = float64(sectorsRead-last.SectorsRead) * 512 / 1024 / elapsed
			stat.WriteKBps = float64(sectorsWritten-last.SectorsWritten) * 512 / 1024 / elapsed
		}
		
		stats = append(stats, stat)
		lastDiskStats[device] = &stat
	}

	// Update the last stats time for next calculation
	lastDiskStatsTime = now

	return stats
}

// ConnectionStats holds connection count statistics by protocol
type ConnectionStats struct {
	TCP   int `json:"tcp"`
	UDP   int `json:"udp"`
	Other int `json:"other"`
	Total int `json:"total"`
}

// getConnectionKey generates a unique key for a connection
func getConnectionKey(protocol, localAddr, foreignAddr, pid string) string {
	return fmt.Sprintf("%s|%s|%s|%s", protocol, localAddr, foreignAddr, pid)
}

// formatDuration formats a duration into human-readable string
func formatDuration(d time.Duration) string {
	seconds := int64(d.Seconds())

	if seconds < 60 {
		return fmt.Sprintf("%ds", seconds)
	}

	minutes := seconds / 60
	if minutes < 60 {
		return fmt.Sprintf("%dm %ds", minutes, seconds%60)
	}

	hours := minutes / 60
	if hours < 24 {
		return fmt.Sprintf("%dh %dm", hours, minutes%60)
	}

	days := hours / 24
	return fmt.Sprintf("%dd %dh", days, hours%24)
}

// trackConnectionDuration tracks and returns duration for a connection
func trackConnectionDuration(connKey string) (string, int64) {
	now := time.Now()

	connectionFirstSeenMu.Lock()
	firstSeen, exists := connectionFirstSeen[connKey]
	if !exists {
		connectionFirstSeen[connKey] = now
		firstSeen = now
	}
	connectionFirstSeenMu.Unlock()

	duration := now.Sub(firstSeen)
	durationSec := int64(duration.Seconds())
	durationStr := formatDuration(duration)

	return durationStr, durationSec
}

// cleanupStaleConnections removes connections that are no longer active
func cleanupStaleConnections(activeKeys map[string]bool) {
	connectionFirstSeenMu.Lock()
	defer connectionFirstSeenMu.Unlock()

	for key := range connectionFirstSeen {
		if !activeKeys[key] {
			delete(connectionFirstSeen, key)
		}
	}
}

// GetNetworkConnections gets network connections using netstat
func GetNetworkConnections() ([]NetworkConnection, ConnectionStats) {
	connections := []NetworkConnection{}
	stats := ConnectionStats{}
	activeKeys := make(map[string]bool) // Track active connections for cleanup

	// Use netstat for consistent output format
	// Format: Proto Recv-Q Send-Q Local-Address Foreign-Address State PID/Program
	cmd := exec.Command("netstat", "-tunap")
	output, err := cmd.Output()
	if err != nil {
		// If netstat fails, try ss with proper parsing
		return getConnectionsWithSS()
	}

	lines := strings.Split(string(output), "\n")
	for i, line := range lines {
		// Skip header lines
		if i == 0 || strings.TrimSpace(line) == "" {
			continue
		}

		fields := strings.Fields(line)
		if len(fields) < 5 {
			continue
		}

		// Skip header line (contains "Proto")
		if fields[0] == "Proto" {
			continue
		}

		protocol := strings.ToLower(fields[0])
		foreignAddr := fields[4]
		foreignIP := ExtractIP(foreignAddr)

		conn := NetworkConnection{
			Protocol:    fields[0],
			LocalAddr:   fields[3],
			ForeignAddr: foreignAddr,
			ForeignIP:   foreignIP,
			IsExternal:  !IsPrivateIP(foreignIP) && foreignIP != "" && foreignIP != "*" && foreignIP != "0.0.0.0",
		}

		// State field (position 5 for TCP, may not exist for UDP)
		if len(fields) > 5 && fields[5] != "-" {
			conn.State = fields[5]
		}

		// Try to extract PID and program name (last field)
		if len(fields) > 6 {
			pidProgram := fields[len(fields)-1]
			if strings.Contains(pidProgram, "/") {
				parts := strings.SplitN(pidProgram, "/", 2)
				conn.PID = parts[0]
				if len(parts) > 1 {
					conn.Program = parts[1]
				}
			}
		}

		// Track connection duration
		connKey := getConnectionKey(conn.Protocol, conn.LocalAddr, conn.ForeignAddr, conn.PID)
		conn.Duration, conn.DurationSec = trackConnectionDuration(connKey)
		activeKeys[connKey] = true

		// Count by protocol
		if strings.HasPrefix(protocol, "tcp") {
			stats.TCP++
		} else if strings.HasPrefix(protocol, "udp") {
			stats.UDP++
		} else {
			stats.Other++
		}
		stats.Total++

		connections = append(connections, conn)

		// Limit to 100 connections to avoid overwhelming the UI
		if len(connections) >= 100 {
			break
		}
	}

	// Cleanup stale connections
	cleanupStaleConnections(activeKeys)

	return connections, stats
}

// getConnectionsWithSS uses ss command (alternative parser)
func getConnectionsWithSS() ([]NetworkConnection, ConnectionStats) {
	connections := []NetworkConnection{}
	stats := ConnectionStats{}
	activeKeys := make(map[string]bool) // Track active connections for cleanup

	// ss format: Netid State Recv-Q Send-Q Local-Address:Port Peer-Address:Port Process
	cmd := exec.Command("ss", "-tunap")
	output, err := cmd.Output()
	if err != nil {
		return connections, stats
	}

	lines := strings.Split(string(output), "\n")
	for i, line := range lines {
		// Skip header line
		if i == 0 || strings.TrimSpace(line) == "" {
			continue
		}

		fields := strings.Fields(line)
		if len(fields) < 5 {
			continue
		}

		// Skip header if it appears
		if fields[0] == "Netid" {
			continue
		}

		protocol := strings.ToLower(fields[0])
		foreignAddr := fields[5]
		foreignIP := ExtractIP(foreignAddr)

		conn := NetworkConnection{
			Protocol:    fields[0],
			State:       fields[1],
			LocalAddr:   fields[4],
			ForeignAddr: foreignAddr,
			ForeignIP:   foreignIP,
			IsExternal:  !IsPrivateIP(foreignIP) && foreignIP != "" && foreignIP != "*" && foreignIP != "0.0.0.0",
		}

		// Extract PID and program from Process field (last field, format: users:(("program",pid=1234,fd=3)))
		if len(fields) > 6 {
			processField := fields[len(fields)-1]
			// Parse users:(("sshd",pid=1234,fd=3))
			if strings.Contains(processField, "pid=") {
				// Extract program name
				if start := strings.Index(processField, "((\""); start >= 0 {
					start += 3
					if end := strings.Index(processField[start:], "\""); end >= 0 {
						conn.Program = processField[start : start+end]
					}
				}
				// Extract PID
				if start := strings.Index(processField, "pid="); start >= 0 {
					start += 4
					pidStr := ""
					for j := start; j < len(processField) && processField[j] >= '0' && processField[j] <= '9'; j++ {
						pidStr += string(processField[j])
					}
					conn.PID = pidStr
				}
			}
		}

		// Track connection duration
		connKey := getConnectionKey(conn.Protocol, conn.LocalAddr, conn.ForeignAddr, conn.PID)
		conn.Duration, conn.DurationSec = trackConnectionDuration(connKey)
		activeKeys[connKey] = true

		// Count by protocol
		if strings.HasPrefix(protocol, "tcp") {
			stats.TCP++
		} else if strings.HasPrefix(protocol, "udp") {
			stats.UDP++
		} else {
			stats.Other++
		}
		stats.Total++

		connections = append(connections, conn)

		// Limit to 100 connections
		if len(connections) >= 100 {
			break
		}
	}

	// Cleanup stale connections
	cleanupStaleConnections(activeKeys)

	return connections, stats
}

// GetTopLargestFiles finds the top N largest files in a specified directory
func GetTopLargestFiles(n int, directory string) []LargeFile {
	files := []LargeFile{}

	// If no directory specified, use user's home directory
	if directory == "" {
		usr, err := user.Current()
		if err != nil {
			return files
		}
		directory = usr.HomeDir
	}

	// Collect all files with their sizes
	err := filepath.Walk(directory, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil // Skip files we can't access
		}

		// Skip directories and hidden files/directories
		if info.IsDir() || strings.HasPrefix(filepath.Base(path), ".") {
			if info.IsDir() && strings.HasPrefix(filepath.Base(path), ".") {
				return filepath.SkipDir
			}
			return nil
		}

		// Get apparent size
		apparentSize := info.Size()

		// Get actual disk usage using syscall
		var actualSize int64
		var isSparse bool

		if stat, ok := info.Sys().(*syscall.Stat_t); ok {
			// Blocks * 512 = actual bytes on disk
			actualSize = stat.Blocks * 512
			// File is sparse if apparent size > actual disk usage
			isSparse = apparentSize > actualSize
		} else {
			// Fallback: use apparent size if syscall fails
			actualSize = apparentSize
			isSparse = false
		}

		files = append(files, LargeFile{
			Path:           path,
			Size:           apparentSize,
			SizeHuman:      formatBytes(apparentSize),
			ActualSize:     actualSize,
			ActualSizeHuman: formatBytes(actualSize),
			IsSparse:       isSparse,
			ModTime:        info.ModTime().Format("2006-01-02 15:04:05"),
		})

		return nil
	})
	
	if err != nil {
		return []LargeFile{}
	}
	
	// Sort by actual disk usage descending (not apparent size)
	// This gives a more accurate view of what's consuming disk space
	sort.Slice(files, func(i, j int) bool {
		return files[i].ActualSize > files[j].ActualSize
	})
	
	// Return top N
	if len(files) > n {
		files = files[:n]
	}
	
	return files
}

// formatBytes converts bytes to human-readable format
func formatBytes(bytes int64) string {
	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%d B", bytes)
	}
	div, exp := int64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(bytes)/float64(div), "KMGTPE"[exp])
}

// GetOpenFileCount gets the number of open file descriptors system-wide
func GetOpenFileCount() int {
	// Read /proc/sys/fs/file-nr which contains:
	// allocated_handles free_handles max_handles
	data, err := os.ReadFile("/proc/sys/fs/file-nr")
	if err != nil {
		return 0
	}

	fields := strings.Fields(string(data))
	if len(fields) < 1 {
		return 0
	}

	allocated, err := strconv.Atoi(fields[0])
	if err != nil {
		return 0
	}

	return allocated
}

// GetSystemMetrics collects all system metrics
func GetSystemMetrics() map[string]interface{} {
	connections, connStats := GetNetworkConnections()

	// Get geolocation for external IPs (async to avoid blocking)
	geoLocations := GetConnectionGeoLocations(connections)

	return map[string]interface{}{
		"network_io":       GetNetworkIO(),
		"disk_io":          GetDiskIO(),
		"connections":      connections,
		"connection_stats": connStats,
		"geo_locations":    geoLocations,
		"open_files":       GetOpenFileCount(),
		"largest_files":    GetTopLargestFiles(10, "/"),
		"timestamp":        time.Now().Format(time.RFC3339),
	}
}

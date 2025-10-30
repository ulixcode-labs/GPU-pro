package handlers

import (
	"runtime"
	"testing"
	"time"

	"github.com/shirou/gopsutil/v3/cpu"
	"github.com/shirou/gopsutil/v3/disk"
	"github.com/shirou/gopsutil/v3/mem"
)

func TestCPUPercent(t *testing.T) {
	t.Log("Testing CPU metrics on", runtime.GOOS)

	// Test with 0 interval (may not work on first call)
	cpuPercent0, err := cpu.Percent(0, false)
	if err != nil {
		t.Errorf("cpu.Percent(0) failed: %v", err)
	}
	t.Logf("CPU with 0 interval: %v", cpuPercent0)

	// Test with 100ms interval (should always work)
	cpuPercent100, err := cpu.Percent(100*time.Millisecond, false)
	if err != nil {
		t.Errorf("cpu.Percent(100ms) failed: %v", err)
	}
	t.Logf("CPU with 100ms interval: %v", cpuPercent100)

	if len(cpuPercent100) == 0 {
		t.Error("Expected at least one CPU percentage value")
	}

	if len(cpuPercent100) > 0 {
		if cpuPercent100[0] < 0 || cpuPercent100[0] > 100 {
			t.Errorf("CPU percentage out of range: %f", cpuPercent100[0])
		}
		t.Logf("✓ CPU: %.2f%%", cpuPercent100[0])
	}

	// Test with 500ms interval
	cpuPercent500, err := cpu.Percent(500*time.Millisecond, false)
	if err != nil {
		t.Errorf("cpu.Percent(500ms) failed: %v", err)
	}
	t.Logf("CPU with 500ms interval: %v", cpuPercent500)

	if len(cpuPercent500) > 0 && cpuPercent500[0] == 0 {
		t.Error("CPU percentage should not be 0 with 500ms interval")
	}
}

func TestMemoryInfo(t *testing.T) {
	t.Log("Testing Memory metrics on", runtime.GOOS)

	memInfo, err := mem.VirtualMemory()
	if err != nil {
		t.Fatalf("mem.VirtualMemory() failed: %v", err)
	}

	if memInfo == nil {
		t.Fatal("memInfo is nil")
	}

	t.Logf("Memory Total: %.2f GB", float64(memInfo.Total)/(1024*1024*1024))
	t.Logf("Memory Used: %.2f GB", float64(memInfo.Used)/(1024*1024*1024))
	t.Logf("Memory Available: %.2f GB", float64(memInfo.Available)/(1024*1024*1024))
	t.Logf("Memory UsedPercent: %.2f%%", memInfo.UsedPercent)

	if memInfo.Total == 0 {
		t.Error("Total memory should not be 0")
	}

	if memInfo.UsedPercent < 0 || memInfo.UsedPercent > 100 {
		t.Errorf("Memory percentage out of range: %f", memInfo.UsedPercent)
	}

	t.Logf("✓ Memory: %.2f%% used", memInfo.UsedPercent)
}

func TestDiskUsage(t *testing.T) {
	t.Log("Testing Disk metrics on", runtime.GOOS)

	path := "/"
	if runtime.GOOS == "windows" {
		path = "C:\\"
	}

	diskUsage, err := disk.Usage(path)
	if err != nil {
		t.Fatalf("disk.Usage(%s) failed: %v", path, err)
	}

	if diskUsage == nil {
		t.Fatal("diskUsage is nil")
	}

	t.Logf("Disk Path: %s", path)
	t.Logf("Disk Total: %.2f GB", float64(diskUsage.Total)/(1024*1024*1024))
	t.Logf("Disk Used: %.2f GB", float64(diskUsage.Used)/(1024*1024*1024))
	t.Logf("Disk Free: %.2f GB", float64(diskUsage.Free)/(1024*1024*1024))
	t.Logf("Disk UsedPercent: %.2f%%", diskUsage.UsedPercent)

	if diskUsage.Total == 0 {
		t.Error("Total disk space should not be 0")
	}

	if diskUsage.UsedPercent < 0 || diskUsage.UsedPercent > 100 {
		t.Errorf("Disk percentage out of range: %f", diskUsage.UsedPercent)
	}

	t.Logf("✓ Disk: %.2f%% used", diskUsage.UsedPercent)
}

func TestSystemInfoCollection(t *testing.T) {
	t.Log("Testing complete system info collection on", runtime.GOOS)

	// Simulate what the handler does
	cpuPercent, err := cpu.Percent(100*time.Millisecond, false)
	if err != nil {
		t.Errorf("CPU collection failed: %v", err)
	}

	memInfo, err := mem.VirtualMemory()
	if err != nil {
		t.Errorf("Memory collection failed: %v", err)
	}

	diskPath := "/"
	if runtime.GOOS == "windows" {
		diskPath = "C:\\"
	}
	diskUsage, err := disk.Usage(diskPath)
	if err != nil {
		t.Errorf("Disk collection failed: %v", err)
	}

	// Build systemInfo map like in the handler
	systemInfo := map[string]interface{}{
		"cpu_percent":    0.0,
		"memory_percent": 0.0,
		"disk_percent":   0.0,
		"timestamp":      time.Now().Format(time.RFC3339),
	}

	if len(cpuPercent) > 0 && cpuPercent[0] > 0 {
		systemInfo["cpu_percent"] = cpuPercent[0]
	}
	if memInfo != nil {
		systemInfo["memory_percent"] = memInfo.UsedPercent
	}
	if diskUsage != nil {
		systemInfo["disk_percent"] = diskUsage.UsedPercent
		systemInfo["disk_used"] = float64(diskUsage.Used) / (1024 * 1024 * 1024)
		systemInfo["disk_total"] = float64(diskUsage.Total) / (1024 * 1024 * 1024)
	}

	t.Logf("\n=== System Info Map ===")
	t.Logf("CPU Percent: %v", systemInfo["cpu_percent"])
	t.Logf("Memory Percent: %v", systemInfo["memory_percent"])
	t.Logf("Disk Percent: %v", systemInfo["disk_percent"])
	t.Logf("Disk Used: %v GB", systemInfo["disk_used"])
	t.Logf("Disk Total: %v GB", systemInfo["disk_total"])
	t.Logf("Timestamp: %v", systemInfo["timestamp"])

	// Verify we have valid data
	cpuVal, ok := systemInfo["cpu_percent"].(float64)
	if !ok || cpuVal == 0 {
		t.Error("❌ CPU percent is 0 or not a float64")
	} else {
		t.Logf("✓ CPU: %.2f%%", cpuVal)
	}

	memVal, ok := systemInfo["memory_percent"].(float64)
	if !ok || memVal == 0 {
		t.Error("❌ Memory percent is 0 or not a float64")
	} else {
		t.Logf("✓ Memory: %.2f%%", memVal)
	}

	diskVal, ok := systemInfo["disk_percent"].(float64)
	if !ok || diskVal == 0 {
		t.Error("❌ Disk percent is 0 or not a float64")
	} else {
		t.Logf("✓ Disk: %.2f%%", diskVal)
	}
}

func TestMultipleCPUReadings(t *testing.T) {
	t.Log("Testing multiple CPU readings over time on", runtime.GOOS)

	readings := []float64{}
	for i := 0; i < 5; i++ {
		cpuPercent, err := cpu.Percent(100*time.Millisecond, false)
		if err != nil {
			t.Errorf("Reading %d failed: %v", i+1, err)
			continue
		}
		if len(cpuPercent) > 0 {
			readings = append(readings, cpuPercent[0])
			t.Logf("Reading %d: %.2f%%", i+1, cpuPercent[0])
		}
		time.Sleep(100 * time.Millisecond)
	}

	if len(readings) != 5 {
		t.Errorf("Expected 5 readings, got %d", len(readings))
	}

	// Check if we have variation (system is not idle)
	allSame := true
	first := readings[0]
	for _, r := range readings {
		if r != first {
			allSame = false
			break
		}
	}

	if !allSame {
		t.Log("✓ CPU readings show variation (system is active)")
	} else {
		t.Log("Note: All CPU readings are the same (system may be idle)")
	}
}

func BenchmarkCPUPercent0(b *testing.B) {
	for i := 0; i < b.N; i++ {
		cpu.Percent(0, false)
	}
}

func BenchmarkCPUPercent100ms(b *testing.B) {
	for i := 0; i < b.N; i++ {
		cpu.Percent(100*time.Millisecond, false)
	}
}

func BenchmarkMemoryInfo(b *testing.B) {
	for i := 0; i < b.N; i++ {
		mem.VirtualMemory()
	}
}

func BenchmarkDiskUsage(b *testing.B) {
	path := "/"
	if runtime.GOOS == "windows" {
		path = "C:\\"
	}
	for i := 0; i < b.N; i++ {
		disk.Usage(path)
	}
}

// +build linux,!nogpu

package monitor

import (
	"fmt"
	"strings"
	"time"

	"github.com/NVIDIA/go-nvml/pkg/nvml"
)

// MetricsCollector collects all available GPU metrics via NVML (Linux)
type MetricsCollector struct {
	previousSamples map[string]map[string]interface{}
	lastSampleTime  map[string]time.Time
}

// NewMetricsCollector creates a new metrics collector
func NewMetricsCollector() *MetricsCollector {
	return &MetricsCollector{
		previousSamples: make(map[string]map[string]interface{}),
		lastSampleTime:  make(map[string]time.Time),
	}
}

// CollectAll collects all available metrics for a GPU
func (mc *MetricsCollector) CollectAll(device nvml.Device, gpuID string) map[string]interface{} {
	data := make(map[string]interface{})
	data["index"] = gpuID
	data["timestamp"] = time.Now().Format(time.RFC3339)

	mc.addBasicInfo(device, data)
	mc.addPerformance(device, data)
	mc.addMemory(device, data, gpuID)
	mc.addPowerThermal(device, data)
	mc.addClocks(device, data)
	mc.addConnectivity(device, data)

	mc.previousSamples[gpuID] = copyMap(data)
	mc.lastSampleTime[gpuID] = time.Now()

	return data
}

func (mc *MetricsCollector) addBasicInfo(device nvml.Device, data map[string]interface{}) {
	if name, ret := device.GetName(); ret == nvml.SUCCESS {
		data["name"] = name
	}

	if uuid, ret := device.GetUUID(); ret == nvml.SUCCESS {
		data["uuid"] = uuid
	}

	if driver, ret := nvml.SystemGetDriverVersion(); ret == nvml.SUCCESS {
		data["driver_version"] = driver
	}

	if vbios, ret := device.GetVbiosVersion(); ret == nvml.SUCCESS {
		data["vbios_version"] = vbios
	}

	// Brand
	if brand, ret := device.GetBrand(); ret == nvml.SUCCESS {
		data["brand"] = getBrandName(brand)
	}

	// Architecture - detect from name if needed
	if name, ok := data["name"].(string); ok {
		data["architecture"] = detectArchFromName(name)
	}

	// CUDA capability
	if major, minor, ret := device.GetCudaComputeCapability(); ret == nvml.SUCCESS {
		data["cuda_compute_capability"] = fmt.Sprintf("%d.%d", major, minor)
	}

	if serial, ret := device.GetSerial(); ret == nvml.SUCCESS {
		data["serial"] = serial
	}
}

func (mc *MetricsCollector) addPerformance(device nvml.Device, data map[string]interface{}) {
	if util, ret := device.GetUtilizationRates(); ret == nvml.SUCCESS {
		data["utilization"] = float64(util.Gpu)
		data["memory_utilization"] = float64(util.Memory)
	}

	if pstate, ret := device.GetPerformanceState(); ret == nvml.SUCCESS {
		data["performance_state"] = fmt.Sprintf("P%d", pstate)
	}

	if mode, ret := device.GetComputeMode(); ret == nvml.SUCCESS {
		modes := map[nvml.ComputeMode]string{
			0: "Default",
			1: "Exclusive Thread",
			2: "Prohibited",
			3: "Exclusive Process",
		}
		if modeName, ok := modes[mode]; ok {
			data["compute_mode"] = modeName
		} else {
			data["compute_mode"] = fmt.Sprintf("Mode %d", mode)
		}
	}

	// Calculate MFU (Model FLOPs Utilization)
	mc.calculateMFU(device, data)
}

func (mc *MetricsCollector) addMemory(device nvml.Device, data map[string]interface{}, gpuID string) {
	if mem, ret := device.GetMemoryInfo(); ret == nvml.SUCCESS {
		data["memory_used"] = float64(mem.Used) / (1024 * 1024)    // MiB
		data["memory_total"] = float64(mem.Total) / (1024 * 1024)  // MiB
		data["memory_free"] = float64(mem.Free) / (1024 * 1024)    // MiB

		// Calculate change rate
		if prev, exists := mc.previousSamples[gpuID]; exists {
			if prevUsed, ok := prev["memory_used"].(float64); ok {
				if lastTime, timeExists := mc.lastSampleTime[gpuID]; timeExists {
					dt := time.Since(lastTime).Seconds()
					if dt > 0 {
						delta := data["memory_used"].(float64) - prevUsed
						data["memory_change_rate"] = delta / dt
					}
				}
			}
		}
	}

	// BAR1 memory
	if bar1, ret := device.GetBAR1MemoryInfo(); ret == nvml.SUCCESS {
		data["bar1_memory_used"] = float64(bar1.Bar1Used) / (1024 * 1024)
		data["bar1_memory_total"] = float64(bar1.Bar1Total) / (1024 * 1024)
	}
}

func (mc *MetricsCollector) addPowerThermal(device nvml.Device, data map[string]interface{}) {
	// Temperature
	if temp, ret := device.GetTemperature(nvml.TEMPERATURE_GPU); ret == nvml.SUCCESS {
		data["temperature"] = float64(temp)
	}

	// Power
	if power, ret := device.GetPowerUsage(); ret == nvml.SUCCESS {
		data["power_draw"] = float64(power) / 1000.0 // Convert mW to W
	}

	if limit, ret := device.GetPowerManagementLimit(); ret == nvml.SUCCESS {
		data["power_limit"] = float64(limit) / 1000.0 // Convert mW to W
	}

	if minLimit, maxLimit, ret := device.GetPowerManagementLimitConstraints(); ret == nvml.SUCCESS {
		data["power_limit_min"] = float64(minLimit) / 1000.0
		data["power_limit_max"] = float64(maxLimit) / 1000.0
	}

	// Fan speed
	if fan, ret := device.GetFanSpeed(); ret == nvml.SUCCESS {
		data["fan_speed"] = float64(fan)
	}

	// Throttle reasons
	if throttle, ret := device.GetCurrentClocksThrottleReasons(); ret == nvml.SUCCESS {
		reasons := []string{}
		throttleMap := map[uint64]string{
			nvml.ClocksThrottleReasonGpuIdle:                    "GPU Idle",
			nvml.ClocksThrottleReasonApplicationsClocksSetting: "App Settings",
			nvml.ClocksThrottleReasonSwPowerCap:                "SW Power Cap",
			nvml.ClocksThrottleReasonHwSlowdown:                "HW Slowdown",
			nvml.ClocksThrottleReasonSwThermalSlowdown:         "SW Thermal",
			nvml.ClocksThrottleReasonHwThermalSlowdown:         "HW Thermal",
			nvml.ClocksThrottleReasonHwPowerBrakeSlowdown:      "Power Brake",
		}
		for flag, label := range throttleMap {
			if throttle&flag != 0 {
				reasons = append(reasons, label)
			}
		}
		if len(reasons) > 0 {
			data["throttle_reasons"] = strings.Join(reasons, ", ")
		} else {
			data["throttle_reasons"] = "None"
		}
	}
}

func (mc *MetricsCollector) addClocks(device nvml.Device, data map[string]interface{}) {
	clockTypes := map[string]nvml.ClockType{
		"clock_graphics": nvml.CLOCK_GRAPHICS,
		"clock_sm":       nvml.CLOCK_SM,
		"clock_memory":   nvml.CLOCK_MEM,
		"clock_video":    nvml.CLOCK_VIDEO,
	}

	for key, clockType := range clockTypes {
		if clock, ret := device.GetClockInfo(clockType); ret == nvml.SUCCESS {
			data[key] = float64(clock)
		}

		if maxClock, ret := device.GetMaxClockInfo(clockType); ret == nvml.SUCCESS {
			data[key+"_max"] = float64(maxClock)
		}

		if appClock, ret := device.GetApplicationsClock(clockType); ret == nvml.SUCCESS {
			data[key+"_app"] = float64(appClock)
		}

		if defaultClock, ret := device.GetDefaultApplicationsClock(clockType); ret == nvml.SUCCESS {
			data[key+"_default"] = float64(defaultClock)
		}
	}
}

func (mc *MetricsCollector) addConnectivity(device nvml.Device, data map[string]interface{}) {
	// PCIe
	if gen, ret := device.GetCurrPcieLinkGeneration(); ret == nvml.SUCCESS {
		data["pcie_gen"] = fmt.Sprintf("%d", gen)
	}

	if maxGen, ret := device.GetMaxPcieLinkGeneration(); ret == nvml.SUCCESS {
		data["pcie_gen_max"] = fmt.Sprintf("%d", maxGen)
	}

	if width, ret := device.GetCurrPcieLinkWidth(); ret == nvml.SUCCESS {
		data["pcie_width"] = fmt.Sprintf("%d", width)
	}

	if maxWidth, ret := device.GetMaxPcieLinkWidth(); ret == nvml.SUCCESS {
		data["pcie_width_max"] = fmt.Sprintf("%d", maxWidth)
	}

	if pci, ret := device.GetPciInfo(); ret == nvml.SUCCESS {
		// Convert BusId int8 array to string
		busIdBytes := make([]byte, 0, len(pci.BusId))
		for _, b := range pci.BusId {
			if b == 0 {
				break
			}
			busIdBytes = append(busIdBytes, byte(b))
		}
		data["pci_bus_id"] = string(busIdBytes)
	}
}

// Helper functions
func getBrandName(brand nvml.BrandType) string {
	brands := map[nvml.BrandType]string{
		nvml.BRAND_UNKNOWN:     "Unknown",
		nvml.BRAND_QUADRO:      "Quadro",
		nvml.BRAND_TESLA:       "Tesla",
		nvml.BRAND_NVS:         "NVS",
		nvml.BRAND_GRID:        "GRID",
		nvml.BRAND_GEFORCE:     "GeForce",
		nvml.BRAND_TITAN:       "Titan",
		nvml.BRAND_NVIDIA_VAPPS: "NVIDIA vApps",
		nvml.BRAND_NVIDIA_VPC:   "NVIDIA VPC",
		nvml.BRAND_NVIDIA_VCS:   "NVIDIA VCS",
		nvml.BRAND_NVIDIA_VWS:   "NVIDIA VWS",
		nvml.BRAND_NVIDIA_VGAMING: "NVIDIA vGaming", // Value 11 (same as BRAND_NVIDIA_CLOUD_GAMING)
	}
	if name, ok := brands[brand]; ok {
		return name
	}
	return fmt.Sprintf("Brand %d", brand)
}

func detectArchFromName(gpuName string) string {
	name := strings.ToUpper(gpuName)

	archPatterns := []struct {
		patterns []string
		arch     string
	}{
		{[]string{"RTX 40", "RTX 4", "L40", "L4"}, "Ada Lovelace"},
		{[]string{"H100", "H200"}, "Hopper"},
		{[]string{"RTX 30", "RTX 3", "A100", "A40", "A30", "A10", "A6000", "A5000", "A4000", "A2000"}, "Ampere"},
		{[]string{"RTX 20", "RTX 2", "GTX 16", "T1000", "T2000", "T600"}, "Turing"},
		{[]string{"GTX 10", "TITAN X", "P100", "P40", "P6"}, "Pascal"},
		{[]string{"GTX 9", "TITAN M", "M60", "M40"}, "Maxwell"},
		{[]string{"GTX 7", "GTX 6", "K80", "K40"}, "Kepler"},
		{[]string{"V100"}, "Volta"},
	}

	for _, ap := range archPatterns {
		for _, pattern := range ap.patterns {
			if strings.Contains(name, pattern) {
				return ap.arch
			}
		}
	}

	return "Unknown"
}

func copyMap(m map[string]interface{}) map[string]interface{} {
	copy := make(map[string]interface{})
	for k, v := range m {
		copy[k] = v
	}
	return copy
}

// calculateMFU calculates Model FLOPs Utilization
func (mc *MetricsCollector) calculateMFU(device nvml.Device, data map[string]interface{}) {
	// Get GPU name to determine peak FLOPs
	gpuName := ""
	if name, ok := data["name"].(string); ok {
		gpuName = strings.ToUpper(name)
	}

	// Get current clock speeds
	var smClock float64 = 0
	if clock, ret := device.GetClockInfo(nvml.CLOCK_SM); ret == nvml.SUCCESS {
		smClock = float64(clock) // MHz
	}

	// Get GPU utilization
	var utilization float64 = 0
	if util, ok := data["utilization"].(float64); ok {
		utilization = util
	}

	// Calculate peak FLOPs based on GPU architecture
	peakTFLOPs := getPeakTFLOPs(gpuName)

	if peakTFLOPs > 0 {
		// Get max SM clock
		var maxSmClock float64 = 0
		if maxClock, ret := device.GetMaxClockInfo(nvml.CLOCK_SM); ret == nvml.SUCCESS {
			maxSmClock = float64(maxClock)
		}

		// Calculate achieved TFLOPs
		// MFU = (current_clock / max_clock) * (utilization / 100) * peak_TFLOPs
		var achievedTFLOPs float64 = 0
		var mfu float64 = 0

		if maxSmClock > 0 && smClock > 0 {
			clockRatio := smClock / maxSmClock
			utilRatio := utilization / 100.0
			achievedTFLOPs = clockRatio * utilRatio * peakTFLOPs
			mfu = (achievedTFLOPs / peakTFLOPs) * 100.0
		} else if utilization > 0 {
			// Fallback: if clock info not available, use utilization directly
			achievedTFLOPs = (utilization / 100.0) * peakTFLOPs
			mfu = utilization
		}

		data["mfu"] = mfu
		data["achieved_tflops"] = achievedTFLOPs
		data["peak_tflops"] = peakTFLOPs
	} else {
		// Unknown GPU, set to 0
		data["mfu"] = 0.0
		data["achieved_tflops"] = 0.0
		data["peak_tflops"] = 0.0
	}
}

// getPeakTFLOPs returns the peak TFLOPs for known GPU models
// These are FP32 (single precision) TFLOPs unless specified
func getPeakTFLOPs(gpuName string) float64 {
	// Hopper Architecture
	if strings.Contains(gpuName, "H100") {
		if strings.Contains(gpuName, "SXM") || strings.Contains(gpuName, "HBM3") {
			return 67.0 // H100 SXM5 80GB FP32 TFLOPs
		}
		return 51.0 // H100 PCIe FP32 TFLOPs
	}
	if strings.Contains(gpuName, "H200") {
		return 67.0 // H200 FP32 TFLOPs
	}

	// Ada Lovelace Architecture
	if strings.Contains(gpuName, "RTX 4090") {
		return 82.6 // RTX 4090 FP32 TFLOPs
	}
	if strings.Contains(gpuName, "RTX 4080") {
		if strings.Contains(gpuName, "SUPER") {
			return 52.2 // RTX 4080 SUPER FP32 TFLOPs
		}
		return 48.7 // RTX 4080 FP32 TFLOPs
	}
	if strings.Contains(gpuName, "RTX 4070") {
		if strings.Contains(gpuName, "TI") || strings.Contains(gpuName, "Ti") {
			if strings.Contains(gpuName, "SUPER") {
				return 44.1 // RTX 4070 Ti SUPER FP32 TFLOPs
			}
			return 40.1 // RTX 4070 Ti FP32 TFLOPs
		}
		if strings.Contains(gpuName, "SUPER") {
			return 35.5 // RTX 4070 SUPER FP32 TFLOPs
		}
		return 29.1 // RTX 4070 FP32 TFLOPs
	}
	if strings.Contains(gpuName, "RTX 4060") {
		if strings.Contains(gpuName, "TI") || strings.Contains(gpuName, "Ti") {
			return 22.1 // RTX 4060 Ti FP32 TFLOPs
		}
		return 15.1 // RTX 4060 FP32 TFLOPs
	}
	if strings.Contains(gpuName, "L40S") {
		return 91.6 // L40S FP32 TFLOPs
	}
	if strings.Contains(gpuName, "L40") {
		return 90.5 // L40 FP32 TFLOPs
	}
	if strings.Contains(gpuName, "L4") {
		return 30.3 // L4 FP32 TFLOPs
	}

	// Ampere Architecture
	if strings.Contains(gpuName, "A100") {
		if strings.Contains(gpuName, "80GB") || strings.Contains(gpuName, "SXM") {
			return 19.5 // A100 80GB FP32 TFLOPs
		}
		return 19.5 // A100 40GB FP32 TFLOPs
	}
	if strings.Contains(gpuName, "A40") {
		return 37.4 // A40 FP32 TFLOPs
	}
	if strings.Contains(gpuName, "A30") {
		return 10.3 // A30 FP32 TFLOPs
	}
	if strings.Contains(gpuName, "A10") {
		if strings.Contains(gpuName, "A10G") {
			return 31.2 // A10G FP32 TFLOPs
		}
		return 31.2 // A10 FP32 TFLOPs
	}
	if strings.Contains(gpuName, "A6000") {
		return 38.7 // A6000 FP32 TFLOPs
	}
	if strings.Contains(gpuName, "A5000") {
		return 27.8 // A5000 FP32 TFLOPs
	}
	if strings.Contains(gpuName, "A4000") {
		return 19.2 // A4000 FP32 TFLOPs
	}
	if strings.Contains(gpuName, "A2000") {
		return 8.0 // A2000 FP32 TFLOPs
	}
	if strings.Contains(gpuName, "RTX 3090") {
		if strings.Contains(gpuName, "TI") || strings.Contains(gpuName, "Ti") {
			return 40.0 // RTX 3090 Ti FP32 TFLOPs
		}
		return 35.6 // RTX 3090 FP32 TFLOPs
	}
	if strings.Contains(gpuName, "RTX 3080") {
		if strings.Contains(gpuName, "TI") || strings.Contains(gpuName, "Ti") {
			return 34.1 // RTX 3080 Ti FP32 TFLOPs
		}
		return 29.8 // RTX 3080 FP32 TFLOPs
	}
	if strings.Contains(gpuName, "RTX 3070") {
		if strings.Contains(gpuName, "TI") || strings.Contains(gpuName, "Ti") {
			return 21.8 // RTX 3070 Ti FP32 TFLOPs
		}
		return 20.3 // RTX 3070 FP32 TFLOPs
	}
	if strings.Contains(gpuName, "RTX 3060") {
		if strings.Contains(gpuName, "TI") || strings.Contains(gpuName, "Ti") {
			return 16.2 // RTX 3060 Ti FP32 TFLOPs
		}
		return 13.0 // RTX 3060 FP32 TFLOPs
	}

	// Turing Architecture
	if strings.Contains(gpuName, "RTX 2080") {
		if strings.Contains(gpuName, "TI") || strings.Contains(gpuName, "Ti") {
			return 13.4 // RTX 2080 Ti FP32 TFLOPs
		}
		if strings.Contains(gpuName, "SUPER") {
			return 11.2 // RTX 2080 SUPER FP32 TFLOPs
		}
		return 10.1 // RTX 2080 FP32 TFLOPs
	}
	if strings.Contains(gpuName, "RTX 2070") {
		if strings.Contains(gpuName, "SUPER") {
			return 9.1 // RTX 2070 SUPER FP32 TFLOPs
		}
		return 7.5 // RTX 2070 FP32 TFLOPs
	}
	if strings.Contains(gpuName, "RTX 2060") {
		if strings.Contains(gpuName, "SUPER") {
			return 7.2 // RTX 2060 SUPER FP32 TFLOPs
		}
		return 6.5 // RTX 2060 FP32 TFLOPs
	}

	// Volta Architecture
	if strings.Contains(gpuName, "V100") {
		if strings.Contains(gpuName, "32GB") || strings.Contains(gpuName, "SXM") {
			return 15.7 // V100 32GB FP32 TFLOPs
		}
		return 14.0 // V100 16GB FP32 TFLOPs
	}
	if strings.Contains(gpuName, "TITAN V") {
		return 15.0 // Titan V FP32 TFLOPs
	}

	// Pascal Architecture
	if strings.Contains(gpuName, "P100") {
		return 9.3 // P100 FP32 TFLOPs
	}
	if strings.Contains(gpuName, "GTX 1080") {
		if strings.Contains(gpuName, "TI") || strings.Contains(gpuName, "Ti") {
			return 11.3 // GTX 1080 Ti FP32 TFLOPs
		}
		return 8.9 // GTX 1080 FP32 TFLOPs
	}
	if strings.Contains(gpuName, "GTX 1070") {
		if strings.Contains(gpuName, "TI") || strings.Contains(gpuName, "Ti") {
			return 8.1 // GTX 1070 Ti FP32 TFLOPs
		}
		return 6.5 // GTX 1070 FP32 TFLOPs
	}
	if strings.Contains(gpuName, "GTX 1060") {
		return 4.4 // GTX 1060 FP32 TFLOPs
	}
	if strings.Contains(gpuName, "TITAN X") {
		if strings.Contains(gpuName, "PASCAL") {
			return 11.0 // Titan X Pascal FP32 TFLOPs
		}
		return 6.1 // Titan X Maxwell FP32 TFLOPs
	}

	// Unknown GPU - return 0 to indicate we can't calculate MFU
	return 0.0
}

package monitor

import (
	"fmt"
	"strings"
	"time"

	"github.com/NVIDIA/go-nvml/pkg/nvml"
)

// MetricsCollector collects all available GPU metrics via NVML
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

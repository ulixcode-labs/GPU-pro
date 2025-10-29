package handlers

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

// getSystemFanSpeeds reads system fan speeds from hwmon sysfs interface (Linux only)
// Returns a map of fan labels to their RPM values
func getSystemFanSpeeds() map[string]int {
	fans := make(map[string]int)
	
	// Path to hardware monitoring devices
	hwmonPath := "/sys/class/hwmon"
	
	// Check if hwmon path exists (Linux only)
	if _, err := os.Stat(hwmonPath); os.IsNotExist(err) {
		return fans
	}
	
	// Iterate through all hwmon devices
	devices, err := ioutil.ReadDir(hwmonPath)
	if err != nil {
		return fans
	}
	
	for _, device := range devices {
		devicePath := filepath.Join(hwmonPath, device.Name())
		
		// Read all files in the device directory
		files, err := ioutil.ReadDir(devicePath)
		if err != nil {
			continue
		}
		
		// Look for fan input files (fan1_input, fan2_input, etc.)
		for _, file := range files {
			if strings.HasPrefix(file.Name(), "fan") && strings.HasSuffix(file.Name(), "_input") {
				fanPath := filepath.Join(devicePath, file.Name())
				
				// Read fan speed
				data, err := ioutil.ReadFile(fanPath)
				if err != nil {
					continue
				}
				
				// Parse RPM value
				rpm, err := strconv.Atoi(strings.TrimSpace(string(data)))
				if err != nil {
					continue
				}
				
				// Try to read the fan label
				labelFile := strings.Replace(file.Name(), "_input", "_label", 1)
				labelPath := filepath.Join(devicePath, labelFile)
				label := file.Name() // Default to filename
				
				if labelData, err := ioutil.ReadFile(labelPath); err == nil {
					label = strings.TrimSpace(string(labelData))
				}
				
				fans[label] = rpm
			}
		}
	}
	
	return fans
}

// getAverageFanSpeed calculates the average fan speed from all system fans
func getAverageFanSpeed(fans map[string]int) int {
	if len(fans) == 0 {
		return 0
	}
	
	total := 0
	for _, rpm := range fans {
		total += rpm
	}
	
	return total / len(fans)
}

// getMaxFanSpeed gets the maximum RPM from all fans to calculate percentage
// Common max RPM is around 2000-3000, but we'll track the actual max seen
func getMaxFanSpeed(fans map[string]int) int {
	max := 0
	for _, rpm := range fans {
		if rpm > max {
			max = rpm
		}
	}
	
	// Use a reasonable default max if no fans are spinning
	if max == 0 {
		return 2000
	}
	
	return max
}

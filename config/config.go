package config

import (
	"os"
	"strconv"
	"strings"
)

// Config holds all application configuration
type Config struct {
	// Server configuration
	Host  string
	Port  int
	Debug bool

	// Monitoring configuration
	UpdateInterval    float64 // Update interval for NVML (sub-second monitoring)
	NvidiaSMIInterval float64 // Update interval for nvidia-smi fallback

	// GPU Monitoring Mode
	NvidiaSMI bool // Force nvidia-smi mode

	// Multi-Node Configuration
	Mode     string   // "default" (single node) or "hub" (aggregate multiple nodes)
	NodeName string   // Node identifier
	NodeURLs []string // Comma-separated URLs for hub mode
}

// Default configuration values
var (
	DefaultHost              = "0.0.0.0"
	DefaultPort              = 8889
	DefaultUpdateInterval    = 0.5 // 500ms
	DefaultNvidiaSMIInterval = 2.0 // 2s
)

// Load reads configuration from environment variables
func Load() *Config {
	cfg := &Config{
		Host:              getEnv("HOST", DefaultHost),
		Port:              getEnvInt("PORT", DefaultPort),
		Debug:             getEnvBool("DEBUG", false),
		UpdateInterval:    getEnvFloat("UPDATE_INTERVAL", DefaultUpdateInterval),
		NvidiaSMIInterval: getEnvFloat("NVIDIA_SMI_INTERVAL", DefaultNvidiaSMIInterval),
		NvidiaSMI:         getEnvBool("NVIDIA_SMI", false),
		Mode:              getEnv("GPU_HOT_MODE", "default"),
		NodeName:          getEnv("NODE_NAME", getHostname()),
	}

	// Parse NODE_URLS
	if nodeURLsStr := os.Getenv("NODE_URLS"); nodeURLsStr != "" {
		urls := strings.Split(nodeURLsStr, ",")
		for _, url := range urls {
			if trimmed := strings.TrimSpace(url); trimmed != "" {
				cfg.NodeURLs = append(cfg.NodeURLs, trimmed)
			}
		}
	}

	return cfg
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intVal, err := strconv.Atoi(value); err == nil {
			return intVal
		}
	}
	return defaultValue
}

func getEnvFloat(key string, defaultValue float64) float64 {
	if value := os.Getenv(key); value != "" {
		if floatVal, err := strconv.ParseFloat(value, 64); err == nil {
			return floatVal
		}
	}
	return defaultValue
}

func getEnvBool(key string, defaultValue bool) bool {
	if value := os.Getenv(key); value != "" {
		return strings.ToLower(value) == "true"
	}
	return defaultValue
}

func getHostname() string {
	hostname, err := os.Hostname()
	if err != nil {
		return "unknown"
	}
	return hostname
}

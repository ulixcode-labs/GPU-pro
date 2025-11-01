package analytics

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"runtime"
	"time"
)

const (
	// Analytics backend URL - update this to your deployed Worker URL
	analyticsURL = "https://gpu-pro-analytics-backend.xing-mathcoder.workers.dev/heartbeat"
	// Heartbeat interval (5 minutes)
	heartbeatInterval = 5 * time.Minute
)

// HeartbeatClient manages analytics heartbeats
type HeartbeatClient struct {
	clientID  string
	version   string
	appType   string
	gpuInfo   string
	ticker    *time.Ticker
	stopChan  chan bool
	isRunning bool
}

// HeartbeatPayload represents the data sent to analytics backend
type HeartbeatPayload struct {
	ClientID string `json:"client_id"`
	Hostname string `json:"hostname,omitempty"`
	AppType  string `json:"app_type,omitempty"`
	GPUInfo  string `json:"gpu_info,omitempty"`
	OSInfo   string `json:"os_info,omitempty"`
	Version  string `json:"version,omitempty"`
}

// NewHeartbeatClient creates a new heartbeat client
// appType should be "webui" or "tui"
func NewHeartbeatClient(version, appType string) *HeartbeatClient {
	return &HeartbeatClient{
		clientID: generateClientID(),
		version:  version,
		appType:  appType,
		stopChan: make(chan bool),
	}
}

// generateClientID creates a unique client identifier based on hostname and MAC
func generateClientID() string {
	// Use hostname as base for client ID
	hostname, err := os.Hostname()
	if err != nil {
		hostname = "unknown"
	}

	// Add some system info to make it more unique
	systemInfo := fmt.Sprintf("%s-%s-%s", hostname, runtime.GOOS, runtime.GOARCH)

	// Hash it to create a stable, anonymized ID
	hash := sha256.Sum256([]byte(systemInfo))
	return hex.EncodeToString(hash[:])[:32]
}

// Start begins sending heartbeats
func (hb *HeartbeatClient) Start() {
	if hb.isRunning {
		return
	}

	hb.isRunning = true
	log.Println("📡 Starting analytics heartbeat (interval: 5 minutes)")

	// Send initial heartbeat
	go hb.sendHeartbeat()

	// Setup ticker for periodic heartbeats
	hb.ticker = time.NewTicker(heartbeatInterval)

	go func() {
		for {
			select {
			case <-hb.ticker.C:
				hb.sendHeartbeat()
			case <-hb.stopChan:
				return
			}
		}
	}()
}

// Stop stops sending heartbeats
func (hb *HeartbeatClient) Stop() {
	if !hb.isRunning {
		return
	}

	hb.isRunning = false
	if hb.ticker != nil {
		hb.ticker.Stop()
	}
	close(hb.stopChan)
	//log.Println("📡 Analytics heartbeat stopped")
}

// SetGPUInfo updates GPU information
func (hb *HeartbeatClient) SetGPUInfo(gpuInfo string) {
	hb.gpuInfo = gpuInfo
}

// sendHeartbeat sends a heartbeat to the analytics backend
func (hb *HeartbeatClient) sendHeartbeat() {
	hostname, _ := os.Hostname()

	payload := HeartbeatPayload{
		ClientID: hb.clientID,
		Hostname: hostname,
		AppType:  hb.appType,
		GPUInfo:  hb.gpuInfo,
		OSInfo:   fmt.Sprintf("%s/%s", runtime.GOOS, runtime.GOARCH),
		Version:  hb.version,
	}

	jsonData, err := json.Marshal(payload)
	if err != nil {
		log.Printf("⚠️  Failed to marshal heartbeat payload: %v", err)
		return
	}

	// Send heartbeat in background (don't block if it fails)
	go func() {
		client := &http.Client{
			Timeout: 10 * time.Second,
		}

		resp, err := client.Post(analyticsURL, "application/json", bytes.NewBuffer(jsonData))
		if err != nil {
			// Silently fail - don't spam logs if analytics is down
			return
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			// Silently fail
			return
		}
	}()
}

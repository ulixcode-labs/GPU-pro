package hub

import (
	"encoding/json"
	"log"
	"sync"
	"time"

	"gpu-pro/analytics"

	"github.com/gorilla/websocket"
)

// NodeInfo holds information about a connected node
type NodeInfo struct {
	URL        string                 `json:"url"`
	Data       map[string]interface{} `json:"data"`
	Status     string                 `json:"status"`
	LastUpdate string                 `json:"last_update"`
	conn       *websocket.Conn
	mu         sync.RWMutex
}

// Hub aggregates GPU data from multiple nodes
type Hub struct {
	nodeURLs        []string
	nodes           map[string]*NodeInfo
	urlToNode       map[string]string
	running         bool
	mu              sync.RWMutex
	connMu          sync.Mutex
	connStarted     bool
	heartbeatClient *analytics.HeartbeatClient
}

// NewHub creates a new hub instance
func NewHub(nodeURLs []string) *Hub {
	hub := &Hub{
		nodeURLs:        nodeURLs,
		nodes:           make(map[string]*NodeInfo),
		urlToNode:       make(map[string]string),
		heartbeatClient: analytics.NewHeartbeatClient("v2.0-hub", "webui"), // GPU Pro hub version, WebUI mode
	}

	// Initialize nodes as offline
	for _, url := range nodeURLs {
		hub.nodes[url] = &NodeInfo{
			URL:        url,
			Data:       nil,
			Status:     "offline",
			LastUpdate: "",
		}
		hub.urlToNode[url] = url
	}

	// Start analytics heartbeat
	hub.heartbeatClient.Start()

	return hub
}

// Start begins connecting to all nodes
func (h *Hub) Start() {
	h.connMu.Lock()
	defer h.connMu.Unlock()

	if h.connStarted {
		return
	}

	h.connStarted = true
	h.running = true

	// Start connections in background
	go h.connectAllNodes()
}

func (h *Hub) connectAllNodes() {
	// Wait a bit for initialization
	time.Sleep(2 * time.Second)

	// Connect to all nodes concurrently
	var wg sync.WaitGroup
	for _, url := range h.nodeURLs {
		wg.Add(1)
		go func(nodeURL string) {
			defer wg.Done()
			h.connectNodeWithRetry(nodeURL)
		}(url)
	}
	wg.Wait()
}

func (h *Hub) connectNodeWithRetry(url string) {
	maxRetries := 5
	retryDelay := 2 * time.Second

	for attempt := 0; attempt < maxRetries; attempt++ {
		err := h.connectNode(url)
		if err == nil {
			return // Success
		}

		if attempt < maxRetries-1 {
			log.Printf("Connection attempt %d/%d failed for %s: %v, retrying in %v...",
				attempt+1, maxRetries, url, err, retryDelay)
			time.Sleep(retryDelay)
		} else {
			log.Printf("Failed to connect to node %s after %d attempts: %v", url, maxRetries, err)
		}
	}
}

func (h *Hub) connectNode(url string) error {
	for h.running {
		// Convert HTTP URL to WebSocket URL
		wsURL := url
		if len(wsURL) > 7 && wsURL[:7] == "http://" {
			wsURL = "ws://" + wsURL[7:]
		} else if len(wsURL) > 8 && wsURL[:8] == "https://" {
			wsURL = "wss://" + wsURL[8:]
		}
		wsURL += "/socket.io/"

		log.Printf("Connecting to node WebSocket: %s", wsURL)

		// Connect to WebSocket
		conn, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
		if err != nil {
			h.markNodeOffline(url)
			time.Sleep(5 * time.Second)
			continue
		}

		log.Printf("Connected to node: %s", url)

		// Store connection
		nodeName := h.getNodeName(url)
		h.mu.Lock()
		if node, ok := h.nodes[nodeName]; ok {
			node.mu.Lock()
			node.conn = conn
			node.Status = "online"
			node.LastUpdate = time.Now().Format(time.RFC3339)
			node.mu.Unlock()
		}
		h.mu.Unlock()

		// Listen for messages
		for {
			_, message, err := conn.ReadMessage()
			if err != nil {
				log.Printf("WebSocket connection closed for node: %s - %v", url, err)
				h.markNodeOffline(url)
				break
			}

			// Parse message
			var data map[string]interface{}
			if err := json.Unmarshal(message, &data); err != nil {
				log.Printf("Failed to parse message from %s: %v", url, err)
				continue
			}

			// Extract node name from data or use URL
			nodeName := url
			if name, ok := data["node_name"].(string); ok && name != "" {
				nodeName = name
				h.mu.Lock()
				h.urlToNode[url] = nodeName
				h.mu.Unlock()
			}

			// Update node data
			h.mu.Lock()
			if _, exists := h.nodes[nodeName]; !exists {
				h.nodes[nodeName] = &NodeInfo{}
			}
			node := h.nodes[nodeName]
			node.mu.Lock()
			node.URL = url
			node.Data = data
			node.Status = "online"
			node.LastUpdate = time.Now().Format(time.RFC3339)
			node.mu.Unlock()
			h.mu.Unlock()
		}

		// Connection closed, retry after delay
		if h.running {
			time.Sleep(5 * time.Second)
		}
	}

	return nil
}

func (h *Hub) markNodeOffline(url string) {
	h.mu.Lock()
	defer h.mu.Unlock()

	nodeName := h.urlToNode[url]
	if node, ok := h.nodes[nodeName]; ok {
		node.mu.Lock()
		node.Status = "offline"
		node.mu.Unlock()
		log.Printf("Marked node %s as offline", nodeName)
	}
}

func (h *Hub) getNodeName(url string) string {
	h.mu.RLock()
	defer h.mu.RUnlock()
	if name, ok := h.urlToNode[url]; ok {
		return name
	}
	return url
}

// GetClusterData gets aggregated data from all nodes
func (h *Hub) GetClusterData() map[string]interface{} {
	h.mu.RLock()
	defer h.mu.RUnlock()

	nodes := make(map[string]interface{})
	totalGPUs := 0
	onlineNodes := 0

	for nodeName, nodeInfo := range h.nodes {
		nodeInfo.mu.RLock()
		if nodeInfo.Status == "online" && nodeInfo.Data != nil {
			gpus := make(map[string]interface{})
			if gpusData, ok := nodeInfo.Data["gpus"].(map[string]interface{}); ok {
				gpus = gpusData
			}

			processes := []interface{}{}
			if procsData, ok := nodeInfo.Data["processes"].([]interface{}); ok {
				processes = procsData
			}

			system := make(map[string]interface{})
			if sysData, ok := nodeInfo.Data["system"].(map[string]interface{}); ok {
				system = sysData
			}

			nodes[nodeName] = map[string]interface{}{
				"status":      "online",
				"gpus":        gpus,
				"processes":   processes,
				"system":      system,
				"last_update": nodeInfo.LastUpdate,
			}

			totalGPUs += len(gpus)
			onlineNodes++
		} else {
			nodes[nodeName] = map[string]interface{}{
				"status":      "offline",
				"gpus":        map[string]interface{}{},
				"processes":   []interface{}{},
				"system":      map[string]interface{}{},
				"last_update": nodeInfo.LastUpdate,
			}
		}
		nodeInfo.mu.RUnlock()
	}

	return map[string]interface{}{
		"mode":  "hub",
		"nodes": nodes,
		"cluster_stats": map[string]interface{}{
			"total_nodes":  len(h.nodes),
			"online_nodes": onlineNodes,
			"total_gpus":   totalGPUs,
		},
	}
}

// Shutdown disconnects from all nodes
func (h *Hub) Shutdown() {
	h.running = false

	// Stop heartbeat client
	if h.heartbeatClient != nil {
		h.heartbeatClient.Stop()
	}

	h.mu.RLock()
	defer h.mu.RUnlock()

	for _, nodeInfo := range h.nodes {
		nodeInfo.mu.Lock()
		if nodeInfo.conn != nil {
			nodeInfo.conn.Close()
		}
		nodeInfo.mu.Unlock()
	}
}

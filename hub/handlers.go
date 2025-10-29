package hub

import (
	"encoding/json"
	"log"
	"sync"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/websocket/v2"
)

// WebSocketClients holds all connected WebSocket clients for hub mode
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

// RegisterHubHandlers registers WebSocket handlers for hub mode
func RegisterHubHandlers(app *fiber.App, h *Hub) {
	wsClients := NewWebSocketClients()
	hubRunning := false
	var hubMu sync.Mutex

	// WebSocket endpoint
	app.Get("/socket.io/", websocket.New(func(c *websocket.Conn) {
		wsClients.Add(c)
		log.Println("Dashboard client connected")

		// Start hub loop if not already running
		hubMu.Lock()
		if !hubRunning {
			hubRunning = true
			go hubLoop(h, wsClients)
		}
		hubMu.Unlock()

		// Start node connections if not already started
		h.Start()

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

// hubLoop is the background loop that emits aggregated cluster data
func hubLoop(h *Hub, wsClients *WebSocketClients) {
	log.Println("Hub monitoring loop started")

	ticker := time.NewTicker(500 * time.Millisecond) // 0.5s to match node update rate
	defer ticker.Stop()

	for range ticker.C {
		// Skip if no clients connected
		if wsClients.Count() == 0 {
			continue
		}

		// Get cluster data
		clusterData := h.GetClusterData()

		// Send to all connected clients
		data, err := json.Marshal(clusterData)
		if err != nil {
			log.Printf("Error marshaling cluster data: %v", err)
			continue
		}

		wsClients.Broadcast(data)
	}
}

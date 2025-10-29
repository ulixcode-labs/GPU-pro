package main

import (
	"embed"
	"fmt"
	"io/fs"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"gpu-pro/config"
	"gpu-pro/handlers"
	"gpu-pro/hub"
	"gpu-pro/monitor"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/filesystem"
	"github.com/gofiber/websocket/v2"
)

//go:embed static templates
var embeddedFiles embed.FS

func main() {
	// Setup logging
	log.SetFlags(log.LstdFlags | log.Lshortfile)

	// Setup signal handling IMMEDIATELY before anything else
	// Using buffered channel to ensure signal isn't lost
	quit := make(chan os.Signal, 2)
	signal.Notify(quit, os.Interrupt, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)
	log.Println("✓ Signal handler registered (SIGQUIT/Ctrl+\\, SIGINT/Ctrl+C, SIGTERM)")

	// Load configuration
	cfg := config.Load()

	if cfg.Debug {
		log.Println("Debug mode enabled")
	}

	// Create Fiber app with disabled prefork to avoid signal issues
	app := fiber.New(fiber.Config{
		DisableStartupMessage: false,
		AppName:               "GPU Pro v2.0",
		DisableKeepalive:      false,
		Prefork:               false, // Disable prefork to prevent signal handling conflicts
	})

	// Serve static files from embedded FS
	staticFS, err := fs.Sub(embeddedFiles, "static")
	if err != nil {
		log.Fatalf("Failed to get static subdirectory: %v", err)
	}

	app.Use("/static", filesystem.New(filesystem.Config{
		Root: http.FS(staticFS),
	}))

	// Index page
	app.Get("/", func(c *fiber.Ctx) error {
		data, err := embeddedFiles.ReadFile("templates/index.html")
		if err != nil {
			return c.Status(500).SendString("Failed to load index.html")
		}
		c.Set("Content-Type", "text/html")
		return c.Send(data)
	})

	// Mode selection
	var monitorOrHub interface{}

	if cfg.Mode == "hub" {
		// Hub mode: aggregate data from multiple nodes
		if len(cfg.NodeURLs) == 0 {
			log.Fatal("Hub mode requires NODE_URLS environment variable")
		}

		log.Println("Starting GPU Pro in HUB mode")
		log.Printf("Connecting to %d node(s): %v", len(cfg.NodeURLs), cfg.NodeURLs)

		h := hub.NewHub(cfg.NodeURLs)
		hub.RegisterHubHandlers(app, h)
		monitorOrHub = h

		// API endpoint for hub mode
		app.Get("/api/gpu-data", func(c *fiber.Ctx) error {
			return c.JSON(fiber.Map{
				"gpus":      fiber.Map{},
				"timestamp": "hub_mode",
			})
		})

	} else {
		// Default mode: monitor local GPUs
		log.Println("Starting GPU Pro (Monitor mode)")
		log.Printf("Node name: %s", cfg.NodeName)

		mon := monitor.NewGPUMonitor()
		handlers.RegisterHandlers(app, mon, cfg)
		monitorOrHub = mon

		// API endpoint for monitor mode
		app.Get("/api/gpu-data", func(c *fiber.Ctx) error {
			gpuData, err := mon.GetGPUData()
			if err != nil {
				return c.JSON(fiber.Map{
					"gpus":      fiber.Map{},
					"timestamp": "error",
				})
			}
			return c.JSON(fiber.Map{
				"gpus":      gpuData,
				"timestamp": "async",
			})
		})
	}

	// Upgrade WebSocket connections
	app.Use("/socket.io/", func(c *fiber.Ctx) error {
		if websocket.IsWebSocketUpgrade(c) {
			return c.Next()
		}
		return fiber.ErrUpgradeRequired
	})

	// Start server
	addr := fmt.Sprintf("%s:%d", cfg.Host, cfg.Port)

	// Start server in a goroutine
	serverErr := make(chan error, 1)
	go func() {
		log.Printf("🚀 GPU Pro server starting on %s", addr)
		log.Printf("📊 Access the dashboard at http://%s", addr)
		log.Printf("⚡ Press Ctrl+\\ to stop the server\n")

		if err := app.Listen(addr); err != nil {
			serverErr <- err
		}
	}()

	// Wait for interrupt signal or server error
	log.Println("Main thread waiting for signals...")
	select {
	case sig := <-quit:
		log.Printf("\n🛑 Received signal: %v (type: %T)", sig, sig)
		log.Println("⚠️  Shutting down server (press Ctrl+\\ again to force quit)...")

		// Setup force quit handler
		go func() {
			sig2 := <-quit
			log.Printf("\n💥 Second signal received (%v) - forcing immediate exit!", sig2)
			os.Exit(1)
		}()

		// Try graceful shutdown with timeout
		done := make(chan bool, 1)
		go func() {
			defer func() {
				if r := recover(); r != nil {
					log.Printf("  ⚠️  Panic during shutdown: %v", r)
				}
			}()

			log.Println("  → Stopping HTTP server...")
			if err := app.Shutdown(); err != nil {
				log.Printf("  ⚠️  Server shutdown error: %v", err)
			}

			log.Println("  → Cleaning up resources...")
			if mon, ok := monitorOrHub.(*monitor.GPUMonitor); ok {
				mon.Shutdown()
			} else if h, ok := monitorOrHub.(*hub.Hub); ok {
				h.Shutdown()
			}

			log.Println("  → Cleanup complete")
			done <- true
		}()

		// Wait for shutdown with timeout
		select {
		case <-done:
			log.Println("✅ Shutdown complete")
		case <-time.After(3 * time.Second):
			log.Println("⚠️  Shutdown timeout - forcing exit")
		}

	case err := <-serverErr:
		log.Printf("❌ Server error: %v", err)
	}

	log.Println("Exiting process...")
	os.Exit(0)
}

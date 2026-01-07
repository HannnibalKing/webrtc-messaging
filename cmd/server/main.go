package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/pion/webrtc/v3"
	"github.com/yourusername/webrtc-messaging/config"
	"github.com/yourusername/webrtc-messaging/pkg/message"
	"github.com/yourusername/webrtc-messaging/pkg/peer"
	"github.com/yourusername/webrtc-messaging/pkg/server"
	"github.com/yourusername/webrtc-messaging/pkg/signal"
)

func main() {
	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	if err := cfg.Validate(); err != nil {
		log.Fatalf("Configuration validation failed: %v", err)
	}

	// Build ICE servers from configuration
	iceServers := buildICEServers(cfg)

	// Initialize hub
	hub := signal.NewHub(cfg.MaxPeersPerRoom, iceServers)

	// Initialize connection manager
	connManager, err := peer.NewConnectionManager(iceServers)
	if err != nil {
		log.Fatalf("Failed to create connection manager: %v", err)
	}

	// Initialize message queue and router
	msgQueue := message.NewMessageQueue(1000)
	msgRouter := message.NewMessageRouter(msgQueue)

	// Register message handlers
	registerMessageHandlers(msgRouter, hub)

	// Initialize server
	addr := fmt.Sprintf("%s:%d", cfg.ServerHost, cfg.ServerPort)
	webServer := server.NewServer(addr, hub, connManager, msgRouter)

	// Handle graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-sigChan
		fmt.Println("\nShutting down server...")
		hub.Close()
		os.Exit(0)
	}()

	// Start server
	if err := webServer.Start(); err != nil {
		log.Fatalf("Server error: %v", err)
	}
}

// buildICEServers builds WebRTC ICE servers from configuration
func buildICEServers(cfg *config.Config) []webrtc.ICEServer {
	var iceServers []webrtc.ICEServer

	// Add STUN servers
	for _, stunServer := range cfg.STUNServers {
		iceServers = append(iceServers, webrtc.ICEServer{
			URLs: []string{stunServer},
		})
	}

	// Add TURN servers
	for _, turnServer := range cfg.TURNServers {
		iceServers = append(iceServers, webrtc.ICEServer{
			URLs: []string{turnServer},
		})
	}

	// Add default STUN servers if none configured
	if len(iceServers) == 0 {
		iceServers = append(iceServers, webrtc.ICEServer{
			URLs: []string{
				"stun:stun.l.google.com:19302",
				"stun:stun1.l.google.com:19302",
			},
		})
	}

	return iceServers
}

// registerMessageHandlers registers handlers for message routing
func registerMessageHandlers(router *message.MessageRouter, hub *signal.Hub) {
	// Handle text messages
	router.RegisterHandler(string(message.TextMessage), func(msg *message.Message) error {
		fmt.Printf("[ROOM %s] %s: %s\n", msg.RoomID, msg.SenderID, msg.Content)
		return hub.BroadcastToRoom(msg.RoomID, msg.SenderID, []byte(msg.Content))
	})

	// Handle system messages
	router.RegisterHandler(string(message.SystemMessage), func(msg *message.Message) error {
		fmt.Printf("[SYSTEM] %s\n", msg.Content)
		return nil
	})
}

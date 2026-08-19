// Package main demonstrates a WebRTC client implementation
package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/url"
	"time"

	"github.com/gorilla/websocket"
	"github.com/pion/webrtc/v3"
)

type Client struct {
	conn *websocket.Conn
	done chan struct{}
}

type SignalingMessage struct {
	Type      string      `json:"type"`
	SDP       string      `json:"sdp,omitempty"`
	RoomID    string      `json:"room_id,omitempty"`
	UserID    string      `json:"user_id,omitempty"`
	Candidate interface{} `json:"candidate,omitempty"`
	Content   string      `json:"content,omitempty"`
}

func main() {
	var (
		userID = flag.String("user", "user-"+fmt.Sprint(time.Now().Unix()), "User ID")
		roomID = flag.String("room", "test-room", "Room ID")
	)
	flag.Parse()

	// Dial WebSocket server
	u := url.URL{Scheme: "ws", Host: "localhost:8080", Path: "/ws"}
	log.Printf("Connecting to %s", u.String())

	conn, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
	if err != nil {
		log.Fatal("Dial error:", err)
	}
	defer conn.Close()

	client := &Client{
		conn: conn,
		done: make(chan struct{}),
	}

	// Read messages in goroutine
	go client.readMessages()

	// Send join room message
	joinMsg := SignalingMessage{
		Type:   "join_room",
		RoomID: *roomID,
		UserID: *userID,
	}

	if err := conn.WriteJSON(joinMsg); err != nil {
		log.Fatal("Send error:", err)
	}

	log.Printf("Joined room %s as user %s", *roomID, *userID)

	// Keep connection alive
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			// Send periodic message
			msg := SignalingMessage{
				Type:    "message",
				Content: fmt.Sprintf("Hello from %s at %s", *userID, time.Now().Format(time.RFC3339)),
			}
			if err := conn.WriteJSON(msg); err != nil {
				log.Fatal("Send error:", err)
			}
		case <-client.done:
			return
		}
	}
}

func (c *Client) readMessages() {
	defer close(c.done)

	for {
		var msg map[string]interface{}
		if err := c.conn.ReadJSON(&msg); err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("WebSocket error: %v", err)
			}
			return
		}

		msgType, _ := msg["type"].(string)
		data, _ := json.MarshalIndent(msg, "", "  ")

		log.Printf("[%s] %s\n", msgType, string(data))

		switch msgType {
		case "joined_room":
			log.Println("Successfully joined room")
			if peers, ok := msg["peers"].([]interface{}); ok {
				log.Printf("Current peers: %d\n", len(peers))
			}

		case "offer":
			// Handle WebRTC offer
			log.Println("Received offer, would create answer")

		case "answer":
			// Handle WebRTC answer
			log.Println("Received answer")

		case "ice_candidate":
			// Handle ICE candidate
			log.Println("Received ICE candidate")

		case "message":
			if msgData, ok := msg["message"].(map[string]interface{}); ok {
				content, _ := msgData["content"].(string)
				senderID, _ := msgData["sender_id"].(string)
				log.Printf("[CHAT] %s: %s", senderID, content)
			}

		case "error":
			errMsg, _ := msg["error"].(string)
			log.Printf("Error: %s", errMsg)
		}
	}
}

// Example WebRTC peer connection setup (not executed by default)
func examplePeerConnection() {
	// Create peer connection with default settings
	peerConnection, err := webrtc.NewPeerConnection(webrtc.Configuration{})
	if err != nil {
		log.Fatal(err)
	}
	defer peerConnection.Close()

	// Set up event handlers
	peerConnection.OnICEConnectionStateChange(func(connectionState webrtc.ICEConnectionState) {
		fmt.Printf("Connection State has changed %s \n", connectionState.String())
	})

	peerConnection.OnConnectionStateChange(func(s webrtc.PeerConnectionState) {
		fmt.Printf("Peer Connection State has changed: %s\n", s.String())
	})

	// Create data channel
	dataChannel, err := peerConnection.CreateDataChannel("chat", nil)
	if err != nil {
		log.Fatal(err)
	}

	dataChannel.OnOpen(func() {
		fmt.Println("Data channel opened")
		dataChannel.SendText("Hello from client!")
	})

	dataChannel.OnMessage(func(msg webrtc.DataChannelMessage) {
		fmt.Printf("Message from server: %s\n", string(msg.Data))
	})

	// Create offer
	offer, err := peerConnection.CreateOffer(nil)
	if err != nil {
		log.Fatal(err)
	}

	// Set local description
	if err = peerConnection.SetLocalDescription(offer); err != nil {
		log.Fatal(err)
	}

	fmt.Println("Offer SDP:", offer.SDP)
}

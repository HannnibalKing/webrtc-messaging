package server

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	"github.com/pion/webrtc/v3"
	"github.com/yourusername/webrtc-messaging/pkg/message"
	"github.com/yourusername/webrtc-messaging/pkg/peer"
	"github.com/yourusername/webrtc-messaging/pkg/signal"
)

// Server represents the WebRTC signaling server
type Server struct {
	hub         *signal.Hub
	connManager *peer.ConnectionManager
	msgRouter   *message.MessageRouter
	msgQueue    *message.MessageQueue

	mux      *http.ServeMux
	addr     string
	upgrader websocket.Upgrader

	// Client connections
	clients   map[*Client]bool
	clientsMu sync.RWMutex

	// Channel for client management
	register   chan *Client
	unregister chan *Client
	broadcast  chan interface{}
}

// Client represents a connected WebSocket client
type Client struct {
	conn   *websocket.Conn
	server *Server
	send   chan interface{}
	roomID string
	peerID string
	userID string
	closed bool
	mu     sync.RWMutex
}

// Event types
type JoinRoomEvent struct {
	RoomID string `json:"room_id"`
	Name   string `json:"name"`
	UserID string `json:"user_id"`
}

type OfferEvent struct {
	Type   string                    `json:"type"`
	Offer  webrtc.SessionDescription `json:"offer"`
	PeerID string                    `json:"peer_id"`
}

type AnswerEvent struct {
	Type   string                    `json:"type"`
	Answer webrtc.SessionDescription `json:"answer"`
	PeerID string                    `json:"peer_id"`
}

type ICECandidateEvent struct {
	Type      string            `json:"type"`
	Candidate peer.ICECandidate `json:"candidate"`
	PeerID    string            `json:"peer_id"`
}

type MessageEvent struct {
	Type    string      `json:"type"`
	Message interface{} `json:"message"`
}

// NewServer creates a new WebRTC server
func NewServer(addr string, hub *signal.Hub, connManager *peer.ConnectionManager, msgRouter *message.MessageRouter) *Server {
	return &Server{
		hub:         hub,
		connManager: connManager,
		msgRouter:   msgRouter,
		msgQueue:    message.NewMessageQueue(1000),
		mux:         http.NewServeMux(),
		addr:        addr,
		upgrader: websocket.Upgrader{
			ReadBufferSize:  1024,
			WriteBufferSize: 1024,
			CheckOrigin: func(r *http.Request) bool {
				return true // Allow all origins for demo
			},
		},
		clients:    make(map[*Client]bool),
		register:   make(chan *Client),
		unregister: make(chan *Client),
		broadcast:  make(chan interface{}, 256),
	}
}

// Start starts the server
func (s *Server) Start() error {
	s.mux.HandleFunc("/", s.handleIndex)
	s.mux.HandleFunc("/ws", s.handleWebSocket)
	s.mux.HandleFunc("/api/health", s.handleHealth)
	s.mux.HandleFunc("/api/rooms", s.handleRooms)

	go s.run()

	fmt.Printf("WebRTC server starting on %s\n", s.addr)
	return http.ListenAndServe(s.addr, s.mux)
}

// run manages client connections
func (s *Server) run() {
	for {
		select {
		case client := <-s.register:
			s.clientsMu.Lock()
			s.clients[client] = true
			s.clientsMu.Unlock()
		case client := <-s.unregister:
			s.clientsMu.Lock()
			if _, ok := s.clients[client]; ok {
				delete(s.clients, client)
				close(client.send)
			}
			s.clientsMu.Unlock()
		case msg := <-s.broadcast:
			s.clientsMu.RLock()
			for client := range s.clients {
				select {
				case client.send <- msg:
				default:
					// Client's send channel is full, close it
					go s.closeClient(client)
				}
			}
			s.clientsMu.RUnlock()
		}
	}
}

// handleWebSocket handles WebSocket connections
func (s *Server) handleWebSocket(w http.ResponseWriter, r *http.Request) {
	conn, err := s.upgrader.Upgrade(w, r, nil)
	if err != nil {
		fmt.Printf("WebSocket upgrade error: %v\n", err)
		return
	}

	client := &Client{
		conn:   conn,
		server: s,
		send:   make(chan interface{}, 256),
	}

	s.register <- client

	go client.readPump()
	go client.writePump()
}

// readPump reads messages from the WebSocket
func (c *Client) readPump() {
	defer func() {
		c.server.unregister <- c
		c.conn.Close()
	}()

	c.conn.SetReadDeadline(time.Now().Add(60 * time.Second))
	c.conn.SetPongHandler(func(string) error {
		c.conn.SetReadDeadline(time.Now().Add(60 * time.Second))
		return nil
	})

	for {
		var event map[string]interface{}
		err := c.conn.ReadJSON(&event)
		if err != nil {
			return
		}

		eventType, ok := event["type"].(string)
		if !ok {
			continue
		}

		switch eventType {
		case "join_room":
			c.handleJoinRoom(event)
		case "leave_room":
			c.handleLeaveRoom(event)
		case "offer":
			c.handleOffer(event)
		case "answer":
			c.handleAnswer(event)
		case "ice_candidate":
			c.handleICECandidate(event)
		case "message":
			c.handleMessage(event)
		}
	}
}

// writePump writes messages to the WebSocket
func (c *Client) writePump() {
	ticker := time.NewTicker(54 * time.Second)
	defer func() {
		ticker.Stop()
		c.conn.Close()
	}()

	for {
		select {
		case msg, ok := <-c.send:
			c.conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if !ok {
				c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			if err := c.conn.WriteJSON(msg); err != nil {
				return
			}
		case <-ticker.C:
			c.conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

// handleJoinRoom handles room join requests
func (c *Client) handleJoinRoom(event map[string]interface{}) {
	roomID, _ := event["room_id"].(string)
	userID, _ := event["user_id"].(string)

	if roomID == "" || userID == "" {
		c.send <- map[string]interface{}{"error": "missing room_id or user_id"}
		return
	}

	room, peer, err := c.server.hub.JoinRoom(roomID, userID)
	if err != nil {
		c.send <- map[string]interface{}{"error": err.Error()}
		return
	}

	c.roomID = roomID
	c.peerID = peer.ID
	c.userID = userID

	// Create peer connection
	peerConn, err := c.server.connManager.CreatePeerConnection(peer.ID)
	if err != nil {
		c.send <- map[string]interface{}{"error": err.Error()}
		return
	}

	// Send confirmation
	c.send <- map[string]interface{}{
		"type":    "joined_room",
		"room_id": roomID,
		"peer_id": peer.ID,
		"peers":   getPeerInfoList(room),
	}

	// Setup ICE candidates
	peerConn.OnICECandidate(func(candidate *webrtc.ICECandidate) {
		if candidate != nil {
			c.send <- map[string]interface{}{
				"type": "ice_candidate",
				"candidate": map[string]interface{}{
					"candidate":       candidate.Candidate,
					"sdp_mline_index": candidate.SDPMLineIndex,
					"sdp_mid":         candidate.SDPMid,
				},
			}
		}
	})
}

// handleLeaveRoom handles room leave requests
func (c *Client) handleLeaveRoom(event map[string]interface{}) {
	if c.roomID == "" {
		return
	}

	c.server.hub.LeaveRoom(c.roomID, c.peerID)
	c.server.connManager.ClosePeerConnection(c.peerID)
	c.roomID = ""
	c.peerID = ""
}

// handleOffer handles WebRTC offers
func (c *Client) handleOffer(event map[string]interface{}) {
	sdpData, _ := event["sdp"].(string)
	offer := webrtc.SessionDescription{
		Type: webrtc.SDPTypeOffer,
		SDP:  sdpData,
	}

	err := c.server.connManager.SetRemoteDescription(c.peerID, offer)
	if err != nil {
		c.send <- map[string]interface{}{"error": err.Error()}
		return
	}

	answer, err := c.server.connManager.CreateAnswer(c.peerID, offer)
	if err != nil {
		c.send <- map[string]interface{}{"error": err.Error()}
		return
	}

	c.send <- map[string]interface{}{
		"type": "answer",
		"sdp":  answer.SDP,
	}
}

// handleAnswer handles WebRTC answers
func (c *Client) handleAnswer(event map[string]interface{}) {
	sdpData, _ := event["sdp"].(string)
	answer := webrtc.SessionDescription{
		Type: webrtc.SDPTypeAnswer,
		SDP:  sdpData,
	}

	c.server.connManager.SetRemoteDescription(c.peerID, answer)
}

// handleICECandidate handles ICE candidates
func (c *Client) handleICECandidate(event map[string]interface{}) {
	candidateData, _ := event["candidate"].(map[string]interface{})
	candidate := peer.ICECandidate{
		Candidate: candidateData["candidate"].(string),
		SDPMid:    candidateData["sdp_mid"].(string),
	}

	c.server.connManager.AddICECandidate(c.peerID, candidate)
}

// handleMessage handles text messages
func (c *Client) handleMessage(event map[string]interface{}) {
	content, _ := event["content"].(string)

	msg := message.NewMessage(message.TextMessage, c.peerID, c.roomID, content)
	c.server.msgQueue.EnqueueMessage(c.roomID, msg)

	// Broadcast to room
	c.server.broadcast <- map[string]interface{}{
		"type":    "message",
		"message": msg,
	}
}

// handleHealth handles health checks
func (s *Server) handleHealth(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"status": "ok",
	})
}

// handleRooms handles room listing
func (s *Server) handleRooms(w http.ResponseWriter, r *http.Request) {
	rooms := s.hub.ListRooms()
	roomInfos := make([]map[string]interface{}, len(rooms))

	for i, room := range rooms {
		roomInfos[i] = map[string]interface{}{
			"id":    room.ID,
			"name":  room.Name,
			"peers": len(room.Peers),
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(roomInfos)
}

// handleIndex serves the index page
func (s *Server) handleIndex(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html")
	fmt.Fprint(w, indexHTML)
}

// closeClient closes a client connection
func (s *Server) closeClient(client *Client) {
	s.unregister <- client
}

// getPeerInfoList returns information about all peers in a room
func getPeerInfoList(room *signal.Room) []map[string]string {
	var peers []map[string]string
	room.mu.RLock()
	defer room.mu.RUnlock()

	for _, peer := range room.Peers {
		peers = append(peers, map[string]string{
			"id":   peer.ID,
			"user": peer.UserID,
		})
	}
	return peers
}

const indexHTML = `<!DOCTYPE html>
<html>
<head>
	<title>WebRTC Messaging</title>
	<style>
		body { font-family: Arial; margin: 20px; }
		input { margin: 10px 0; padding: 8px; width: 300px; }
		button { padding: 8px 20px; cursor: pointer; }
		#video { max-width: 400px; margin: 10px 0; }
	</style>
</head>
<body>
	<h1>WebRTC Real-time Messaging</h1>
	<input type="text" id="userIdInput" placeholder="Enter your name">
	<input type="text" id="roomIdInput" placeholder="Enter room ID">
	<button onclick="joinRoom()">Join Room</button>
	<video id="video" autoplay playsinline></video>
	<div id="messages"></div>
	<script>
		// WebRTC implementation will go here
		console.log("WebRTC messaging app loaded");
	</script>
</body>
</html>
`

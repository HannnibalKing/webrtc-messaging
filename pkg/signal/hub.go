package signal

import (
	"fmt"
	"sync"

	"github.com/google/uuid"
	"github.com/pion/webrtc/v3"
)

// Room represents a communication room with multiple peers
type Room struct {
	ID        string
	Name      string
	Peers     map[string]*Peer
	mu        sync.RWMutex
	CreatedAt int64
	MaxPeers  int
}

// Peer represents a peer in the room
type Peer struct {
	ID             string
	UserID         string
	PeerConnection *webrtc.PeerConnection
	DataChannels   map[string]*webrtc.DataChannel
	mu             sync.RWMutex
}

// Hub manages all rooms and peers
type Hub struct {
	rooms    map[string]*Room
	mu       sync.RWMutex
	maxPeers int

	// Channels for peer events
	PeerJoined  chan *PeerEvent
	PeerLeft    chan *PeerEvent
	RoomCreated chan *RoomEvent
	RoomClosed  chan *RoomEvent

	// ICE server configuration
	ICEServers []webrtc.ICEServer
}

// PeerEvent represents a peer join/leave event
type PeerEvent struct {
	RoomID string
	PeerID string
	UserID string
}

// RoomEvent represents a room creation/closure event
type RoomEvent struct {
	RoomID string
	Name   string
}

// NewHub creates a new signaling hub
func NewHub(maxPeers int, iceServers []webrtc.ICEServer) *Hub {
	return &Hub{
		rooms:       make(map[string]*Room),
		maxPeers:    maxPeers,
		PeerJoined:  make(chan *PeerEvent, 100),
		PeerLeft:    make(chan *PeerEvent, 100),
		RoomCreated: make(chan *RoomEvent, 100),
		RoomClosed:  make(chan *RoomEvent, 100),
		ICEServers:  iceServers,
	}
}

// CreateRoom creates a new room
func (h *Hub) CreateRoom(name string, maxPeers int) (*Room, error) {
	h.mu.Lock()
	defer h.mu.Unlock()

	roomID := uuid.New().String()
	room := &Room{
		ID:       roomID,
		Name:     name,
		Peers:    make(map[string]*Peer),
		MaxPeers: maxPeers,
	}

	h.rooms[roomID] = room
	h.RoomCreated <- &RoomEvent{RoomID: roomID, Name: name}

	return room, nil
}

// GetRoom retrieves a room by ID
func (h *Hub) GetRoom(roomID string) (*Room, error) {
	h.mu.RLock()
	defer h.mu.RUnlock()

	room, exists := h.rooms[roomID]
	if !exists {
		return nil, fmt.Errorf("room not found: %s", roomID)
	}

	return room, nil
}

// JoinRoom adds a peer to a room
func (h *Hub) JoinRoom(roomID, userID string) (*Room, *Peer, error) {
	h.mu.RLock()
	room, exists := h.rooms[roomID]
	h.mu.RUnlock()

	if !exists {
		return nil, nil, fmt.Errorf("room not found: %s", roomID)
	}

	room.mu.Lock()
	defer room.mu.Unlock()

	if len(room.Peers) >= room.MaxPeers {
		return nil, nil, fmt.Errorf("room is full")
	}

	peerID := uuid.New().String()
	peer := &Peer{
		ID:           peerID,
		UserID:       userID,
		DataChannels: make(map[string]*webrtc.DataChannel),
	}

	room.Peers[peerID] = peer

	h.PeerJoined <- &PeerEvent{RoomID: roomID, PeerID: peerID, UserID: userID}

	return room, peer, nil
}

// LeaveRoom removes a peer from a room
func (h *Hub) LeaveRoom(roomID, peerID string) error {
	h.mu.RLock()
	room, exists := h.rooms[roomID]
	h.mu.RUnlock()

	if !exists {
		return fmt.Errorf("room not found: %s", roomID)
	}

	room.mu.Lock()
	defer room.mu.Unlock()

	peer, peerExists := room.Peers[peerID]
	if !peerExists {
		return fmt.Errorf("peer not found: %s", peerID)
	}

	delete(room.Peers, peerID)

	h.PeerLeft <- &PeerEvent{RoomID: roomID, PeerID: peerID, UserID: peer.UserID}

	// Close peer connection
	if peer.PeerConnection != nil {
		peer.PeerConnection.Close()
	}

	// Close room if empty
	if len(room.Peers) == 0 {
		h.mu.Lock()
		delete(h.rooms, roomID)
		h.mu.Unlock()
		h.RoomClosed <- &RoomEvent{RoomID: roomID}
	}

	return nil
}

// ListRooms lists all active rooms
func (h *Hub) ListRooms() []*Room {
	h.mu.RLock()
	defer h.mu.RUnlock()

	rooms := make([]*Room, 0, len(h.rooms))
	for _, room := range h.rooms {
		rooms = append(rooms, room)
	}

	return rooms
}

// GetRoomPeers gets all peers in a room
func (h *Hub) GetRoomPeers(roomID string) ([]*Peer, error) {
	room, err := h.GetRoom(roomID)
	if err != nil {
		return nil, err
	}

	room.mu.RLock()
	defer room.mu.RUnlock()

	peers := make([]*Peer, 0, len(room.Peers))
	for _, peer := range room.Peers {
		peers = append(peers, peer)
	}

	return peers, nil
}

// BroadcastToRoom sends a message to all peers in a room except sender
func (h *Hub) BroadcastToRoom(roomID, senderPeerID string, data []byte) error {
	peers, err := h.GetRoomPeers(roomID)
	if err != nil {
		return err
	}

	for _, peer := range peers {
		if peer.ID == senderPeerID {
			continue
		}

		peer.mu.RLock()
		for _, dc := range peer.DataChannels {
			if dc != nil && dc.ReadyState() == webrtc.DataChannelStateOpen {
				dc.Send(webrtc.DataChannelMessage{Data: data})
			}
		}
		peer.mu.RUnlock()
	}

	return nil
}

// Close closes the hub and all rooms
func (h *Hub) Close() error {
	h.mu.Lock()
	defer h.mu.Unlock()

	for _, room := range h.rooms {
		room.mu.Lock()
		for _, peer := range room.Peers {
			if peer.PeerConnection != nil {
				peer.PeerConnection.Close()
			}
		}
		room.mu.Unlock()
	}

	h.rooms = make(map[string]*Room)

	return nil
}

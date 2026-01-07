package peer

import (
	"encoding/json"
	"fmt"
	"sync"

	"github.com/pion/webrtc/v3"
)

// ConnectionManager manages WebRTC peer connections
type ConnectionManager struct {
	peerConnections map[string]*webrtc.PeerConnection
	mu              sync.RWMutex
	mediaEngine     *webrtc.MediaEngine
	settingEngine   *webrtc.SettingEngine
}

// SignalingMessage represents a WebRTC signaling message
type SignalingMessage struct {
	Type      string        `json:"type"` // offer, answer, ice-candidate
	SDP       string        `json:"sdp,omitempty"`
	Candidate *ICECandidate `json:"candidate,omitempty"`
	PeerID    string        `json:"peer_id"`
}

// ICECandidate represents an ICE candidate
type ICECandidate struct {
	Candidate     string  `json:"candidate"`
	SDPMLineIndex *uint16 `json:"sdp_mline_index"`
	SDPMid        string  `json:"sdp_mid"`
}

// NewConnectionManager creates a new connection manager
func NewConnectionManager(iceServers []webrtc.ICEServer) (*ConnectionManager, error) {
	mediaEngine := &webrtc.MediaEngine{}

	// Register default codecs
	if err := mediaEngine.RegisterDefaultCodecs(); err != nil {
		return nil, err
	}

	settingEngine := &webrtc.SettingEngine{
		LoggerFactory: webrtc.NewDefaultLoggerFactory(),
	}

	// Add ICE servers
	for _, server := range iceServers {
		settingEngine.AddICEServer(server)
	}

	return &ConnectionManager{
		peerConnections: make(map[string]*webrtc.PeerConnection),
		mediaEngine:     mediaEngine,
		settingEngine:   settingEngine,
	}, nil
}

// CreatePeerConnection creates a new peer connection
func (cm *ConnectionManager) CreatePeerConnection(peerID string) (*webrtc.PeerConnection, error) {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	// Check if connection already exists
	if _, exists := cm.peerConnections[peerID]; exists {
		return nil, fmt.Errorf("peer connection already exists for peer: %s", peerID)
	}

	api := webrtc.NewAPI(
		webrtc.WithMediaEngine(cm.mediaEngine),
		webrtc.WithSettingEngine(cm.settingEngine),
	)

	peerConnection, err := api.NewPeerConnection(webrtc.Configuration{})
	if err != nil {
		return nil, err
	}

	cm.peerConnections[peerID] = peerConnection

	return peerConnection, nil
}

// GetPeerConnection retrieves a peer connection
func (cm *ConnectionManager) GetPeerConnection(peerID string) (*webrtc.PeerConnection, error) {
	cm.mu.RLock()
	defer cm.mu.RUnlock()

	pc, exists := cm.peerConnections[peerID]
	if !exists {
		return nil, fmt.Errorf("peer connection not found for peer: %s", peerID)
	}

	return pc, nil
}

// ClosePeerConnection closes a peer connection
func (cm *ConnectionManager) ClosePeerConnection(peerID string) error {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	pc, exists := cm.peerConnections[peerID]
	if !exists {
		return fmt.Errorf("peer connection not found for peer: %s", peerID)
	}

	delete(cm.peerConnections, peerID)
	return pc.Close()
}

// CreateOffer creates a WebRTC offer
func (cm *ConnectionManager) CreateOffer(peerID string) (*SignalingMessage, error) {
	pc, err := cm.GetPeerConnection(peerID)
	if err != nil {
		return nil, err
	}

	offer, err := pc.CreateOffer(nil)
	if err != nil {
		return nil, err
	}

	if err = pc.SetLocalDescription(offer); err != nil {
		return nil, err
	}

	return &SignalingMessage{
		Type:   "offer",
		SDP:    offer.SDP,
		PeerID: peerID,
	}, nil
}

// CreateAnswer creates a WebRTC answer
func (cm *ConnectionManager) CreateAnswer(peerID string, offer webrtc.SessionDescription) (*SignalingMessage, error) {
	pc, err := cm.GetPeerConnection(peerID)
	if err != nil {
		return nil, err
	}

	if err = pc.SetRemoteDescription(offer); err != nil {
		return nil, err
	}

	answer, err := pc.CreateAnswer(nil)
	if err != nil {
		return nil, err
	}

	if err = pc.SetLocalDescription(answer); err != nil {
		return nil, err
	}

	return &SignalingMessage{
		Type:   "answer",
		SDP:    answer.SDP,
		PeerID: peerID,
	}, nil
}

// AddICECandidate adds an ICE candidate
func (cm *ConnectionManager) AddICECandidate(peerID string, candidate ICECandidate) error {
	pc, err := cm.GetPeerConnection(peerID)
	if err != nil {
		return err
	}

	candidateInit := webrtc.ICECandidateInit{
		Candidate:     candidate.Candidate,
		SDPMLineIndex: candidate.SDPMLineIndex,
		SDPMid:        candidate.SDPMid,
	}

	return pc.AddICECandidate(candidateInit)
}

// SetRemoteDescription sets the remote session description
func (cm *ConnectionManager) SetRemoteDescription(peerID string, description webrtc.SessionDescription) error {
	pc, err := cm.GetPeerConnection(peerID)
	if err != nil {
		return err
	}

	return pc.SetRemoteDescription(description)
}

// MarshalSignalingMessage marshals a signaling message to JSON
func MarshalSignalingMessage(msg *SignalingMessage) ([]byte, error) {
	return json.Marshal(msg)
}

// UnmarshalSignalingMessage unmarshals a JSON signaling message
func UnmarshalSignalingMessage(data []byte) (*SignalingMessage, error) {
	var msg SignalingMessage
	if err := json.Unmarshal(data, &msg); err != nil {
		return nil, err
	}
	return &msg, nil
}

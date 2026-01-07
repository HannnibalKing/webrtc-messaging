# WebRTC System - Implementation Examples

## Web Client Integration

### Basic HTML Page

```html
<!DOCTYPE html>
<html>
<head>
    <title>WebRTC Chat</title>
</head>
<body>
    <div id="app">
        <input id="userName" placeholder="Your name">
        <input id="roomId" placeholder="Room ID">
        <button onclick="joinRoom()">Join</button>
        <video id="localVideo" autoplay muted></video>
        <div id="remoteVideos"></div>
        <div id="messages"></div>
        <input id="message" placeholder="Type message">
        <button onclick="sendMessage()">Send</button>
    </div>
    <script src="js/webrtc.js"></script>
    <script src="js/client.js"></script>
</body>
</html>
```

### Minimal JavaScript Example

```javascript
const webrtc = new WebRTCManager();
let ws;

async function joinRoom() {
    const userName = document.getElementById('userName').value;
    const roomId = document.getElementById('roomId').value;

    // Get local media
    await webrtc.getLocalStream(true, true);
    document.getElementById('localVideo').srcObject = webrtc.localStream;

    // Connect to WebSocket
    ws = new WebSocket('ws://localhost:8080/ws');
    ws.onmessage = (e) => handleMessage(JSON.parse(e.data));

    // Join room
    ws.send(JSON.stringify({
        type: 'join_room',
        room_id: roomId,
        user_id: userName
    }));
}

function handleMessage(msg) {
    if (msg.type === 'offer') {
        webrtc.createAnswer(msg.peer_id, msg.sdp)
            .then(answer => {
                ws.send(JSON.stringify({
                    type: 'answer',
                    peer_id: msg.peer_id,
                    sdp: answer.sdp
                }));
            });
    }
}

function sendMessage() {
    const content = document.getElementById('message').value;
    ws.send(JSON.stringify({
        type: 'message',
        content: content
    }));
}
```

## Go Client Examples

### Simple Client Connection

```go
package main

import (
    "log"
    "net/url"

    "github.com/gorilla/websocket"
)

func main() {
    u := url.URL{Scheme: "ws", Host: "localhost:8080", Path: "/ws"}
    
    conn, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
    if err != nil {
        log.Fatal(err)
    }
    defer conn.Close()

    // Join room
    conn.WriteJSON(map[string]string{
        "type": "join_room",
        "room_id": "test-room",
        "user_id": "alice",
    })

    // Read messages
    for {
        var msg map[string]interface{}
        conn.ReadJSON(&msg)
        log.Printf("Message: %v\n", msg)
    }
}
```

### WebRTC Peer Connection Handling

```go
package main

import (
    "github.com/pion/webrtc/v3"
    "github.com/yourusername/webrtc-messaging/pkg/peer"
)

func setupPeerConnection() {
    connManager, err := peer.NewConnectionManager([]webrtc.ICEServer{
        {URLs: []string{"stun:stun.l.google.com:19302"}},
    })
    if err != nil {
        log.Fatal(err)
    }

    // Create peer connection
    pc, err := connManager.CreatePeerConnection("peer-123")
    if err != nil {
        log.Fatal(err)
    }

    // Handle ICE candidates
    pc.OnICECandidate(func(candidate *webrtc.ICECandidate) {
        if candidate != nil {
            log.Printf("New ICE candidate: %s", candidate.Candidate)
            // Send to peer via signaling server
        }
    })

    // Handle remote stream
    pc.OnTrack(func(tr *webrtc.TrackRemote, receiver *webrtc.RTPReceiver) {
        log.Printf("Track received: %s %s", tr.Kind(), tr.Codec().MimeType)
    })
}
```

### Creating an Offer

```go
func createPeerConnection(connManager *peer.ConnectionManager, peerId string) error {
    // Create offer
    offer, err := connManager.CreateOffer(peerId)
    if err != nil {
        return err
    }

    // Send offer to peer via WebSocket
    ws.WriteJSON(map[string]interface{}{
        "type": "offer",
        "peer_id": peerId,
        "sdp": offer.SDP,
    })

    return nil
}
```

### Receiving and Answering Offers

```go
func handleOffer(connManager *peer.ConnectionManager, offer map[string]interface{}) error {
    peerId := offer["peer_id"].(string)
    sdp := offer["sdp"].(string)

    // Create answer
    answer, err := connManager.CreateAnswer(peerId, webrtc.SessionDescription{
        Type: webrtc.SDPTypeOffer,
        SDP:  sdp,
    })
    if err != nil {
        return err
    }

    // Send answer back
    ws.WriteJSON(map[string]interface{}{
        "type": "answer",
        "peer_id": peerId,
        "sdp": answer.SDP,
    })

    return nil
}
```

## Integration Patterns

### Slack-like Integration

```go
// In your Slack bot handler
func handleSlackCommand(cmd string) {
    room := createSignalingRoom(cmd)
    shareLink := fmt.Sprintf("https://yourapp.com?room=%s", room.ID)
    
    // Send link to Slack
    sendToSlack(shareLink)
}
```

### Discord Bot Integration

```go
// Discord bot command
func discordVideoCommand(s *discordgo.Session, m *discordgo.MessageCreate) {
    room, err := hub.CreateRoom("discord-meeting", 10)
    if err != nil {
        return
    }
    
    link := fmt.Sprintf("https://yourapp.com?room=%s", room.ID)
    s.ChannelMessageSend(m.ChannelID, fmt.Sprintf(
        "Click to join video call: %s", link))
}
```

### Mobile Web Integration

```html
<!-- Responsive design for mobile -->
<meta name="viewport" content="width=device-width, initial-scale=1.0">
<script>
    // Request full-screen for mobile
    document.getElementById('videoContainer')
        .addEventListener('click', () => {
            element.requestFullscreen?.();
        });
</script>
```

### Authentication Middleware

```go
func authMiddleware(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        token := r.Header.Get("Authorization")
        if !isValidToken(token) {
            http.Error(w, "Unauthorized", http.StatusUnauthorized)
            return
        }
        next.ServeHTTP(w, r)
    })
}

// Wrap WebSocket handler
mux.HandleFunc("/ws", authMiddleware(wsHandler))
```

## Performance Optimization

### Bandwidth Optimization

```javascript
// Reduce video quality for slow connections
const constraints = {
    video: {
        width: { ideal: 320 },  // Low resolution
        height: { ideal: 240 }
    },
    audio: true
};
```

### Media Server-Side Filtering

```go
func filterStatsReport(report webrtc.StatsReport) {
    if report.Type == "inbound-rtp" {
        inboundStats := report.*(webrtc.InboundRTPStreamStats)
        if inboundStats.PacketsLost > threshold {
            // Reduce quality
            log.Println("High packet loss, reducing quality")
        }
    }
}
```

### Connection Pooling

```go
type ConnectionPool struct {
    conns chan *webrtc.PeerConnection
    max   int
}

func (p *ConnectionPool) Get() (*webrtc.PeerConnection, error) {
    select {
    case conn := <-p.conns:
        return conn, nil
    default:
        return webrtc.NewPeerConnection(webrtc.Configuration{})
    }
}
```

## Monitoring and Logging

### Custom Metrics Collection

```go
type Metrics struct {
    ActiveRooms  int64
    ActivePeers  int64
    MessagesQty  int64
    AveLatency   float64
}

func (h *Hub) CollectMetrics() Metrics {
    h.mu.RLock()
    defer h.mu.RUnlock()
    
    peers := 0
    for _, room := range h.rooms {
        peers += len(room.Peers)
    }
    
    return Metrics{
        ActiveRooms: int64(len(h.rooms)),
        ActivePeers: int64(peers),
    }
}
```

### Prometheus Integration

```go
import "github.com/prometheus/client_golang/prometheus"

var (
    activeRooms = prometheus.NewGauge(prometheus.GaugeOpts{
        Name: "webrtc_active_rooms",
    })
    activePeers = prometheus.NewGauge(prometheus.GaugeOpts{
        Name: "webrtc_active_peers",
    })
)

func (h *Hub) recordMetrics() {
    activeRooms.Set(float64(len(h.rooms)))
    peers := 0
    for _, room := range h.rooms {
        peers += len(room.Peers)
    }
    activePeers.Set(float64(peers))
}
```

## Testing

### Unit Test Example

```go
package signal

import (
    "testing"
)

func TestHubCreateRoom(t *testing.T) {
    hub := NewHub(10, []webrtc.ICEServer{})
    
    room, err := hub.CreateRoom("test", 5)
    if err != nil {
        t.Fatal(err)
    }
    
    if room.Name != "test" {
        t.Errorf("Expected name 'test', got '%s'", room.Name)
    }
}

func TestHubJoinRoom(t *testing.T) {
    hub := NewHub(10, []webrtc.ICEServer{})
    room, _ := hub.CreateRoom("test", 5)
    
    _, peer, err := hub.JoinRoom(room.ID, "alice")
    if err != nil {
        t.Fatal(err)
    }
    
    if peer.UserID != "alice" {
        t.Errorf("Expected user 'alice', got '%s'", peer.UserID)
    }
}
```

### WebSocket Client Test

```go
func TestWebSocketConnection(t *testing.T) {
    // Start test server
    server := httptest.NewServer(
        http.HandlerFunc(wsHandler),
    )
    defer server.Close()

    // Connect client
    wsURL := "ws" + strings.TrimPrefix(server.URL, "http")
    conn, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
    if err != nil {
        t.Fatal(err)
    }
    defer conn.Close()

    // Test join room
    conn.WriteJSON(map[string]string{
        "type": "join_room",
        "room_id": "test",
        "user_id": "alice",
    })

    var response map[string]interface{}
    conn.ReadJSON(&response)

    if response["type"] != "joined_room" {
        t.Error("Expected joined_room response")
    }
}
```

## Deployment Scripts

### Build Script (build.sh)

```bash
#!/bin/bash
set -e

echo "Building WebRTC server..."
go build -o bin/server cmd/server/main.go

echo "Building CLI client..."
go build -o bin/client cmd/client/main.go

echo "Build complete!"
```

### Run Script (run.sh)

```bash
#!/bin/bash

# Load environment
source .env

# Start server
./bin/server &
SERVER_PID=$!

echo "Server started with PID $SERVER_PID"
echo "Access at http://localhost:$SERVER_PORT"

# Keep running
wait $SERVER_PID
```

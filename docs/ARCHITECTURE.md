# WebRTC System Architecture

## High-Level Overview

The WebRTC messaging system is a distributed, peer-to-peer (P2P) real-time communication platform built with Go on the backend and JavaScript on the frontend.

```
┌─────────────────────────────────────────────────────────┐
│                    Web Browsers                          │
│  ┌──────────────┐  ┌──────────────┐  ┌──────────────┐  │
│  │  Client A    │  │  Client B    │  │  Client C    │  │
│  │ (WebRTC)     │  │ (WebRTC)     │  │ (WebRTC)     │  │
│  └──────┬───────┘  └──────┬───────┘  └──────┬───────┘  │
│         │                 │                  │          │
│         └─────────────────┼──────────────────┘          │
│                           │                             │
└───────────────────────────┼─────────────────────────────┘
                            │ WebSocket
                     ┌──────▼──────┐
                     │   Signaling │
                     │   Server    │
                     │   (Go)      │
                     └──────┬──────┘
                            │
                    ┌───────┴───────┐
                    │               │
              ┌─────▼────┐    ┌─────▼────┐
              │ Signal   │    │ Message  │
              │ Manager  │    │ Router   │
              └──────────┘    └──────────┘
```

## Component Architecture

### 1. Frontend (Web Client)

**Location:** `web/`

**Components:**
- `index.html` - UI layout
- `css/style.css` - Styling
- `js/webrtc.js` - WebRTC peer connection management
- `js/client.js` - Application logic and signaling

**Technologies:**
- HTML5 Web APIs
- WebRTC API (RTCPeerConnection, RTCDataChannel)
- WebSocket API
- Modern CSS Grid/Flexbox

**Key Features:**
- Real-time video/audio streaming
- Text messaging
- Peer discovery and connection
- Statistics monitoring

### 2. Signaling Server (Go Backend)

**Location:** `cmd/server/main.go`

**Responsibilities:**
- Accept WebSocket connections
- Manage room membership
- Relay WebRTC signaling messages (offers, answers, ICE candidates)
- Broadcast messages between peers

**Key Components:**

#### Signal Hub (`pkg/signal/hub.go`)
- Manages all active rooms
- Tracks peers in each room
- Handles room creation/destruction
- Broadcasting to peers

```go
Hub
├── Room (per room)
│   ├── Peer A
│   ├── Peer B
│   └── Peer C
├── Room (per room)
└── ...
```

#### Peer Connection Manager (`pkg/peer/connection.go`)
- Creates and manages WebRTC peer connections
- Handles SDP offer/answer negotiation
- Processes ICE candidates
- Manages connection state

#### Message Router (`pkg/message/router.go`)
- Queues messages for delivery
- Routes messages to appropriate handlers
- Supports message types: text, system, data

#### Media Stream Manager (`pkg/media/stream.go`)
- Configures audio/video codecs
- Handles media track management
- Provides connection statistics

#### HTTP/WebSocket Server (`pkg/server/handler.go`)
- Handles WebSocket upgrades
- Processes client messages
- Sends server-initiated messages
- REST API endpoints

### 3. Data Flow

#### Connection Establishment

```
Client A                    Signaling Server              Client B
   │                              │                          │
   ├─────── join_room ──────────→ │                          │
   │                              │ (Create Room)            │
   │                              │                          │
   │                              │ ←──── join_room ──────── │
   │                              │                          │
   │ ←─ joined_room ────────────── │ (Sends existing peers)  │
   │                              │                          │
   │ (Create PC) ─────┐           │  ┌─── joined_room ──────→│
   │                 │           │  │ (Sends existing peers) │
   │                 ▼           │  │                        │
   ├──── offer (B's peer_id) ──→ │ ──────→ offer ──────────→ │
   │                              │       (A's peer_id)      │
   │                              │                        (Create PC)
   │ ←─ answer (B's peer_id) ─────┤ ←────── answer ───────── │
   │                              │
   ├── ice_candidate ────────────→ │ ─────→ ice_candidate ──→ │
   │                              │
   │ ←─ ice_candidate ───────────── ←──── ice_candidate ──── │
   │                              │
   ▼                              ▼                          ▼
(P2P connection established - no further server involvement)
```

#### Message Exchange

```
Client A                    Signaling Server              Client B
   │                              │                          │
   ├─ message (content) ────────→ │                          │
   │                              │ (Queue message)          │
   │                              │                          │
   │                              │ ├─ message ────────────→ │
   │                              │ │ (Client C)             │
   │                              │ └─ message ────────────→ │
   │                              │   (Other peers)          │
   │                              │                          │
   ▼                              ▼                          ▼
(Message appears in chat for all peers)
```

## Package Organization

```
webrtc-messaging/
├── cmd/
│   ├── server/
│   │   └── main.go          # Server entry point
│   └── client/
│       └── main.go          # CLI client example
├── pkg/
│   ├── signal/
│   │   └── hub.go           # Room & peer management
│   ├── peer/
│   │   └── connection.go    # WebRTC peer connections
│   ├── media/
│   │   └── stream.go        # Media codec handling
│   ├── message/
│   │   └── router.go        # Message routing
│   └── server/
│       └── handler.go       # HTTP/WebSocket handlers
├── config/
│   └── config.go            # Configuration management
├── web/
│   ├── index.html           # Web UI
│   ├── css/
│   │   └── style.css        # Styling
│   └── js/
│       ├── webrtc.js        # WebRTC utilities
│       └── client.js        # Client logic
├── docs/
│   ├── GETTING_STARTED.md
│   ├── API_REFERENCE.md
│   ├── ARCHITECTURE.md
│   ├── EXAMPLES.md
│   └── DEPLOYMENT.md
└── config files
    ├── go.mod
    ├── go.sum
    ├── .env.example
    ├── Dockerfile
    └── docker-compose.yml
```

## Technology Stack

### Backend
- **Language:** Go 1.21+
- **WebRTC:** Pion WebRTC v3
- **Networking:** Gorilla WebSocket
- **Configuration:** godotenv

### Frontend
- **Core:** HTML5, CSS3, JavaScript (ES6+)
- **WebRTC:** Browser native WebRTC API
- **Protocol:** WebSocket

### Infrastructure
- **Containerization:** Docker
- **Orchestration:** Docker Compose (optional Kubernetes)
- **Networking:** Standard HTTP/WebSocket ports

## Concurrency Model

### Server-Side

**Thread Safety:**
- `sync.RWMutex` protects shared data structures
- Separate goroutines per WebSocket client
- Non-blocking channel operations

```go
Hub {
    rooms: map[string]*Room  // Protected by mu
    mu:    sync.RWMutex
    
    // Event channels (thread-safe)
    PeerJoined chan *PeerEvent
    PeerLeft   chan *PeerEvent
}
```

**Message Handling:**
```
For each WebSocket client:
  ├── readPump()  // Read messages from client
  ├── writePump() // Write messages to client
  └── handlers    // Process message types
```

### Client-Side

**Asynchronous:** All operations use JavaScript Promises/async-await

```javascript
async function joinRoom() {
    const stream = await webrtc.getLocalStream()
    await webrtc.createOffer(peerId)
    // ...
}
```

## Security Considerations

### Media Encryption
- **DTLS (Datagram Transport Layer Security):** Encrypts media streams
- **SRTP (Secure Real-time Transport Protocol):** Encrypts audio/video
- Implemented by browser and Pion WebRTC library

### Signaling Security
- ⚠️ **Development:** Plain WebSocket (HTTP)
- 🔐 **Production:** Use WSS (HTTPS)
- Implement authentication/authorization

### Privacy
- No server-side media recording
- P2P media transfer (server is unaware of media content)
- Server only sees signaling (SDP, ICE candidates, metadata)

## Scalability Architecture

### Vertical Scaling
- Increase server resources (CPU, RAM)
- Limit by: Room count, peer connections per room

### Horizontal Scaling
- Load balancing across multiple signaling servers
- Sticky sessions for WebSocket connections
- Shared message broker (Redis) for inter-server communication

### Typical Capacity (Per Server)
- **Rooms:** 1,000+ concurrent rooms
- **Peers:** 10,000+ concurrent peers (with P2P media)
- **Bandwidth:** Limited by infrastructure (100 Mbps typical)
- **Memory:** ~50MB per 100 peers

### Scaling Bottlenecks
1. **CPU:** Video encoding (if server processes media)
2. **Memory:** Peer connection management
3. **Bandwidth:** ICE candidate exchange, signaling messages
4. **TURN Server:** NAT traversal capacity

## Future Architecture Enhancements

```
Proposed Scalable Architecture:

┌─────────────────────────────────────────┐
│         Load Balancer (sticky)          │
└────────────┬────────────────────────────┘
             │
    ┌────────┼────────┐
    │        │        │
┌───▼──┐ ┌──▼───┐ ┌──▼───┐
│ SRV1 │ │ SRV2 │ │ SRV3 │  (Signaling Servers)
└───┬──┘ └──┬───┘ └──┬───┘
    │       │       │
    └───────┼───────┘
            │
        ┌───▼────┐
        │ Redis  │  (Message Broker)
        └────────┘
            │
    ┌───────┴────────┐
    │                │
┌───▼────┐      ┌───▼────┐
│ TURN 1  │      │ TURN 2  │  (TURN Servers)
└────────┘      └────────┘
```

## Performance Characteristics

### Latency
- **Signaling:** < 100ms (typical)
- **Media RTT:** 20-100ms (depends on network)
- **Message delivery:** < 500ms

### Bandwidth Usage
- **Signaling:** ~1-10 KB per connection
- **Media (HD):** 2-4 Mbps per peer (depends on codec)
- **ICE candidates:** 1-10 KB per peer

### CPU Usage
- **Per peer connection:** ~5-10% (one core)
- **Video codec:** Scales with resolution (H264 > VP8)

## Monitoring & Observability

### Server Metrics
- Active rooms count
- Active peers count
- Message throughput
- Connection state distribution

### Client Metrics
- RTT (round-trip time)
- Bandwidth (available outgoing)
- Packets lost
- Connection state

### Logging
- Go stdlib `log` package
- Configurable log levels
- Server operation traces
- Error tracking

## Deployment Topologies

### Single Server
```
Internet → Load Balancer → Signaling Server → Application
```

### High Availability
```
Internet → Load Balancer → [Signaling Server 1]
                       ↘   [Signaling Server 2]
                           [Message Broker]
```

### Global Distribution
```
Users (Region A) → CDN/LB → [Signaling Servers A]
Users (Region B) → CDN/LB → [Signaling Servers B]
                    ↓ Federated
                [Message Broker]
```

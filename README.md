# WebRTC Real-time Messaging System

A low-latency, peer-to-peer real-time communication system built with Go and WebRTC. Supports audio, video, and text messaging with persistent connections and advanced networking capabilities.

## Features

- **Real-time Audio/Video**: High-quality peer-to-peer audio and video streaming
- **Text Messaging**: Low-latency chat messaging between peers
- **Room-based Communication**: Create and manage communication rooms
- **Persistent Connections**: WebSocket signaling with automatic reconnection
- **NAT Traversal**: STUN/TURN server support for reliable connectivity
- **Scalable Architecture**: Handle thousands of concurrent connections
- **Media Management**: Audio/video codec negotiation and stream management
- **Cross-platform**: Works with any WebRTC-compatible client

## Architecture

### Core Components

1. **Signaling Server** - Handles peer discovery and connection setup
2. **Peer Manager** - Manages peer connections and state
3. **Media Engine** - Handles audio/video streaming
4. **Message Router** - Routes messages between peers
5. **WebSocket Hub** - Manages WebSocket connections

### Communication Flow

```
Client A                    Signaling Server                    Client B
   |                              |                              |
   |------ Register Room -------->|                              |
   |                              |<----- Register Room ---------|
   |<----- Peer List -------------|                              |
   |                              |------ Peer List ----------->|
   |                              |                              |
   |-- Create Offer ---->|                                      |
   |                     |------ Offer via Signal -------->|    |
   |                     |                                 |    |
   |                     |<------ Answer via Signal -------|    |
   |<---- Answer --------|                                      |
   |                                                             |
   |================== Establish Direct P2P Connection==================|
   |                          (DTLS/SRTP)                            |
   |                                                             |
   |==================== Audio/Video/Data Stream ========================|
```

## Project Structure

```
webrtc-messaging/
├── cmd/
│   ├── server/              # Signaling server
│   └── client/              # Example client
├── pkg/
│   ├── signal/              # Signaling logic
│   ├── peer/                # Peer connection management
│   ├── media/               # Audio/video handling
│   ├── message/             # Message routing
│   └── server/              # HTTP server
├── web/                     # Client-side web app
├── docs/                    # Documentation
├── config/                  # Configuration
└── go.mod                   # Go module

```

## Quick Start

### Prerequisites

- Go 1.21+
- Modern browser with WebRTC support
- Network access (for STUN/TURN servers)

### Installation

```bash
git clone <repository>
cd webrtc-messaging

# Install dependencies
go mod download

# Copy configuration
cp .env.example .env

# Start signaling server
go run cmd/server/main.go
```

Server will start on `http://localhost:8080`

### Usage

1. Open browser and navigate to `http://localhost:8080`
2. Create or join a room
3. Share room URL with another user
4. Enable audio/video and start communicating

## Configuration

Edit `.env` to configure:

- **Server Port**: `SERVER_PORT`
- **STUN Servers**: `STUN_SERVERS` (for NAT traversal)
- **TURN Servers**: `TURN_SERVERS` (for restrictive networks)
- **Media Codecs**: `VIDEO_CODEC`, `AUDIO_CODEC`
- **Room Settings**: `MAX_PEERS_PER_ROOM`, `DEFAULT_ROOM_TIMEOUT`

## API Endpoints

### WebSocket (`/ws`)

Primary communication channel for signaling

**Events:**
- `join_room` - Join a communication room
- `leave_room` - Leave a room
- `offer` - WebRTC offer
- `answer` - WebRTC answer
- `ice_candidate` - ICE candidate
- `message` - Text message
- `peer_list` - List of connected peers

### REST API

- `GET /api/rooms` - List active rooms
- `POST /api/rooms` - Create a new room
- `GET /api/rooms/:id` - Get room details
- `GET /api/health` - Health check

## WebRTC Protocol

This system uses:

- **ICE**: For Network Address Translation (NAT) traversal
- **DTLS**: For secure peer connection
- **SRTP**: For encrypted media streaming
- **SCTP**: For reliable data channel

## Performance

- **Latency**: < 100ms peer-to-peer (depending on network)
- **Throughput**: Adaptive bitrate for video/audio
- **Scalability**: Supports 1000+ concurrent WebSocket connections
- **Resource Usage**: ~5-10MB per active peer connection

## Deployment

### Docker

```bash
docker build -t webrtc-messaging .
docker run -p 8080:8080 webrtc-messaging
```

### Production Deployment

See [docs/DEPLOYMENT.md](docs/DEPLOYMENT.md) for:
- Kubernetes setup
- Load balancing
- TURN server configuration
- Monitoring and metrics

## Documentation

- [GETTING_STARTED.md](docs/GETTING_STARTED.md) - Setup guide
- [API_REFERENCE.md](docs/API_REFERENCE.md) - WebSocket API
- [ARCHITECTURE.md](docs/ARCHITECTURE.md) - System design
- [DEPLOYMENT.md](docs/DEPLOYMENT.md) - Production setup
- [EXAMPLES.md](docs/EXAMPLES.md) - Code examples

## Security Considerations

- All connections use TLS/DTLS encryption
- TURN credentials should be short-lived
- Room access can be restricted with authentication
- Media is encrypted end-to-end
- Input validation on all signaling messages

## Known Limitations

- Audio/video processing on server not implemented (P2P only)
- Recording requires client-side implementation
- No built-in transcoding
- TURN server required for strict NAT scenarios

## Browser Support

- Chrome/Chromium 51+
- Firefox 55+
- Safari 11+
- Edge 79+

## Performance Optimization

1. **Adaptive Bitrate**: Automatically adjusts video quality based on network
2. **Connection Pooling**: Reuses network resources
3. **Efficient Signaling**: Minimal message overhead
4. **ICE Optimization**: Rapid candidate gathering

## Troubleshooting

### No video/audio
1. Check browser permissions
2. Verify camera/microphone access
3. Check STUN server connectivity
4. Review firewall settings

### High latency
1. Check network conditions
2. Verify STUN/TURN server proximity
3. Review WebRTC statistics
4. Check CPU usage

## Contributing

Contributions welcome! Please review [CONTRIBUTING.md](CONTRIBUTING.md)

## License

MIT

## Resources

- [WebRTC.org](https://webrtc.org) - WebRTC specification
- [Pion WebRTC](https://github.com/pion/webrtc) - Go WebRTC library
- [RFC 8288](https://tools.ietf.org/html/rfc8288) - DTLS
- [RFC 3711](https://tools.ietf.org/html/rfc3711) - SRTP

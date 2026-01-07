# WebRTC Real-time Messaging System - Getting Started

## Prerequisites

- Go 1.21 or higher
- A modern web browser with WebRTC support
- Git and GitHub CLI (optional, for deployment)

## Quick Start (5 minutes)

### 1. Clone the Repository

```bash
cd c:\Users\HellKnight\Documents\GitHub\Go
git clone https://github.com/yourusername/webrtc-messaging.git
cd webrtc-messaging
```

### 2. Install Dependencies

```bash
go mod download
go mod tidy
```

### 3. Configure Environment

Copy the example configuration:

```bash
copy .env.example .env
```

Edit `.env` with your settings:

```env
# Server Configuration
SERVER_HOST=localhost
SERVER_PORT=8080

# WebRTC Configuration
STUN_SERVERS=stun:stun.l.google.com:19302,stun:stun1.l.google.com:19302
TURN_SERVERS=

# WebSocket Configuration
WS_PATH=/ws
WS_READ_BUFFER=1024
WS_WRITE_BUFFER=1024

# Media Configuration
AUDIO_ENABLED=true
VIDEO_ENABLED=true
VIDEO_CODEC=VP8
AUDIO_CODEC=OPUS

# Room Configuration
MAX_PEERS_PER_ROOM=10
DEFAULT_ROOM_TIMEOUT=3600
ALLOW_ANONYMOUS=true
```

### 4. Run the Server

```bash
go run cmd/server/main.go
```

Expected output:
```
WebRTC server starting on localhost:8080
```

### 5. Open in Browser

Navigate to `http://localhost:8080` and start chatting!

## Web Client Usage

### Joining a Room

1. Enter your name in "Your Name" field
2. Enter or create a room ID
3. Click "Join Room"
4. Allow camera/microphone permissions when prompted

### Using Media Controls

- **📢 Audio ON/OFF**: Toggle microphone
- **📹 Video ON/OFF**: Toggle camera
- **🖥️ Screen Share**: Share your screen (coming soon)

### Sending Messages

1. Type your message in the chat box
2. Press Enter or click Send
3. Messages appear in real-time for all room members

### Viewing Statistics

Monitor connection quality in the stats panel:
- **Bandwidth**: Available outgoing bitrate
- **Latency**: Round-trip time
- **Packets Lost**: Number of lost packets
- **Connection State**: Current WebRTC connection state

## Running the CLI Client

### Basic Usage

```bash
go run cmd/client/main.go -server ws://localhost:8080/ws -user YourName -room test-room
```

### Command-line Options

- `-server`: WebSocket server URL (default: `ws://localhost:8080/ws`)
- `-user`: Your username (default: auto-generated)
- `-room`: Room to join (default: `test-room`)

### Example

```bash
go run cmd/client/main.go -server ws://192.168.1.100:8080/ws -user alice -room conference
```

## API Endpoints

### Health Check

```bash
curl http://localhost:8080/api/health
```

Response:
```json
{"status": "ok"}
```

### List Rooms

```bash
curl http://localhost:8080/api/rooms
```

Response:
```json
[
  {
    "id": "room-123",
    "name": "Conference Room",
    "peers": 3
  }
]
```

## Docker Deployment

### Build Image

```bash
docker build -t webrtc-messaging:latest .
```

### Run Container

```bash
docker run -p 8080:8080 \
  -e SERVER_PORT=8080 \
  -e STUN_SERVERS="stun:stun.l.google.com:19302" \
  webrtc-messaging:latest
```

### Using Docker Compose

```bash
docker-compose up -d
```

Access at `http://localhost:8080`

## Troubleshooting

### Camera/Microphone Not Working

- Check browser permissions for your site
- Ensure devices aren't in use by another application
- Try reloading the page

### No Connection to Other Peers

1. Check server is running: `curl http://localhost:8080/api/health`
2. Verify firewall allows port 8080
3. Check browser console for JavaScript errors (F12)
4. Ensure both peers are in the same room

### High Latency or Packet Loss

1. Check network connection stability
2. Reduce video resolution
3. Close other bandwidth-heavy applications
4. Consider adding TURN servers for better NAT traversal

### WebSocket Connection Fails

1. Verify server URL is correct
2. Check CORS settings if on different domain
3. Look at server logs for detailed error messages
4. Ensure WebSocket port (8080) is accessible

## Next Steps

- Read [API Reference](docs/API_REFERENCE.md) for detailed message formats
- Explore [Architecture](docs/ARCHITECTURE.md) for system design
- Check [Examples](docs/EXAMPLES.md) for integration patterns
- See [Deployment](docs/DEPLOYMENT.md) for production setup

## Performance Tuning

### Optimize Video Quality

Modify `config/.env`:

```env
# Lower resolution for better performance
VIDEO_CODEC=VP8
# Add more TURN servers for better connectivity
TURN_SERVERS=turn:your-turn-server:3478
```

### Scale Multiple Rooms

The system supports unlimited concurrent rooms limited only by:
- Server memory (each peer connection ~10-50 MB)
- CPU (video encoding/decoding)
- Bandwidth

### Monitor Performance

Check `/api/rooms` endpoint to see active rooms and peer counts.

## Logs

Logs are output to stdout. Set `LOG_LEVEL` for more/less detail:

```env
LOG_LEVEL=DEBUG  # verbose
LOG_LEVEL=INFO   # normal
LOG_LEVEL=ERROR  # errors only
```

## Security Notes

- ⚠️ The `ALLOW_ANONYMOUS=true` setting allows any user - disable in production
- 🔐 Use HTTPS/WSS in production (not plain HTTP/WS)
- 🔐 Implement authentication before exposing to internet
- 🔐 All media is encrypted via DTLS/SRTP

## Support

For issues and questions:
1. Check [Troubleshooting](#troubleshooting) section
2. Review server logs for error messages
3. Open an issue on GitHub
4. Check browser developer console (F12) for client-side errors

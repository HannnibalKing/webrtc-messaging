# WebRTC Real-time Messaging System - Quick Reference

## Starting the Server

```bash
# Development
go run cmd/server/main.go

# Production (with build)
go build -o server cmd/server/main.go
./server
```

## Web Client Access

```
http://localhost:8080
```

## Docker Quick Start

```bash
# Build
docker build -t webrtc-messaging:latest .

# Run
docker run -p 8080:8080 webrtc-messaging:latest

# Docker Compose
docker-compose up -d
```

## Configuration

Copy `.env.example` to `.env` and customize:

```bash
cp .env.example .env
```

Key settings:
- `SERVER_PORT=8080` - Server port
- `SERVER_HOST=localhost` - Server host
- `STUN_SERVERS=stun:stun.l.google.com:19302` - STUN servers
- `MAX_PEERS_PER_ROOM=10` - Max peers per room
- `LOG_LEVEL=INFO` - Logging level

## CLI Client Usage

```bash
# Start CLI client
go run cmd/client/main.go -server ws://localhost:8080/ws -user alice -room test-room
```

## API Endpoints

### Health Check
```bash
curl http://localhost:8080/api/health
```

Response: `{"status":"ok"}`

### List Rooms
```bash
curl http://localhost:8080/api/rooms
```

Response: List of active rooms with peer counts

## WebSocket Protocol

### Join Room
```json
{
  "type": "join_room",
  "room_id": "room-123",
  "user_id": "alice"
}
```

### Send Message
```json
{
  "type": "message",
  "content": "Hello!"
}
```

### WebRTC Offer
```json
{
  "type": "offer",
  "peer_id": "peer-uuid",
  "sdp": "v=0\r\no=- ..."
}
```

## Troubleshooting

### Server won't start
- Check port 8080 is available
- Verify Go installation: `go version`
- Check environment: `echo $PATH`

### No video/audio
- Browser permissions needed (check address bar)
- Camera/mic might be in use
- Try refreshing page

### Can't connect to other peers
- Verify both in same room
- Check server is running: `curl http://localhost:8080/api/health`
- Look at browser console (F12) for errors

### High latency
- Check network connection
- Try reducing video quality
- Check for packet loss in stats

## Performance Monitoring

Check stats panel in web UI:
- **Bandwidth** - Available bitrate
- **Latency** - Round-trip time
- **Packets Lost** - Connection quality
- **Connection State** - Current state

## Development Workflow

```bash
# 1. Install dependencies
go mod download

# 2. Run server in development
go run cmd/server/main.go

# 3. Build production binary
go build -o bin/server cmd/server/main.go

# 4. Run tests (when added)
go test ./...

# 5. Deploy
docker build -t webrtc-messaging .
docker run -p 8080:8080 webrtc-messaging
```

## File Organization

- `cmd/` - Executable entry points
- `pkg/` - Reusable packages
- `config/` - Configuration management
- `web/` - Web client files
- `docs/` - Documentation
- `.env.example` - Configuration template
- `Dockerfile` - Container build
- `docker-compose.yml` - Multi-container setup
- `go.mod` / `go.sum` - Dependencies

## Useful Commands

```bash
# Format code
go fmt ./...

# Run linter
golangci-lint run

# View dependencies
go mod graph

# Update dependencies
go get -u ./...

# Build for Linux
GOOS=linux GOARCH=amd64 go build -o server cmd/server/main.go

# Build for macOS
GOOS=darwin GOARCH=amd64 go build -o server cmd/server/main.go

# Build for Windows
GOOS=windows GOARCH=amd64 go build -o server.exe cmd/server/main.go
```

## Performance Tips

1. **Reduce video resolution** for better performance
2. **Enable hardware acceleration** in browser
3. **Use VP8 codec** for better compatibility
4. **Configure TURN servers** for better NAT traversal
5. **Monitor stats** to detect issues early
6. **Limit peers per room** to prevent overload
7. **Use WebSocket compression** for signaling efficiency

## Security Notes

⚠️ **Development Settings**
- `ALLOW_ANONYMOUS=true` allows anyone to join
- Plain HTTP (not HTTPS)
- No authentication required

🔐 **Production Requirements**
- Use HTTPS/WSS only
- Implement authentication (JWT tokens)
- Set `ALLOW_ANONYMOUS=false`
- Configure TURN servers
- Enable rate limiting
- Monitor and log connections

## Additional Resources

- [Getting Started](docs/GETTING_STARTED.md) - Complete setup guide
- [API Reference](docs/API_REFERENCE.md) - Message formats
- [Architecture](docs/ARCHITECTURE.md) - System design
- [Examples](docs/EXAMPLES.md) - Code samples
- [Deployment](docs/DEPLOYMENT.md) - Production guide

## Support

For issues:
1. Check [Troubleshooting](docs/GETTING_STARTED.md#troubleshooting)
2. Review server logs
3. Check browser console (F12)
4. Open GitHub issue

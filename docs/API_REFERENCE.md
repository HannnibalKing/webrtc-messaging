# WebRTC Signaling Protocol - API Reference

## Message Types

All messages are JSON objects sent over WebSocket connection.

### Server URL
```
ws://[SERVER_HOST]:[SERVER_PORT]/ws
```

## Client → Server Messages

### Join Room

Sent when a client wants to join a room.

```json
{
  "type": "join_room",
  "room_id": "room-123",
  "user_id": "alice"
}
```

**Response:**
```json
{
  "type": "joined_room",
  "room_id": "room-123",
  "peer_id": "peer-uuid",
  "peers": [
    {
      "id": "peer-uuid-1",
      "user": "bob"
    },
    {
      "id": "peer-uuid-2",
      "user": "charlie"
    }
  ]
}
```

### Leave Room

Sent when a client disconnects from a room.

```json
{
  "type": "leave_room",
  "room_id": "room-123"
}
```

### WebRTC Offer

Sent to initiate a peer connection.

```json
{
  "type": "offer",
  "peer_id": "peer-uuid",
  "sdp": "v=0\r\no=- 123456789 2 IN IP4 127.0.0.1\r\n..."
}
```

### WebRTC Answer

Sent in response to an offer.

```json
{
  "type": "answer",
  "peer_id": "peer-uuid",
  "sdp": "v=0\r\no=- 987654321 2 IN IP4 127.0.0.1\r\n..."
}
```

### ICE Candidate

Sent to exchange ICE candidates for NAT traversal.

```json
{
  "type": "ice_candidate",
  "peer_id": "peer-uuid",
  "candidate": {
    "candidate": "candidate:842163049 1 udp 1677729535 192.168.1.100 54321 typ srflx",
    "sdp_mline_index": 0,
    "sdp_mid": "0"
  }
}
```

### Text Message

Sent to broadcast a message to all peers in the room.

```json
{
  "type": "message",
  "content": "Hello everyone!"
}
```

## Server → Client Messages

### Peer Joined

Broadcast when a new peer joins the room.

```json
{
  "type": "peer_joined",
  "room_id": "room-123",
  "peer_id": "peer-uuid",
  "user_id": "charlie"
}
```

### Peer Left

Broadcast when a peer leaves the room.

```json
{
  "type": "peer_left",
  "room_id": "room-123",
  "peer_id": "peer-uuid",
  "user_id": "charlie"
}
```

### Offer (Relayed)

Relayed offer from another peer.

```json
{
  "type": "offer",
  "peer_id": "peer-uuid",
  "sdp": "v=0\r\no=- 123456789 2 IN IP4 127.0.0.1\r\n..."
}
```

### Answer (Relayed)

Relayed answer from another peer.

```json
{
  "type": "answer",
  "peer_id": "peer-uuid",
  "sdp": "v=0\r\no=- 987654321 2 IN IP4 127.0.0.1\r\n..."
}
```

### ICE Candidate (Relayed)

Relayed ICE candidate from another peer.

```json
{
  "type": "ice_candidate",
  "peer_id": "peer-uuid",
  "candidate": {
    "candidate": "candidate:842163049 1 udp 1677729535 192.168.1.100 54321 typ srflx",
    "sdp_mline_index": 0,
    "sdp_mid": "0"
  }
}
```

### Message (Relayed)

Relayed message from another peer.

```json
{
  "type": "message",
  "message": {
    "id": "msg-uuid",
    "type": "text",
    "sender_id": "peer-uuid",
    "room_id": "room-123",
    "content": "Hello everyone!",
    "timestamp": "2024-01-15T10:30:00Z"
  }
}
```

### Error

Sent when an error occurs.

```json
{
  "type": "error",
  "error": "room is full"
}
```

## REST API Endpoints

### GET /api/health

Health check endpoint.

**Response:**
```json
{
  "status": "ok"
}
```

**HTTP Status:** 200

### GET /api/rooms

List all active rooms.

**Response:**
```json
[
  {
    "id": "room-123",
    "name": "Conference",
    "peers": 3
  },
  {
    "id": "room-456",
    "name": "Study Group",
    "peers": 2
  }
]
```

**HTTP Status:** 200

## WebRTC Configuration Parameters

### Media Codecs

**Video Codecs:**
- `VP8` - Default, good compatibility
- `VP9` - Better compression, higher CPU
- `H264` - Hardware acceleration support

**Audio Codecs:**
- `OPUS` - Default, excellent quality

### ICE Servers

**STUN Servers (NAT Traversal):**
```
stun:stun.l.google.com:19302
stun:stun1.l.google.com:19302
stun:stun2.l.google.com:19302
stun:stun3.l.google.com:19302
stun:stun4.l.google.com:19302
```

**TURN Servers (Required for Symmetric NAT):**
```
turn:turnserver.example.com:3478?transport=udp
turn:turnserver.example.com:3478?transport=tcp
turns:turnserver.example.com:5349?transport=tcp
```

## Example Workflows

### Peer-to-Peer Connection Establishment

1. **Peer A joins room**
   ```
   → join_room (room-123, alice)
   ← joined_room (peers=[])
   ```

2. **Peer B joins room**
   ```
   → join_room (room-123, bob)
   ← joined_room (peers=[peer-a])
   (Peer A receives) → peer_joined (peer-b)
   ```

3. **Peer A creates offer to Peer B**
   ```
   → offer (to: peer-b, sdp: ...)
   (Peer B receives) → offer
   ```

4. **Peer B creates answer to Peer A**
   ```
   → answer (to: peer-a, sdp: ...)
   (Peer A receives) → answer
   ```

5. **ICE Candidate Exchange**
   ```
   A → ice_candidate
   (B receives) → ice_candidate
   B → ice_candidate
   (A receives) → ice_candidate
   ... (multiple candidates)
   ```

6. **Connection Established**
   - Both peers now have active peer connection
   - Can exchange data channel messages
   - Can share audio/video

### Chat Message Flow

1. **User sends message**
   ```
   → message (content: "Hello")
   ```

2. **Server broadcasts to all peers in room**
   ```
   (All peers receive) → message
   ```

3. **Message appears in client UI**

## Error Codes

| Error | Meaning | Resolution |
|-------|---------|-----------|
| `room not found` | Room ID doesn't exist | Join existing room or create new one |
| `room is full` | Maximum peers reached | Try different room |
| `missing room_id` | Join request incomplete | Provide valid room_id |
| `missing user_id` | Join request incomplete | Provide valid user_id |
| `peer not found` | Invalid peer ID | Check peer ID is correct |
| `invalid candidate` | Malformed ICE candidate | Verify candidate format |

## Rate Limiting

No built-in rate limiting. Implement at application level:
- Max messages per second per peer
- Max join attempts per minute
- Max peer connections per IP

## Timeouts

- **WebSocket idle timeout:** 60 seconds (ping/pong)
- **Message queue retention:** All room messages
- **Room timeout:** Configurable (default 1 hour)

## Performance Metrics

Monitor via `/api/rooms`:
- Active room count
- Peers per room
- Connection duration

Additional metrics via client-side WebRTC API:
- Round-trip time (latency)
- Available bandwidth
- Packets lost
- Jitter

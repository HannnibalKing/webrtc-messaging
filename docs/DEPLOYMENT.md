# WebRTC Deployment Guide

## Production Deployment Checklist

- [ ] Use HTTPS/WSS (not HTTP/WS)
- [ ] Enable CORS security headers
- [ ] Implement authentication/authorization
- [ ] Configure TURN servers for NAT traversal
- [ ] Set up monitoring and alerting
- [ ] Configure database for persistence (optional)
- [ ] Load balance across multiple servers
- [ ] Set resource limits and rate limiting
- [ ] Enable logging and tracing
- [ ] Run security scan and penetration tests

## Docker Deployment

### Single Container

```bash
# Build image
docker build -t webrtc-messaging:latest .

# Run container
docker run -d \
  -p 8080:8080 \
  -e SERVER_PORT=8080 \
  -e STUN_SERVERS="stun:stun.l.google.com:19302" \
  --name webrtc-server \
  webrtc-messaging:latest

# View logs
docker logs -f webrtc-server

# Stop container
docker stop webrtc-server
docker rm webrtc-server
```

### Docker Compose Stack

```bash
# Start services
docker-compose up -d

# View logs
docker-compose logs -f server

# Stop services
docker-compose down

# Scale signaling servers
docker-compose up -d --scale server=3
```

## Kubernetes Deployment

### Deployment Manifest

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: webrtc-signaling
spec:
  replicas: 3
  selector:
    matchLabels:
      app: webrtc-signaling
  template:
    metadata:
      labels:
        app: webrtc-signaling
    spec:
      containers:
      - name: signaling-server
        image: webrtc-messaging:latest
        ports:
        - containerPort: 8080
        env:
        - name: SERVER_PORT
          value: "8080"
        - name: STUN_SERVERS
          value: "stun:stun.l.google.com:19302"
        resources:
          requests:
            memory: "256Mi"
            cpu: "250m"
          limits:
            memory: "512Mi"
            cpu: "500m"
        livenessProbe:
          httpGet:
            path: /api/health
            port: 8080
          initialDelaySeconds: 10
          periodSeconds: 10
        readinessProbe:
          httpGet:
            path: /api/health
            port: 8080
          initialDelaySeconds: 5
          periodSeconds: 5
---
apiVersion: v1
kind: Service
metadata:
  name: webrtc-signaling
spec:
  selector:
    app: webrtc-signaling
  ports:
  - protocol: TCP
    port: 80
    targetPort: 8080
  type: LoadBalancer
---
apiVersion: autoscaling/v2
kind: HorizontalPodAutoscaler
metadata:
  name: webrtc-hpa
spec:
  scaleTargetRef:
    apiVersion: apps/v1
    kind: Deployment
    name: webrtc-signaling
  minReplicas: 3
  maxReplicas: 10
  metrics:
  - type: Resource
    resource:
      name: cpu
      target:
        type: Utilization
        averageUtilization: 70
```

### Deploy to Kubernetes

```bash
# Create namespace
kubectl create namespace webrtc

# Apply manifests
kubectl apply -f k8s/deployment.yaml -n webrtc

# Check status
kubectl get pods -n webrtc
kubectl get svc -n webrtc

# View logs
kubectl logs -f deployment/webrtc-signaling -n webrtc

# Scale manually
kubectl scale deployment webrtc-signaling --replicas=5 -n webrtc
```

## Cloud Platforms

### AWS Elastic Container Service (ECS)

```bash
# Push to ECR
aws ecr get-login-password --region us-east-1 | \
  docker login --username AWS --password-stdin ACCOUNT_ID.dkr.ecr.us-east-1.amazonaws.com

docker tag webrtc-messaging:latest \
  ACCOUNT_ID.dkr.ecr.us-east-1.amazonaws.com/webrtc-messaging:latest

docker push ACCOUNT_ID.dkr.ecr.us-east-1.amazonaws.com/webrtc-messaging:latest

# Create ECS service with ALB load balancer
```

### Google Cloud Run

```bash
# Build and push
gcloud builds submit --tag gcr.io/PROJECT_ID/webrtc-messaging

# Deploy
gcloud run deploy webrtc-messaging \
  --image gcr.io/PROJECT_ID/webrtc-messaging \
  --port 8080 \
  --memory 512Mi \
  --cpu 1 \
  --min-instances 3 \
  --max-instances 100
```

### Azure Container Instances

```bash
# Build image
az acr build --registry myregistry --image webrtc-messaging:latest .

# Deploy container
az container create \
  --resource-group mygroup \
  --name webrtc-server \
  --image myregistry.azurecr.io/webrtc-messaging:latest \
  --ports 8080 \
  --cpu 2 \
  --memory 1
```

## HTTPS/WSS Configuration

### Self-Signed Certificate (Development)

```bash
# Generate certificate
openssl req -x509 -newkey rsa:4096 \
  -keyout key.pem -out cert.pem -days 365 -nodes

# Configure server
export TLS_CERT=cert.pem
export TLS_KEY=key.pem
```

### Let's Encrypt (Production)

```bash
# Using Certbot
sudo certbot certonly --standalone -d yourdomain.com

# Certificate path: /etc/letsencrypt/live/yourdomain.com/
```

### Nginx Reverse Proxy

```nginx
upstream webrtc_backend {
    server 127.0.0.1:8080;
}

server {
    listen 443 ssl http2;
    server_name yourdomain.com;

    ssl_certificate /etc/letsencrypt/live/yourdomain.com/fullchain.pem;
    ssl_certificate_key /etc/letsencrypt/live/yourdomain.com/privkey.pem;

    location / {
        proxy_pass http://webrtc_backend;
        proxy_http_version 1.1;
        proxy_set_header Upgrade $http_upgrade;
        proxy_set_header Connection "upgrade";
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
    }

    location /ws {
        proxy_pass http://webrtc_backend/ws;
        proxy_http_version 1.1;
        proxy_set_header Upgrade $http_upgrade;
        proxy_set_header Connection "upgrade";
        proxy_read_timeout 86400;
    }
}
```

## TURN Server Setup

### Coturn Installation

```bash
# Ubuntu/Debian
sudo apt-get install coturn

# Configure /etc/coturn/turnserver.conf
listening-port=3478
listening-ip=0.0.0.0
external-ip=YOUR_PUBLIC_IP
realm=yourdomain.com
server-name=yourdomain.com
user=username:password
```

### Access in Application

```go
config.TURNServers = []string{
    "turn:your-turn-server.com:3478?transport=udp",
    "turn:your-turn-server.com:3478?transport=tcp",
    "turns:your-turn-server.com:5349?transport=tcp",
}
```

## Monitoring and Logging

### Application Metrics

```go
import (
    "github.com/prometheus/client_golang/prometheus"
    "github.com/prometheus/client_golang/prometheus/promhttp"
)

// Register metrics
prometheus.MustRegister(activeConnections)
prometheus.MustRegister(messageCount)

// Expose metrics
http.Handle("/metrics", promhttp.Handler())
```

### Log Aggregation (ELK Stack)

```yaml
# filebeat.yml
filebeat.inputs:
- type: log
  enabled: true
  paths:
    - /var/log/webrtc/*.log

output.elasticsearch:
  hosts: ["elasticsearch:9200"]

logging.level: info
```

### Monitoring Dashboard (Grafana)

Query examples:
```
# Active connections
rate(connection_total[1m])

# Message throughput
rate(messages_total[5m])

# Server latency
histogram_quantile(0.95, latency_seconds_bucket)
```

## Scaling Strategies

### Horizontal Scaling with Load Balancer

```
                    ┌─────────────────┐
                    │  Load Balancer  │
                    │  (Sticky Sess)  │
                    └────────┬────────┘
                    ┌────────┼────────┐
                    │        │        │
                ┌───▼──┐ ┌──▼───┐ ┌──▼───┐
                │ SRV1 │ │ SRV2 │ │ SRV3 │
                └───┬──┘ └──┬───┘ └──┬───┘
                    │       │       │
                    └───────┼───────┘
                            │
                        ┌───▼────┐
                        │ Redis  │ (Session/Message Store)
                        └────────┘
```

### Configuration for Load Balancer

```nginx
upstream webrtc {
    # Sticky sessions for WebSocket
    hash $remote_addr consistent;
    
    server server1.example.com:8080 weight=1;
    server server2.example.com:8080 weight=1;
    server server3.example.com:8080 weight=1;
}
```

## Performance Tuning

### System Limits

```bash
# /etc/security/limits.conf
* soft nofile 65536
* hard nofile 65536
* soft nproc 32768
* hard nproc 32768

# Apply
ulimit -n 65536
```

### Network Tuning

```bash
# Increase buffer sizes
sysctl -w net.core.rmem_max=134217728
sysctl -w net.core.wmem_max=134217728
sysctl -w net.ipv4.tcp_rmem="4096 87380 67108864"
sysctl -w net.ipv4.tcp_wmem="4096 65536 67108864"
```

### Go Runtime Tuning

```go
import "runtime"

func init() {
    // Tune garbage collection
    runtime.GOGC = 50  // More aggressive GC
    
    // Set max procs
    runtime.GOMAXPROCS(runtime.NumCPU())
}
```

## Health Checks & Recovery

### Liveness Probe

```bash
curl http://localhost:8080/api/health
```

### Readiness Probe

```go
func isReady() bool {
    return hub.ListRooms() != nil  // Check if hub is initialized
}
```

### Automatic Restart Policy

```yaml
# Docker
--restart=always

# Kubernetes
restartPolicy: Always
```

## Backup & Recovery

### State Backup

```go
func (h *Hub) ExportState() map[string]interface{} {
    h.mu.RLock()
    defer h.mu.RUnlock()
    
    roomData := make([]map[string]interface{}, 0)
    for _, room := range h.rooms {
        roomData = append(roomData, map[string]interface{}{
            "id": room.ID,
            "name": room.Name,
            "peers": len(room.Peers),
        })
    }
    
    return map[string]interface{}{
        "timestamp": time.Now(),
        "rooms": roomData,
    }
}
```

## Security Hardening

### Rate Limiting

```go
import "github.com/go-redis/redis_rate"

limiter := redis_rate.NewLimiter(redisClient)

// Limit to 100 messages per minute
limit := redis_rate.Limit{Rate: 100, Burst: 10, Period: time.Minute}
if err := limiter.Allow(ctx, "user-123", limit); err != nil {
    // Rate limit exceeded
}
```

### Authentication

```go
func validateToken(token string) (string, error) {
    claims := jwt.MapClaims{}
    jwt.ParseWithClaims(token, claims, func(token *jwt.Token) (interface{}, error) {
        return []byte("secret"), nil
    })
    
    if userID, ok := claims["user_id"].(string); ok {
        return userID, nil
    }
    return "", errors.New("invalid token")
}
```

## Incident Response

### Common Issues

| Issue | Cause | Solution |
|-------|-------|----------|
| High memory | Goroutine leak | Add goroutine profiling, check channel closures |
| Connection timeouts | Firewall/NAT | Configure TURN servers, check firewall rules |
| Audio/video lag | Bandwidth | Reduce resolution, increase STUN timeouts |
| WebSocket closes | Server restart | Implement reconnection logic in client |

### Debug Mode

```bash
# Enable debug logging
RUST_LOG=debug cargo run

# Enable CPU profiling
go run -cpuprofile=cpu.prof cmd/server/main.go

# Analyze
go tool pprof cpu.prof
```

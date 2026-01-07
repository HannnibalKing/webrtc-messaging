package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/joho/godotenv"
)

// Config holds all WebRTC server configuration
type Config struct {
	// Server
	ServerPort int
	ServerHost string

	// WebRTC
	STUNServers []string
	TURNServers []string

	// WebSocket
	WSPath           string
	WSReadBuffer     int
	WSWriteBuffer    int
	WSMaxConnections int

	// Media
	AudioEnabled bool
	VideoEnabled bool
	VideoCodec   string
	AudioCodec   string

	// Rooms
	MaxPeersPerRoom    int
	DefaultRoomTimeout int
	AllowAnonymous     bool

	// Logging
	LogLevel  string
	LogFormat string
}

// Load loads configuration from environment variables
func Load() (*Config, error) {
	_ = godotenv.Load()

	cfg := &Config{
		ServerPort:         getEnvInt("SERVER_PORT", 8080),
		ServerHost:         getEnv("SERVER_HOST", "0.0.0.0"),
		STUNServers:        parseServers(getEnv("STUN_SERVERS", "stun:stun.l.google.com:19302")),
		TURNServers:        parseServers(getEnv("TURN_SERVERS", "")),
		WSPath:             getEnv("WS_PATH", "/ws"),
		WSReadBuffer:       getEnvInt("WS_READ_BUFFER", 1024),
		WSWriteBuffer:      getEnvInt("WS_WRITE_BUFFER", 1024),
		WSMaxConnections:   getEnvInt("WS_MAX_CONNECTIONS", 1000),
		AudioEnabled:       getEnvBool("AUDIO_ENABLED", true),
		VideoEnabled:       getEnvBool("VIDEO_ENABLED", true),
		VideoCodec:         getEnv("VIDEO_CODEC", "VP8"),
		AudioCodec:         getEnv("AUDIO_CODEC", "OPUS"),
		MaxPeersPerRoom:    getEnvInt("MAX_PEERS_PER_ROOM", 100),
		DefaultRoomTimeout: getEnvInt("DEFAULT_ROOM_TIMEOUT", 3600),
		AllowAnonymous:     getEnvBool("ALLOW_ANONYMOUS", true),
		LogLevel:           getEnv("LOG_LEVEL", "info"),
		LogFormat:          getEnv("LOG_FORMAT", "json"),
	}

	if err := cfg.Validate(); err != nil {
		return nil, err
	}

	return cfg, nil
}

// Validate validates the configuration
func (c *Config) Validate() error {
	if c.ServerPort < 1 || c.ServerPort > 65535 {
		return fmt.Errorf("invalid SERVER_PORT: %d (must be 1-65535)", c.ServerPort)
	}

	if c.MaxPeersPerRoom < 1 {
		return fmt.Errorf("MAX_PEERS_PER_ROOM must be at least 1")
	}

	if c.DefaultRoomTimeout < 60 {
		return fmt.Errorf("DEFAULT_ROOM_TIMEOUT must be at least 60 seconds")
	}

	return nil
}

func getEnv(key, defaultValue string) string {
	value, exists := os.LookupEnv(key)
	if !exists {
		return defaultValue
	}
	return value
}

func getEnvInt(key string, defaultValue int) int {
	value, exists := os.LookupEnv(key)
	if !exists {
		return defaultValue
	}

	intValue, err := strconv.Atoi(value)
	if err != nil {
		return defaultValue
	}

	return intValue
}

func getEnvBool(key string, defaultValue bool) bool {
	value, exists := os.LookupEnv(key)
	if !exists {
		return defaultValue
	}

	boolValue, err := strconv.ParseBool(value)
	if err != nil {
		return defaultValue
	}

	return boolValue
}

func parseServers(serverString string) []string {
	if serverString == "" {
		return []string{}
	}

	servers := strings.Split(serverString, ",")
	for i := range servers {
		servers[i] = strings.TrimSpace(servers[i])
	}
	return servers
}

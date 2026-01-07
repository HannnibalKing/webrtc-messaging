package message

import (
	"encoding/json"
	"fmt"
	"sync"
	"time"
)

// MessageType defines the type of message
type MessageType string

const (
	TextMessage   MessageType = "text"
	SystemMessage MessageType = "system"
	DataMessage   MessageType = "data"
)

// Message represents a message between peers
type Message struct {
	ID        string                 `json:"id"`
	Type      MessageType            `json:"type"`
	SenderID  string                 `json:"sender_id"`
	RoomID    string                 `json:"room_id"`
	Content   string                 `json:"content"`
	Timestamp time.Time              `json:"timestamp"`
	Metadata  map[string]interface{} `json:"metadata,omitempty"`
}

// MessageQueue manages message delivery
type MessageQueue struct {
	messages map[string][]*Message
	mu       sync.RWMutex
	maxSize  int
}

// MessageRouter routes messages between peers
type MessageRouter struct {
	queue  *MessageQueue
	mu     sync.RWMutex
	routes map[string][]MessageHandler
}

// MessageHandler is a function that handles messages
type MessageHandler func(*Message) error

// NewMessageQueue creates a new message queue
func NewMessageQueue(maxSize int) *MessageQueue {
	return &MessageQueue{
		messages: make(map[string][]*Message),
		maxSize:  maxSize,
	}
}

// EnqueueMessage adds a message to the queue
func (mq *MessageQueue) EnqueueMessage(roomID string, msg *Message) error {
	mq.mu.Lock()
	defer mq.mu.Unlock()

	if _, exists := mq.messages[roomID]; !exists {
		mq.messages[roomID] = make([]*Message, 0, mq.maxSize)
	}

	if len(mq.messages[roomID]) >= mq.maxSize {
		// Remove oldest message
		mq.messages[roomID] = mq.messages[roomID][1:]
	}

	mq.messages[roomID] = append(mq.messages[roomID], msg)
	return nil
}

// DequeueMessages retrieves messages for a room
func (mq *MessageQueue) DequeueMessages(roomID string, limit int) []*Message {
	mq.mu.RLock()
	defer mq.mu.RUnlock()

	messages, exists := mq.messages[roomID]
	if !exists {
		return []*Message{}
	}

	if limit > len(messages) {
		limit = len(messages)
	}

	result := make([]*Message, limit)
	copy(result, messages[len(messages)-limit:])
	return result
}

// ClearRoom removes all messages for a room
func (mq *MessageQueue) ClearRoom(roomID string) {
	mq.mu.Lock()
	defer mq.mu.Unlock()
	delete(mq.messages, roomID)
}

// NewMessageRouter creates a new message router
func NewMessageRouter(queue *MessageQueue) *MessageRouter {
	return &MessageRouter{
		queue:  queue,
		routes: make(map[string][]MessageHandler),
	}
}

// RegisterHandler registers a handler for a message type
func (mr *MessageRouter) RegisterHandler(msgType string, handler MessageHandler) {
	mr.mu.Lock()
	defer mr.mu.Unlock()

	if _, exists := mr.routes[msgType]; !exists {
		mr.routes[msgType] = make([]MessageHandler, 0)
	}

	mr.routes[msgType] = append(mr.routes[msgType], handler)
}

// Route routes a message to appropriate handlers
func (mr *MessageRouter) Route(msg *Message) error {
	mr.mu.RLock()
	handlers, exists := mr.routes[string(msg.Type)]
	mr.mu.RUnlock()

	if !exists {
		return fmt.Errorf("no handlers registered for message type: %s", msg.Type)
	}

	for _, handler := range handlers {
		if err := handler(msg); err != nil {
			return err
		}
	}

	return nil
}

// ParseMessage parses a JSON message
func ParseMessage(data []byte) (*Message, error) {
	var msg Message
	if err := json.Unmarshal(data, &msg); err != nil {
		return nil, err
	}
	return &msg, nil
}

// SerializeMessage serializes a message to JSON
func SerializeMessage(msg *Message) ([]byte, error) {
	return json.Marshal(msg)
}

// NewMessage creates a new message
func NewMessage(msgType MessageType, senderID, roomID, content string) *Message {
	return &Message{
		Type:      msgType,
		SenderID:  senderID,
		RoomID:    roomID,
		Content:   content,
		Timestamp: time.Now(),
		Metadata:  make(map[string]interface{}),
	}
}

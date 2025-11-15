package app

import "encoding/json"

type Message struct {
	Type     string          `json:"type"`                // chat, join, leave, private, etc.
	SenderID string          `json:"sender_id"`           // UserID of the sender
	TargetID string          `json:"target_id,omitempty"` // For private messages
	Payload  json.RawMessage `json:"payload"`             // The actual data (e.g., chat content)
}

type SimpleChatPayload struct {
	Content string `json:"content"`
}

package realtime

import (
	"RealTime/internal/logger"

	"go.uber.org/zap"
)

type MessageHandler interface {
	Handle(hub *Hub, message *Message)
}

type ChatHandler struct {
}

type NewsHandler struct {
}

func (ChatHandler) Handle(hub *Hub, message *Message) {
	hub.BroadcastToAll(message)
}

func (NewsHandler) Handle(hub *Hub, message *Message) {
	hub.BroadcastToAll(message)
}

type PingHandler struct {
}

func (PingHandler) Handle(hub *Hub, message *Message) {
	logger.Logger.Info("App-Ping received.", zap.String("Sender ID", message.SenderID))
}

type PrivateHandler struct {
}

func (PrivateHandler) Handle(hub *Hub, message *Message) {
	if message.TargetID != "" {
		hub.SendToClient(message.TargetID, message)
	} else {
		logger.Logger.Info("Received 'private' message without TargetID from %s", zap.String("Sender ID", message.SenderID))
	}
}

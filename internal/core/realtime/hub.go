package realtime

import (
	"RealTime/internal/logger"
	"encoding/json"
	"fmt"
	"time"

	"go.uber.org/zap"
)

type Hub struct {
	clients    map[string]*Client
	broadcast  chan *Message
	register   chan *Client
	unregister chan *Client
	dispatcher *Dispatcher
}

func NewHub() *Hub {
	return &Hub{
		clients:    make(map[string]*Client),
		broadcast:  make(chan *Message, 256),
		register:   make(chan *Client),
		unregister: make(chan *Client),
		dispatcher: NewDispatcher(),
	}
}

func (h *Hub) Register(client *Client) {
	h.register <- client
}

func (h *Hub) Unregister(client *Client) {
	h.unregister <- client
}

func (h *Hub) Run() {
	logger.Logger.Info("Hub started")
	for {
		select {
		case client := <-h.register:
			handleRegisterEvent(client, h)
		case client := <-h.unregister:
			handleUnregisterEvent(client, h)
		case message := <-h.broadcast:
			h.dispatcher.Dispatch(h, message)
		}
	}
}

func (h *Hub) BroadcastToAll(msg *Message) {

	jsonMessage, err := json.Marshal(msg)
	if err != nil {
		logger.Logger.Error("Error marshaling message for broadcast: %v", zap.Error(err))
		return
	}

	for id, client := range h.clients {
		select {
		case client.send <- jsonMessage:
		default:
			logger.Logger.Info("Client %s send channel blocked (full). Unregistering...", zap.String("client_id", id))
			close(client.send)
			delete(h.clients, id)
		}
	}
}

func (h *Hub) SendToClient(targetID string, msg *Message) {
	client, ok := h.clients[targetID]
	if !ok {
		logger.Logger.Info("Target client %s not found for private message.", zap.String("Target ID", targetID))
		return
	}

	jsonMessage, err := json.Marshal(msg)
	if err != nil {
		logger.Logger.Info("Error marshaling message for private send to %s: %v", zap.String("Target ID", targetID), zap.Error(err))
		return
	}

	select {
	case client.send <- jsonMessage:
	default:
		logger.Logger.Info("Target client %s send channel blocked. Unregistering...", zap.String("Target ID", targetID))
		close(client.send)
		delete(h.clients, targetID)
	}
}

func (h *Hub) Broadcast(msg *Message) {
	select {
	case h.broadcast <- msg:
	default:
		logger.Logger.Warn("Hub broadcast channel is saturated. Message dropped.", zap.String("message_type", msg.Type))
	}
}

func handleRegisterEvent(client *Client, hub *Hub) {
	hub.clients[client.ID] = client
	logger.Logger.Info("Client registered", zap.String("client_id", client.ID), zap.Int("total_clients", len(hub.clients)))

	welcomeMsg := fmt.Sprintf(
		`{"type": "welcome", "user_id": "%s", "user_name": "%s", "message": "Welcome!"}`,
		client.ID,
		client.UserName,
	)
	joinMsg := &Message{
		Type:     "join",
		SenderID: client.ID,
		Payload:  []byte(welcomeMsg),
	}
	hub.broadcast <- joinMsg
	select {
	case client.send <- []byte(welcomeMsg):
	default:
		logger.Logger.Info("Client %s send channel blocked on register. Unregistering.", zap.String("client_id", client.ID))

		close(client.send)
		delete(hub.clients, client.ID)
	}
}

func handleUnregisterEvent(client *Client, hub *Hub) {
	if _, ok := hub.clients[client.ID]; ok {
		close(client.send)
		delete(hub.clients, client.ID)
		leaveMsg := &Message{
			Type:     "leave",
			SenderID: client.ID,
		}
		hub.broadcast <- leaveMsg
		logger.Logger.Info("Client unregistered: %s. Total clients: %d", zap.String("client_id", client.ID), zap.Int("total_clients", len(hub.clients)))
	}
}

func (h *Hub) SendNewsUpdates() {
	ticker := time.NewTicker(10 * time.Second)

	go func() {
		for {
			select {
			case <-ticker.C:
				h.BroadcastToAll(&Message{Type: "market_news", Payload: []byte("{\n  \"article\": {\n    \"topic\": \"Crypto\",\n    \"headline\": \"Bitcoin surges past $65k resistance level.\"\n  }\n}")})
			}
		}
	}()

}

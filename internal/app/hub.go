package app

import (
	"RealTime/internal/log"
	"encoding/json"
	"fmt"

	"go.uber.org/zap"
)

type Hub struct {
	clients    map[string]*Client
	broadcast  chan *Message
	register   chan *Client
	unregister chan *Client
}

func NewHub() *Hub {
	return &Hub{
		clients:    make(map[string]*Client),
		broadcast:  make(chan *Message, 256),
		register:   make(chan *Client),
		unregister: make(chan *Client),
	}
}

func (h *Hub) Register(client *Client) {
	h.register <- client
}

func (h *Hub) Unregister(client *Client) {
	h.unregister <- client
}

func (h *Hub) Run() {
	log.Logger.Info("Hub started")
	for {
		select {
		case client := <-h.register:
			h.clients[client.ID] = client
			log.Logger.Info("Client registered", zap.String("client_id", client.ID), zap.Int("total_clients", len(h.clients)))

			welcomeMsg := fmt.Sprintf(
				`{"type": "welcome", "user_id": "%s", "message": "Welcome!"}`,
				client.ID,
			)
			joinMsg := &Message{
				Type:     "join",
				SenderID: client.ID,
			}
			h.broadcast <- joinMsg
			select {
			case client.send <- []byte(welcomeMsg):
				// Sent successfully
			default:
				log.Logger.Info("Client %s send channel blocked on register. Unregistering.", zap.String("client_id", client.ID))

				close(client.send)
				delete(h.clients, client.ID)
			}
		case client := <-h.unregister:
			if _, ok := h.clients[client.ID]; ok {
				close(client.send)
				delete(h.clients, client.ID)
				leaveMsg := &Message{
					Type:     "leave",
					SenderID: client.ID,
				}
				h.broadcast <- leaveMsg
				log.Logger.Info("Client unregistered: %s. Total clients: %d", zap.String("client_id", client.ID), zap.Int("total_clients", len(h.clients)))
			}
		case message := <-h.broadcast:
			switch message.Type {
			case "chat":
				h.broadcastToAll(message)
			case "ping":
				log.Logger.Info("App-Ping from %s received.", zap.String("Sender ID", message.SenderID))
			case "private":
				if message.TargetID != "" {
					h.sendToClient(message.TargetID, message)
				} else {
					log.Logger.Info("Received 'private' message without TargetID from %s", zap.String("Sender ID", message.SenderID))
				}
			case "join", "leave":
				h.broadcastToAll(message)
			}

		}
	}
}

func (h *Hub) broadcastToAll(msg *Message) {

	jsonMessage, err := json.Marshal(msg)
	if err != nil {
		log.Logger.Error("Error marshaling message for broadcast: %v", zap.Error(err))
		return
	}

	for id, client := range h.clients {
		select {
		case client.send <- jsonMessage:
		default:
			log.Logger.Info("Client %s send channel blocked (full). Unregistering...", zap.String("client_id", id))
			close(client.send)
			delete(h.clients, id)
		}
	}
}

func (h *Hub) sendToClient(targetID string, msg *Message) {
	client, ok := h.clients[targetID]
	if !ok {
		log.Logger.Info("Target client %s not found for private message.", zap.String("Target ID", targetID))
		return
	}

	jsonMessage, err := json.Marshal(msg)
	if err != nil {
		log.Logger.Info("Error marshaling message for private send to %s: %v", zap.String("Target ID", targetID), zap.Error(err))
		return
	}

	select {
	case client.send <- jsonMessage:
	default:
		log.Logger.Info("Target client %s send channel blocked. Unregistering...", zap.String("Target ID", targetID))
		close(client.send)
		delete(h.clients, targetID)
	}
}

func (h *Hub) Broadcast(msg *Message) {
	select {
	case h.broadcast <- msg:
	default:
		log.Logger.Warn("Hub broadcast channel is saturated. Message dropped.", zap.String("message_type", msg.Type))
	}
}

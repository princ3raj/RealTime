package ws

import (
	"RealTime/internal/auth"
	realtime2 "RealTime/internal/core/realtime"
	"RealTime/internal/log"
	"net/http"

	"github.com/gorilla/websocket"
	"go.uber.org/zap"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin:     func(r *http.Request) bool { return true },
}

func NewWsHandlerFactory(hub *realtime2.Hub, jwtSecret string) http.HandlerFunc {

	return func(w http.ResponseWriter, r *http.Request) {

		tokenString := r.URL.Query().Get("token")
		if tokenString == "" {
			http.Error(w, "Authentication token required.", http.StatusUnauthorized)
			return
		}

		userID, userName, err := auth.ValidateWsToken(tokenString, jwtSecret)
		if err != nil {
			log.Logger.Warn("Authentication failed for token",
				zap.Error(err),
				zap.String("token_prefix", tokenString[:min(len(tokenString), 10)]),
			)
			http.Error(w, "Invalid or expired token.", http.StatusUnauthorized)
			return
		}

		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			log.Logger.Error("Upgrade failed", zap.Error(err))
			return
		}

		client := realtime2.NewClient(hub, conn, userID, userName)

		hub.Register(client)
		hub.SendNewsUpdates()

		go client.WritePump()
		client.ReadPump()
	}
}

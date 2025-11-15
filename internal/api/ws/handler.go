package ws

import (
	"RealTime/internal/app"
	"RealTime/internal/auth"
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

func NewWsHandlerFactory(hub *app.Hub, jwtSecret string) http.HandlerFunc {

	return func(w http.ResponseWriter, r *http.Request) {

		tokenString := r.URL.Query().Get("token")
		if tokenString == "" {
			http.Error(w, "Authentication token required.", http.StatusUnauthorized)
			return
		}

		userID, err := auth.ValidateWsToken(tokenString, jwtSecret)
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

		client := app.NewClient(hub, conn, userID)

		hub.Register(client)

		go client.WritePump()
		client.ReadPump()
	}
}

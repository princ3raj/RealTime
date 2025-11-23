package wiring

import (
	"RealTime/internal/config"
	"RealTime/internal/core/realtime"
	"RealTime/internal/core/service"
	"RealTime/internal/repository/postgres"
	transport "RealTime/internal/transport/http"
	"RealTime/internal/transport/ws"
	"database/sql"
	"log"
	"net/http"
)

type WsApp struct {
	Handler   http.Handler
	ChatHub   *realtime.Hub
	NotifyHub *realtime.Hub
}

type NoOpPublisher struct{}

func NewNoOpPublisher() *NoOpPublisher {
	return &NoOpPublisher{}
}

// Publish implements service.Publisher implicitly
func (n *NoOpPublisher) Publish(topic string, message []byte) error {
	// Just log it so we see it working in the terminal
	log.Printf("[Mock Event] Topic: %s | Payload: %s", topic, string(message))
	return nil
}

func BuildRestApi(db *sql.DB, cfg *config.Config) (http.Handler, error) {
	userStore := postgres.NewUserStore(db)
	publisher := NewNoOpPublisher()
	userService := service.NewUserService(userStore, publisher)

	deps := &transport.AppDependencies{
		UserService: userService,
		Config:      cfg,
	}

	router := transport.NewRootRouter(deps)

	return router, nil
}

func BuildWsServer(cfg *config.Config) (*WsApp, error) {
	chatHub := realtime.NewHub()
	notifyHub := realtime.NewHub()
	newsFeedHub := realtime.NewHub()

	chatHandler := ws.NewWsHandlerFactory(chatHub, cfg.JWTSecret)
	notifyHandler := ws.NewWsHandlerFactory(notifyHub, cfg.JWTSecret)
	newsFeedHandler := ws.NewWsHandlerFactory(newsFeedHub, cfg.JWTSecret)

	mux := http.NewServeMux()
	mux.HandleFunc("/ws/chat", chatHandler)
	mux.HandleFunc("/ws/group/chat", chatHandler)
	mux.HandleFunc("/ws/notifications", notifyHandler)
	mux.HandleFunc("/ws/news", newsFeedHandler)

	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, err := w.Write([]byte("WS Server OK"))
		if err != nil {
			return
		}
	})

	wsApp := &WsApp{
		Handler:   mux,
		ChatHub:   chatHub,
		NotifyHub: notifyHub,
	}

	return wsApp, nil
}

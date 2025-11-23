package wiring

import (
	"RealTime/internal/config"
	"RealTime/internal/core/realtime"
	"RealTime/internal/core/service"
	"RealTime/internal/repository/postgres/store"
	"RealTime/internal/transport/contracts"
	transport "RealTime/internal/transport/http"
	"RealTime/internal/transport/ws"
	"database/sql"
	"net/http"

	"RealTime/internal/log" // <-- 2. IMPORT log for publisher stub

	"go.uber.org/zap"
)

// WsApp remains the same. It's the contract for the WS server binary.
type WsApp struct {
	Handler   http.Handler
	ChatHub   *realtime.Hub
	NotifyHub *realtime.Hub
}

// --- Stub Publisher (for decoupled REST API) ---

// NoOpPublisher is a temporary stub that satisfies the http.Publisher interface.
// This allows the REST API to compile and run without a real message broker.
// We will replace this with a real Redis/Kafka publisher when ready.
type NoOpPublisher struct{}

func (n *NoOpPublisher) Publish(topic string, message []byte) error {
	// A real implementation would send this to Redis/Kafka.
	// For now, we just log the intent.
	log.Logger.Info("StubPublisher: message published (but not sent)",
		zap.String("topic", topic),
		zap.Int("size_bytes", len(message)),
	)
	return nil
}

// NewNoOpPublisher provides the stub implementation.
func NewNoOpPublisher() contracts.Publisher {
	return &NoOpPublisher{}
}

// BuildRestApi assembles the entire REST API server.
func BuildRestApi(db *sql.DB, cfg *config.Config) (http.Handler, error) {

	// 1. Build Data Layer (Concrete Adapter)
	// This implements the user.Storer interface.
	userStore := store.NewUserStore(db)

	// 2. Build Domain Layer (Concrete Service)
	// This implements the http.UserServiceProvider interface.
	userService := service.NewUserService(userStore)

	// 3. Build Publisher (Stubbed Adapter)
	// This implements the http.Publisher interface.
	publisher := NewNoOpPublisher()

	// 4. Assemble Dependencies
	deps := &contracts.AppDependencies{
		UserService: userService, // Concrete *user.Service satisfies http.UserServiceProvider
		Publisher:   publisher,   // Concrete *NoOpPublisher satisfies http.Publisher
		Config:      cfg,
	}

	// 5. Build API Layer (Root Router)
	// We call the NewRootRouter, which assembles all resource sub-routers.
	router := transport.NewRootRouter(deps)

	return router, nil
}

// BuildWsServer assembles the WebSocket server.
func BuildWsServer(cfg *config.Config) (*WsApp, error) {

	chatHub := realtime.NewHub()
	notifyHub := realtime.NewHub()
	newsFeedHub := realtime.NewHub()

	// Use the handler factory from the ws package
	chatHandler := ws.NewWsHandlerFactory(chatHub, cfg.JWTSecret)
	notifyHandler := ws.NewWsHandlerFactory(notifyHub, cfg.JWTSecret)
	newsFeedHandler := ws.NewWsHandlerFactory(newsFeedHub, cfg.JWTSecret)

	mux := http.NewServeMux()
	mux.HandleFunc("/ws/chat", chatHandler)
	mux.HandleFunc("/ws/group/chat", chatHandler)
	mux.HandleFunc("/ws/notifications", notifyHandler)
	mux.HandleFunc("/ws/news", newsFeedHandler)

	// Stub the health handler since it's not defined
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

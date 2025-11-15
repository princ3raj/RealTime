package wiring

import (
	"RealTime/internal/api/contracts"
	"RealTime/internal/api/rest" // <-- 1. IMPORT the API layer's contracts
	"RealTime/internal/api/ws"
	"RealTime/internal/app"
	"RealTime/internal/config"
	"RealTime/internal/domain/user"
	"RealTime/internal/store"
	"database/sql"
	"net/http"

	"RealTime/internal/log" // <-- 2. IMPORT log for publisher stub

	"go.uber.org/zap"
)

// WsApp remains the same. It's the contract for the WS server binary.
type WsApp struct {
	Handler   http.Handler
	ChatHub   *app.Hub
	NotifyHub *app.Hub
}

// --- Stub Publisher (for decoupled REST API) ---

// NoOpPublisher is a temporary stub that satisfies the rest.Publisher interface.
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
	// This implements the rest.UserServiceProvider interface.
	userService := user.NewService(userStore)

	// 3. Build Publisher (Stubbed Adapter)
	// This implements the rest.Publisher interface.
	publisher := NewNoOpPublisher()

	// 4. Assemble Dependencies
	deps := &contracts.AppDependencies{
		UserService: userService, // Concrete *user.Service satisfies rest.UserServiceProvider
		Publisher:   publisher,   // Concrete *NoOpPublisher satisfies rest.Publisher
		Config:      cfg,
	}

	// 5. Build API Layer (Root Router)
	// We call the NewRootRouter, which assembles all resource sub-routers.
	router := rest.NewRootRouter(deps)

	return router, nil
}

// BuildWsServer assembles the WebSocket server.
func BuildWsServer(cfg *config.Config) (*WsApp, error) {

	chatHub := app.NewHub()
	notifyHub := app.NewHub()

	// Use the handler factory from the ws package
	chatHandler := ws.NewWsHandlerFactory(chatHub, cfg.JWTSecret)
	notifyHandler := ws.NewWsHandlerFactory(notifyHub, cfg.JWTSecret)

	mux := http.NewServeMux()
	mux.HandleFunc("/ws/chat", chatHandler)
	mux.HandleFunc("/ws/group/chat", chatHandler)
	mux.HandleFunc("/ws/notifications", notifyHandler)

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

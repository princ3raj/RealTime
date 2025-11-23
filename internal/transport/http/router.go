package http

import (
	"RealTime/internal/config"
	rest "RealTime/internal/transport/http/v1/client"
	"RealTime/internal/transport/http/v1/user"
	"net/http"

	"github.com/gorilla/mux"
)

type AppDependencies struct {
	UserService user.ServiceProvider
	Config      *config.Config
}

func NewRootRouter(deps *AppDependencies) http.Handler {
	rootRouter := mux.NewRouter()

	setUpUserRoutes(rootRouter, deps)

	rootRouter.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, err := w.Write([]byte("REST API OK"))
		if err != nil {
			return
		}
	}).Methods("GET")

	rootRouter.HandleFunc("/client", rest.ClientHandler).Methods("GET")

	return rootRouter
}

func setUpUserRoutes(rootRouter *mux.Router, deps *AppDependencies) {
	handlerCfg := user.HandlerConfig{
		JWTSecret:    deps.Config.JWTSecret,
		TokenTimeout: deps.Config.TokenTimeout,
	}

	userRouter := user.NewUserRouter(deps.UserService, handlerCfg)

	rootRouter.PathPrefix("/api/v1/users").Handler(
		http.StripPrefix("/api/v1/users", userRouter),
	)
}

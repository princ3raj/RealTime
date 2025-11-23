package http

import (
	"RealTime/internal/transport/contracts"
	"RealTime/internal/transport/http/v1/client"
	"RealTime/internal/transport/http/v1/user"
	"net/http"

	"github.com/gorilla/mux"
)

// NewRootRouter is the main assembler for the API.
func NewRootRouter(deps *contracts.AppDependencies) http.Handler {
	rootRouter := mux.NewRouter()

	setUpUserRoutes(rootRouter, deps)

	rootRouter.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("REST API OK"))
	}).Methods("GET")

	rootRouter.HandleFunc("/client", rest.ClientHandler).Methods("GET")

	return rootRouter
}

func setUpUserRoutes(rootRouter *mux.Router, deps *contracts.AppDependencies) {
	userRouter := user.NewUserRouter(deps)
	v1UserRouter := rootRouter.PathPrefix("/api/v1/users").Subrouter()
	v1UserRouter.PathPrefix("/").Handler(
		http.StripPrefix("/api/v1/users", userRouter),
	)

}

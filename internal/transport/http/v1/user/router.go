package user

import (
	"RealTime/internal/transport/contracts"
	"net/http"

	"github.com/gorilla/mux"
)

func NewUserRouter(deps *contracts.AppDependencies) http.Handler {
	api := NewUserAPI(deps)

	router := mux.NewRouter()

	router.HandleFunc("/register", api.RegisterHandler).Methods("POST")
	router.HandleFunc("/login", api.LoginHandler).Methods("POST")

	return router
}

package user

import (
	"net/http"

	"github.com/gorilla/mux"
)

func NewUserRouter(userService ServiceProvider, cfg HandlerConfig) http.Handler {
	api := NewUserAPI(userService, cfg)

	router := mux.NewRouter()

	router.HandleFunc("/register", api.RegisterHandler).Methods("POST")
	router.HandleFunc("/login", api.LoginHandler).Methods("POST")

	return router
}

package rest

import "net/http"

func ClientHandler(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "ws_client.html")
}

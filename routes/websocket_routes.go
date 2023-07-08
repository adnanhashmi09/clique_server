package routes

import (
	"encoding/json"
	"net/http"

	"github.com/adnanhashmi09/clique_server/internal/ws"
	"github.com/go-chi/chi/v5"

	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		// origin := r.Header.Get("Origin")
		// return origin == "http://localhost:3000"
		return true
	},
}

func WSRoutes(r chi.Router, wsHandler *ws.Handler) {
	r.Get("/join_channel/{channel_id}", wsHandler.JoinChannel)
	r.Get("/ping", func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode("PONG")
	})

	r.Post("/create_direct_channel", wsHandler.CreateDirectChannel)
}

package routes

import (
	"github.com/adnanhashmi09/clique_server/internal/ws"
	"github.com/go-chi/chi/v5"
)

func RoomRoutes(r chi.Router, wsHandler *ws.Handler) {
	r.Post("/create_room", wsHandler.CreateRoom)
	r.Post("/join_room", wsHandler.JoinRoom)
	r.Post("/leave_room", wsHandler.LeaveRoom)
}

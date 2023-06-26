package routes

import (
	"encoding/json"
	"fmt"
	"net/http"

	"log"

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

	r.Get("/test/{id}", func(w http.ResponseWriter, r *http.Request) {

		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			log.Println("Error upgrading connection to websockets. ", err)
			http.Error(w, fmt.Sprintln("Cannot establish a websocket connection."), http.StatusInternalServerError)
			return
		}
		handleWsConnection((conn))
	})
}

func handleWsConnection(ws *websocket.Conn) {
	log.Println("new incoming connection from client: ", ws.RemoteAddr())
	readLoop(ws)
}

func readLoop(ws *websocket.Conn) {

	for {
		_, msg, err := ws.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("error in websocket hub readMessage function: %v", err)
			}
			break
		}

		log.Println(string(msg))
		ws.WriteJSON([]byte("Thank you for sending a message :)"))
	}
}

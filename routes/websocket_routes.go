package routes

import (
	"github.com/adnanhashmi09/clique_server/internal/ws"
	"github.com/go-chi/chi/v5"
)

func WSRoutes(r chi.Router, wsHandler *ws.Handler) {
	// r.Post("/create_room", wsHandler.CreateRoom)
}

// func handleWsConnection(ws *websocket.Conn) {
// 	log.Println("new incoming connection from client: ", ws.RemoteAddr())
// 	readLoop(ws)
// }

// func readLoop(ws *websocket.Conn) {
// 	buf := make([]byte, 1024)

// 	for {
// 		n, err := ws.Read(buf)
// 		if err != nil {
// 			if err == io.EOF {
// 				break
// 			}

// 			log.Println("Error: ", err)
// 		}

// 		msg := buf[:n]
// 		log.Println(string(msg))
// 		ws.Write([]byte("Thank you for sending a message :)"))
// 	}
// }

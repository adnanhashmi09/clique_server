package routes

import (
	"io"
	"log"
	// "net/http"

	// "github.com/go-chi/chi/v5"
	"golang.org/x/net/websocket"
)

func handleWsConnection(ws *websocket.Conn) {
	log.Println("new incoming connection from client: ", ws.RemoteAddr())
	readLoop(ws)
}

func readLoop(ws *websocket.Conn) {
	buf := make([]byte, 1024)

	for {
		n, err := ws.Read(buf)
		if err != nil {
			if err == io.EOF {
				break
			}

			log.Println("Error: ", err)
		}

		msg := buf[:n]
		log.Println(string(msg))
		ws.Write([]byte("Thank you for sending a message :)"))
	}
}

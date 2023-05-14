package main

import (
	"log"
	"net/http"

	"github.com/adnanhashmi09/clique_server/db"
	"github.com/adnanhashmi09/clique_server/internal/user"
	"github.com/adnanhashmi09/clique_server/internal/ws"
	"github.com/adnanhashmi09/clique_server/routes"
)

func main() {
	dbConn, err := db.NewDatabaseConn()
	if err != nil {
		log.Fatalf("Could not initialize database connection. %s", err)
	}
	userRepo := user.NewRepository(dbConn.GetDb())
	userServ := user.NewService(userRepo)
	userHandler := user.NewHandler(userServ)

	wsRepo := ws.NewRepository(dbConn.GetDb())
	wsService := ws.NewService(wsRepo)
	hub := ws.NewHub()
	wsHandler := ws.NewHandler(hub, wsService)

	r := routes.RouterInit(userHandler, wsHandler)

	log.Println("Starting server on Port:5050 ")
	log.Fatal(http.ListenAndServe(":5050", r))
}

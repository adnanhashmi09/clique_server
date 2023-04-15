package main

import (
	"log"
	"net/http"

	"github.com/adnanhashmi09/clique_server/db"
	"github.com/adnanhashmi09/clique_server/internal/user"
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

	r := routes.RouterInit(userHandler)

	log.Println("Starting server on Port:5050 ")
	log.Fatal(http.ListenAndServe(":5050", r))
}

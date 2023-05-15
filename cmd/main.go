package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/adnanhashmi09/clique_server/db"
	"github.com/adnanhashmi09/clique_server/internal/user"
	"github.com/adnanhashmi09/clique_server/internal/ws"
	"github.com/adnanhashmi09/clique_server/routes"
	"github.com/adnanhashmi09/clique_server/utils"
)

func main() {

	utils.EnvVariablesInit()
	port := utils.Get_Env_Variable("PORT")

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

	log.Printf("Starting server on Port:%v \n", port)
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%s", port), r))
}

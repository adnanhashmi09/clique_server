package routes

import (
	"net/http"

	"github.com/adnanhashmi09/clique_server/internal/user"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/cors"
	"golang.org/x/net/websocket"
)

func RouterInit(userHandler *user.Handler) http.Handler {
	r := chi.NewRouter()

	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"https://*", "http://*"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: true,
		MaxAge:           300, // Maximum value not ignored by any of major browsers
	}))

	r.Route("/auth", func(r chi.Router) {
		AuthRoutes(r, userHandler)
	})

	r.Mount("/ws", websocket.Handler(handleWsConnection))

	return r
}

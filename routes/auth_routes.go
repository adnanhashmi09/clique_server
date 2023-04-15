package routes

import (
	"github.com/adnanhashmi09/clique_server/internal/user"

	"github.com/go-chi/chi/v5"
)

func AuthRoutes(r chi.Router, userHandler *user.Handler) {
	r.Post("/signup", userHandler.CreateUser)

	r.Post("/login", userHandler.Login)

	r.Post("/logout", userHandler.Logout)

}

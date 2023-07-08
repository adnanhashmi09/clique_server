package routes

import (
	"github.com/adnanhashmi09/clique_server/internal/user"

	"github.com/go-chi/chi/v5"
)

func UserRoutes(r chi.Router, userHandler *user.Handler) {
	r.Get("/all", userHandler.AllInfo)
}

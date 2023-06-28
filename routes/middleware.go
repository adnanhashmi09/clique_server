package routes

import (
	"context"
	"log"
	"net/http"

	"github.com/adnanhashmi09/clique_server/utils"
)

func verify_jwt(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cookie, err := r.Cookie("jwt")
		if err != nil {
			log.Println("here", err)
			http.Error(w, "not authorized", http.StatusInternalServerError)
			return
		}

		if cookie.Name != "jwt" {
			http.Error(w, "not authorized", http.StatusInternalServerError)
			return
		}

		claims, err := utils.Verify_JWT_Token(cookie.Value)

		if err != nil {
			http.Error(w, "not authorized", http.StatusInternalServerError)
			return
		}

		ctx := context.WithValue(r.Context(), "requesting_user_id", claims.ID)
		r = r.WithContext(ctx)

		next.ServeHTTP(w, r)
	})
}

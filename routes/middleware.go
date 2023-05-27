package routes

import (
	"github.com/adnanhashmi09/clique_server/utils"
	"log"
	"net/http"
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
		log.Println(cookie.Value)
		claims, err := utils.Verify_JWT_Token(cookie.Value)

		if err != nil {
			http.Error(w, "not authorized", http.StatusInternalServerError)
			return
		}

		log.Println(claims)
		next.ServeHTTP(w, r)
	})
}

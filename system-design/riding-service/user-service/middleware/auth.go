package middleware

import (
	"encoding/json"
	"net/http"
	"strings"

	"userservice/utils"
)

func JWTAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			json.NewEncoder(w).Encode(map[string]string{
				"error": "Authorization header required",
			})
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		bearerToken := strings.Split(authHeader, " ")
		if len(bearerToken) != 2 || bearerToken[0] != "Bearer" {
			json.NewEncoder(w).Encode(map[string]string{
				"error": "Invalid token format. Use: Bearer <token>",
			})
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		token, err := utils.ValidateJWT(bearerToken[1])
		if err != nil || !token.Valid {
			json.NewEncoder(w).Encode(map[string]string{
				"error": "Invalid or expired token",
			})
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		next.ServeHTTP(w, r)
	})
}

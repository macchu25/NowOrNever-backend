package middleware

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"

	"backend-go/utils"
)

type ContextKey string

const UserContextKey ContextKey = "userClaims"

func RequireAuth(next http.HandlerFunc) http.HandlerFunc {
	return EnableCORS(func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusUnauthorized)
			json.NewEncoder(w).Encode(map[string]string{
				"status":  "error",
				"message": "Authorization header missing",
			})
			return
		}

		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusUnauthorized)
			json.NewEncoder(w).Encode(map[string]string{
				"status":  "error",
				"message": "Invalid Authorization header format (expected Bearer <token>)",
			})
			return
		}

		claims, err := utils.ValidateToken(parts[1])
		if err != nil {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusUnauthorized)
			json.NewEncoder(w).Encode(map[string]string{
				"status":  "error",
				"message": "Invalid or expired token: " + err.Error(),
			})
			return
		}

		ctx := context.WithValue(r.Context(), UserContextKey, claims)
		next(w, r.WithContext(ctx))
	})
}

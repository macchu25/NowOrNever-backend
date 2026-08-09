package middleware

import (
	"net/http"
	"os"
	"strings"
)

func EnableCORS(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		origin := r.Header.Get("Origin")
		allowedOriginsEnv := os.Getenv("ALLOWED_ORIGINS")

		if origin != "" {
			isAllowed := false
			if allowedOriginsEnv == "" || allowedOriginsEnv == "*" {
				isAllowed = true
			} else {
				originsList := strings.Split(allowedOriginsEnv, ",")
				for _, o := range originsList {
					o = strings.TrimSpace(o)
					if o == origin || strings.HasSuffix(origin, ".vercel.app") || o == "*" {
						isAllowed = true
						break
					}
				}
			}

			if isAllowed {
				w.Header().Set("Access-Control-Allow-Origin", origin)
			} else {
				w.Header().Set("Access-Control-Allow-Origin", "*")
			}
		} else {
			w.Header().Set("Access-Control-Allow-Origin", "*")
		}

		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS, PATCH")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, X-Requested-With, Accept, Origin")
		w.Header().Set("Access-Control-Allow-Credentials", "true")

		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}

		next(w, r)
	}
}

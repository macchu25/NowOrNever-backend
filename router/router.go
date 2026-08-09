package router

import (
	"net/http"

	"backend-go/controllers"
	"backend-go/middleware"
)

func SetupRouter() *http.ServeMux {
	mux := http.NewServeMux()

	// Public Health & Telemetry
	mux.HandleFunc("GET /health", middleware.EnableCORS(controllers.HealthCheckHandler))
	mux.HandleFunc("GET /api/v1/stats", middleware.EnableCORS(controllers.StatsTelemetryHandler))

	// Auth Routes
	mux.HandleFunc("POST /api/v1/auth/register", middleware.EnableCORS(controllers.RegisterHandler))
	mux.HandleFunc("POST /api/v1/auth/login", middleware.EnableCORS(controllers.LoginHandler))
	mux.HandleFunc("POST /api/v1/auth/google", middleware.EnableCORS(controllers.GoogleAuthHandler))
	mux.HandleFunc("GET /api/v1/auth/me", middleware.RequireAuth(controllers.MeHandler))

	// Exercise DB Routes
	mux.HandleFunc("GET /api/v1/exercises", middleware.EnableCORS(controllers.GetExercisesHandler))
	mux.HandleFunc("GET /api/v1/exercises/detail", middleware.EnableCORS(controllers.GetExerciseByIDHandler))

	// Data Routes
	mux.HandleFunc("GET /api/v1/data", middleware.EnableCORS(controllers.DataHandler))
	mux.HandleFunc("POST /api/v1/data", middleware.EnableCORS(controllers.DataHandler))

	return mux
}

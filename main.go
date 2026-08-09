package main

import (
	"log"
	"net/http"
	"time"

	"backend-go/config"
	"backend-go/router"
)

// Main entry point for Now or Never Railway Go Backend
func main() {
	config.LoadConfig()

	appRouter := router.SetupRouter()

	server := &http.Server{
		Addr:         ":" + config.AppConfig.Port,
		Handler:      appRouter,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
	}

	log.Printf("🚀 [Now or Never Go Backend] Server running on port %s\n", config.AppConfig.Port)
	if err := server.ListenAndServe(); err != nil {
		log.Fatalf("❌ Server startup error: %v", err)
	}
}

package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"runtime"
	"sync"
	"time"
)

type Response struct {
	Status    string      `json:"status"`
	Timestamp string      `json:"timestamp"`
	Message   string      `json:"message"`
	Data      interface{} `json:"data,omitempty"`
}

type ServiceStats struct {
	GoVersion    string `json:"goVersion"`
	NumGoroutine int    `json:"numGoroutine"`
	MemoryAlloc  string `json:"memoryAllocMB"`
	TotalRequests uint64 `json:"totalRequests"`
	Uptime       string `json:"uptime"`
}

type DataPayload struct {
	Title       string `json:"title"`
	Category    string `json:"category"`
	Value       int    `json:"value"`
	Description string `json:"description"`
}

var (
	startTime     = time.Now()
	requestCount  uint64
	countMutex    sync.Mutex
	storedData    = []DataPayload{
		{Title: "Workout Session #1", Category: "Strength", Value: 45, Description: "Upper body hyper-trophy workout"},
		{Title: "Cardio Blitz", Category: "Cardio", Value: 30, Description: "High-intensity interval training"},
		{Title: "Recovery Stretch", Category: "Mobility", Value: 20, Description: "Full body flexibility session"},
	}
	dataMutex sync.Mutex
)

func enableCORS(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusOK)
			return
		}

		countMutex.Lock()
		requestCount++
		countMutex.Unlock()

		next(w, r)
	}
}

func sendJSON(w http.ResponseWriter, statusCode int, payload Response) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(payload)
}

func handleHealth(w http.ResponseWriter, r *http.Request) {
	sendJSON(w, http.StatusOK, Response{
		Status:    "success",
		Timestamp: time.Now().Format(time.RFC3339),
		Message:   "Go Backend service is healthy and operating normally",
	})
}

func handleStats(w http.ResponseWriter, r *http.Request) {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)

	stats := ServiceStats{
		GoVersion:    runtime.Version(),
		NumGoroutine: runtime.NumGoroutine(),
		MemoryAlloc:  fmt.Sprintf("%.2f MB", float64(m.Alloc)/(1024*1024)),
		TotalRequests: requestCount,
		Uptime:       time.Since(startTime).Round(time.Second).String(),
	}

	sendJSON(w, http.StatusOK, Response{
		Status:    "success",
		Timestamp: time.Now().Format(time.RFC3339),
		Message:   "Real-time system telemetry fetched successfully",
		Data:      stats,
	})
}

func handleData(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		dataMutex.Lock()
		items := make([]DataPayload, len(storedData))
		copy(items, storedData)
		dataMutex.Unlock()

		sendJSON(w, http.StatusOK, Response{
			Status:    "success",
			Timestamp: time.Now().Format(time.RFC3339),
			Message:   "Data items retrieved successfully",
			Data:      items,
		})

	case http.MethodPost:
		var payload DataPayload
		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
			sendJSON(w, http.StatusBadRequest, Response{
				Status:    "error",
				Timestamp: time.Now().Format(time.RFC3339),
				Message:   "Invalid JSON payload format",
			})
			return
		}

		if payload.Title == "" {
			sendJSON(w, http.StatusBadRequest, Response{
				Status:    "error",
				Timestamp: time.Now().Format(time.RFC3339),
				Message:   "Field 'title' is required",
			})
			return
		}

		dataMutex.Lock()
		storedData = append(storedData, payload)
		dataMutex.Unlock()

		sendJSON(w, http.StatusCreated, Response{
			Status:    "success",
			Timestamp: time.Now().Format(time.RFC3339),
			Message:   "New data item added successfully",
			Data:      payload,
		})

	default:
		sendJSON(w, http.StatusMethodNotAllowed, Response{
			Status:    "error",
			Timestamp: time.Now().Format(time.RFC3339),
			Message:   "Method not allowed",
		})
	}
}

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	mux := http.NewServeMux()
	mux.HandleFunc("GET /health", enableCORS(handleHealth))
	mux.HandleFunc("GET /api/v1/stats", enableCORS(handleStats))
	mux.HandleFunc("GET /api/v1/data", enableCORS(handleData))
	mux.HandleFunc("POST /api/v1/data", enableCORS(handleData))

	server := &http.Server{
		Addr:         ":" + port,
		Handler:      mux,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
	}

	log.Printf("🚀 [Go Backend] Server listening on http://localhost:%s\n", port)
	if err := server.ListenAndServe(); err != nil {
		log.Fatalf("❌ Go Backend server error: %v", err)
	}
}

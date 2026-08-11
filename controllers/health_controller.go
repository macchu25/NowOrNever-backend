package controllers

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"runtime"
	"sync/atomic"
	"time"

	"backend-go/config"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/readpref"
)

var (
	startTime    = time.Now()
	requestCount uint64
)

type ServiceStats struct {
	GoVersion      string `json:"goVersion"`
	NumGoroutine   int    `json:"numGoroutine"`
	MemoryAlloc    string `json:"memoryAllocMB"`
	TotalRequests  uint64 `json:"totalRequests"`
	Uptime         string `json:"uptime"`
	MongoDBStatus  string `json:"mongoDBStatus"`
	TotalDocuments int64  `json:"totalDocuments"`
	Environment    string `json:"environment"`
}

func HealthCheckHandler(w http.ResponseWriter, r *http.Request) {
	mongoStatus := "Disconnected"
	if config.MongoConnected && config.MongoClient != nil {
		ctx, cancel := context.WithTimeout(r.Context(), 3*time.Second)
		defer cancel()
		if err := config.MongoClient.Ping(ctx, readpref.Primary()); err == nil {
			mongoStatus = "Connected (MongoDB Atlas)"
		}
	}

	sendJSONResponse(w, http.StatusOK, map[string]interface{}{
		"status":    "success",
		"timestamp": time.Now().Format(time.RFC3339),
		"message":   "Now or Never Go Backend Service is healthy",
		"data": map[string]interface{}{
			"mongoDB":     mongoStatus,
			"railwayHost": os.Getenv("RAILWAY_STATIC_URL"),
			"uptime":      time.Since(startTime).Round(time.Second).String(),
		},
	})
}

func StatsTelemetryHandler(w http.ResponseWriter, r *http.Request) {
	currentRequests := atomic.AddUint64(&requestCount, 1)

	var m runtime.MemStats
	runtime.ReadMemStats(&m)

	mongoStatus := "Disconnected"
	var totalDocs int64 = 0

	if config.MongoConnected && config.DB != nil {
		ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
		defer cancel()
		if count, err := config.DB.Collection("activities").CountDocuments(ctx, bson.M{}); err == nil {
			totalDocs = count
			mongoStatus = "Connected"
		}
	}

	envName := "Local"
	if os.Getenv("RAILWAY_STATIC_URL") != "" || os.Getenv("RAILWAY_ENVIRONMENT") != "" {
		envName = "Railway Production"
	}

	stats := ServiceStats{
		GoVersion:      runtime.Version(),
		NumGoroutine:   runtime.NumGoroutine(),
		MemoryAlloc:    fmt.Sprintf("%.2f MB", float64(m.Alloc)/(1024*1024)),
		TotalRequests:  currentRequests,
		Uptime:         time.Since(startTime).Round(time.Second).String(),
		MongoDBStatus:  mongoStatus,
		TotalDocuments: totalDocs,
		Environment:    envName,
	}

	sendJSONResponse(w, http.StatusOK, map[string]interface{}{
		"status":    "success",
		"timestamp": time.Now().Format(time.RFC3339),
		"message":   "System telemetry & MongoDB stats fetched successfully",
		"data":      stats,
	})
}

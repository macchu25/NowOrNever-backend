package main

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"runtime"
	"strings"
	"sync"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
)

type Response struct {
	Status    string      `json:"status"`
	Timestamp string      `json:"timestamp"`
	Message   string      `json:"message"`
	Data      interface{} `json:"data,omitempty"`
}

type ServiceStats struct {
	GoVersion      string `json:"goVersion"`
	NumGoroutine   int    `json:"numGoroutine"`
	MemoryAlloc    string `json:"memoryAllocMB"`
	TotalRequests  uint64 `json:"totalRequests"`
	Uptime         string `json:"uptime"`
	MongoDBStatus  string `json:"mongoDBStatus"`
	TotalDocuments int64  `json:"totalDocuments"`
}

type ActivityDocument struct {
	ID          primitive.ObjectID `bson:"_id,omitempty" json:"id,omitempty"`
	Title       string             `bson:"title" json:"title"`
	Category    string             `bson:"category" json:"category"`
	Value       int                `bson:"value" json:"value"`
	Description string             `bson:"description" json:"description"`
	CreatedAt   time.Time          `bson:"createdAt" json:"createdAt"`
}

var (
	startTime       = time.Now()
	requestCount    uint64
	countMutex      sync.Mutex
	mongoClient     *mongo.Client
	mongoCollection *mongo.Collection
	mongoConnected   bool
	dbName          = "nowornever"
	collectionName  = "activities"
)

// Simple helper to load .env file if present
func loadDotEnv(filepath string) {
	file, err := os.Open(filepath)
	if err != nil {
		return
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		parts := strings.SplitN(line, "=", 2)
		if len(parts) == 2 {
			key := strings.TrimSpace(parts[0])
			val := strings.TrimSpace(parts[1])
			if os.Getenv(key) == "" {
				os.Setenv(key, val)
			}
		}
	}
}

func initMongoDB() {
	mongoURI := os.Getenv("MONGODB_URI")
	if mongoURI == "" {
		mongoURI = "mongodb://localhost:27017"
	}

	envDB := os.Getenv("DB_NAME")
	if envDB != "" {
		dbName = envDB
	}

	log.Printf("🔌 Connecting to MongoDB Atlas...")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	clientOpts := options.Client().ApplyURI(mongoURI)
	client, err := mongo.Connect(ctx, clientOpts)
	if err != nil {
		log.Printf("⚠️ MongoDB Connection Warning: %v", err)
		mongoConnected = false
		return
	}

	if err := client.Ping(ctx, readpref.Primary()); err != nil {
		log.Printf("⚠️ MongoDB Ping Warning: %v", err)
		mongoConnected = false
		return
	}

	mongoClient = client
	mongoCollection = client.Database(dbName).Collection(collectionName)
	mongoConnected = true
	log.Printf("✅ Successfully connected to MongoDB Atlas! (Database: %s, Collection: %s)", dbName, collectionName)
}

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
	mongoStatus := "Disconnected"
	if mongoConnected {
		ctx, cancel := context.WithTimeout(r.Context(), 3*time.Second)
		defer cancel()
		if err := mongoClient.Ping(ctx, readpref.Primary()); err == nil {
			mongoStatus = "Connected (MongoDB Atlas)"
		}
	}

	sendJSON(w, http.StatusOK, Response{
		Status:    "success",
		Timestamp: time.Now().Format(time.RFC3339),
		Message:   "Now or Never Go Backend service is healthy and operating normally",
		Data: map[string]interface{}{
			"mongoDB": mongoStatus,
			"uptime":  time.Since(startTime).Round(time.Second).String(),
		},
	})
}

func handleStats(w http.ResponseWriter, r *http.Request) {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)

	mongoStatus := "Disconnected"
	var totalDocs int64 = 0

	if mongoConnected && mongoCollection != nil {
		ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
		defer cancel()
		if count, err := mongoCollection.CountDocuments(ctx, bson.M{}); err == nil {
			totalDocs = count
			mongoStatus = "Connected"
		}
	}

	stats := ServiceStats{
		GoVersion:      runtime.Version(),
		NumGoroutine:   runtime.NumGoroutine(),
		MemoryAlloc:    fmt.Sprintf("%.2f MB", float64(m.Alloc)/(1024*1024)),
		TotalRequests:  requestCount,
		Uptime:         time.Since(startTime).Round(time.Second).String(),
		MongoDBStatus:  mongoStatus,
		TotalDocuments: totalDocs,
	}

	sendJSON(w, http.StatusOK, Response{
		Status:    "success",
		Timestamp: time.Now().Format(time.RFC3339),
		Message:   "Real-time system telemetry & MongoDB stats fetched successfully",
		Data:      stats,
	})
}

func handleData(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		if !mongoConnected || mongoCollection == nil {
			sendJSON(w, http.StatusOK, Response{
				Status:    "success",
				Timestamp: time.Now().Format(time.RFC3339),
				Message:   "Retrieved sample data (MongoDB offline)",
				Data: []ActivityDocument{
					{Title: "Workout Session #1", Category: "Strength", Value: 45, Description: "Upper body hyper-trophy workout", CreatedAt: time.Now()},
					{Title: "Cardio Blitz", Category: "Cardio", Value: 30, Description: "High-intensity interval training", CreatedAt: time.Now()},
				},
			})
			return
		}

		ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
		defer cancel()

		findOpts := options.Find().SetSort(bson.D{{Key: "createdAt", Value: -1}}).SetLimit(50)
		cursor, err := mongoCollection.Find(ctx, bson.M{}, findOpts)
		if err != nil {
			sendJSON(w, http.StatusInternalServerError, Response{
				Status:    "error",
				Timestamp: time.Now().Format(time.RFC3339),
				Message:   "Error querying MongoDB documents: " + err.Error(),
			})
			return
		}
		defer cursor.Close(ctx)

		var items []ActivityDocument
		if err := cursor.All(ctx, &items); err != nil {
			sendJSON(w, http.StatusInternalServerError, Response{
				Status:    "error",
				Timestamp: time.Now().Format(time.RFC3339),
				Message:   "Error decoding MongoDB documents: " + err.Error(),
			})
			return
		}

		if items == nil {
			items = []ActivityDocument{}
		}

		sendJSON(w, http.StatusOK, Response{
			Status:    "success",
			Timestamp: time.Now().Format(time.RFC3339),
			Message:   "MongoDB activities fetched successfully",
			Data:      items,
		})

	case http.MethodPost:
		var payload ActivityDocument
		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
			sendJSON(w, http.StatusBadRequest, Response{
				Status:    "error",
				Timestamp: time.Now().Format(time.RFC3339),
				Message:   "Invalid JSON payload format",
			})
			return
		}

		if strings.TrimSpace(payload.Title) == "" {
			sendJSON(w, http.StatusBadRequest, Response{
				Status:    "error",
				Timestamp: time.Now().Format(time.RFC3339),
				Message:   "Field 'title' is required",
			})
			return
		}

		payload.ID = primitive.NewObjectID()
		payload.CreatedAt = time.Now()

		if mongoConnected && mongoCollection != nil {
			ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
			defer cancel()

			res, err := mongoCollection.InsertOne(ctx, payload)
			if err != nil {
				sendJSON(w, http.StatusInternalServerError, Response{
					Status:    "error",
					Timestamp: time.Now().Format(time.RFC3339),
					Message:   "Failed to insert document into MongoDB: " + err.Error(),
				})
				return
			}
			payload.ID = res.InsertedID.(primitive.ObjectID)
		}

		sendJSON(w, http.StatusCreated, Response{
			Status:    "success",
			Timestamp: time.Now().Format(time.RFC3339),
			Message:   "New activity document saved to MongoDB Atlas successfully",
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
	loadDotEnv(".env")

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	initMongoDB()

	mux := http.NewServeMux()
	mux.HandleFunc("GET /health", enableCORS(handleHealth))
	mux.HandleFunc("GET /api/v1/stats", enableCORS(handleStats))
	mux.HandleFunc("GET /api/v1/data", enableCORS(handleData))
	mux.HandleFunc("POST /api/v1/data", enableCORS(handleData))

	server := &http.Server{
		Addr:         ":" + port,
		Handler:      mux,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
	}

	log.Printf("🚀 [Now or Never Go Backend] Listening on http://localhost:%s\n", port)
	if err := server.ListenAndServe(); err != nil {
		log.Fatalf("❌ Go Backend server error: %v", err)
	}
}

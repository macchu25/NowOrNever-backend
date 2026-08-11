package config

import (
	"bufio"
	"context"
	"log"
	"os"
	"strings"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
)

type Config struct {
	Port         string
	MongoURI     string
	DBName       string
	JWTSecret    string
	GoogleClientID string
	AllowedOrigins string
}

var AppConfig Config
var MongoClient *mongo.Client
var DB *mongo.Database
var MongoConnected bool

func LoadConfig() {
	loadDotEnv(".env")

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	mongoURI := os.Getenv("MONGODB_URI")
	if mongoURI == "" {
		log.Println("⚠️ MONGODB_URI not set in environment. Falling back to local MongoDB (mongodb://localhost:27017)...")
		mongoURI = "mongodb://localhost:27017"
	}

	dbName := os.Getenv("DB_NAME")
	if dbName == "" {
		dbName = "nowornever"
	}

	jwtSecret := os.Getenv("JWT_SECRET")
	if jwtSecret == "" {
		log.Println("⚠️ JWT_SECRET not set in environment. Using default development secret...")
		jwtSecret = "nowornever_dev_jwt_secret_change_in_production"
	}

	AppConfig = Config{
		Port:           port,
		MongoURI:       mongoURI,
		DBName:         dbName,
		JWTSecret:      jwtSecret,
		GoogleClientID: os.Getenv("GOOGLE_CLIENT_ID"),
		AllowedOrigins: os.Getenv("ALLOWED_ORIGINS"),
	}

	initMongoDB()
}

func initMongoDB() {
	log.Printf("🔌 Connecting to MongoDB Atlas...")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	clientOpts := options.Client().ApplyURI(AppConfig.MongoURI)
	client, err := mongo.Connect(ctx, clientOpts)
	if err != nil {
		log.Printf("⚠️ MongoDB Connection Error: %v", err)
		MongoConnected = false
		return
	}

	if err := client.Ping(ctx, readpref.Primary()); err != nil {
		log.Printf("⚠️ MongoDB Ping Warning: %v", err)
		MongoConnected = false
		return
	}

	MongoClient = client
	DB = client.Database(AppConfig.DBName)
	MongoConnected = true
	log.Printf("✅ MongoDB Atlas Connected! Database: %s", AppConfig.DBName)
}

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

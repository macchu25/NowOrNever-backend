package controllers

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"backend-go/config"
	"backend-go/models"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func DataHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		if !config.MongoConnected || config.DB == nil {
			sendJSONResponse(w, http.StatusOK, map[string]interface{}{
				"status":    "success",
				"timestamp": time.Now().Format(time.RFC3339),
				"message":   "Retrieved sample data (MongoDB offline)",
				"data": []models.Activity{
					{Title: "Workout Session #1", Category: "Strength", Value: 45, Description: "Upper body hyper-trophy workout", CreatedAt: time.Now()},
					{Title: "Cardio Blitz", Category: "Cardio", Value: 30, Description: "High-intensity interval training", CreatedAt: time.Now()},
				},
			})
			return
		}

		ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
		defer cancel()

		coll := config.DB.Collection("activities")
		findOpts := options.Find().SetSort(bson.D{{Key: "createdAt", Value: -1}}).SetLimit(50)
		cursor, err := coll.Find(ctx, bson.M{}, findOpts)
		if err != nil {
			sendJSONResponse(w, http.StatusInternalServerError, map[string]interface{}{
				"status":  "error",
				"message": "Error querying MongoDB documents: " + err.Error(),
			})
			return
		}
		defer cursor.Close(ctx)

		var items []models.Activity
		if err := cursor.All(ctx, &items); err != nil {
			sendJSONResponse(w, http.StatusInternalServerError, map[string]interface{}{
				"status":  "error",
				"message": "Error decoding MongoDB documents: " + err.Error(),
			})
			return
		}

		if items == nil {
			items = []models.Activity{}
		}

		sendJSONResponse(w, http.StatusOK, map[string]interface{}{
			"status":    "success",
			"timestamp": time.Now().Format(time.RFC3339),
			"message":   "MongoDB activities fetched successfully",
			"data":      items,
		})

	case http.MethodPost:
		var payload models.Activity
		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
			sendJSONResponse(w, http.StatusBadRequest, map[string]interface{}{
				"status":  "error",
				"message": "Invalid JSON payload format",
			})
			return
		}

		if strings.TrimSpace(payload.Title) == "" {
			sendJSONResponse(w, http.StatusBadRequest, map[string]interface{}{
				"status":  "error",
				"message": "Field 'title' is required",
			})
			return
		}

		payload.ID = primitive.NewObjectID()
		payload.CreatedAt = time.Now()

		if config.MongoConnected && config.DB != nil {
			ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
			defer cancel()

			res, err := config.DB.Collection("activities").InsertOne(ctx, payload)
			if err != nil {
				sendJSONResponse(w, http.StatusInternalServerError, map[string]interface{}{
					"status":  "error",
					"message": "Failed to insert document into MongoDB: " + err.Error(),
				})
				return
			}
			payload.ID = res.InsertedID.(primitive.ObjectID)
		}

		sendJSONResponse(w, http.StatusCreated, map[string]interface{}{
			"status":    "success",
			"timestamp": time.Now().Format(time.RFC3339),
			"message":   "New activity document saved to MongoDB Atlas successfully",
			"data":      payload,
		})

	default:
		sendJSONResponse(w, http.StatusMethodNotAllowed, map[string]interface{}{
			"status":  "error",
			"message": "Method not allowed",
		})
	}
}

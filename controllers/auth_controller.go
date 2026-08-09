package controllers

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"backend-go/config"
	"backend-go/middleware"
	"backend-go/models"
	"backend-go/utils"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

func sendJSONResponse(w http.ResponseWriter, status int, payload interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(payload)
}

func RegisterHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		sendJSONResponse(w, http.StatusMethodNotAllowed, map[string]string{"status": "error", "message": "Method not allowed"})
		return
	}

	var req models.RegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		sendJSONResponse(w, http.StatusBadRequest, map[string]string{"status": "error", "message": "Invalid JSON request payload"})
		return
	}

	req.Email = strings.ToLower(strings.TrimSpace(req.Email))
	req.Name = strings.TrimSpace(req.Name)

	if req.Email == "" || req.Password == "" || req.Name == "" {
		sendJSONResponse(w, http.StatusBadRequest, map[string]string{"status": "error", "message": "All fields (name, email, password) are required"})
		return
	}

	userID := primitive.NewObjectID()
	newUser := models.User{
		ID:        userID,
		Email:     req.Email,
		Name:      req.Name,
		Avatar:    "https://api.dicebear.com/7.x/bottts/svg?seed=" + req.Name,
		Provider:  "local",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	if config.MongoConnected && config.DB != nil {
		usersColl := config.DB.Collection("users")
		ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
		defer cancel()

		var existingUser models.User
		err := usersColl.FindOne(ctx, bson.M{"email": req.Email}).Decode(&existingUser)
		if err == nil {
			sendJSONResponse(w, http.StatusConflict, map[string]string{"status": "error", "message": "Email is already registered. Please login instead."})
			return
		}

		hashedPassword, err := utils.HashPassword(req.Password)
		if err != nil {
			sendJSONResponse(w, http.StatusInternalServerError, map[string]string{"status": "error", "message": "Error hashing password"})
			return
		}
		newUser.PasswordHash = hashedPassword

		_, err = usersColl.InsertOne(ctx, newUser)
		if err != nil {
			sendJSONResponse(w, http.StatusInternalServerError, map[string]string{"status": "error", "message": "Failed to create user in database"})
			return
		}
	}

	token, err := utils.GenerateToken(newUser.ID, newUser.Email, newUser.Name)
	if err != nil {
		sendJSONResponse(w, http.StatusInternalServerError, map[string]string{"status": "error", "message": "Failed to generate token"})
		return
	}

	sendJSONResponse(w, http.StatusCreated, map[string]interface{}{
		"status":  "success",
		"message": "User registered successfully",
		"token":   token,
		"user":    newUser,
	})
}

func LoginHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		sendJSONResponse(w, http.StatusMethodNotAllowed, map[string]string{"status": "error", "message": "Method not allowed"})
		return
	}

	var req models.LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		sendJSONResponse(w, http.StatusBadRequest, map[string]string{"status": "error", "message": "Invalid JSON request payload"})
		return
	}

	req.Email = strings.ToLower(strings.TrimSpace(req.Email))

	if req.Email == "" || req.Password == "" {
		sendJSONResponse(w, http.StatusBadRequest, map[string]string{"status": "error", "message": "Email and password are required"})
		return
	}

	var user models.User
	if config.MongoConnected && config.DB != nil {
		usersColl := config.DB.Collection("users")
		ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
		defer cancel()

		err := usersColl.FindOne(ctx, bson.M{"email": req.Email}).Decode(&user)
		if err != nil {
			if err == mongo.ErrNoDocuments {
				sendJSONResponse(w, http.StatusUnauthorized, map[string]string{"status": "error", "message": "Invalid email or password"})
				return
			}
			sendJSONResponse(w, http.StatusInternalServerError, map[string]string{"status": "error", "message": "Database error"})
			return
		}

		if user.PasswordHash == "" && user.Provider == "google" {
			sendJSONResponse(w, http.StatusBadRequest, map[string]string{"status": "error", "message": "This account was created via Google Login. Please sign in with Google."})
			return
		}

		if !utils.CheckPasswordHash(req.Password, user.PasswordHash) {
			sendJSONResponse(w, http.StatusUnauthorized, map[string]string{"status": "error", "message": "Invalid email or password"})
			return
		}
	} else {
		// Memory fallback user for testing if DB offline
		user = models.User{
			ID:        primitive.NewObjectID(),
			Email:     req.Email,
			Name:      strings.Split(req.Email, "@")[0],
			Avatar:    "https://api.dicebear.com/7.x/bottts/svg?seed=" + req.Email,
			Provider:  "local",
			CreatedAt: time.Now(),
		}
	}

	token, err := utils.GenerateToken(user.ID, user.Email, user.Name)
	if err != nil {
		sendJSONResponse(w, http.StatusInternalServerError, map[string]string{"status": "error", "message": "Failed to generate token"})
		return
	}

	sendJSONResponse(w, http.StatusOK, map[string]interface{}{
		"status":  "success",
		"message": "Logged in successfully",
		"token":   token,
		"user":    user,
	})
}

func GoogleAuthHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		sendJSONResponse(w, http.StatusMethodNotAllowed, map[string]string{"status": "error", "message": "Method not allowed"})
		return
	}

	var req models.GoogleAuthRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		sendJSONResponse(w, http.StatusBadRequest, map[string]string{"status": "error", "message": "Invalid JSON request payload"})
		return
	}

	req.Email = strings.ToLower(strings.TrimSpace(req.Email))
	if req.Email == "" {
		req.Email = "google.user@gmail.com"
	}

	name := req.Name
	if name == "" {
		name = "Google Member"
	}
	avatar := req.Avatar
	if avatar == "" {
		avatar = "https://lh3.googleusercontent.com/a/default-user=s96-c"
	}

	userID := primitive.NewObjectID()
	user := models.User{
		ID:        userID,
		Email:     req.Email,
		Name:      name,
		Avatar:    avatar,
		GoogleID:  req.GoogleID,
		Provider:  "google",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	if config.MongoConnected && config.DB != nil {
		usersColl := config.DB.Collection("users")
		ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
		defer cancel()

		var existingUser models.User
		err := usersColl.FindOne(ctx, bson.M{"email": req.Email}).Decode(&existingUser)

		if err == mongo.ErrNoDocuments {
			usersColl.InsertOne(ctx, user)
		} else if err == nil {
			user = existingUser
			updateFields := bson.M{"updatedAt": time.Now()}
			if req.Avatar != "" {
				updateFields["avatar"] = req.Avatar
			}
			usersColl.UpdateOne(ctx, bson.M{"_id": user.ID}, bson.M{"$set": updateFields})
		}
	}

	token, err := utils.GenerateToken(user.ID, user.Email, user.Name)
	if err != nil {
		sendJSONResponse(w, http.StatusInternalServerError, map[string]string{"status": "error", "message": "Failed to generate token"})
		return
	}

	sendJSONResponse(w, http.StatusOK, map[string]interface{}{
		"status":  "success",
		"message": "Google authentication successful",
		"token":   token,
		"user":    user,
	})
}

func MeHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		sendJSONResponse(w, http.StatusMethodNotAllowed, map[string]string{"status": "error", "message": "Method not allowed"})
		return
	}

	claims, ok := r.Context().Value(middleware.UserContextKey).(*utils.JWTClaims)
	if !ok || claims == nil {
		sendJSONResponse(w, http.StatusUnauthorized, map[string]string{"status": "error", "message": "Unauthorized"})
		return
	}

	objID, err := primitive.ObjectIDFromHex(claims.UserID)
	if err != nil {
		sendJSONResponse(w, http.StatusBadRequest, map[string]string{"status": "error", "message": "Invalid user ID"})
		return
	}

	user := models.User{
		ID:        objID,
		Email:     claims.Email,
		Name:      claims.Name,
		Avatar:    "https://api.dicebear.com/7.x/bottts/svg?seed=" + claims.Name,
		Provider:  "local",
		CreatedAt: time.Now(),
	}

	if config.MongoConnected && config.DB != nil {
		usersColl := config.DB.Collection("users")
		ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
		defer cancel()

		var dbUser models.User
		err = usersColl.FindOne(ctx, bson.M{"_id": objID}).Decode(&dbUser)
		if err == nil {
			user = dbUser
		}
	}

	sendJSONResponse(w, http.StatusOK, map[string]interface{}{
		"status": "success",
		"user":   user,
	})
}

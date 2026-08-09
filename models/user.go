package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type User struct {
	ID           primitive.ObjectID `bson:"_id,omitempty" json:"id,omitempty"`
	Email        string             `bson:"email" json:"email"`
	PasswordHash string             `bson:"passwordHash,omitempty" json:"-"`
	Name         string             `bson:"name" json:"name"`
	Avatar       string             `bson:"avatar,omitempty" json:"avatar,omitempty"`
	GoogleID     string             `bson:"googleId,omitempty" json:"googleId,omitempty"`
	Provider     string             `bson:"provider" json:"provider"` // "local" or "google"
	CreatedAt    time.Time          `bson:"createdAt" json:"createdAt"`
	UpdatedAt    time.Time          `bson:"updatedAt" json:"updatedAt"`
}

type RegisterRequest struct {
	Name     string `json:"name"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type GoogleAuthRequest struct {
	Credential string `json:"credential"` // Google ID Token
	Email      string `json:"email,omitempty"`
	Name       string `json:"name,omitempty"`
	Avatar     string `json:"avatar,omitempty"`
	GoogleID   string `json:"googleId,omitempty"`
}

type AuthResponse struct {
	Token string `json:"token"`
	User  User   `json:"user"`
}

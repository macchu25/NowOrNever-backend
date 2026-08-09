package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Survey struct {
	Height            float64   `bson:"height" json:"height"`                   // cm
	Weight            float64   `bson:"weight" json:"weight"`                   // kg
	Age               int       `bson:"age" json:"age"`                         // Age in years
	Gender            string    `bson:"gender" json:"gender"`                   // "male", "female", "other"
	Goal              string    `bson:"goal" json:"goal"`                       // "muscle", "fat_loss", "stamina", "maintain"
	ActivityLevel     string    `bson:"activityLevel" json:"activityLevel"`     // "sedentary", "light", "moderate", "active"
	WorkoutLocation   string    `bson:"workoutLocation" json:"workoutLocation"` // "home", "gym", "hybrid"
	Equipment         string    `bson:"equipment" json:"equipment"`             // "bodyweight", "dumbbells", "full_gym"
	Allergies         []string  `bson:"allergies" json:"allergies"`             // ["seafood", "dairy", "peanuts", ...]
	DaysPerWeek       int       `bson:"daysPerWeek" json:"daysPerWeek"`         // 3, 4, 5, 6
	ExperienceLevel   string    `bson:"experienceLevel" json:"experienceLevel"` // beginner, intermediate, advanced
	SessionMinutes    int       `bson:"sessionMinutes" json:"sessionMinutes"`
	AvailableDays     []string  `bson:"availableDays" json:"availableDays"`
	Preferences       []string  `bson:"preferences" json:"preferences"`
	DislikedExercises []string  `bson:"dislikedExercises" json:"dislikedExercises"`
	Limitations       []string  `bson:"limitations" json:"limitations"`
	PushUps           int       `bson:"pushUps" json:"pushUps"`
	Squats            int       `bson:"squats" json:"squats"`
	PlankSeconds      int       `bson:"plankSeconds" json:"plankSeconds"`
	BMI               float64   `bson:"bmi" json:"bmi"`   // Calculated BMI
	TDEE              float64   `bson:"tdee" json:"tdee"` // Calculated TDEE
	CompletedAt       time.Time `bson:"completedAt" json:"completedAt"`
}

type User struct {
	ID           primitive.ObjectID `bson:"_id,omitempty" json:"id,omitempty"`
	Email        string             `bson:"email" json:"email"`
	PasswordHash string             `bson:"passwordHash,omitempty" json:"-"`
	Name         string             `bson:"name" json:"name"`
	Avatar       string             `bson:"avatar,omitempty" json:"avatar,omitempty"`
	GoogleID     string             `bson:"googleId,omitempty" json:"googleId,omitempty"`
	Provider     string             `bson:"provider" json:"provider"` // "local" or "google"
	Survey       *Survey            `bson:"survey,omitempty" json:"survey,omitempty"`
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

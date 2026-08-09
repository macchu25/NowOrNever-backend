package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Exercise struct {
	ID               primitive.ObjectID `json:"id,omitempty" bson:"_id,omitempty"`
	ExerciseID       string             `json:"exerciseId" bson:"exerciseId"`
	Name             string             `json:"name" bson:"name"`
	NameVi           string             `json:"nameVi" bson:"nameVi"`
	BodyPart         string             `json:"bodyPart" bson:"bodyPart"`
	Equipment        string             `json:"equipment" bson:"equipment"`
	Target           string             `json:"target" bson:"target"`
	GIFUrl           string             `json:"gifUrl" bson:"gifUrl"`
	Image            string             `json:"image" bson:"image"`
	SecondaryMuscles []string           `json:"secondaryMuscles" bson:"secondaryMuscles"`
	Instructions     []string           `json:"instructions" bson:"instructions"`
	Sets             int                `json:"sets" bson:"sets"`
	Reps             int                `json:"reps" bson:"reps"`
	RestSeconds      int                `json:"restSeconds" bson:"restSeconds"`
	CaloriesBurned   int                `json:"caloriesBurned" bson:"caloriesBurned"`
	Difficulty       string             `json:"difficulty" bson:"difficulty"`
	Location         string             `json:"location" bson:"location"` // "home", "gym", "both"
	CreatedAt        time.Time          `json:"createdAt" bson:"createdAt"`
}

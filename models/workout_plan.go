package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type PlanExercise struct {
	ExerciseID  string `bson:"exerciseId" json:"exerciseId"`
	Name        string `bson:"name" json:"name"`
	NameVi      string `bson:"nameVi" json:"nameVi"`
	BodyPart    string `bson:"bodyPart" json:"bodyPart"`
	Equipment   string `bson:"equipment" json:"equipment"`
	Sets        int    `bson:"sets" json:"sets"`
	Reps        int    `bson:"reps" json:"reps"`
	RestSeconds int    `bson:"restSeconds" json:"restSeconds"`
	TargetRPE   int    `bson:"targetRpe" json:"targetRpe"`
	Reason      string `bson:"reason" json:"reason"`
}

type WorkoutDay struct {
	Day       int            `bson:"day" json:"day"`
	Focus     string         `bson:"focus" json:"focus"`
	RestDay   bool           `bson:"restDay" json:"restDay"`
	Exercises []PlanExercise `bson:"exercises" json:"exercises"`
}

type WorkoutPlan struct {
	ID              primitive.ObjectID `bson:"_id,omitempty" json:"id,omitempty"`
	UserID          primitive.ObjectID `bson:"userId,omitempty" json:"userId,omitempty"`
	Goal            string             `bson:"goal" json:"goal"`
	Level           string             `bson:"level" json:"level"`
	DaysPerWeek     int                `bson:"daysPerWeek" json:"daysPerWeek"`
	SessionMinutes  int                `bson:"sessionMinutes" json:"sessionMinutes"`
	Schedule        []WorkoutDay       `bson:"schedule" json:"schedule"`
	Warnings        []string           `bson:"warnings" json:"warnings"`
	ProgressionRule string             `bson:"progressionRule" json:"progressionRule"`
	GeneratedBy     string             `bson:"generatedBy" json:"generatedBy"`
	CreatedAt       time.Time          `bson:"createdAt" json:"createdAt"`
}

type GenerateWorkoutPlanRequest struct {
	Survey Survey `json:"survey"`
}

package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Activity struct {
	ID          primitive.ObjectID `bson:"_id,omitempty" json:"id,omitempty"`
	UserID      primitive.ObjectID `bson:"userId,omitempty" json:"userId,omitempty"`
	Title       string             `bson:"title" json:"title"`
	Category    string             `bson:"category" json:"category"`
	Value       int                `bson:"value" json:"value"`
	Description string             `bson:"description" json:"description"`
	CreatedAt   time.Time          `bson:"createdAt" json:"createdAt"`
}

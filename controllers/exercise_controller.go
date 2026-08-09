package controllers

import (
	"context"
	"net/http"
	"strings"
	"time"

	"backend-go/config"
	"backend-go/models"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// Default seed exercises from ExerciseDB schema
var defaultExercises = []models.Exercise{
	{
		ExerciseID: "ex_001",
		Name:       "Barbell Bench Press",
		NameVi:     "Đẩy Ngực Ngang Với Đòn Tạ (Bench Press)",
		BodyPart:   "chest",
		Equipment:  "barbell",
		Target:     "pectorals",
		GIFUrl:     "https://v2.exercisedb.io/image/3s-A1gB-Tz9m7g",
		Image:      "/images/workout_full_body.png",
		SecondaryMuscles: []string{"delts", "triceps"},
		Instructions: []string{
			"Nằm ngửa trên ghế tập bench ngang, chân đặt chắc chắn trên sàn.",
			"Nắm thanh đòn tạ rộng hơn chiều rộng vai một chút, nhấc đòn tạ ra khỏi giá đỡ.",
			"Hạ thanh tạ từ từ xuống giữa ngực hít vào.",
			"Đẩy thanh tạ mạnh lên vị trí ban đầu và thở ra.",
		},
		Sets:           4,
		Reps:           12,
		RestSeconds:    60,
		CaloriesBurned: 120,
		Difficulty:     "Intermediate",
	},
	{
		ExerciseID: "ex_002",
		Name:       "Bodyweight Squat",
		NameVi:     "Squat Tự Do Cá Nhân (Bodyweight Squat)",
		BodyPart:   "upper legs",
		Equipment:  "body weight",
		Target:     "quads",
		GIFUrl:     "https://v2.exercisedb.io/image/9s-B2gB-Tz9m7h",
		Image:      "/images/workout_leg.png",
		SecondaryMuscles: []string{"glutes", "hamstrings", "calves"},
		Instructions: []string{
			"Đứng thẳng, 2 chân rộng bằng vai, mũi chân hơi hướng ra ngoài.",
			"Đẩy hông ra sau và hạ đùi xuống song song với sàn nhà.",
			"Giữ lưng thẳng, ngực mở rộng.",
			"Dùng lực gót chân đẩy người đứng thẳng dậy.",
		},
		Sets:           4,
		Reps:           15,
		RestSeconds:    45,
		CaloriesBurned: 90,
		Difficulty:     "Beginner",
	},
	{
		ExerciseID: "ex_003",
		Name:       "Dumbbell Bicep Curl",
		NameVi:     "Cuốn Tay Trước Với Tạ Đơn (Bicep Curl)",
		BodyPart:   "upper arms",
		Equipment:  "dumbbell",
		Target:     "biceps",
		GIFUrl:     "https://v2.exercisedb.io/image/1s-C3gB-Tz9m7i",
		Image:      "/images/workout_upper.png",
		SecondaryMuscles: []string{"forearms"},
		Instructions: []string{
			"Đứng thẳng, mỗi tay cầm 1 quả tạ đơn, lòng bàn tay hướng về trước.",
			"Giữ bắp tay cố định sát thân người, cuốn tạ lên về phía vai.",
			"Gồng chặt cơ bắp tay ở đỉnh động tác trong 1 giây.",
			"Hạ tạ từ từ về vị trí ban đầu.",
		},
		Sets:           3,
		Reps:           12,
		RestSeconds:    45,
		CaloriesBurned: 70,
		Difficulty:     "Beginner",
	},
	{
		ExerciseID: "ex_004",
		Name:       "Abdominal Crunch / Plank",
		NameVi:     "Gập Bụng & Giữ Plank (Core Crunch)",
		BodyPart:   "waist",
		Equipment:  "body weight",
		Target:     "abs",
		GIFUrl:     "https://v2.exercisedb.io/image/4s-D4gB-Tz9m7j",
		Image:      "/images/workout_core.png",
		SecondaryMuscles: []string{"obliques", "lower back"},
		Instructions: []string{
			"Nằm trên thảm, đầu gối gập 90 độ, 2 tay đặt sau đầu.",
			"Gồng cơ bụng cuộn người nâng vai rời khỏi thảm.",
			"Giữ 1 giây ở đỉnh động tác rồi từ từ hạ người xuống.",
		},
		Sets:           4,
		Reps:           20,
		RestSeconds:    30,
		CaloriesBurned: 85,
		Difficulty:     "Beginner",
	},
	{
		ExerciseID: "ex_005",
		Name:       "Lat Pulldown",
		NameVi:     "Kéo Xô Nối Cáp Rộng Tay (Lat Pulldown)",
		BodyPart:   "back",
		Equipment:  "cable",
		Target:     "lats",
		GIFUrl:     "https://v2.exercisedb.io/image/5s-E5gB-Tz9m7k",
		Image:      "/images/workout_full_body.png",
		SecondaryMuscles: []string{"biceps", "rhomboids", "rear delts"},
		Instructions: []string{
			"Sit down at lat pulldown machine, grip bar wider than shoulder width.",
			"Pull bar down towards upper chest while squeezing shoulder blades together.",
			"Slowly return bar back up to starting position.",
		},
		Sets:           4,
		Reps:           12,
		RestSeconds:    60,
		CaloriesBurned: 110,
		Difficulty:     "Intermediate",
	},
}

func GetExercisesHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		sendJSONResponse(w, http.StatusMethodNotAllowed, map[string]string{"status": "error", "message": "Method not allowed"})
		return
	}

	bodyPart := strings.ToLower(r.URL.Query().Get("bodyPart"))
	query := strings.ToLower(r.URL.Query().Get("q"))

	var results []models.Exercise

	if config.MongoConnected && config.DB != nil {
		ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
		defer cancel()

		coll := config.DB.Collection("exercises")
		filter := bson.M{}
		if bodyPart != "" && bodyPart != "all" {
			filter["bodyPart"] = bodyPart
		}

		cursor, err := coll.Find(ctx, filter)
		if err == nil {
			cursor.All(ctx, &results)
		}
	}

	// Fallback to default exercises if database empty or offline
	if len(results) == 0 {
		for _, ex := range defaultExercises {
			matchBodyPart := bodyPart == "" || bodyPart == "all" || strings.ToLower(ex.BodyPart) == bodyPart
			matchQuery := query == "" || strings.Contains(strings.ToLower(ex.Name), query) || strings.Contains(strings.ToLower(ex.NameVi), query)
			if matchBodyPart && matchQuery {
				results = append(results, ex)
			}
		}
	}

	sendJSONResponse(w, http.StatusOK, map[string]interface{}{
		"status":  "success",
		"count":   len(results),
		"data":    results,
	})
}

func GetExerciseByIDHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		sendJSONResponse(w, http.StatusMethodNotAllowed, map[string]string{"status": "error", "message": "Method not allowed"})
		return
	}

	id := r.URL.Query().Get("id")
	if id == "" {
		sendJSONResponse(w, http.StatusBadRequest, map[string]string{"status": "error", "message": "Exercise ID query parameter required"})
		return
	}

	var found *models.Exercise

	if config.MongoConnected && config.DB != nil {
		ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
		defer cancel()

		coll := config.DB.Collection("exercises")
		objID, err := primitive.ObjectIDFromHex(id)
		if err == nil {
			var ex models.Exercise
			if err := coll.FindOne(ctx, bson.M{"_id": objID}).Decode(&ex); err == nil {
				found = &ex
			}
		}
	}

	if found == nil {
		for _, ex := range defaultExercises {
			if ex.ExerciseID == id || ex.ID.Hex() == id {
				found = &ex
				break
			}
		}
	}

	if found == nil {
		sendJSONResponse(w, http.StatusNotFound, map[string]string{"status": "error", "message": "Exercise not found"})
		return
	}

	sendJSONResponse(w, http.StatusOK, map[string]interface{}{
		"status": "success",
		"data":   found,
	})
}

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

// Default seed exercises from ExerciseDB & MuscleWiki schemas for Home & Gym
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
		Location:   "gym",
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
		Name:       "Bodyweight Push-Up",
		NameVi:     "Chống Đẩy Hít Đất Tự Do (Push-Up)",
		BodyPart:   "chest",
		Equipment:  "body weight",
		Target:     "pectorals",
		GIFUrl:     "https://v2.exercisedb.io/image/9s-B2gB-Tz9m7h",
		Image:      "/images/workout_full_body.png",
		Location:   "home",
		SecondaryMuscles: []string{"triceps", "core"},
		Instructions: []string{
			"Chống 2 tay xuống thảm rộng bằng vai, thân người tạo thành đường thẳng.",
			"Hạ ngực xuống sát thảm, khuỷu tay mở góc 45 độ.",
			"Dùng lực cơ ngực đẩy người trở lại vị trí ban đầu.",
		},
		Sets:           4,
		Reps:           15,
		RestSeconds:    45,
		CaloriesBurned: 95,
		Difficulty:     "Beginner",
	},
	{
		ExerciseID: "ex_003",
		Name:       "Dumbbell Chest Flyes",
		NameVi:     "Banh Ngực Với Tạ Đơn (Dumbbell Flyes)",
		BodyPart:   "chest",
		Equipment:  "dumbbell",
		Target:     "pectorals",
		GIFUrl:     "https://v2.exercisedb.io/image/1s-C3gB-Tz9m7i",
		Image:      "/images/workout_full_body.png",
		Location:   "both",
		SecondaryMuscles: []string{"front delts"},
		Instructions: []string{
			"Nằm trên ghế tập, mỗi tay cầm 1 quả tạ đơn duỗi vuông góc.",
			"Banh rộng 2 tay ra 2 bên dạng hình cánh cung hít vào.",
			"Khép tạ lại về giữa ngực gồng chặt cơ ngực và thở ra.",
		},
		Sets:           3,
		Reps:           12,
		RestSeconds:    45,
		CaloriesBurned: 85,
		Difficulty:     "Intermediate",
	},
	{
		ExerciseID: "ex_004",
		Name:       "Bodyweight Squat",
		NameVi:     "Squat Tự Do Cá Nhân (Bodyweight Squat)",
		BodyPart:   "upper legs",
		Equipment:  "body weight",
		Target:     "quads",
		GIFUrl:     "https://v2.exercisedb.io/image/9s-B2gB-Tz9m7h",
		Image:      "/images/workout_leg.png",
		Location:   "home",
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
		ExerciseID: "ex_005",
		Name:       "Barbell Back Squat",
		NameVi:     "Gánh Tạ Đòn Đùi Sau (Barbell Back Squat)",
		BodyPart:   "upper legs",
		Equipment:  "barbell",
		Target:     "quads",
		GIFUrl:     "https://v2.exercisedb.io/image/3s-A1gB-Tz9m7g",
		Image:      "/images/workout_leg.png",
		Location:   "gym",
		SecondaryMuscles: []string{"glutes", "lower back"},
		Instructions: []string{
			"Đặt thanh đòn tạ lên phần cơ cầu vai (Upper Traps).",
			"Gồng chắc cơ bụng, hạ đùi xuống sâu dưới góc 90 độ.",
			"Đẩy mạnh đùi đứng dậy hết hành trình.",
		},
		Sets:           4,
		Reps:           10,
		RestSeconds:    90,
		CaloriesBurned: 150,
		Difficulty:     "Advanced",
	},
	{
		ExerciseID: "ex_006",
		Name:       "Dumbbell Romanian Deadlift",
		NameVi:     "Deadlift Tạ Đơn Đùi Sau (Dumbbell RDL)",
		BodyPart:   "upper legs",
		Equipment:  "dumbbell",
		Target:     "hamstrings",
		GIFUrl:     "https://v2.exercisedb.io/image/1s-C3gB-Tz9m7i",
		Image:      "/images/workout_leg.png",
		Location:   "both",
		SecondaryMuscles: []string{"glutes", "lower back"},
		Instructions: []string{
			"Đứng thẳng cầm 2 quả tạ đơn trước đùi, chân rộng bằng hông.",
			"Đẩy hông ra sau tối đa, hạ tạ dọc theo cẳng chân cho đến khi đùi sau căng dãn.",
			"Kéo hông về trước để đứng thẳng.",
		},
		Sets:           4,
		Reps:           12,
		RestSeconds:    60,
		CaloriesBurned: 110,
		Difficulty:     "Intermediate",
	},
	{
		ExerciseID: "ex_007",
		Name:       "Dumbbell Bicep Curl",
		NameVi:     "Cuốn Tay Trước Với Tạ Đơn (Bicep Curl)",
		BodyPart:   "upper arms",
		Equipment:  "dumbbell",
		Target:     "biceps",
		GIFUrl:     "https://v2.exercisedb.io/image/1s-C3gB-Tz9m7i",
		Image:      "/images/workout_upper.png",
		Location:   "both",
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
		ExerciseID: "ex_008",
		Name:       "Bench Tricep Dips",
		NameVi:     "Nhún Tay Sau Với Ghế (Bench Dips)",
		BodyPart:   "upper arms",
		Equipment:  "body weight",
		Target:     "triceps",
		GIFUrl:     "https://v2.exercisedb.io/image/9s-B2gB-Tz9m7h",
		Image:      "/images/workout_upper.png",
		Location:   "home",
		SecondaryMuscles: []string{"chest", "shoulders"},
		Instructions: []string{
			"Đặt 2 tay lên mép ghế/bậc thảm sau lưng, chân duỗi về trước.",
			"Hạ hông xuống sâu cho khuỷu tay gập góc 90 độ.",
			"Dùng lực tay sau duỗi thẳng tay nâng người lên.",
		},
		Sets:           3,
		Reps:           15,
		RestSeconds:    45,
		CaloriesBurned: 75,
		Difficulty:     "Beginner",
	},
	{
		ExerciseID: "ex_009",
		Name:       "Abdominal Crunch / Plank",
		NameVi:     "Gập Bụng & Giữ Plank (Core Crunch)",
		BodyPart:   "waist",
		Equipment:  "body weight",
		Target:     "abs",
		GIFUrl:     "https://v2.exercisedb.io/image/4s-D4gB-Tz9m7j",
		Image:      "/images/workout_core.png",
		Location:   "home",
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
		ExerciseID: "ex_010",
		Name:       "Lat Pulldown",
		NameVi:     "Kéo Xô Nối Cáp Rộng Tay (Lat Pulldown)",
		BodyPart:   "back",
		Equipment:  "cable",
		Target:     "lats",
		GIFUrl:     "https://v2.exercisedb.io/image/5s-E5gB-Tz9m7k",
		Image:      "/images/workout_full_body.png",
		Location:   "gym",
		SecondaryMuscles: []string{"biceps", "rhomboids", "rear delts"},
		Instructions: []string{
			"Nằm ngồi trên máy kéo cáp xô, tay nắm thanh kéo rộng hơn vai.",
			"Kéo thanh cáp xuống sát ngực trên, gồng chặt 2 bả vai.",
			"Trả thanh cáp lên từ từ.",
		},
		Sets:           4,
		Reps:           12,
		RestSeconds:    60,
		CaloriesBurned: 110,
		Difficulty:     "Intermediate",
	},
	{
		ExerciseID: "ex_011",
		Name:       "Resistance Band Standing Row",
		NameVi:     "Kéo Xô Đứng Với Dây Kháng Lực (Band Row)",
		BodyPart:   "back",
		Equipment:  "band",
		Target:     "lats",
		GIFUrl:     "https://v2.exercisedb.io/image/5s-E5gB-Tz9m7k",
		Image:      "/images/workout_full_body.png",
		Location:   "home",
		SecondaryMuscles: []string{"biceps", "rear delts"},
		Instructions: []string{
			"Móc dây kháng lực vào điểm cố định hoặc dẫm 2 chân lên giữa dây.",
			"Kéo 2 tay cầm dây về phía hông, ép sát bả vai.",
			"Thả dây về từ từ.",
		},
		Sets:           4,
		Reps:           15,
		RestSeconds:    45,
		CaloriesBurned: 80,
		Difficulty:     "Beginner",
	},
	{
		ExerciseID: "ex_012",
		Name:       "Dumbbell Shoulder Press",
		NameVi:     "Đẩy Vai Ngồi Với Tạ Đơn (Shoulder Press)",
		BodyPart:   "shoulders",
		Equipment:  "dumbbell",
		Target:     "delts",
		GIFUrl:     "https://v2.exercisedb.io/image/1s-C3gB-Tz9m7i",
		Image:      "/images/workout_upper.png",
		Location:   "both",
		SecondaryMuscles: []string{"triceps", "upper chest"},
		Instructions: []string{
			"Ngồi thẳng lưng, 2 tay cầm tạ đơn ngang chiều cao tai.",
			"Đẩy tạ thẳng lên qua đầu đến khi duỗi gần hết tay.",
			"Hạ tạ từ từ về vị trí ban đầu.",
		},
		Sets:           4,
		Reps:           12,
		RestSeconds:    60,
		CaloriesBurned: 100,
		Difficulty:     "Intermediate",
	},
}

func GetExercisesHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		sendJSONResponse(w, http.StatusMethodNotAllowed, map[string]string{"status": "error", "message": "Method not allowed"})
		return
	}

	bodyPart := strings.ToLower(r.URL.Query().Get("bodyPart"))
	equipment := strings.ToLower(r.URL.Query().Get("equipment"))
	location := strings.ToLower(r.URL.Query().Get("location"))
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
		if equipment != "" && equipment != "all" {
			filter["equipment"] = equipment
		}
		if location != "" && location != "all" {
			filter["location"] = location
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
			matchEquipment := equipment == "" || equipment == "all" || strings.Contains(strings.ToLower(ex.Equipment), equipment)
			matchLocation := location == "" || location == "all" || ex.Location == "both" || ex.Location == location
			matchQuery := query == "" || strings.Contains(strings.ToLower(ex.Name), query) || strings.Contains(strings.ToLower(ex.NameVi), query) || strings.Contains(strings.ToLower(ex.Equipment), query)

			if matchBodyPart && matchEquipment && matchLocation && matchQuery {
				results = append(results, ex)
			}
		}
	}

	sendJSONResponse(w, http.StatusOK, map[string]interface{}{
		"status": "success",
		"count":  len(results),
		"data":   results,
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

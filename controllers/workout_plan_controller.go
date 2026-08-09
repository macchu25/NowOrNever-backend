package controllers

import (
	"context"
	"encoding/json"
	"net/http"
	"sort"
	"strings"
	"time"

	"backend-go/config"
	"backend-go/middleware"
	"backend-go/models"
	"backend-go/utils"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func containsFold(values []string, candidate string) bool {
	for _, value := range values {
		if strings.Contains(strings.ToLower(candidate), strings.ToLower(value)) || strings.Contains(strings.ToLower(value), strings.ToLower(candidate)) {
			return true
		}
	}
	return false
}

func exerciseAllowed(ex models.Exercise, survey models.Survey) bool {
	if containsFold(survey.DislikedExercises, ex.ExerciseID) || containsFold(survey.DislikedExercises, ex.Name) {
		return false
	}
	locationOK := survey.WorkoutLocation == "hybrid" || ex.Location == "both" || ex.Location == survey.WorkoutLocation
	if !locationOK {
		return false
	}
	equipment := strings.ToLower(survey.Equipment)
	equipmentOK := strings.Contains(strings.ToLower(ex.Equipment), "body weight")
	switch equipment {
	case "dumbbells":
		equipmentOK = equipmentOK || strings.Contains(strings.ToLower(ex.Equipment), "dumbbell")
	case "full_gym":
		equipmentOK = true
	}
	if !equipmentOK {
		return false
	}
	limitations := strings.Join(survey.Limitations, " ")
	if strings.Contains(limitations, "knee") && (strings.Contains(strings.ToLower(ex.Name), "squat") || strings.Contains(strings.ToLower(ex.Name), "lunge")) {
		return false
	}
	if strings.Contains(limitations, "shoulder") && (ex.BodyPart == "shoulders" || strings.Contains(strings.ToLower(ex.Name), "press")) {
		return false
	}
	if survey.ExperienceLevel == "beginner" && strings.EqualFold(ex.Difficulty, "advanced") {
		return false
	}
	return true
}

func buildWorkoutPlan(survey models.Survey) models.WorkoutPlan {
	if survey.DaysPerWeek < 2 || survey.DaysPerWeek > 6 {
		survey.DaysPerWeek = 3
	}
	if survey.SessionMinutes < 20 || survey.SessionMinutes > 120 {
		survey.SessionMinutes = 45
	}
	if survey.ExperienceLevel == "" {
		survey.ExperienceLevel = "beginner"
	}

	eligible := make([]models.Exercise, 0, len(defaultExercises))
	for _, ex := range defaultExercises {
		if exerciseAllowed(ex, survey) {
			eligible = append(eligible, ex)
		}
	}
	sort.SliceStable(eligible, func(i, j int) bool {
		left := containsFold(survey.Preferences, eligible[i].BodyPart) || containsFold(survey.Preferences, eligible[i].Name)
		right := containsFold(survey.Preferences, eligible[j].BodyPart) || containsFold(survey.Preferences, eligible[j].Name)
		return left && !right
	})

	focuses := []string{"Toàn thân", "Thân trên", "Thân dưới", "Sức bền", "Toàn thân", "Kỹ thuật & phục hồi"}
	exercisesPerDay := survey.SessionMinutes / 9
	if exercisesPerDay < 3 {
		exercisesPerDay = 3
	}
	if exercisesPerDay > 6 {
		exercisesPerDay = 6
	}
	targetRPE := 6
	sets := 3
	if survey.ExperienceLevel == "intermediate" {
		targetRPE, sets = 7, 4
	}
	if survey.ExperienceLevel == "advanced" {
		targetRPE, sets = 8, 4
	}

	schedule := make([]models.WorkoutDay, 0, survey.DaysPerWeek)
	for day := 0; day < survey.DaysPerWeek; day++ {
		items := make([]models.PlanExercise, 0, exercisesPerDay)
		for offset := 0; offset < len(eligible) && len(items) < exercisesPerDay; offset++ {
			ex := eligible[(day*exercisesPerDay+offset)%len(eligible)]
			items = append(items, models.PlanExercise{
				ExerciseID: ex.ExerciseID, Name: ex.Name, NameVi: ex.NameVi, BodyPart: ex.BodyPart,
				Equipment: ex.Equipment, Sets: sets, Reps: ex.Reps, RestSeconds: ex.RestSeconds,
				TargetRPE: targetRPE, Reason: "Phù hợp mục tiêu, thiết bị và khả năng hiện tại của bạn.",
			})
		}
		schedule = append(schedule, models.WorkoutDay{Day: day + 1, Focus: focuses[day], Exercises: items})
	}
	warnings := []string{"Khởi động 5–8 phút trước mỗi buổi và dừng lại nếu đau nhói, chóng mặt hoặc khó thở bất thường."}
	if len(survey.Limitations) > 0 {
		warnings = append(warnings, "Kế hoạch đã loại một số động tác theo hạn chế bạn khai báo; hãy hỏi chuyên gia y tế nếu triệu chứng kéo dài.")
	}
	return models.WorkoutPlan{
		Goal: survey.Goal, Level: survey.ExperienceLevel, DaysPerWeek: survey.DaysPerWeek,
		SessionMinutes: survey.SessionMinutes, Schedule: schedule, Warnings: warnings,
		ProgressionRule: "Khi hoàn thành đủ số lần ở RPE thấp hơn mục tiêu trong 2 buổi liên tiếp, tăng 1–2 lần lặp hoặc 2–5% mức tạ.",
		GeneratedBy:     "rules-v1", CreatedAt: time.Now(),
	}
}

func GenerateWorkoutPlanHandler(w http.ResponseWriter, r *http.Request) {
	var request models.GenerateWorkoutPlanRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		sendJSONResponse(w, http.StatusBadRequest, map[string]string{"status": "error", "message": "Invalid plan request"})
		return
	}
	plan := buildWorkoutPlan(request.Survey)
	if len(plan.Schedule) == 0 || len(plan.Schedule[0].Exercises) == 0 {
		sendJSONResponse(w, http.StatusUnprocessableEntity, map[string]string{"status": "error", "message": "Không tìm thấy bài tập phù hợp. Hãy kiểm tra thiết bị và hạn chế vận động."})
		return
	}
	claims, _ := r.Context().Value(middleware.UserContextKey).(*utils.JWTClaims)
	if claims != nil {
		if userID, err := primitive.ObjectIDFromHex(claims.UserID); err == nil {
			plan.ID, plan.UserID = primitive.NewObjectID(), userID
			if config.MongoConnected && config.DB != nil {
				ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
				defer cancel()
				_, _ = config.DB.Collection("workout_plans").InsertOne(ctx, plan)
			}
		}
	}
	sendJSONResponse(w, http.StatusCreated, map[string]interface{}{"status": "success", "plan": plan})
}

func GetCurrentWorkoutPlanHandler(w http.ResponseWriter, r *http.Request) {
	claims, _ := r.Context().Value(middleware.UserContextKey).(*utils.JWTClaims)
	if claims == nil || !config.MongoConnected || config.DB == nil {
		sendJSONResponse(w, http.StatusNotFound, map[string]string{"status": "error", "message": "Chưa có kế hoạch đã lưu"})
		return
	}
	userID, err := primitive.ObjectIDFromHex(claims.UserID)
	if err != nil {
		sendJSONResponse(w, http.StatusBadRequest, map[string]string{"status": "error", "message": "Invalid user ID"})
		return
	}
	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()
	var plan models.WorkoutPlan
	err = config.DB.Collection("workout_plans").FindOne(ctx, bson.M{"userId": userID}, options.FindOne().SetSort(bson.D{{Key: "createdAt", Value: -1}})).Decode(&plan)
	if err != nil {
		sendJSONResponse(w, http.StatusNotFound, map[string]string{"status": "error", "message": "Chưa có kế hoạch đã lưu"})
		return
	}
	sendJSONResponse(w, http.StatusOK, map[string]interface{}{"status": "success", "plan": plan})
}

package careers

import (
	"encoding/json"
	"hired-valley-backend/config"
	"hired-valley-backend/controllers/authentication"
	"hired-valley-backend/models/career"
	"hired-valley-backend/models/users"
	"hired-valley-backend/services"
	"net/http"
	"os"
)

type CareerPlanRequest struct {
	ShortTermGoals string `json:"short_term_goals"`
	LongTermGoals  string `json:"long_term_goals"`
}

func GenerateCareerPlanHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Проверка токена
	claims, err := authentication.ValidateToken(r)
	if err != nil {
		http.Error(w, "Unauthorized: "+err.Error(), http.StatusUnauthorized)
		return
	}

	var req CareerPlanRequest
	err = json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	apiKey := os.Getenv("AIML_API_KEY")
	if apiKey == "" {
		http.Error(w, "API key is missing", http.StatusInternalServerError)
		return
	}

	// Генерация карьерного плана
	plan, err := services.GenerateCareerPlan(apiKey, req.ShortTermGoals, req.LongTermGoals)
	if err != nil {
		http.Error(w, "Failed to generate career plan: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Сохранение плана в базе данных
	careerPlan := career.PlanCareer{
		UserID:         claims.UserID,
		ShortTermGoals: req.ShortTermGoals,
		LongTermGoals:  req.LongTermGoals,
		Steps:          plan,
	}
	if err := config.DB.Create(&careerPlan).Error; err != nil {
		http.Error(w, "Failed to save plan", http.StatusInternalServerError)
		return
	}

	// Поиск менторов по навыкам и интересам
	var mentors []users.User
	err = config.DB.
		Preload("Skills").
		Preload("Interests").
		Where("role = ?", "mentor").
		Where("visibility = ?", "public").
		Joins("LEFT JOIN user_skills ON user_skills.user_id = users.id").
		Joins("LEFT JOIN user_interests ON user_interests.user_id = users.id").
		Where("user_skills.name ILIKE ? OR user_interests.name ILIKE ?",
			"%"+req.ShortTermGoals+"%", "%"+req.LongTermGoals+"%").
		Group("users.id").
		Find(&mentors).Error

	if err != nil {
		http.Error(w, "Failed to retrieve mentors: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Проверка на наличие найденных менторов
	if len(mentors) == 0 {
		http.Error(w, "No mentors found matching the criteria", http.StatusNotFound)
		return
	}

	// Формирование ответа
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"plan_id": careerPlan.ID,
		"steps":   plan,
		"mentors": mentors,
	})
}

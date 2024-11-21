package careers

import (
	"encoding/json"
	"hired-valley-backend/config"
	"hired-valley-backend/controllers/authentication"
	"hired-valley-backend/models/career"
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

	// Получаем запрос
	var req CareerPlanRequest
	err = json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Генерация prompt
	prompt := services.GenerateCareerPrompt(req.ShortTermGoals, req.LongTermGoals)

	// Получение API-ключа
	apiKey := os.Getenv("AIML_API_KEY")
	if apiKey == "" {
		http.Error(w, "API key is missing", http.StatusInternalServerError)
		return
	}

	// Вызов AI/ML API
	response, err := services.GenerateRecommendations(apiKey, prompt)
	if err != nil {
		http.Error(w, "Failed to generate plan: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Сохраняем план в базе данных
	careerPlan := career.PlanCareer{
		UserID:         claims.UserID,
		ShortTermGoals: req.ShortTermGoals,
		LongTermGoals:  req.LongTermGoals,
		Steps:          response,
	}
	if err := config.DB.Create(&careerPlan).Error; err != nil {
		http.Error(w, "Failed to save plan", http.StatusInternalServerError)
		return
	}

	// Возвращаем успешный ответ
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"plan_id": careerPlan.ID,
		"steps":   response,
	})
}

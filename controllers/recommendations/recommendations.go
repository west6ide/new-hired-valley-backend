package recommendations

import (
	"encoding/json"
	"hired-valley-backend/config"
	"hired-valley-backend/controllers/authentication"
	"hired-valley-backend/models/users"
	"hired-valley-backend/services"
	"net/http"
	"os"
)

func GetRecommendationsHandler(w http.ResponseWriter, r *http.Request) {
	// Извлекаем данные токена
	claims, err := authentication.ValidateToken(r)
	if err != nil {
		http.Error(w, "Unauthorized: "+err.Error(), http.StatusUnauthorized)
		return
	}

	// Получаем userID из токена
	userID := claims.UserID

	// Ищем пользователя в базе данных
	var user users.User
	if err := config.DB.Preload("Skills").Preload("Interests").First(&user, userID).Error; err != nil {
		http.Error(w, "User not found", http.StatusNotFound)
		return
	}

	// Генерируем рекомендации через OpenAI API
	apiKey := os.Getenv("OPENAI_API_KEY")
	recs, err := services.GenerateAIRecommendationsForUser(config.DB, apiKey, userID)
	if err != nil {
		http.Error(w, "Failed to generate recommendations: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Возвращаем рекомендации
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"recommendations": recs,
	})
}

package recommendations

import (
	"encoding/json"
	"fmt"
	"gorm.io/gorm"
	"hired-valley-backend/controllers/authentication"
	"hired-valley-backend/models/users"
	"hired-valley-backend/services"
	"net/http"
	"os"
)

func GenerateRecommendationsHandler(w http.ResponseWriter, r *http.Request, db *gorm.DB) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Проверяем токен и извлекаем данные
	claims, err := authentication.ValidateToken(r)
	if err != nil {
		http.Error(w, "Unauthorized: "+err.Error(), http.StatusUnauthorized)
		return
	}

	// Получаем пользователя из базы данных
	var user users.User
	err = db.Preload("Skills").Preload("Interests").First(&user, claims.UserID).Error
	if err != nil {
		http.Error(w, "User not found", http.StatusNotFound)
		return
	}

	// Генерация prompt
	prompt := services.GeneratePrompt(user.Skills, user.Interests)

	// Получение API-ключа
	apiKey := os.Getenv("AIML_API_KEY")
	if apiKey == "" {
		http.Error(w, "API key is missing in server configuration", http.StatusInternalServerError)
		return
	}

	// Вызов AI/ML API
	response, err := services.GenerateRecommendations(apiKey, prompt)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"error": fmt.Sprintf("Failed to generate recommendations: %v", err),
		})
		return
	}

	// Возвращаем рекомендации
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"recommendations": response,
	})
}

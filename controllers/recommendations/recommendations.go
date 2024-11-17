package recommendations

import (
	"encoding/json"
	"hired-valley-backend/config"
	"hired-valley-backend/controllers/authentication"
	"hired-valley-backend/models/users"
	"hired-valley-backend/services"
	"net/http"
)

// GetRecommendations обрабатывает запросы для получения рекомендаций
func GetRecommendations(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		return
	}

	// Проверяем токен пользователя
	claims, err := authentication.ValidateToken(r)
	if err != nil {
		http.Error(w, "Unauthorized: "+err.Error(), http.StatusUnauthorized)
		return
	}

	// Извлекаем данные пользователя из базы по UserID из токена
	var user users.User
	if err := config.DB.Preload("Skills").Preload("Interests").First(&user, claims.UserID).Error; err != nil {
		http.Error(w, "User not found", http.StatusNotFound)
		return
	}

	// Генерируем рекомендации
	recommendations, err := services.GenerateRecommendations(user)
	if err != nil {
		http.Error(w, "Failed to get recommendations: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(recommendations)
}

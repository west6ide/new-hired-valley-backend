package controllers

import (
	"encoding/json"
	"fmt"
	"gorm.io/gorm/clause"
	"hired-valley-backend/config"
	"hired-valley-backend/models/users"
	"net/http"
)

// Структура для ответа, как будто выданного ИИ
type RecommendationResponse struct {
	ContentID   uint   `json:"content_id"`
	Title       string `json:"title"`
	Description string `json:"description"`
	AIComment   string `json:"ai_comment"`
}
type Content struct {
	ID          uint     `gorm:"primaryKey" json:"id"`
	Title       string   `json:"title"`
	Description string   `json:"description"`
	Category    string   `json:"category"`
	Tags        []string `gorm:"type:text[]" json:"tags"` // Массив тегов
}

// Обработчик для рекомендаций, совместимый с http.HandlerFunc
func GetRecommendationsHandler(w http.ResponseWriter, r *http.Request) {
	userID := r.URL.Query().Get("userID")
	var user users.User
	var recommendations []RecommendationResponse

	// Получаем профиль пользователя
	if err := config.DB.First(&user, userID).Error; err != nil {
		http.Error(w, "User not found", http.StatusNotFound)
		return
	}

	// Находим контент, релевантный пользователю
	var content []Content
	config.DB.Where("category IN ?", user.Interests).
		Or(clause.Expr{SQL: "tags && ?", Vars: []interface{}{user.Skills}}).
		Find(&content)

	// Формируем ответ с "ИИ-комментарием"
	for _, item := range content {
		recommendations = append(recommendations, RecommendationResponse{
			ContentID:   item.ID,
			Title:       item.Title,
			Description: item.Description,
			AIComment:   fmt.Sprintf("Recommended for you based on your interest in %s.", item.Category),
		})
	}

	// Кодирование данных в JSON и отправка ответа
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(map[string]interface{}{"recommendations": recommendations}); err != nil {
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
	}
}

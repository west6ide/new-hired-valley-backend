// controllers/recommendations.go

package controllers

import (
	"encoding/json"
	"hired-valley-backend/config"
	"hired-valley-backend/models"
	"log"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
)

// SaveRecommendations сохраняет рекомендации для пользователя в базе данных
func SaveRecommendations(userID uint, recommendations []models.RecommendedContent) error {
	for _, recommendation := range recommendations {
		rec := models.RecommendedContent{
			UserID:     userID,
			Content:    recommendation.Content,
			Similarity: recommendation.Similarity,
		}
		if err := config.DB.Create(&rec).Error; err != nil {
			log.Printf("Ошибка сохранения рекомендации: %v", err)
			return err
		}
	}
	return nil
}

// GetUserRecommendationsHandler возвращает последние рекомендации для конкретного пользователя
func GetUserRecommendationsHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	userID, err := strconv.Atoi(vars["userID"])
	if err != nil {
		http.Error(w, "Неверный формат userID", http.StatusBadRequest)
		return
	}

	var recommendations []models.RecommendedContent
	if err := config.DB.Where("user_id = ?", userID).Order("created_at desc").Limit(10).Find(&recommendations).Error; err != nil {
		http.Error(w, "Ошибка при получении рекомендаций", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(recommendations)
}

// GetRecommendationsHandlerWithFilters возвращает рекомендации с возможностью фильтрации и сортировки
func GetRecommendationsHandlerWithFilters(w http.ResponseWriter, r *http.Request) {
	userID := r.URL.Query().Get("user_id")
	limit := r.URL.Query().Get("limit")
	sort := r.URL.Query().Get("sort")

	var recommendations []models.RecommendedContent
	query := config.DB

	// Фильтр по userID
	if userID != "" {
		query = query.Where("user_id = ?", userID)
	}

	// Лимит результатов
	limitInt := 10
	if limit != "" {
		limitInt, _ = strconv.Atoi(limit)
	}
	query = query.Limit(limitInt)

	// Сортировка
	if sort == "asc" {
		query = query.Order("created_at asc")
	} else {
		query = query.Order("created_at desc")
	}

	if err := query.Find(&recommendations).Error; err != nil {
		http.Error(w, "Ошибка при получении рекомендаций", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(recommendations)
}

// TempCreateRecommendationsHandler временный обработчик для создания рекомендаций (только для тестирования)
func TempCreateRecommendationsHandler(w http.ResponseWriter, r *http.Request) {
	var recommendations []models.RecommendedContent
	if err := json.NewDecoder(r.Body).Decode(&recommendations); err != nil {
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}

	userID := recommendations[0].UserID // Предполагается, что все рекомендации для одного пользователя
	if err := SaveRecommendations(userID, recommendations); err != nil {
		http.Error(w, "Failed to save recommendations", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]string{"message": "Recommendations saved successfully"})
}

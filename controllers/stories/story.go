package stories

import (
	"encoding/json"
	"errors"
	"gorm.io/gorm"
	"hired-valley-backend/config"
	"hired-valley-backend/controllers/authentication"
	"hired-valley-backend/models/story"
	"net/http"
	"strconv"
	"time"
)

// Создание истории
func CreateStory(w http.ResponseWriter, r *http.Request) {
	claims, err := authentication.ValidateToken(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}

	if r.Method != http.MethodPost {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		return
	}

	var newStory story.Story
	if err := json.NewDecoder(r.Body).Decode(&newStory); err != nil {
		http.Error(w, "Bad request", http.StatusBadRequest)
		return
	}

	if newStory.ContentURL == "" {
		http.Error(w, "ContentURL is required", http.StatusBadRequest)
		return
	}

	newStory.UserID = claims.UserID
	newStory.CreatedAt = time.Now().UTC()
	newStory.ExpireAt = newStory.CreatedAt.Add(24 * time.Hour)

	if result := config.DB.Create(&newStory); result.Error != nil {
		http.Error(w, "Failed to create story", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(newStory)
}

// Получение всех историй пользователя
func GetUserStories(w http.ResponseWriter, r *http.Request) {
	claims, err := authentication.ValidateToken(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}

	var stories []story.Story
	config.DB.Where("user_id = ? AND expire_at > ? AND is_archived = ?", claims.UserID, time.Now().UTC(), false).Find(&stories)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(stories)
}

// Просмотр одной истории
func ViewStory(w http.ResponseWriter, r *http.Request) {
	claims, err := authentication.ValidateToken(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}

	storyIDStr := r.URL.Query().Get("id")
	storyID, err := strconv.Atoi(storyIDStr)
	if err != nil {
		http.Error(w, "Invalid story ID", http.StatusBadRequest)
		return
	}

	var currentStory story.Story
	if result := config.DB.First(&currentStory, storyID); errors.Is(result.Error, gorm.ErrRecordNotFound) {
		http.Error(w, "Story not found", http.StatusNotFound)
		return
	}

	if currentStory.UserID != claims.UserID && currentStory.Privacy != "public" {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}

	// Атомарное обновление просмотров
	if err := config.DB.Model(&currentStory).UpdateColumn("views", gorm.Expr("views + ?", 1)).Error; err != nil {
		http.Error(w, "Failed to update views", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(currentStory)
}

// Архивирование истории
func ArchiveStory(w http.ResponseWriter, r *http.Request) {
	claims, err := authentication.ValidateToken(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}

	storyIDStr := r.URL.Query().Get("id")
	storyID, err := strconv.Atoi(storyIDStr)
	if err != nil {
		http.Error(w, "Invalid story ID", http.StatusBadRequest)
		return
	}

	var currentStory story.Story
	if result := config.DB.First(&currentStory, storyID); errors.Is(result.Error, gorm.ErrRecordNotFound) {
		http.Error(w, "Story not found", http.StatusNotFound)
		return
	}

	if currentStory.UserID != claims.UserID {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}

	currentStory.IsArchived = true
	if err := config.DB.Save(&currentStory).Error; err != nil {
		http.Error(w, "Failed to archive story", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(currentStory)
}

// UpdateStory - обновление информации об истории
func UpdateStory(w http.ResponseWriter, r *http.Request) {
	claims, err := authentication.ValidateToken(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}

	if r.Method != http.MethodPut {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		return
	}

	storyIDStr := r.URL.Query().Get("id")
	storyID, err := strconv.Atoi(storyIDStr)
	if err != nil {
		http.Error(w, "Invalid story ID", http.StatusBadRequest)
		return
	}

	var currentStory story.Story
	if result := config.DB.First(&currentStory, storyID); errors.Is(result.Error, gorm.ErrRecordNotFound) {
		http.Error(w, "Story not found", http.StatusNotFound)
		return
	}

	if currentStory.UserID != claims.UserID {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}

	if time.Now().After(currentStory.ExpireAt) {
		http.Error(w, "Cannot update an expired story", http.StatusBadRequest)
		return
	}

	var updatedStory story.Story
	if err := json.NewDecoder(r.Body).Decode(&updatedStory); err != nil {
		http.Error(w, "Invalid input", http.StatusBadRequest)
		return
	}

	// Обновляем только разрешенные поля
	if updatedStory.ContentURL != "" {
		currentStory.ContentURL = updatedStory.ContentURL
	}
	if updatedStory.Privacy != "" {
		currentStory.Privacy = updatedStory.Privacy
	}

	if err := config.DB.Save(&currentStory).Error; err != nil {
		http.Error(w, "Failed to update story", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(currentStory)
}

// Удаление истории
func DeleteStory(w http.ResponseWriter, r *http.Request) {
	claims, err := authentication.ValidateToken(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}

	storyIDStr := r.URL.Query().Get("id")
	storyID, err := strconv.Atoi(storyIDStr)
	if err != nil {
		http.Error(w, "Invalid story ID", http.StatusBadRequest)
		return
	}

	var currentStory story.Story
	if result := config.DB.First(&currentStory, storyID); errors.Is(result.Error, gorm.ErrRecordNotFound) {
		http.Error(w, "Story not found", http.StatusNotFound)
		return
	}

	if currentStory.UserID != claims.UserID {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}

	if err := config.DB.Delete(&currentStory).Error; err != nil {
		http.Error(w, "Failed to delete story", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

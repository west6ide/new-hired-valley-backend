package stories

import (
	"encoding/json"
	"gorm.io/gorm"
	"hired-valley-backend/controllers/authentication" // Импортируем authentication для проверки токенов
	"hired-valley-backend/models/story"
	"net/http"
	"strconv"
	"time"
)

func CreateStory(w http.ResponseWriter, r *http.Request, db *gorm.DB) {
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

	// Устанавливаем userID из токена и время истечения истории
	newStory.UserID = claims.UserID
	newStory.CreatedAt = time.Now()
	newStory.ExpireAt = newStory.CreatedAt.Add(24 * time.Hour)

	if result := db.Create(&newStory); result.Error != nil {
		http.Error(w, "Failed to create stories", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(newStory)
}

func GetUserStories(w http.ResponseWriter, r *http.Request, db *gorm.DB) {
	claims, err := authentication.ValidateToken(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}

	if r.Method != http.MethodGet {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		return
	}

	var stories []story.Story
	db.Where("user_id = ? AND expire_at > ? AND is_archived = ?", claims.UserID, time.Now(), false).Find(&stories)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(stories)
}

func ViewStory(w http.ResponseWriter, r *http.Request, db *gorm.DB) {
	claims, err := authentication.ValidateToken(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}

	if r.Method != http.MethodPost {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		return
	}

	storyIDStr := r.URL.Query().Get("id")
	storyID, err := strconv.Atoi(storyIDStr)
	if err != nil {
		http.Error(w, "Invalid stories ID", http.StatusBadRequest)
		return
	}

	var currentStory story.Story
	if result := db.First(&currentStory, storyID); result.Error != nil {
		http.Error(w, "Story not found", http.StatusNotFound)
		return
	}

	// Проверка, что пользователь просматривает свою историю или она публичная
	if currentStory.UserID != claims.UserID && currentStory.Privacy != "public" {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}

	currentStory.Views += 1
	db.Save(&currentStory)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(currentStory)
}

func ArchiveStory(w http.ResponseWriter, r *http.Request, db *gorm.DB) {
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
		http.Error(w, "Invalid stories ID", http.StatusBadRequest)
		return
	}

	var currentStory story.Story
	if result := db.First(&currentStory, storyID); result.Error != nil {
		http.Error(w, "Story not found", http.StatusNotFound)
		return
	}

	// Проверка, что пользователь архивирует свою собственную историю
	if currentStory.UserID != claims.UserID {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}

	currentStory.IsArchived = true
	db.Save(&currentStory)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(currentStory)
}

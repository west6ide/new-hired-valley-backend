package story

import (
	"encoding/json"
	"errors"
	"gorm.io/gorm"
	"hired-valley-backend/models/story"
	"net/http"
	"strconv"
	"strings"
	"time"
)

func authenticate(r *http.Request) (uint, error) {
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		return 0, errors.New("authorization header missing")
	}

	// Проверка формата заголовка Authorization
	parts := strings.Split(authHeader, " ")
	if len(parts) != 2 || parts[0] != "Bearer" {
		return 0, errors.New("invalid authorization header format")
	}

	token := parts[1]
	// Проверка и извлечение userID из токена
	// Здесь предполагается использование какой-либо библиотеки для проверки JWT токена.
	// Для примера, считаем, что token == "valid-token" и привязываем userID = 1
	if token == "valid-token" {
		return 1, nil
	}

	return 0, errors.New("invalid token")
}

func CreateStory(w http.ResponseWriter, r *http.Request, db *gorm.DB) {
	userID, err := authenticate(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}

	if r.Method != http.MethodPost {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		return
	}

	var story story.Story
	if err := json.NewDecoder(r.Body).Decode(&story); err != nil {
		http.Error(w, "Bad request", http.StatusBadRequest)
		return
	}

	// Устанавливаем userID и время истечения
	story.UserID = userID
	story.CreatedAt = time.Now()
	story.ExpireAt = story.CreatedAt.Add(24 * time.Hour)

	if result := db.Create(&story); result.Error != nil {
		http.Error(w, "Failed to create story", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(story)
}

func GetUserStories(w http.ResponseWriter, r *http.Request, db *gorm.DB) {
	userID, err := authenticate(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}

	if r.Method != http.MethodGet {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		return
	}

	var stories []story.Story
	db.Where("user_id = ? AND expire_at > ? AND is_archived = ?", userID, time.Now(), false).Find(&stories)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(stories)
}

func ViewStory(w http.ResponseWriter, r *http.Request, db *gorm.DB) {
	userID, err := authenticate(r)
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
		http.Error(w, "Invalid story ID", http.StatusBadRequest)
		return
	}

	var story story.Story
	if result := db.First(&story, storyID); result.Error != nil {
		http.Error(w, "Story not found", http.StatusNotFound)
		return
	}

	// Проверка, что пользователь просматривает свою историю или она публичная
	if story.UserID != userID && story.Privacy != "public" {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}

	story.Views += 1
	db.Save(&story)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(story)
}

func ArchiveStory(w http.ResponseWriter, r *http.Request, db *gorm.DB) {
	userID, err := authenticate(r)
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

	var story story.Story
	if result := db.First(&story, storyID); result.Error != nil {
		http.Error(w, "Story not found", http.StatusNotFound)
		return
	}

	// Проверка, что пользователь архивирует свою собственную историю
	if story.UserID != userID {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}

	story.IsArchived = true
	db.Save(&story)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(story)
}

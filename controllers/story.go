package controllers

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"
	"time"

	"hired-valley-backend/config"
	"hired-valley-backend/models/users"
)

func CreateStory(w http.ResponseWriter, r *http.Request) {
	var story users.Story

	if err := json.NewDecoder(r.Body).Decode(&story); err != nil {
		http.Error(w, "Invalid input", http.StatusBadRequest)
		return
	}

	// Логируем UserID для отладки
	log.Printf("Creating story for UserID: %d", story.UserID)

	var user users.User
	if err := config.DB.First(&user, story.UserID).Error; err != nil {
		log.Printf("User not found with ID: %d", story.UserID)
		http.Error(w, "User not found", http.StatusNotFound)
		return
	}

	story.CreatedAt = time.Now()
	story.ExpiresAt = time.Now().Add(24 * time.Hour)

	if err := config.DB.Create(&story).Error; err != nil {
		log.Printf("Ошибка при сохранении сториса в базе данных: %v", err)
		http.Error(w, "Error saving story", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(story)
}

// GetActiveStories - обработчик для получения всех активных историй
func GetActiveStories(w http.ResponseWriter, r *http.Request) {
	var stories []users.Story
	config.DB.Where("expires_at > ? AND is_archived = ?", time.Now(), false).Find(&stories)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(stories)
}

// ArchiveStory - обработчик для архивации истории
func ArchiveStory(w http.ResponseWriter, r *http.Request) {
	idStr := r.URL.Query().Get("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "Invalid story ID", http.StatusBadRequest)
		return
	}

	var story users.Story
	if err := config.DB.First(&story, id).Error; err != nil {
		http.Error(w, "Story not found", http.StatusNotFound)
		return
	}

	story.IsArchived = true
	if err := config.DB.Save(&story).Error; err != nil {
		log.Printf("Ошибка при архивации истории: %v", err)
		http.Error(w, "Failed to archive story", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(story)
}

// GetArchivedStories - обработчик для получения архивированных историй
func GetArchivedStories(w http.ResponseWriter, r *http.Request) {
	userIDStr := r.URL.Query().Get("user_id")
	userID, err := strconv.Atoi(userIDStr)
	if err != nil {
		http.Error(w, "Invalid user ID", http.StatusBadRequest)
		return
	}

	var stories []users.Story
	config.DB.Where("user_id = ? AND is_archived = ?", userID, true).Find(&stories)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(stories)
}

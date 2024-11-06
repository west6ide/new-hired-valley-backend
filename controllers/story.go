package controllers

import (
	"encoding/json"
	"hired-valley-backend/config"
	"hired-valley-backend/models/users"
	"log"
	"net/http"
	"strconv"
	"time"
)

// Создание истории
func CreateStory(w http.ResponseWriter, r *http.Request) {
	var story users.Story

	if err := json.NewDecoder(r.Body).Decode(&story); err != nil {
		http.Error(w, "Invalid input", http.StatusBadRequest)
		return
	}

	// Проверяем, существует ли пользователь с данным UserID
	var user users.User
	if err := config.DB.First(&user, story.UserID).Error; err != nil {
		log.Printf("User not found with ID: %d", story.UserID)
		http.Error(w, "User not found", http.StatusNotFound)
		return
	}

	story.CreatedAt = time.Now()
	story.ExpiresAt = time.Now().Add(24 * time.Hour) // История истечет через 24 часа

	if err := config.DB.Create(&story).Error; err != nil {
		log.Printf("Error saving story: %v", err)
		http.Error(w, "Error saving story", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(story)
}

// Получение всех активных (не истекших) публичных историй
func GetActiveStories(w http.ResponseWriter, r *http.Request) {
	var stories []users.Story
	config.DB.Where("expires_at > ? AND is_archived = ? AND is_public = ?", time.Now(), false, true).Find(&stories)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(stories)
}

// Получение всех активных историй пользователя
func GetUserStories(w http.ResponseWriter, r *http.Request) {
	userIDStr := r.URL.Query().Get("user_id")
	userID, err := strconv.Atoi(userIDStr)
	if err != nil {
		http.Error(w, "Invalid user ID", http.StatusBadRequest)
		return
	}

	var stories []users.Story
	config.DB.Where("user_id = ? AND expires_at > ?", userID, time.Now()).Find(&stories)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(stories)
}

// Архивирование истории
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
		log.Printf("Error archiving story: %v", err)
		http.Error(w, "Failed to archive story", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(story)
}

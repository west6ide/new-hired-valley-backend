package controllers

import (
	"encoding/json"
	"github.com/gorilla/mux"
	"hired-valley-backend/config"
	"hired-valley-backend/models"
	"net/http"
	"strconv"
	"time"
)

// CreateStory - обработчик для создания истории
func CreateStory(w http.ResponseWriter, r *http.Request) {
	var story models.Story
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&story); err != nil {
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	story.CreatedAt = time.Now()
	story.ExpiresAt = story.CreatedAt.Add(24 * time.Hour)

	if err := config.DB.Create(&story).Error; err != nil {
		http.Error(w, "Failed to create story", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(story)
}

// GetActiveStories - обработчик для получения всех активных историй
func GetActiveStories(w http.ResponseWriter, r *http.Request) {
	var stories []models.Story
	config.DB.Where("expires_at > ? AND is_archived = ?", time.Now(), false).Find(&stories)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(stories)
}

// ArchiveStory - обработчик для архивации истории
func ArchiveStory(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		http.Error(w, "Invalid story ID", http.StatusBadRequest)
		return
	}

	var story models.Story
	if err := config.DB.First(&story, id).Error; err != nil {
		http.Error(w, "Story not found", http.StatusNotFound)
		return
	}

	story.IsArchived = true
	config.DB.Save(&story)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(story)
}

// GetArchivedStories - обработчик для получения архивированных историй
func GetArchivedStories(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	userID, err := strconv.Atoi(vars["user_id"])
	if err != nil {
		http.Error(w, "Invalid user ID", http.StatusBadRequest)
		return
	}

	var stories []models.Story
	config.DB.Where("user_id = ? AND is_archived = ?", userID, true).Find(&stories)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(stories)
}

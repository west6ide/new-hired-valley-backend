package story

import (
	"encoding/json"
	"gorm.io/gorm"
	"hired-valley-backend/models/story"
	"net/http"
	"strconv"
	"time"
)

func CreateStory(w http.ResponseWriter, r *http.Request, db *gorm.DB) {
	if r.Method != http.MethodPost {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		return
	}

	var story story.Story
	if err := json.NewDecoder(r.Body).Decode(&story); err != nil {
		http.Error(w, "Bad request", http.StatusBadRequest)
		return
	}

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
	if r.Method != http.MethodGet {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		return
	}

	userIDStr := r.URL.Query().Get("user_id")
	userID, err := strconv.Atoi(userIDStr)
	if err != nil {
		http.Error(w, "Invalid user ID", http.StatusBadRequest)
		return
	}

	var stories []story.Story
	db.Where("user_id = ? AND expire_at > ? AND is_archived = ?", userID, time.Now(), false).Find(&stories)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(stories)
}

func ViewStory(w http.ResponseWriter, r *http.Request, db *gorm.DB) {
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

	story.Views += 1
	db.Save(&story)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(story)
}

func ArchiveStory(w http.ResponseWriter, r *http.Request, db *gorm.DB) {
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

	story.IsArchived = true
	db.Save(&story)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(story)
}

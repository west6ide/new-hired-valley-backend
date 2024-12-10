package stories

import (
	"encoding/json"
	"errors"
	"fmt"
	"gorm.io/gorm"
	"hired-valley-backend/controllers/authentication"
	"hired-valley-backend/models/story"
	"net/http"
	"strconv"
	"time"
)

// AddReaction - добавление реакции
func AddReaction(w http.ResponseWriter, r *http.Request, db *gorm.DB) {
	// Проверяем Google OAuth токен
	googleUser, err := authentication.ValidateGoogleToken(r)
	if err != nil {
		http.Error(w, "Unauthorized: "+err.Error(), http.StatusUnauthorized)
		return
	}

	if r.Method != http.MethodPost {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		return
	}

	var reaction story.Reaction
	if err := json.NewDecoder(r.Body).Decode(&reaction); err != nil {
		http.Error(w, "Bad request", http.StatusBadRequest)
		return
	}

	reaction.UserID = googleUser.UserID
	reaction.CreatedAt = time.Now().UTC()

	if result := db.Create(&reaction); result.Error != nil {
		http.Error(w, "Failed to add reaction", http.StatusInternalServerError)
		return
	}

	// Уведомление владельца истории
	var storyOwner story.Story
	if err := db.First(&storyOwner, reaction.StoryID).Error; err == nil {
		sendNotification(db, storyOwner.UserID, fmt.Sprintf("User %d reacted to your story", googleUser.UserID))
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(reaction)
}

// GetReactions - получение всех реакций для истории
func GetReactions(w http.ResponseWriter, r *http.Request, db *gorm.DB) {
	storyIDStr := r.URL.Query().Get("story_id")
	storyID, err := strconv.Atoi(storyIDStr)
	if err != nil {
		http.Error(w, "Invalid story ID", http.StatusBadRequest)
		return
	}

	var reactions []story.Reaction
	db.Where("story_id = ?", storyID).Find(&reactions)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(reactions)
}

// UpdateReaction - обновление реакции
func UpdateReaction(w http.ResponseWriter, r *http.Request, db *gorm.DB) {
	// Проверяем Google OAuth токен
	googleUser, err := authentication.ValidateGoogleToken(r)
	if err != nil {
		http.Error(w, "Unauthorized: "+err.Error(), http.StatusUnauthorized)
		return
	}

	if r.Method != http.MethodPut {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		return
	}

	reactionIDStr := r.URL.Query().Get("id")
	reactionID, err := strconv.Atoi(reactionIDStr)
	if err != nil {
		http.Error(w, "Invalid reaction ID", http.StatusBadRequest)
		return
	}

	var existingReaction story.Reaction
	if result := db.First(&existingReaction, reactionID); errors.Is(result.Error, gorm.ErrRecordNotFound) {
		http.Error(w, "Reaction not found", http.StatusNotFound)
		return
	}

	if existingReaction.UserID != googleUser.UserID {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}

	var updatedReaction story.Reaction
	if err := json.NewDecoder(r.Body).Decode(&updatedReaction); err != nil {
		http.Error(w, "Bad request", http.StatusBadRequest)
		return
	}

	existingReaction.Emoji = updatedReaction.Emoji
	if err := db.Save(&existingReaction).Error; err != nil {
		http.Error(w, "Failed to update reaction", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(existingReaction)
}

// DeleteReaction - удаление реакции
func DeleteReaction(w http.ResponseWriter, r *http.Request, db *gorm.DB) {
	// Проверяем Google OAuth токен
	googleUser, err := authentication.ValidateGoogleToken(r)
	if err != nil {
		http.Error(w, "Unauthorized: "+err.Error(), http.StatusUnauthorized)
		return
	}

	if r.Method != http.MethodDelete {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		return
	}

	reactionIDStr := r.URL.Query().Get("id")
	reactionID, err := strconv.Atoi(reactionIDStr)
	if err != nil {
		http.Error(w, "Invalid reaction ID", http.StatusBadRequest)
		return
	}

	var reaction story.Reaction
	if result := db.First(&reaction, reactionID); errors.Is(result.Error, gorm.ErrRecordNotFound) {
		http.Error(w, "Reaction not found", http.StatusNotFound)
		return
	}

	if reaction.UserID != googleUser.UserID {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}

	if err := db.Delete(&reaction).Error; err != nil {
		http.Error(w, "Failed to delete reaction", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

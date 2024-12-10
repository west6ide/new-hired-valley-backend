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

// CreateComment - добавление нового комментария
func CreateComment(w http.ResponseWriter, r *http.Request, db *gorm.DB) {
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

	var comment story.Comment
	if err := json.NewDecoder(r.Body).Decode(&comment); err != nil {
		http.Error(w, "Bad request", http.StatusBadRequest)
		return
	}

	// Устанавливаем UserID текущего пользователя
	comment.UserID = googleUser.UserID
	comment.CreatedAt = time.Now().UTC()

	if result := db.Create(&comment); result.Error != nil {
		http.Error(w, "Failed to add comment", http.StatusInternalServerError)
		return
	}

	// Уведомление владельца истории
	var storyOwner story.Story
	if err := db.First(&storyOwner, comment.StoryID).Error; err == nil {
		sendNotification(db, storyOwner.UserID, fmt.Sprintf("User %d commented on your story", googleUser.UserID))
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(comment)
}

// GetComments - получение всех комментариев для истории
func GetComments(w http.ResponseWriter, r *http.Request, db *gorm.DB) {
	// Проверяем Google OAuth токен
	_, err := authentication.ValidateGoogleToken(r)
	if err != nil {
		http.Error(w, "Unauthorized: "+err.Error(), http.StatusUnauthorized)
		return
	}

	storyIDStr := r.URL.Query().Get("story_id")
	storyID, err := strconv.Atoi(storyIDStr)
	if err != nil {
		http.Error(w, "Invalid story ID", http.StatusBadRequest)
		return
	}

	var comments []story.Comment
	db.Where("story_id = ?", storyID).Find(&comments)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(comments)
}

// UpdateComment - обновление комментария
func UpdateComment(w http.ResponseWriter, r *http.Request, db *gorm.DB) {
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

	commentIDStr := r.URL.Query().Get("id")
	commentID, err := strconv.Atoi(commentIDStr)
	if err != nil {
		http.Error(w, "Invalid comment ID", http.StatusBadRequest)
		return
	}

	var existingComment story.Comment
	if result := db.First(&existingComment, commentID); errors.Is(result.Error, gorm.ErrRecordNotFound) {
		http.Error(w, "Comment not found", http.StatusNotFound)
		return
	}

	if existingComment.UserID != googleUser.UserID {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}

	var updatedComment story.Comment
	if err := json.NewDecoder(r.Body).Decode(&updatedComment); err != nil {
		http.Error(w, "Bad request", http.StatusBadRequest)
		return
	}

	existingComment.Content = updatedComment.Content
	if err := db.Save(&existingComment).Error; err != nil {
		http.Error(w, "Failed to update comment", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(existingComment)
}

// DeleteComment - удаление комментария
func DeleteComment(w http.ResponseWriter, r *http.Request, db *gorm.DB) {
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

	commentIDStr := r.URL.Query().Get("id")
	commentID, err := strconv.Atoi(commentIDStr)
	if err != nil {
		http.Error(w, "Invalid comment ID", http.StatusBadRequest)
		return
	}

	var comment story.Comment
	if result := db.First(&comment, commentID); errors.Is(result.Error, gorm.ErrRecordNotFound) {
		http.Error(w, "Comment not found", http.StatusNotFound)
		return
	}

	if comment.UserID != googleUser.UserID {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}

	if err := db.Delete(&comment).Error; err != nil {
		http.Error(w, "Failed to delete comment", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

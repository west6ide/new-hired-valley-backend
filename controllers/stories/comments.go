package stories

import (
	"encoding/json"
	"errors"
	"fmt"
	"gorm.io/gorm"
	"hired-valley-backend/models/story"
	"net/http"
	"strconv"
	"time"
)

// CreateComment - добавление нового комментария
func CreateComment(w http.ResponseWriter, r *http.Request, db *gorm.DB) {
	if r.Method != http.MethodPost {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		return
	}

	var comment story.Comment
	if err := json.NewDecoder(r.Body).Decode(&comment); err != nil {
		http.Error(w, "Bad request", http.StatusBadRequest)
		return
	}

	// Проверка существования истории, к которой добавляется комментарий
	var storyOwner story.Story
	if err := db.First(&storyOwner, comment.StoryID).Error; err != nil {
		http.Error(w, "Story not found", http.StatusNotFound)
		return
	}

	// Устанавливаем текущее время создания комментария
	comment.CreatedAt = time.Now().UTC()

	// Сохраняем комментарий в базе данных
	if result := db.Create(&comment); result.Error != nil {
		http.Error(w, "Failed to add comment", http.StatusInternalServerError)
		return
	}

	// Уведомление владельца истории
	sendNotification(db, storyOwner.UserID, fmt.Sprintf("User %d commented on your story", comment.UserID))

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(comment)
}

func GetComments(w http.ResponseWriter, r *http.Request, db *gorm.DB) {
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

	var updatedComment story.Comment
	if err := json.NewDecoder(r.Body).Decode(&updatedComment); err != nil {
		http.Error(w, "Bad request", http.StatusBadRequest)
		return
	}

	// Проверка, что пользователь редактирует свой комментарий
	if existingComment.UserID != updatedComment.UserID {
		http.Error(w, "Forbidden", http.StatusForbidden)
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

	var existingComment story.Comment
	if result := db.First(&existingComment, commentID); errors.Is(result.Error, gorm.ErrRecordNotFound) {
		http.Error(w, "Comment not found", http.StatusNotFound)
		return
	}

	var requestBody struct {
		UserID uint `json:"user_id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&requestBody); err != nil {
		http.Error(w, "Bad request", http.StatusBadRequest)
		return
	}

	// Проверка, что пользователь удаляет свой комментарий
	if existingComment.UserID != requestBody.UserID {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}

	if err := db.Delete(&existingComment).Error; err != nil {
		http.Error(w, "Failed to delete comment", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

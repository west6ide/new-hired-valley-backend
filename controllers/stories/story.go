package stories

import (
	"encoding/json"
	"errors"
	"fmt"
	"gorm.io/gorm"
	"hired-valley-backend/config"
	"hired-valley-backend/controllers/authentication" // Импортируем authentication для проверки токенов
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

// Просмотр истории
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

// Helper to send notifications
func sendNotification(db *gorm.DB, userID uint, message string) {
	notification := story.Notification{
		UserID:    userID,
		Message:   message,
		CreatedAt: time.Now().UTC(),
	}
	db.Create(&notification)
}

// Add reaction to a story
func AddReaction(w http.ResponseWriter, r *http.Request, db *gorm.DB) {
	claims, err := authentication.ValidateToken(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnauthorized)
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

	reaction.UserID = claims.UserID
	reaction.CreatedAt = time.Now().UTC()

	if result := db.Create(&reaction); result.Error != nil {
		http.Error(w, "Failed to add reaction", http.StatusInternalServerError)
		return
	}

	// Notify story owner
	var storyOwner story.Story
	if err := db.First(&storyOwner, reaction.StoryID).Error; err == nil {
		sendNotification(db, storyOwner.UserID, fmt.Sprintf("User %d reacted to your story", claims.UserID))
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(reaction)
}

// Add comment to a story
func AddComment(w http.ResponseWriter, r *http.Request, db *gorm.DB) {
	claims, err := authentication.ValidateToken(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnauthorized)
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

	comment.UserID = claims.UserID
	comment.CreatedAt = time.Now().UTC()

	if result := db.Create(&comment); result.Error != nil {
		http.Error(w, "Failed to add comment", http.StatusInternalServerError)
		return
	}

	// Notify story owner
	var storyOwner story.Story
	if err := db.First(&storyOwner, comment.StoryID).Error; err == nil {
		sendNotification(db, storyOwner.UserID, fmt.Sprintf("User %d commented on your story", claims.UserID))
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(comment)
}

// Get comments for a story
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

// Get notifications for the user
func GetNotifications(w http.ResponseWriter, r *http.Request, db *gorm.DB) {
	claims, err := authentication.ValidateToken(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}

	var notifications []story.Notification
	db.Where("user_id = ?", claims.UserID).Find(&notifications)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(notifications)
}

// Mark notification as read
func MarkNotificationAsRead(w http.ResponseWriter, r *http.Request, db *gorm.DB) {
	claims, err := authentication.ValidateToken(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}

	notificationIDStr := r.URL.Query().Get("id")
	notificationID, err := strconv.Atoi(notificationIDStr)
	if err != nil {
		http.Error(w, "Invalid notification ID", http.StatusBadRequest)
		return
	}

	db.Model(&story.Notification{}).Where("id = ? AND user_id = ?", notificationID, claims.UserID).
		Update("is_read", true)

	w.WriteHeader(http.StatusOK)
}

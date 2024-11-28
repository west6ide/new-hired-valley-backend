package stories

import (
	"encoding/json"
	"gorm.io/gorm"
	"hired-valley-backend/controllers/authentication"
	"hired-valley-backend/models/story"
	"net/http"
	"strconv"
	"time"
)

// Helper to send notifications
func sendNotification(db *gorm.DB, userID uint, message string) {
	notification := story.Notification{
		UserID:    userID,
		Message:   message,
		IsRead:    false,
		CreatedAt: time.Now().UTC(),
	}
	db.Create(&notification)
}

// GetNotifications - получение всех уведомлений для пользователя
func GetNotifications(w http.ResponseWriter, r *http.Request, db *gorm.DB) {
	claims, err := authentication.ValidateToken(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}

	var notifications []story.Notification
	db.Where("user_id = ?", claims.UserID).Order("created_at DESC").Find(&notifications)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(notifications)
}

// MarkNotificationAsRead - отметить уведомление как прочитанное
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

// DeleteNotification - удаление уведомления
func DeleteNotification(w http.ResponseWriter, r *http.Request, db *gorm.DB) {
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

	if result := db.Where("id = ? AND user_id = ?", notificationID, claims.UserID).Delete(&story.Notification{}); result.Error != nil {
		http.Error(w, "Failed to delete notification", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

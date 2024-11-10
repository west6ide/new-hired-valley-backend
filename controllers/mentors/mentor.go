package mentors

import (
	"encoding/json"
	"hired-valley-backend/config"
	"hired-valley-backend/controllers/authentication"
	"hired-valley-backend/models/users"
	"net/http"
	"strings"
)

// Helper function to parse URL path parameters
func getPathParam(path string, param string) string {
	parts := strings.Split(path, "/")
	for i, part := range parts {
		if part == param && i+1 < len(parts) {
			return parts[i+1]
		}
	}
	return ""
}

// CreateMentorProfile creates a mentor profile for users with the "mentor" role
func CreateMentorProfile(w http.ResponseWriter, r *http.Request) {
	claims, err := authentication.ValidateToken(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}

	if claims.Role != "mentor" {
		http.Error(w, "Only mentors can create profiles", http.StatusForbidden)
		return
	}

	var profile users.MentorProfile
	if err := json.NewDecoder(r.Body).Decode(&profile); err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	profile.UserID = claims.UserID
	if err := config.DB.Create(&profile).Error; err != nil {
		http.Error(w, "Failed to create profile", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(profile)
}

// GetMentorProfile retrieves a mentor profile by ID
func GetMentorProfile(w http.ResponseWriter, r *http.Request) {
	claims, err := authentication.ValidateToken(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}

	var profile users.MentorProfile
	if err := config.DB.Where("user_id = ?", claims.UserID).First(&profile).Error; err != nil {
		http.Error(w, "Profile not found", http.StatusNotFound)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(profile)
}

// UpdateMentorProfile updates an existing mentor profile
func UpdateMentorProfile(w http.ResponseWriter, r *http.Request) {
	claims, err := authentication.ValidateToken(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}

	var profile users.MentorProfile
	if err := config.DB.Where("user_id = ?", claims.UserID).First(&profile).Error; err != nil {
		http.Error(w, "Profile not found", http.StatusNotFound)
		return
	}

	if err := json.NewDecoder(r.Body).Decode(&profile); err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	if err := config.DB.Save(&profile).Error; err != nil {
		http.Error(w, "Failed to update profile", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(profile)
}

// DeleteMentorProfile deletes a mentor profile by ID
func DeleteMentorProfile(w http.ResponseWriter, r *http.Request) {
	claims, err := authentication.ValidateToken(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}

	if err := config.DB.Where("user_id = ?", claims.UserID).Delete(&users.MentorProfile{}).Error; err != nil {
		http.Error(w, "Failed to delete profile", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"message": "Profile deleted"})
}

// CRUD Handlers for AvailableTime
// AddAvailableTime создает новый интервал доступности для ментора
func AddAvailableTime(w http.ResponseWriter, r *http.Request) {
	claims, err := authentication.ValidateToken(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}

	if claims.Role != "mentor" {
		http.Error(w, "Only mentors can add available times", http.StatusForbidden)
		return
	}

	var availableTime users.AvailableTime
	if err := json.NewDecoder(r.Body).Decode(&availableTime); err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	// Устанавливаем связь с ментором
	availableTime.MentorID = claims.UserID
	if availableTime.StartTime.After(availableTime.EndTime) {
		http.Error(w, "Start time must be before end time", http.StatusBadRequest)
		return
	}

	if err := config.DB.Create(&availableTime).Error; err != nil {
		http.Error(w, "Failed to create available time", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(availableTime)
}

// GetAvailableTimes получает список доступных интервалов для ментора
func GetAvailableTimes(w http.ResponseWriter, r *http.Request) {
	claims, err := authentication.ValidateToken(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}

	var schedule []users.AvailableTime
	if err := config.DB.Where("mentor_id = ?", claims.UserID).Find(&schedule).Error; err != nil {
		http.Error(w, "Failed to retrieve schedule", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(schedule)
}

// UpdateAvailableTime обновляет существующий интервал доступности
func UpdateAvailableTime(w http.ResponseWriter, r *http.Request) {
	claims, err := authentication.ValidateToken(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}

	var availableTime users.AvailableTime
	id := getPathParam(r.URL.Path, "available-times")

	if err := config.DB.First(&availableTime, id).Error; err != nil {
		http.Error(w, "Available time not found", http.StatusNotFound)
		return
	}

	// Проверка, что интервал принадлежит текущему ментору
	if availableTime.MentorID != claims.UserID {
		http.Error(w, "You are not authorized to update this available time", http.StatusForbidden)
		return
	}

	if err := json.NewDecoder(r.Body).Decode(&availableTime); err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	if availableTime.StartTime.After(availableTime.EndTime) {
		http.Error(w, "Start time must be before end time", http.StatusBadRequest)
		return
	}

	if err := config.DB.Save(&availableTime).Error; err != nil {
		http.Error(w, "Failed to update available time", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(availableTime)
}

// DeleteAvailableTime удаляет существующий интервал доступности
func DeleteAvailableTime(w http.ResponseWriter, r *http.Request) {
	claims, err := authentication.ValidateToken(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}

	var availableTime users.AvailableTime
	id := getPathParam(r.URL.Path, "available-times")

	if err := config.DB.First(&availableTime, id).Error; err != nil {
		http.Error(w, "Available time not found", http.StatusNotFound)
		return
	}

	// Проверка, что интервал принадлежит текущему ментору
	if availableTime.MentorID != claims.UserID {
		http.Error(w, "You are not authorized to delete this available time", http.StatusForbidden)
		return
	}

	if err := config.DB.Delete(&availableTime).Error; err != nil {
		http.Error(w, "Failed to delete available time", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"message": "Available time deleted"})
}

package mentors

import (
	"encoding/json"
	"gorm.io/gorm"
	"hired-valley-backend/config"
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

// CRUD Handlers for MentorProfile

func CreateMentorProfile(w http.ResponseWriter, r *http.Request) {
	userID := r.Header.Get("userID")
	if userID == "" {
		http.Error(w, "User not authenticated", http.StatusUnauthorized)
		return
	}

	var user users.User
	if err := config.DB.First(&user, userID).Error; err != nil {
		http.Error(w, "User not found", http.StatusNotFound)
		return
	}

	if user.Role != "mentor" {
		http.Error(w, "Only mentors can create mentor profiles", http.StatusForbidden)
		return
	}

	var profile users.MentorProfile
	if err := json.NewDecoder(r.Body).Decode(&profile); err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	profile.UserID = user.ID
	if err := config.DB.Create(&profile).Error; err != nil {
		http.Error(w, "Failed to create mentor profile", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(profile)
}

func GetMentorProfile(w http.ResponseWriter, r *http.Request) {
	id := getPathParam(r.URL.Path, "mentors")
	var profile users.MentorProfile

	if err := config.DB.Preload("Skills").Preload("Schedule").First(&profile, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			http.Error(w, "Profile not found", http.StatusNotFound)
			return
		}
		http.Error(w, "Failed to retrieve profile", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(profile)
}

func UpdateMentorProfile(w http.ResponseWriter, r *http.Request) {
	id := getPathParam(r.URL.Path, "mentors")
	var profile users.MentorProfile

	if err := config.DB.First(&profile, id).Error; err != nil {
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

func DeleteMentorProfile(w http.ResponseWriter, r *http.Request) {
	id := getPathParam(r.URL.Path, "mentors")
	var profile users.MentorProfile

	if err := config.DB.First(&profile, id).Error; err != nil {
		http.Error(w, "Profile not found", http.StatusNotFound)
		return
	}

	if err := config.DB.Delete(&profile).Error; err != nil {
		http.Error(w, "Failed to delete profile", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"message": "Profile deleted"})
}

// CRUD Handlers for AvailableTime

func AddAvailableTime(w http.ResponseWriter, r *http.Request) {
	var availableTime users.AvailableTime

	if err := json.NewDecoder(r.Body).Decode(&availableTime); err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

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

func GetAvailableTimes(w http.ResponseWriter, r *http.Request) {
	mentorID := getPathParam(r.URL.Path, "mentors")
	var schedule []users.AvailableTime

	if err := config.DB.Where("mentor_id = ?", mentorID).Find(&schedule).Error; err != nil {
		http.Error(w, "Failed to retrieve schedule", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(schedule)
}

func UpdateAvailableTime(w http.ResponseWriter, r *http.Request) {
	id := getPathParam(r.URL.Path, "available-times")
	var availableTime users.AvailableTime

	if err := config.DB.First(&availableTime, id).Error; err != nil {
		http.Error(w, "Available time not found", http.StatusNotFound)
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

func DeleteAvailableTime(w http.ResponseWriter, r *http.Request) {
	id := getPathParam(r.URL.Path, "available-times")
	var availableTime users.AvailableTime

	if err := config.DB.First(&availableTime, id).Error; err != nil {
		http.Error(w, "Available time not found", http.StatusNotFound)
		return
	}

	if err := config.DB.Delete(&availableTime).Error; err != nil {
		http.Error(w, "Failed to delete available time", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"message": "Available time deleted"})
}

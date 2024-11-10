package mentor

import (
	"encoding/json"
	"hired-valley-backend/config"
	"hired-valley-backend/controllers/authentication"
	"hired-valley-backend/models/users"
	"net/http"
	"strconv"
	"strings"
)

// Проверка, что пользователь является наставником
func isMentor(role string) bool {
	return role == "mentor"
}

// CreateMentorProfile — создание профиля наставника
func CreateMentorProfile(w http.ResponseWriter, r *http.Request) {
	claims, err := authentication.ValidateToken(r)
	if err != nil || !isMentor(claims.Role) {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	var profile users.MentorProfile
	if err := json.NewDecoder(r.Body).Decode(&profile); err != nil {
		http.Error(w, "Invalid input", http.StatusBadRequest)
		return
	}

	profile.UserID = claims.UserID
	if err := config.DB.Create(&profile).Error; err != nil {
		http.Error(w, "Error creating mentor profile", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(profile)
}

// GetMentorProfile — получение профиля наставника
func GetMentorProfile(w http.ResponseWriter, r *http.Request) {
	claims, err := authentication.ValidateToken(r)
	if err != nil || !isMentor(claims.Role) {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	var profile users.MentorProfile
	if err := config.DB.Where("user_id = ?", claims.UserID).First(&profile).Error; err != nil {
		http.Error(w, "Profile not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(profile)
}

// UpdateMentorProfile — обновление профиля наставника
func UpdateMentorProfile(w http.ResponseWriter, r *http.Request) {
	claims, err := authentication.ValidateToken(r)
	if err != nil || !isMentor(claims.Role) {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	var profile users.MentorProfile
	if err := config.DB.Where("user_id = ?", claims.UserID).First(&profile).Error; err != nil {
		http.Error(w, "Profile not found", http.StatusNotFound)
		return
	}

	if err := json.NewDecoder(r.Body).Decode(&profile); err != nil {
		http.Error(w, "Invalid input", http.StatusBadRequest)
		return
	}

	if err := config.DB.Save(&profile).Error; err != nil {
		http.Error(w, "Error updating profile", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(profile)
}

// DeleteMentorProfile — удаление профиля наставника
func DeleteMentorProfile(w http.ResponseWriter, r *http.Request) {
	claims, err := authentication.ValidateToken(r)
	if err != nil || !isMentor(claims.Role) {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	if err := config.DB.Where("user_id = ?", claims.UserID).Delete(&users.MentorProfile{}).Error; err != nil {
		http.Error(w, "Error deleting profile", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"message": "Profile deleted successfully"})
}

// CreateAvailableTime — создание доступного времени для наставника
func CreateAvailableTime(w http.ResponseWriter, r *http.Request) {
	claims, err := authentication.ValidateToken(r)
	if err != nil || !isMentor(claims.Role) {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	var timeSlot users.AvailableTime
	if err := json.NewDecoder(r.Body).Decode(&timeSlot); err != nil {
		http.Error(w, "Invalid input", http.StatusBadRequest)
		return
	}

	timeSlot.MentorID = claims.UserID
	if err := config.DB.Create(&timeSlot).Error; err != nil {
		http.Error(w, "Error creating available time", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(timeSlot)
}

// GetAvailableTimes — получение всех доступных временных интервалов для наставника
func GetAvailableTimes(w http.ResponseWriter, r *http.Request) {
	claims, err := authentication.ValidateToken(r)
	if err != nil || !isMentor(claims.Role) {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	var timeSlots []users.AvailableTime
	if err := config.DB.Where("mentor_id = ?", claims.UserID).Find(&timeSlots).Error; err != nil {
		http.Error(w, "No available times found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(timeSlots)
}

// UpdateAvailableTime — обновление временного интервала
func UpdateAvailableTime(w http.ResponseWriter, r *http.Request) {
	claims, err := authentication.ValidateToken(r)
	if err != nil || !isMentor(claims.Role) {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	idStr := strings.TrimPrefix(r.URL.Path, "/mentor/available_times/")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "Invalid time slot ID", http.StatusBadRequest)
		return
	}

	var timeSlot users.AvailableTime
	if err := config.DB.Where("id = ? AND mentor_id = ?", id, claims.UserID).First(&timeSlot).Error; err != nil {
		http.Error(w, "Time slot not found", http.StatusNotFound)
		return
	}

	if err := json.NewDecoder(r.Body).Decode(&timeSlot); err != nil {
		http.Error(w, "Invalid input", http.StatusBadRequest)
		return
	}

	if err := config.DB.Save(&timeSlot).Error; err != nil {
		http.Error(w, "Error updating time slot", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(timeSlot)
}

// DeleteAvailableTime — удаление временного интервала
func DeleteAvailableTime(w http.ResponseWriter, r *http.Request) {
	claims, err := authentication.ValidateToken(r)
	if err != nil || !isMentor(claims.Role) {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	idStr := strings.TrimPrefix(r.URL.Path, "/mentor/available_times/")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "Invalid time slot ID", http.StatusBadRequest)
		return
	}

	if err := config.DB.Where("id = ? AND mentor_id = ?", id, claims.UserID).Delete(&users.AvailableTime{}).Error; err != nil {
		http.Error(w, "Error deleting time slot", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"message": "Time slot deleted successfully"})
}

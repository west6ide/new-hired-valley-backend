package mentors

import (
	"encoding/json"
	"fmt"
	"hired-valley-backend/config"
	"hired-valley-backend/controllers/authentication"
	"hired-valley-backend/models/users"
	"net/http"
	"strconv"
)

func MentorsHandler(w http.ResponseWriter, r *http.Request) {
	user, err := authentication.ValidateToken(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}

	if r.Method == http.MethodPost {
		var mentorProfile users.MentorProfile
		if err := json.NewDecoder(r.Body).Decode(&mentorProfile); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		if user.Role != "mentor" {
			http.Error(w, "User is not authorized to create a mentor profile", http.StatusForbidden)
			return
		}

		mentorProfile.UserID = user.UserID
		config.DB.Create(&mentorProfile)
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(mentorProfile)
	} else if r.Method == http.MethodGet {
		var mentors []users.MentorProfile
		query := config.DB.Preload("User")
		skills := r.URL.Query().Get("skills")
		if skills != "" {
			query = query.Where("skills LIKE ?", fmt.Sprintf("%%%s%%", skills))
		}
		priceRange := r.URL.Query().Get("price_range")
		if priceRange != "" {
			price, _ := strconv.ParseFloat(priceRange, 64)
			query = query.Where("price_per_hour <= ?", price)
		}
		query.Find(&mentors)
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(mentors)
	} else {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func BookSlotHandler(w http.ResponseWriter, r *http.Request) {
	user, err := authentication.ValidateToken(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}

	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var slot users.Slot
	if err := json.NewDecoder(r.Body).Decode(&slot); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if user.Role != "mentor" {
		http.Error(w, "User is not authorized to book slots", http.StatusForbidden)
		return
	}

	var existingSlot users.Slot
	config.DB.First(&existingSlot, "mentor_id = ? AND start_time = ? AND end_time = ?", slot.MentorID, slot.StartTime, slot.EndTime)

	if existingSlot.ID != 0 && existingSlot.IsBooked {
		http.Error(w, "Slot is already booked", http.StatusConflict)
		return
	}

	slot.IsBooked = true
	config.DB.Create(&slot)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(slot)
}

func NotificationsHandler(w http.ResponseWriter, r *http.Request) {
	_, err := authentication.ValidateToken(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}

	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var notification users.NotificationMentor
	if err := json.NewDecoder(r.Body).Decode(&notification); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	config.DB.Create(&notification)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(notification)
}

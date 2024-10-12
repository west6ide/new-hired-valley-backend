// search.go

package authentication

import (
	"encoding/json"
	"hired-valley-backend/config"
	"hired-valley-backend/models"
	"net/http"
)

// SearchUsers: Поиск пользователей по навыкам, интересам и другим параметрам
func SearchUsers(w http.ResponseWriter, r *http.Request) {
	skill := r.URL.Query().Get("skill")
	interest := r.URL.Query().Get("interest")
	position := r.URL.Query().Get("position")

	var users []models.User
	query := config.DB

	if skill != "" {
		query = query.Joins("JOIN user_skills ON users.id = user_skills.user_id").Joins("JOIN skills ON skills.id = user_skills.skill_id").Where("skills.name = ?", skill)
	}
	if interest != "" {
		query = query.Joins("JOIN user_interests ON users.id = user_interests.user_id").Joins("JOIN interests ON interests.id = user_interests.interest_id").Where("interests.name = ?", interest)
	}
	if position != "" {
		query = query.Where("position = ?", position)
	}

	if err := query.Find(&users).Error; err != nil {
		http.Error(w, "Error fetching users", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(users)
}

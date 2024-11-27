// search.go

package authentication

import (
	"encoding/json"
	"hired-valley-backend/config"
	"hired-valley-backend/models/users"
	"net/http"
)

// SearchUsers: Поиск пользователей по навыкам, интересам, геолокации и профессиональным данным
func SearchUsers(w http.ResponseWriter, r *http.Request) {
	// Извлечение параметров запроса
	skill := r.URL.Query().Get("skill")
	interest := r.URL.Query().Get("interest")
	position := r.URL.Query().Get("position")
	city := r.URL.Query().Get("city")
	company := r.URL.Query().Get("company")
	industry := r.URL.Query().Get("industry")

	var users []users.User
	query := config.DB

	// Фильтр по навыкам
	if skill != "" {
		query = query.Joins("JOIN user_skills ON users.id = user_skills.user_id").
			Joins("JOIN skills ON skills.id = user_skills.skill_id").
			Where("skills.name = ?", skill)
	}

	// Фильтр по интересам
	if interest != "" {
		query = query.Joins("JOIN user_interests ON users.id = user_interests.user_id").
			Joins("JOIN interests ON interests.id = user_interests.interest_id").
			Where("interests.name = ?", interest)
	}

	// Фильтр по позиции
	if position != "" {
		query = query.Where("position = ?", position)
	}

	// Фильтр по городу
	if city != "" {
		query = query.Where("city = ?", city)
	}

	// Фильтр по компании
	if company != "" {
		query = query.Where("company = ?", company)
	}

	// Фильтр по индустрии
	if industry != "" {
		query = query.Where("industry = ?", industry)
	}

	// Выполнение запроса
	if err := query.Find(&users).Error; err != nil {
		http.Error(w, "Error fetching users", http.StatusInternalServerError)
		return
	}

	// Возвращаем список пользователей
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(users)
}

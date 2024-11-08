package mentors

import (
	"encoding/json"
	"github.com/dgrijalva/jwt-go"
	"hired-valley-backend/config"
	"hired-valley-backend/controllers/authentication"
	"hired-valley-backend/models/users"
	"net/http"
	"strings"
)

// Создание доступности
func CreateAvailability(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Получение токена из заголовка Authorization
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		http.Error(w, "Authorization header required", http.StatusUnauthorized)
		return
	}

	// Убираем "Bearer " из начала заголовка
	tokenString := strings.TrimPrefix(authHeader, "Bearer ")
	claims := &authentication.Claims{}

	token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
		return authentication.JwtKey, nil
	})

	if err != nil || !token.Valid {
		http.Error(w, "Invalid token", http.StatusUnauthorized)
		return
	}

	// Проверка, что пользователь имеет роль наставника
	var user users.User
	if err := config.DB.First(&user, claims.UserID).Error; err != nil {
		http.Error(w, "User not found", http.StatusUnauthorized)
		return
	}

	if user.Role != "mentor" {
		http.Error(w, "Only mentors can create availability", http.StatusForbidden)
		return
	}

	// Создание доступности
	var availability users.Availability
	if err := json.NewDecoder(r.Body).Decode(&availability); err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	// Присваиваем профильному полю MentorID значение из авторизованного пользователя
	availability.MentorID = user.ID

	if err := config.DB.Create(&availability).Error; err != nil {
		http.Error(w, "Could not create availability", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(availability)
}

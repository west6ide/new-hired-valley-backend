package authentication

import (
	"encoding/json"
	"github.com/dgrijalva/jwt-go"
	"hired-valley-backend/config"
	"hired-valley-backend/models"
	"net/http"
	"strings"
)

// Обработчик для обновления профиля
func UpdateProfile(w http.ResponseWriter, r *http.Request) {
	// Извлекаем токен из заголовка Authorization
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		http.Error(w, "Authorization header required", http.StatusUnauthorized)
		return
	}

	// Убираем "Bearer " из начала заголовка
	tokenString := strings.TrimPrefix(authHeader, "Bearer ")

	// Проверяем и декодируем токен
	claims := &Claims{}
	token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
		return jwtKey, nil
	})

	if err != nil || !token.Valid {
		http.Error(w, "Invalid token", http.StatusUnauthorized)
		return
	}

	// Извлекаем email пользователя из токена
	email := claims.Email

	// Поиск пользователя по email в базе данных
	var user models.User
	if err := config.DB.Where("email = ?", email).First(&user).Error; err != nil {
		http.Error(w, "User not found", http.StatusNotFound)
		return
	}

	// Декодируем тело запроса и обновляем профиль пользователя
	var updatedProfile models.User
	if err := json.NewDecoder(r.Body).Decode(&updatedProfile); err != nil {
		http.Error(w, "Invalid input", http.StatusBadRequest)
		return
	}

	// Обновляем данные профиля
	user.Position = updatedProfile.Position
	user.City = updatedProfile.City
	user.Income = updatedProfile.Income
	user.Skills = updatedProfile.Skills
	user.Interests = updatedProfile.Interests
	user.Visibility = updatedProfile.Visibility

	// Сохраняем изменения в базе данных
	if err := config.DB.Save(&user).Error; err != nil {
		http.Error(w, "Error updating profile", http.StatusInternalServerError)
		return
	}

	// Возвращаем обновлённые данные профиля
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(user)
}

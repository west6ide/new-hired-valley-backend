package authentication

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/dgrijalva/jwt-go"
	"hired-valley-backend/config"
	"hired-valley-backend/models"
)

func UpdateProfile(w http.ResponseWriter, r *http.Request) {
	// Извлечение токена из заголовка Authorization
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		http.Error(w, "Authorization header required", http.StatusUnauthorized)
		return
	}

	// Убираем "Bearer " из заголовка
	tokenString := strings.TrimPrefix(authHeader, "Bearer ")

	// Проверка и декодирование токена
	claims := &Claims{}
	token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
		return jwtKey, nil
	})
	if err != nil || !token.Valid {
		http.Error(w, "Invalid token", http.StatusUnauthorized)
		return
	}

	// Извлечение email пользователя из токена
	email := claims.Email

	// Поиск пользователя по email в базе данных
	var user models.User
	if err := config.DB.Where("email = ?", email).First(&user).Error; err != nil {
		http.Error(w, "User not found", http.StatusNotFound)
		return
	}

	// Декодирование тела запроса для обновления профиля
	var updatedProfile models.User
	if err := json.NewDecoder(r.Body).Decode(&updatedProfile); err != nil {
		http.Error(w, "Invalid input", http.StatusBadRequest)
		return
	}

	// Обновление полей пользователя
	user.Position = updatedProfile.Position
	user.City = updatedProfile.City
	user.Income = updatedProfile.Income
	user.Skills = updatedProfile.Skills
	user.Interests = updatedProfile.Interests
	user.Visibility = updatedProfile.Visibility

	// Сохранение изменений в базе данных
	if err := config.DB.Save(&user).Error; err != nil {
		http.Error(w, "Error updating profile", http.StatusInternalServerError)
		return
	}

	// Возвращаем обновлённый профиль
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(user)
}

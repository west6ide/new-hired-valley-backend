package authentication

import (
	"encoding/json"
	"fmt"
	"github.com/dgrijalva/jwt-go"
	"golang.org/x/crypto/bcrypt"
	"hired-valley-backend/config"
	"hired-valley-backend/models/authenticationUsers"
	"net/http"
	"strings"
)

// ChangePassword: смена пароля пользователя
func ChangePassword(w http.ResponseWriter, r *http.Request) {
	// Шаг 1: Получение текущего пользователя через токен
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		http.Error(w, "Authorization header required", http.StatusUnauthorized)
		return
	}

	// Извлекаем токен из заголовка
	tokenString := strings.TrimPrefix(authHeader, "Bearer ")
	claims := &Claims{}

	// Парсим токен и проверяем его валидность
	token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
		return jwtKey, nil
	})

	if err != nil || !token.Valid {
		http.Error(w, "Invalid token", http.StatusUnauthorized)
		return
	}

	// Поиск пользователя по email из токена
	var user authenticationUsers.User
	if err := config.DB.Where("email = ? AND provider = ?", claims.Email, "local").First(&user).Error; err != nil {
		http.Error(w, "User not found", http.StatusUnauthorized)
		return
	}

	// Отладочный вывод для проверки правильности хэша
	fmt.Println("Hashed password from DB:", user.Password)

	// Шаг 2: Получение старого и нового пароля из запроса
	var passwordChangeRequest struct {
		CurrentPassword string `json:"current_password"`
		NewPassword     string `json:"new_password"`
	}
	if err := json.NewDecoder(r.Body).Decode(&passwordChangeRequest); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Отладочный вывод для проверки введенного пароля
	fmt.Println("Provided current password:", passwordChangeRequest.CurrentPassword)

	// Шаг 3: Проверка текущего пароля
	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(passwordChangeRequest.CurrentPassword))
	if err != nil {
		fmt.Println("Error comparing passwords:", err)
		http.Error(w, "Current password is incorrect", http.StatusUnauthorized)
		return
	}

	// Шаг 4: Хэширование нового пароля
	hashedNewPassword, err := bcrypt.GenerateFromPassword([]byte(passwordChangeRequest.NewPassword), bcrypt.DefaultCost)
	if err != nil {
		http.Error(w, "Error hashing new password", http.StatusInternalServerError)
		return
	}

	// Шаг 5: Обновление пароля в базе данных
	user.Password = string(hashedNewPassword)
	if err := config.DB.Save(&user).Error; err != nil {
		http.Error(w, "Error updating password", http.StatusInternalServerError)
		return
	}

	// Шаг 6: Возвращаем сообщение об успешной смене пароля
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"message": "Password changed successfully"})
}

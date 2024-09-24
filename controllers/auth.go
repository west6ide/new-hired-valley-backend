package controllers

import (
	"encoding/json"
	"github.com/dgrijalva/jwt-go"
	"golang.org/x/crypto/bcrypt"
	"hired-valley-backend/config"
	"hired-valley-backend/models"
	"net/http"
	"strings"
	"time"
)

var jwtKey = []byte("12345678") // Ваш секретный ключ

type Claims struct {
	Email string `json:"email"`
	jwt.StandardClaims
}

// Register: Обычная регистрация с паролем и генерация JWT токена
// Register: Обычная регистрация с паролем
func Register(w http.ResponseWriter, r *http.Request) {
	var user models.User
	if err := json.NewDecoder(r.Body).Decode(&user); err != nil {
		http.Error(w, "Invalid input", http.StatusBadRequest)
		return
	}

	// Проверяем, существует ли пользователь с таким email
	var existingUser models.User
	if err := config.DB.Where("email = ?", user.Email).First(&existingUser).Error; err == nil {
		http.Error(w, "Email already registered", http.StatusConflict)
		return
	}

	// Хэшируем пароль
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
	if err != nil {
		http.Error(w, "Error hashing password", http.StatusInternalServerError)
		return
	}
	user.Password = string(hashedPassword)

	// Создание JWT токена
	expirationTime := time.Now().Add(24 * time.Hour)
	claims := &Claims{
		Email: user.Email,
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: expirationTime.Unix(),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(jwtKey)
	if err != nil {
		http.Error(w, "Error generating token", http.StatusInternalServerError)
		return
	}

	// Сохраняем пользователя и токен в базе данных
	user.Token = tokenString
	if err := config.DB.Create(&user).Error; err != nil {
		http.Error(w, "Error creating user", http.StatusInternalServerError)
		return
	}

	// Убираем токен из структуры пользователя перед отправкой
	user.Token = ""

	// Возвращаем пользователя и токен отдельно
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"user":  user,
		"token": tokenString,
	})
}

// Login: Вход с паролем и генерация JWT
// Login: Вход с паролем и генерация JWT
func Login(w http.ResponseWriter, r *http.Request) {
	var inputUser models.User
	if err := json.NewDecoder(r.Body).Decode(&inputUser); err != nil {
		http.Error(w, "Invalid input", http.StatusBadRequest)
		return
	}

	var user models.User
	// Поиск пользователя по email
	if err := config.DB.Where("email = ?", inputUser.Email).First(&user).Error; err != nil {
		http.Error(w, "User not found", http.StatusUnauthorized)
		return
	}

	// Проверка пароля
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(inputUser.Password)); err != nil {
		http.Error(w, "Invalid password", http.StatusUnauthorized)
		return
	}

	// Создание JWT токена
	expirationTime := time.Now().Add(24 * time.Hour)
	claims := &Claims{
		Email: user.Email,
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: expirationTime.Unix(),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(jwtKey)
	if err != nil {
		http.Error(w, "Error generating token", http.StatusInternalServerError)
		return
	}

	// Обновляем токен в базе данных
	user.Token = tokenString
	if err := config.DB.Save(&user).Error; err != nil {
		http.Error(w, "Error updating user token", http.StatusInternalServerError)
		return
	}

	// Возвращаем токен клиенту
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"token": tokenString})
}

// GetProfile: Получение профиля пользователя по токену
func GetProfile(w http.ResponseWriter, r *http.Request) {
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		http.Error(w, "Authorization header required", http.StatusUnauthorized)
		return
	}

	// Убираем "Bearer " из начала заголовка
	tokenString := strings.TrimPrefix(authHeader, "Bearer ")
	claims := &Claims{}

	token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
		return jwtKey, nil
	})

	if err != nil || !token.Valid {
		http.Error(w, "Invalid token", http.StatusUnauthorized)
		return
	}

	// Поиск пользователя по email из токена
	var user models.User
	if err := config.DB.Where("email = ?", claims.Email).First(&user).Error; err != nil {
		http.Error(w, "User not found", http.StatusUnauthorized)
		return
	}

	// Возвращаем информацию о пользователе
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(user)
}

// Logout: Инвалидировать сессию (удаление токена)
func Logout(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"message": "Logged out successfully"})
}

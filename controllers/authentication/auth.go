package authentication

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/dgrijalva/jwt-go"
	"golang.org/x/crypto/bcrypt"
	"hired-valley-backend/config"
	"hired-valley-backend/models/users"
	"net/http"
	"os"
	"strings"
	"time"
)

var JwtKey = []byte(os.Getenv("JWT_SECRET"))

type Claims struct {
	Email  string `json:"email"`
	Role   string `json:"role"`
	UserID uint   `json:"user_id"`
	jwt.StandardClaims
}

func Register(w http.ResponseWriter, r *http.Request) {
	var user users.User
	if err := json.NewDecoder(r.Body).Decode(&user); err != nil {
		http.Error(w, "Invalid input", http.StatusBadRequest)
		return
	}

	// Хэшируем пароль
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
	if err != nil {
		http.Error(w, "Error hashing password", http.StatusInternalServerError)
		return
	}
	user.Password = string(hashedPassword)
	user.Provider = "local"

	// Создаем запись в базе данных
	if err := config.DB.Create(&user).Error; err != nil {
		http.Error(w, fmt.Sprintf("Error creating user: %v", err), http.StatusInternalServerError)
		return
	}

	fmt.Printf("User registered with ID: %d\n", user.ID)

	// Генерируем токен
	tokenString, err := generateToken(user.ID, user.Email, user.Role)
	if err != nil {
		http.Error(w, "Error generating token", http.StatusInternalServerError)
		return
	}

	// Сохраняем токен в базе данных
	if err := config.DB.Model(&user).Update("token", tokenString).Error; err != nil {
		http.Error(w, "Error saving token", http.StatusInternalServerError)
		return
	}

	// Возвращаем токен
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"token": tokenString})
}

func Login(w http.ResponseWriter, r *http.Request) {
	var inputUser users.User
	if err := json.NewDecoder(r.Body).Decode(&inputUser); err != nil {
		http.Error(w, "Invalid input", http.StatusBadRequest)
		return
	}

	// Проверяем пользователя в базе данных
	var user users.User
	if err := config.DB.Where("email = ?", inputUser.Email).First(&user).Error; err != nil {
		http.Error(w, "User not found", http.StatusUnauthorized)
		return
	}

	// Сравниваем хэш пароля
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(inputUser.Password)); err != nil {
		http.Error(w, "Invalid password", http.StatusUnauthorized)
		return
	}

	fmt.Printf("User logged in with ID: %d\n", user.ID)

	// Генерируем токен
	tokenString, err := generateToken(user.ID, user.Email, user.Role)
	if err != nil {
		http.Error(w, "Error generating token", http.StatusInternalServerError)
		return
	}

	// Сохраняем токен в базе данных
	if err := config.DB.Model(&user).Update("token", tokenString).Error; err != nil {
		http.Error(w, "Error saving token", http.StatusInternalServerError)
		return
	}

	// Возвращаем токен
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"token": tokenString})
}

func ValidateToken(r *http.Request) (*Claims, error) {
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		return nil, errors.New("missing authorization header")
	}

	tokenString := strings.TrimPrefix(authHeader, "Bearer ")
	claims := &Claims{}
	token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
		return JwtKey, nil
	})
	if err != nil || !token.Valid {
		return nil, errors.New("invalid token")
	}

	var user users.User
	if err := config.DB.Where("id = ? AND token = ?", claims.UserID, tokenString).First(&user).Error; err != nil {
		return nil, errors.New("token mismatch")
	}

	fmt.Printf("Token validated with userID: %d\n", claims.UserID)
	return claims, nil
}

func generateToken(userID uint, email, role string) (string, error) {
	expirationTime := time.Now().Add(24 * time.Hour)
	claims := &Claims{
		UserID: userID,
		Email:  email,
		Role:   role,
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: expirationTime.Unix(),
		},
	}

	fmt.Printf("Generating token with userID: %d\n", userID)

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(JwtKey)
}

func GetProfile(w http.ResponseWriter, r *http.Request) {
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		http.Error(w, "Authorization header required", http.StatusUnauthorized)
		return
	}

	tokenString := strings.TrimPrefix(authHeader, "Bearer ")
	claims := &Claims{}

	token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
		return JwtKey, nil
	})

	if err != nil || !token.Valid {
		http.Error(w, "Invalid token", http.StatusUnauthorized)
		return
	}

	var user users.User
	if err := config.DB.Where("email = ? AND provider = ?", claims.Email, "local").First(&user).Error; err != nil {
		http.Error(w, "User not found", http.StatusUnauthorized)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(user)
}

func Logout(w http.ResponseWriter, r *http.Request) {
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		http.Error(w, "Authorization header required", http.StatusUnauthorized)
		return
	}

	claims := &Claims{}

	if err := config.DB.Model(&users.User{}).Where("id = ?", claims.UserID).Update("token", "").Error; err != nil {
		http.Error(w, "Error logging out", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"message": "Logged out successfully"})
}

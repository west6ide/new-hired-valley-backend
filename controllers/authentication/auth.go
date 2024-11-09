package authentication

import (
	"github.com/dgrijalva/jwt-go"
	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
	"hired-valley-backend/config"
	"hired-valley-backend/models/users"
	"log"
	"net/http"
	"os"
	"strings"
	"time"
)

var JwtKey = []byte(os.Getenv("JWT_SECRET"))

type Claims struct {
	Email  string `json:"email"`
	Role   string `json:"role"`
	UserID uint   `gorm:"primaryKey"`
	jwt.StandardClaims
}

// Register: регистрация с паролем и выбором роли
func Register(c *gin.Context) {
	var user users.User
	if err := c.ShouldBindJSON(&user); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input"})
		return
	}

	log.Printf("Попытка регистрации пользователя: %+v", user)

	// Проверка на существование пользователя с таким email
	var existingUser users.User
	if err := config.DB.Where("email = ? AND provider = ?", user.Email, "local").First(&existingUser).Error; err == nil {
		c.JSON(http.StatusConflict, gin.H{"error": "Email already registered"})
		return
	}

	// Хэшируем пароль
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error hashing password"})
		return
	}
	user.Password = string(hashedPassword)
	user.Provider = "local"

	// Валидация роли
	if user.Role != "user" && user.Role != "mentor" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid role. Allowed roles: user, mentor"})
		return
	}

	// Создание JWT токена
	expirationTime := time.Now().Add(24 * time.Hour)
	claims := &Claims{
		UserID: user.ID,
		Email:  user.Email,
		Role:   user.Role,
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: expirationTime.Unix(),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(JwtKey)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error generating token"})
		return
	}

	// Сохраняем пользователя и токен
	user.AccessToken = tokenString
	if err := config.DB.Create(&user).Error; err != nil {
		log.Printf("Ошибка при создании пользователя: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error creating user"})
		return
	}

	user.AccessToken = ""
	c.JSON(http.StatusCreated, gin.H{
		"user":  user,
		"token": tokenString,
	})
}

// Login: Вход с паролем и генерация JWT
func Login(c *gin.Context) {
	var inputUser users.User
	if err := c.ShouldBindJSON(&inputUser); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input"})
		return
	}

	var user users.User
	if err := config.DB.Where("email = ? AND provider = ?", inputUser.Email, "local").First(&user).Error; err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not found"})
		return
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(inputUser.Password)); err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid password"})
		return
	}

	expirationTime := time.Now().Add(24 * time.Hour)
	claims := &Claims{
		UserID: user.ID,
		Email:  user.Email,
		Role:   user.Role,
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: expirationTime.Unix(),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(JwtKey)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error generating token"})
		return
	}

	user.AccessToken = tokenString
	if err := config.DB.Save(&user).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error updating user token"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"token": tokenString})
}

// GetProfile: Получение профиля по токену
func GetProfile(c *gin.Context) {
	authHeader := c.GetHeader("Authorization")
	if authHeader == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Authorization header required"})
		return
	}

	tokenString := strings.TrimPrefix(authHeader, "Bearer ")
	claims := &Claims{}

	token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
		return JwtKey, nil
	})

	if err != nil || !token.Valid {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token"})
		return
	}

	var user users.User
	if err := config.DB.Where("email = ? AND provider = ?", claims.Email, "local").First(&user).Error; err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not found"})
		return
	}

	c.JSON(http.StatusOK, user)
}

// Logout: Завершение сеанса
func Logout(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"message": "Logged out successfully"})
}

func AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Authorization header required"})
			c.Abort()
			return
		}

		tokenString := strings.TrimPrefix(authHeader, "Bearer ")
		claims := &Claims{}
		token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
			return JwtKey, nil
		})

		if err != nil || !token.Valid {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid or expired token"})
			c.Abort()
			return
		}

		// Проверка и логирование аутентифицированного пользователя
		log.Printf("Authenticated user ID: %v, Role: %v", claims.UserID, claims.Role)

		c.Set("userID", claims.UserID)
		c.Set("userRole", claims.Role)
		c.Next()
	}
}

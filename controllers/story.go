package controllers

import (
	"encoding/json"
	"github.com/dgrijalva/jwt-go"
	"hired-valley-backend/controllers/authentication"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"hired-valley-backend/config"
	"hired-valley-backend/models/users"
)

// CreateStory - обработчик для создания истории
func CreateStory(w http.ResponseWriter, r *http.Request) {
	// Получаем заголовок авторизации
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		http.Error(w, "Authorization header required", http.StatusUnauthorized)
		return
	}

	// Убираем "Bearer " из начала заголовка
	tokenString := strings.TrimPrefix(authHeader, "Bearer ")
	claims := &authentication.Claims{}

	// Парсим и валидируем токен
	token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
		return authentication.JwtKey, nil
	})

	if err != nil || !token.Valid {
		http.Error(w, "Invalid token", http.StatusUnauthorized)
		return
	}

	// Создаем новую историю, привязывая её к идентификатору пользователя из токена
	var newStory users.Story
	if err := json.NewDecoder(r.Body).Decode(&newStory); err != nil {
		http.Error(w, "Invalid input", http.StatusBadRequest)
		return
	}

	newStory.UserID = claims.UserID                     // Привязка истории к пользователю
	newStory.CreatedAt = time.Now()                     // Устанавливаем время создания истории
	newStory.ExpiresAt = time.Now().Add(24 * time.Hour) // Устанавливаем срок действия истории на 24 часа

	// Сохраняем историю в базе данных
	if err := config.DB.Create(&newStory).Error; err != nil {
		http.Error(w, "Error saving story", http.StatusInternalServerError)
		return
	}

	// Возвращаем ответ с созданной историей
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(newStory)
}

// GetActiveStories - обработчик для получения всех активных историй
func GetActiveStories(w http.ResponseWriter, r *http.Request) {
	var stories []users.Story
	config.DB.Where("expires_at > ? AND is_archived = ?", time.Now(), false).Find(&stories)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(stories)
}

// ArchiveStory - обработчик для архивации истории
func ArchiveStory(w http.ResponseWriter, r *http.Request) {
	idStr := r.URL.Query().Get("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "Invalid story ID", http.StatusBadRequest)
		return
	}

	var story users.Story
	if err := config.DB.First(&story, id).Error; err != nil {
		http.Error(w, "Story not found", http.StatusNotFound)
		return
	}

	story.IsArchived = true
	if err := config.DB.Save(&story).Error; err != nil {
		log.Printf("Ошибка при архивации истории: %v", err)
		http.Error(w, "Failed to archive story", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(story)
}

// GetArchivedStories - обработчик для получения архивированных историй
func GetArchivedStories(w http.ResponseWriter, r *http.Request) {
	userIDStr := r.URL.Query().Get("user_id")
	userID, err := strconv.Atoi(userIDStr)
	if err != nil {
		http.Error(w, "Invalid user ID", http.StatusBadRequest)
		return
	}

	var stories []users.Story
	config.DB.Where("user_id = ? AND is_archived = ?", userID, true).Find(&stories)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(stories)
}

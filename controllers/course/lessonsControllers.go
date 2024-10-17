package course

import (
	"encoding/json"
	"github.com/dgrijalva/jwt-go"
	"hired-valley-backend/config"
	"hired-valley-backend/controllers/authentication"
	"hired-valley-backend/models/courses"
	"net/http"
	"strconv"
	"strings"
)

// Листинг уроков курса
func ListLessons(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	courseIDStr := r.URL.Query().Get("course_id")
	courseID, err := strconv.ParseUint(courseIDStr, 10, 32)
	if err != nil {
		http.Error(w, "Invalid course ID", http.StatusBadRequest)
		return
	}

	var lessons []courses.Lesson
	if err := config.DB.Where("course_id = ?", uint(courseID)).Find(&lessons).Error; err != nil {
		http.Error(w, "Failed to list lessons", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(lessons)
}

// CreateLesson функция для создания нового урока в курсе
func CreateLesson(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Извлекаем токен из заголовка Authorization
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		http.Error(w, "Authorization header required", http.StatusUnauthorized)
		return
	}

	// Удаляем префикс "Bearer "
	tokenString := strings.TrimPrefix(authHeader, "Bearer ")

	// Парсим токен и извлекаем claims
	claims := &authentication.Claims{}
	token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
		return []byte(authentication.JwtKey), nil
	})
	if err != nil || !token.Valid {
		http.Error(w, "Invalid token", http.StatusUnauthorized)
		return
	}

	// Проверяем, что пользователь является инструктором
	if claims.Role != "instructor" {
		http.Error(w, "Only instructors can create lessons", http.StatusForbidden)
		return
	}

	// Извлекаем CourseID из URL или тела запроса
	courseIDStr := r.URL.Query().Get("course_id")
	courseID, err := strconv.Atoi(courseIDStr)
	if err != nil {
		http.Error(w, "Invalid course ID", http.StatusBadRequest)
		return
	}

	// Декодируем данные урока из тела запроса
	var lesson courses.Lesson
	if err := json.NewDecoder(r.Body).Decode(&lesson); err != nil {
		http.Error(w, "Invalid input", http.StatusBadRequest)
		return
	}

	// Присваиваем CourseID уроку
	lesson.CourseID = uint(courseID)

	// Сохраняем урок в базе данных
	if err := config.DB.Create(&lesson).Error; err != nil {
		http.Error(w, "Failed to create lesson", http.StatusInternalServerError)
		return
	}

	// Возвращаем успешный ответ с данными урока
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(lesson)
}

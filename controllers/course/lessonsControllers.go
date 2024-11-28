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

// ListLessons - получение всех уроков курса
func ListLessons(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	courseIDStr := r.URL.Query().Get("course_id")
	courseID, err := strconv.Atoi(courseIDStr)
	if err != nil || courseID <= 0 {
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

// GetLessonByID - получение урока по ID
func GetLessonByID(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	lessonIDStr := strings.TrimPrefix(r.URL.Path, "/lessons/")
	lessonID, err := strconv.Atoi(lessonIDStr)
	if err != nil || lessonID <= 0 {
		http.Error(w, "Invalid lesson ID", http.StatusBadRequest)
		return
	}

	var lesson courses.Lesson
	if err := config.DB.First(&lesson, uint(lessonID)).Error; err != nil {
		http.Error(w, "Lesson not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(lesson)
}

// CreateLesson - создание нового урока
func CreateLesson(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		http.Error(w, "Authorization header required", http.StatusUnauthorized)
		return
	}

	tokenString := strings.TrimPrefix(authHeader, "Bearer ")
	claims := &authentication.Claims{}
	token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
		return []byte(authentication.JwtKey), nil
	})
	if err != nil || !token.Valid {
		http.Error(w, "Invalid token", http.StatusUnauthorized)
		return
	}

	if claims.Role != "mentor" {
		http.Error(w, "Only mentors can create lessons", http.StatusForbidden)
		return
	}

	var lesson courses.Lesson
	if err := json.NewDecoder(r.Body).Decode(&lesson); err != nil {
		http.Error(w, "Invalid input", http.StatusBadRequest)
		return
	}

	if err := config.DB.Create(&lesson).Error; err != nil {
		http.Error(w, "Failed to create lesson", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(lesson)
}

// UpdateLesson - обновление урока
func UpdateLesson(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPut {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	lessonIDStr := strings.TrimPrefix(r.URL.Path, "/lessons/")
	lessonID, err := strconv.Atoi(lessonIDStr)
	if err != nil || lessonID <= 0 {
		http.Error(w, "Invalid lesson ID", http.StatusBadRequest)
		return
	}

	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		http.Error(w, "Authorization header required", http.StatusUnauthorized)
		return
	}

	tokenString := strings.TrimPrefix(authHeader, "Bearer ")
	claims := &authentication.Claims{}
	token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
		return []byte(authentication.JwtKey), nil
	})
	if err != nil || !token.Valid {
		http.Error(w, "Invalid token", http.StatusUnauthorized)
		return
	}

	var lesson courses.Lesson
	if err := config.DB.First(&lesson, uint(lessonID)).Error; err != nil {
		http.Error(w, "Lesson not found", http.StatusNotFound)
		return
	}

	if claims.Role != "mentor" || claims.UserID != lesson.InstructorID {
		http.Error(w, "Permission denied", http.StatusForbidden)
		return
	}

	var updatedLesson courses.Lesson
	if err := json.NewDecoder(r.Body).Decode(&updatedLesson); err != nil {
		http.Error(w, "Invalid input", http.StatusBadRequest)
		return
	}

	lesson.Title = updatedLesson.Title
	lesson.Content = updatedLesson.Content
	lesson.VimeoVideoLink = updatedLesson.VimeoVideoLink

	if err := config.DB.Save(&lesson).Error; err != nil {
		http.Error(w, "Failed to update lesson", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(lesson)
}

// DeleteLesson - удаление урока
func DeleteLesson(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	lessonIDStr := strings.TrimPrefix(r.URL.Path, "/lessons/")
	lessonID, err := strconv.Atoi(lessonIDStr)
	if err != nil || lessonID <= 0 {
		http.Error(w, "Invalid lesson ID", http.StatusBadRequest)
		return
	}

	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		http.Error(w, "Authorization header required", http.StatusUnauthorized)
		return
	}

	tokenString := strings.TrimPrefix(authHeader, "Bearer ")
	claims := &authentication.Claims{}
	token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
		return []byte(authentication.JwtKey), nil
	})
	if err != nil || !token.Valid {
		http.Error(w, "Invalid token", http.StatusUnauthorized)
		return
	}

	var lesson courses.Lesson
	if err := config.DB.First(&lesson, uint(lessonID)).Error; err != nil {
		http.Error(w, "Lesson not found", http.StatusNotFound)
		return
	}

	if claims.Role != "mentor" || claims.UserID != lesson.InstructorID {
		http.Error(w, "Permission denied", http.StatusForbidden)
		return
	}

	if err := config.DB.Delete(&lesson).Error; err != nil {
		http.Error(w, "Failed to delete lesson", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

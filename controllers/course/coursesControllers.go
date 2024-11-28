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

// ListCourses - получение всех курсов
// ListCourses - получение всех курсов менторов (фильтр по instructor_id)
func ListCourses(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Получение параметра instructor_id из запроса
	instructorIDStr := r.URL.Query().Get("instructor_id")
	if instructorIDStr == "" {
		http.Error(w, "Instructor ID is required", http.StatusBadRequest)
		return
	}

	instructorID, err := strconv.Atoi(instructorIDStr)
	if err != nil || instructorID <= 0 {
		http.Error(w, "Invalid Instructor ID", http.StatusBadRequest)
		return
	}

	var courses []courses.Course
	if err := config.DB.Where("instructor_id = ?", instructorID).Find(&courses).Error; err != nil {
		http.Error(w, "Failed to list courses", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(courses)
}

// GetCourseByID - получение курса по ID
func GetCourseByID(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Получение ID из URL
	idStr := strings.TrimPrefix(r.URL.Path, "/courses/")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "Invalid course ID", http.StatusBadRequest)
		return
	}

	var course courses.Course
	if err := config.DB.First(&course, id).Error; err != nil {
		http.Error(w, "Course not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(course)
}

// CreateCourse - создание нового курса
func CreateCourse(w http.ResponseWriter, r *http.Request) {
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
		http.Error(w, "Only mentors can create courses", http.StatusForbidden)
		return
	}

	var course courses.Course
	if err := json.NewDecoder(r.Body).Decode(&course); err != nil {
		http.Error(w, "Invalid input", http.StatusBadRequest)
		return
	}

	course.InstructorID = claims.UserID
	if err := config.DB.Create(&course).Error; err != nil {
		http.Error(w, "Failed to create course", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(course)
}

// UpdateCourse - обновление курса
func UpdateCourse(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPut {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	idStr := strings.TrimPrefix(r.URL.Path, "/courses/")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "Invalid course ID", http.StatusBadRequest)
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

	var course courses.Course
	if err := config.DB.First(&course, id).Error; err != nil {
		http.Error(w, "Course not found", http.StatusNotFound)
		return
	}

	if claims.Role != "mentor" || course.InstructorID != claims.UserID {
		http.Error(w, "Permission denied", http.StatusForbidden)
		return
	}

	var updatedCourse courses.Course
	if err := json.NewDecoder(r.Body).Decode(&updatedCourse); err != nil {
		http.Error(w, "Invalid input", http.StatusBadRequest)
		return
	}

	course.Title = updatedCourse.Title
	course.Description = updatedCourse.Description
	course.Price = updatedCourse.Price

	if err := config.DB.Save(&course).Error; err != nil {
		http.Error(w, "Failed to update course", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(course)
}

// DeleteCourse - удаление курса
func DeleteCourse(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	idStr := strings.TrimPrefix(r.URL.Path, "/courses/")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "Invalid course ID", http.StatusBadRequest)
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

	var course courses.Course
	if err := config.DB.First(&course, id).Error; err != nil {
		http.Error(w, "Course not found", http.StatusNotFound)
		return
	}

	if claims.Role != "mentor" || course.InstructorID != claims.UserID {
		http.Error(w, "Permission denied", http.StatusForbidden)
		return
	}

	if err := config.DB.Delete(&course).Error; err != nil {
		http.Error(w, "Failed to delete course", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

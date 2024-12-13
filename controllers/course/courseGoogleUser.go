package course

import (
	"encoding/json"
	"hired-valley-backend/config"
	"hired-valley-backend/controllers/authentication"
	"hired-valley-backend/models/courses"
	"hired-valley-backend/models/users"
	"net/http"
	"strconv"
)

// Проверка роли пользователя
func isMentor(userID uint) (bool, error) {
	var user users.User
	if err := config.DB.First(&user, userID).Error; err != nil {
		return false, err
	}
	return user.Role == "mentor", nil
}

// ListCourses - получение всех курсов менторов (фильтр по instructor_id)
func ListCoursesGoogle(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Проверяем Google OAuth токен
	claims, err := authentication.ValidateGoogleToken(r)
	if err != nil {
		http.Error(w, "Unauthorized: "+err.Error(), http.StatusUnauthorized)
		return
	}

	// Проверка роли пользователя
	isMentor, err := isMentor(claims.UserID)
	if err != nil || !isMentor {
		http.Error(w, "Permission denied", http.StatusForbidden)
		return
	}

	var courses []courses.Course
	if err := config.DB.Where("instructor_id = ?", claims.UserID).Find(&courses).Error; err != nil {
		http.Error(w, "Failed to list courses", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(courses)
}

// GetCourseByID - получение курса по ID
func GetCourseByIDGoogle(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Проверяем Google OAuth токен
	claims, err := authentication.ValidateGoogleToken(r)
	if err != nil {
		http.Error(w, "Unauthorized: "+err.Error(), http.StatusUnauthorized)
		return
	}

	// Получение ID из параметров запроса
	idStr := r.URL.Query().Get("id")
	if idStr == "" {
		http.Error(w, "Course ID is required", http.StatusBadRequest)
		return
	}

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

	if course.InstructorID != claims.UserID {
		http.Error(w, "Permission denied", http.StatusForbidden)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(course)
}

// CreateCourse - создание нового курса
func CreateCourseGoogle(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Проверяем Google OAuth токен
	claims, err := authentication.ValidateGoogleToken(r)
	if err != nil {
		http.Error(w, "Unauthorized: "+err.Error(), http.StatusUnauthorized)
		return
	}

	// Проверка роли пользователя
	isMentor, err := isMentor(claims.UserID)
	if err != nil || !isMentor {
		http.Error(w, "Only mentors can create courses", http.StatusForbidden)
		return
	}

	var course courses.Course
	if err := json.NewDecoder(r.Body).Decode(&course); err != nil {
		http.Error(w, "Invalid input", http.StatusBadRequest)
		return
	}

	if len(course.Tags) == 0 {
		http.Error(w, "Tags are required", http.StatusBadRequest)
		return
	}

	course.InstructorID = claims.UserID
	if err := config.DB.Create(&course).Error; err != nil {
		http.Error(w, "Failed to create course: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(course)
}

// UpdateCourse - обновление курса
func UpdateCourseGoogle(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPut {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Проверяем Google OAuth токен
	claims, err := authentication.ValidateGoogleToken(r)
	if err != nil {
		http.Error(w, "Unauthorized: "+err.Error(), http.StatusUnauthorized)
		return
	}

	// Получение ID из параметров запроса
	idStr := r.URL.Query().Get("id")
	if idStr == "" {
		http.Error(w, "Course ID is required", http.StatusBadRequest)
		return
	}

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

	if course.InstructorID != claims.UserID {
		http.Error(w, "Permission denied", http.StatusForbidden)
		return
	}

	// Декодируем обновляемые данные
	var updates map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&updates); err != nil {
		http.Error(w, "Invalid input", http.StatusBadRequest)
		return
	}

	// Обновляем только те поля, которые были переданы
	if title, ok := updates["title"].(string); ok {
		course.Title = title
	}
	if description, ok := updates["description"].(string); ok {
		course.Description = description
	}
	if price, ok := updates["price"].(float64); ok {
		course.Price = price
	}

	// Сохраняем обновления
	if err := config.DB.Save(&course).Error; err != nil {
		http.Error(w, "Failed to update course", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(course)
}

// DeleteCourse - удаление курса
func DeleteCourseGoogle(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Проверяем Google OAuth токен
	claims, err := authentication.ValidateGoogleToken(r)
	if err != nil {
		http.Error(w, "Unauthorized: "+err.Error(), http.StatusUnauthorized)
		return
	}

	// Получение ID из параметров запроса
	idStr := r.URL.Query().Get("id")
	if idStr == "" {
		http.Error(w, "Course ID is required", http.StatusBadRequest)
		return
	}

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

	if course.InstructorID != claims.UserID {
		http.Error(w, "Permission denied", http.StatusForbidden)
		return
	}

	if err := config.DB.Delete(&course).Error; err != nil {
		http.Error(w, "Failed to delete course", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

package course

import (
	"encoding/json"
	"hired-valley-backend/config"
	"hired-valley-backend/models/courses"
	"net/http"
	"strconv"
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

// Функция для создания одного или нескольких уроков для конкретного курса
func CreateLessons(w http.ResponseWriter, r *http.Request) {
	// Извлекаем ID курса из URL
	courseIDStr := r.URL.Query().Get("course_id")
	courseID, err := strconv.ParseUint(courseIDStr, 10, 32)
	if err != nil {
		http.Error(w, "Invalid course ID", http.StatusBadRequest)
		return
	}

	// Декодируем массив уроков из тела запроса
	var lessons []courses.Lesson
	if err := json.NewDecoder(r.Body).Decode(&lessons); err != nil {
		http.Error(w, "Invalid input", http.StatusBadRequest)
		return
	}

	// Привязываем все уроки к курсу
	for i := range lessons {
		lessons[i].CourseID = uint(courseID)
	}

	// Сохраняем уроки в базе данных
	if err := config.DB.Create(&lessons).Error; err != nil {
		http.Error(w, "Failed to create lessons", http.StatusInternalServerError)
		return
	}

	// Возвращаем успешный ответ с данными уроков
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(lessons)
}

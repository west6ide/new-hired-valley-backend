package courses

import (
	"encoding/json"
	"errors"
	"github.com/gorilla/mux"
	"gorm.io/gorm"
	"hired-valley-backend/config"
	"hired-valley-backend/models"
	"net/http"
	"strconv"
)

// Получение всех курсов
func GetCourses(w http.ResponseWriter, r *http.Request) {
	var courses models.Course
	config.DB.Preload("Modules").Find(&courses)
	json.NewEncoder(w).Encode(courses)
}

// Получение курса по ID
func GetCourseByID(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	id, _ := strconv.Atoi(params["id"])
	var course models.Course
	if err := config.DB.Preload("Modules").First(&course, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			http.Error(w, "Курс не найден", http.StatusNotFound)
			return
		}
		http.Error(w, "Ошибка сервера", http.StatusInternalServerError)
		return
	}
	json.NewEncoder(w).Encode(course)
}

// Создание курса
func CreateCourse(w http.ResponseWriter, r *http.Request) {
	var course models.Course
	if err := json.NewDecoder(r.Body).Decode(&course); err != nil {
		http.Error(w, "Неверный формат данных", http.StatusBadRequest)
		return
	}
	config.DB.Create(&course)
	json.NewEncoder(w).Encode(course)
}

// Обновление курса
func UpdateCourse(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	id, _ := strconv.Atoi(params["id"])
	var course models.Course
	if err := config.DB.First(&course, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			http.Error(w, "Курс не найден", http.StatusNotFound)
			return
		}
		http.Error(w, "Ошибка сервера", http.StatusInternalServerError)
		return
	}
	if err := json.NewDecoder(r.Body).Decode(&course); err != nil {
		http.Error(w, "Неверный формат данных", http.StatusBadRequest)
		return
	}
	config.DB.Save(&course)
	json.NewEncoder(w).Encode(course)
}

// Удаление курса
func DeleteCourse(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	id, _ := strconv.Atoi(params["id"])
	var course models.Course
	if err := config.DB.First(&course, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			http.Error(w, "Курс не найден", http.StatusNotFound)
			return
		}
		http.Error(w, "Ошибка сервера", http.StatusInternalServerError)
		return
	}
	config.DB.Delete(&course)
	json.NewEncoder(w).Encode("Курс успешно удален")
}

package courses

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
	"gorm.io/gorm"
	"hired-valley-backend/config"
	"hired-valley-backend/models"
)

// Получение всех курсов
func GetCourses(w http.ResponseWriter, r *http.Request) {
	var courses []models.Course
	if err := config.DB.Preload("Modules").Find(&courses).Error; err != nil {
		http.Error(w, "Ошибка при получении курсов", http.StatusInternalServerError)
		return
	}
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
		log.Printf("Ошибка при декодировании тела запроса: %v", err)
		http.Error(w, "Неверный формат данных", http.StatusBadRequest)
		return
	}

	// Проверка обязательных полей
	if course.Title == "" || course.Description == "" || course.InstructorID == 0 {
		http.Error(w, "Необходимо заполнить все обязательные поля", http.StatusBadRequest)
		return
	}

	// Проверяем, существует ли инструктор с данным ID
	var instructor models.User
	if err := config.DB.First(&instructor, course.InstructorID).Error; err != nil {
		log.Printf("Инструктор не найден: %v", err)
		http.Error(w, "Инструктор не найден", http.StatusNotFound)
		return
	}

	// Сохраняем курс в базу данных
	if err := config.DB.Create(&course).Error; err != nil {
		log.Printf("Ошибка при сохранении курса: %v", err)
		http.Error(w, "Ошибка при сохранении курса", http.StatusInternalServerError)
		return
	}

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
	json.NewEncoder(w).Encode(map[string]string{"message": "Курс успешно удален"})
}

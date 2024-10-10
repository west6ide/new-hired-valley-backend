package courses

import (
	"encoding/json"
	"hired-valley-backend/config"
	"hired-valley-backend/models"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
	"gorm.io/gorm"
)

// Функция для загрузки видео в Vimeo и создания модуля курса
func CreateModule(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	courseID, _ := strconv.Atoi(params["course_id"])

	var module models.Module
	if err := json.NewDecoder(r.Body).Decode(&module); err != nil {
		http.Error(w, "Неверный формат данных", http.StatusBadRequest)
		return
	}

	// Загрузка видео на Vimeo
	videoURL, err := UploadToVimeo(module.VideoURL) // Передаем путь к видео
	if err != nil {
		http.Error(w, "Ошибка загрузки видео", http.StatusInternalServerError)
		return
	}

	// Сохранение URL видео, полученного от Vimeo
	module.VideoURL = videoURL
	module.CourseID = uint(courseID)

	config.DB.Create(&module)
	json.NewEncoder(w).Encode(module)
}

// Обновление модуля курса и загрузка нового видео на Vimeo (если нужно)
func UpdateModule(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	moduleID, _ := strconv.Atoi(params["id"])

	var module models.Module
	if err := config.DB.First(&module, moduleID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			http.Error(w, "Модуль не найден", http.StatusNotFound)
			return
		}
		http.Error(w, "Ошибка сервера", http.StatusInternalServerError)
		return
	}

	if err := json.NewDecoder(r.Body).Decode(&module); err != nil {
		http.Error(w, "Неверный формат данных", http.StatusBadRequest)
		return
	}

	// Если было загружено новое видео, обновляем его на Vimeo
	if module.VideoURL != "" {
		videoURL, err := UploadToVimeo(module.VideoURL) // Передаем новый путь к видео
		if err != nil {
			http.Error(w, "Ошибка загрузки видео", http.StatusInternalServerError)
			return
		}
		module.VideoURL = videoURL
	}

	config.DB.Save(&module)
	json.NewEncoder(w).Encode(module)
}

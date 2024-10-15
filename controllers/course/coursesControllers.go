package course

import (
	"encoding/json"
	"hired-valley-backend/config"
	"hired-valley-backend/models/courses"
	"mime/multipart"
	"net/http"
)

// Листинг курсов
func ListCourses(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var courses []courses.Course
	if err := config.DB.Find(&courses).Error; err != nil {
		http.Error(w, "Failed to list courses", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(courses)
}

// Создание курса
func CreateCourse(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var course courses.Course
	if err := json.NewDecoder(r.Body).Decode(&course); err != nil {
		http.Error(w, "Invalid input", http.StatusBadRequest)
		return
	}

	if err := config.DB.Create(&course).Error; err != nil {
		http.Error(w, "Failed to create course", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"message": "Course created"})
}

// Загрузка видео
func UploadVideo(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	file, _, err := r.FormFile("video")
	if err != nil {
		http.Error(w, "Failed to upload video", http.StatusBadRequest)
		return
	}
	defer file.Close()

	vimeoLink, err := uploadVideoToVimeo(file)
	if err != nil {
		http.Error(w, "Failed to upload video to Vimeo", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"vimeo_link": vimeoLink})
}

// Загрузка видео на Vimeo (заглушка)
func uploadVideoToVimeo(file multipart.File) (string, error) {
	// Логика для загрузки видео на Vimeo через API
	return "https://vimeo.com/video_id", nil
}

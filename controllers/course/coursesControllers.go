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

func CreateCourse(w http.ResponseWriter, r *http.Request) {
	var course courses.Course
	err := json.NewDecoder(r.Body).Decode(&course)
	if err != nil {
		http.Error(w, "Invalid input", http.StatusBadRequest)
		return
	}

	// Проверка обязательных полей
	if course.Title == "" || course.InstructorID == 0 {
		http.Error(w, "Missing required fields", http.StatusBadRequest)
		return
	}

	// Создание курса в базе данных
	if err := config.DB.Create(&course).Error; err != nil {
		http.Error(w, "Failed to create course", http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(course)
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

package course

import (
	"encoding/json"
	"fmt"
	"github.com/dgrijalva/jwt-go"
	"github.com/silentsokolov/go-vimeo/vimeo"
	"golang.org/x/oauth2"
	"hired-valley-backend/config"
	"hired-valley-backend/controllers/authentication"
	"hired-valley-backend/models/courses"
	"io"
	"mime/multipart"
	"net/http"
	"os"
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
	if claims.Role != "mentor" {
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

// UploadVideoToLesson загружает видео в Vimeo и сохраняет ссылку в lesson
func UploadVideoToLesson(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Авторизация через токен
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
		http.Error(w, "Only instructors can upload videos", http.StatusForbidden)
		return
	}

	// Получаем lesson_id
	lessonIDStr := r.URL.Query().Get("lesson_id")
	lessonID, err := strconv.Atoi(lessonIDStr)
	if err != nil || lessonID <= 0 {
		http.Error(w, "Invalid lesson ID", http.StatusBadRequest)
		return
	}

	// Проверяем существование урока
	var lesson courses.Lesson
	if err := config.DB.First(&lesson, lessonID).Error; err != nil {
		http.Error(w, "Lesson not found", http.StatusNotFound)
		return
	}

	// Получаем файл из запроса
	file, header, err := r.FormFile("video")
	if err != nil {
		http.Error(w, "Failed to read video file", http.StatusBadRequest)
		return
	}
	defer file.Close()

	// Загружаем видео на Vimeo
	vimeoLink, err := uploadVideoToVimeo(file, header.Filename)
	if err != nil {
		http.Error(w, "Failed to upload video to Vimeo: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Сохраняем ссылку на Vimeo в базе данных
	lesson.VimeoVideoLink = vimeoLink
	if err := config.DB.Save(&lesson).Error; err != nil {
		http.Error(w, "Failed to save video link to lesson", http.StatusInternalServerError)
		return
	}

	// Возвращаем успешный ответ
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{
		"message":    "Video uploaded successfully",
		"vimeo_link": vimeoLink,
	})
}

// uploadVideoToVimeo загружает файл на Vimeo
func uploadVideoToVimeo(file multipart.File, fileName string) (string, error) {
	client := createVimeoClient()

	// Создаём временный файл
	tempFilePath := "/tmp/" + fileName
	tempFile, err := os.Create(tempFilePath)
	if err != nil {
		return "", fmt.Errorf("cannot create temporary file: %v", err)
	}
	defer os.Remove(tempFilePath) // Удаляем временный файл после использования
	defer tempFile.Close()

	// Копируем данные из multipart.File в временный файл
	if _, err := io.Copy(tempFile, file); err != nil {
		return "", fmt.Errorf("cannot copy file data: %v", err)
	}

	// Открываем временный файл как *os.File
	osFile, err := os.Open(tempFilePath)
	if err != nil {
		return "", fmt.Errorf("cannot open temporary file: %v", err)
	}
	defer osFile.Close()

	// Загружаем файл на Vimeo
	video, _, err := client.Users.UploadVideo("", osFile)
	if err != nil {
		return "", fmt.Errorf("failed to upload video to Vimeo: %v", err)
	}

	// Возвращаем ссылку на видео
	return video.Link, nil
}

func createVimeoClient() *vimeo.Client {
	token := "149d5f7f0963345bf59f85515b2df582" // Замените на токен, полученный в Vimeo Developer
	ts := oauth2.StaticTokenSource(&oauth2.Token{AccessToken: token})
	tc := oauth2.NewClient(oauth2.NoContext, ts)
	return vimeo.NewClient(tc, nil)
}

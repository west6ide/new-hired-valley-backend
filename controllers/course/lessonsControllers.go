package course

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/dgrijalva/jwt-go"
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

	// Извлекаем параметры и файл
	lessonIDStr := r.URL.Query().Get("lesson_id")
	lessonID, err := strconv.Atoi(lessonIDStr)
	if err != nil || lessonID <= 0 {
		http.Error(w, "Invalid lesson ID", http.StatusBadRequest)
		return
	}

	file, header, err := r.FormFile("video")
	if err != nil {
		http.Error(w, "Failed to read video file", http.StatusBadRequest)
		return
	}
	defer file.Close()

	// Загрузка видео на Vimeo
	vimeoLink, err := uploadVideoToVimeo(file, header.Filename)
	if err != nil {
		http.Error(w, "Failed to upload video to Vimeo: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Сохраняем ссылку в базе данных
	var lesson courses.Lesson
	if err := config.DB.First(&lesson, lessonID).Error; err != nil {
		http.Error(w, "Lesson not found", http.StatusNotFound)
		return
	}

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

const vimeoUploadURL = "https://api.vimeo.com/me/videos"

// uploadVideoToVimeo выполняет загрузку файла на Vimeo и возвращает ссылку на видео
func uploadVideoToVimeo(file multipart.File, fileName string) (string, error) {
	// Создаём временный файл для сохранения данных
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

	// Перечитываем временный файл как *os.File
	osFile, err := os.Open(tempFilePath)
	if err != nil {
		return "", fmt.Errorf("cannot open temporary file: %v", err)
	}
	defer osFile.Close()

	// Создаём HTTP-запрос с multipart/form-data
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	// Добавляем файл в запрос
	part, err := writer.CreateFormFile("file_data", fileName)
	if err != nil {
		return "", fmt.Errorf("cannot create form file: %v", err)
	}
	if _, err := io.Copy(part, osFile); err != nil {
		return "", fmt.Errorf("cannot copy file data: %v", err)
	}

	writer.Close()

	// Создаём HTTP-запрос
	req, err := http.NewRequest("POST", vimeoUploadURL, body)
	if err != nil {
		return "", fmt.Errorf("cannot create request: %v", err)
	}

	// Добавляем заголовки
	req.Header.Set("Authorization", "Bearer your_access_token") // Замените на реальный токен
	req.Header.Set("Content-Type", writer.FormDataContentType())

	// Выполняем запрос
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("request failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		respBody, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("failed to upload video: %v, response: %s", resp.StatusCode, string(respBody))
	}

	// Обрабатываем успешный ответ
	var response map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return "", fmt.Errorf("cannot decode response: %v", err)
	}

	// Извлекаем ссылку на видео
	link, ok := response["link"].(string)
	if !ok {
		return "", fmt.Errorf("video link not found in response")
	}

	return link, nil
}

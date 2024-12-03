package course

import (
	"context"
	"encoding/json"
	"fmt"
	"golang.org/x/oauth2"
	"google.golang.org/api/option"
	"google.golang.org/api/youtube/v3"
	"hired-valley-backend/config"
	"hired-valley-backend/controllers/authentication"
	"hired-valley-backend/models/courses"
	"hired-valley-backend/models/users"
	"mime/multipart"
	"net/http"
	"strconv"
	"time"
)

var (
	oauthToken string
)

// ListLessons - получение всех уроков курса
func ListLessons(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	courseIDStr := r.URL.Query().Get("course_id")
	courseID, err := strconv.Atoi(courseIDStr)
	if err != nil || courseID <= 0 {
		http.Error(w, "Invalid course ID", http.StatusBadRequest)
		return
	}

	var lessons []courses.Lesson
	if err := config.DB.Where("course_id = ?", uint(courseID)).Find(&lessons).Error; err != nil {
		http.Error(w, "Failed to list lessons", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(lessons)
}

// GetLessonByID - получение урока по ID
func GetLessonByID(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	lessonIDStr := r.URL.Query().Get("id")
	if lessonIDStr == "" {
		http.Error(w, "Lesson ID is required", http.StatusBadRequest)
		return
	}

	lessonID, err := strconv.Atoi(lessonIDStr)
	if err != nil || lessonID <= 0 {
		http.Error(w, "Invalid lesson ID", http.StatusBadRequest)
		return
	}

	var lesson courses.Lesson
	if err := config.DB.First(&lesson, uint(lessonID)).Error; err != nil {
		http.Error(w, "Lesson not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(lesson)
}

// CreateLesson - создание нового урока
func CreateLesson(w http.ResponseWriter, r *http.Request) {
	claims, err := authentication.ValidateToken(r)
	if err != nil || claims.Role != "mentor" {
		http.Error(w, "Unauthorized or forbidden", http.StatusUnauthorized)
		return
	}

	var lesson courses.Lesson
	if err := json.NewDecoder(r.Body).Decode(&lesson); err != nil {
		http.Error(w, "Invalid input", http.StatusBadRequest)
		return
	}

	lesson.InstructorID = claims.UserID
	if err := config.DB.Create(&lesson).Error; err != nil {
		http.Error(w, "Failed to create lesson", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(lesson)
}

// UpdateLesson - обновление урока
func UpdateLesson(w http.ResponseWriter, r *http.Request) {
	claims, err := authentication.ValidateToken(r)
	if err != nil || claims.Role != "mentor" {
		http.Error(w, "Unauthorized or forbidden", http.StatusUnauthorized)
		return
	}

	lessonIDStr := r.URL.Query().Get("id")
	if lessonIDStr == "" {
		http.Error(w, "Lesson ID is required", http.StatusBadRequest)
		return
	}

	lessonID, err := strconv.Atoi(lessonIDStr)
	if err != nil || lessonID <= 0 {
		http.Error(w, "Invalid lesson ID", http.StatusBadRequest)
		return
	}

	var lesson courses.Lesson
	if err := config.DB.First(&lesson, uint(lessonID)).Error; err != nil {
		http.Error(w, "Lesson not found", http.StatusNotFound)
		return
	}

	if lesson.InstructorID != claims.UserID {
		http.Error(w, "Permission denied", http.StatusForbidden)
		return
	}

	if err := json.NewDecoder(r.Body).Decode(&lesson); err != nil {
		http.Error(w, "Invalid input", http.StatusBadRequest)
		return
	}

	if err := config.DB.Save(&lesson).Error; err != nil {
		http.Error(w, "Failed to update lesson", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(lesson)
}

// DeleteLesson - удаление урока
func DeleteLesson(w http.ResponseWriter, r *http.Request) {
	claims, err := authentication.ValidateToken(r)
	if err != nil || claims.Role != "mentor" {
		http.Error(w, "Unauthorized or forbidden", http.StatusUnauthorized)
		return
	}

	lessonIDStr := r.URL.Query().Get("id")
	if lessonIDStr == "" {
		http.Error(w, "Lesson ID is required", http.StatusBadRequest)
		return
	}

	lessonID, err := strconv.Atoi(lessonIDStr)
	if err != nil || lessonID <= 0 {
		http.Error(w, "Invalid lesson ID", http.StatusBadRequest)
		return
	}

	var lesson courses.Lesson
	if err := config.DB.First(&lesson, uint(lessonID)).Error; err != nil {
		http.Error(w, "Lesson not found", http.StatusNotFound)
		return
	}

	if lesson.InstructorID != claims.UserID {
		http.Error(w, "Permission denied", http.StatusForbidden)
		return
	}

	if err := config.DB.Delete(&lesson).Error; err != nil {
		http.Error(w, "Failed to delete lesson", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// UploadVideoToLesson - загрузка видео на YouTube и сохранение ссылки в Lesson
func UploadVideoToLesson(w http.ResponseWriter, r *http.Request) {
	claims, err := authentication.ValidateToken(r)
	if err != nil || claims.Role != "mentor" {
		http.Error(w, "Unauthorized or forbidden", http.StatusUnauthorized)
		return
	}

	lessonIDStr := r.URL.Query().Get("lesson_id")
	if lessonIDStr == "" {
		http.Error(w, "Lesson ID is required", http.StatusBadRequest)
		return
	}

	lessonID, err := strconv.Atoi(lessonIDStr)
	if err != nil || lessonID <= 0 {
		http.Error(w, "Invalid lesson ID", http.StatusBadRequest)
		return
	}

	var lesson courses.Lesson
	if err := config.DB.First(&lesson, uint(lessonID)).Error; err != nil {
		http.Error(w, "Lesson not found", http.StatusNotFound)
		return
	}

	if lesson.InstructorID != claims.UserID {
		http.Error(w, "Permission denied", http.StatusForbidden)
		return
	}

	// Получаем YouTube токены из базы данных
	var youtubeUser users.YoutubeUser
	if err := config.DB.Where("user_id = ?", claims.UserID).First(&youtubeUser).Error; err != nil {
		http.Error(w, "Google user not found", http.StatusUnauthorized)
		return
	}

	// Читаем видео из формы
	file, header, err := r.FormFile("video")
	if err != nil {
		http.Error(w, "Failed to read video file", http.StatusBadRequest)
		return
	}
	defer file.Close()

	// Загрузка видео на YouTube
	videoID, err := uploadVideoToYouTube(file, header.Filename, youtubeUser.AccessToken, youtubeUser.RefreshToken, youtubeUser.Expiry)
	if err != nil {
		http.Error(w, "Failed to upload video to YouTube: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Обновляем информацию об уроке
	lesson.VideoLink = fmt.Sprintf("https://www.youtube.com/watch?v=%s", videoID)
	if err := config.DB.Save(&lesson).Error; err != nil {
		http.Error(w, "Failed to save lesson", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"message":   "Video uploaded successfully",
		"video_url": lesson.VideoLink,
		"lesson_id": lessonIDStr,
		"video_id":  videoID,
	})
}

func uploadVideoToYouTube(file multipart.File, fileName string, accessToken string, refreshToken string, expiry time.Time) (string, error) {
	ctx := context.Background()

	// Создаем токен вручную
	token := &oauth2.Token{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		Expiry:       expiry,
	}

	// Создаем TokenSource для автоматического обновления токенов
	tokenSource := authentication.GoogleOauthConfig.TokenSource(ctx, token)

	// Создаем сервис YouTube
	service, err := youtube.NewService(ctx, option.WithTokenSource(tokenSource))
	if err != nil {
		return "", fmt.Errorf("failed to create YouTube service: %v", err)
	}

	// Подготавливаем метаданные видео
	video := &youtube.Video{
		Snippet: &youtube.VideoSnippet{
			Title:       fileName,
			Description: "Uploaded via Hired Valley platform",
		},
		Status: &youtube.VideoStatus{
			PrivacyStatus: "unlisted",
		},
	}

	// Загрузка видео
	call := service.Videos.Insert([]string{"snippet", "status"}, video)
	response, err := call.Media(file).Do()
	if err != nil {
		return "", fmt.Errorf("failed to upload video: %v", err)
	}

	return response.Id, nil
}

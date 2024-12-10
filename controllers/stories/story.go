package stories

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"golang.org/x/oauth2"
	"google.golang.org/api/drive/v3"
	"google.golang.org/api/option"
	"gorm.io/gorm"
	"hired-valley-backend/config"
	"hired-valley-backend/controllers/authentication"
	"hired-valley-backend/models/story"
	"hired-valley-backend/models/users"
	"mime/multipart"
	"net/http"
	"strconv"
	"time"
)

// CreateStory - создаёт историю и загружает файл в Google Drive
func CreateStory(w http.ResponseWriter, r *http.Request) {
	// Проверяем Google OAuth токен
	token, err := authentication.ValidateGoogleToken(r)
	if err != nil {
		http.Error(w, "Unauthorized or forbidden: "+err.Error(), http.StatusUnauthorized)
		return
	}

	// Читаем файл из запроса
	file, header, err := r.FormFile("file")
	if err != nil {
		http.Error(w, "Failed to read file", http.StatusBadRequest)
		return
	}
	defer file.Close()

	// Получаем пользователя из базы данных
	claims, err := authentication.ValidateToken(r)
	if err != nil {
		http.Error(w, "Unauthorized: "+err.Error(), http.StatusUnauthorized)
		return
	}

	var userRecord users.GoogleUser
	if err := config.DB.First(&userRecord, claims.UserID).Error; err != nil {
		http.Error(w, "User not found", http.StatusNotFound)
		return
	}

	// Укажите ID папки Google Drive, куда будут загружаться файлы
	folderId := "1c4YaW6Qd3c8PyFXV43qSVQzWgPLysAs2"

	// Загрузка файла в Google Drive
	fileID, webViewLink, err := uploadFileToDrive(file, header.Filename, token.AccessToken, folderId)
	if err != nil {
		http.Error(w, "Failed to upload file to Google Drive: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Сохраняем историю в базе данных
	newStory := story.Story{
		ContentURL:  webViewLink,
		DriveFileID: fileID,
		UserID:      claims.UserID,
		CreatedAt:   time.Now().UTC(),
		ExpireAt:    time.Now().UTC().Add(24 * time.Hour),
	}

	if err := config.DB.Create(&newStory).Error; err != nil {
		http.Error(w, "Failed to create story", http.StatusInternalServerError)
		return
	}

	// Успешный ответ
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"message":   "Story created successfully",
		"story_url": webViewLink,
		"story_id":  newStory.ID,
		"story":     newStory,
	})
}

// uploadFileToDrive - загружает файл в Google Drive с использованием Google OAuth токена
func uploadFileToDrive(file multipart.File, fileName, accessToken, folderId string) (string, string, error) {
	ctx := context.Background()

	// Создаём токен OAuth вручную
	token := &oauth2.Token{
		AccessToken: accessToken,
	}

	// Создаем источник токена
	tokenSource := authentication.GoogleOauthConfig.TokenSource(ctx, token)

	// Создаём Google Drive сервис
	service, err := drive.NewService(ctx, option.WithTokenSource(tokenSource))
	if err != nil {
		return "", "", fmt.Errorf("failed to create Drive service: %v", err)
	}

	// Подготавливаем метаданные файла
	driveFile := &drive.File{
		Name:    fileName,
		Parents: []string{folderId}, // Укажите ID папки
	}

	// Загружаем файл
	uploadedFile, err := service.Files.Create(driveFile).Media(file).Do()
	if err != nil {
		return "", "", fmt.Errorf("failed to upload file: %v", err)
	}

	// Возвращаем ID файла и ссылку на просмотр
	return uploadedFile.Id, uploadedFile.WebViewLink, nil
}

// Получение всех историй пользователя
func GetUserStories(w http.ResponseWriter, r *http.Request) {
	claims, err := authentication.ValidateToken(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}

	var stories []story.Story
	config.DB.Where("user_id = ? AND expire_at > ? AND is_archived = ?", claims.UserID, time.Now().UTC(), false).Find(&stories)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(stories)
}

// Просмотр одной истории
func ViewStory(w http.ResponseWriter, r *http.Request) {
	claims, err := authentication.ValidateToken(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}

	storyIDStr := r.URL.Query().Get("id")
	storyID, err := strconv.Atoi(storyIDStr)
	if err != nil {
		http.Error(w, "Invalid story ID", http.StatusBadRequest)
		return
	}

	var currentStory story.Story
	if result := config.DB.First(&currentStory, storyID); errors.Is(result.Error, gorm.ErrRecordNotFound) {
		http.Error(w, "Story not found", http.StatusNotFound)
		return
	}

	if currentStory.UserID != claims.UserID && currentStory.Privacy != "public" {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}

	// Атомарное обновление просмотров
	if err := config.DB.Model(&currentStory).UpdateColumn("views", gorm.Expr("views + ?", 1)).Error; err != nil {
		http.Error(w, "Failed to update views", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(currentStory)
}

// Архивирование истории
func ArchiveStory(w http.ResponseWriter, r *http.Request) {
	claims, err := authentication.ValidateToken(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}

	storyIDStr := r.URL.Query().Get("id")
	storyID, err := strconv.Atoi(storyIDStr)
	if err != nil {
		http.Error(w, "Invalid story ID", http.StatusBadRequest)
		return
	}

	var currentStory story.Story
	if result := config.DB.First(&currentStory, storyID); errors.Is(result.Error, gorm.ErrRecordNotFound) {
		http.Error(w, "Story not found", http.StatusNotFound)
		return
	}

	if currentStory.UserID != claims.UserID {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}

	currentStory.IsArchived = true
	if err := config.DB.Save(&currentStory).Error; err != nil {
		http.Error(w, "Failed to archive story", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(currentStory)
}

// UpdateStory - обновление информации об истории
func UpdateStory(w http.ResponseWriter, r *http.Request) {
	claims, err := authentication.ValidateToken(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}

	if r.Method != http.MethodPut {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		return
	}

	storyIDStr := r.URL.Query().Get("id")
	storyID, err := strconv.Atoi(storyIDStr)
	if err != nil {
		http.Error(w, "Invalid story ID", http.StatusBadRequest)
		return
	}

	var currentStory story.Story
	if result := config.DB.First(&currentStory, storyID); errors.Is(result.Error, gorm.ErrRecordNotFound) {
		http.Error(w, "Story not found", http.StatusNotFound)
		return
	}

	if currentStory.UserID != claims.UserID {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}

	if time.Now().After(currentStory.ExpireAt) {
		http.Error(w, "Cannot update an expired story", http.StatusBadRequest)
		return
	}

	var updatedStory story.Story
	if err := json.NewDecoder(r.Body).Decode(&updatedStory); err != nil {
		http.Error(w, "Invalid input", http.StatusBadRequest)
		return
	}

	// Обновляем только разрешенные поля
	if updatedStory.ContentURL != "" {
		currentStory.ContentURL = updatedStory.ContentURL
	}
	if updatedStory.Privacy != "" {
		currentStory.Privacy = updatedStory.Privacy
	}

	if err := config.DB.Save(&currentStory).Error; err != nil {
		http.Error(w, "Failed to update story", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(currentStory)
}

// Удаление истории
func DeleteStory(w http.ResponseWriter, r *http.Request) {
	claims, err := authentication.ValidateToken(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}

	storyIDStr := r.URL.Query().Get("id")
	storyID, err := strconv.Atoi(storyIDStr)
	if err != nil {
		http.Error(w, "Invalid story ID", http.StatusBadRequest)
		return
	}

	var currentStory story.Story
	if result := config.DB.First(&currentStory, storyID); errors.Is(result.Error, gorm.ErrRecordNotFound) {
		http.Error(w, "Story not found", http.StatusNotFound)
		return
	}

	if currentStory.UserID != claims.UserID {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}

	if err := config.DB.Delete(&currentStory).Error; err != nil {
		http.Error(w, "Failed to delete story", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

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
	"log"
	"mime/multipart"
	"net/http"
	"os"
	"strconv"
	"time"
)

// Google Drive credentials file
var serviceAccountFile = os.Getenv("DRIVE_JSON")

// CreateStory - создает историю и загружает файл в Google Drive
func CreateStory(w http.ResponseWriter, r *http.Request) {
	// Validate Google OAuth Token
	googleUser, err := authentication.ValidateGoogleToken(r)
	if err != nil {
		http.Error(w, "Unauthorized: "+err.Error(), http.StatusUnauthorized)
		return
	}

	// Parse multipart form
	file, header, err := r.FormFile("file")
	if err != nil {
		http.Error(w, "Failed to read file: "+err.Error(), http.StatusBadRequest)
		return
	}
	defer file.Close()

	// Folder ID in Google Drive
	folderID := os.Getenv("GOOGLE_DRIVE_FOLDER_ID")

	// Upload file to Google Drive
	fileID, webViewLink, err := uploadFileToGoogleDrive(file, header.Filename, googleUser.AccessToken, folderID)
	if err != nil {
		http.Error(w, "Failed to upload file to Google Drive: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Save story to the database
	newStory := story.Story{
		ContentURL:  webViewLink,
		DriveFileID: fileID,
		UserID:      googleUser.UserID, // Используем UserID из GoogleUser
		CreatedAt:   time.Now().UTC(),
		ExpireAt:    time.Now().UTC().Add(24 * time.Hour),
	}

	if err := config.DB.Create(&newStory).Error; err != nil {
		http.Error(w, "Failed to save story: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Respond with success
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"message":   "Story created successfully",
		"story_url": webViewLink,
		"story_id":  newStory.ID,
		"story":     newStory,
	})
}

// uploadFileToGoogleDrive - загружает файл в Google Drive
func uploadFileToGoogleDrive(file multipart.File, fileName string, accessToken string, folderID string) (string, string, error) {
	ctx := context.Background()

	// Создаем токен OAuth вручную
	token := &oauth2.Token{
		AccessToken: accessToken,
	}

	// Создаем источник токена
	tokenSource := authentication.GoogleOauthConfig.TokenSource(ctx, token)

	// Создаем Google Drive сервис
	service, err := drive.NewService(ctx, option.WithTokenSource(tokenSource))
	if err != nil {
		return "", "", fmt.Errorf("failed to create Drive service: %v", err)
	}

	// Подготавливаем метаданные файла
	driveFile := &drive.File{
		Name:    fileName,
		Parents: []string{folderID},
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
	// Проверяем Google OAuth токен
	googleUser, err := authentication.ValidateGoogleToken(r)
	if err != nil {
		http.Error(w, "Unauthorized: "+err.Error(), http.StatusUnauthorized)
		return
	}

	// Логируем ID пользователя для проверки
	log.Printf("Fetching stories for Google User ID: %d", googleUser.UserID)

	// Получаем все истории пользователя, без фильтрации по is_archived
	var stories []story.Story
	result := config.DB.Where("user_id = ? AND expire_at > ?", googleUser.UserID, time.Now().UTC()).Find(&stories)
	if result.Error != nil {
		log.Printf("Database query error: %v", result.Error)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	// Логируем количество найденных историй
	log.Printf("Stories found: %d", result.RowsAffected)

	// Отправляем результат клиенту
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(stories)
}

// Просмотр одной истории
func ViewStory(w http.ResponseWriter, r *http.Request) {
	// Проверяем Google OAuth токен
	googleUser, err := authentication.ValidateGoogleToken(r)
	if err != nil {
		http.Error(w, "Unauthorized: "+err.Error(), http.StatusUnauthorized)
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

	if currentStory.UserID != googleUser.UserID && currentStory.Privacy != "public" {
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
	// Проверяем Google OAuth токен
	googleUser, err := authentication.ValidateGoogleToken(r)
	if err != nil {
		http.Error(w, "Unauthorized: "+err.Error(), http.StatusUnauthorized)
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

	if currentStory.UserID != googleUser.UserID {
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

// Обновление информации об истории
func UpdateStory(w http.ResponseWriter, r *http.Request) {
	// Проверяем Google OAuth токен
	googleUser, err := authentication.ValidateGoogleToken(r)
	if err != nil {
		http.Error(w, "Unauthorized: "+err.Error(), http.StatusUnauthorized)
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

	if currentStory.UserID != googleUser.UserID {
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
	// Проверяем Google OAuth токен
	googleUser, err := authentication.ValidateGoogleToken(r)
	if err != nil {
		http.Error(w, "Unauthorized: "+err.Error(), http.StatusUnauthorized)
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

	if currentStory.UserID != googleUser.UserID {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}

	if err := config.DB.Delete(&currentStory).Error; err != nil {
		http.Error(w, "Failed to delete story", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

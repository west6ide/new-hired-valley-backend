package stories

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/drive/v3"
	"google.golang.org/api/option"
	"gorm.io/gorm"
	"hired-valley-backend/config"
	"hired-valley-backend/controllers/authentication"
	"hired-valley-backend/models/story"
	"hired-valley-backend/models/users"
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
	// Validate JWT Token and get claims
	claims, err := authentication.ValidateToken(r)
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

	// Get GoogleUser record
	var googleUser users.GoogleUser
	if err := config.DB.Where("user_id = ?", claims.UserID).First(&googleUser).Error; err != nil {
		http.Error(w, "Google account not linked or user not found", http.StatusForbidden)
		return
	}

	// Check if access token exists
	if googleUser.AccessToken == "" {
		http.Error(w, "No Google OAuth token found for the user", http.StatusForbidden)
		return
	}

	// Folder ID in Google Drive
	folderID := os.Getenv("GOOGLE_DRIVE_FOLDER_ID") // Укажите ID вашей папки

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
		UserID:      claims.UserID,
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

// getGoogleDriveClient - возвращает клиента Google Drive API
func getGoogleDriveClient() (*http.Client, error) {
	b, err := os.ReadFile(serviceAccountFile)
	if err != nil {
		return nil, fmt.Errorf("failed to read service account file: %v", err)
	}

	// Authenticate with the service account
	config, err := google.JWTConfigFromJSON(b, drive.DriveScope)
	if err != nil {
		return nil, fmt.Errorf("failed to create JWT config: %v", err)
	}

	return config.Client(context.Background()), nil
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

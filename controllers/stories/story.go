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
	"io"
	"net/http"
	"strconv"
	"time"
)

// GoogleDriveUploadWithOAuth - загружает файл в Google Drive и возвращает ссылки
func GoogleDriveUploadWithOAuth(file io.Reader, filename, mimeType, folderId, accessToken string) (string, string, error) {
	token := &oauth2.Token{
		AccessToken: accessToken,
	}
	client := authentication.GoogleOauthConfig.Client(context.Background(), token)

	// Инициализация службы Google Drive
	srv, err := drive.NewService(context.Background(), option.WithHTTPClient(client))
	if err != nil {
		return "", "", fmt.Errorf("unable to retrieve Drive client: %v", err)
	}

	// Создание метаданных файла
	driveFile := &drive.File{
		Name:     filename,
		MimeType: mimeType,
		Parents:  []string{folderId}, // ID папки
	}

	// Загрузка файла
	uploadedFile, err := srv.Files.Create(driveFile).Media(file).Do()
	if err != nil {
		return "", "", fmt.Errorf("failed to upload file: %v", err)
	}

	return uploadedFile.Id, uploadedFile.WebViewLink, nil
}

// CreateStory - создаёт историю и загружает файл в Google Drive
func CreateStory(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		return
	}

	claims, err := authentication.ValidateToken(r)
	if err != nil {
		http.Error(w, "Unauthorized: "+err.Error(), http.StatusUnauthorized)
		return
	}

	const maxUploadSize = 10 * 1024 * 1024 // 10 MB
	r.Body = http.MaxBytesReader(w, r.Body, maxUploadSize)

	if err := r.ParseMultipartForm(maxUploadSize); err != nil {
		http.Error(w, "File too large or invalid form data", http.StatusRequestEntityTooLarge)
		return
	}

	file, header, err := r.FormFile("file")
	if err != nil {
		http.Error(w, "File is required", http.StatusBadRequest)
		return
	}
	defer file.Close()

	// Проверка MIME-типа файла
	buffer := make([]byte, 512)
	if _, err := file.Read(buffer); err != nil {
		http.Error(w, "Failed to read file", http.StatusInternalServerError)
		return
	}
	fileType := http.DetectContentType(buffer)

	// Возврат курсора файла
	if _, err := file.Seek(0, io.SeekStart); err != nil {
		http.Error(w, "Failed to reset file pointer", http.StatusInternalServerError)
		return
	}

	// Получение токена пользователя из базы данных
	var userRecord users.User
	if err := config.DB.First(&userRecord, claims.UserID).Error; err != nil {
		http.Error(w, "User not found", http.StatusNotFound)
		return
	}

	if userRecord.AccessToken == "" {
		http.Error(w, "Google account is not linked", http.StatusForbidden)
		return
	}

	// Укажите папку Google Drive
	folderId := "1c4YaW6Qd3c8PyFXV43qSVQzWgPLysAs2" // Укажите ID папки в Google Drive

	// Загрузка файла в Google Drive через пользовательский OAuth
	fileID, webViewLink, err := GoogleDriveUploadWithOAuth(file, header.Filename, fileType, folderId, userRecord.AccessToken)
	if err != nil {
		http.Error(w, "Failed to upload file to Google Drive: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Сохранение истории в базе данных
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
	w.WriteHeader(http.StatusCreated)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"status":  "success",
		"message": "Story created successfully",
		"story":   newStory,
	})
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

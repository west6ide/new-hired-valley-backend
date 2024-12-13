package contentsControl

import (
	"context"
	"encoding/json"
	"fmt"
	"golang.org/x/oauth2"
	"google.golang.org/api/option"
	"google.golang.org/api/youtube/v3"
	"hired-valley-backend/config"
	"hired-valley-backend/controllers/authentication"
	"hired-valley-backend/models/content"
	"mime/multipart"
	"net/http"
	"strconv"
	"strings"
)

// UploadContent - загрузка контента на YouTube и сохранение записи в базе данных
func UploadContent(w http.ResponseWriter, r *http.Request) {
	// Проверка авторизации
	claims, err := authentication.ValidateGoogleToken(r)
	if err != nil {
		http.Error(w, "Unauthorized: "+err.Error(), http.StatusUnauthorized)
		return
	}

	// Получение текстовых данных из form-data
	title := r.FormValue("title")
	description := r.FormValue("description")
	category := r.FormValue("category")
	tagsStr := r.FormValue("tags") // Теги передаются через запятую, например: "tag1,tag2"

	if title == "" || description == "" || category == "" {
		http.Error(w, "Missing required fields (title, description, category)", http.StatusBadRequest)
		return
	}

	// Парсинг тегов в массив строк
	tags := strings.Split(tagsStr, ",")
	for i := range tags {
		tags[i] = strings.TrimSpace(tags[i])
	}

	// Получение файла видео
	file, header, err := r.FormFile("video")
	if err != nil {
		http.Error(w, "Video file is required", http.StatusBadRequest)
		return
	}
	defer file.Close()

	// Загрузка видео на YouTube
	videoID, err := uploadVideoToYouTube(file, header.Filename, claims.AccessToken, title, description)
	if err != nil {
		http.Error(w, "Failed to upload video to YouTube: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Проверяем, что videoID получен
	if videoID == "" {
		http.Error(w, "Failed to get video ID from YouTube", http.StatusInternalServerError)
		return
	}

	// Создание записи контента
	content := content.Content{
		Title:       title,
		Description: description,
		Category:    category,
		Tags:        tags,
		VideoLink:   fmt.Sprintf("https://www.youtube.com/watch?v=%s", videoID), // Убедитесь, что это поле не пустое
		YouTubeID:   videoID,
		AuthorID:    claims.UserID,
	}

	// Сохранение записи в базе данных
	if err := config.DB.Create(&content).Error; err != nil {
		http.Error(w, "Failed to save content: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Успешный ответ
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(content)
}

// uploadVideoToYouTube - загрузка видео на YouTube
func uploadVideoToYouTube(file multipart.File, fileName, accessToken, title, description string) (string, error) {
	ctx := context.Background()

	token := &oauth2.Token{AccessToken: accessToken}
	tokenSource := authentication.GoogleOauthConfig.TokenSource(ctx, token)

	service, err := youtube.NewService(ctx, option.WithTokenSource(tokenSource))
	if err != nil {
		return "", fmt.Errorf("failed to create YouTube service: %v", err)
	}

	video := &youtube.Video{
		Snippet: &youtube.VideoSnippet{
			Title:       title,
			Description: description,
			CategoryId:  "22", // По умолчанию "People & Blogs"
		},
		Status: &youtube.VideoStatus{
			PrivacyStatus: "unlisted", // Видео будет скрытым
		},
	}

	call := service.Videos.Insert([]string{"snippet", "status"}, video)
	response, err := call.Media(file).Do()
	if err != nil {
		return "", fmt.Errorf("failed to upload video: %v", err)
	}

	return response.Id, nil
}

// ListContent - получение списка контента
func ListContent(w http.ResponseWriter, r *http.Request) {
	category := r.URL.Query().Get("category")
	tags := r.URL.Query()["tags"]

	query := config.DB.Model(&content.Content{})
	if category != "" {
		query = query.Where("category = ?", category)
	}
	if len(tags) > 0 {
		query = query.Where("tags && ?", tags) // Используем массив тегов
	}

	var contents []content.Content
	if err := query.Find(&contents).Error; err != nil {
		http.Error(w, "Failed to fetch content", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(contents)
}

// GetContentByID - получение контента по ID
func GetContentByID(w http.ResponseWriter, r *http.Request) {
	idStr := r.URL.Query().Get("id")
	if idStr == "" {
		http.Error(w, "Content ID is required", http.StatusBadRequest)
		return
	}

	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "Invalid Content ID", http.StatusBadRequest)
		return
	}

	var content content.Content
	if err := config.DB.First(&content, id).Error; err != nil {
		http.Error(w, "Content not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(content)
}

// DeleteContent - удаление контента
func DeleteContent(w http.ResponseWriter, r *http.Request) {
	claims, err := authentication.ValidateToken(r)
	if err != nil {
		http.Error(w, "Unauthorized: "+err.Error(), http.StatusUnauthorized)
		return
	}

	idStr := r.URL.Query().Get("id")
	if idStr == "" {
		http.Error(w, "Content ID is required", http.StatusBadRequest)
		return
	}

	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "Invalid Content ID", http.StatusBadRequest)
		return
	}

	var content content.Content
	if err := config.DB.First(&content, id).Error; err != nil {
		http.Error(w, "Content not found", http.StatusNotFound)
		return
	}

	if content.AuthorID != claims.UserID {
		http.Error(w, "Permission denied", http.StatusForbidden)
		return
	}

	if err := config.DB.Delete(&content).Error; err != nil {
		http.Error(w, "Failed to delete content", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

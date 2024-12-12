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
	claims, err := authentication.ValidateGoogleToken(r)
	if err != nil {
		http.Error(w, "Unauthorized: "+err.Error(), http.StatusUnauthorized)
		return
	}

	// Декодируем JSON-запрос
	var input struct {
		Title       string   `json:"title"`
		Description string   `json:"description"`
		Category    string   `json:"category"`
		Tags        []string `json:"tags"`
	}

	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		http.Error(w, "Invalid input: "+err.Error(), http.StatusBadRequest)
		return
	}

	if input.Title == "" || input.Description == "" || input.Category == "" {
		http.Error(w, "Missing required fields", http.StatusBadRequest)
		return
	}

	// Получение файла из form-data
	file, header, err := r.FormFile("video")
	if err != nil {
		http.Error(w, "Video file is required", http.StatusBadRequest)
		return
	}
	defer file.Close()

	videoID, err := uploadVideoToYouTube(file, header.Filename, claims.AccessToken, input.Title, input.Description)
	if err != nil {
		http.Error(w, "Failed to upload video to YouTube: "+err.Error(), http.StatusInternalServerError)
		return
	}

	content := content.Content{
		Title:       input.Title,
		Description: input.Description,
		Category:    input.Category,
		Tags:        input.Tags,
		VideoLink:   fmt.Sprintf("https://www.youtube.com/watch?v=%s", videoID),
		YouTubeID:   videoID,
		AuthorID:    claims.UserID,
	}

	if err := config.DB.Create(&content).Error; err != nil {
		http.Error(w, "Failed to save content: "+err.Error(), http.StatusInternalServerError)
		return
	}

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
	tagsStr := r.URL.Query().Get("tags")

	query := config.DB.Model(&content.Content{})
	if category != "" {
		query = query.Where("category = ?", category)
	}
	if tagsStr != "" {
		tags := strings.Split(tagsStr, ",")
		query = query.Where("tags @> ?", tags)
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

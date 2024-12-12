package contentsControl

import (
	"encoding/json"
	"hired-valley-backend/config"
	"hired-valley-backend/controllers/authentication"
	"hired-valley-backend/models/content"
	"hired-valley-backend/models/users"
	"net/http"
	"strconv"
)

func CreateContent(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	claims, err := authentication.ValidateToken(r)
	if err != nil {
		http.Error(w, "Unauthorized: "+err.Error(), http.StatusUnauthorized)
		return
	}

	var content content.Content
	if err := json.NewDecoder(r.Body).Decode(&content); err != nil {
		http.Error(w, "Invalid input", http.StatusBadRequest)
		return
	}

	if len(content.Tags) == 0 || content.VideoURL == "" {
		http.Error(w, "Tags and Video URL are required", http.StatusBadRequest)
		return
	}

	content.UserID = claims.UserID

	if err := config.DB.Create(&content).Error; err != nil {
		http.Error(w, "Failed to create content", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(content)
}

func GetPersonalizedContent(w http.ResponseWriter, r *http.Request) {
	claims, err := authentication.ValidateToken(r)
	if err != nil {
		http.Error(w, "Unauthorized: "+err.Error(), http.StatusUnauthorized)
		return
	}

	// Получаем интересы пользователя
	var user users.User
	err = config.DB.Preload("Interests").First(&user, claims.UserID).Error
	if err != nil {
		http.Error(w, "User not found", http.StatusNotFound)
		return
	}

	interestTags := []string{}
	for _, interest := range user.Interests {
		interestTags = append(interestTags, interest.Name)
	}

	// Поиск контента по интересам
	var contentList []content.Content
	if err := config.DB.Where("tags && ?", interestTags).Find(&contentList).Error; err != nil {
		http.Error(w, "Failed to fetch content", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(contentList)
}

func DeleteContent(w http.ResponseWriter, r *http.Request) {
	claims, err := authentication.ValidateToken(r)
	if err != nil {
		http.Error(w, "Unauthorized: "+err.Error(), http.StatusUnauthorized)
		return
	}

	contentIDStr := r.URL.Query().Get("id")
	contentID, err := strconv.Atoi(contentIDStr)
	if err != nil {
		http.Error(w, "Invalid content ID", http.StatusBadRequest)
		return
	}

	var content content.Content
	if err := config.DB.First(&content, contentID).Error; err != nil {
		http.Error(w, "Content not found", http.StatusNotFound)
		return
	}

	if content.UserID != claims.UserID {
		http.Error(w, "Permission denied", http.StatusForbidden)
		return
	}

	if err := config.DB.Delete(&content).Error; err != nil {
		http.Error(w, "Failed to delete content", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

package recommendations

import (
	"encoding/json"
	"fmt"
	"hired-valley-backend/config"
	"hired-valley-backend/controllers/authentication"
	"hired-valley-backend/models/content"
	"hired-valley-backend/models/courses"
	"hired-valley-backend/models/users"
	"net/http"
	"os"
	"strings"
)

// PersonalizedRecommendationsHandler - обработчик для персонализированных рекомендаций
func PersonalizedRecommendationsHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Проверяем Google OAuth токен
	claims, err := authentication.ValidateGoogleToken(r)
	if err != nil {
		http.Error(w, "Unauthorized: "+err.Error(), http.StatusUnauthorized)
		return
	}

	// Получение данных пользователя
	var user users.User
	if err := config.DB.First(&user, claims.UserID).Error; err != nil {
		http.Error(w, "User not found", http.StatusNotFound)
		return
	}

	// Ищем курсы, соответствующие интересам пользователя
	var matchedCourses []courses.Course
	if err := config.DB.Where("tags && ?", user.Interests).Find(&matchedCourses).Error; err != nil {
		http.Error(w, "Failed to fetch courses", http.StatusInternalServerError)
		return
	}

	// Ищем контент, соответствующий интересам пользователя
	var matchedContent []content.Content
	if err := config.DB.Where("tags && ?", user.Interests).Find(&matchedContent).Error; err != nil {
		http.Error(w, "Failed to fetch content", http.StatusInternalServerError)
		return
	}

	// Подключаем AI для улучшения рекомендаций
	apiKey := os.Getenv("AIML_API_KEY")
	if apiKey == "" {
		http.Error(w, "AI API key is missing", http.StatusInternalServerError)
		return
	}

	aiRequestBody := map[string]interface{}{
		"profile": map[string]interface{}{
			"industry":  user.Industry,
			"skills":    user.Skills,
			"interests": user.Interests,
		},
		"existing_recommendations": map[string]interface{}{
			"courses": matchedCourses,
			"content": matchedContent,
		},
	}

	// Отправляем запрос в AI
	aiResponse, err := callAIMLAPI(apiKey, aiRequestBody)
	if err != nil {
		http.Error(w, "Failed to fetch AI recommendations: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Формируем итоговый ответ
	response := map[string]interface{}{
		"personalized_courses": matchedCourses,
		"personalized_content": matchedContent,
		"ai_suggestions":       aiResponse,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// callAIMLAPI - отправка запроса к AIML API
func callAIMLAPI(apiKey string, requestBody map[string]interface{}) (map[string]interface{}, error) {
	url := "https://api.aimlapi.com/recommendations"
	requestJSON, _ := json.Marshal(requestBody)

	req, err := http.NewRequest("POST", url, strings.NewReader(string(requestJSON)))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %v", err)
	}
	req.Header.Set("Authorization", "Bearer "+apiKey)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to make request: %v", err)
	}
	defer resp.Body.Close()

	var response map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, fmt.Errorf("failed to decode response: %v", err)
	}

	return response, nil
}

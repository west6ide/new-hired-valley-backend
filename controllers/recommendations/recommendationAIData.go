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

func PersonalizedRecommendationsHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Проверяем Google OAuth токен
	claims, err := authentication.ValidateToken(r)
	if err != nil {
		http.Error(w, "Unauthorized: "+err.Error(), http.StatusUnauthorized)
		return
	}

	// Получение данных пользователя с предзагрузкой навыков и интересов
	var user users.User
	if err := config.DB.Preload("Skills").Preload("Interests").First(&user, claims.UserID).Error; err != nil {
		http.Error(w, "User not found", http.StatusNotFound)
		return
	}

	// Преобразуем навыки и интересы в массив строк
	skills := convertSkillsToStrings(user.Skills)
	interests := convertInterestsToStrings(user.Interests)

	// Ищем курсы, соответствующие интересам пользователя
	var matchedCourses []courses.Course
	if err := config.DB.Where("tags && ?", interests).Find(&matchedCourses).Error; err != nil {
		http.Error(w, "Failed to fetch courses", http.StatusInternalServerError)
		return
	}

	// Ищем контент, соответствующий интересам пользователя
	var matchedContent []content.Content
	if err := config.DB.Where("tags && ?", interests).Find(&matchedContent).Error; err != nil {
		http.Error(w, "Failed to fetch content", http.StatusInternalServerError)
		return
	}

	// Подготовка запроса для AI
	apiKey := os.Getenv("AIML_API_KEY")
	if apiKey == "" {
		http.Error(w, "AI API key is missing", http.StatusInternalServerError)
		return
	}

	aiRequestBody := map[string]interface{}{
		"model": "gpt-4-turbo-2024-04-09",
		"messages": []map[string]string{
			{"role": "system", "content": "You are an AI assistant recommending courses and content based on user preferences."},
			{"role": "user", "content": fmt.Sprintf("The user works in %s, has skills in %s, and is interested in %s. Recommend personalized courses and content.", user.Industry, strings.Join(skills, ", "), strings.Join(interests, ", "))},
		},
		"max_tokens":  500, // Ограничиваем длину ответа
		"temperature": 0.7,
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

// Функция для вызова AIML API
func callAIMLAPI(apiKey string, requestBody map[string]interface{}) (map[string]interface{}, error) {
	url := "https://api.aimlapi.com/chat/completions"

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

	if resp.StatusCode != http.StatusOK {
		var errorResponse map[string]interface{}
		json.NewDecoder(resp.Body).Decode(&errorResponse)
		return nil, fmt.Errorf("API error: %v, Status: %d", errorResponse, resp.StatusCode)
	}

	var response map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, fmt.Errorf("failed to decode response: %v", err)
	}

	return response, nil
}

// Конвертация навыков в массив строк
func convertSkillsToStrings(skills []users.Skill) []string {
	var skillStrings []string
	for _, skill := range skills {
		skillStrings = append(skillStrings, skill.Name)
	}
	return skillStrings
}

// Конвертация интересов в массив строк
func convertInterestsToStrings(interests []users.Interest) []string {
	var interestStrings []string
	for _, interest := range interests {
		interestStrings = append(interestStrings, interest.Name)
	}
	return interestStrings
}

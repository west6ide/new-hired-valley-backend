package recommendations

import (
	"bytes"
	"encoding/json"
	"fmt"
	"hired-valley-backend/config"
	"hired-valley-backend/controllers/authentication"
	"hired-valley-backend/models/courses"
	"hired-valley-backend/models/users"
	"net/http"
	"os"
)

// RecommendationRequest представляет запрос к AI API
type RecommendationRequest struct {
	Prompt string `json:"prompt"`
}

// RecommendationResponse представляет ответ AI API
type RecommendationResponse struct {
	Recommendations []courses.Course `json:"recommendations"`
}

// GenerateRecommendationsHandler обрабатывает запросы на рекомендации
func GenerateRecommendationsHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Проверка токена пользователя
	claims, err := authentication.ValidateToken(r)
	if err != nil {
		http.Error(w, "Unauthorized: "+err.Error(), http.StatusUnauthorized)
		return
	}

	// Извлечение данных пользователя
	var user users.User
	err = config.DB.Preload("Skills").Preload("Interests").First(&user, claims.UserID).Error
	if err != nil {
		http.Error(w, "User not found", http.StatusNotFound)
		return
	}

	// Генерация prompt для AI
	prompt := generatePromptForCourses(user.Skills, user.Interests)

	// Получение рекомендаций от AI
	apiKey := os.Getenv("AIML_API_KEY")
	recommendations, err := fetchRecommendationsFromAI(apiKey, prompt)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to fetch recommendations: %v", err), http.StatusInternalServerError)
		return
	}

	// Возврат рекомендаций
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"recommendations": recommendations,
	})
}

// fetchRecommendationsFromAI вызывает AI API и возвращает рекомендации
func fetchRecommendationsFromAI(apiKey string, prompt string) ([]courses.Course, error) {
	requestBody, _ := json.Marshal(RecommendationRequest{Prompt: prompt})

	req, err := http.NewRequest("POST", "https://api.aimlapi.com/recommend", bytes.NewBuffer(requestBody))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", "Bearer "+apiKey)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("AI API error: %s", resp.Status)
	}

	var response RecommendationResponse
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, err
	}

	return response.Recommendations, nil
}

// generatePromptForCourses создает текстовый запрос для AI
func generatePromptForCourses(skills []users.Skill, interests []users.Interest) string {
	var skillNames, interestNames []string

	for _, skill := range skills {
		skillNames = append(skillNames, skill.Name)
	}

	for _, interest := range interests {
		interestNames = append(interestNames, interest.Name)
	}

	return fmt.Sprintf(
		"Generate course recommendations for a user with skills: %v and interests: %v",
		skillNames, interestNames,
	)
}

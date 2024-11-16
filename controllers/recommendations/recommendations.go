package recommendations

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/google/uuid"
	"gorm.io/gorm"
	"hired-valley-backend/config"
	"hired-valley-backend/controllers/authentication"
	"hired-valley-backend/models/recommend"
	"hired-valley-backend/models/users"
	"io/ioutil"
	"net/http"
	"strings"
)

const vanusAPIURL = "https://app.ai.vanus.ai/api/v1/0b973b9cd13f4635acae25277820b407" // Замените на ваш реальный API-ключ

// GetUserByID извлекает профиль пользователя по его userID
func GetUserByID(userID uint) (users.User, error) {
	var user users.User
	// Загрузка пользователя вместе с навыками и интересами
	result := config.DB.Preload("Skills").Preload("Interests").First(&user, userID)
	if errors.Is(result.Error, gorm.ErrRecordNotFound) {
		return users.User{}, errors.New("user not found")
	} else if result.Error != nil {
		return users.User{}, result.Error
	}
	return user, nil
}

// cleanJSONResponse обрабатывает сырые данные, чтобы исправить возможные ошибки в формате
func cleanJSONResponse(rawResponse string) string {
	// Убираем некорректные пробелы или символы после чисел
	cleaned := strings.ReplaceAll(rawResponse, "1. ", "1.0 ")
	return cleaned
}

// sendRequestToVanusAI отправляет запрос в Vanus AI
func sendRequestToVanusAI(req recommend.RecommendationRequest) (*recommend.RecommendationResponse, error) {
	reqBody, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("failed to create request body: %w", err)
	}

	fmt.Printf("Request to Vanus AI: %s\n", string(reqBody))

	httpReq, err := http.NewRequest("POST", vanusAPIURL, bytes.NewBuffer(reqBody))
	if err != nil {
		return nil, fmt.Errorf("failed to create HTTP request: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("x-vanusai-model", "gpt-3.5-turbo")
	httpReq.Header.Set("x-vanusai-sessionid", uuid.New().String())

	client := &http.Client{}
	resp, err := client.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("failed to send request to Vanus AI: %w", err)
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	fmt.Printf("Raw response from Vanus AI: %s\n", string(body))

	cleanedBody := cleanJSONResponse(string(body))

	var vanusResponse recommend.RecommendationResponse
	if err := json.Unmarshal([]byte(cleanedBody), &vanusResponse); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &vanusResponse, nil
}

// GetPersonalizedRecommendations создает запрос на основе данных пользователя
func GetPersonalizedRecommendations(user users.User) ([]recommend.Content, error) {
	skills := []string{}
	for _, skill := range user.Skills {
		skills = append(skills, skill.Name)
	}

	interests := []string{}
	for _, interest := range user.Interests {
		interests = append(interests, interest.Name)
	}

	fmt.Printf("Skills: %v, Interests: %v, Role: %s\n", skills, interests, user.Role)

	prompt := fmt.Sprintf(
		`Based on the user's professional profile, suggest relevant content from the knowledge base.
		Skills: %s, Interests: %s, Role: %s. Select the most relevant materials.`,
		strings.Join(skills, ", "),
		strings.Join(interests, ", "),
		user.Role,
	)

	request := recommend.RecommendationRequest{
		Prompt: prompt,
		Stream: false,
	}

	// Получаем ответ из Vanus AI
	response, err := sendRequestToVanusAI(request)
	if err != nil {
		return nil, err
	}

	fmt.Printf("Vanus AI Response: %+v\n", response)

	var recommendations []recommend.Content
	for _, content := range response.Content {
		recommendations = append(recommendations, recommend.Content{
			Title:       content.Title,
			Description: content.Description,
			ContentURL:  content.ContentURL,
			Category:    content.Category,
			SkillLevel:  content.SkillLevel,
			Tags:        content.Tags,
		})
	}

	return recommendations, nil
}

// GetRecommendationsHandler обрабатывает запрос для получения рекомендаций
func GetRecommendationsHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		return
	}

	claims, err := authentication.ValidateToken(r)
	if err != nil {
		http.Error(w, "Unauthorized: "+err.Error(), http.StatusUnauthorized)
		return
	}

	// Извлекаем профиль пользователя
	user, err := GetUserByID(claims.UserID)
	if err != nil {
		http.Error(w, "User not found", http.StatusNotFound)
		return
	}

	// Получаем рекомендации
	recs, err := GetPersonalizedRecommendations(user)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to get recommendations: %v", err), http.StatusInternalServerError)
		return
	}

	// Отправляем результат
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(recs)
}

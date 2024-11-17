package services

import (
	"bytes"
	"encoding/json"
	"fmt"
	"hired-valley-backend/models/recommend"
	"hired-valley-backend/models/users"
	"net/http"
	"os"
)

const openAIEndpoint = "https://api.openai.com/v1/completions"

// OpenAIRequest структура для отправки запроса к OpenAI
type OpenAIRequest struct {
	Model     string `json:"model"`
	Prompt    string `json:"prompt"`
	MaxTokens int    `json:"max_tokens"`
}

// OpenAIResponse структура для обработки ответа от OpenAI
type OpenAIResponse struct {
	Choices []struct {
		Text string `json:"text"`
	} `json:"choices"`
}

// GenerateRecommendations создает рекомендации на основе профиля пользователя
func GenerateRecommendations(user users.User) ([]recommend.Recommendation, error) {
	prompt := buildPrompt(user)

	reqBody := OpenAIRequest{
		Model:     "text-davinci-003",
		Prompt:    prompt,
		MaxTokens: 300,
	}

	payload, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to encode request: %w", err)
	}

	req, err := http.NewRequest("POST", openAIEndpoint, bytes.NewBuffer(payload))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+os.Getenv("OPENAI_API_KEY"))

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected response status: %s", resp.Status)
	}

	var openAIResp OpenAIResponse
	if err := json.NewDecoder(resp.Body).Decode(&openAIResp); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	recommendations := parseOpenAIResponse(openAIResp)
	return recommendations, nil
}

func buildPrompt(user users.User) string {
	return fmt.Sprintf(
		"Based on the user's profile: Role: %s, Skills: %v, Interests: %v, suggest up to 5 relevant learning materials.",
		user.Role, user.Skills, user.Interests,
	)
}

func parseOpenAIResponse(resp OpenAIResponse) []recommend.Recommendation {
	var recommendations []recommend.Recommendation
	for _, choice := range resp.Choices {
		recommendations = append(recommendations, recommend.Recommendation{
			Title:       "Recommended Content",
			Description: choice.Text,
			URL:         "https://example.com", // Можно заменить на динамическую ссылку
		})
	}
	return recommendations
}

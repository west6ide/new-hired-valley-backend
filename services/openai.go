package services

import (
	"bytes"
	"encoding/json"
	"fmt"
	"hired-valley-backend/models/users"
	"io/ioutil"
	"net/http"
	"strings"
)

const AIMLAPIEndpoint = "https://api.aimlapi.com/chat/completions"

// Структура запроса к AI/ML API
type CompletionRequest struct {
	Model     string `json:"model"`
	Prompt    string `json:"prompt"`
	MaxTokens int    `json:"max_tokens"`
}

// Структура ответа от AI/ML API
type CompletionResponse struct {
	Text string `json:"text"`
}

// Генерация универсального prompt на основе навыков и интересов
func GeneratePrompt(skills []users.Skill, interests []users.Interest) string {
	skillNames := []string{}
	for _, skill := range skills {
		skillNames = append(skillNames, skill.Name)
	}

	interestNames := []string{}
	for _, interest := range interests {
		interestNames = append(interestNames, interest.Name)
	}

	return fmt.Sprintf(`
You are a professional career assistant. Based on the user's skills and interests, suggest at least 5 useful resources, courses, or strategies to help the user grow professionally.

Skills: %s
Interests: %s

Provide the recommendations in a numbered list.
`, strings.Join(skillNames, ", "), strings.Join(interestNames, ", "))
}

// Взаимодействие с AI/ML API
func GenerateRecommendations(apiKey, prompt string) (string, error) {
	requestData := CompletionRequest{
		Model:     "gpt-4",
		Prompt:    prompt,
		MaxTokens: 200,
	}

	jsonData, err := json.Marshal(requestData)
	if err != nil {
		return "", fmt.Errorf("failed to serialize request: %w", err)
	}

	req, err := http.NewRequest("POST", AIMLAPIEndpoint, bytes.NewBuffer(jsonData))
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	// Установка заголовков
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+apiKey)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	// Проверка статуса ответа
	if resp.StatusCode != http.StatusOK {
		body, _ := ioutil.ReadAll(resp.Body)
		return "", fmt.Errorf("API returned status %d: %s", resp.StatusCode, string(body))
	}

	// Декодирование ответа
	var completionResp CompletionResponse
	err = json.NewDecoder(resp.Body).Decode(&completionResp)
	if err != nil {
		return "", fmt.Errorf("failed to decode API response: %w", err)
	}

	return completionResp.Text, nil
}

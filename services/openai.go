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

type ChatCompletionRequest struct {
	Model     string                  `json:"model"`
	Messages  []ChatCompletionMessage `json:"messages"`
	MaxTokens int                     `json:"max_tokens"`
}

// Структура сообщения в запросе
type ChatCompletionMessage struct {
	Role    string `json:"role"`    // "system", "user", или "assistant"
	Content string `json:"content"` // Текст сообщения
}

// Структура ответа от AI/ML API
type ChatCompletionResponse struct {
	Choices []struct {
		Message ChatCompletionMessage `json:"message"`
	} `json:"choices"`
}

// Генерация универсального prompt
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
Based on the user's skills and interests, suggest at least 5 resources or strategies to help their career growth.

Skills: %s
Interests: %s

Provide the recommendations in a numbered list.
`, strings.Join(skillNames, ", "), strings.Join(interestNames, ", "))
}

// Взаимодействие с AI/ML API
func GenerateRecommendations(apiKey, prompt string) (string, error) {
	// Формируем массив сообщений для запроса
	requestData := ChatCompletionRequest{
		Model: "gpt-4",
		Messages: []ChatCompletionMessage{
			{Role: "system", Content: "You are a career recommendation assistant."},
			{Role: "user", Content: prompt},
		},
		MaxTokens: 200,
	}

	// Сериализуем запрос в JSON
	jsonData, err := json.Marshal(requestData)
	if err != nil {
		return "", fmt.Errorf("failed to serialize request: %w", err)
	}

	// Создаём HTTP-запрос
	req, err := http.NewRequest("POST", AIMLAPIEndpoint, bytes.NewBuffer(jsonData))
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	// Устанавливаем заголовки
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+apiKey)

	// Отправляем запрос
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	// Проверяем статус ответа
	if resp.StatusCode != http.StatusOK {
		body, _ := ioutil.ReadAll(resp.Body)
		return "", fmt.Errorf("API returned status %d: %s", resp.StatusCode, string(body))
	}

	// Распаковываем ответ
	var completionResp ChatCompletionResponse
	err = json.NewDecoder(resp.Body).Decode(&completionResp)
	if err != nil {
		return "", fmt.Errorf("failed to decode API response: %w", err)
	}

	// Возвращаем текст первой рекомендации
	if len(completionResp.Choices) > 0 {
		return completionResp.Choices[0].Message.Content, nil
	}

	return "", fmt.Errorf("no recommendations returned by the API")
}

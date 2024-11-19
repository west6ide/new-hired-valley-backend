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

// Структура сообщения
type ChatCompletionMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// Структура ответа
type ChatCompletionResponse struct {
	Choices []struct {
		Message ChatCompletionMessage `json:"message"`
	} `json:"choices"`
}

// Генерация укороченного prompt для бесплатного тарифа
func GeneratePrompt(skills []users.Skill, interests []users.Interest) string {
	skillNames := []string{}
	for _, skill := range skills {
		skillNames = append(skillNames, skill.Name)
	}

	interestNames := []string{}
	for _, interest := range interests {
		interestNames = append(interestNames, interest.Name)
	}

	// Укороченный prompt
	return fmt.Sprintf("Skills: %s. Interests: %s. Recommend 5 career resources.",
		strings.Join(skillNames, ", "), strings.Join(interestNames, ", "))
}

// Проверка длины массива сообщений
func checkMessagesLength(messages []ChatCompletionMessage) error {
	totalLength := 0
	for _, message := range messages {
		totalLength += len(message.Content)
	}

	if totalLength > 256 {
		return fmt.Errorf("total length of 'messages' exceeds 256 characters: %d", totalLength)
	}
	return nil
}

// Взаимодействие с AI/ML API
func GenerateRecommendations(apiKey, prompt string) (string, error) {
	// Формируем массив сообщений
	messages := []ChatCompletionMessage{
		{Role: "system", Content: "You are a career assistant."},
		{Role: "user", Content: prompt},
	}

	// Проверяем длину сообщений
	if err := checkMessagesLength(messages); err != nil {
		return "", fmt.Errorf("messages array too long: %w", err)
	}

	// Подготовка данных запроса
	requestData := ChatCompletionRequest{
		Model:     "gpt-4",
		Messages:  messages,
		MaxTokens: 100,
	}

	jsonData, err := json.Marshal(requestData)
	if err != nil {
		return "", fmt.Errorf("failed to serialize request: %w", err)
	}

	req, err := http.NewRequest("POST", AIMLAPIEndpoint, bytes.NewBuffer(jsonData))
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	// Устанавливаем заголовки
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+apiKey)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	// Читаем тело ответа
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response body: %w", err)
	}

	// Проверяем статус ответа
	if resp.StatusCode != http.StatusCreated {
		return "", fmt.Errorf("API returned status %d: %s", resp.StatusCode, string(body))
	}

	// Распаковываем успешный ответ
	var completionResp ChatCompletionResponse
	err = json.Unmarshal(body, &completionResp)
	if err != nil {
		return "", fmt.Errorf("failed to decode API response: %w", err)
	}

	// Проверяем, есть ли в ответе рекомендации
	if len(completionResp.Choices) > 0 {
		return completionResp.Choices[0].Message.Content, nil
	}

	return "", fmt.Errorf("no recommendations returned by the API")
}

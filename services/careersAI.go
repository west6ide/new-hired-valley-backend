package services

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
)

const AIMLAPIEndpointToCareer = "https://api.aimlapi.com/chat/completions"

// Структура сообщения для CareerCompletion
type CareerCompletionMessageCareer struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// Структура запроса для CareerCompletion
type CareerCompletionRequestCareer struct {
	Model     string                          `json:"model"`
	Messages  []CareerCompletionMessageCareer `json:"messages"`
	MaxTokens int                             `json:"max_tokens"`
}

// Структура ответа от CareerCompletion
type CareerCompletionResponseCareer struct {
	Choices []struct {
		Message CareerCompletionMessageCareer `json:"message"`
	} `json:"choices"`
}

// Генерация карьерного плана через отдельный API
func GenerateCareerPlan(apiKey string, shortTermGoals, longTermGoals string) (string, error) {
	// Формируем массив сообщений
	messages := []CareerCompletionMessageCareer{
		{Role: "system", Content: "You are a career strategy assistant."},
		{Role: "user", Content: fmt.Sprintf("Short-term goals: %s", shortTermGoals)},
		{Role: "user", Content: fmt.Sprintf("Long-term goals: %s", longTermGoals)},
		{Role: "user", Content: "Create a step-by-step career strategy with recommended skills, resources, and mentors."},
	}

	// Проверяем длину сообщений
	totalLength := 0
	for _, msg := range messages {
		totalLength += len(msg.Content)
	}
	if totalLength > 256 {
		return "", fmt.Errorf("messages array too long: total length exceeds 256 characters: %d", totalLength)
	}

	// Создаём запрос
	requestData := CareerCompletionRequestCareer{
		Model:     "gpt-4",
		Messages:  messages,
		MaxTokens: 150,
	}

	jsonData, err := json.Marshal(requestData)
	if err != nil {
		return "", fmt.Errorf("failed to serialize request: %w", err)
	}

	req, err := http.NewRequest("POST", AIMLAPIEndpointToCareer, bytes.NewBuffer(jsonData))
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+apiKey)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		body, _ := ioutil.ReadAll(resp.Body)
		return "", fmt.Errorf("API returned status %d: %s", resp.StatusCode, string(body))
	}

	var completionResp CareerCompletionResponseCareer
	err = json.NewDecoder(resp.Body).Decode(&completionResp)
	if err != nil {
		return "", fmt.Errorf("failed to decode API response: %w", err)
	}

	if len(completionResp.Choices) > 0 {
		return completionResp.Choices[0].Message.Content, nil
	}

	return "", fmt.Errorf("no career strategy returned by the API")
}

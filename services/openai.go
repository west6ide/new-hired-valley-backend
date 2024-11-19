package services

import (
	"context"
	"fmt"
	"github.com/sashabaranov/go-openai"
	"gorm.io/gorm"
	"hired-valley-backend/models/users"
	"strings"
)

func GenerateAIRecommendationsForUser(db *gorm.DB, apiKey string, userID uint) ([]string, error) {
	// Загружаем пользователя и его данные из базы
	var user users.User
	if err := db.Preload("Skills").Preload("Interests").First(&user, userID).Error; err != nil {
		return nil, fmt.Errorf("failed to find user: %w", err)
	}

	// Инициализируем OpenAI клиент
	client := openai.NewClient(apiKey)

	// Формируем запрос (prompt) для OpenAI
	prompt := createPrompt(user)

	// Отправляем запрос в OpenAI
	resp, err := client.CreateCompletion(context.Background(), openai.CompletionRequest{
		Model:       openai.GPT4,
		Prompt:      prompt,
		MaxTokens:   200,
		Temperature: 0.7,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get AI recommendations: %w", err)
	}

	return parseAIResponse(resp.Choices[0].Text), nil
}

func createPrompt(user users.User) string {
	return fmt.Sprintf(`
	Based on the following skills and interests, suggest at least 5 useful courses or resources to help the user improve:
	Skills: %s
	Interests: %s
	Provide the recommendations in a list format.
	`, skillsToString(user.Skills), interestsToString(user.Interests))
}

func parseAIResponse(response string) []string {
	lines := strings.Split(response, "\n")
	var recommendations []string
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed != "" {
			recommendations = append(recommendations, trimmed)
		}
	}
	return recommendations
}

func skillsToString(skills []users.Skill) string {
	var names []string
	for _, skill := range skills {
		names = append(names, skill.Name)
	}
	return strings.Join(names, ", ")
}

func interestsToString(interests []users.Interest) string {
	var names []string
	for _, interest := range interests {
		names = append(names, interest.Name)
	}
	return strings.Join(names, ", ")
}

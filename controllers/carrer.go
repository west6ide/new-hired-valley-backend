// controllers/recommendations.go
package controllers

import (
	"encoding/json"
	"fmt"
	"log"
	"math"
	"net/http"
	"os"
	"sort"
	"strings"

	"github.com/cohere-ai/cohere-go"
	"hired-valley-backend/models/users"
)

// Структура для хранения рекомендованного контента
type RecommendedContent struct {
	Content    string
	Similarity float64
}

// Инициализация клиента Cohere
var cohereClient *cohere.Client

func InitCohereClient() {
	var err error
	cohereClient, err = cohere.CreateClient(os.Getenv("COHERE_API_KEY"))
	if err != nil {
		log.Fatalf("Ошибка создания клиента Cohere: %v", err)
	}
}

// Обработчик для получения персонализированных рекомендаций
func GetRecommendationsHandler(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value("userID") // Подразумевается, что userID извлекается из контекста
	user, err := users.GetUserByID(userID)
	if err != nil {
		http.Error(w, "Пользователь не найден", http.StatusNotFound)
		return
	}

	userProfileDescription := createUserProfileDescription(*user) // Разыменовываем указатель

	userEmbedding, err := generateUserProfileEmbedding(userProfileDescription)
	if err != nil {
		http.Error(w, "Ошибка генерации профиля", http.StatusInternalServerError)
		return
	}

	// Пример контента для рекомендаций
	contentDescriptions := map[string]string{
		"Golang Advanced Course":   "Advanced course on Golang for cloud-based applications.",
		"Career Development Guide": "Guide on career development strategies for tech professionals.",
		"Mentorship for Engineers": "Mentorship program for engineering leadership.",
		"Cloud Services Tutorial":  "Tutorial on deploying applications to the cloud.",
		"Tech Networking Tips":     "Tips on networking for software engineers.",
	}

	contentEmbeddings := make(map[string][]float64)
	for title, description := range contentDescriptions {
		embedding, err := generateContentEmbedding(description)
		if err != nil {
			log.Printf("Ошибка генерации вектора для %s: %v", title, err)
			continue
		}
		contentEmbeddings[title] = embedding
	}

	recommendations := getPersonalizedContent(userEmbedding, contentEmbeddings)

	// Отправка ответа с рекомендациями
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(recommendations)
}

// Создание текстового описания профиля пользователя
func createUserProfileDescription(user users.User) string {
	skills := make([]string, len(user.Skills))
	for i, skill := range user.Skills {
		skills[i] = skill.Name
	}

	interests := make([]string, len(user.Interests))
	for i, interest := range user.Interests {
		interests[i] = interest.Name
	}

	return fmt.Sprintf("Position: %s, City: %s, Skills: %s, Interests: %s, Preferences: %s",
		user.Position, user.City, strings.Join(skills, ", "), strings.Join(interests, ", "), user.ContentPreferences)
}

// Генерация вектора профиля пользователя
func generateUserProfileEmbedding(profileDescription string) ([]float64, error) {
	response, err := cohereClient.Embed(cohere.EmbedOptions{
		Texts: []string{profileDescription},
		Model: "embed-english-v2.0",
	})
	if err != nil {
		return nil, err
	}
	return response.Embeddings[0], nil
}

// Генерация вектора для контента
func generateContentEmbedding(contentDescription string) ([]float64, error) {
	response, err := cohereClient.Embed(cohere.EmbedOptions{
		Texts: []string{contentDescription},
		Model: "embed-english-v2.0",
	})
	if err != nil {
		return nil, err
	}
	return response.Embeddings[0], nil
}

// Косинусное сходство
func calculateCosineSimilarity(vecA, vecB []float64) float64 {
	var dotProduct, magnitudeA, magnitudeB float64
	for i := range vecA {
		dotProduct += vecA[i] * vecB[i]
		magnitudeA += vecA[i] * vecA[i]
		magnitudeB += vecB[i] * vecB[i]
	}
	return dotProduct / (math.Sqrt(magnitudeA) * math.Sqrt(magnitudeB))
}

// Получение персонализированного контента
func getPersonalizedContent(userEmbedding []float64, contentEmbeddings map[string][]float64) []RecommendedContent {
	var recommendations []RecommendedContent
	for content, embedding := range contentEmbeddings {
		similarity := calculateCosineSimilarity(userEmbedding, embedding)
		recommendations = append(recommendations, RecommendedContent{Content: content, Similarity: similarity})
	}
	sort.Slice(recommendations, func(i, j int) bool {
		return recommendations[i].Similarity > recommendations[j].Similarity
	})
	return recommendations[:5] // Топ-5 рекомендаций
}

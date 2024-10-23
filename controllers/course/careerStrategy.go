package course

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
)

// Структура запроса для AI генерации карьерной стратегии
type CareerStrategyRequest struct {
	Goals  string `json:"goals"`
	Income int    `json:"income"`
}

// Структура ответа от AI
type CareerStrategyResponse struct {
	Plan    string   `json:"plan"`
	Skills  []string `json:"skills"`
	Courses []string `json:"courses"`
	Mentors []string `json:"mentors"`
}

// Вызов OpenAI API для генерации карьерной стратегии
func callOpenAICareerStrategy(goals string, income int) (string, error) {
	apiKey := os.Getenv("OPENAI_API_KEY") // Ensure that the OpenAI API key is set in the environment variables
	if apiKey == "" {
		return "", fmt.Errorf("OpenAI API key is not set")
	}

	url := "https://api.openai.com/v1/completions"
	payload := fmt.Sprintf(`{
		"model": "text-davinci-003",
		"prompt": "Generate a career strategy for a user whose goal is %s and wants to earn %d USD annually.",
		"max_tokens": 500
	}`, goals, income)

	req, err := http.NewRequest("POST", url, bytes.NewBuffer([]byte(payload)))
	if err != nil {
		return "", err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+apiKey)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	return string(body), nil
}

// AI генерация карьерной стратегии с OpenAI
func GenerateCareerStrategy(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Извлекаем и декодируем запрос
	var request CareerStrategyRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Вызов функции OpenAI для генерации плана
	plan, err := callOpenAICareerStrategy(request.Goals, request.Income)
	if err != nil {
		http.Error(w, "Failed to generate career strategy", http.StatusInternalServerError)
		return
	}

	// Заглушки для примера (должно быть также интегрировано с реальными рекомендациями AI)
	skills := []string{"Skill 1", "Skill 2", "Skill 3"}
	courses := []string{"Course A", "Course B"}
	mentors := []string{"Mentor X", "Mentor Y"}

	// Формируем ответ
	response := CareerStrategyResponse{
		Plan:    plan,
		Skills:  skills,
		Courses: courses,
		Mentors: mentors,
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

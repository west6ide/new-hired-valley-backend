package recommendations

import (
	"encoding/json"
	"fmt"
	"hired-valley-backend/config"
	"hired-valley-backend/controllers/authentication"
	"hired-valley-backend/models/content"
	"hired-valley-backend/models/courses"
	"hired-valley-backend/models/users"
	"net/http"
	"os"
	"strings"

	"github.com/lib/pq"
)

// PersonalizedRecommendationsHandler - обработчик для персонализированных рекомендаций
func PersonalizedRecommendationsHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Проверяем токен
	claims, err := authentication.ValidateToken(r)
	if err != nil {
		http.Error(w, "Unauthorized: "+err.Error(), http.StatusUnauthorized)
		return
	}

	// Получаем данные пользователя с предзагрузкой навыков и интересов
	var user users.User
	if err := config.DB.Preload("Skills").Preload("Interests").First(&user, claims.UserID).Error; err != nil {
		http.Error(w, "User not found", http.StatusNotFound)
		return
	}

	// Конвертируем навыки и интересы в массив строк
	skills := extractSkillNames(user.Skills)
	interests := extractInterestNames(user.Interests)

	// Получаем курсы, контент и менторов из базы данных
	matchedCourses, matchedContent, matchedMentors, err := fetchDataFromDatabase(interests, skills)
	if err != nil {
		http.Error(w, "Failed to fetch data: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Подготовка данных для отправки в AI API
	apiKey := os.Getenv("AIML_API_KEY")
	if apiKey == "" {
		http.Error(w, "AI API key is missing", http.StatusInternalServerError)
		return
	}

	aiRequestBody := prepareAIRequest(user, matchedCourses, matchedContent, matchedMentors, skills, interests)
	if err := validateRequestSize(aiRequestBody); err != nil {
		http.Error(w, "Request size exceeds limit: "+err.Error(), http.StatusBadRequest)
		return
	}

	aiResponse, err := callAIMLAPI(apiKey, aiRequestBody)
	if err != nil {
		http.Error(w, "Failed to fetch AI recommendations: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Формируем итоговый ответ
	response := map[string]interface{}{
		"personalized_courses": matchedCourses,
		"personalized_content": matchedContent,
		"personalized_mentors": matchedMentors,
		"ai_suggestions":       aiResponse,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// fetchDataFromDatabase - выборка данных из базы на основе интересов и навыков
func fetchDataFromDatabase(interests, skills []string) ([]courses.Course, []content.Content, []users.User, error) {
	var matchedCourses []courses.Course
	if err := config.DB.Where("tags && ?", pq.Array(interests)).Find(&matchedCourses).Error; err != nil {
		return nil, nil, nil, fmt.Errorf("failed to fetch courses: %v", err)
	}

	var matchedContent []content.Content
	if err := config.DB.Where("tags && ?", pq.Array(interests)).Find(&matchedContent).Error; err != nil {
		return nil, nil, nil, fmt.Errorf("failed to fetch content: %v", err)
	}

	var matchedMentors []users.User
	if err := config.DB.
		Joins("JOIN user_skills ON users.id = user_skills.user_id").
		Joins("JOIN skills ON skills.id = user_skills.skill_id").
		Where("skills.name IN ?", skills).
		Where("users.role = ?", "mentor").
		Group("users.id").
		Find(&matchedMentors).Error; err != nil {
		return nil, nil, nil, fmt.Errorf("failed to fetch mentors: %v", err)
	}

	return matchedCourses, matchedContent, matchedMentors, nil
}

// prepareAIRequest - подготовка тела запроса для AI API
func prepareAIRequest(user users.User, courses []courses.Course, content []content.Content, mentors []users.User, skills, interests []string) map[string]interface{} {
	coursesList := summarizeTitles(courses)
	contentList := summarizeTitles(content)
	mentorsList := summarizeNames(mentors)

	return map[string]interface{}{
		"model": "gpt-4-turbo-2024-04-09",
		"messages": []map[string]string{
			{"role": "system", "content": "You are an AI assistant specializing in personalized recommendations."},
			{"role": "user", "content": fmt.Sprintf(
				"The user works in %s, has skills in %s, and is interested in %s. Based on this, recommend relevant courses, content, and mentors.",
				user.Industry, strings.Join(skills, ", "), strings.Join(interests, ", "),
			)},
			{"role": "user", "content": fmt.Sprintf("Relevant Courses: %s.", coursesList)},
			{"role": "user", "content": fmt.Sprintf("Relevant Content: %s.", contentList)},
			{"role": "user", "content": fmt.Sprintf("Relevant Mentors: %s.", mentorsList)},
		},
		"max_tokens": 500,
	}
}

// callAIMLAPI - отправка запроса к AI API
func callAIMLAPI(apiKey string, requestBody map[string]interface{}) (map[string]interface{}, error) {
	url := "https://api.aimlapi.com/chat/completions"
	requestJSON, _ := json.Marshal(requestBody)

	req, err := http.NewRequest("POST", url, strings.NewReader(string(requestJSON)))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %v", err)
	}
	req.Header.Set("Authorization", "Bearer "+apiKey)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to make request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		var errorResponse map[string]interface{}
		json.NewDecoder(resp.Body).Decode(&errorResponse)
		return nil, fmt.Errorf("API error: %v, Status: %d", errorResponse, resp.StatusCode)
	}

	var response map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, fmt.Errorf("failed to decode response: %v", err)
	}

	return response, nil
}

// validateRequestSize - проверка размера запроса
func validateRequestSize(request map[string]interface{}) error {
	messages, ok := request["messages"].([]map[string]string)
	if !ok {
		return fmt.Errorf("invalid messages format")
	}

	totalLength := 0
	for _, msg := range messages {
		totalLength += len(msg["content"])
		if totalLength > 256 {
			return fmt.Errorf("messages array exceeds 256 characters")
		}
	}
	return nil
}

// Вспомогательные функции для извлечения данных
func summarizeTitles(items interface{}) string {
	switch v := items.(type) {
	case []courses.Course:
		var titles []string
		for i, item := range v {
			titles = append(titles, item.Title)
			if i >= 2 { // Ограничиваем вывод 3 элементами
				break
			}
		}
		return strings.Join(titles, ", ")
	case []content.Content:
		var titles []string
		for i, item := range v {
			titles = append(titles, item.Title)
			if i >= 2 { // Ограничиваем вывод 3 элементами
				break
			}
		}
		return strings.Join(titles, ", ")
	default:
		return ""
	}
}

func summarizeNames(mentors []users.User) string {
	var names []string
	for i, mentor := range mentors {
		names = append(names, mentor.Name)
		if i >= 2 { // Ограничиваем вывод 3 элементами
			break
		}
	}
	return strings.Join(names, ", ")
}

func extractSkillNames(skills []users.Skill) []string {
	var skillStrings []string
	for _, skill := range skills {
		skillStrings = append(skillStrings, skill.Name)
	}
	return skillStrings
}

func extractInterestNames(interests []users.Interest) []string {
	var interestStrings []string
	for _, interest := range interests {
		interestStrings = append(interestStrings, interest.Name)
	}
	return interestStrings
}

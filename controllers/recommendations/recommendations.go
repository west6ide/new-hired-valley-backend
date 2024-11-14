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
)

const vanusAPIURL = "https://app.ai.vanus.ai/api/v1/0b973b9cd13f4635acae25277820b407" // Замените на ваш реальный API-ключ

// GetUserByID извлекает профиль пользователя по его userID и загружает связанные навыки и интересы
func GetUserByID(claims *authentication.Claims) (users.User, error) {
	var user users.User
	result := config.DB.Preload("Skills").Preload("Interests").Preload("Stories").First(&user, claims.UserID)

	if errors.Is(result.Error, gorm.ErrRecordNotFound) {
		return users.User{}, errors.New("user not found")
	} else if result.Error != nil {
		return users.User{}, result.Error
	}

	return user, nil
}

// GetPersonalizedRecommendations отправляет запрос в Vanus AI для получения рекомендаций
func GetPersonalizedRecommendations(user users.User) ([]recommend.Content, error) {
	sessionID := uuid.New().String()

	// Формируем промпт на основе профиля пользователя
	prompt := fmt.Sprintf(`Based on the user's professional profile, suggest relevant content from the knowledge base. Please consider the following aspects:

1. **User Skills**: %s — If possible, offer materials that help deepen existing skills or develop related areas.
   
2. **Interests**: %s — select content that matches the user's interests and further develops knowledge in the selected topics.

3. **Professional role or position**: %s — provide materials that will help the user improve skills important for his current role, as well as content that will help in career growth.

4. **Difficulty level**: Select materials that match the user's level of training. If the profile lists skills for beginners, recommend entry-level materials. If the user has experience, suggest more advanced topics and courses to expand their knowledge.

5. **Interaction History**: Avoid re-offering materials that the user has already viewed or studied, except for updated versions or materials with an in-depth level of study.

### Examples of materials that can be recommended:
   - Articles, training courses, practical guides
   - Video and audio materials
   - Research and case studies suitable for in-depth study of topics

### Restrictions:
- Avoid irrelevant content, especially from areas that are not related to the specified skills or interests.
   - If the user is interested in related topics, add content that expands their knowledge in this area.
   
Please select and submit up to 5 of the most appropriate recommendations from the knowledge base, taking into account the above parameters and providing the user with a brief description and a link to each material.`,
		user.Skills, user.Interests, user.Role)

	fmt.Printf("Generated prompt: %s\n", prompt) // Логируем для отладки

	requestBody, err := json.Marshal(map[string]interface{}{
		"prompt": prompt,
		"stream": false,
	})
	if err != nil {
		return nil, fmt.Errorf("ошибка при создании тела запроса: %w", err)
	}

	req, err := http.NewRequest("POST", vanusAPIURL, bytes.NewBuffer(requestBody))
	if err != nil {
		return nil, fmt.Errorf("ошибка при создании запроса: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("x-vanusai-model", "gpt-4")
	req.Header.Set("x-vanusai-sessionid", sessionID)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("ошибка при отправке запроса: %w", err)
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("ошибка при чтении ответа: %w", err)
	}

	fmt.Printf("Response from Vanus AI: %s\n", string(body)) // Логируем ответ от Vanus AI

	var recommendations []recommend.Content
	if err := json.Unmarshal(body, &recommendations); err != nil {
		return nil, fmt.Errorf("ошибка при декодировании ответа: %w", err)
	}

	fmt.Printf("Parsed recommendations: %+v\n", recommendations) // Логируем полученные рекомендации
	return recommendations, nil
}

// GetRecommendationsHandler обрабатывает запрос для получения рекомендаций
func GetRecommendationsHandler(w http.ResponseWriter, r *http.Request) {
	userID, err := authentication.ValidateToken(r)
	if err != nil {
		http.Error(w, "Unauthorized: "+err.Error(), http.StatusUnauthorized)
		return
	}

	fmt.Printf("User ID from token: %d\n", userID.UserID) // Логирование для отладки

	user, err := GetUserByID(userID)
	if err != nil {
		http.Error(w, "User not found", http.StatusNotFound)
		return
	}

	recs, err := GetPersonalizedRecommendations(user)
	if err != nil {
		http.Error(w, "Failed to get recommendations", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(recs)
}

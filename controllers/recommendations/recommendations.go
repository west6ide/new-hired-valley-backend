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
func GetUserByID(userID *authentication.Claims) (users.User, error) {
	var user users.User
	// Поиск пользователя по ID с загрузкой навыков и интересов
	result := config.DB.Preload("Skills").Preload("Interests").Preload("Stories").First(&user, userID)

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
	prompt := fmt.Sprintf("На основе навыков: %v, интересов: %v и роли: %s предложите релевантный контент.",
		user.Skills, user.Interests, user.Role)

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
	req.Header.Set("x-vanusai-model", "gpt-3.5-turbo")
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

	var recommendations []recommend.Content
	if err := json.Unmarshal(body, &recommendations); err != nil {
		return nil, fmt.Errorf("ошибка при декодировании ответа: %w", err)
	}

	return recommendations, nil
}

// GetRecommendationsHandler обрабатывает запрос для получения рекомендаций
func GetRecommendationsHandler(w http.ResponseWriter, r *http.Request) {
	// Проверка и извлечение userID из токена
	userID, err := authentication.ValidateToken(r)
	if err != nil {
		http.Error(w, "Unauthorized: "+err.Error(), http.StatusUnauthorized)
		return
	}

	// Получение профиля пользователя по userID
	user, err := GetUserByID(userID)
	if err != nil {
		http.Error(w, "User not found", http.StatusNotFound)
		return
	}

	// Получение персонализированных рекомендаций
	recs, err := GetPersonalizedRecommendations(user)
	if err != nil {
		http.Error(w, "Failed to get recommendations", http.StatusInternalServerError)
		return
	}

	// Возвращаем рекомендации в JSON формате
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(recs)
}

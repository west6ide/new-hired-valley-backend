package recommendations

import (
	"bytes"
	"encoding/json"
	"fmt"
	"hired-valley-backend/models/recommend"
	"io/ioutil"
	"net/http"
)

const vanusAPIURL = "https://app.ai.vanus.ai/api/v1/0b973b9cd13f4635acae25277820b407" // Замените на ваш реальный API-ключ

// GetRecommendationsHandler обрабатывает запросы для получения рекомендаций
func GetRecommendationsHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		return
	}

	// Извлекаем промпт из тела запроса
	var reqBody recommend.RecommendationRequest
	if err := json.NewDecoder(r.Body).Decode(&reqBody); err != nil {
		http.Error(w, "Bad request", http.StatusBadRequest)
		return
	}

	// Отправляем запрос в Vanus AI
	response, err := sendRequestToVanusAI(reqBody)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to get recommendations: %v", err), http.StatusInternalServerError)
		return
	}

	// Возвращаем ответ клиенту
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// sendRequestToVanusAI отправляет запрос к Vanus AI и возвращает ответ
func sendRequestToVanusAI(req recommend.RecommendationRequest) (*recommend.RecommendationResponse, error) {
	reqBody, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("failed to create request body: %w", err)
	}

	httpReq, err := http.NewRequest("POST", vanusAPIURL, bytes.NewBuffer(reqBody))
	if err != nil {
		return nil, fmt.Errorf("failed to create HTTP request: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("x-vanusai-model", "gpt-4")
	httpReq.Header.Set("x-vanusai-sessionid", "123e4567-e89b-12d3-a456-426614174000") // Генерируйте уникальный ID

	client := &http.Client{}
	resp, err := client.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("failed to send request to Vanus AI: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	var vanusResponse recommend.RecommendationResponse
	if err := json.Unmarshal(respBody, &vanusResponse); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &vanusResponse, nil
}

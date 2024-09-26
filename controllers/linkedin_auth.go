package controllers

import (
	"context"
	"encoding/json"
	"fmt"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/linkedin"
	"hired-valley-backend/config"
	"hired-valley-backend/models"
	"io/ioutil"
	"log"
	"net/http"
	"os"
)

var linkedinOAuthConfig = &oauth2.Config{
	ClientID:     os.Getenv("LINKEDIN_CLIENT_ID"),
	ClientSecret: os.Getenv("LINKEDIN_CLIENT_SECRET"),
	RedirectURL:  os.Getenv("LINKEDIN_REDIRECT_URL"),
	Scopes:       []string{"r_liteprofile", "r_emailaddress"}, // Убедитесь, что здесь указаны корректные scopes
	Endpoint:     linkedin.Endpoint,
}

// Обработчик для начала авторизации через LinkedIn
func HandleLinkedInLogin(w http.ResponseWriter, r *http.Request) {
	url := linkedinOAuthConfig.AuthCodeURL("state", oauth2.AccessTypeOffline)
	http.Redirect(w, r, url, http.StatusTemporaryRedirect)
}

// Обработчик для получения токена и данных пользователя
func HandleLinkedInCallback(w http.ResponseWriter, r *http.Request) {
	// Получение кода авторизации
	code := r.URL.Query().Get("code")
	if code == "" {
		http.Error(w, "Код авторизации отсутствует", http.StatusBadRequest)
		return
	}

	// Получение токена
	token, err := linkedinOAuthConfig.Exchange(context.Background(), code)
	if err != nil {
		http.Error(w, "Не удалось получить токен: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Создание клиента для запросов к LinkedIn API
	client := linkedinOAuthConfig.Client(context.Background(), token)

	// Запрос данных профиля пользователя
	resp, err := client.Get("https://api.linkedin.com/v2/me")
	if err != nil {
		http.Error(w, "Не удалось получить данные профиля: "+err.Error(), http.StatusInternalServerError)
		return
	}
	defer resp.Body.Close()

	// Проверка кода ответа от LinkedIn API
	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := ioutil.ReadAll(resp.Body)
		bodyString := string(bodyBytes)
		log.Printf("Ошибка при запросе данных профиля: статус код %d, тело ответа: %s", resp.StatusCode, bodyString)
		http.Error(w, fmt.Sprintf("Не удалось получить корректный ответ от LinkedIn API. Статус: %d, Ответ: %s", resp.StatusCode, bodyString), http.StatusInternalServerError)
		return
	}

	// Логируем ответ для отладки
	log.Println("Ответ от LinkedIn (профиль):", resp)

	// Декодирование данных профиля
	var profileData map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&profileData); err != nil {
		http.Error(w, "Ошибка при декодировании данных профиля: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Логируем декодированные данные профиля для отладки
	log.Println("Декодированные данные профиля:", profileData)

	// Создаем пользователя с данными из профиля
	var linkedInUser models.LinkedInUser
	if id, ok := profileData["id"].(string); ok {
		linkedInUser.LinkedInID = id
	}
	if firstName, ok := profileData["localizedFirstName"].(string); ok {
		linkedInUser.FirstName = firstName
	}
	if lastName, ok := profileData["localizedLastName"].(string); ok {
		linkedInUser.LastName = lastName
	}

	// Запрос email пользователя
	emailResp, err := client.Get("https://api.linkedin.com/v2/emailAddress?q=members&projection=(elements*(handle~))")
	if err != nil {
		http.Error(w, "Не удалось получить email: "+err.Error(), http.StatusInternalServerError)
		return
	}
	defer emailResp.Body.Close()

	// Проверка кода ответа от LinkedIn API
	if emailResp.StatusCode != http.StatusOK {
		bodyBytes, _ := ioutil.ReadAll(emailResp.Body)
		bodyString := string(bodyBytes)
		log.Printf("Ошибка при запросе email: статус код %d, тело ответа: %s", emailResp.StatusCode, bodyString)
		http.Error(w, fmt.Sprintf("Не удалось получить корректный ответ от LinkedIn API для email. Статус: %d, Ответ: %s", emailResp.StatusCode, bodyString), http.StatusInternalServerError)
		return
	}

	// Логируем ответ для отладки
	log.Println("Ответ от LinkedIn (email):", emailResp)

	// Структура для email ответа
	type EmailResponse struct {
		Elements []struct {
			HandleTilde struct {
				EmailAddress string `json:"emailAddress"`
			} `json:"handle~"`
		} `json:"elements"`
	}

	var emailData EmailResponse
	if err := json.NewDecoder(emailResp.Body).Decode(&emailData); err != nil {
		http.Error(w, "Ошибка при декодировании email: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Сохранение email в структуре пользователя
	if len(emailData.Elements) > 0 {
		linkedInUser.Email = emailData.Elements[0].HandleTilde.EmailAddress
	}

	// Сохранение токена в структуре пользователя
	linkedInUser.AccessToken = token.AccessToken

	// Логируем данные пользователя перед сохранением
	log.Println("Данные пользователя перед сохранением:", linkedInUser)

	// Сохранение данных в базу данных
	if err := config.DB.Create(&linkedInUser).Error; err != nil {
		http.Error(w, "Ошибка при сохранении данных в базу: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Отображение данных пользователя
	fmt.Fprintf(w, "Добро пожаловать, %s %s! Ваш email: %s", linkedInUser.FirstName, linkedInUser.LastName, linkedInUser.Email)
}

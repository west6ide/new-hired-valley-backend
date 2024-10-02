package controllers

import (
	"context"
	"encoding/json"
	"fmt"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/linkedin"
	"hired-valley-backend/config"
	"hired-valley-backend/models"
	"net/http"
	"os"
)

var linkedinOAuthConfig = &oauth2.Config{
	ClientID:     os.Getenv("LINKEDIN_CLIENT_ID"),
	ClientSecret: os.Getenv("LINKEDIN_CLIENT_SECRET"),
	RedirectURL:  os.Getenv("LINKEDIN_REDIRECT_URL"),
	Scopes:       []string{"openid", "profile", "email", "w_member_social"},
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

	// Декодирование данных профиля
	var linkedInUser models.LinkedInUser
	if err := json.NewDecoder(resp.Body).Decode(&linkedInUser); err != nil {
		http.Error(w, "Ошибка при декодировании данных профиля: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Запрос email пользователя
	emailResp, err := client.Get("https://api.linkedin.com/v2/emailAddress?q=members&projection=(elements*(handle~))")
	if err != nil {
		http.Error(w, "Не удалось получить email: "+err.Error(), http.StatusInternalServerError)
		return
	}
	defer emailResp.Body.Close()

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

	// Сохранение данных пользователя в базу данных
	if err := config.DB.Create(&linkedInUser).Error; err != nil {
		http.Error(w, "Ошибка при сохранении данных пользователя в базу: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Сохранение токена в базу данных
	oauthToken := models.OAuthToken{
		UserID:      linkedInUser.ID, // предполагается, что ID пользователя хранится в модели LinkedInUser
		AccessToken: token.AccessToken,
		TokenType:   token.TokenType,
		Expiry:      token.Expiry,
	}

	if err := config.DB.Create(&oauthToken).Error; err != nil {
		http.Error(w, "Ошибка при сохранении токена в базу: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Отображение данных пользователя
	fmt.Fprintf(w, "Добро пожаловать, %s %s! Ваш email: %s", linkedInUser.FirstName, linkedInUser.LastName, linkedInUser.Email)
}

// Получение данных пользователя из токена
func GetUserFromToken(w http.ResponseWriter, r *http.Request) {
	// Получение userID из запроса (например, из параметров или сессии)
	userID := uint(1) // Пример, заменить на реальный userID

	var token models.OAuthToken

	// Поиск токена по userID
	if err := config.DB.Where("user_id = ?", userID).First(&token).Error; err != nil {
		http.Error(w, "Не удалось найти токен пользователя: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Восстановление токена в структуру oauth2.Token
	oauthToken := &oauth2.Token{
		AccessToken: token.AccessToken,
		TokenType:   token.TokenType,
		Expiry:      token.Expiry,
	}

	// Создание нового клиента с использованием сохраненного токена
	client := linkedinOAuthConfig.Client(context.Background(), oauthToken)

	// Выполнение запроса к LinkedIn API для получения данных профиля
	resp, err := client.Get("https://api.linkedin.com/v2/me")
	if err != nil {
		http.Error(w, "Не удалось получить данные профиля: "+err.Error(), http.StatusInternalServerError)
		return
	}
	defer resp.Body.Close()

	// Декодирование данных профиля
	var linkedInUser models.LinkedInUser
	if err := json.NewDecoder(resp.Body).Decode(&linkedInUser); err != nil {
		http.Error(w, "Ошибка при декодировании данных профиля: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Запрос email пользователя
	emailResp, err := client.Get("https://api.linkedin.com/v2/emailAddress?q=members&projection=(elements*(handle~))")
	if err != nil {
		http.Error(w, "Не удалось получить email: "+err.Error(), http.StatusInternalServerError)
		return
	}
	defer emailResp.Body.Close()

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

	// Отображение данных пользователя
	fmt.Fprintf(w, "Добро пожаловать, %s %s! Ваш email: %s", linkedInUser.FirstName, linkedInUser.LastName, linkedInUser.Email)
}

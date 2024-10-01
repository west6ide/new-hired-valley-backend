package controllers

import (
	"context"
	"encoding/json"
	"fmt"
	"golang.org/x/oauth2"
	"hired-valley-backend/config"
	"hired-valley-backend/models"
	"net/http"
	"os"
)

var linkedinOAuthConfig = &oauth2.Config{
	ClientID:     os.Getenv("LINKEDIN_CLIENT_ID"),
	ClientSecret: os.Getenv("LINKEDIN_CLIENT_SECRET"),
	RedirectURL:  os.Getenv("LINKEDIN_REDIRECT_URL"),
	Scopes:       []string{"openid", "profile", "email"},
	Endpoint: oauth2.Endpoint{
		AuthURL:   "https://www.linkedin.com/oauth/v2/authorization",
		TokenURL:  "https://www.linkedin.com/oauth/v2/accessToken",
		AuthStyle: oauth2.AuthStyleInParams,
	},
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
	linkedInUser, err := fetchLinkedInUserProfile(client)
	if err != nil {
		http.Error(w, "Ошибка при запросе профиля пользователя: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Запрос email пользователя
	linkedInUser.Email, err = fetchLinkedInEmail(client)
	if err != nil {
		http.Error(w, "Ошибка при запросе email пользователя: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Сохранение пользователя в базу данных
	err = saveOrUpdateUser(linkedInUser, token.AccessToken)
	if err != nil {
		http.Error(w, "Ошибка при сохранении данных пользователя: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Отображение токена пользователя после успешной авторизации
	fmt.Fprintf(w, "Ваш токен: %s", token.AccessToken)
}

// fetchLinkedInUserProfile запрашивает данные профиля LinkedIn пользователя
func fetchLinkedInUserProfile(client *http.Client) (models.LinkedInUser, error) {
	resp, err := client.Get("https://api.linkedin.com/v2/me")
	if err != nil {
		return models.LinkedInUser{}, err
	}
	defer resp.Body.Close()

	var linkedInUser models.LinkedInUser
	if err := json.NewDecoder(resp.Body).Decode(&linkedInUser); err != nil {
		return models.LinkedInUser{}, err
	}

	return linkedInUser, nil
}

// fetchLinkedInEmail запрашивает email пользователя через LinkedIn API
func fetchLinkedInEmail(client *http.Client) (string, error) {
	resp, err := client.Get("https://api.linkedin.com/v2/emailAddress?q=members&projection=(elements*(handle~))")
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	var emailData struct {
		Elements []struct {
			HandleTilde struct {
				EmailAddress string `json:"emailAddress"`
			} `json:"handle~"`
		} `json:"elements"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&emailData); err != nil {
		return "", err
	}

	if len(emailData.Elements) > 0 {
		return emailData.Elements[0].HandleTilde.EmailAddress, nil
	}

	return "", fmt.Errorf("не удалось найти email пользователя")
}

// saveOrUpdateUser сохраняет или обновляет пользователя в базе данных
func saveOrUpdateUser(linkedInUser models.LinkedInUser, accessToken string) error {
	var existingUser models.LinkedInUser

	// Проверяем, существует ли пользователь
	if err := config.DB.Where("email = ?", linkedInUser.Email).First(&existingUser).Error; err == nil {
		// Пользователь существует — обновляем его данные
		existingUser.FirstName = linkedInUser.FirstName
		existingUser.LastName = linkedInUser.LastName
		existingUser.AccessToken = accessToken
		return config.DB.Save(&existingUser).Error
	}

	// Пользователь не найден — создаём нового
	linkedInUser.AccessToken = accessToken
	return config.DB.Create(&linkedInUser).Error
}

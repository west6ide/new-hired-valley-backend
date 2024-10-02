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
	var linkedInUserProfile struct {
		ID        string `json:"id"`
		FirstName struct {
			Localized map[string]string `json:"localized"`
		} `json:"firstName"`
		LastName struct {
			Localized map[string]string `json:"localized"`
		} `json:"lastName"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&linkedInUserProfile); err != nil {
		http.Error(w, "Ошибка при декодировании данных профиля: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Получение email пользователя
	emailResp, err := client.Get("https://api.linkedin.com/v2/emailAddress?q=members&projection=(elements*(handle~))")
	if err != nil {
		http.Error(w, "Не удалось получить email: "+err.Error(), http.StatusInternalServerError)
		return
	}
	defer emailResp.Body.Close()

	// Структура для email ответа
	var emailData struct {
		Elements []struct {
			HandleTilde struct {
				EmailAddress string `json:"emailAddress"`
			} `json:"handle~"`
		} `json:"elements"`
	}

	if err := json.NewDecoder(emailResp.Body).Decode(&emailData); err != nil {
		http.Error(w, "Ошибка при декодировании email: "+err.Error(), http.StatusInternalServerError)
		return
	}

	email := ""
	if len(emailData.Elements) > 0 {
		email = emailData.Elements[0].HandleTilde.EmailAddress
	}

	// Сохранение данных пользователя в базу данных
	linkedInUser := models.LinkedInUser{
		FirstName:   linkedInUserProfile.FirstName.Localized["en_US"],
		LastName:    linkedInUserProfile.LastName.Localized["en_US"],
		Email:       email,
		LinkedInID:  linkedInUserProfile.ID,
		AccessToken: token.AccessToken, // Сохранение AccessToken
	}

	if err := config.DB.Where("linkedin_id = ?", linkedInUser.LinkedInID).FirstOrCreate(&linkedInUser).Error; err != nil {
		http.Error(w, "Ошибка при сохранении данных пользователя в базу: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Отображение данных пользователя
	fmt.Fprintf(w, "Добро пожаловать, %s %s! Ваш email: %s", linkedInUser.FirstName, linkedInUser.LastName, linkedInUser.Email, linkedInUser.LinkedInID)
}

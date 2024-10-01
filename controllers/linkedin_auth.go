package controllers

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/gorilla/sessions"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/linkedin"
	"hired-valley-backend/config"
	"hired-valley-backend/models"
	"net/http"
	"os"
)

var (
	linkedinOAuthConfig = &oauth2.Config{
		ClientID:     os.Getenv("LINKEDIN_CLIENT_ID"),
		ClientSecret: os.Getenv("LINKEDIN_CLIENT_SECRET"),
		RedirectURL:  os.Getenv("LINKEDIN_REDIRECT_URL"),
		Scopes:       []string{"openid", "r_liteprofile", "r_emailaddress"},
		Endpoint:     linkedin.Endpoint,
	}
	storeLinkedin = sessions.NewCookieStore([]byte("something-very-secret"))
)

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
	var linkedInUser map[string]interface{}
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

	// Получение данных пользователя
	linkedInID := linkedInUser["id"].(string)
	firstName := linkedInUser["localizedFirstName"].(string)
	lastName := linkedInUser["localizedLastName"].(string)
	email := emailData.Elements[0].HandleTilde.EmailAddress

	// Сохранение или обновление данных пользователя в базе данных
	var user models.LinkedInUser
	config.DB.Where("linkedin_id = ?", linkedInID).First(&user)

	if user.LinkedInID == "" {
		// Пользователь не найден, создаем нового
		user = models.LinkedInUser{
			LinkedInID:  linkedInID,
			FirstName:   firstName,
			LastName:    lastName,
			Email:       email,
			AccessToken: token.AccessToken,
		}
		config.DB.Create(&user)
	} else {
		// Обновление данных пользователя
		user.FirstName = firstName
		user.LastName = lastName
		user.Email = email
		user.AccessToken = token.AccessToken
		config.DB.Save(&user)
	}

	// Сохранение информации о пользователе в сессии
	session, _ := storeLinkedin.Get(r, "session-name")
	session.Values["user"] = user
	session.Save(r, w)

	// Отображение данных пользователя
	fmt.Fprintf(w, "Добро пожаловать, %s %s! Ваш email: %s", user.FirstName, user.LastName, user.Email)
}

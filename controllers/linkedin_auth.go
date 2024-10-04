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
	Scopes:       []string{"r_liteprofile", "r_emailaddress"}, // Необходимо корректировать в зависимости от разрешений
	Endpoint:     linkedin.Endpoint,
}

// Обработчик для начала авторизации через LinkedIn
func HandleLinkedInLogin(w http.ResponseWriter, r *http.Request) {
	// Генерация уникального состояния (state) для предотвращения CSRF атак
	state := "state" // В реальном приложении лучше генерировать случайное значение
	url := linkedinOAuthConfig.AuthCodeURL(state, oauth2.AccessTypeOffline)
	http.Redirect(w, r, url, http.StatusTemporaryRedirect)
}

// Обработчик для получения токена и данных пользователя через LinkedIn API
func HandleLinkedInCallback(w http.ResponseWriter, r *http.Request) {
	// Проверка параметра state для защиты от CSRF атак
	if r.FormValue("state") != "state" {
		http.Error(w, "Invalid state parameter", http.StatusBadRequest)
		return
	}

	// Получение кода авторизации
	code := r.URL.Query().Get("code")
	if code == "" {
		http.Error(w, "Код авторизации отсутствует", http.StatusBadRequest)
		return
	}

	// Получение токена по коду авторизации
	token, err := linkedinOAuthConfig.Exchange(context.Background(), code)
	if err != nil {
		http.Error(w, "Не удалось получить токен: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Создание HTTP-клиента для запросов с использованием полученного токена
	client := linkedinOAuthConfig.Client(context.Background(), token)

	// Запрос на получение профиля пользователя
	userInfo, err := getLinkedInUserInfo(client)
	if err != nil {
		http.Error(w, "Не удалось получить данные пользователя: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Сохранение или обновление данных пользователя в базе данных
	err = saveLinkedInUserToDB(userInfo, token.AccessToken)
	if err != nil {
		http.Error(w, "Ошибка при сохранении пользователя: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Успешная авторизация и отображение приветственного сообщения
	fmt.Fprintf(w, "Добро пожаловать, %s %s! Ваш email: %s", userInfo.FirstName, userInfo.LastName, userInfo.Email)
}

// Структура для декодирования данных пользователя из LinkedIn
type LinkedInUserInfo struct {
	FirstName string `json:"localizedFirstName"`
	LastName  string `json:"localizedLastName"`
	Email     string
	Sub       string `json:"id"` // LinkedIn ID
}

// Получение данных пользователя с LinkedIn API
func getLinkedInUserInfo(client *http.Client) (*LinkedInUserInfo, error) {
	// Запрос на получение профиля пользователя
	profileResp, err := client.Get("https://api.linkedin.com/v2/me")
	if err != nil {
		return nil, err
	}
	defer profileResp.Body.Close()

	// Декодирование профиля пользователя
	var profileData map[string]interface{}
	if err := json.NewDecoder(profileResp.Body).Decode(&profileData); err != nil {
		return nil, err
	}

	// Запрос на получение email-адреса пользователя
	emailResp, err := client.Get("https://api.linkedin.com/v2/emailAddress?q=members&projection=(elements*(handle~))")
	if err != nil {
		return nil, err
	}
	defer emailResp.Body.Close()

	// Декодирование email-адреса
	var emailData map[string]interface{}
	if err := json.NewDecoder(emailResp.Body).Decode(&emailData); err != nil {
		return nil, err
	}

	emailElements := emailData["elements"].([]interface{})
	email := emailElements[0].(map[string]interface{})["handle~"].(map[string]interface{})["emailAddress"].(string)

	// Создание структуры пользователя с полученными данными
	userInfo := &LinkedInUserInfo{
		FirstName: profileData["localizedFirstName"].(string),
		LastName:  profileData["localizedLastName"].(string),
		Email:     email,
		Sub:       profileData["id"].(string),
	}

	return userInfo, nil
}

// Сохранение или обновление пользователя в базе данных
func saveLinkedInUserToDB(userInfo *LinkedInUserInfo, accessToken string) error {
	// Проверка на существование пользователя в базе данных по email
	var user models.User
	if err := config.DB.Where("email = ?", userInfo.Email).First(&user).Error; err != nil {
		// Если пользователя нет, создаем нового
		user = models.User{
			Email:       userInfo.Email,
			Name:        userInfo.FirstName + " " + userInfo.LastName,
			Provider:    "linkedin",
			AccessToken: accessToken,
		}
		if err := config.DB.Create(&user).Error; err != nil {
			return err
		}
	}

	// Проверка на существование LinkedInUser
	var linkedInUser models.LinkedInUser
	config.DB.Where("sub = ?", userInfo.Sub).First(&linkedInUser)

	if linkedInUser.Sub == "" {
		// Если LinkedInUser не найден, создаем нового
		linkedInUser = models.LinkedInUser{
			UserID:      user.ID,
			Sub:         userInfo.Sub,
			FirstName:   userInfo.FirstName,
			LastName:    userInfo.LastName,
			Email:       userInfo.Email,
			AccessToken: accessToken,
		}
		if err := config.DB.Create(&linkedInUser).Error; err != nil {
			return err
		}
	} else {
		// Обновление данных пользователя LinkedIn
		linkedInUser.FirstName = userInfo.FirstName
		linkedInUser.LastName = userInfo.LastName
		linkedInUser.Email = userInfo.Email
		linkedInUser.AccessToken = accessToken
		if err := config.DB.Save(&linkedInUser).Error; err != nil {
			return err
		}
	}

	return nil
}

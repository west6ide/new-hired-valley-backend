package controllers

import (
	"context"
	"encoding/json"
	"fmt"
	"golang.org/x/oauth2"
	"hired-valley-backend/config"
	"hired-valley-backend/models"
	"io/ioutil"
	"net/http"
	"os"

	"gorm.io/gorm"
)

// LinkedInOAuthConfig — конфигурация для LinkedIn OAuth2
var LinkedInOAuthConfig = &oauth2.Config{
	ClientID:     os.Getenv("LINKEDIN_CLIENT_ID"),     // Установи эти переменные окружения
	ClientSecret: os.Getenv("LINKEDIN_CLIENT_SECRET"), // Установи эти переменные окружения
	RedirectURL:  os.Getenv("LINKEDIN_REDIRECT_URL"),  // URL для перенаправления после авторизации
	Scopes:       []string{"r_liteprofile", "r_emailaddress"},
	Endpoint: oauth2.Endpoint{
		AuthURL:  "https://www.linkedin.com/oauth/v2/authorization",
		TokenURL: "https://www.linkedin.com/oauth/v2/accessToken",
	},
}

// HandleLinkedInLogin обрабатывает запрос на авторизацию через LinkedIn
func HandleLinkedInLogin(w http.ResponseWriter, r *http.Request) {
	// Генерация URL для авторизации
	url := LinkedInOAuthConfig.AuthCodeURL("state")
	http.Redirect(w, r, url, http.StatusTemporaryRedirect)
}

// HandleLinkedInCallback обрабатывает редирект от LinkedIn после успешной авторизации
func HandleLinkedInCallback(w http.ResponseWriter, r *http.Request) {
	// Получение кода авторизации
	code := r.URL.Query().Get("code")

	// Обмен кода на токен
	token, err := LinkedInOAuthConfig.Exchange(context.Background(), code)
	if err != nil {
		http.Error(w, "Не удалось получить токен", http.StatusInternalServerError)
		return
	}

	// Использование токена для получения данных пользователя
	client := LinkedInOAuthConfig.Client(context.Background(), token)
	resp, err := client.Get("https://api.linkedin.com/v2/me")
	if err != nil {
		http.Error(w, "Ошибка получения профиля", http.StatusInternalServerError)
		return
	}
	defer resp.Body.Close()

	// Чтение данных ответа
	data, _ := ioutil.ReadAll(resp.Body)

	// Парсинг данных пользователя
	var linkedInUser models.LinkedInUser
	err = json.Unmarshal(data, &linkedInUser)
	if err != nil {
		http.Error(w, "Ошибка парсинга профиля", http.StatusInternalServerError)
		return
	}

	// Проверка наличия пользователя в базе данных
	var existingUser models.LinkedInUser
	result := config.DB.Where("linked_in_id = ?", linkedInUser.LinkedInID).First(&existingUser)

	// Если пользователь не найден, создаем нового
	if result.Error == gorm.ErrRecordNotFound {
		newUser := models.LinkedInUser{
			LinkedInID: linkedInUser.LinkedInID,
			FirstName:  linkedInUser.FirstName,
			LastName:   linkedInUser.LastName,
			Email:      linkedInUser.Email,
		}
		config.DB.Create(&newUser)
	} else if result.Error != nil {
		http.Error(w, "Ошибка базы данных", http.StatusInternalServerError)
		return
	}

	// Здесь можно создать сессию и сохранить информацию о пользователе

	fmt.Fprintf(w, "Добро пожаловать, %s %s!", linkedInUser.FirstName, linkedInUser.LastName)
}

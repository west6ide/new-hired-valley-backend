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

// Обработчик для получения токена и данных пользователя через /v2/userinfo
func HandleLinkedInCallback(w http.ResponseWriter, r *http.Request) {
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

	// Создание клиента для запросов
	client := linkedinOAuthConfig.Client(context.Background(), token)

	// Запрос к /v2/userinfo для получения данных пользователя
	userinfoResp, err := client.Get("https://api.linkedin.com/v2/userinfo")
	if err != nil {
		http.Error(w, "Не удалось получить данные пользователя: "+err.Error(), http.StatusInternalServerError)
		return
	}
	defer userinfoResp.Body.Close()

	// Декодирование данных пользователя
	var userInfo map[string]interface{}
	if err := json.NewDecoder(userinfoResp.Body).Decode(&userInfo); err != nil {
		http.Error(w, "Ошибка при декодировании данных пользователя: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Проверка на существование пользователя в основной таблице User
	user, err := createOrGetUser(userInfo["email"].(string), userInfo["given_name"].(string), userInfo["family_name"].(string), token.AccessToken)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Проверка на существование пользователя в базе данных LinkedInUser
	var linkedInUser models.LinkedInUser
	if err := config.DB.Where("sub = ?", userInfo["sub"]).First(&linkedInUser).Error; err != nil {
		// Если пользователя нет, создаем запись в таблице LinkedInUser
		linkedInUser = models.LinkedInUser{
			UserID:      user.ID, // Связь с таблицей User
			Sub:         userInfo["sub"].(string),
			FirstName:   userInfo["given_name"].(string),
			LastName:    userInfo["family_name"].(string),
			Email:       userInfo["email"].(string),
			AccessToken: token.AccessToken,
		}
		if err := config.DB.Create(&linkedInUser).Error; err != nil {
			http.Error(w, "Ошибка при сохранении пользователя: "+err.Error(), http.StatusInternalServerError)
			return
		}
	}

	// Успешная авторизация и сохранение данных
	fmt.Fprintf(w, "Добро пожаловать, %s %s! Ваш email: %s", linkedInUser.FirstName, linkedInUser.LastName, linkedInUser.Email)
}

// Функция для создания или получения пользователя в основной таблице User
func createOrGetUser(email, firstName, lastName, accessToken string) (*models.User, error) {
	// Проверка, существует ли пользователь в таблице User
	var user models.User
	if err := config.DB.Where("email = ?", email).First(&user).Error; err != nil {
		// Если пользователя нет, создаем нового
		user = models.User{
			ID:          user.ID,
			Email:       email,
			Name:        firstName + " " + lastName,
			Provider:    "linkedin",  // Указываем, что провайдер LinkedIn
			AccessToken: accessToken, // Сохраняем токен доступа
		}
		if err := config.DB.Create(&user).Error; err != nil {
			return nil, fmt.Errorf("ошибка при создании пользователя в таблице User: %v", err)
		}
	}
	return &user, nil
}

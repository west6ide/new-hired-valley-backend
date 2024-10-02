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

	// Проверка на существование пользователя в базе данных
	var user models.LinkedInUser
	if err := config.DB.Where("sub = ?", userInfo["sub"]).First(&user).Error; err != nil {
		// Если пользователя нет, создаем его
		user = models.LinkedInUser{
			Sub:       userInfo["sub"].(string),
			FirstName: userInfo["given_name"].(string),
			LastName:  userInfo["family_name"].(string),
			Email:     userInfo["email"].(string),
		}
		if err := config.DB.Create(&user).Error; err != nil {
			http.Error(w, "Ошибка при сохранении пользователя: "+err.Error(), http.StatusInternalServerError)
			return
		}
	}

	// Сохранение токена OAuth в базу данных
	oauthToken := models.OAuthToken{
		UserID:      user.ID, // Связь с пользователем
		AccessToken: token.AccessToken,
		TokenType:   token.TokenType,
		Expiry:      token.Expiry,
	}

	// Если токен уже существует, обновляем его, иначе создаем новый
	if err := config.DB.Where("user_id = ?", user.ID).Assign(&oauthToken).FirstOrCreate(&oauthToken).Error; err != nil {
		http.Error(w, "Ошибка при сохранении токена: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Успешная авторизация и сохранение данных
	fmt.Fprintf(w, "Добро пожаловать, %s %s! Ваш email: %s", user.FirstName, user.LastName, user.Email)
}

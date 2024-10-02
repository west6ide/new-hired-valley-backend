package controllers

import (
	"context"
	"encoding/json"
	"fmt"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/linkedin"
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

// Обработчик для получения данных через /v2/userinfo
func HandleLinkedInCallback(w http.ResponseWriter, r *http.Request) {
	code := r.URL.Query().Get("code")
	if code == "" {
		http.Error(w, "Код авторизации отсутствует", http.StatusBadRequest)
		return
	}

	// Получаем токен
	token, err := linkedinOAuthConfig.Exchange(context.Background(), code)
	if err != nil {
		http.Error(w, "Не удалось получить токен: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Создаем клиент для запросов
	client := linkedinOAuthConfig.Client(context.Background(), token)

	// Выполняем запрос к /v2/userinfo
	userinfoResp, err := client.Get("https://api.linkedin.com/v2/userinfo")
	if err != nil {
		http.Error(w, "Не удалось получить данные пользователя: "+err.Error(), http.StatusInternalServerError)
		return
	}
	defer userinfoResp.Body.Close()

	// Декодируем ответ
	var userInfo map[string]interface{}
	if err := json.NewDecoder(userinfoResp.Body).Decode(&userInfo); err != nil {
		http.Error(w, "Ошибка при декодировании данных пользователя: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Обрабатываем полученные данные
	fmt.Fprintf(w, "Добро пожаловать, %v!", userInfo["sub"])
}

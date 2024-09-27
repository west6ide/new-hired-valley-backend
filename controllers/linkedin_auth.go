package controllers

import (
	"context"
	"encoding/json"
	"fmt"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/linkedin"
	"gorm.io/gorm"
	"hired-valley-backend/models"
	"net/http"
	"os"
)

// OAuth2 конфигурация для LinkedIn
var linkedinOAuthConfig = &oauth2.Config{
	ClientID:     os.Getenv("LINKEDIN_CLIENT_ID"),
	ClientSecret: os.Getenv("LINKEDIN_CLIENT_SECRET"),
	RedirectURL:  os.Getenv("LINKEDIN_REDIRECT_URL"),
	Scopes:       []string{"r_liteprofile", "r_emailaddress"},
	Endpoint:     linkedin.Endpoint,
}

// Структура для профиля LinkedIn
type LinkedInProfile struct {
	ID        string `json:"id"`
	FirstName string `json:"localizedFirstName"`
	LastName  string `json:"localizedLastName"`
	Email     string `json:"emailAddress"`
}

// Структура для ответа с email
type EmailResponse struct {
	Elements []struct {
		Handle struct {
			EmailAddress string `json:"emailAddress"`
		} `json:"handle~"`
	} `json:"elements"`
}

// Функция для получения email пользователя из LinkedIn
func getEmail(accessToken string) (string, error) {
	req, err := http.NewRequest("GET", "https://api.linkedin.com/v2/emailAddress?q=members&projection=(elements*(handle~))", nil)
	if err != nil {
		return "", err
	}
	req.Header.Set("Authorization", "Bearer "+accessToken)
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	var emailResp EmailResponse
	if err := json.NewDecoder(resp.Body).Decode(&emailResp); err != nil {
		return "", err
	}
	if len(emailResp.Elements) > 0 {
		return emailResp.Elements[0].Handle.EmailAddress, nil
	}
	return "", fmt.Errorf("email not found")
}

// Функция для получения профиля пользователя из LinkedIn
func getProfile(accessToken string) (*LinkedInProfile, error) {
	req, err := http.NewRequest("GET", "https://api.linkedin.com/v2/me", nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+accessToken)
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var profile LinkedInProfile
	if err := json.NewDecoder(resp.Body).Decode(&profile); err != nil {
		return nil, err
	}
	return &profile, nil
}

// Функция для обработки авторизации
func HandleLinkedInLogin(w http.ResponseWriter, r *http.Request) {
	url := linkedinOAuthConfig.AuthCodeURL("state", oauth2.AccessTypeOffline)
	http.Redirect(w, r, url, http.StatusTemporaryRedirect)
}

// Обработчик для обратного вызова LinkedIn
func HandleLinkedInCallback(w http.ResponseWriter, r *http.Request, db *gorm.DB) {
	code := r.URL.Query().Get("code")
	token, err := linkedinOAuthConfig.Exchange(context.Background(), code)
	if err != nil {
		http.Error(w, "Не удалось обменять код на токен", http.StatusInternalServerError)
		return
	}

	// Получение профиля пользователя LinkedIn
	profile, err := getProfile(token.AccessToken)
	if err != nil {
		http.Error(w, "Не удалось получить профиль LinkedIn", http.StatusInternalServerError)
		return
	}

	// Получение email пользователя LinkedIn
	email, err := getEmail(token.AccessToken)
	if err != nil {
		http.Error(w, "Не удалось получить email LinkedIn", http.StatusInternalServerError)
		return
	}
	profile.Email = email

	// Сохранение данных пользователя в базу данных
	user := models.LinkedInUser{
		LinkedInID: profile.ID,
		FirstName:  profile.FirstName,
		LastName:   profile.LastName,
		Email:      profile.Email,
	}
	if err := db.Create(&user).Error; err != nil {
		http.Error(w, "Не удалось сохранить пользователя в базу данных", http.StatusInternalServerError)
		return
	}

	// Ответ клиенту
	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, "Пользователь успешно авторизован: %s %s", profile.FirstName, profile.LastName)
}

package controllers

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/dgrijalva/jwt-go"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/linkedin"
	"hired-valley-backend/config"
	"hired-valley-backend/models"
	"net/http"
	"os"
	"time"
)

var linkedinOAuthConfig = &oauth2.Config{
	ClientID:     os.Getenv("LINKEDIN_CLIENT_ID"),
	ClientSecret: os.Getenv("LINKEDIN_CLIENT_SECRET"),
	RedirectURL:  os.Getenv("LINKEDIN_REDIRECT_URL"),
	Scopes:       []string{"openid", "profile", "email"},
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

	// Получение id_token из токена
	idToken, err := extractIDToken(token.Extra("id_token").(string))
	if err != nil {
		http.Error(w, "Ошибка при извлечении id_token: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Декодирование данных пользователя
	var linkedInUser models.LinkedInUser
	if err := json.Unmarshal(idToken, &linkedInUser); err != nil {
		http.Error(w, "Ошибка при декодировании данных пользователя: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Сохранение данных в базу данных
	if err := config.DB.Create(&linkedInUser).Error; err != nil {
		http.Error(w, "Ошибка при сохранении данных в базу: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Отображение данных пользователя
	fmt.Fprintf(w, "Добро пожаловать, %s %s! Ваш email: %s", linkedInUser.FirstName, linkedInUser.LastName, linkedInUser.Email)
}

// extractIDToken извлекает информацию пользователя из id_token
func extractIDToken(idToken string) (json.RawMessage, error) {
	token, err := jwt.Parse(idToken, nil)
	if err != nil {
		return nil, err
	}
	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		userInfo := map[string]interface{}{
			"firstName": claims["given_name"],
			"lastName":  claims["family_name"],
			"email":     claims["email"],
			"exp":       time.Unix(int64(claims["exp"].(float64)), 0),
		}
		return json.Marshal(userInfo)
	}
	return nil, fmt.Errorf("invalid token")
}

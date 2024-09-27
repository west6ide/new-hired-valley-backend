package controllers

import (
	"context"
	"crypto/rsa"
	"encoding/json"
	"fmt"
	"github.com/dgrijalva/jwt-go" // Import for JWT handling
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
	state := r.URL.Query().Get("state")
	if code == "" || state == "" {
		http.Error(w, "Код авторизации или состояние отсутствуют", http.StatusBadRequest)
		return
	}

	// Получение токена
	token, err := linkedinOAuthConfig.Exchange(context.Background(), code)
	if err != nil {
		http.Error(w, "Не удалось получить токен: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Извлечение id_token из ответа
	idToken, err := extractIDToken(token)
	if err != nil {
		http.Error(w, "Не удалось извлечь id_token: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Декодирование id_token
	var linkedInUser models.LinkedInUser
	if err := decodeIDToken(idToken, &linkedInUser); err != nil {
		http.Error(w, "Ошибка при декодировании id_token: "+err.Error(), http.StatusInternalServerError)
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

// Функция для извлечения id_token из access_token
func extractIDToken(token *oauth2.Token) (string, error) {
	idToken := token.Extra("id_token")
	if idTokenStr, ok := idToken.(string); ok {
		return idTokenStr, nil
	}
	return "", fmt.Errorf("id_token не найден")
}

func decodeIDToken(idToken string, user *models.LinkedInUser) error {
	// Парсинг JWT
	tkn, err := jwt.Parse(idToken, func(token *jwt.Token) (interface{}, error) {
		// Проверка алгоритма
		if _, ok := token.Method.(*jwt.SigningMethodRSA); !ok {
			return nil, fmt.Errorf("недействительный метод подписи")
		}

		// Получение открытого ключа LinkedIn
		keySet, err := getLinkedInPublicKeys()
		if err != nil {
			return nil, err
		}

		// Получение kid из токена
		if kid, ok := token.Header["kid"].(string); ok {
			if key, ok := keySet[kid]; ok {
				return key, nil
			}
		}
		return nil, fmt.Errorf("ключ не найден")
	})

	if err != nil {
		return err
	}

	if claims, ok := tkn.Claims.(jwt.MapClaims); ok && tkn.Valid {
		user.FirstName = claims["given_name"].(string)
		user.LastName = claims["family_name"].(string)
		user.Email = claims["email"].(string)
		return nil
	}

	return fmt.Errorf("недействительный токен")
}

// Функция для получения публичных ключей LinkedIn
func getLinkedInPublicKeys() (map[string]*rsa.PublicKey, error) {
	resp, err := http.Get("https://api.linkedin.com/v2/publicKeys")
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("ошибка при получении публичных ключей: %s", resp.Status)
	}

	var keySet map[string]*rsa.PublicKey
	if err := json.NewDecoder(resp.Body).Decode(&keySet); err != nil {
		return nil, err
	}

	return keySet, nil
}

package controllers

import (
	"fmt"
	"hired-valley-backend/config"
	"hired-valley-backend/models"
	"net/http"

	"github.com/markbates/goth"
	"github.com/markbates/goth/gothic"
	"github.com/markbates/goth/providers/linkedin"
	"os"
)

func init() {
	// Инициализация LinkedIn провайдера с использованием Goth
	goth.UseProviders(
		linkedin.New(os.Getenv("LINKEDIN_CLIENT_ID"), os.Getenv("LINKEDIN_CLIENT_SECRET"), os.Getenv("LINKEDIN_REDIRECT_URL"), "r_liteprofile", "r_emailaddress"),
	)
}

// Обработчик для начала авторизации через LinkedIn
func HandleLinkedInLogin(w http.ResponseWriter, r *http.Request) {
	gothic.BeginAuthHandler(w, r) // Начало процесса авторизации через Goth
}

// Обработчик для получения токена и данных пользователя после редиректа
func HandleLinkedInCallback(w http.ResponseWriter, r *http.Request) {
	// Получение пользователя из сессии после успешной авторизации
	user, err := gothic.CompleteUserAuth(w, r)
	if err != nil {
		http.Error(w, "Ошибка при завершении авторизации: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Создание структуры LinkedInUser из полученных данных
	linkedInUser := models.LinkedInUser{
		FirstName: user.FirstName,
		LastName:  user.LastName,
		Email:     user.Email,
	}

	// Сохранение данных пользователя в базу данных
	if err := config.DB.Create(&linkedInUser).Error; err != nil {
		http.Error(w, "Ошибка при сохранении данных пользователя в базу: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Сохранение токена в базу данных
	oauthToken := models.OAuthToken{
		UserID:      linkedInUser.ID,
		AccessToken: user.AccessToken,
		TokenType:   "Bearer", // Goth не возвращает тип токена напрямую
	}

	if err := config.DB.Create(&oauthToken).Error; err != nil {
		http.Error(w, "Ошибка при сохранении токена в базу: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Отображение данных пользователя
	fmt.Fprintf(w, "Добро пожаловать, %s %s! Ваш email: %s", linkedInUser.FirstName, linkedInUser.LastName, linkedInUser.Email)
}

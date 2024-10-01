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
)

var linkedinOAuthConfig = &oauth2.Config{
	ClientID:     os.Getenv("LINKEDIN_CLIENT_ID"),
	ClientSecret: os.Getenv("LINKEDIN_CLIENT_SECRET"),
	RedirectURL:  os.Getenv("LINKEDIN_REDIRECT_URL"),
	Scopes:       []string{"r_liteprofile", "r_emailaddress"}, // Добавляем правильные scopes
	Endpoint: oauth2.Endpoint{
		AuthURL:   "https://www.linkedin.com/oauth/v2/authorization",
		TokenURL:  "https://www.linkedin.com/oauth/v2/accessToken",
		AuthStyle: oauth2.AuthStyleInParams,
	},
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

	// Лог токена для отладки
	fmt.Println("Token:", token)

	// Создание клиента для запросов к LinkedIn API
	client := linkedinOAuthConfig.Client(context.Background(), token)

	// Запрос данных профиля пользователя
	resp, err := client.Get("https://api.linkedin.com/v2/me")
	if err != nil {
		http.Error(w, "Не удалось получить данные профиля: "+err.Error(), http.StatusInternalServerError)
		return
	}
	defer resp.Body.Close()

	// Лог статуса ответа
	fmt.Println("Profile response status:", resp.Status)

	// Декодирование данных профиля
	var linkedInUser models.LinkedInUser
	if err := json.NewDecoder(resp.Body).Decode(&linkedInUser); err != nil {
		http.Error(w, "Ошибка при декодировании данных профиля: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Лог профиля пользователя для отладки
	fmt.Println("LinkedIn User Profile:", linkedInUser)

	// Запрос email пользователя
	email, err := getUserEmail(client)
	if err != nil {
		http.Error(w, "Ошибка при запросе email пользователя: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Сохранение email в структуре пользователя
	linkedInUser.Email = email

	// Проверка, существует ли пользователь
	var existingUser models.LinkedInUser
	if err := config.DB.Where("email = ?", linkedInUser.Email).First(&existingUser).Error; err == nil {
		// Если пользователь существует, обновляем его данные и токен
		existingUser.FirstName = linkedInUser.FirstName
		existingUser.LastName = linkedInUser.LastName
		existingUser.AccessToken = token.AccessToken // Сохраняем новый токен
		if err := config.DB.Save(&existingUser).Error; err != nil {
			http.Error(w, "Ошибка при обновлении данных пользователя в базу: "+err.Error(), http.StatusInternalServerError)
			return
		}
	} else {
		// Сохранение нового пользователя с токеном в базу данных
		linkedInUser.AccessToken = token.AccessToken // Сохраняем токен при создании
		if err := config.DB.Create(&linkedInUser).Error; err != nil {
			http.Error(w, "Ошибка при сохранении данных пользователя в базу: "+err.Error(), http.StatusInternalServerError)
			return
		}
	}

	// Отображение токена пользователя после успешной авторизации
	fmt.Fprintf(w, "Ваш токен: %s", token.AccessToken)
}

// Вспомогательная функция для получения email пользователя через LinkedIn API
func getUserEmail(client *http.Client) (string, error) {
	resp, err := client.Get("https://api.linkedin.com/v2/emailAddress?q=members&projection=(elements*(handle~))")
	if err != nil {
		return "", fmt.Errorf("не удалось выполнить запрос: %v", err)
	}
	defer resp.Body.Close()

	// Лог ответа для отладки
	body, _ := ioutil.ReadAll(resp.Body)
	fmt.Println("Response Body: ", string(body))

	// Декодирование данных email
	var emailData struct {
		Elements []struct {
			HandleTilde struct {
				EmailAddress string `json:"emailAddress"`
			} `json:"handle~"`
		} `json:"elements"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&emailData); err != nil {
		return "", fmt.Errorf("ошибка при декодировании данных email: %v", err)
	}

	// Проверка наличия email
	if len(emailData.Elements) > 0 {
		return emailData.Elements[0].HandleTilde.EmailAddress, nil
	}

	return "", fmt.Errorf("не удалось найти email пользователя")
}

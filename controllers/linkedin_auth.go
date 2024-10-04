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
	Scopes:       []string{"openid", "profile", "email", "w_member_social"}, // Сохраняем ваши Scopes
	Endpoint:     linkedin.Endpoint,
}

// Обработчик для начала авторизации через LinkedIn
func HandleLinkedInLogin(w http.ResponseWriter, r *http.Request) {
	state := "state" // Можно заменить на случайное значение для защиты от CSRF
	url := linkedinOAuthConfig.AuthCodeURL(state, oauth2.AccessTypeOffline)
	http.Redirect(w, r, url, http.StatusTemporaryRedirect)
}

// Обработчик для получения токена и данных пользователя через LinkedIn API
func HandleLinkedInCallback(w http.ResponseWriter, r *http.Request) {
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

	// Создание клиента для запросов
	client := linkedinOAuthConfig.Client(context.Background(), token)

	// Запрос на получение профиля пользователя
	profile, err := getLinkedInProfile(client)
	if err != nil {
		http.Error(w, "Не удалось получить профиль: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Запрос на получение email-адреса пользователя
	email, err := getLinkedInEmail(client)
	if err != nil {
		// Если не удалось получить email, отправляем форму для ввода email вручную
		showEmailForm(w, profile, token.AccessToken)
		return
	}

	// Сохранение данных пользователя в базу данных
	err = saveLinkedInUserToDB(profile, email, token.AccessToken)
	if err != nil {
		http.Error(w, "Ошибка при сохранении пользователя: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Успешная авторизация и отображение приветственного сообщения
	fmt.Fprintf(w, "Добро пожаловать, %s %s! Ваш email: %s", profile.FirstName, profile.LastName, email)
}

// Функция для отображения формы, если email не был получен
func showEmailForm(w http.ResponseWriter, profile *LinkedInProfile, accessToken string) {
	html := fmt.Sprintf(`
		<html>
		<body>
			<p>Мы не смогли получить ваш email. Пожалуйста, введите его вручную для завершения регистрации.</p>
			<form action="/submit-email" method="post">
				<input type="hidden" name="first_name" value="%s">
				<input type="hidden" name="last_name" value="%s">
				<input type="hidden" name="sub" value="%s">
				<input type="hidden" name="access_token" value="%s">
				<label for="email">Email:</label>
				<input type="email" id="email" name="email" required>
				<button type="submit">Завершить регистрацию</button>
			</form>
		</body>
		</html>`, profile.FirstName, profile.LastName, profile.ID, accessToken)
	fmt.Fprint(w, html)
}

// Обработчик для получения введенного пользователем email
func HandleSubmitEmail(w http.ResponseWriter, r *http.Request) {
	// Получение данных из формы
	firstName := r.FormValue("first_name")
	lastName := r.FormValue("last_name")
	sub := r.FormValue("sub")
	email := r.FormValue("email")
	accessToken := r.FormValue("access_token")

	// Проверка наличия всех обязательных данных
	if firstName == "" || lastName == "" || sub == "" || email == "" || accessToken == "" {
		http.Error(w, "Все поля обязательны для заполнения", http.StatusBadRequest)
		return
	}

	// Создаем профиль LinkedIn с данными из формы
	profile := &LinkedInProfile{
		FirstName: firstName,
		LastName:  lastName,
		ID:        sub,
	}

	// Сохраняем данные в базу
	err := saveLinkedInUserToDB(profile, email, accessToken)
	if err != nil {
		http.Error(w, "Ошибка при сохранении данных в базу: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Успешная регистрация
	fmt.Fprintf(w, "Регистрация завершена. Добро пожаловать, %s %s! Ваш email: %s", firstName, lastName, email)
}

// Структура для хранения данных профиля LinkedIn
type LinkedInProfile struct {
	FirstName string `json:"localizedFirstName"`
	LastName  string `json:"localizedLastName"`
	ID        string `json:"id"` // LinkedIn ID (sub)
}

// Получение данных профиля с LinkedIn API
func getLinkedInProfile(client *http.Client) (*LinkedInProfile, error) {
	resp, err := client.Get("https://api.linkedin.com/v2/me")
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

// Получение email-адреса пользователя с LinkedIn API
func getLinkedInEmail(client *http.Client) (string, error) {
	resp, err := client.Get("https://api.linkedin.com/v2/emailAddress?q=members&projection=(elements*(handle~))")
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	var emailData struct {
		Elements []struct {
			Handle struct {
				EmailAddress string `json:"emailAddress"`
			} `json:"handle~"`
		} `json:"elements"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&emailData); err != nil {
		return "", err
	}

	// Проверка, есть ли email в ответе
	if len(emailData.Elements) == 0 || emailData.Elements[0].Handle.EmailAddress == "" {
		return "", fmt.Errorf("email не найден")
	}

	return emailData.Elements[0].Handle.EmailAddress, nil
}

// Сохранение или обновление пользователя в базе данных
func saveLinkedInUserToDB(profile *LinkedInProfile, email string, accessToken string) error {
	// Проверка на существование пользователя в базе данных по email
	var user models.User
	if err := config.DB.Where("email = ?", email).First(&user).Error; err != nil {
		// Если пользователя нет, создаем нового
		user = models.User{
			Email:       email,
			Name:        profile.FirstName + " " + profile.LastName,
			Provider:    "linkedin",
			AccessToken: accessToken,
		}
		if err := config.DB.Create(&user).Error; err != nil {
			return err
		}
	}

	// Проверка на существование LinkedInUser
	var linkedInUser models.LinkedInUser
	config.DB.Where("sub = ?", profile.ID).First(&linkedInUser)

	if linkedInUser.Sub == "" {
		// Если LinkedInUser не найден, создаем нового
		linkedInUser = models.LinkedInUser{
			UserID:      user.ID,
			Sub:         profile.ID,
			FirstName:   profile.FirstName,
			LastName:    profile.LastName,
			Email:       email,
			AccessToken: accessToken,
		}
		if err := config.DB.Create(&linkedInUser).Error; err != nil {
			return err
		}
	} else {
		// Обновление данных пользователя LinkedIn
		linkedInUser.FirstName = profile.FirstName
		linkedInUser.LastName = profile.LastName
		linkedInUser.Email = email
		linkedInUser.AccessToken = accessToken
		if err := config.DB.Save(&linkedInUser).Error; err != nil {
			return err
		}
	}

	return nil
}

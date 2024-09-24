package controllers

import (
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"hired-valley-backend/config"
	"hired-valley-backend/models"
	"io/ioutil"
	"net/http"
	"os"
)

var googleOauthConfig = &oauth2.Config{
	RedirectURL:  os.Getenv("GOOGLE_REDIRECT_URL"), // Используем переменную окружения для URL callback
	ClientID:     os.Getenv("GOOGLE_CLIENT_ID"),
	ClientSecret: os.Getenv("GOOGLE_CLIENT_SECRET"),
	Scopes:       []string{"https://www.googleapis.com/auth/userinfo.email", "https://www.googleapis.com/auth/userinfo.profile"},
	Endpoint:     google.Endpoint,
}

// Generate a new random state string
func generateState() (string, error) {
	b := make([]byte, 32)
	_, err := rand.Read(b)
	if err != nil {
		return "", err
	}
	return base64.URLEncoding.EncodeToString(b), nil
}

// HandleGoogleLogin initiates Google OAuth login
func HandleGoogleLogin(w http.ResponseWriter, r *http.Request) {
	state, err := generateState()
	if err != nil {
		http.Error(w, "Error generating state", http.StatusInternalServerError)
		return
	}

	http.SetCookie(w, &http.Cookie{
		Name:  "oauth_state",
		Value: state,
	})

	url := googleOauthConfig.AuthCodeURL(state, oauth2.AccessTypeOffline, oauth2.SetAuthURLParam("prompt", "select_account"))
	http.Redirect(w, r, url, http.StatusTemporaryRedirect)
}

func HandleGoogleCallback(w http.ResponseWriter, r *http.Request) {
	state, err := r.Cookie("oauth_state")
	if err != nil || r.FormValue("state") != state.Value {
		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
		return
	}

	token, err := googleOauthConfig.Exchange(r.Context(), r.FormValue("code"))
	if err != nil {
		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
		return
	}

	resp, err := http.Get("https://www.googleapis.com/oauth2/v2/userinfo?access_token=" + token.AccessToken)
	if err != nil {
		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
		return
	}
	defer resp.Body.Close()

	content, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
		return
	}

	var googleUser models.GoogleUser
	if err := json.Unmarshal(content, &googleUser); err != nil {
		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
		return
	}

	// Проверка, существует ли пользователь с данным email
	var existingUser models.GoogleUser
	result := config.DB.Where("email = ?", googleUser.Email).First(&existingUser)

	if result.Error != nil {
		// Если пользователь не найден, сохраняем его в базе
		config.DB.Create(&googleUser)
	} else {
		// Если найден, обновляем информацию о пользователе
		existingUser.FirstName = googleUser.FirstName
		existingUser.LastName = googleUser.LastName
		config.DB.Save(&existingUser)
		googleUser = existingUser
	}

	// Сохранение данных пользователя в сессии
	session, _ := config.Store.Get(r, "session-name")
	session.Values["user"] = googleUser
	session.Save(r, w)

	// Отображаем информацию о пользователе после авторизации
	html := fmt.Sprintf(`<html><body>
        <h1>Добро пожаловать, %s %s!</h1>
        <p>Email: %s</p>
        <img src="%s" alt="User Picture"/>
        <br>
        <a href="/logout">Выйти</a>
        </body></html>`,
		googleUser.FirstName, googleUser.LastName, googleUser.Email)

	fmt.Fprint(w, html)
}

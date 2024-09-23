package controllers

import (
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"hired-valley-backend/models"
	"io/ioutil"
	"net/http"
	"os"
)

var googleOauthConfig = &oauth2.Config{
	RedirectURL:  os.Getenv("GOOGLE_REDIRECT_URL"),
	ClientID:     os.Getenv("GOOGLE_CLIENT_ID"),
	ClientSecret: os.Getenv("GOOGLE_CLIENT_SECRET"),
	Scopes:       []string{"https://www.googleapis.com/auth/userinfo.email", "https://www.googleapis.com/auth/userinfo.profile"},
	Endpoint:     google.Endpoint,
}

// Генерация случайного состояния
func generateState() (string, error) {
	b := make([]byte, 32)
	_, err := rand.Read(b)
	if err != nil {
		return "", err
	}
	return base64.URLEncoding.EncodeToString(b), nil
}

// Начало Google OAuth авторизации
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

// Callback после успешной авторизации в Google
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

	http.SetCookie(w, &http.Cookie{
		Name:  "user",
		Value: googleUser.FirstName,
	})

	http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
}

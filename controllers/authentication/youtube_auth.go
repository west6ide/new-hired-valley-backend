package authentication

import (
	"context"
	"encoding/json"
	"errors"
	"golang.org/x/oauth2"
	"gorm.io/gorm"
	"hired-valley-backend/config"
	"hired-valley-backend/models/users"
	"io/ioutil"
	"log"
	"net/http"
	"time"
)

//var (
//	googleOauthConfig = &oauth2.Config{
//		RedirectURL:  os.Getenv("GOOGLE_REDIRECT_URL"),
//		ClientID:     os.Getenv("GOOGLE_CLIENT_ID"),
//		ClientSecret: os.Getenv("GOOGLE_CLIENT_SECRET"),
//		Scopes:       []string{"https://www.googleapis.com/auth/youtube.upload", "https://www.googleapis.com/auth/userinfo.profile", "https://www.googleapis.com/auth/userinfo.email"},
//		Endpoint:     google.Endpoint,
//	}
//	store = sessions.NewCookieStore([]byte(os.Getenv("SESSION_SECRET")))
//)

func init() {
	// Проверка переменных окружения
	if GoogleOauthConfig.ClientID == "" || GoogleOauthConfig.ClientSecret == "" || GoogleOauthConfig.RedirectURL == "" {
		log.Fatal("Не установлены переменные окружения для Google OAuth")
	}
}

// HandleGoogleLogin - инициирует вход через Google OAuth
func HandleGoogleYoutubeLogin(w http.ResponseWriter, r *http.Request) {
	state := "google"
	url := GoogleOauthConfig.AuthCodeURL(state)
	http.Redirect(w, r, url, http.StatusTemporaryRedirect)
}

// HandleGoogleCallback - обрабатывает ответ от Google и сохраняет пользователя
func HandleGoogleYoutubeCallback(w http.ResponseWriter, r *http.Request) {
	state := "google"
	if r.FormValue("state") != state {
		http.Error(w, "Invalid state", http.StatusUnauthorized)
		return
	}

	token, err := GoogleOauthConfig.Exchange(r.Context(), r.FormValue("code"))
	if err != nil {
		log.Printf("Error exchanging token: %v", err)
		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
		return
	}

	resp, err := http.Get("https://www.googleapis.com/oauth2/v2/userinfo?access_token=" + token.AccessToken)
	if err != nil {
		log.Printf("Error fetching user info: %v", err)
		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
		return
	}
	defer resp.Body.Close()

	content, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Printf("Error reading user info: %v", err)
		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
		return
	}

	var userInfo map[string]interface{}
	if err := json.Unmarshal(content, &userInfo); err != nil {
		log.Printf("Error parsing user info: %v", err)
		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
		return
	}

	email := userInfo["email"].(string)
	googleID := userInfo["id"].(string)
	firstName := userInfo["given_name"].(string)
	lastName := userInfo["family_name"].(string)

	// Проверка или создание пользователя
	var user users.User
	if err := config.DB.Where("email = ?", email).First(&user).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			user = users.User{
				Email:    email,
				Name:     firstName + " " + lastName,
				Provider: "google",
			}
			config.DB.Create(&user)
		} else {
			http.Error(w, "Database error", http.StatusInternalServerError)
			return
		}
	}

	// Проверка или обновление GoogleUser
	var youtubeUser users.YoutubeUser
	if err := config.DB.Where("google_id = ?", googleID).First(&youtubeUser).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			youtubeUser = users.YoutubeUser{
				UserID:       user.ID,
				GoogleID:     googleID,
				Email:        email,
				AccessToken:  token.AccessToken,
				RefreshToken: token.RefreshToken,
				Expiry:       token.Expiry,
			}
			config.DB.Create(&youtubeUser)
		} else {
			http.Error(w, "Database error", http.StatusInternalServerError)
			return
		}
	} else {
		youtubeUser.AccessToken = token.AccessToken
		youtubeUser.RefreshToken = token.RefreshToken
		youtubeUser.Expiry = token.Expiry
		config.DB.Save(&youtubeUser)
	}

	session, _ := store.Get(r, "session-name")
	session.Values["user_id"] = user.ID
	session.Save(r, w)

	http.Redirect(w, r, "/welcome", http.StatusTemporaryRedirect)
}

// RefreshToken обновляет токен YouTube API
func RefreshToken(user *users.YoutubeUser) (string, error) {
	if user.Expiry.After(time.Now()) {
		return user.AccessToken, nil
	}

	token := &oauth2.Token{
		RefreshToken: user.RefreshToken,
	}
	tokenSource := GoogleOauthConfig.TokenSource(context.Background(), token)
	newToken, err := tokenSource.Token()
	if err != nil {
		return "", err
	}

	user.AccessToken = newToken.AccessToken
	user.RefreshToken = newToken.RefreshToken
	user.Expiry = newToken.Expiry
	config.DB.Save(user)

	return newToken.AccessToken, nil
}

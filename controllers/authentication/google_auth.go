package authentication

import (
	"encoding/json"
	"github.com/gorilla/sessions"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"hired-valley-backend/config"
	"hired-valley-backend/models"
	"io/ioutil"
	"log"
	"net/http"
	"os"
)

var (
	googleOauthConfig = &oauth2.Config{
		RedirectURL:  os.Getenv("GOOGLE_REDIRECT_URL"),
		ClientID:     os.Getenv("GOOGLE_CLIENT_ID"),
		ClientSecret: os.Getenv("GOOGLE_CLIENT_SECRET"),
		Scopes:       []string{"https://www.googleapis.com/auth/userinfo.email", "https://www.googleapis.com/auth/userinfo.profile"},
		Endpoint:     google.Endpoint,
	}
	store = sessions.NewCookieStore([]byte("something-very-secret"))
)

func init() {
	// Проверка, что все переменные окружения заданы
	if googleOauthConfig.ClientID == "" || googleOauthConfig.ClientSecret == "" || googleOauthConfig.RedirectURL == "" {
		log.Fatal("Не установлены переменные окружения для Google OAuth")
	}
}

// HandleGoogleLogin initiates Google OAuth login
func HandleGoogleLogin(w http.ResponseWriter, r *http.Request) {
	state := "google"
	url := googleOauthConfig.AuthCodeURL(state)
	http.Redirect(w, r, url, http.StatusTemporaryRedirect)
}

// HandleGoogleCallback processes the OAuth callback and retrieves user info from Google
func HandleGoogleCallback(w http.ResponseWriter, r *http.Request) {
	state := "google"
	if r.FormValue("state") != state {
		log.Println("Invalid OAuth state")
		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
		return
	}

	token, err := googleOauthConfig.Exchange(r.Context(), r.FormValue("code"))
	if err != nil {
		log.Printf("Error while exchanging code for token: %s", err.Error())
		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
		return
	}

	resp, err := http.Get("https://www.googleapis.com/oauth2/v2/userinfo?access_token=" + token.AccessToken)
	if err != nil {
		log.Printf("Error while fetching user info: %s", err.Error())
		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
		return
	}
	defer resp.Body.Close()

	content, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Printf("Error reading response: %s", err.Error())
		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
		return
	}

	// Convert JSON response to structure
	var userInfo map[string]interface{}
	if err := json.Unmarshal(content, &userInfo); err != nil {
		log.Printf("Error parsing user info: %s", err.Error())
		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
		return
	}

	// Extract user info with type assertion
	googleID, ok := userInfo["id"].(string)
	if !ok {
		log.Println("Error extracting Google ID")
		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
		return
	}

	email, ok := userInfo["email"].(string)
	if !ok {
		log.Println("Error extracting email")
		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
		return
	}

	firstName, ok := userInfo["given_name"].(string)
	if !ok {
		log.Println("Error extracting first name")
		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
		return
	}

	lastName, ok := userInfo["family_name"].(string)
	if !ok {
		log.Println("Error extracting last name")
		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
		return
	}

	// Проверка, существует ли пользователь с таким email
	var user models.User
	if err := config.DB.Where("email = ?", email).First(&user).Error; err != nil {
		log.Printf("Error fetching user: %s", err.Error())
		http.Error(w, "Error fetching user", http.StatusInternalServerError)
		return
	}

	if user.ID == 0 {
		// Если пользователь не найден, создаем нового
		user = models.User{
			Email:       email,
			Name:        firstName + " " + lastName,
			Provider:    "google",
			AccessToken: token.AccessToken,
		}
		if err := config.DB.Create(&user).Error; err != nil {
			log.Printf("Error creating user: %s", err.Error())
			http.Error(w, "Error creating user", http.StatusInternalServerError)
			return
		}
	}

	// Проверка в таблице GoogleUser
	var googleUser models.GoogleUser
	if err := config.DB.Where("google_id = ?", googleID).First(&googleUser).Error; err != nil {
		log.Printf("Error fetching Google user: %s", err.Error())
		http.Error(w, "Error fetching Google user", http.StatusInternalServerError)
		return
	}

	if googleUser.GoogleID == "" {
		// Если GoogleUser не найден, создаем нового
		googleUser = models.GoogleUser{
			UserID:      user.ID,
			GoogleID:    googleID,
			Email:       email,
			FirstName:   firstName,
			LastName:    lastName,
			AccessToken: token.AccessToken,
		}
		if err := config.DB.Create(&googleUser).Error; err != nil {
			log.Printf("Error creating Google user: %s", err.Error())
			http.Error(w, "Error creating Google user", http.StatusInternalServerError)
			return
		}
	} else {
		// Обновляем информацию, если GoogleUser существует
		googleUser.Email = email
		googleUser.FirstName = firstName
		googleUser.LastName = lastName
		googleUser.AccessToken = token.AccessToken
		if err := config.DB.Save(&googleUser).Error; err != nil {
			log.Printf("Error updating Google user: %s", err.Error())
			http.Error(w, "Error updating Google user", http.StatusInternalServerError)
			return
		}
	}

	// Сохраняем данные пользователя в сессии
	session, _ := store.Get(r, "session-name")
	session.Values["user"] = user
	if err := session.Save(r, w); err != nil {
		log.Printf("Error saving session: %s", err.Error())
		http.Error(w, "Error saving session", http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/welcome", http.StatusTemporaryRedirect)
}

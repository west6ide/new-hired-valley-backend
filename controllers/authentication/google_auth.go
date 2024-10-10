package authentication

import (
	"encoding/json"
	"fmt"
	"github.com/gorilla/sessions"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"hired-valley-backend/config"
	"hired-valley-backend/models/authenticationUsers"
	"io/ioutil"
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
		fmt.Println("State is not valid")
		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
		return
	}

	token, err := googleOauthConfig.Exchange(oauth2.NoContext, r.FormValue("code"))
	if err != nil {
		fmt.Printf("Could not get token: %s\n", err.Error())
		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
		return
	}

	resp, err := http.Get("https://www.googleapis.com/oauth2/v2/userinfo?access_token=" + token.AccessToken)
	if err != nil {
		fmt.Printf("Could not create get request: %s\n", err.Error())
		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
		return
	}

	defer resp.Body.Close()
	content, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Printf("Could not parse response: %s\n", err.Error())
		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
		return
	}

	// Convert JSON response to structure
	var userInfo map[string]interface{}
	if err := json.Unmarshal(content, &userInfo); err != nil {
		fmt.Printf("Could not parse user info: %s\n", err.Error())
		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
		return
	}

	// Extract user info
	googleID := userInfo["id"].(string)
	email := userInfo["email"].(string)
	firstName := userInfo["given_name"].(string)
	lastName := userInfo["family_name"].(string)

	// Проверка, существует ли пользователь с таким email в таблице User
	var user authenticationUsers.User
	config.DB.Where("email = ?", email).First(&user)

	if user.ID == 0 {
		// Если пользователь не найден, создаем нового пользователя
		user = authenticationUsers.User{
			Email:       email,
			Name:        firstName + " " + lastName,
			Provider:    "google",
			AccessToken: token.AccessToken,
		}
		config.DB.Create(&user)
	}

	// Проверка в таблице GoogleUser
	var googleUser authenticationUsers.GoogleUser
	config.DB.Where("google_id = ?", googleID).First(&googleUser)

	if googleUser.GoogleID == "" {
		// Если GoogleUser не найден, создаем нового
		googleUser = authenticationUsers.GoogleUser{
			UserID:      user.ID, // Связь с таблицей User
			GoogleID:    googleID,
			Email:       email,
			FirstName:   firstName,
			LastName:    lastName,
			AccessToken: token.AccessToken,
		}
		config.DB.Create(&googleUser)
	} else {
		// Если GoogleUser найден, обновляем его информацию
		googleUser.Email = email
		googleUser.FirstName = firstName
		googleUser.LastName = lastName
		googleUser.AccessToken = token.AccessToken
		config.DB.Save(&googleUser)
	}

	// Сохраняем данные пользователя в сессии
	session, _ := store.Get(r, "session-name")
	session.Values["user"] = user
	session.Save(r, w)

	fmt.Fprintf(w, "Response: %s", content)
}

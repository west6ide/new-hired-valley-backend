package controllers

import (
	"encoding/json"
	"fmt"
	"github.com/gorilla/sessions"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"hired-valley-backend/config"
	"hired-valley-backend/models"
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

	// Save or update user info in database
	var user models.GoogleUser
	config.DB.Where("google_id = ?", googleID).First(&user)

	if user.GoogleID == "" {
		// User not found, create new
		user = models.GoogleUser{
			GoogleID:    googleID,
			Email:       email,
			FirstName:   firstName,
			LastName:    lastName,
			AccessToken: token.AccessToken,
		}
		config.DB.Create(&user)
	} else {
		// User found, update info
		user.Email = email
		user.FirstName = firstName
		user.LastName = lastName
		user.AccessToken = token.AccessToken
		config.DB.Save(&user)
	}

	// Save user info in session
	session, _ := store.Get(r, "session-name")
	session.Values["user"] = user
	session.Save(r, w)

	fmt.Fprintf(w, "Response: %s", content)
}

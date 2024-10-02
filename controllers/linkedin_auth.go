package controllers

import (
	"crypto/rand"
	"encoding/json"
	"fmt"
	"hired-valley-backend/config"
	"hired-valley-backend/models"
	"io/ioutil"
	"net/http"

	"github.com/gorilla/sessions"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/linkedin"
)

const (
	userInfoURL  = "https://api.linkedin.com/v2/me"
	emailInfoURL = "https://api.linkedin.com/v2/emailAddress?q=members&projection=(elements*(handle~))"
)

var (
	validPermissions = map[string]bool{
		"openid":  true,
		"profile": true,
		"email":   true,
	}
	authConf      *oauth2.Config
	storeLinkedin = sessions.NewCookieStore([]byte("golinkedinapi"))
)

// LinkedinProfile holds the user's profile data from LinkedIn
type LinkedinProfile struct {
	ProfileID  string `json:"id"`
	FirstName  string `json:"localizedFirstName"`
	LastName   string `json:"localizedLastName"`
	Email      string `json:"email" gorm:"not null;unique"`
	Headline   string `json:"headline"`
	Industry   string `json:"industry"`
	Summary    string `json:"summary"`
	PictureURL string `json:"profilePicture(displayImage~:playableStreams)"`
}

// ParseJSON converts a JSON string to a pointer to a LinkedinProfile.
func parseJSON(s string) (*LinkedinProfile, error) {
	linkedinProfile := &LinkedinProfile{}
	err := json.Unmarshal([]byte(s), linkedinProfile)
	if err != nil {
		return nil, err
	}
	return linkedinProfile, nil
}

// generateState generates a random state string for CSRF protection
func generateState() string {
	b := make([]byte, 32)
	rand.Read(b)
	return string(b)
}

// validState validates the state stored client-side with the request's state.
func validState(r *http.Request) bool {
	session, _ := storeLinkedin.Get(r, "golinkedinapi")
	retrievedState := session.Values["state"]
	return retrievedState == r.URL.Query().Get("state")
}

// GetLoginURL generates the LinkedIn login URL with state parameter.
func GetLoginURL(w http.ResponseWriter, r *http.Request) string {
	state := generateState()
	session, _ := storeLinkedin.Get(r, "golinkedinapi")
	session.Values["state"] = state
	session.Save(r, w)
	return authConf.AuthCodeURL(state)
}

// GetProfileData retrieves the user's LinkedIn profile data.
func GetProfileData(w http.ResponseWriter, r *http.Request) (*LinkedinProfile, error) {
	if !validState(r) {
		return nil, fmt.Errorf("invalid state")
	}

	params := r.URL.Query()
	tok, err := authConf.Exchange(oauth2.NoContext, params.Get("code"))
	if err != nil {
		return nil, err
	}
	client := authConf.Client(oauth2.NoContext, tok)

	// Get user profile data
	resp, err := client.Get(userInfoURL)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	data, _ := ioutil.ReadAll(resp.Body)
	profile, err := parseJSON(string(data))
	if err != nil {
		return nil, err
	}

	// Get user email
	respEmail, err := client.Get(emailInfoURL)
	if err != nil {
		return nil, err
	}
	defer respEmail.Body.Close()
	emailData, _ := ioutil.ReadAll(respEmail.Body)
	var email struct {
		Elements []struct {
			Handle struct {
				Email string `json:"emailAddress"`
			} `json:"handle~"`
		} `json:"elements"`
	}
	if err := json.Unmarshal(emailData, &email); err != nil {
		return nil, err
	}

	if len(email.Elements) > 0 {
		profile.Email = email.Elements[0].Handle.Email
	}

	return profile, nil
}

// InitConfig initializes the OAuth configuration for LinkedIn.
func InitConfig(clientID, clientSecret, redirectURL string) {
	authConf = &oauth2.Config{
		ClientID:     clientID,
		ClientSecret: clientSecret,
		Endpoint:     linkedin.Endpoint,
		RedirectURL:  redirectURL,
		Scopes:       []string{"openid", "profile", "email"},
	}
}

// HandleLinkedInLogin redirects the user to LinkedIn for authentication.
func HandleLinkedInLogin(w http.ResponseWriter, r *http.Request) {
	loginURL := GetLoginURL(w, r)
	http.Redirect(w, r, loginURL, http.StatusTemporaryRedirect)
}

func HandleLinkedInCallback(w http.ResponseWriter, r *http.Request) {
	if err := validState(r); err != true {
		http.Error(w, "Invalid state", http.StatusBadRequest)
		return
	}

	profile, err := GetProfileData(w, r)
	if err != nil {
		http.Error(w, fmt.Sprintf("Error retrieving user data from LinkedIn: %v", err), http.StatusInternalServerError)
		return
	}

	// Save user data in the database
	user := models.LinkedInUser{
		FirstName:  profile.FirstName,
		LastName:   profile.LastName,
		Email:      profile.Email,
		LinkedInID: profile.ProfileID,
	}

	var existingUser models.User
	if err := config.DB.Where("linked_in_id = ?", profile.ProfileID).First(&existingUser).Error; err == nil {
		config.DB.Model(&existingUser).Updates(user)
	} else {
		if err := config.DB.Create(&user).Error; err != nil {
			http.Error(w, fmt.Sprintf("Error saving user data: %v", err), http.StatusInternalServerError)
			return
		}
	}

	// Set session
	session, _ := config.Store.Get(r, "session-name")
	session.Values["user"] = user
	session.Save(r, w)

	fmt.Fprintf(w, "Hello, %s! Your LinkedIn ID: %s", profile.FirstName, profile.ProfileID)
}

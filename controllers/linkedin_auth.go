package controllers

import (
	"encoding/json"
	"fmt"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/linkedin"
	"hired-valley-backend/config"
	"hired-valley-backend/models"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"strings"

	_ "github.com/lib/pq"
)

var linkedinOAuthConfig = &oauth2.Config{
	ClientID:     os.Getenv("LINKEDIN_CLIENT_ID"),
	ClientSecret: os.Getenv("LINKEDIN_CLIENT_SECRET"),
	RedirectURL:  os.Getenv("LINKEDIN_REDIRECT_URL"),
	Scopes:       []string{"openid", "profile", "email"},
	Endpoint:     linkedin.Endpoint,
}

func HandleLinkedInHome(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "<a href='/login'>Login with LinkedIn</a>")
}

func HandleLinkedInLogin(w http.ResponseWriter, r *http.Request) {
	linkedinAuthURL := "https://www.linkedin.com/oauth/v2/authorization?response_type=code&client_id=" + linkedinOAuthConfig.ClientID + "&redirect_uri=" + url.QueryEscape(linkedinOAuthConfig.RedirectURL) + "&state=123456&scope=r_liteprofile%20r_emailaddress"
	http.Redirect(w, r, linkedinAuthURL, http.StatusTemporaryRedirect)
}

func HandleLinkedInCallback(w http.ResponseWriter, r *http.Request) {
	code := r.URL.Query().Get("code")
	if code == "" {
		http.Error(w, "Code not found", http.StatusBadRequest)
		return
	}
	token, err := getAccessToken(code)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	user, err := getUserInfo(token)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	saveUserToDB(user)
	fmt.Fprintf(w, "User: %s", user.Email)
}

func getAccessToken(code string) (string, error) {
	tokenURL := "https://www.linkedin.com/oauth/v2/accessToken"
	data := url.Values{}
	data.Set("grant_type", "authorization_code")
	data.Set("code", code)
	data.Set("redirect_uri", linkedinOAuthConfig.RedirectURL)
	data.Set("client_id", linkedinOAuthConfig.ClientID)
	data.Set("client_secret", linkedinOAuthConfig.ClientSecret)

	req, err := http.NewRequest("POST", tokenURL, strings.NewReader(data.Encode()))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	var tokenResponse struct {
		AccessToken string `json:"access_token"`
	}
	err = json.Unmarshal(body, &tokenResponse)
	if err != nil {
		return "", err
	}
	return tokenResponse.AccessToken, nil
}

func getUserInfo(accessToken string) (*models.LinkedInUser, error) {
	profileURL := "https://api.linkedin.com/v2/me"
	emailURL := "https://api.linkedin.com/v2/emailAddress?q=members&projection=(elements*(handle~))"

	req, err := http.NewRequest("GET", profileURL, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+accessToken)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	profileBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var profile struct {
		FirstName struct {
			Localized map[string]string `json:"localized"`
		} `json:"firstName"`
		LastName struct {
			Localized map[string]string `json:"lastName"`
		} `json:"lastName"`
		ID string `json:"id"`
	}
	err = json.Unmarshal(profileBody, &profile)
	if err != nil {
		return nil, err
	}

	req, err = http.NewRequest("GET", emailURL, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+accessToken)

	resp, err = client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	emailBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var emailResponse struct {
		Elements []struct {
			Handle struct {
				Email string `json:"emailAddress"`
			} `json:"handle~"`
		} `json:"elements"`
	}
	err = json.Unmarshal(emailBody, &emailResponse)
	if err != nil {
		return nil, err
	}

	user := &models.LinkedInUser{
		LinkedInID: profile.ID,
		FirstName:  profile.FirstName.Localized["en_US"],
		LastName:   profile.LastName.Localized["en_US"],
		Email:      emailResponse.Elements[0].Handle.Email,
	}
	return user, nil
}

func saveUserToDB(user *models.LinkedInUser) {
	// Сохраняем пользователя в базу данных с помощью GORM
	result := config.DB.FirstOrCreate(&user, models.LinkedInUser{LinkedInID: user.LinkedInID})
	if result.Error != nil {
		log.Println("Error saving user:", result.Error)
	}
}

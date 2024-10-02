package controllers

import (
	"bytes"
	"crypto/rand"
	"encoding/json"
	"fmt"
	"hired-valley-backend/config"
	"hired-valley-backend/models"
	"io/ioutil"
	"net/http"
	"net/url"

	"github.com/gorilla/sessions"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/linkedin"
)

const (
	// Используем правильный URL для получения данных профиля
	fullRequestURL  = "https://api.linkedin.com/v2/me"
	basicRequestURL = "https://api.linkedin.com/v2/me?projection=(id,firstName,lastName,headline,picture(displayImage~:playableStreams),industry,summary,emailAddress)"
)

var (
	validPermissions = map[string]bool{
		"openid":          true,
		"profile":         true,
		"w_member_social": true,
		"email":           true,
	}
	authConf      *oauth2.Config
	storeLinkedin = sessions.NewCookieStore([]byte("golinkedinapi"))
	requestedURL  string
)

// LinkedinProfile is used within this package as it is less useful than native types.
type LinkedinProfile struct {
	// ProfileID represents the Unique ID every Linkedin profile has.
	ProfileID string `json:"id"`
	// FirstName represents the user's first name.
	FirstName string `json:"first_name"`
	// LastName represents the user's last name.
	LastName string `json:"last-name"`
	// MaidenName represents the user's maiden name, if they have one.
	MaidenName string `json:"maiden-name"`
	// FormattedName represents the user's formatted name, based on locale.
	FormattedName string `json:"formatted-name"`
	// PhoneticFirstName represents the user's first name, spelled phonetically.
	PhoneticFirstName string `json:"phonetic-first-name"`
	// PhoneticFirstName represents the user's last name, spelled phonetically.
	PhoneticLastName string `json:"phonetic-last-name"`
	// Headline represents a short, attention grabbing description of the user.
	Headline string `json:"headline"`
	// Location represents where the user is located.
	Location Location `json:"location"`
	// Industry represents what industry the user is working in.
	Industry string `json:"industry"`
	// CurrentShare represents the user's last shared post.
	CurrentShare string `json:"current-share"`
	// NumConnections represents the user's number of connections, up to a maximum of 500.
	// The user's connections may be over this, however it will not be shown. (i.e. 500+ connections)
	NumConnections int `json:"num-connections"`
	// IsConnectionsCapped represents whether or not the user's connections are capped.
	IsConnectionsCapped bool `json:"num-connections-capped"`
	// Summary represents a long-form text description of the user's capabilities.
	Summary string `json:"summary"`
	// Specialties is a short-form text description of the user's specialties.
	Specialties string `json:"specialties"`
	// Positions is a Positions struct that describes the user's previously held positions.
	Positions Positions `json:"positions"`
	// PictureURL represents a URL pointing to the user's profile picture.
	PictureURL string `json:"picture-url"`
	// EmailAddress represents the user's e-mail address, however you must specify 'r_emailaddress'
	// to be able to retrieve this.
	EmailAddress string `json:"email-address"`
}

// Positions represents the result given by json:"positions"
type Positions struct {
	total  int
	values []Position
}

// Location specifies the users location
type Location struct {
	UserLocation string
	CountryCode  string
}

// Position represents a job held by the authorized user.
type Position struct {
	// ID represents a unique ID representing the position
	ID string
	// Title represents a user's position's title, for example Jeff Bezos's title would be 'CEO'
	Title string
	// Summary represents a short description of the user's position.
	Summary string
	// StartDate represents when the user's position started.
	StartDate string
	// EndDate represents the user's position's end date, if any.
	EndDate string
	// IsCurrent represents if the position is currently held or not.
	// If this is false, EndDate will not be returned, and will therefore equal ""
	IsCurrent bool
	// Company represents the Company where the user is employed.
	Company PositionCompany
}

// PositionCompany represents a company that is described within a user's Profile.
// This is different from Company, which fully represents a company's data.
type PositionCompany struct {
	// ID represents a unique ID representing the company
	ID string
	// Name represents the name of the company
	Name string
	// Type represents the type of the company, either 'public' or 'private'
	Type string
	// Industry represents which industry the company is in.
	Industry string
	// Ticker represents the stock market ticker symbol of the company.
	// This will be blank if the company is privately held.
	Ticker string
}

// ParseJSON converts a JSON string to a pointer to a LinkedinProfile.
func parseJSON(s string) (*LinkedinProfile, error) {
	linkedinProfile := &LinkedinProfile{}
	bytes := bytes.NewBuffer([]byte(s))
	err := json.NewDecoder(bytes).Decode(linkedinProfile)
	if err != nil {
		return nil, err
	}
	return linkedinProfile, nil
}

// generateState generates a random set of bytes to ensure state is preserved.
// This prevents such things as XSS occuring.
func generateState() string {
	b := make([]byte, 32)
	rand.Read(b)
	return string(b)
}

// getSessionValue grabs the value of an interface in this case being the session.Values["string"]
// This will return "" if f is nil.
func getSessionValue(f interface{}) string {
	if f != nil {
		if foo, ok := f.(string); ok {
			return foo
		}
	}
	return ""
}

// validState validates the state stored client-side with the request's state.
func validState(r *http.Request) bool {
	session, _ := storeLinkedin.Get(r, "golinkedinapi")
	retrievedState := session.Values["state"]
	return getSessionValue(retrievedState) == r.URL.Query().Get("state")
}

// GetLoginURL provides a state-specific login URL for the user to login to.
func GetLoginURL(w http.ResponseWriter, r *http.Request) string {
	state := generateState()
	session, _ := storeLinkedin.Get(r, "golinkedinapi")
	session.Values["state"] = state
	defer session.Save(r, w)
	return authConf.AuthCodeURL(state)
}

// GetProfileData gather's the user's Linkedin profile data and returns it as a pointer to a LinkedinProfile struct.
// CAUTION: GetLoginURL must be called before this, as GetProfileData() has a state check.
func GetProfileData(w http.ResponseWriter, r *http.Request) (*LinkedinProfile, error) {
	if validState(r) == false {
		err := fmt.Errorf("State comparison failed")
		return &LinkedinProfile{}, err
	}
	params := r.URL.Query()
	// Authenticate
	tok, err := authConf.Exchange(oauth2.NoContext, params.Get("code"))
	if err != nil {
		return &LinkedinProfile{}, err
	}
	client := authConf.Client(oauth2.NoContext, tok)
	// Retrieve data
	resp, err := client.Get(fullRequestURL)
	if err != nil {
		return &LinkedinProfile{}, err
	}
	// Store data to struct and return.
	data, _ := ioutil.ReadAll(resp.Body)
	defer resp.Body.Close()
	formattedData, err := parseJSON(string(data))
	if err != nil {
		return &LinkedinProfile{}, err
	}
	return formattedData, nil
}

// InitConfig initializes the config needed by the client.
// permissions is a string of all scopes desired by the user.
func InitConfig(permissions []string, clientID string, clientSecret string, redirectURL string) {
	var isEmail, isBasic bool
	if permissions == nil {
		panic(fmt.Errorf("You must specify some scope to request"))
	}
	for _, elem := range permissions {
		if isEmail && isBasic {
			requestedURL = fullRequestURL + ",r_emailaddress"
		} else if isBasic {
			requestedURL = basicRequestURL
		}
		if validPermissions[elem] != true {
			panic(fmt.Errorf("All elements of permissions must be valid Linkedin permissions as specified in the API docs"))
		}
	}
	_, err := url.ParseRequestURI(redirectURL)
	if err != nil {
		panic(fmt.Errorf("redirectURL specified must be a valid FQDN. Please ensure you added https:// to the front"))
	}
	authConf = &oauth2.Config{ClientID: clientID,
		ClientSecret: clientSecret,
		Endpoint:     linkedin.Endpoint,
		RedirectURL:  redirectURL,
		Scopes:       permissions,
	}
	if isEmail && isBasic {
		requestedURL = fullRequestURL
	} else if isBasic {
		requestedURL = basicRequestURL
	}
}

// TODO: Config Struct
// HandleLinkedInLogin перенаправляет пользователя на страницу авторизации LinkedIn.
func HandleLinkedInLogin(w http.ResponseWriter, r *http.Request) {
	// Получаем URL для авторизации через LinkedIn
	loginURL := GetLoginURL(w, r)
	// Перенаправляем пользователя на страницу LinkedIn для авторизации
	http.Redirect(w, r, loginURL, http.StatusTemporaryRedirect)
}

func HandleLinkedInCallback(w http.ResponseWriter, r *http.Request) {
	code := r.URL.Query().Get("code")
	state := r.URL.Query().Get("state")
	if code == "" || state == "" {
		http.Error(w, "Отсутствует параметр 'code' или 'state'", http.StatusBadRequest)
		return
	}

	fmt.Println("Получен код:", code)
	fmt.Println("Состояние из запроса:", state)

	if !validState(r) {
		http.Error(w, "Недействительное состояние", http.StatusBadRequest)
		return
	}

	profile, err := GetProfileData(w, r)
	if err != nil {
		http.Error(w, fmt.Sprintf("Ошибка при получении данных пользователя через LinkedIn: %v", err), http.StatusInternalServerError)
		return
	}

	// Сохранение данных пользователя в базу данных
	user := models.LinkedInUser{
		FirstName:  profile.FirstName,
		LastName:   profile.LastName,
		Email:      profile.EmailAddress,
		LinkedInID: profile.ProfileID,
	}

	// Попробуйте найти существующего пользователя по LinkedIn ID
	var existingUser models.User
	if err := config.DB.Where("linked_in_id = ?", profile.ProfileID).First(&existingUser).Error; err == nil {
		// Пользователь уже существует, обновите его данные
		config.DB.Model(&existingUser).Updates(user)
	} else {
		// Пользователь не найден, создайте нового
		if err := config.DB.Create(&user).Error; err != nil {
			http.Error(w, fmt.Sprintf("Ошибка при сохранении данных пользователя: %v", err), http.StatusInternalServerError)
			return
		}
	}

	// Установите сессию
	session, _ := config.Store.Get(r, "session-name")
	session.Values["user"] = user
	session.Save(r, w)

	fmt.Fprintf(w, "Привет, %s! Ваш LinkedIn ID: %s", profile.FirstName, profile.ProfileID)
}

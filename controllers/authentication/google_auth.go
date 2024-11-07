package authentication

import (
	"encoding/json"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/session"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"gorm.io/gorm"
	"hired-valley-backend/config"
	"hired-valley-backend/models/users"
	"io/ioutil"
	"log"
	"os"
)

// Настройка конфигурации OAuth для Google
var googleOauthConfig = &oauth2.Config{
	RedirectURL:  os.Getenv("GOOGLE_REDIRECT_URL"),
	ClientID:     os.Getenv("GOOGLE_CLIENT_ID"),
	ClientSecret: os.Getenv("GOOGLE_CLIENT_SECRET"),
	Scopes:       []string{"https://www.googleapis.com/auth/userinfo.email", "https://www.googleapis.com/auth/userinfo.profile"},
	Endpoint:     google.Endpoint,
}

// Инициализация менеджера сессий Fiber
var store = session.New()

// Инициализация переменных окружения
func init() {
	if googleOauthConfig.ClientID == "" || googleOauthConfig.ClientSecret == "" || googleOauthConfig.RedirectURL == "" {
		log.Fatal("Не установлены переменные окружения для Google OAuth")
	}
}

// HandleGoogleLogin инициирует авторизацию через Google OAuth
func HandleGoogleLogin(c *fiber.Ctx) error {
	state := "google" // Используем простой state для примера, лучше генерировать случайное значение
	url := googleOauthConfig.AuthCodeURL(state)
	return c.Redirect(url)
}

// HandleGoogleCallback обрабатывает коллбэк OAuth и получает информацию о пользователе от Google
func HandleGoogleCallback(c *fiber.Ctx) error {
	state := "google"
	if c.Query("state") != state {
		log.Println("Invalid OAuth state")
		return c.Redirect("/")
	}

	token, err := googleOauthConfig.Exchange(c.Context(), c.Query("code"))
	if err != nil {
		log.Printf("Error while exchanging code for token: %s", err.Error())
		return c.Redirect("/")
	}

	// Запрос информации о пользователе
	resp, err := googleOauthConfig.Client(c.Context(), token).Get("https://www.googleapis.com/oauth2/v2/userinfo")
	if err != nil {
		log.Printf("Error while fetching user info: %s", err.Error())
		return c.Redirect("/")
	}
	defer resp.Body.Close()

	content, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Printf("Error reading response: %s", err.Error())
		return c.Redirect("/")
	}

	var userInfo map[string]interface{}
	if err := json.Unmarshal(content, &userInfo); err != nil {
		log.Printf("Error parsing user info: %s", err.Error())
		return c.Redirect("/")
	}

	googleID, ok := userInfo["id"].(string)
	if !ok {
		log.Println("Error extracting Google ID")
		return c.Redirect("/")
	}

	email, ok := userInfo["email"].(string)
	if !ok {
		log.Println("Error extracting email")
		return c.Redirect("/")
	}

	firstName, ok := userInfo["given_name"].(string)
	if !ok {
		log.Println("Error extracting first name")
		return c.Redirect("/")
	}

	lastName, ok := userInfo["family_name"].(string)
	if !ok {
		log.Println("Error extracting last name")
		return c.Redirect("/")
	}

	// Проверка пользователя в базе данных
	var user users.User
	if err := config.DB.Where("email = ?", email).First(&user).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			log.Printf("User with email %s not found, creating a new one", email)
			user = users.User{
				Email:       email,
				Name:        firstName + " " + lastName,
				Provider:    "google",
				AccessToken: token.AccessToken,
			}
			if err := config.DB.Create(&user).Error; err != nil {
				log.Printf("Error creating user: %v", err)
				return c.Status(fiber.StatusInternalServerError).SendString("Error creating user")
			}
		} else {
			log.Printf("Error finding user with email %s: %v", email, err)
			return c.Status(fiber.StatusInternalServerError).SendString("Error finding user")
		}
	}

	// Проверка или создание GoogleUser
	var googleUser users.GoogleUser
	if err := config.DB.Where("google_id = ?", googleID).First(&googleUser).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			log.Printf("GoogleUser with ID %s not found, creating a new one", googleID)
			googleUser = users.GoogleUser{
				UserID:      user.ID,
				GoogleID:    googleID,
				Email:       email,
				FirstName:   firstName,
				LastName:    lastName,
				AccessToken: token.AccessToken,
			}
			if err := config.DB.Create(&googleUser).Error; err != nil {
				log.Printf("Error creating GoogleUser: %v", err)
				return c.Status(fiber.StatusInternalServerError).SendString("Error creating GoogleUser")
			}
		} else {
			log.Printf("Error finding GoogleUser with Google ID %s: %v", googleID, err)
			return c.Status(fiber.StatusInternalServerError).SendString("Error finding GoogleUser")
		}
	} else {
		googleUser.Email = email
		googleUser.FirstName = firstName
		googleUser.LastName = lastName
		googleUser.AccessToken = token.AccessToken
		if err := config.DB.Save(&googleUser).Error; err != nil {
			log.Printf("Error updating GoogleUser: %v", err)
			return c.Status(fiber.StatusInternalServerError).SendString("Error updating GoogleUser")
		}
	}

	// Создание сессии пользователя
	sess, err := store.Get(c)
	if err != nil {
		log.Printf("Error getting session: %s", err.Error())
		return c.Status(fiber.StatusInternalServerError).SendString("Error getting session")
	}

	sess.Set("user_id", user.ID)
	if err := sess.Save(); err != nil {
		log.Printf("Error saving session: %s", err.Error())
		return c.Status(fiber.StatusInternalServerError).SendString("Error saving session")
	}

	// Перенаправляем пользователя на защищенную страницу
	return c.Redirect("/welcome")
}

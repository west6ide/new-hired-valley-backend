package controllers

import (
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"github.com/dgrijalva/jwt-go"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"gorm.io/gorm"
	"hired-valley-backend/config"
	"hired-valley-backend/models"
	"io"
	"net/http"
	"os"
	"time"
)

var jwtKeyGoogle = []byte(os.Getenv("JWT_GOOGLE_SECRET")) // Инициализация jwtKey

// GenerateJWT создает JWT токен для пользователя
func GenerateJWT(email string) (string, error) {
	expirationTime := time.Now().Add(24 * time.Hour) // Токен будет действителен 24 часа
	claims := &Claims{
		Email: email,
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: expirationTime.Unix(),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(jwtKeyGoogle)
	if err != nil {
		return "", err
	}
	return tokenString, nil
}

var googleOauthConfig = &oauth2.Config{
	RedirectURL:  os.Getenv("GOOGLE_REDIRECT_URL"),
	ClientID:     os.Getenv("GOOGLE_CLIENT_ID"),
	ClientSecret: os.Getenv("GOOGLE_CLIENT_SECRET"),
	Scopes:       []string{"https://www.googleapis.com/auth/userinfo.email", "https://www.googleapis.com/auth/userinfo.profile"},
	Endpoint:     google.Endpoint,
}

// Генерация случайного состояния (state) для защиты от CSRF-атак
func generateState() (string, error) {
	b := make([]byte, 32)
	_, err := rand.Read(b)
	if err != nil {
		return "", err
	}
	return base64.URLEncoding.EncodeToString(b), nil
}

// Инициация авторизации через Google
func HandleGoogleLogin(w http.ResponseWriter, r *http.Request) {
	state, err := generateState()
	if err != nil {
		http.Error(w, "Error generating state", http.StatusInternalServerError)
		return
	}

	// Сохраняем state в cookie
	http.SetCookie(w, &http.Cookie{
		Name:     "oauth_state",
		Value:    state,
		Expires:  time.Now().Add(10 * time.Minute),
		HttpOnly: true,
		Secure:   os.Getenv("ENV") == "production", // Установить Secure: true для продакшена
		SameSite: http.SameSiteLaxMode,
	})

	url := googleOauthConfig.AuthCodeURL(state, oauth2.AccessTypeOffline, oauth2.SetAuthURLParam("prompt", "select_account"))
	http.Redirect(w, r, url, http.StatusTemporaryRedirect)
}

// Callback после авторизации Google
func HandleGoogleCallback(w http.ResponseWriter, r *http.Request) {
	// Проверяем сохраненное состояние
	stateCookie, err := r.Cookie("oauth_state")
	if err != nil || r.FormValue("state") != stateCookie.Value {
		http.Error(w, "Invalid state parameter", http.StatusBadRequest)
		return
	}

	// Обмен кода на токен
	token, err := googleOauthConfig.Exchange(r.Context(), r.FormValue("code"))
	if err != nil {
		http.Error(w, "Failed to exchange token: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Получение информации о пользователе с помощью токена
	resp, err := http.Get("https://www.googleapis.com/oauth2/v2/userinfo?access_token=" + token.AccessToken)
	if err != nil {
		http.Error(w, "Failed to get user info", http.StatusInternalServerError)
		return
	}
	defer resp.Body.Close()

	content, err := io.ReadAll(resp.Body) // Используем io.ReadAll вместо устаревшего ioutil.ReadAll
	if err != nil {
		http.Error(w, "Failed to read user info", http.StatusInternalServerError)
		return
	}

	// Десериализация данных пользователя
	var googleUser models.GoogleUser
	if err := json.Unmarshal(content, &googleUser); err != nil {
		http.Error(w, "Failed to parse user info", http.StatusInternalServerError)
		return
	}

	// Сохранение пользователя в базу данных или обновление существующего
	var existingUser models.GoogleUser
	result := config.DB.Where("email = ?", googleUser.Email).First(&existingUser)

	if result.Error == gorm.ErrRecordNotFound {
		if err := config.DB.Create(&googleUser).Error; err != nil { // Проверяем ошибки создания записи
			http.Error(w, "Failed to save user", http.StatusInternalServerError)
			return
		}
	} else if result.Error != nil {
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	} else {
		existingUser.FirstName = googleUser.FirstName
		existingUser.LastName = googleUser.LastName
		if err := config.DB.Save(&existingUser).Error; err != nil { // Проверяем ошибки обновления
			http.Error(w, "Failed to update user", http.StatusInternalServerError)
			return
		}
		googleUser = existingUser
	}

	// Генерация JWT токена
	jwtToken, err := GenerateJWT(googleUser.Email)
	if err != nil {
		http.Error(w, "Failed to generate token", http.StatusInternalServerError)
		return
	}

	// Возвращаем данные пользователя и токен в ответе
	response := map[string]interface{}{
		"token": jwtToken,
		"user":  googleUser,
	}

	// Устанавливаем заголовки и отправляем JSON-ответ
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

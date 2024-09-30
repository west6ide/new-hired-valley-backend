package main

import (
	"fmt"
	"hired-valley-backend/config"
	"hired-valley-backend/controllers"
	"hired-valley-backend/controllers/httpCors"
	"hired-valley-backend/models"
	"net/http"
	"os"

	"github.com/markbates/goth/gothic"
)

func main() {
	// Получение порта из переменных окружения
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080" // Значение по умолчанию
	}

	// Инициализация базы данных
	config.InitDB()
	err := config.DB.AutoMigrate(&models.GoogleUser{}, &models.User{}, &models.LinkedInUser{})
	if err != nil {
		return
	} // Добавление миграции для LinkedInUser

	// Инициализация провайдеров авторизации

	// Настройка маршрутов
	http.HandleFunc("/", handleHome)
	http.HandleFunc("/login/google", controllers.HandleGoogleLogin)
	http.HandleFunc("/callback/google", controllers.HandleGoogleCallback)
	http.HandleFunc("/auth/linkedin", handleLinkedInLogin)             // Маршрут для авторизации через LinkedIn
	http.HandleFunc("/auth/linkedin/callback", handleLinkedInCallback) // Маршрут для обратного вызова LinkedIn
	http.HandleFunc("/register", controllers.Register)
	http.HandleFunc("/login", controllers.Login)
	http.HandleFunc("/api/profile", controllers.GetProfile)
	http.HandleFunc("/api/logout", controllers.Logout)

	// Маршрут для обработки запросов к провайдерам Goth
	http.HandleFunc("/auth/", gothic.BeginAuthHandler)

	// Запуск сервера
	c := httpCors.CorsSettings()
	handler := c.Handler(http.DefaultServeMux)
	http.ListenAndServe(":"+port, handler)
}

// Обработчик домашней страницы
func handleHome(w http.ResponseWriter, r *http.Request) {
	session, _ := config.Store.Get(r, "session-name")
	user := session.Values["user"]

	if user != nil {
		usr := user.(models.GoogleUser)
		html := fmt.Sprintf(`<html><body>
                   <p>Добро пожаловать, %s!</p>
                   <a href="/logout">Выйти</a><br>
                   <form action="/google-logout" method="post">
                       <button type="submit">Выйти из Google</button>
                   </form>
                 </body></html>`, usr.FirstName)
		fmt.Fprint(w, html)
	} else {
		html := `<html><body>
                   <a href="/login/google">Войти через Google</a><br>
                   <a href="/auth/linkedin">Войти через LinkedIn</a>
                 </body></html>`
		fmt.Fprint(w, html)
	}
}

// Обработчик для начала авторизации через LinkedIn
func handleLinkedInLogin(w http.ResponseWriter, r *http.Request) {
	gothic.BeginAuthHandler(w, r) // Использует провайдера LinkedIn по маршруту
}

// Обработчик для получения токена и данных пользователя после обратного вызова
func handleLinkedInCallback(w http.ResponseWriter, r *http.Request) {
	// Завершение процесса авторизации и получение данных пользователя
	user, err := gothic.CompleteUserAuth(w, r)
	if err != nil {
		http.Error(w, "Ошибка при завершении авторизации: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Создание структуры LinkedInUser из полученных данных
	linkedInUser := models.LinkedInUser{
		FirstName: user.FirstName,
		LastName:  user.LastName,
		Email:     user.Email,
	}

	// Сохранение данных пользователя в базу данных
	if err := config.DB.Create(&linkedInUser).Error; err != nil {
		http.Error(w, "Ошибка при сохранении данных пользователя в базу: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Сохранение токена в базу данных
	oauthToken := models.OAuthToken{
		UserID:      linkedInUser.ID,
		AccessToken: user.AccessToken,
		TokenType:   "Bearer",
	}

	if err := config.DB.Create(&oauthToken).Error; err != nil {
		http.Error(w, "Ошибка при сохранении токена в базу: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Отображение данных пользователя
	fmt.Fprintf(w, "Добро пожаловать, %s %s! Ваш email: %s", linkedInUser.FirstName, linkedInUser.LastName, linkedInUser.Email)
}

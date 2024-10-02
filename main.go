package main

import (
	"fmt"
	"hired-valley-backend/config"
	"hired-valley-backend/controllers"
	"hired-valley-backend/controllers/httpCors"
	"hired-valley-backend/models"
	"net/http"
	"os"
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

	// Инициализация конфигурации OAuth для LinkedIn
	clientID := os.Getenv("LINKEDIN_CLIENT_ID")
	clientSecret := os.Getenv("LINKEDIN_CLIENT_SECRET")
	redirectURL := os.Getenv("LINKEDIN_REDIRECT_URL")                        // URL для редиректа после успешной авторизации
	permissions := []string{"openid", "profile", "w_member_social", "email"} // Необходимые разрешения

	controllers.InitConfig(permissions, clientID, clientSecret, redirectURL)

	// Настройка маршрутов
	http.HandleFunc("/", handleHome)
	http.HandleFunc("/login/google", controllers.HandleGoogleLogin)
	http.HandleFunc("/callback/google", controllers.HandleGoogleCallback)

	// Маршруты для LinkedIn OAuth
	http.HandleFunc("/login/linkedin", controllers.HandleLinkedInLogin)
	http.HandleFunc("/callback/linkedin", controllers.HandleLinkedInCallback)

	http.HandleFunc("/register", controllers.Register)
	http.HandleFunc("/login", controllers.Login)
	http.HandleFunc("/api/profile", controllers.GetProfile)
	http.HandleFunc("/api/logout", controllers.Logout)

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
                   <a href="/login/linkedin">Войти через LinkedIn</a>
                 </body></html>`
		fmt.Fprint(w, html)
	}
}

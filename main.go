package main

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"hired-valley-backend/config"
	"hired-valley-backend/controllers"
	"hired-valley-backend/controllers/httpCors"
	"hired-valley-backend/models"
	"hired-valley-backend/routes"
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
	config.DB.AutoMigrate(&models.GoogleUser{}, &models.User{}, &models.LinkedInUser{}) // Добавление миграции для LinkedInUser

	r := gin.Default()
	routes.RegisterStoryRoutes(r, config.DB)

	// Настройка маршрутов
	http.HandleFunc("/", handleHome)
	http.HandleFunc("/login/google", controllers.HandleGoogleLogin)
	http.HandleFunc("/callback/google", controllers.HandleGoogleCallback)
	http.HandleFunc("/login/linkedin", controllers.HandleLinkedInLogin)       // LinkedIn login
	http.HandleFunc("/callback/linkedin", controllers.HandleLinkedInCallback) // LinkedIn callback
	http.HandleFunc("/register", controllers.Register)
	http.HandleFunc("/login", controllers.Login)
	http.HandleFunc("/api/profile", controllers.GetProfile)
	http.HandleFunc("/api/logout", controllers.Logout)

	// Запуск сервера
	c := httpCors.CorsSettings()
	handler := c.Handler(http.DefaultServeMux)
	http.ListenAndServe(":8080", handler)
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

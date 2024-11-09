package main

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"hired-valley-backend/config"
	"hired-valley-backend/controllers/authentication"
	"hired-valley-backend/controllers/mentors"
	"hired-valley-backend/models/courses"
	"hired-valley-backend/models/users"
	"log"
	"net/http"
	"os"
	"time"
)

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080" // Устанавливаем порт по умолчанию
	}

	r := gin.Default()

	// Инициализируем базу данных
	err := config.InitDB()
	if err != nil {
		log.Fatalf("Ошибка инициализации базы данных: %v", err)
	}

	// Выполняем миграцию базы данных
	err = config.DB.AutoMigrate(
		&users.User{},
		&users.GoogleUser{},
		&users.LinkedInUser{},
		&courses.Course{},
		&courses.Lesson{},
		&users.Story{},
		&users.MentorProfile{},
		&users.MentorSkill{},
		&users.AvailableTime{},
		&users.SocialLinks{},
	)
	if err != nil {
		log.Fatalf("Ошибка миграции базы данных: %v", err)
	}

	// Проверка подключения к базе данных
	sqlDB, err := config.DB.DB()
	if err != nil {
		log.Fatalf("Ошибка получения подключения к базе данных: %v", err)
	}

	err = sqlDB.Ping()
	if err != nil {
		log.Fatalf("Ошибка подключения к базе данных: %v", err)
	} else {
		log.Println("Подключение к базе данных успешно")
	}

	r.GET("/", handleHome)
	//http.HandleFunc("/login/google", authentication.HandleGoogleLogin)
	//http.HandleFunc("/callback/google", authentication.HandleGoogleCallback)
	//http.HandleFunc("/login/linkedin", authentication.HandleLinkedInLogin)
	//http.HandleFunc("/callback/linkedin", authentication.HandleLinkedInCallback)

	// Защищенные маршруты с AuthMiddleware

	r.POST("/register", authentication.Register)
	r.POST("/login", authentication.Login)
	r.GET("/profile", authentication.GetProfile)
	r.POST("/logout", authentication.Logout)

	auth := r.Group("/")
	auth.Use(authentication.AuthMiddleware())
	//http.HandleFunc("/profile/update", authentication.UpdateProfile)
	//http.HandleFunc("/users/search", authentication.SearchUsers)
	//
	//http.HandleFunc("/list/courses", course.ListCourses)
	//http.HandleFunc("/create/courses", course.CreateCourse)
	//http.HandleFunc("/upload/video", course.UploadVideo)
	//http.HandleFunc("/list/lessons", course.ListLessons)
	//http.HandleFunc("/create/lessons", course.CreateLesson)
	//
	//http.HandleFunc("/create/stories", controllers.CreateStory)
	//http.HandleFunc("/list/stories", controllers.GetActiveStories)
	//http.HandleFunc("/stories/archive", controllers.ArchiveStory) // Параметр id передается как query параметр
	auth.POST("/mentors", mentors.CreateMentorProfile)
	auth.GET("/mentors/:id", mentors.GetMentorProfile)
	auth.PUT("/mentors/:id", mentors.UpdateMentorProfile)
	auth.DELETE("/mentors/:id", mentors.DeleteMentorProfile)

	// CRUD для AvailableTime (изменены маршруты, чтобы избежать конфликта)
	auth.POST("/mentors/:id/availability", mentors.AddAvailableTime)
	auth.GET("/mentors/:id/availability", mentors.GetAvailableTimes)
	auth.PUT("/availability/:id", mentors.UpdateAvailableTime)
	auth.DELETE("/availability/:id", mentors.DeleteAvailableTime)

	//// Запускаем сервер
	//log.Printf("Сервер запущен на порту %s", port)
	//err = http.ListenAndServe(":"+port, nil)
	//if err != nil {
	//	log.Fatalf("Ошибка запуска сервера: %v", err)
	//}

	r.Run(":" + port)
}

// Обработчик домашней страницы
func handleHome(c *gin.Context) {
	session, _ := config.Store.Get(c.Request, "session-name")
	user := session.Values["user"]

	if user != nil {
		switch usr := user.(type) {
		case users.GoogleUser:
			html := fmt.Sprintf(`<html><body>
				<p>Добро пожаловать, %s!</p>
				<a href="/logout">Выйти</a><br>
			</body></html>`, usr.FirstName)
			c.Data(http.StatusOK, "text/html; charset=utf-8", []byte(html))
		case users.LinkedInUser:
			html := fmt.Sprintf(`<html><body>
				<p>Добро пожаловать, %s!</p>
				<a href="/logout">Выйти</a><br>
			</body></html>`, usr.FirstName)
			c.Data(http.StatusOK, "text/html; charset=utf-8", []byte(html))
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Неизвестный тип пользователя"})
		}
	} else {
		html := `<html><body>
                   <a href="/login/google">Войти через Google</a><br>
                   <a href="/login/linkedin">Войти через LinkedIn</a>
                 </body></html>`
		c.Data(http.StatusOK, "text/html; charset=utf-8", []byte(html))
	}
}

func RemoveExpiredStories() {
	for {
		config.DB.Where("expires_at <= ?", time.Now()).Delete(&users.Story{})
		time.Sleep(1 * time.Hour) // Запуск каждые 1 час
	}
}

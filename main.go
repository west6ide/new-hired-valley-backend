package main

import (
	"fmt"
	"hired-valley-backend/config"
	"hired-valley-backend/controllers/authentication"
	"hired-valley-backend/controllers/courses"
	"hired-valley-backend/models"
	"log"
	"net/http"
	"os"
)

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080" // Устанавливаем порт по умолчанию
	}

	// Инициализируем базу данных
	err := config.InitDB()
	if err != nil {
		log.Fatalf("Ошибка инициализации базы данных: %v", err)
	}

	// Выполняем миграцию базы данных
	err = config.DB.AutoMigrate(
		&models.GoogleUser{},
		&models.User{},
		&models.LinkedInUser{},
		&models.Course{},
		&models.Module{},
		&models.Review{},
		&models.Progress{},
	)
	if err != nil {
		log.Fatalf("Ошибка миграции базы данных: %v", err)
	}

	// Настраиваем маршруты
	http.HandleFunc("/", handleHome)
	http.HandleFunc("/login/google", authentication.HandleGoogleLogin)
	http.HandleFunc("/callback/google", authentication.HandleGoogleCallback)
	http.HandleFunc("/login/linkedin", authentication.HandleLinkedInLogin)
	http.HandleFunc("/callback/linkedin", authentication.HandleLinkedInCallback)

	http.HandleFunc("/register", authentication.Register)
	http.HandleFunc("/login", authentication.Login)
	http.HandleFunc("/api/profile", authentication.GetProfile)
	http.HandleFunc("/api/logout", authentication.Logout)
	http.HandleFunc("/change-password", authentication.ChangePassword)

	http.HandleFunc("/courses", courses.GetCourses)
	http.HandleFunc("/courses/create", courses.CreateCourse)
	http.HandleFunc("/courses/{id}", courses.GetCourseByID)
	http.HandleFunc("/courses/{id}/update", courses.UpdateCourse)
	http.HandleFunc("/courses/{id}/delete", courses.DeleteCourse)

	http.HandleFunc("/courses/{course_id}/modules", courses.CreateModule)
	http.HandleFunc("/modules/{id}", courses.UpdateModule)

	// Запускаем сервер
	log.Printf("Сервер запущен на порту %s", port)
	err = http.ListenAndServe(":"+port, nil)
	if err != nil {
		log.Fatalf("Ошибка запуска сервера: %v", err)
	}
}

func handleHome(w http.ResponseWriter, r *http.Request) {
	session, _ := config.Store.Get(r, "session-name")
	user := session.Values["user"]

	if user != nil {
		switch usr := user.(type) {
		case models.GoogleUser:
			html := fmt.Sprintf(`<html><body>
				<p>Добро пожаловать, %s!</p>
				<a href="/logout">Выйти</a><br>
				<form action="/google-logout" method="post">
					<button type="submit">Выйти из Google</button>
				</form>
			</body></html>`, usr.FirstName)
			fmt.Fprint(w, html)
		case models.LinkedInUser:
			html := fmt.Sprintf(`<html><body>
				<p>Добро пожаловать, %s!</p>
				<a href="/logout">Выйти</a><br>
				<form action="/linkedin-logout" method="post">
					<button type="submit">Выйти из LinkedIn</button>
				</form>
			</body></html>`, usr.FirstName)
			fmt.Fprint(w, html)
		default:
			http.Error(w, "Неизвестный тип пользователя", http.StatusInternalServerError)
		}
	} else {
		html := `<html><body>
                   <a href="/login/google">Войти через Google</a><br>
                   <a href="/login/linkedin">Войти через LinkedIn</a>
                 </body></html>`
		fmt.Fprint(w, html)
	}
}

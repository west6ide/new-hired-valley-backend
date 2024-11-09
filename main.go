package main

import (
	"fmt"
	"hired-valley-backend/config"
	"hired-valley-backend/controllers"
	"hired-valley-backend/controllers/authentication"
	"hired-valley-backend/controllers/course"
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

	http.HandleFunc("/", handleHome)
	http.HandleFunc("/login/google", authentication.HandleGoogleLogin)
	http.HandleFunc("/callback/google", authentication.HandleGoogleCallback)
	http.HandleFunc("/login/linkedin", authentication.HandleLinkedInLogin)
	http.HandleFunc("/callback/linkedin", authentication.HandleLinkedInCallback)

	http.HandleFunc("/register", authentication.Register)
	http.HandleFunc("/login", authentication.Login)
	http.HandleFunc("/profile", authentication.GetProfile)
	http.HandleFunc("/logout", authentication.Logout)

	http.HandleFunc("/profile/update", authentication.UpdateProfile)
	http.HandleFunc("/users/search", authentication.SearchUsers)

	http.HandleFunc("/list/courses", course.ListCourses)
	http.HandleFunc("/create/courses", course.CreateCourse)
	http.HandleFunc("/upload/video", course.UploadVideo)
	http.HandleFunc("/list/lessons", course.ListLessons)
	http.HandleFunc("/create/lessons", course.CreateLesson)

	http.HandleFunc("/create/stories", controllers.CreateStory)
	http.HandleFunc("/list/stories", controllers.GetActiveStories)
	http.HandleFunc("/stories/archive", controllers.ArchiveStory) // Параметр id передается как query параметр

	http.HandleFunc("/mentors", mentors.CreateMentorProfile)
	http.HandleFunc("/mentors/:id", mentors.GetMentorProfile)
	http.HandleFunc("/mentors/:id", mentors.UpdateMentorProfile)
	http.HandleFunc("/mentors/:id", mentors.DeleteMentorProfile)

	// CRUD для AvailableTime (изменены маршруты, чтобы избежать конфликта)
	http.HandleFunc("/mentors/:id/availability", mentors.AddAvailableTime)
	http.HandleFunc("/mentors/:id/availability", mentors.GetAvailableTimes)
	http.HandleFunc("/availability/:id", mentors.UpdateAvailableTime)
	http.HandleFunc("/availability/:id", mentors.DeleteAvailableTime)

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
		case users.GoogleUser:
			html := fmt.Sprintf(`<html><body>
				<p>Добро пожаловать, %s!</p>
				<a href="/logout">Выйти</a><br>
				<form action="/google-logout" method="post">
					<button type="submit">Выйти из Google</button>
				</form>
			</body></html>`, usr.FirstName)
			fmt.Fprint(w, html)
		case users.LinkedInUser:
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

func RemoveExpiredStories() {
	for {
		config.DB.Where("expires_at <= ?", time.Now()).Delete(&users.Story{})
		time.Sleep(1 * time.Hour) // Запуск каждые 1 час
	}
}

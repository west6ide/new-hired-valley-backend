package main

import (
	"fmt"
	"hired-valley-backend/config"
	"hired-valley-backend/controllers/authentication"
	"hired-valley-backend/controllers/careers"
	"hired-valley-backend/controllers/course"
	"hired-valley-backend/controllers/recommendations"
	"hired-valley-backend/controllers/stories"
	"hired-valley-backend/models/career"
	"hired-valley-backend/models/courses"
	"hired-valley-backend/models/recommend"
	"hired-valley-backend/models/story"
	"hired-valley-backend/models/users"
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
		&users.User{},
		&users.GoogleUser{},
		&users.LinkedInUser{},
		&courses.Course{},
		&courses.Lesson{},
		&story.Story{},
		&story.Reaction{},
		&story.ViewStory{},
		&story.Comment{},
		&story.Notification{},
		&recommend.Recommendation{},
		&career.PlanCareer{},
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

	// authorization endpoints
	http.HandleFunc("/", handleHome)
	http.HandleFunc("/login/google", authentication.HandleGoogleLogin)
	http.HandleFunc("/callback/google", authentication.HandleGoogleCallback)
	http.HandleFunc("/login/linkedin", authentication.HandleLinkedInLogin)
	http.HandleFunc("/callback/linkedin", authentication.HandleLinkedInCallback)
	http.HandleFunc("/register", authentication.Register)
	http.HandleFunc("/login", authentication.Login)
	http.HandleFunc("/profile", authentication.GetProfile)
	http.HandleFunc("/logout", authentication.Logout)

	//users profile endpoints
	http.HandleFunc("/profile/update", authentication.UpdateProfile)
	http.HandleFunc("/users/search", authentication.SearchUsers)

	//courses endpoints
	http.HandleFunc("/list/courses", course.ListCourses)
	http.HandleFunc("/create/courses", course.CreateCourse)
	http.HandleFunc("/get/courses", course.GetCourseByID)
	http.HandleFunc("/update/courses", course.UpdateCourse)
	http.HandleFunc("/delete/courses", course.DeleteCourse)

	//lessons endpoints
	http.HandleFunc("/list/lessons", course.ListLessons)
	http.HandleFunc("/create/lessons", course.CreateLesson)
	http.HandleFunc("/get/lessons", course.GetLessonByID)
	http.HandleFunc("/update/lessons", course.UpdateLesson)
	http.HandleFunc("/delete/lessons", course.DeleteLesson)
	http.HandleFunc("/upload-video-to-lesson", course.UploadVideoToLesson)

	//stories endpoints
	http.HandleFunc("/create/stories", stories.CreateStory)
	http.HandleFunc("/stories/get/user", stories.GetUserStories)
	http.HandleFunc("/update/stories", stories.UpdateStory)
	http.HandleFunc("/delete/stories", stories.DeleteStory)
	http.HandleFunc("/stories/view", stories.ViewStory)
	http.HandleFunc("/stories/archive", stories.ArchiveStory)

	//reactions endpoints
	http.HandleFunc("/add/stories/reactions", func(w http.ResponseWriter, r *http.Request) {
		stories.AddReaction(w, r, config.DB)
	})
	http.HandleFunc("/get/stories/reactions", func(w http.ResponseWriter, r *http.Request) {
		stories.GetReactions(w, r, config.DB)
	})
	http.HandleFunc("/update/stories/reactions", func(w http.ResponseWriter, r *http.Request) {
		stories.UpdateReaction(w, r, config.DB)
	})
	http.HandleFunc("/delete/stories/reactions", func(w http.ResponseWriter, r *http.Request) {
		stories.DeleteReaction(w, r, config.DB)
	})

	// comments endpoints
	http.HandleFunc("/create/stories/comments", func(w http.ResponseWriter, r *http.Request) {
		stories.CreateComment(w, r, config.DB)
	})
	http.HandleFunc("/get/stories/comments", func(w http.ResponseWriter, r *http.Request) {
		stories.GetComments(w, r, config.DB)
	})
	http.HandleFunc("/update/stories/comments", func(w http.ResponseWriter, r *http.Request) {
		stories.UpdateComment(w, r, config.DB)
	})
	http.HandleFunc("/delete/stories/comments", func(w http.ResponseWriter, r *http.Request) {
		stories.DeleteComment(w, r, config.DB)
	})

	//notifications endpoints
	http.HandleFunc("/get/notifications", func(w http.ResponseWriter, r *http.Request) {
		stories.GetNotifications(w, r, config.DB)
	})
	http.HandleFunc("/notifications/read", func(w http.ResponseWriter, r *http.Request) {
		stories.MarkNotificationAsRead(w, r, config.DB)
	})
	http.HandleFunc("/delete/notifications", func(w http.ResponseWriter, r *http.Request) {
		stories.DeleteNotification(w, r, config.DB)
	})

	// AI  endpoints
	http.HandleFunc("/generate-recommendations", recommendations.GenerateRecommendationsHandler)
	http.HandleFunc("/careersPlan", careers.GenerateCareerPlanHandler)

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

package controllers

import (
	"github.com/markbates/goth"
	"github.com/markbates/goth/gothic"
	"github.com/markbates/goth/providers/linkedin"
	"html/template"
	"net/http"
	"os"
	"sort"
)

// homeHandler обрабатывает запросы к главной странице
func HomeHandler(res http.ResponseWriter, request *http.Request) {
	providers := getAuthProviders()
	t, err := template.New("index").Parse(indexTemplate)
	if err != nil {
		http.Error(res, err.Error(), http.StatusInternalServerError)
		return
	}
	t.Execute(res, providers)
}

// authHandler инициирует процесс авторизации
func AuthHandler(res http.ResponseWriter, req *http.Request) {
	gothic.BeginAuthHandler(res, req)
}

// authCallbackHandler обрабатывает callback после авторизации
func AuthCallbackHandler(res http.ResponseWriter, req *http.Request) {
	user, err := gothic.CompleteUserAuth(res, req)
	if err != nil {
		http.Error(res, err.Error(), http.StatusInternalServerError)
		return
	}
	t, err := template.New("user").Parse(userTemplate)
	if err != nil {
		http.Error(res, err.Error(), http.StatusInternalServerError)
		return
	}
	t.Execute(res, user)
}

// logoutHandler выполняет выход пользователя
func LogoutHandler(res http.ResponseWriter, req *http.Request) {
	gothic.Logout(res, req)
	http.Redirect(res, req, "/", http.StatusTemporaryRedirect)
}

// initAuthProviders инициализирует провайдеры авторизации
func InitAuthProviders() {
	goth.UseProviders(
		linkedin.New(
			os.Getenv("LINKEDIN_CLIENT_ID"),
			os.Getenv("LINKEDIN_CLIENT_SECRET"),
			"https://new-hired-valley-backend.onrender.com/callback/linkedin",
		),
		// Добавьте здесь другие провайдеры при необходимости
	)
}

// getAuthProviders возвращает список доступных провайдеров
func getAuthProviders() []string {
	providers := []string{"linkedin"}
	sort.Strings(providers)
	return providers
}

// indexTemplate — шаблон для главной страницы
var indexTemplate = `{{range .}}
    <p><a href="/auth/{{.}}">Войти через {{.}}</a></p>
{{end}}`

// userTemplate — шаблон для отображения информации о пользователе
var userTemplate = `
<p><a href="/logout/{{.Provider}}">Выйти</a></p>
<p>Имя: {{.Name}}</p>
<p>Email: {{.Email}}</p>
<p>Никнейм: {{.NickName}}</p>
<p>Местоположение: {{.Location}}</p>
<p>URL аватара: {{.AvatarURL}} <img src="{{.AvatarURL}}" alt="Аватар"></p>
<p>ID пользователя: {{.UserID}}</p>
<p>Токен доступа: {{.AccessToken}}</p>
<p>Дата истечения токена: {{.ExpiresAt}}</p>
<p>Токен обновления: {{.RefreshToken}}</p>
`

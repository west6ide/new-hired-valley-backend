package mentors

import (
	"encoding/json"
	"github.com/dgrijalva/jwt-go"
	"hired-valley-backend/config"
	"hired-valley-backend/controllers/authentication"
	"hired-valley-backend/models/users"
	"log"
	"net/http"
	"strconv"
	"strings"
)

// Создание профиля наставника
func CreateMentorProfile(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Получение ID авторизованного пользователя из токена
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		http.Error(w, "Authorization header required", http.StatusUnauthorized)
		return
	}

	// Убираем "Bearer " из начала заголовка
	tokenString := strings.TrimPrefix(authHeader, "Bearer ")
	claims := &authentication.Claims{}

	token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
		return authentication.JwtKey, nil
	})

	if err != nil || !token.Valid {
		http.Error(w, "Invalid token", http.StatusUnauthorized)
		return
	}

	// Проверка, что пользователь имеет роль наставника
	var user users.User
	if err := config.DB.First(&user, claims.UserID).Error; err != nil {
		http.Error(w, "User not found", http.StatusUnauthorized)
		return
	}

	if user.Role != "mentor" {
		http.Error(w, "Only mentors can create a mentor profile", http.StatusForbidden)
		return
	}

	// Создание профиля наставника, если роль пользователя - "mentor"
	var profile users.MentorProfile
	if err := json.NewDecoder(r.Body).Decode(&profile); err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	// Присваиваем профильному полю UserID значение из авторизованного пользователя
	profile.UserID = user.ID

	if err := config.DB.Create(&profile).Error; err != nil {
		http.Error(w, "Could not create mentor profile", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(profile)
}

// Получить наставников по фильтрам
func GetMentors(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var mentors []users.MentorProfile
	industry := r.URL.Query().Get("industry")
	specialization := r.URL.Query().Get("specialization")

	query := config.DB.Model(&users.MentorProfile{})
	if industry != "" {
		query = query.Where("industry = ?", industry)
	}
	if specialization != "" {
		query = query.Where("specialization = ?", specialization)
	}
	if err := query.Find(&mentors).Error; err != nil {
		http.Error(w, "Could not retrieve mentors", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(mentors)
}

// Создание доступности
func CreateAvailability(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Получение токена из заголовка Authorization
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		http.Error(w, "Authorization header required", http.StatusUnauthorized)
		return
	}

	// Убираем "Bearer " из начала заголовка
	tokenString := strings.TrimPrefix(authHeader, "Bearer ")
	claims := &authentication.Claims{}

	token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
		return authentication.JwtKey, nil
	})

	if err != nil || !token.Valid {
		http.Error(w, "Invalid token", http.StatusUnauthorized)
		return
	}

	var user users.User
	if err := config.DB.First(&user, claims.UserID).Error; err != nil {
		http.Error(w, "User not found", http.StatusUnauthorized)
		return
	}

	if user.Role != "mentor" {
		http.Error(w, "Only mentors can create availability", http.StatusForbidden)
		return
	}

	// Создание доступности
	var availability users.Availability
	if err := json.NewDecoder(r.Body).Decode(&availability); err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	// Логирование полученных значений
	log.Printf("Доступность наставника ID %d с временем %s - %s", availability.MentorID, availability.StartTime, availability.EndTime)

	// Присваиваем профильному полю MentorID значение из авторизованного пользователя
	availability.MentorID = user.ID

	// Сохранение доступности в базе данных
	if err := config.DB.Create(&availability).Error; err != nil {
		http.Error(w, "Could not create availability", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(availability)
}

// Создание сеанса наставничества
func CreateMentorshipSession(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		http.Error(w, "Authorization header required", http.StatusUnauthorized)
		return
	}

	tokenString := strings.TrimPrefix(authHeader, "Bearer ")
	claims := &authentication.Claims{}

	token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
		return authentication.JwtKey, nil
	})

	if err != nil || !token.Valid {
		http.Error(w, "Invalid token", http.StatusUnauthorized)
		return
	}

	var session users.MentorshipSession
	if err := json.NewDecoder(r.Body).Decode(&session); err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	// Проверка доступности времени
	var availability users.Availability
	if err := config.DB.Where("mentor_id = ? AND start_time <= ? AND end_time >= ?", session.MentorID, session.Date, session.Date).First(&availability).Error; err != nil {
		http.Error(w, "No available slot", http.StatusNotFound)
		return
	}

	session.UserID = claims.UserID

	if err := config.DB.Create(&session).Error; err != nil {
		http.Error(w, "Could not create session", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(session)
}

// Обновить статус сеанса
func UpdateSessionStatus(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPut {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Извлечение ID из URL
	pathParts := strings.Split(r.URL.Path, "/")
	if len(pathParts) < 3 {
		http.Error(w, "Invalid session ID", http.StatusBadRequest)
		return
	}

	sessionID, err := strconv.Atoi(pathParts[2])
	if err != nil {
		http.Error(w, "Invalid session ID", http.StatusBadRequest)
		return
	}

	var session users.MentorshipSession
	if err := config.DB.First(&session, sessionID).Error; err != nil {
		http.Error(w, "Session not found", http.StatusNotFound)
		return
	}

	newStatus := r.FormValue("status")
	session.Status = newStatus

	if err := config.DB.Save(&session).Error; err != nil {
		http.Error(w, "Could not update session status", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(session)
}

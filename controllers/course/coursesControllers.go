package course

import (
	"encoding/json"
	"github.com/dgrijalva/jwt-go"
	"hired-valley-backend/config"
	"hired-valley-backend/controllers/authentication"
	"hired-valley-backend/models/courses"
	"net/http"
	"strings"
)

// Листинг курсов
func ListCourses(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var courses []courses.Course
	if err := config.DB.Find(&courses).Error; err != nil {
		http.Error(w, "Failed to list courses", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(courses)
}

// Функция для создания курса, доступна только для пользователей с ролью instructor
func CreateCourse(w http.ResponseWriter, r *http.Request) {
	// Извлекаем токен из заголовка Authorization
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		http.Error(w, "Authorization header required", http.StatusUnauthorized)
		return
	}

	// Удаляем префикс "Bearer "
	tokenString := strings.TrimPrefix(authHeader, "Bearer ")

	// Парсим токен и извлекаем claims
	claims := &authentication.Claims{}
	token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
		return []byte(authentication.JwtKey), nil
	})

	if err != nil || !token.Valid {
		http.Error(w, "Invalid token", http.StatusUnauthorized)
		return
	}

	// Проверяем, что роль пользователя — instructor
	if claims.Role != "mentor" {
		http.Error(w, "Only instructors can create courses", http.StatusForbidden)
		return
	}

	// Декодируем данные курса из тела запроса
	var course courses.Course
	if err := json.NewDecoder(r.Body).Decode(&course); err != nil {
		http.Error(w, "Invalid input", http.StatusBadRequest)
		return
	}

	// Устанавливаем идентификатор инструктора из токена
	course.InstructorID = claims.UserID

	// Сохраняем курс в базе данных
	if err := config.DB.Create(&course).Error; err != nil {
		http.Error(w, "Failed to create course", http.StatusInternalServerError)
		return
	}

	// Возвращаем успешный ответ с данными курса
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(course)
}

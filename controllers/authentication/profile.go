package authentication

import (
	"encoding/json"
	"gorm.io/gorm"
	"net/http"
	"strings"

	"github.com/dgrijalva/jwt-go"
	"hired-valley-backend/config"
	"hired-valley-backend/models"
)

func UpdateProfile(w http.ResponseWriter, r *http.Request) {
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		http.Error(w, "Authorization header required", http.StatusUnauthorized)
		return
	}

	tokenString := strings.TrimPrefix(authHeader, "Bearer ")
	claims := &Claims{}
	token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
		return jwtKey, nil
	})
	if err != nil || !token.Valid {
		http.Error(w, "Invalid token", http.StatusUnauthorized)
		return
	}

	email := claims.Email

	var user models.User
	if err := config.DB.Where("email = ?", email).First(&user).Error; err != nil {
		http.Error(w, "User not found", http.StatusNotFound)
		return
	}

	var updatedProfile models.User
	if err := json.NewDecoder(r.Body).Decode(&updatedProfile); err != nil {
		http.Error(w, "Invalid input", http.StatusBadRequest)
		return
	}

	// Обработка добавления навыков
	for _, skill := range updatedProfile.Skills {
		var existingSkill models.Skill
		if err := config.DB.Where("name = ?", skill.Name).First(&existingSkill).Error; err != nil {
			if err == gorm.ErrRecordNotFound {
				// Если навык не найден, создаем его
				newSkill := models.Skill{Name: skill.Name}
				config.DB.Create(&newSkill)
				user.Skills = append(user.Skills, newSkill)
			} else {
				http.Error(w, "Error updating skills", http.StatusInternalServerError)
				return
			}
		} else {
			// Если навык найден, добавляем его к пользователю
			user.Skills = append(user.Skills, existingSkill)
		}
	}

	// Обработка добавления интересов
	for _, interest := range updatedProfile.Interests {
		var existingInterest models.Interest
		if err := config.DB.Where("name = ?", interest.Name).First(&existingInterest).Error; err != nil {
			if err == gorm.ErrRecordNotFound {
				// Если интерес не найден, создаем его
				newInterest := models.Interest{Name: interest.Name}
				config.DB.Create(&newInterest)
				user.Interests = append(user.Interests, newInterest)
			} else {
				http.Error(w, "Error updating interests", http.StatusInternalServerError)
				return
			}
		} else {
			// Если интерес найден, добавляем его к пользователю
			user.Interests = append(user.Interests, existingInterest)
		}
	}

	// Обновление остальных данных профиля
	user.Position = updatedProfile.Position
	user.City = updatedProfile.City
	user.Income = updatedProfile.Income
	user.Visibility = updatedProfile.Visibility

	if err := config.DB.Save(&user).Error; err != nil {
		http.Error(w, "Error updating profile", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(user)
}

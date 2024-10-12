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

	// Начало транзакции
	tx := config.DB.Begin()

	// Обработка добавления навыков
	var updatedSkills []models.Skill
	for _, skill := range updatedProfile.Skills {
		var existingSkill models.Skill
		if err := tx.Where("name = ?", skill.Name).First(&existingSkill).Error; err != nil {
			if err == gorm.ErrRecordNotFound {
				// Если навык не найден, создаем его
				newSkill := models.Skill{Name: skill.Name}
				if err := tx.Create(&newSkill).Error; err != nil {
					tx.Rollback()
					http.Error(w, "Error creating new skill", http.StatusInternalServerError)
					return
				}
				updatedSkills = append(updatedSkills, newSkill)
			} else {
				tx.Rollback()
				http.Error(w, "Error finding skill", http.StatusInternalServerError)
				return
			}
		} else {
			// Добавляем уже существующий навык в список
			updatedSkills = append(updatedSkills, existingSkill)
		}
	}
	// Обновляем навыки пользователя
	if err := tx.Model(&user).Association("Skills").Replace(updatedSkills); err != nil {
		tx.Rollback()
		http.Error(w, "Error updating skills", http.StatusInternalServerError)
		return
	}

	// Обработка добавления интересов
	var updatedInterests []models.Interest
	for _, interest := range updatedProfile.Interests {
		var existingInterest models.Interest
		if err := tx.Where("name = ?", interest.Name).First(&existingInterest).Error; err != nil {
			if err == gorm.ErrRecordNotFound {
				// Если интерес не найден, создаем его
				newInterest := models.Interest{Name: interest.Name}
				if err := tx.Create(&newInterest).Error; err != nil {
					tx.Rollback()
					http.Error(w, "Error creating new interest", http.StatusInternalServerError)
					return
				}
				updatedInterests = append(updatedInterests, newInterest)
			} else {
				tx.Rollback()
				http.Error(w, "Error finding interest", http.StatusInternalServerError)
				return
			}
		} else {
			// Добавляем уже существующий интерес в список
			updatedInterests = append(updatedInterests, existingInterest)
		}
	}
	// Обновляем интересы пользователя
	if err := tx.Model(&user).Association("Interests").Replace(updatedInterests); err != nil {
		tx.Rollback()
		http.Error(w, "Error updating interests", http.StatusInternalServerError)
		return
	}

	// Обновляем остальные данные профиля
	user.Position = updatedProfile.Position
	user.City = updatedProfile.City
	user.Income = updatedProfile.Income
	user.Visibility = updatedProfile.Visibility

	if err := tx.Save(&user).Error; err != nil {
		tx.Rollback()
		http.Error(w, "Error updating profile", http.StatusInternalServerError)
		return
	}

	// Коммит транзакции
	tx.Commit()

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(user)
}

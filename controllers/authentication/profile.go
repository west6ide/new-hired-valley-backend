package authentication

import (
	"encoding/json"
	"gorm.io/gorm"
	"hired-valley-backend/models/users"
	"net/http"
	"strings"

	"github.com/dgrijalva/jwt-go"
	"hired-valley-backend/config"
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
		return JwtKey, nil
	})
	if err != nil || !token.Valid {
		http.Error(w, "Invalid token", http.StatusUnauthorized)
		return
	}

	email := claims.Email

	var user users.User
	if err := config.DB.Where("email = ?", email).First(&user).Error; err != nil {
		http.Error(w, "User not found", http.StatusNotFound)
		return
	}

	var updatedProfile users.User
	if err := json.NewDecoder(r.Body).Decode(&updatedProfile); err != nil {
		http.Error(w, "Invalid input", http.StatusBadRequest)
		return
	}

	// Начало транзакции
	tx := config.DB.Begin()

	// Обновление навыков
	var updatedSkills []users.Skill
	for _, skill := range updatedProfile.Skills {
		var existingSkill users.Skill
		if err := tx.Where("name = ?", skill.Name).First(&existingSkill).Error; err != nil {
			if err == gorm.ErrRecordNotFound {
				newSkill := users.Skill{Name: skill.Name}
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
			updatedSkills = append(updatedSkills, existingSkill)
		}
	}

	// Привязка обновленных навыков к пользователю
	if err := tx.Model(&user).Association("Skills").Replace(updatedSkills); err != nil {
		tx.Rollback()
		http.Error(w, "Error updating skills", http.StatusInternalServerError)
		return
	}

	// Обновление интересов
	var updatedInterests []users.Interest
	for _, interest := range updatedProfile.Interests {
		var existingInterest users.Interest
		if err := tx.Where("name = ?", interest.Name).First(&existingInterest).Error; err != nil {
			if err == gorm.ErrRecordNotFound {
				newInterest := users.Interest{Name: interest.Name}
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
			updatedInterests = append(updatedInterests, existingInterest)
		}
	}

	// Привязка обновленных интересов к пользователю
	if err := tx.Model(&user).Association("Interests").Replace(updatedInterests); err != nil {
		tx.Rollback()
		http.Error(w, "Error updating interests", http.StatusInternalServerError)
		return
	}

	// Обновление других полей профиля
	user.Position = updatedProfile.Position
	user.City = updatedProfile.City
	user.Income = updatedProfile.Income
	user.Visibility = updatedProfile.Visibility
	user.Company = updatedProfile.Company
	user.Industry = updatedProfile.Industry

	// Проверка и изменение роли (только для администраторов)
	if claims.Role == "admin" {
		if updatedProfile.Role == "user" || updatedProfile.Role == "mentor" || updatedProfile.Role == "admin" {
			user.Role = updatedProfile.Role
		} else {
			tx.Rollback()
			http.Error(w, "Invalid role", http.StatusBadRequest)
			return
		}
	} else {
		// Запрет на изменение роли для обычных пользователей
		if updatedProfile.Role != "" && updatedProfile.Role != user.Role {
			tx.Rollback()
			http.Error(w, "Only admin can change role", http.StatusForbidden)
			return
		}
	}

	// Сохранение обновленного пользователя
	if err := tx.Save(&user).Error; err != nil {
		tx.Rollback()
		http.Error(w, "Error updating profile", http.StatusInternalServerError)
		return
	}

	// Коммит транзакции
	tx.Commit()

	// Возврат обновленного профиля
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(user)
}

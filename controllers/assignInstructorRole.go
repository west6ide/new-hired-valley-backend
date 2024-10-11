package controllers

import (
	"encoding/json"
	"github.com/gorilla/mux"
	"hired-valley-backend/config"
	"hired-valley-backend/models"
	"net/http"
	"strconv"
)

// Назначение роли "instructor" пользователю
func AssignInstructorRole(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	userID, _ := strconv.Atoi(params["user_id"])

	// Поиск пользователя по ID
	var user models.User
	if err := config.DB.First(&user, userID).Error; err != nil {
		http.Error(w, "Пользователь не найден", http.StatusNotFound)
		return
	}

	// Назначаем пользователю роль "instructor"
	user.Role = "instructor"
	if err := config.DB.Save(&user).Error; err != nil {
		http.Error(w, "Ошибка при обновлении пользователя", http.StatusInternalServerError)
		return
	}

	// Возвращаем обновленного пользователя
	json.NewEncoder(w).Encode(user)
}

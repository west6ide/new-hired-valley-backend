package mentors

import (
	"encoding/json"
	"fmt"
	"hired-valley-backend/config"
	"hired-valley-backend/controllers/authentication"
	"hired-valley-backend/models/users"
	"net/http"
	"strconv"
	"time"
)

func MentorsHandler(w http.ResponseWriter, r *http.Request) {
	user, err := authentication.ValidateToken(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}

	if r.Method == http.MethodPost {
		// Создаем новый профиль ментора
		var mentorProfile users.MentorProfile
		if err := json.NewDecoder(r.Body).Decode(&mentorProfile); err != nil {
			http.Error(w, "Invalid request body", http.StatusBadRequest)
			return
		}

		// Проверяем роль пользователя
		if user.Role != "mentor" {
			http.Error(w, "User is not authorized to create a mentor profile", http.StatusForbidden)
			return
		}

		// Присваиваем user.ID в поле id и user_id
		mentorProfile.ID = user.ID // Записываем user.ID в поле id
		mentorProfile.UserID = user.ID

		// Создаем профиль ментора в базе данных
		if err := config.DB.Create(&mentorProfile).Error; err != nil {
			http.Error(w, "Error creating mentor profile", http.StatusInternalServerError)
			return
		}

		// Прелоадим данные пользователя для ответа
		if err := config.DB.Preload("User").First(&mentorProfile, mentorProfile.ID).Error; err != nil {
			http.Error(w, "Error fetching created mentor profile", http.StatusInternalServerError)
			return
		}

		// Возвращаем успешный ответ
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(mentorProfile)

	} else if r.Method == http.MethodGet {
		// Обработчик GET-запроса для получения списка менторов
		var mentors []users.MentorProfile
		query := config.DB.Preload("User")

		// Фильтрация по навыкам
		skills := r.URL.Query().Get("skills")
		if skills != "" {
			query = query.Where("skills LIKE ?", fmt.Sprintf("%%%s%%", skills))
		}

		// Фильтрация по цене
		priceRange := r.URL.Query().Get("price_range")
		if priceRange != "" {
			price, _ := strconv.ParseFloat(priceRange, 64)
			query = query.Where("price_per_hour <= ?", price)
		}

		// Получаем список менторов
		if err := query.Find(&mentors).Error; err != nil {
			http.Error(w, "Error fetching mentors", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(mentors)

	} else {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func CreateSlotHandler(w http.ResponseWriter, r *http.Request) {
	// Проверка метода запроса
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Валидация токена и получение данных пользователя
	user, err := authentication.ValidateToken(r)
	if err != nil {
		http.Error(w, "Unauthorized: "+err.Error(), http.StatusUnauthorized)
		return
	}

	// Проверка роли пользователя
	if user.Role != "mentor" {
		http.Error(w, "Only mentors can create slots", http.StatusForbidden)
		return
	}

	// Получение профиля ментора по токену пользователя
	var mentorProfile users.MentorProfile
	if err := config.DB.First(&mentorProfile, "user_id = ?", user.ID).Error; err != nil {
		http.Error(w, "Mentor profile not found. Create a mentor profile first.", http.StatusBadRequest)
		return
	}

	// Декодирование данных слота
	var input struct {
		StartTime string `json:"start_time"`
		EndTime   string `json:"end_time"`
	}
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		http.Error(w, "Invalid input", http.StatusBadRequest)
		return
	}

	// Конвертация строк в time.Time
	startTime, err := time.Parse(time.RFC3339, input.StartTime)
	if err != nil {
		http.Error(w, "Invalid start_time format. Use RFC3339 format.", http.StatusBadRequest)
		return
	}
	endTime, err := time.Parse(time.RFC3339, input.EndTime)
	if err != nil {
		http.Error(w, "Invalid end_time format. Use RFC3339 format.", http.StatusBadRequest)
		return
	}

	// Создание нового слота
	slot := users.Slot{
		MentorID:  mentorProfile.ID, // Используем mentor_id из профиля
		StartTime: startTime,
		EndTime:   endTime,
		IsBooked:  false,
	}

	// Проверка на дубликаты
	var existingSlot users.Slot
	if err := config.DB.Where("mentor_id = ? AND start_time = ? AND end_time = ?", slot.MentorID, slot.StartTime, slot.EndTime).First(&existingSlot).Error; err == nil {
		http.Error(w, "Slot already exists for this time range", http.StatusConflict)
		return
	}

	// Сохранение слота
	if err := config.DB.Create(&slot).Error; err != nil {
		fmt.Printf("Error creating slot: %v\n", err)
		http.Error(w, fmt.Sprintf("Error creating slot: %v", err), http.StatusInternalServerError)
		return
	}

	// Возвращаем успешный ответ
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(slot)
}

func BookSlotHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	user, err := authentication.ValidateToken(r) // Пользователь, который бронирует слот
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}

	var slot users.Slot
	if err := json.NewDecoder(r.Body).Decode(&slot); err != nil {
		http.Error(w, "Invalid input", http.StatusBadRequest)
		return
	}

	var mentor users.MentorProfile
	if err := config.DB.First(&mentor, slot.MentorID).Error; err != nil {
		http.Error(w, "Mentor not found", http.StatusBadRequest)
		return
	}

	var existingSlot users.Slot
	if err := config.DB.First(&existingSlot, "mentor_id = ? AND start_time = ? AND end_time = ?", slot.MentorID, slot.StartTime, slot.EndTime).Error; err == nil && existingSlot.IsBooked {
		http.Error(w, "Slot is already booked", http.StatusConflict)
		return
	}

	// Заполнение данных о слоте
	slot.IsBooked = true
	slot.UserID = &user.ID // Сохраняем ID пользователя, который забронировал слот
	if err := config.DB.Create(&slot).Error; err != nil {
		fmt.Printf("Error creating slot: %v\n", err)
		http.Error(w, "Error booking slot", http.StatusInternalServerError)
		return
	}

	// Уведомление для ментора
	notification := users.NotificationMentor{
		UserID:  mentor.UserID,
		Message: fmt.Sprintf("Your slot from %s to %s has been booked by %s.", slot.StartTime, slot.EndTime, user.Name),
		IsRead:  false,
	}
	config.DB.Create(&notification)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(slot)
}

func SlotsHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	_, err := authentication.ValidateToken(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}

	var slots []users.Slot
	mentorID := r.URL.Query().Get("mentor_id")
	if mentorID != "" {
		query := config.DB.Where("mentor_id = ?", mentorID)
		if err := query.Find(&slots).Error; err != nil {
			http.Error(w, "Error fetching slots", http.StatusInternalServerError)
			return
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(slots)
}
func MentorBookedSlotsHandler(w http.ResponseWriter, r *http.Request) {
	user, err := authentication.ValidateToken(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}

	// Проверка роли ментора
	if user.Role != "mentor" {
		http.Error(w, "Only mentors can view booked slots", http.StatusForbidden)
		return
	}

	// Получаем слоты для ментора с забронированными пользователями
	var slots []users.Slot
	if err := config.DB.Preload("User").Where("mentor_id = ? AND is_booked = ?", user.ID, true).Find(&slots).Error; err != nil {
		http.Error(w, "Error fetching booked slots", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(slots)
}

func NotificationsHandler(w http.ResponseWriter, r *http.Request) {
	user, err := authentication.ValidateToken(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}

	switch r.Method {
	case http.MethodGet:
		var notifications []users.NotificationMentor
		if err := config.DB.Where("user_id = ?", user.ID).Find(&notifications).Error; err != nil {
			http.Error(w, "Error fetching notifications", http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(notifications)

	case http.MethodPost:
		var notification users.NotificationMentor
		if err := json.NewDecoder(r.Body).Decode(&notification); err != nil {
			http.Error(w, "Invalid request body", http.StatusBadRequest)
			return
		}

		// Assign user ID to the notification
		notification.UserID = user.ID
		if err := config.DB.Create(&notification).Error; err != nil {
			http.Error(w, "Error creating notification", http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(notification)

	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

package controllers

import (
	"github.com/gofiber/fiber/v2"
	"hired-valley-backend/config"
	"hired-valley-backend/models/users"
)

// Создать профиль наставника
func CreateMentorProfile(c *fiber.Ctx) error {
	var profile users.MentorProfile
	if err := c.BodyParser(&profile); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request"})
	}

	if err := config.DB.Create(&profile).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Could not create mentor profile"})
	}

	return c.JSON(profile)
}

// Найти наставников по фильтрам
func GetMentors(c *fiber.Ctx) error {
	var mentors []users.MentorProfile
	industry := c.Query("industry")
	specialization := c.Query("specialization")

	query := config.DB.Model(&users.MentorProfile{})
	if industry != "" {
		query = query.Where("industry = ?", industry)
	}
	if specialization != "" {
		query = query.Where("specialization = ?", specialization)
	}
	if err := query.Find(&mentors).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Could not retrieve mentors"})
	}
	return c.JSON(mentors)
}

// Создать сеанс наставничества
func CreateMentorshipSession(c *fiber.Ctx) error {
	var session users.MentorshipSession
	if err := c.BodyParser(&session); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request"})
	}

	// Проверка доступности времени
	var availability users.Availability
	if err := config.DB.Where("mentor_id = ? AND start_time <= ? AND end_time >= ?", session.MentorID, session.Date, session.Date).First(&availability).Error; err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "No available slot"})
	}

	if err := config.DB.Create(&session).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Could not create session"})
	}

	return c.JSON(session)
}

// Обновить статус сеанса
func UpdateSessionStatus(c *fiber.Ctx) error {
	sessionID := c.Params("id")
	var session users.MentorshipSession

	if err := config.DB.First(&session, sessionID).Error; err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "Session not found"})
	}

	newStatus := c.FormValue("status")
	session.Status = newStatus

	if err := config.DB.Save(&session).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Could not update session status"})
	}

	return c.JSON(session)
}

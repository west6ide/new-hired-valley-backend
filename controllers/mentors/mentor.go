package mentors

import (
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"hired-valley-backend/config"
	"hired-valley-backend/models/users"
	"net/http"
)

// CRUD для MentorProfile

func CreateMentorProfile(c *gin.Context) {
	var profile users.MentorProfile
	if err := c.ShouldBindJSON(&profile); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
		return
	}

	if err := config.DB.Create(&profile).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create profile"})
		return
	}

	c.JSON(http.StatusCreated, profile)
}

func GetMentorProfile(c *gin.Context) {
	var profile users.MentorProfile
	id := c.Param("id")

	if err := config.DB.Preload("Skills").Preload("Schedule").First(&profile, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "Profile not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve profile"})
		return
	}

	c.JSON(http.StatusOK, profile)
}

func UpdateMentorProfile(c *gin.Context) {
	var profile users.MentorProfile
	id := c.Param("id")

	if err := config.DB.First(&profile, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Profile not found"})
		return
	}

	if err := c.ShouldBindJSON(&profile); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
		return
	}

	if err := config.DB.Save(&profile).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update profile"})
		return
	}

	c.JSON(http.StatusOK, profile)
}

func DeleteMentorProfile(c *gin.Context) {
	var profile users.MentorProfile
	id := c.Param("id")

	if err := config.DB.First(&profile, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Profile not found"})
		return
	}

	if err := config.DB.Delete(&profile).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete profile"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Profile deleted"})
}

// CRUD для AvailableTime

func AddAvailableTime(c *gin.Context) {
	var availableTime users.AvailableTime

	if err := c.ShouldBindJSON(&availableTime); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
		return
	}

	if availableTime.StartTime.After(availableTime.EndTime) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Start time must be before end time"})
		return
	}

	if err := config.DB.Create(&availableTime).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create available time"})
		return
	}

	c.JSON(http.StatusCreated, availableTime)
}

func GetAvailableTimes(c *gin.Context) {
	var schedule []users.AvailableTime
	mentorID := c.Param("mentorID")

	if err := config.DB.Where("mentor_id = ?", mentorID).Find(&schedule).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve schedule"})
		return
	}

	c.JSON(http.StatusOK, schedule)
}

func UpdateAvailableTime(c *gin.Context) {
	var availableTime users.AvailableTime
	id := c.Param("id")

	if err := config.DB.First(&availableTime, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Available time not found"})
		return
	}

	if err := c.ShouldBindJSON(&availableTime); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
		return
	}

	if availableTime.StartTime.After(availableTime.EndTime) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Start time must be before end time"})
		return
	}

	if err := config.DB.Save(&availableTime).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update available time"})
		return
	}

	c.JSON(http.StatusOK, availableTime)
}

func DeleteAvailableTime(c *gin.Context) {
	var availableTime users.AvailableTime
	id := c.Param("id")

	if err := config.DB.First(&availableTime, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Available time not found"})
		return
	}

	if err := config.DB.Delete(&availableTime).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete available time"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Available time deleted"})
}

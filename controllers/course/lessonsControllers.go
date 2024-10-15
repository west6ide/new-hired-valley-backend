package course

import (
	"github.com/gin-gonic/gin"
	"hired-valley-backend/config"
	"hired-valley-backend/models/courses"
	"net/http"
	"strconv"
)

// Листинг уроков курса
func ListLessons(c *gin.Context) {
	courseID := c.Param("id")
	var lessons []courses.Lesson
	if err := config.DB.Where("course_id = ?", courseID).Find(&lessons).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to list lessons"})
		return
	}

	c.JSON(http.StatusOK, lessons)
}

// Создание урока
func CreateLesson(c *gin.Context) {
	courseIDStr := c.Param("id")

	// Преобразуем courseID из строки в uint
	courseID, err := strconv.ParseUint(courseIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid course ID"})
		return
	}

	var lesson courses.Lesson
	if err := c.BindJSON(&lesson); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input"})
		return
	}

	lesson.CourseID = uint(courseID) // Преобразуем к uint перед присвоением

	if err := config.DB.Create(&lesson).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create lesson"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Lesson created"})
}

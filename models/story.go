package models

//import (
//	"hired-valley-backend/config"
//	"hired-valley-backend/controllers"
//	"net/http"
//	"path/filepath"
//	"time"
//
//	"github.com/gin-gonic/gin"
//)
//
//type CreateStoryInput struct {
//	UserID uint `json:"user_id"`
//}
//
//func GetStories(c *gin.Context) {
//	var stories []controllers.Claims
//	now := time.Now()
//
//	// Получение всех активных историй (которые не истекли)
//	if err := config.DB.Where("expiry_at > ?", now).Preload("User").Find(&stories).Error; err != nil {
//		c.JSON(http.StatusBadRequest, gin.H{"error": "Could not fetch stories"})
//		return
//	}
//
//	c.JSON(http.StatusOK, gin.H{"data": stories})
//}
//
//func CreateStory(c *gin.Context) {
//	var input CreateStoryInput
//
//	// Проверка на валидность входных данных
//	if err := c.ShouldBindJSON(&input); err != nil {
//		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
//		return
//	}
//
//	// Получение файла
//	file, err := c.FormFile("file")
//	if err != nil {
//		c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to upload file"})
//		return
//	}
//
//	// Сохранение файла
//	filename := filepath.Base(file.Filename)
//	filepath := "uploads/" + filename
//	if err := c.SaveUploadedFile(file, filepath); err != nil {
//		c.JSON(http.StatusInternalServerError, gin.H{"error": "Could not save file"})
//		return
//	}
//
//	// Создание истории
//	story := GoogleUser{
//		ID: input.UserID,
//	}
//
//	if err := config.DB.Create(&story).Error; err != nil {
//		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create story"})
//		return
//	}
//
//	c.JSON(http.StatusOK, gin.H{"data": story})
//}
//
//func DeleteExpiredStories(c *gin.Context) {
//	now := time.Now()
//
//	// Удаление всех историй, срок которых истек
//	if err := config.DB.Where("expiry_at <= ?", now).Delete(&controllers.Claims{}).Error; err != nil {
//		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete expired stories"})
//		return
//	}
//
//	c.JSON(http.StatusOK, gin.H{"message": "Expired stories deleted"})
//}

// routes/story_routes.go

package routes

import (
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"hired-valley-backend/models"
	"net/http"
)

func RegisterStoryRoutes(router *gin.Engine, db *gorm.DB) {
	router.POST("/stories", func(c *gin.Context) {
		var story models.Story
		if err := c.ShouldBindJSON(&story); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		if err := models.CreateStory(db, &story); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "could not create story"})
			return
		}
		c.JSON(http.StatusCreated, story)
	})

	router.GET("/stories", func(c *gin.Context) {
		stories, err := models.GetActiveStories(db)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "could not fetch stories"})
			return
		}
		c.JSON(http.StatusOK, stories)
	})

	router.GET("/stories/:id", func(c *gin.Context) {
		id := c.Param("id")
		var story models.Story
		err := models.GetStoryByID(db, id)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "story not found"})
			return
		}
		c.JSON(http.StatusOK, story)
	})
}

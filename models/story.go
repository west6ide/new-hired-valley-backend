// models/story.go

package models

import (
	"time"

	"gorm.io/gorm"
)

type Story struct {
	ID        uint      `json:"id" gorm:"primaryKey"`
	UserID    uint      `json:"user_id"`
	Content   string    `json:"content"`
	CreatedAt time.Time `json:"created_at"`
	ExpiresAt time.Time `json:"expires_at"`
}

// Создание истории в базе данных
func CreateStory(db *gorm.DB, story *Story) error {
	story.CreatedAt = time.Now()
	story.ExpiresAt = story.CreatedAt.Add(24 * time.Hour)
	return db.Create(story).Error
}

// Получение всех актуальных историй
func GetActiveStories(db *gorm.DB) ([]Story, error) {
	var stories []Story
	err := db.Find(&stories).Error
	if err != nil {
		return nil, err
	}

	var validStories []Story
	for _, story := range stories {
		if story.ExpiresAt.After(time.Now()) {
			validStories = append(validStories, story)
		}
	}

	return validStories, nil
}

// Получение истории по ID
func GetStoryByID(db *gorm.DB, id string) *Story {
	var story Story
	if err := db.First(&story, id).Error; err != nil {
		return nil
	}
	if story.ExpiresAt.Before(time.Now()) {
		return nil // История истекла
	}
	return &story
}

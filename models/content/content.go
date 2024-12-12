package content

import (
	"gorm.io/gorm"
	"time"
)

type Content struct {
	ID          uint           `gorm:"primaryKey"`
	Title       string         `json:"title" gorm:"not null"`
	Description string         `json:"description" gorm:"type:text"`
	Tags        []string       `json:"tags" gorm:"type:text[]"`
	Category    string         `json:"category"`
	VideoLink   string         `json:"video_link"`
	YouTubeID   string         `json:"youtube_id"`
	AuthorID    uint           `json:"author_id"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `gorm:"index"`
}

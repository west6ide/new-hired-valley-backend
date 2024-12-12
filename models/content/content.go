package content

import (
	"github.com/lib/pq"
	"gorm.io/gorm"
	"time"
)

type Content struct {
	ID          uint           `gorm:"primaryKey"`
	Title       string         `json:"title" gorm:"not null"`
	Description string         `json:"description" gorm:"type:text"`
	Tags        pq.StringArray `json:"tags" gorm:"type:text[]"` // Храним как массив строк
	Category    string         `json:"category"`
	VideoLink   string         `json:"video_link"`
	YouTubeID   string         `json:"youtube_id"`
	AuthorID    uint           `json:"author_id"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `gorm:"index"`
}

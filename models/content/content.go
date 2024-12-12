package content

import (
	"gorm.io/gorm"
	"time"
)

type Content struct {
	ID          uint      `gorm:"primaryKey"`
	Title       string    `gorm:"not null"`
	Description string    `gorm:"type:text"`
	VideoURL    string    `gorm:"not null"`
	UserID      uint      `gorm:"not null"`    // ID пользователя, создавшего контент
	Tags        []string  `gorm:"type:text[]"` // Теги, определяющие тематику контента
	CreatedAt   time.Time `gorm:"default:current_timestamp"`
	UpdatedAt   time.Time
	DeletedAt   gorm.DeletedAt `gorm:"index"`
}

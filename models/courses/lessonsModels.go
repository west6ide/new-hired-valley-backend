package courses

import (
	"gorm.io/gorm"
	"time"
)

type Lesson struct {
	ID             uint      `gorm:"primaryKey"`
	CourseID       uint      `gorm:"not null"`
	Title          string    `gorm:"not null"`
	VimeoVideoLink string    `gorm:"type:varchar(255)"`
	CreatedAt      time.Time `gorm:"default:current_timestamp"`
	UpdatedAt      time.Time
	DeletedAt      gorm.DeletedAt `gorm:"index"`
}

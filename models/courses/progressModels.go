package courses

import (
	"gorm.io/gorm"
	"time"
)

type Progress struct {
	ID          uint       `gorm:"primaryKey"`
	UserID      uint       `gorm:"not null"`
	LessonID    uint       `gorm:"not null"`
	Completed   bool       `gorm:"default:false"`
	CompletedAt *time.Time // This field can be null
	UpdatedAt   time.Time
	DeletedAt   gorm.DeletedAt `gorm:"index"`
}

package courses

import (
	"gorm.io/gorm"
	"time"
)

type Course struct {
	ID           uint      `gorm:"primaryKey"`
	Title        string    `gorm:"not null"`
	Description  string    `gorm:"type:text"`
	InstructorID uint      `gorm:"not null"`
	CreatedAt    time.Time `gorm:"default:current_timestamp"`
	UpdatedAt    time.Time
	DeletedAt    gorm.DeletedAt `gorm:"index"`
	Lessons      []Lesson       `gorm:"foreignKey:CourseID"`
}

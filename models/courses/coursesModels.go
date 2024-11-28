package courses

import (
	"gorm.io/gorm"
	"hired-valley-backend/models/users"
	"time"
)

type Course struct {
	ID           uint       `gorm:"primaryKey"`
	Title        string     `json:"title" gorm:"not null"`
	Description  string     `json:"description" gorm:"type:text"`
	Price        float64    `json:"price"`
	InstructorID uint       `json:"instructor_id" gorm:"not null"`    // Внешний ключ на инструктора
	Instructor   users.User `json:"-" gorm:"foreignKey:InstructorID"` // Связь с таблицей users
	CreatedAt    time.Time  `gorm:"default:current_timestamp"`
	UpdatedAt    time.Time
	DeletedAt    gorm.DeletedAt `gorm:"index"`
	Lessons      []Lesson       `gorm:"foreignKey:CourseID"` // Связь с уроками
}
